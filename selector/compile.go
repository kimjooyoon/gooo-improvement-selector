package selector

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseProgram(source string) (Program, error) {
	program := Program{Schema: "gooo/improvement-selector/source/v1"}
	seenActivities := map[string]bool{}
	seenCandidates := map[string]bool{}
	for lineNumber, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "program":
			if len(fields) != 3 || fields[1] != "gooo-improvement-selector" || fields[2] != "v1" {
				return Program{}, fmt.Errorf("line %d: invalid program declaration", lineNumber+1)
			}
		case "activity":
			if len(fields) != 4 {
				return Program{}, fmt.Errorf("line %d: activity requires id, proof choice, and indicator class", lineNumber+1)
			}
			if seenActivities[fields[1]] {
				return Program{}, fmt.Errorf("line %d: duplicate activity %q", lineNumber+1, fields[1])
			}
			seenActivities[fields[1]] = true
			program.Activities = append(program.Activities, fields[1])
		case "candidate":
			if len(fields) != 3 {
				return Program{}, fmt.Errorf("line %d: candidate requires id and semantic token", lineNumber+1)
			}
			if seenCandidates[fields[1]] {
				return Program{}, fmt.Errorf("line %d: duplicate candidate %q", lineNumber+1, fields[1])
			}
			seenCandidates[fields[1]] = true
			program.Candidates = append(program.Candidates, CandidateDecl{ID: fields[1], Semantic: fields[2]})
		default:
			return Program{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber+1, fields[0])
		}
	}
	if len(program.Activities) != 12 {
		return Program{}, fmt.Errorf("source must declare exactly 12 activities, got %d", len(program.Activities))
	}
	if len(program.Candidates) != 4 {
		return Program{}, fmt.Errorf("source must declare exactly 4 candidates, got %d", len(program.Candidates))
	}
	return program, nil
}

func Compile(source string, contract Contract, contractDigest string) (SemanticIR, error) {
	if err := ValidateContract(contract); err != nil {
		return SemanticIR{}, err
	}
	program, err := ParseProgram(source)
	if err != nil {
		return SemanticIR{}, err
	}
	sourceDigest := DigestBytes([]byte(source))
	ir := SemanticIR{
		Schema:         "gooo/improvement-selector/semantic-ir/v1",
		SourceDigest:   sourceDigest,
		ContractDigest: contractDigest,
		Candidates:     program.Candidates,
	}
	for ordinal, cell := range contract.Cells {
		activity := cell.Activity
		if ordinal >= len(program.Activities) {
			return SemanticIR{}, fmt.Errorf("contract has more activities than source")
		}
		if activity != program.Activities[ordinal] {
			return SemanticIR{}, fmt.Errorf("activity %q is not bound to source activity %q", activity, program.Activities[ordinal])
		}
		ir.Nodes = append(ir.Nodes, IRNode{
			ID:             "ir/" + strconv.Itoa(cell.Ordinal) + "/" + cell.ID,
			Activity:       activity,
			SourceLine:     sourceLineForActivity(source, activity),
			ProofChoice:    cell.ProofChoice,
			IndicatorClass: cell.IndicatorClass,
			MetricID:       cell.MetricID,
			Artifact:       cell.Artifact,
			Evaluator:      cell.Evaluator,
		})
	}
	return ir, nil
}

func sourceLineForActivity(source, activity string) int {
	for lineNumber, raw := range strings.Split(source, "\n") {
		fields := strings.Fields(strings.TrimSpace(raw))
		if len(fields) >= 2 && fields[0] == "activity" && fields[1] == activity {
			return lineNumber + 1
		}
	}
	return 0
}
