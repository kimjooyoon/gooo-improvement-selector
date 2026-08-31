package selector

import (
	"fmt"
	"path/filepath"
)

type generatedCandidate struct {
	Schema      string   `json:"schema"`
	CandidateID string   `json:"candidate_id"`
	Semantic    string   `json:"semantic"`
	Operations  []string `json:"operations"`
	SourceDigest string  `json:"source_digest"`
	IRDigest    string   `json:"ir_digest"`
	InputDigest string   `json:"input_digest"`
	ContractDigest string `json:"contract_digest"`
}

func GenerateCandidates(ir SemanticIR, scenario Scenario, inputDigest, contractDigest, irDigest, outDir string) (CandidateManifest, error) {
	if ir.Schema != "gooo/improvement-selector/semantic-ir/v1" {
		return CandidateManifest{}, fmt.Errorf("unexpected IR schema %q", ir.Schema)
	}
	if scenario.ScenarioID == "" || scenario.Class == "" || scenario.Toolchain == "" {
		return CandidateManifest{}, fmt.Errorf("scenario identity is incomplete")
	}
	if inputDigest == "" || contractDigest == "" || irDigest == "" {
		return CandidateManifest{}, fmt.Errorf("candidate input digests are incomplete")
	}
	if err := EnsureCallerOwnedOutput(outDir); err != nil {
		return CandidateManifest{}, err
	}
	manifest := CandidateManifest{
		Schema:         "gooo/improvement-selector/candidate-manifest/v1",
		ScenarioID:     scenario.ScenarioID,
		SourceDigest:   ir.SourceDigest,
		IRDigest:       irDigest,
		InputDigest:    inputDigest,
		ContractDigest: contractDigest,
	}
	for _, candidate := range ir.Candidates {
		generated := generatedCandidate{
			Schema:         "gooo/improvement-selector/candidate-artifact/v1",
			CandidateID:    candidate.ID,
			Semantic:       candidate.Semantic,
			Operations:     []string{"DECLARE_ONLY", "NO_INPUT_REPOSITORY_WRITE", "NO_APPLY"},
			SourceDigest:   ir.SourceDigest,
			IRDigest:       irDigest,
			InputDigest:    inputDigest,
			ContractDigest: contractDigest,
		}
		data, err := marshalJSON(generated)
		if err != nil {
			return CandidateManifest{}, err
		}
		name := "candidate-" + candidate.ID + ".json"
		path := filepath.Join(outDir, name)
		if err := WriteBytes(path, data); err != nil {
			return CandidateManifest{}, err
		}
		artifactDigest, artifactData, err := DigestFile(path)
		if err != nil {
			return CandidateManifest{}, err
		}
		manifest.Candidates = append(manifest.Candidates, CandidateArtifact{
			Schema:         "gooo/improvement-selector/candidate-artifact/v1",
			CandidateID:    candidate.ID,
			SemanticDigest: DigestBytes([]byte(candidate.Semantic)),
			SourceDigest:   ir.SourceDigest,
			IRDigest:       irDigest,
			InputDigest:    inputDigest,
			ContractDigest: contractDigest,
			ArtifactPath:   name,
			ArtifactDigest: artifactDigest,
			GeneratedBytes: len(artifactData),
		})
	}
	return manifest, nil
}

