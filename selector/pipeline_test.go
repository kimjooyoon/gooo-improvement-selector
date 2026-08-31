package selector

import "testing"

func TestLexicographicCompareUsesExactOrderedAxes(t *testing.T) {
	a := ResourceVector{MemoryKib: -10, BuildWallMs: -5, TestWallMs: -4, ConformanceWallMs: -3}
	b := ResourceVector{MemoryKib: -5, BuildWallMs: -3, TestWallMs: -2, ConformanceWallMs: -1}
	if got := LexicographicCompare(a, b); got >= 0 {
		t.Fatalf("LexicographicCompare(a,b) = %d, want negative", got)
	}
	if CrossesAxes(a, b) {
		t.Fatal("ordered improvements should not cross axes")
	}
}

func TestCrossingResourceAxesRemainIncomparable(t *testing.T) {
	a := ResourceVector{MemoryKib: -10, BuildWallMs: 5}
	b := ResourceVector{MemoryKib: 5, BuildWallMs: -10}
	if !CrossesAxes(a, b) {
		t.Fatal("crossing resource axes must be incomparable")
	}
}

func TestUnknownSelectionContainsSixCoordinates(t *testing.T) {
	contract := testContract()
	manifest := CandidateManifest{
		Schema: "gooo/improvement-selector/candidate-manifest/v1",
		ScenarioID: "missing",
		InputDigest: "sha256:input",
		ContractDigest: "sha256:contract",
		Candidates: []CandidateArtifact{{CandidateID: "alpha"}},
	}
	receipt := MeasurementReceipt{
		Schema: "gooo/improvement-selector/measurement-receipt/v1",
		ScenarioID: "missing",
		InputDigest: "sha256:input",
		ContractDigest: "sha256:contract",
		PromotionAllowed: true,
		Entries: []ReceiptEntry{{CandidateID: "alpha", AuthorityAllowed: true, SemanticConformance: true, CounterexamplePreserved: true, PairIssue: "EXACT_BEFORE_AFTER_PAIR_MISSING"}},
	}
	selection, err := SelectImprovement(contract, manifest, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if selection.State != StateUnknown || selection.Decision != DecisionIncomparableUnknown {
		t.Fatalf("selection = %s/%s, want UNKNOWN/INCOMPARABLE_UNKNOWN", selection.State, selection.Decision)
	}
	if len(selection.Candidates) != 1 || selection.Candidates[0].Unknown == nil {
		t.Fatal("candidate unknown detail is absent")
	}
	if err := ValidateUnknown(selection.Candidates[0].Unknown); err != nil {
		t.Fatal(err)
	}
}

func testContract() Contract {
	contract := Contract{
		Schema: "gooo/improvement-selector/denominator/v1",
		DenominatorID: "test",
		Total: 12,
		ProofChoices: []string{"FOUNDATION", "COHERENCE", "REGRESSION"},
		IndicatorClasses: []string{"DRIVER", "OUTCOME", "GUARDRAIL"},
		SelectorOrder: append([]string(nil), SelectorOrder...),
		Precedence: append([]string(nil), Precedence...),
		UnknownFields: []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"},
		NoScores: true,
		NoPercentages: true,
		NoNaturalInference: true,
	}
	for ordinal := 1; ordinal <= 12; ordinal++ {
		proof := "FOUNDATION"
		if ordinal > 4 {
			proof = "COHERENCE"
		}
		if ordinal > 8 {
			proof = "REGRESSION"
		}
		indicator := "DRIVER"
		if ordinal%3 == 2 {
			indicator = "OUTCOME"
		}
		if ordinal%3 == 0 {
			indicator = "GUARDRAIL"
		}
		contract.Cells = append(contract.Cells, Cell{
			Ordinal: ordinal,
			ID: "cell",
			Activity: "activity",
			ProofChoice: proof,
			IndicatorClass: indicator,
			MetricID: "metric",
			MetricPath: "path",
			Artifact: "artifact",
			Evaluator: "evaluator",
		})
	}
	return contract
}
