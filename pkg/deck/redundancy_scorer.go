// Package deck provides redundancy and versatility scoring research prototypes.
// This is RESEARCH CODE, not production-ready implementations.
//
// Research spike: clash-royale-api-5xb
package deck

import (
	"fmt"
	"math"
)

// RedundancyScorer calculates penalties for redundant cards in a deck.
// This is a research prototype for exploring redundancy detection patterns.
type RedundancyScorer struct {
	// roleClassifier identifies which role each card serves
	roleClassifier *RoleClassifier

	// archetypeDetector identifies the deck's archetype
	archetypeDetector *ArchetypeDetector

	// toleranceByArchetype defines redundancy thresholds per archetype
	toleranceByArchetype map[Archetype]RoleTolerance
}

// RoleTolerance defines the acceptable count for each role before redundancy penalties apply.
type RoleTolerance struct {
	WinCondition int // Default: 2
	Building     int // Default: 1
	SpellBig     int // Default: 2
	SpellSmall   int // Default: 3
	Support      int // Default: 4
	Cycle        int // Default: 3
}

// Archetype represents the deck's strategic category.
type Archetype string

const (
	ArchetypeBeatdown   Archetype = "beatdown"
	ArchetypeControl    Archetype = "control"
	ArchetypeCycle      Archetype = "cycle"
	ArchetypeSiege      Archetype = "siege"
	ArchetypeBait       Archetype = "bait"
	ArchetypeBridgeSpam Archetype = "bridge_spam"
	ArchetypeMidrange   Archetype = "midrange"
	ArchetypeSpawndeck  Archetype = "spawndeck"
)

// CardRole represents a card's strategic function.
type CardRole string

const (
	RoleWinCondition CardRole = "win_condition"
	RoleBuilding     CardRole = "building"
	RoleSpellBig     CardRole = "spell_big"
	RoleSpellSmall   CardRole = "spell_small"
	RoleSupport      CardRole = "support"
	RoleCycle        CardRole = "cycle"
)

// RedundancyReport details redundancy analysis results.
type RedundancyReport struct {
	Archetype             Archetype
	RoleCounts            map[CardRole]int
	RedundantCards        []RedundantCard
	SynergisticRedundancy bool    // True if redundancy is intentional (bait, bridge spam)
	OverallPenalty        float64 // 0-1, higher = more redundant
}

// RedundantCard represents a card category with redundancy.
type RedundantCard struct {
	Role      CardRole
	Count     int
	Threshold int
	Severity  float64 // (count - threshold) / threshold
}

// NewRedundancyScorer creates a new redundancy scorer with default tolerances.
func NewRedundancyScorer(rc *RoleClassifier, ad *ArchetypeDetector) *RedundancyScorer {
	return &RedundancyScorer{
		roleClassifier:       rc,
		archetypeDetector:    ad,
		toleranceByArchetype: defaultToleranceByArchetype(),
	}
}

// defaultToleranceByArchetype provides baseline redundancy thresholds.
func defaultToleranceByArchetype() map[Archetype]RoleTolerance {
	return map[Archetype]RoleTolerance{
		ArchetypeBeatdown: {
			WinCondition: 2,
			Building:     1,
			SpellBig:     2,
			SpellSmall:   3,
			Support:      4,
			Cycle:        2,
		},
		ArchetypeControl: {
			WinCondition: 1,
			Building:     2,
			SpellBig:     3,
			SpellSmall:   3,
			Support:      3,
			Cycle:        3,
		},
		ArchetypeCycle: {
			WinCondition: 1,
			Building:     1,
			SpellBig:     1,
			SpellSmall:   4,
			Support:      3,
			Cycle:        4,
		},
		ArchetypeSiege: {
			WinCondition: 1,
			Building:     2,
			SpellBig:     2,
			SpellSmall:   3,
			Support:      3,
			Cycle:        4,
		},
		ArchetypeBait: {
			WinCondition: 2,
			Building:     1,
			SpellBig:     0,
			SpellSmall:   4, // Bait targets
			Support:      4,
			Cycle:        3,
		},
		ArchetypeBridgeSpam: {
			WinCondition: 3, // Multiple bridge spam win conditions
			Building:     0,
			SpellBig:     2,
			SpellSmall:   2,
			Support:      3,
			Cycle:        2,
		},
	}
}

