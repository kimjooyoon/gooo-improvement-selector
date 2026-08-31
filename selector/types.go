package selector

import "encoding/json"

const (
	StateClosed  = "CLOSED"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"

	DecisionPromote            = "PROMOTE"
	DecisionDefer              = "DEFER"
	DecisionRefute             = "REFUTE"
	DecisionIncomparableUnknown = "INCOMPARABLE_UNKNOWN"
)

var SelectorOrder = []string{
	"AUTHORITY_GUARDRAIL",
	"SEMANTIC_CONFORMANCE",
	"COUNTEREXAMPLE_PRESERVATION",
	"EXACT_PAIRED_RESOURCE_DELTAS",
}

var Precedence = []string{StateRefuted, StateUnknown, StateClosed}

type Cell struct {
	Ordinal         int    `json:"ordinal"`
	ID              string `json:"id"`
	Activity        string `json:"activity"`
	ProofChoice     string `json:"proof_choice"`
	IndicatorClass  string `json:"indicator_class"`
	MetricID        string `json:"metric_id"`
	MetricPath      string `json:"metric_path"`
	Artifact        string `json:"artifact"`
	Evaluator       string `json:"evaluator"`
}

type Contract struct {
	Schema             string   `json:"schema"`
	DenominatorID      string   `json:"denominator_id"`
	Total              int      `json:"total"`
	ProofChoices       []string `json:"proof_choices"`
	IndicatorClasses   []string `json:"indicator_classes"`
	Cells              []Cell   `json:"cells"`
	SelectorOrder      []string `json:"selector_order"`
	Precedence         []string `json:"precedence"`
	UnknownFields      []string `json:"unknown_fields"`
	NoScores           bool     `json:"no_scores"`
	NoPercentages      bool     `json:"no_percentages"`
	NoNaturalInference bool     `json:"no_natural_language_inference"`
}

type CandidateDecl struct {
	ID       string `json:"id"`
	Semantic string `json:"semantic"`
}

type Program struct {
	Schema     string         `json:"schema"`
	Activities []string       `json:"activities"`
	Candidates []CandidateDecl `json:"candidates"`
}

