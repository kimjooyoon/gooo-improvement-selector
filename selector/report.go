package selector

import (
	"fmt"
	"strings"
)

func BuildReport(ir SemanticIR, contract Contract, irDigest, contractDigest string, selections []Selection, runtime RuntimeMetrics) (Report, error) {
	if ir.Schema != "gooo/improvement-selector/semantic-ir/v1" || len(ir.Nodes) != 12 || len(ir.Candidates) != 4 {
		return Report{}, fmt.Errorf("semantic IR cardinality is not fixed")
	}
	if err := ValidateContract(contract); err != nil {
		return Report{}, err
	}
	if len(selections) != 9 {
		return Report{}, fmt.Errorf("scenario denominator must contain exactly 9 scenarios")
	}
	counts := map[string]int{"normal": 0, "unknown": 0, "refuted": 0}
	seen := map[string]bool{}
	for _, selection := range selections {
		if seen[selection.ScenarioID] {
			return Report{}, fmt.Errorf("scenario %q is repeated", selection.ScenarioID)
		}
		seen[selection.ScenarioID] = true
		if _, ok := counts[selection.Class]; !ok {
			return Report{}, fmt.Errorf("unexpected scenario class %q", selection.Class)
		}
		counts[selection.Class]++
		if len(selection.SelectorOrder) != len(SelectorOrder) || len(selection.Precedence) != len(Precedence) || !selection.Unweighted {
			return Report{}, fmt.Errorf("selection contract is incomplete for %q", selection.ScenarioID)
		}
		for _, candidate := range selection.Candidates {
			if candidate.State == StateUnknown {
				if err := ValidateUnknown(candidate.Unknown); err != nil {
					return Report{}, fmt.Errorf("candidate %q: %w", candidate.CandidateID, err)
				}
			}
		}
		for i := range selection.Unknown {
			if err := ValidateUnknown(&selection.Unknown[i]); err != nil {
				return Report{}, fmt.Errorf("selection %q: %w", selection.ScenarioID, err)
			}
		}
	}
	if counts["normal"] != 2 || counts["unknown"] != 3 || counts["refuted"] != 4 {
		return Report{}, fmt.Errorf("scenario denominator must be normal=2 unknown=3 refuted=4")
	}
	for _, selection := range selections {
		switch selection.Class {
		case "normal":
			if selection.State != StateClosed || (selection.Decision != DecisionPromote && selection.Decision != DecisionDefer) {
				return Report{}, fmt.Errorf("normal scenario %q is not CLOSED", selection.ScenarioID)
			}
		case "unknown":
			if selection.State != StateUnknown || selection.Decision != DecisionIncomparableUnknown {
				return Report{}, fmt.Errorf("unknown scenario %q is not INCOMPARABLE_UNKNOWN", selection.ScenarioID)
			}
		case "refuted":
			if selection.State != StateRefuted || selection.Decision != DecisionRefute {
				return Report{}, fmt.Errorf("refuted scenario %q is not REFUTE", selection.ScenarioID)
			}
		}
	}
	if runtime.RepositoryWrites != 0 || runtime.LocalTestExecutions != 0 || runtime.CrossProjectRequiredGates != 0 {
		return Report{}, fmt.Errorf("runtime authority boundary is open")
	}
	if runtime.TestExecuted < 1 || runtime.TestReused < 0 || runtime.TestSkipped < 0 {
		return Report{}, fmt.Errorf("runtime test counts are invalid")
	}
	result := Report{
		Schema:         "gooo/improvement-selector/human-report/v1",
		Decision:       "CONFORMANCE_CLOSED",
		SourceDigest:   ir.SourceDigest,
		IRDigest:       irDigest,
		ContractDigest: contractDigest,
		Bindings: []ArtifactBinding{
			{Stage: "GOOO_SOURCE", Artifact: "examples/improvement-selector.gooo", ProducedBy: "source-authority", Digest: ir.SourceDigest},
			{Stage: "SEMANTIC_IR", Artifact: "semantic-ir.json", ProducedBy: "selector.Compile", Digest: irDigest},
			{Stage: "CANDIDATE_ARTIFACT", Artifact: "candidates/*/candidate-*.json", ProducedBy: "selector.GenerateCandidates", Digest: "manifest-bound"},
			{Stage: "MEASUREMENT_RECEIPT", Artifact: "candidates/*/receipt.json", ProducedBy: "selector.MeasureScenario", Digest: "pair-key-bound"},
			{Stage: "LEXICOGRAPHIC_SELECTOR", Artifact: "candidates/*/selection.json", ProducedBy: "selector.SelectImprovement", Digest: "decision-bound"},
			{Stage: "HUMAN_REPORT", Artifact: "human-report.md", ProducedBy: "selector.RenderMarkdown", Digest: "report-bound"},
		},
		ScenarioCounts: counts,
		Selections:     selections,
		Runtime:        runtime,
		Precedence:     append([]string(nil), Precedence...),
		SelectorOrder:  append([]string(nil), SelectorOrder...),
		Unweighted:     true,
	}
	result.FixedCells.Numerator = 12
	result.FixedCells.Denominator = 12
	return result, nil
}