func MeasureScenario(scenario Scenario, manifest CandidateManifest, inputDigest, contractDigest string) (MeasurementReceipt, error) {
	if manifest.Schema != "gooo/improvement-selector/candidate-manifest/v1" {
		return MeasurementReceipt{}, fmt.Errorf("unexpected candidate manifest schema %q", manifest.Schema)
	}
	if scenario.ScenarioID != manifest.ScenarioID {
		return MeasurementReceipt{}, fmt.Errorf("scenario and candidate manifest differ")
	}
	if manifest.InputDigest != inputDigest || manifest.ContractDigest != contractDigest {
		return MeasurementReceipt{}, fmt.Errorf("candidate manifest digest differs from measurement input")
	}
	byID := make(map[string]ScenarioObservation, len(scenario.Observations))
	for _, observation := range scenario.Observations {
		if observation.CandidateID == "" {
			return MeasurementReceipt{}, fmt.Errorf("scenario contains an empty candidate id")
		}
		if _, exists := byID[observation.CandidateID]; exists {
			return MeasurementReceipt{}, fmt.Errorf("scenario repeats candidate %q", observation.CandidateID)
		}
		byID[observation.CandidateID] = observation
	}
	receipt := MeasurementReceipt{
		Schema:           "gooo/improvement-selector/measurement-receipt/v1",
		ScenarioID:       scenario.ScenarioID,
		Class:            scenario.Class,
		Toolchain:        scenario.Toolchain,
		InputDigest:      inputDigest,
		ContractDigest:   contractDigest,
		PromotionAllowed: scenario.PromotionAllowed,
	}
	for _, candidate := range manifest.Candidates {
		observation, ok := byID[candidate.CandidateID]
		if !ok {
			return MeasurementReceipt{}, fmt.Errorf("scenario does not observe candidate %q", candidate.CandidateID)
		}
		entry := ReceiptEntry{
			CandidateID:             candidate.CandidateID,
			AuthorityAllowed:        observation.AuthorityAllowed != nil && *observation.AuthorityAllowed,
			SemanticConformance:     observation.SemanticConformance != nil && *observation.SemanticConformance,
			CounterexamplePreserved: observation.CounterexamplePreserved != nil && *observation.CounterexamplePreserved,
			Malformed:               observation.Malformed || observation.AuthorityAllowed == nil || observation.SemanticConformance == nil || observation.CounterexamplePreserved == nil,
		}
		if observation.Pair == nil {
			entry.PairIssue = "EXACT_BEFORE_AFTER_PAIR_MISSING"
		} else if !keyMatches(observation.Pair.Key.ScenarioID, scenario.ScenarioID) ||
			!keyMatches(observation.Pair.Key.Toolchain, scenario.Toolchain) ||
			!keyMatches(observation.Pair.Key.InputDigest, inputDigest) ||
			!keyMatches(observation.Pair.Key.ContractDigest, contractDigest) {
			entry.PairIssue = "EXACT_PAIR_KEY_MISMATCH"
		} else if err := ValidateVector(observation.Pair.Before); err != nil {
			entry.PairIssue = "MALFORMED_BEFORE_RESOURCE_VECTOR"
		} else if err := ValidateVector(observation.Pair.After); err != nil {
			entry.PairIssue = "MALFORMED_AFTER_RESOURCE_VECTOR"
		} else {
			entry.PairPresent = true
			pair := *observation.Pair
			pair.Key = PairKey{ScenarioID: scenario.ScenarioID, Toolchain: scenario.Toolchain, InputDigest: inputDigest, ContractDigest: contractDigest}
			entry.Pair = &pair
			delta := Delta(pair.Before, pair.After)
			entry.Delta = &delta
		}
		receipt.Entries = append(receipt.Entries, entry)
	}
	return receipt, nil
}

func keyMatches(observed, expected string) bool {
	return observed == expected || observed == "CURRENT"
}

