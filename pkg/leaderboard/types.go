package leaderboard

import "time"

// DeckEntry represents a single deck entry in the leaderboard with all evaluation data
type DeckEntry struct {
	ID                int       `json:"id"`                 // Auto-incremented primary key
	DeckHash          string    `json:"deck_hash"`          // SHA256 hash of sorted card names for deduplication
	Cards             []string  `json:"cards"`              // 8 card names
	OverallScore      float64   `json:"overall_score"`      // 0-10 overall evaluation score
	AttackScore       float64   `json:"attack_score"`       // 0-10 attack category score
	DefenseScore      float64   `json:"defense_score"`      // 0-10 defense category score
	SynergyScore      float64   `json:"synergy_score"`      // 0-10 synergy category score
	VersatilityScore  float64   `json:"versatility_score"`  // 0-10 versatility category score
	F2PScore          float64   `json:"f2p_score"`          // 0-10 F2P friendly score
	PlayabilityScore  float64   `json:"playability_score"`  // 0-10 playability score
	Archetype         string    `json:"archetype"`          // Detected archetype (beatdown, control, etc.)
	ArchetypeConf     float64   `json:"archetype_conf"`     // Archetype confidence 0-1
	Strategy          string    `json:"strategy,omitempty"` // Strategy used for generation (optional)
	AvgElixir         float64   `json:"avg_elixir"`         // Average elixir cost
	EvaluatedAt       time.Time `json:"evaluated_at"`       // When this deck was evaluated
	PlayerTag         string    `json:"player_tag"`         // Player tag this deck is for
	EvaluationVersion string    `json:"evaluation_version"` // Version of evaluation algorithm used
}

// LeaderboardStats tracks metadata about the leaderboard and evaluation progress
type LeaderboardStats struct {
	PlayerTag           string    `json:"player_tag"`            // Player tag this leaderboard is for
	TotalDecksEvaluated int       `json:"total_decks_evaluated"` // Total number of decks evaluated
	TotalUniqueDecks    int       `json:"total_unique_decks"`    // Total unique decks (after deduplication)
	LastUpdated         time.Time `json:"last_updated"`          // Last time leaderboard was updated
	AvgEvalTimeMs       float64   `json:"avg_eval_time_ms"`      // Average evaluation time in milliseconds
	TopScore            float64   `json:"top_score"`             // Highest overall score in leaderboard
	AvgScore            float64   `json:"avg_score"`             // Average overall score across all decks
}

// QueryOptions defines filtering and sorting options for leaderboard queries
type QueryOptions struct {
	Limit           int      // Maximum number of results to return (0 = no limit)
	Offset          int      // Number of results to skip
	MinScore        float64  // Minimum overall score filter (0 = no filter)
	MaxScore        float64  // Maximum overall score filter (0 = no filter)
	Archetype       string   // Filter by archetype (empty = no filter)
	Strategy        string   // Filter by strategy (empty = no filter)
	MinAvgElixir    float64  // Minimum average elixir (0 = no filter)
	MaxAvgElixir    float64  // Maximum average elixir (0 = no filter)
	SortBy          string   // Sort field: "overall_score", "attack_score", etc. (default: overall_score)
	SortOrder       string   // Sort order: "asc" or "desc" (default: desc)
	RequireAllCards []string // Filter decks containing ALL of these cards (optional)
	RequireAnyCards []string // Filter decks containing ANY of these cards (optional)
	ExcludeCards    []string // Filter out decks containing ANY of these cards (optional)
}

// DefaultQueryOptions returns sensible defaults for leaderboard queries
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Limit:     10,
		Offset:    0,
		MinScore:  0,
		MaxScore:  0, // 0 means no max filter
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
}

// TopNQueryOptions creates options for retrieving top N decks by overall score
func TopNQueryOptions(n int) QueryOptions {
	return QueryOptions{
		Limit:     n,
		Offset:    0,
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
}

// ArchetypeQueryOptions creates options for retrieving decks of a specific archetype
func ArchetypeQueryOptions(archetype string, limit int) QueryOptions {
	return QueryOptions{
		Limit:     limit,
		Offset:    0,
		Archetype: archetype,
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
}