func RenderMarkdown(report Report) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Gooo Improvement Selector CI Report")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "decision: `%s`\n", report.Decision)
	fmt.Fprintf(&b, "precedence: `%s`\n", strings.Join(report.Precedence, " > "))
	fmt.Fprintf(&b, "selector_order: `%s`\n", strings.Join(report.SelectorOrder, " -> "))
	fmt.Fprintln(&b, "unweighted: `true`")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Fixed denominator")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "cells: `%d/%d`\n", report.FixedCells.Numerator, report.FixedCells.Denominator)
	fmt.Fprintf(&b, "source_digest: `%s`\n", report.SourceDigest)
	fmt.Fprintf(&b, "semantic_ir_digest: `%s`\n", report.IRDigest)
	fmt.Fprintf(&b, "contract_digest: `%s`\n", report.ContractDigest)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "pipeline: `GOOO_SOURCE -> SEMANTIC_IR -> CANDIDATE_ARTIFACT -> MEASUREMENT_RECEIPT -> LEXICOGRAPHIC_SELECTOR -> HUMAN_REPORT`")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Scenario matrix")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| scenario | class | state | decision | winner | candidate decisions |")
	fmt.Fprintln(&b, "|---|---|---|---|---|---|")
	for _, selection := range report.Selections {
		decisions := make([]string, 0, len(selection.Candidates))
		for _, candidate := range selection.Candidates {
			decisions = append(decisions, candidate.CandidateID+":"+candidate.Decision)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", selection.ScenarioID, selection.Class, selection.State, selection.Decision, selection.Winner, strings.Join(decisions, ","))
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Exact integer runtime observations")
	fmt.Fprintln(&b)
	r := report.Runtime
	fmt.Fprintf(&b, "| field | value |\n|---|---:|\n")
	for _, item := range []struct{name string; value int}{
		{"build_wall_ms", r.BuildWallMs},
		{"test_wall_ms", r.TestWallMs},
		{"conformance_wall_ms", r.ConformanceWallMs},
		{"build_peak_rss_kib", r.BuildPeakRSSKib},
		{"test_peak_rss_kib", r.TestPeakRSSKib},
		{"conformance_peak_rss_kib", r.ConformancePeakRSSKib},
		{"test_executed", r.TestExecuted},
		{"test_reused", r.TestReused},
		{"test_skipped", r.TestSkipped},
		{"generated_artifact_files", r.GeneratedArtifactFiles},
		{"generated_artifact_bytes", r.GeneratedArtifactBytes},
		{"descendant_directories", r.DescendantDirectories},
		{"regular_files_root_readme_excluded", r.RegularFiles},
		{"go_files", r.GoFiles},
		{"go_lines", r.GoLines},
		{"gooo_files", r.GoooFiles},
		{"gooo_lines", r.GoooLines},
		{"repository_writes", r.RepositoryWrites},
		{"local_test_executions", r.LocalTestExecutions},
		{"cross_project_required_gates", r.CrossProjectRequiredGates},
	} {
		fmt.Fprintf(&b, "| %s | %d |\n", item.name, item.value)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Unknown coordinates")
	fmt.Fprintln(&b)
	for _, selection := range report.Selections {
		for _, unknown := range selection.Unknown {
			fmt.Fprintf(&b, "- `%s`: stage=`%s`, step=`%s`, reason=`%s`, unknown_class=`%s`, next_operation=`%s`, blocked_by=`%s`\n", selection.ScenarioID, unknown.Stage, unknown.Step, unknown.Reason, unknown.UnknownClass, unknown.NextOperation, strings.Join(unknown.BlockedBy, ","))
		}
		for _, candidate := range selection.Candidates {
			if candidate.Unknown != nil {
				u := candidate.Unknown
				fmt.Fprintf(&b, "- `%s/%s`: stage=`%s`, step=`%s`, reason=`%s`, unknown_class=`%s`, next_operation=`%s`, blocked_by=`%s`\n", selection.ScenarioID, candidate.CandidateID, u.Stage, u.Step, u.Reason, u.UnknownClass, u.NextOperation, strings.Join(u.BlockedBy, ","))
			}
		}
	}
	return b.String()
}