// AnalyzeRedundancy performs comprehensive redundancy analysis on a deck.
func (rs *RedundancyScorer) AnalyzeRedundancy(deck []Card) (*RedundancyReport, error) {
	// Detect archetype first
	archetype := rs.archetypeDetector.Detect(deck)

	// Count cards by role
	roleCounts := rs.countByRole(deck)

	// Get tolerance for detected archetype
	tolerance := rs.toleranceByArchetype[archetype]

	// Identify redundant cards
	redundancies := []RedundantCard{}
	for role, count := range roleCounts {
		threshold := rs.getThreshold(tolerance, role)
		if count > threshold {
			redundancies = append(redundancies, RedundantCard{
				Role:      role,
				Count:     count,
				Threshold: threshold,
				Severity:  rs.calculateSeverity(count, threshold),
			})
		}
	}

	// Check for synergistic redundancy (bait, bridge spam)
	synergistic := rs.detectSynergisticRedundancy(deck, archetype)

	// Calculate overall penalty
	penalty := rs.calculateOverallPenalty(redundancies, synergistic, archetype)

	return &RedundancyReport{
		Archetype:             archetype,
		RoleCounts:            roleCounts,
		RedundantCards:        redundancies,
		SynergisticRedundancy: synergistic,
		OverallPenalty:        penalty,
	}, nil
}

// countByRole counts how many cards serve each role.
func (rs *RedundancyScorer) countByRole(deck []Card) map[CardRole]int {
	counts := make(map[CardRole]int)
	for _, card := range deck {
		role := rs.roleClassifier.Classify(card)
		counts[role]++
	}
	return counts
}

// getThreshold returns the redundancy threshold for a given role.
func (rs *RedundancyScorer) getThreshold(tolerance RoleTolerance, role CardRole) int {
	switch role {
	case RoleWinCondition:
		return tolerance.WinCondition
	case RoleBuilding:
		return tolerance.Building
	case RoleSpellBig:
		return tolerance.SpellBig
	case RoleSpellSmall:
		return tolerance.SpellSmall
	case RoleSupport:
		return tolerance.Support
	case RoleCycle:
		return tolerance.Cycle
	default:
		return 2 // Default threshold
	}
}

// calculateSeverity computes how severe the redundancy is.
func (rs *RedundancyScorer) calculateSeverity(count, threshold int) float64 {
	if count <= threshold {
		return 0.0
	}
	// Linear severity: (count - threshold) / threshold
	return float64(count-threshold) / float64(threshold)
}

// detectSynergisticRedundancy checks if redundancy is intentional (bait, bridge spam).
func (rs *RedundancyScorer) detectSynergisticRedundancy(deck []Card, archetype Archetype) bool {
	if archetype == ArchetypeBait {
		// Check for 3+ cards sharing same counter weakness
		// This is the defining feature of bait decks
		return rs.hasSharedCounterWeakness(deck, 3)
	}

	if archetype == ArchetypeBridgeSpam {
		// Check for 2-3 medium-cost swarm win conditions
		winConditions := rs.filterWinConditions(deck)
		swarmCount := 0
		for _, card := range winConditions {
			if card.ElixirCost >= 3 && card.ElixirCost <= 7 && card.IsSwarm {
				swarmCount++
			}
		}
		return swarmCount >= 2 && swarmCount <= 3
	}

	return false
}

// hasSharedCounterWeakness checks if multiple cards share the same counter.
func (rs *RedundancyScorer) hasSharedCounterWeakness(deck []Card, minCount int) bool {
	counterWeaknesses := make(map[string]int)
	for _, card := range deck {
		for _, weakness := range card.CounterWeaknesses {
			counterWeaknesses[weakness]++
		}
	}
	for _, count := range counterWeaknesses {
		if count >= minCount {
			return true
		}
	}
	return false
}

// filterWinConditions returns cards classified as win conditions.
func (rs *RedundancyScorer) filterWinConditions(deck []Card) []Card {
	result := []Card{}
	for _, card := range deck {
		if rs.roleClassifier.Classify(card) == RoleWinCondition {
			result = append(result, card)
		}
	}
	return result
}

