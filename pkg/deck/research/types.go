package research

import (
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	MethodBaseline   = "baseline"
	MethodGenetic    = "genetic"
	MethodConstraint = "constraint"
	MethodRoleFirst  = "role-first"
)

// Method builds one deck from a candidate pool.
type Method interface {
	Name() string
	Build(cards []deck.CardCandidate, cfg MethodConfig) (DeckResult, error)
}

// MethodConfig controls per-method behavior.
type MethodConfig struct {
	Seed    int64
	TopN    int
	DataDir string
}

// BenchmarkConfig controls benchmark execution.
type BenchmarkConfig struct {
	Tags      []string
	Seed      int64
	TopN      int
	Methods   []string
	OutputDir string
	DataDir   string
}

// DeckMetrics is the archetype-free score breakdown.
type DeckMetrics struct {
	Composite            float64  `json:"composite"`
	Synergy              float64  `json:"synergy"`
	Coverage             float64  `json:"coverage"`
	RoleFit              float64  `json:"role_fit"`
	ElixirFit            float64  `json:"elixir_fit"`
	CardQuality          float64  `json:"card_quality"`
	ConstraintViolations []string `json:"constraint_violations"`
	RuntimeMs            int64    `json:"runtime_ms"`
}

// DeckResult is one method's output for one player.
type DeckResult struct {
	Method              string      `json:"method"`
	Deck                []string    `json:"deck"`
	Metrics             DeckMetrics `json:"metrics"`
	DetectedArchetype   string      `json:"detected_archetype"`
	ArchetypeConfidence float64     `json:"archetype_confidence"`
}

// PlayerResult captures all method runs for one player.
type PlayerResult struct {
	PlayerTag   string       `json:"player_tag"`
	PlayerName  string       `json:"player_name"`
	MethodRuns  []DeckResult `json:"method_runs"`
	Winner      string       `json:"winner"`
	WinnerScore float64      `json:"winner_score"`
}

// BenchmarkAggregate summarizes performance by method.
type BenchmarkAggregate struct {
	Method                   string  `json:"method"`
	Runs                     int     `json:"runs"`
	MeanComposite            float64 `json:"mean_composite"`
	MedianComposite          float64 `json:"median_composite"`
	MeanRuntimeMs            float64 `json:"mean_runtime_ms"`
	ConstraintViolationCount int     `json:"constraint_violation_count"`
}

// BenchmarkReport is the exported benchmark artifact.
type BenchmarkReport struct {
	Timestamp       time.Time            `json:"timestamp"`
	Seed            int64                `json:"seed"`
	Methods         []string             `json:"methods"`
	PlayerRuns      []PlayerResult       `json:"player_runs"`
	Aggregates      []BenchmarkAggregate `json:"aggregates"`
	Recommendations []string             `json:"recommendations"`
}
