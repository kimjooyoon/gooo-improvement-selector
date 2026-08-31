package selector

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestFile(path string) (string, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	return DigestBytes(data), data, nil
}

func ReadJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')
	return WriteBytes(path, data)
}

func marshalJSON(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func WriteBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".gooo-selector-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func EnsureCallerOwnedOutput(path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return fmt.Errorf("candidate output must be an existing absolute temporary directory")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("candidate output directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("candidate output is not a directory")
	}
	resolvedOutput, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolve candidate output: %w", err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	workingDirectory, err = filepath.EvalSymlinks(workingDirectory)
	if err != nil {
		return err
	}
	if repositoryRoot := findGitRoot(workingDirectory); repositoryRoot != "" && pathWithin(repositoryRoot, resolvedOutput) {
		return fmt.Errorf("candidate output may not be inside the input repository")
	}
	return nil
}

func findGitRoot(start string) string {
	current := start
	for {
		if info, err := os.Stat(filepath.Join(current, ".git")); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func pathWithin(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && len(relative) > 3 && relative[:3] != ".."+string(filepath.Separator)
}

func ValidateContract(contract Contract) error {
	if contract.Schema != "gooo/improvement-selector/denominator/v1" {
		return fmt.Errorf("unexpected contract schema %q", contract.Schema)
	}
	if contract.Total != 12 || len(contract.Cells) != 12 {
		return fmt.Errorf("contract must contain exactly 12 cells")
	}
	if len(contract.ProofChoices) != 3 || len(contract.IndicatorClasses) != 3 {
		return fmt.Errorf("contract category cardinality is not fixed")
	}
	if len(contract.SelectorOrder) != len(SelectorOrder) || len(contract.Precedence) != len(Precedence) {
		return fmt.Errorf("contract order is incomplete")
	}
	for i, item := range SelectorOrder {
		if contract.SelectorOrder[i] != item {
			return fmt.Errorf("selector order mismatch at %d", i)
		}
	}
	for i, item := range Precedence {
		if contract.Precedence[i] != item {
			return fmt.Errorf("precedence mismatch at %d", i)
		}
	}
	for _, item := range []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"} {
		found := false
		for _, field := range contract.UnknownFields {
			if field == item {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown field %q is not contracted", item)
		}
	}
	if !contract.NoScores || !contract.NoPercentages || !contract.NoNaturalInference {
		return fmt.Errorf("forbidden inference outputs are not disabled")
	}
	seen := map[int]bool{}
	proofCounts := map[string]int{}
	indicatorCounts := map[string]int{}
	for _, cell := range contract.Cells {
		if cell.Ordinal < 1 || cell.Ordinal > 12 || seen[cell.Ordinal] {
			return fmt.Errorf("invalid cell ordinal %d", cell.Ordinal)
		}
		seen[cell.Ordinal] = true
		proofCounts[cell.ProofChoice]++
		indicatorCounts[cell.IndicatorClass]++
		if cell.ID == "" || cell.Activity == "" || cell.MetricID == "" || cell.MetricPath == "" || cell.Artifact == "" || cell.Evaluator == "" {
			return fmt.Errorf("cell %d is incomplete", cell.Ordinal)
		}
	}
	for _, choice := range contract.ProofChoices {
		if proofCounts[choice] != 4 {
			return fmt.Errorf("proof choice %q must occupy four cells", choice)
		}
	}
	for _, class := range contract.IndicatorClasses {
		if indicatorCounts[class] != 4 {
			return fmt.Errorf("indicator class %q must occupy four cells", class)
		}
	}
	return nil
}

func ValidateUnknown(u *UnknownDetail) error {
	if u == nil {
		return fmt.Errorf("unknown detail is absent")
	}
	if u.Stage == "" || u.Step == "" || u.Reason == "" || u.UnknownClass == "" || u.NextOperation == "" || len(u.BlockedBy) == 0 {
		return fmt.Errorf("unknown detail does not contain six required fields")
	}
	return nil
}

func ValidateVector(v ResourceVector) error {
	if v.MemoryKib < 0 || v.BuildWallMs < 0 || v.TestWallMs < 0 || v.ConformanceWallMs < 0 {
		return fmt.Errorf("resource vector contains a negative value")
	}
	return nil
}

func Delta(before, after ResourceVector) ResourceVector {
	return ResourceVector{
		MemoryKib:         after.MemoryKib - before.MemoryKib,
		BuildWallMs:       after.BuildWallMs - before.BuildWallMs,
		TestWallMs:        after.TestWallMs - before.TestWallMs,
		ConformanceWallMs: after.ConformanceWallMs - before.ConformanceWallMs,
	}
}

func vectorAxes(v ResourceVector) [4]int {
	return [4]int{v.MemoryKib, v.BuildWallMs, v.TestWallMs, v.ConformanceWallMs}
}

func LexicographicCompare(a, b ResourceVector) int {
	aa, bb := vectorAxes(a), vectorAxes(b)
	for i := range aa {
		if aa[i] < bb[i] {
			return -1
		}
		if aa[i] > bb[i] {
			return 1
		}
	}
	return 0
}

func CrossesAxes(a, b ResourceVector) bool {
	aa, bb := vectorAxes(a), vectorAxes(b)
	less, greater := false, false
	for i := range aa {
		less = less || aa[i] < bb[i]
		greater = greater || aa[i] > bb[i]
	}
	return less && greater
}