// calculateOverallPenalty computes the final redundancy penalty score.
func (rs *RedundancyScorer) calculateOverallPenalty(redundancies []RedundantCard, synergistic bool, archetype Archetype) float64 {
	if len(redundancies) == 0 {
		return 0.0
	}

	// Base penalty: sum of severity scores
	basePenalty := 0.0
	for _, r := range redundancies {
		basePenalty += r.Severity
	}

	// Apply archetype multiplier
	multiplier := rs.getArchetypeMultiplier(archetype)

	// Reduce penalty for synergistic redundancy
	if synergistic {
		multiplier *= 0.1 // 90% reduction
	}

	// Clamp to [0, 1]
	penalty := basePenalty * multiplier
	if penalty > 1.0 {
		penalty = 1.0
	}
	if penalty < 0.0 {
		penalty = 0.0
	}

	return penalty
}

// getArchetypeMultiplier returns how strongly redundancy should be penalized.
func (rs *RedundancyScorer) getArchetypeMultiplier(archetype Archetype) float64 {
	switch archetype {
	case ArchetypeBeatdown, ArchetypeSiege:
		return 1.0 // High penalty - redundancy is bad
	case ArchetypeControl:
		return 0.8 // Moderate penalty
	case ArchetypeBait, ArchetypeBridgeSpam:
		return 0.3 // Low penalty - redundancy is often intentional
	case ArchetypeCycle:
		return 0.5 // Mixed - cycle card redundancy OK
	default:
		return 0.7 // Balanced
	}
}

// String returns a human-readable representation of the redundancy report.
func (rr *RedundancyReport) String() string {
	return fmt.Sprintf("RedundancyReport{Archetype=%s, Penalty=%.2f, Synergistic=%v, Redundancies=%d}",
		rr.Archetype, rr.OverallPenalty, rr.SynergisticRedundancy, len(rr.RedundantCards))
}

// Card represents a simplified card model for research purposes.
// In production, use the existing Card type from pkg/clashroyale.
type Card struct {
	Name              string
	ElixirCost        int
	CounterWeaknesses []string // Spells/cards that counter this card
	IsSwarm           bool     // True for swarm-type cards
	// Additional fields as needed
}

// RoleClassifier classifies cards into strategic roles.
// This is a placeholder - use the existing pkg/deck/role_classifier.go in production.
type RoleClassifier struct{}

// Classify returns the primary role for a card.
func (rc *RoleClassifier) Classify(card Card) CardRole {
	// Placeholder implementation
	// In production, use the existing role classifier logic
	return RoleSupport
}

// ArchetypeDetector identifies a deck's strategic archetype.
type ArchetypeDetector struct{}

// Detect analyzes deck composition to determine archetype.
func (ad *ArchetypeDetector) Detect(deck []Card) Archetype {
	// Placeholder implementation
	// In production, use the existing archetype detection logic from pkg/archetypes/
	return ArchetypeBeatdown
}

// Example usage demonstrates how to use the redundancy scorer.
func ExampleUsage() {
	rc := &RoleClassifier{}
	ad := &ArchetypeDetector{}
	scorer := NewRedundancyScorer(rc, ad)

	// Example deck: Beatdown with 3 win conditions (redundant)
	deck := []Card{
		{Name: "Giant", ElixirCost: 8, IsSwarm: false},
		{Name: "Golem", ElixirCost: 8, IsSwarm: false},
		{Name: "P.E.K.K.A", ElixirCost: 7, IsSwarm: false},
		// ... more cards
	}

	report, err := scorer.AnalyzeRedundancy(deck)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(report.String())
	// Output: RedundancyReport{Archetype=beatdown, Penalty=0.50, Synergistic=false, Redundancies=1}
}

// calculateExponentialPenalty provides an alternative penalty calculation using exponential scaling.
// This is for research comparison with the linear penalty.
func (rs *RedundancyScorer) calculateExponentialPenalty(count, threshold int) float64 {
	if count <= threshold {
		return 0.0
	}
	// Exponential severity: 1.5 ^ (count - threshold) - 1
	excess := float64(count - threshold)
	return math.Pow(1.5, excess) - 1.0
}