type IRNode struct {
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	SourceLine     int    `json:"source_line"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	MetricID       string `json:"metric_id"`
	Artifact       string `json:"artifact"`
	Evaluator      string `json:"evaluator"`
}

type SemanticIR struct {
	Schema         string            `json:"schema"`
	SourceDigest   string            `json:"source_digest"`
	ContractDigest string            `json:"contract_digest"`
	Nodes          []IRNode          `json:"nodes"`
	Candidates     []CandidateDecl   `json:"candidates"`
}

type CandidateArtifact struct {
	Schema         string `json:"schema"`
	CandidateID    string `json:"candidate_id"`
	SemanticDigest string `json:"semantic_digest"`
	SourceDigest   string `json:"source_digest"`
	IRDigest       string `json:"ir_digest"`
	InputDigest    string `json:"input_digest"`
	ContractDigest string `json:"contract_digest"`
	ArtifactPath   string `json:"artifact_path"`
	ArtifactDigest string `json:"artifact_digest"`
	GeneratedBytes  int    `json:"generated_bytes"`
}

type CandidateManifest struct {
	Schema         string             `json:"schema"`
	ScenarioID     string             `json:"scenario_id"`
	SourceDigest   string             `json:"source_digest"`
	IRDigest       string             `json:"ir_digest"`
	InputDigest    string             `json:"input_digest"`
	ContractDigest string             `json:"contract_digest"`
	Candidates     []CandidateArtifact `json:"candidates"`
}

type ResourceVector struct {
	MemoryKib           int `json:"memory_kib"`
	BuildWallMs         int `json:"build_wall_ms"`
	TestWallMs          int `json:"test_wall_ms"`
	ConformanceWallMs   int `json:"conformance_wall_ms"`
}

type PairKey struct {
	ScenarioID   string `json:"scenario_id"`
	Toolchain    string `json:"toolchain"`
	InputDigest  string `json:"input_digest"`
	ContractDigest string `json:"contract_digest"`
}

type ResourcePair struct {
	Key    PairKey        `json:"key"`
	Before ResourceVector `json:"before"`
	After  ResourceVector `json:"after"`
}

type ScenarioObservation struct {
	CandidateID            string         `json:"candidate_id"`
	AuthorityAllowed       *bool          `json:"authority_allowed"`
	SemanticConformance    *bool          `json:"semantic_conformance"`
	CounterexamplePreserved *bool         `json:"counterexample_preserved"`
	Malformed              bool           `json:"malformed"`
	Pair                   *ResourcePair  `json:"pair"`
}

type Scenario struct {
	Schema          string                 `json:"schema"`
	ScenarioID      string                 `json:"scenario_id"`
	Class           string                 `json:"class"`
	Toolchain       string                 `json:"toolchain"`
	PromotionAllowed bool                  `json:"promotion_allowed"`
	Observations    []ScenarioObservation  `json:"observations"`
}

type ReceiptEntry struct {
	CandidateID             string          `json:"candidate_id"`
	AuthorityAllowed        bool            `json:"authority_allowed"`
	SemanticConformance     bool            `json:"semantic_conformance"`
	CounterexamplePreserved bool            `json:"counterexample_preserved"`
	Malformed               bool            `json:"malformed"`
	PairPresent             bool            `json:"pair_present"`
	PairIssue               string          `json:"pair_issue,omitempty"`
	Pair                    *ResourcePair   `json:"pair,omitempty"`
	Delta                   *ResourceVector `json:"delta,omitempty"`
}

type MeasurementReceipt struct {
	Schema         string         `json:"schema"`
	ScenarioID     string         `json:"scenario_id"`
	Class          string         `json:"class"`
	Toolchain      string         `json:"toolchain"`
	InputDigest    string         `json:"input_digest"`
	ContractDigest string         `json:"contract_digest"`
	PromotionAllowed bool          `json:"promotion_allowed"`
	Entries        []ReceiptEntry `json:"entries"`
}

type UnknownDetail struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CandidateDecision struct {
	CandidateID             string           `json:"candidate_id"`
	State                   string           `json:"state"`
	Decision                string           `json:"decision"`
	Reason                  string           `json:"reason"`
	AuthorityAllowed        bool             `json:"authority_allowed"`
	SemanticConformance     bool             `json:"semantic_conformance"`
	CounterexamplePreserved bool             `json:"counterexample_preserved"`
	Delta                   *ResourceVector  `json:"delta,omitempty"`
	Unknown                 *UnknownDetail   `json:"unknown,omitempty"`
}

type Selection struct {
	Schema         string              `json:"schema"`
	ScenarioID     string              `json:"scenario_id"`
	Class          string              `json:"class"`
	State          string              `json:"state"`
	Decision       string              `json:"decision"`
	Winner         string              `json:"winner,omitempty"`
	Precedence     []string            `json:"precedence"`
	SelectorOrder  []string            `json:"selector_order"`
	Unweighted     bool                `json:"unweighted"`
	Candidates     []CandidateDecision `json:"candidates"`
	Unknown        []UnknownDetail     `json:"unknown,omitempty"`
}

type RuntimeMetrics struct {
	Schema                    string `json:"schema"`
	BuildWallMs               int    `json:"build_wall_ms"`
	TestWallMs                int    `json:"test_wall_ms"`
	ConformanceWallMs         int    `json:"conformance_wall_ms"`
	BuildPeakRSSKib           int    `json:"build_peak_rss_kib"`
	TestPeakRSSKib            int    `json:"test_peak_rss_kib"`
	ConformancePeakRSSKib     int    `json:"conformance_peak_rss_kib"`
	TestExecuted              int    `json:"test_executed"`
	TestReused                int    `json:"test_reused"`
	TestSkipped               int    `json:"test_skipped"`
	GeneratedArtifactFiles    int    `json:"generated_artifact_files"`
	GeneratedArtifactBytes    int    `json:"generated_artifact_bytes"`
	DescendantDirectories     int    `json:"descendant_directories"`
	RegularFiles              int    `json:"regular_files"`
	GoFiles                   int    `json:"go_files"`
	GoLines                   int    `json:"go_lines"`
	GoooFiles                 int    `json:"gooo_files"`
	GoooLines                 int    `json:"gooo_lines"`
	RepositoryWrites          int    `json:"repository_writes"`
	LocalTestExecutions       int    `json:"local_test_executions"`
	CrossProjectRequiredGates int    `json:"cross_project_required_gates"`
}

type Report struct {
	Schema         string           `json:"schema"`
	Decision       string           `json:"decision"`
	SourceDigest   string           `json:"source_digest"`
	IRDigest       string           `json:"ir_digest"`
	ContractDigest string           `json:"contract_digest"`
	Bindings       []ArtifactBinding `json:"bindings"`
	FixedCells     struct {
		Numerator   int `json:"numerator"`
		Denominator int `json:"denominator"`
	} `json:"fixed_cells"`
	ScenarioCounts map[string]int `json:"scenario_counts"`
	Selections     []Selection     `json:"selections"`
	Runtime        RuntimeMetrics  `json:"runtime"`
	Precedence     []string        `json:"precedence"`
	SelectorOrder  []string        `json:"selector_order"`
	Unweighted     bool            `json:"unweighted"`
}

type ArtifactBinding struct {
	Stage     string `json:"stage"`
	Artifact  string `json:"artifact"`
	ProducedBy string `json:"produced_by"`
	Digest    string `json:"digest"`
}

func (r RuntimeMetrics) MarshalJSON() ([]byte, error) {
	type alias RuntimeMetrics
	return json.Marshal(alias(r))
}