func SelectImprovement(contract Contract, manifest CandidateManifest, receipt MeasurementReceipt) (Selection, error) {
	if err := ValidateContract(contract); err != nil {
		return Selection{}, err
	}
	if receipt.Schema != "gooo/improvement-selector/measurement-receipt/v1" || receipt.ScenarioID != manifest.ScenarioID {
		return Selection{}, fmt.Errorf("receipt identity is invalid")
	}
	if receipt.InputDigest != manifest.InputDigest || receipt.ContractDigest != manifest.ContractDigest {
		return Selection{}, fmt.Errorf("receipt and candidate manifest digests differ")
	}
	if len(receipt.Entries) != len(manifest.Candidates) {
		return Selection{}, fmt.Errorf("receipt candidate cardinality differs")
	}
	selection := Selection{
		Schema:        "gooo/improvement-selector/selection/v1",
		ScenarioID:    receipt.ScenarioID,
		Class:         receipt.Class,
		State:         StateClosed,
		Decision:      DecisionDefer,
		Precedence:    append([]string(nil), Precedence...),
		SelectorOrder: append([]string(nil), SelectorOrder...),
		Unweighted:    true,
	}
	for _, entry := range receipt.Entries {
		decision := CandidateDecision{
			CandidateID:             entry.CandidateID,
			State:                   StateClosed,
			Decision:                DecisionDefer,
			Reason:                  "EXACT_EVIDENCE_NOT_SELECTED",
			AuthorityAllowed:        entry.AuthorityAllowed,
			SemanticConformance:     entry.SemanticConformance,
			CounterexamplePreserved: entry.CounterexamplePreserved,
			Delta:                   entry.Delta,
		}
		switch {
		case entry.Malformed:
			decision.State = StateRefuted
			decision.Decision = DecisionRefute
			decision.Reason = "MALFORMED_GUARDRAIL_DECLARATION"
		case !entry.AuthorityAllowed:
			decision.State = StateRefuted
			decision.Decision = DecisionRefute
			decision.Reason = "AUTHORITY_GUARDRAIL_FAILED"
		case !entry.SemanticConformance:
			decision.State = StateRefuted
			decision.Decision = DecisionRefute
			decision.Reason = "SEMANTIC_CONFORMANCE_FAILED"
		case !entry.CounterexamplePreserved:
			decision.State = StateRefuted
			decision.Decision = DecisionRefute
			decision.Reason = "COUNTEREXAMPLE_NOT_PRESERVED"
		case !entry.PairPresent:
			decision.State = StateUnknown
			decision.Decision = DecisionIncomparableUnknown
			decision.Reason = entry.PairIssue
			unknownClass := "DIRECT_MISSING"
			if entry.PairIssue == "EXACT_PAIR_KEY_MISMATCH" {
				unknownClass = "DIGEST_MISMATCH"
			}
			if entry.PairIssue == "MALFORMED_BEFORE_RESOURCE_VECTOR" || entry.PairIssue == "MALFORMED_AFTER_RESOURCE_VECTOR" {
				unknownClass = "MALFORMED_EVIDENCE"
			}
			decision.Unknown = &UnknownDetail{
				Stage:         "MEASUREMENT_RECEIPT",
				Step:          "REQUIRE_EXACT_BEFORE_AFTER_PAIR",
				Reason:        entry.PairIssue,
				UnknownClass:  unknownClass,
				NextOperation: "PROVIDE_EXACT_BEFORE_AFTER_PAIR",
				BlockedBy:     []string{entry.CandidateID},
			}
		}
		selection.Candidates = append(selection.Candidates, decision)
	}
	for _, candidate := range selection.Candidates {
		if candidate.Unknown != nil {
			selection.Unknown = append(selection.Unknown, *candidate.Unknown)
		}
	}
	if hasState(selection.Candidates, StateRefuted) {
		selection.State = StateRefuted
		selection.Decision = DecisionRefute
		return selection, validateSelection(selection)
	}
	if len(selection.Unknown) > 0 {
		selection.State = StateUnknown
		selection.Decision = DecisionIncomparableUnknown
		return selection, validateSelection(selection)
	}
	if !receipt.PromotionAllowed {
		selection.State = StateClosed
		selection.Decision = DecisionDefer
		for i := range selection.Candidates {
			selection.Candidates[i].Reason = "PROMOTION_DEFERRED_BY_CONTRACT"
		}
		return selection, validateSelection(selection)
	}
	eligible := make([]int, 0, len(selection.Candidates))
	for i, candidate := range selection.Candidates {
		if candidate.State == StateClosed && candidate.Delta != nil {
			eligible = append(eligible, i)
		}
	}
	if len(eligible) == 0 {
		selection.State = StateUnknown
		selection.Decision = DecisionIncomparableUnknown
		selection.Unknown = append(selection.Unknown, UnknownDetail{
			Stage:         "SELECTOR",
			Step:          "REQUIRE_ELIGIBLE_CANDIDATE",
			Reason:        "NO_ELIGIBLE_CANDIDATE",
			UnknownClass:  "DEPENDENCY_BLOCKED",
			NextOperation: "PROVIDE_CONFORMANT_CANDIDATE",
			BlockedBy:     []string{"candidate-set"},
		})
		return selection, validateSelection(selection)
	}
	for _, left := range eligible {
		for _, right := range eligible {
			if left == right {
				continue
			}
			if CrossesAxes(*selection.Candidates[left].Delta, *selection.Candidates[right].Delta) {
				selection.State = StateUnknown
				selection.Decision = DecisionIncomparableUnknown
				selection.Unknown = append(selection.Unknown, UnknownDetail{
					Stage:         "SELECTOR",
					Step:          "COMPARE_EXACT_RESOURCE_DELTAS",
					Reason:        "RESOURCE_AXES_CROSS",
					UnknownClass:  "INCOMPARABLE",
					NextOperation: "PROVIDE_NON_CROSSING_EXACT_RESOURCE_PAIRS",
					BlockedBy:     []string{selection.Candidates[left].CandidateID, selection.Candidates[right].CandidateID},
				})
				return selection, validateSelection(selection)
			}
		}
	}
	best := eligible[0]
	unique := true
	for _, index := range eligible[1:] {
		comparison := LexicographicCompare(*selection.Candidates[index].Delta, *selection.Candidates[best].Delta)
		if comparison < 0 {
			best = index
			unique = true
		} else if comparison == 0 {
			unique = false
		}
	}
	if !unique {
		selection.State = StateClosed
		selection.Decision = DecisionDefer
		for i := range selection.Candidates {
			selection.Candidates[i].Reason = "EXACT_RESOURCE_DELTA_TIE"
		}
		return selection, validateSelection(selection)
	}
	selection.State = StateClosed
	selection.Decision = DecisionPromote
	selection.Winner = selection.Candidates[best].CandidateID
	for i := range selection.Candidates {
		if i == best {
			selection.Candidates[i].Decision = DecisionPromote
			selection.Candidates[i].Reason = "LEXICOGRAPHIC_RESOURCE_DELTA_MINIMUM"
		} else {
			selection.Candidates[i].Decision = DecisionDefer
			selection.Candidates[i].Reason = "LEXICOGRAPHIC_RESOURCE_DELTA_NOT_MINIMUM"
		}
	}
	return selection, validateSelection(selection)
}

func hasState(candidates []CandidateDecision, state string) bool {
	for _, candidate := range candidates {
		if candidate.State == state {
			return true
		}
	}
	return false
}

func validateSelection(selection Selection) error {
	if selection.State == StateUnknown {
		if len(selection.Unknown) == 0 {
			return fmt.Errorf("unknown selection has no unknown detail")
		}
		for i := range selection.Unknown {
			if err := ValidateUnknown(&selection.Unknown[i]); err != nil {
				return err
			}
		}
	}
	for _, candidate := range selection.Candidates {
		if candidate.State == StateUnknown {
			if err := ValidateUnknown(candidate.Unknown); err != nil {
				return err
			}
		}
	}
	return nil
}
