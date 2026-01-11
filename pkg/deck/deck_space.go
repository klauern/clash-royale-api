package deck

import (
	"fmt"
	"math"
	"math/big"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// DeckSpaceStats represents statistics about possible deck combinations
type DeckSpaceStats struct {
	TotalCards        int                 // Total cards available for deck building
	CardsByRole       map[CardRole]int    // Number of cards available per role
	TotalCombinations *big.Int            // Total possible combinations without constraints
	ValidCombinations *big.Int            // Valid combinations respecting role constraints
	ByElixirRange     map[string]*big.Int // Combinations by average elixir range
	ByArchetype       map[string]*big.Int // Estimated combinations by archetype
}

// ElixirRange represents an average elixir cost range
type ElixirRange struct {
	Min   float64
	Max   float64
	Label string
}

// Standard elixir ranges for deck categorization
var StandardElixirRanges = []ElixirRange{
	{Min: 0.0, Max: 2.5, Label: "Very Fast (0-2.5)"},
	{Min: 2.5, Max: 3.0, Label: "Fast (2.5-3.0)"},
	{Min: 3.0, Max: 3.5, Label: "Medium-Fast (3.0-3.5)"},
	{Min: 3.5, Max: 4.0, Label: "Medium (3.5-4.0)"},
	{Min: 4.0, Max: 4.5, Label: "Medium-Slow (4.0-4.5)"},
	{Min: 4.5, Max: 5.0, Label: "Slow (4.5-5.0)"},
	{Min: 5.0, Max: 10.0, Label: "Very Slow (5.0+)"},
}

// DeckSpaceCalculator calculates possible deck combinations from a card collection
type DeckSpaceCalculator struct {
	cards       []CardCandidate
	cardsByRole map[CardRole][]CardCandidate
}

// NewDeckSpaceCalculator creates a new calculator from a player's card collection
func NewDeckSpaceCalculator(player *clashroyale.Player) (*DeckSpaceCalculator, error) {
	if player == nil {
		return nil, fmt.Errorf("player cannot be nil")
	}

	calc := &DeckSpaceCalculator{
		cards:       make([]CardCandidate, 0, len(player.Cards)),
		cardsByRole: make(map[CardRole][]CardCandidate),
	}

	// Convert player cards to candidates and categorize by role
	for _, card := range player.Cards {
		role := config.GetCardRoleWithEvolution(card.Name, card.EvolutionLevel)

		candidate := CardCandidate{
			Name:              card.Name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            card.ElixirCost,
			Role:              &role,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}

		calc.cards = append(calc.cards, candidate)
		calc.cardsByRole[role] = append(calc.cardsByRole[role], candidate)
	}

	return calc, nil
}

// CalculateStats computes comprehensive statistics about possible deck combinations
func (dsc *DeckSpaceCalculator) CalculateStats() *DeckSpaceStats {
	stats := &DeckSpaceStats{
		TotalCards:    len(dsc.cards),
		CardsByRole:   make(map[CardRole]int),
		ByElixirRange: make(map[string]*big.Int),
		ByArchetype:   make(map[string]*big.Int),
	}

	// Count cards by role
	for role, cards := range dsc.cardsByRole {
		stats.CardsByRole[role] = len(cards)
	}

	// Calculate total combinations C(n, 8)
	stats.TotalCombinations = combinations(len(dsc.cards), 8)

	// Calculate valid combinations with role constraints
	stats.ValidCombinations = dsc.calculateConstrainedCombinations()

	// Calculate combinations by elixir range (estimation)
	stats.ByElixirRange = dsc.estimateCombinationsByElixir()

	// Estimate by archetype (rough approximation)
	stats.ByArchetype = dsc.estimateCombinationsByArchetype()

	return stats
}

// calculateConstrainedCombinations calculates valid combinations respecting role constraints
// Default composition: 1 win condition, 1 building, 1 big spell, 1 small spell, 2 support, 2 cycle
func (dsc *DeckSpaceCalculator) calculateConstrainedCombinations() *big.Int {
	// Define the default composition requirements (from builder.go:157-170, 496-501)
	composition := map[CardRole]int{
		RoleWinCondition: 1,
		RoleBuilding:     1,
		RoleSpellBig:     1,
		RoleSpellSmall:   1,
		RoleSupport:      2,
		RoleCycle:        2,
	}

	// Calculate combinations for each role requirement
	result := big.NewInt(1)

	for role, count := range composition {
		cardsInRole := len(dsc.cardsByRole[role])
		if cardsInRole < count {
			// Not enough cards of this role - no valid combinations
			return big.NewInt(0)
		}

		// Multiply by C(cardsInRole, count)
		roleCombos := combinations(cardsInRole, count)
		result.Mul(result, roleCombos)
	}

	return result
}

// estimateCombinationsByElixir estimates combinations that fall within each elixir range
// This is a rough approximation based on card distribution
func (dsc *DeckSpaceCalculator) estimateCombinationsByElixir() map[string]*big.Int {
	estimates := make(map[string]*big.Int)

	// For now, return a placeholder estimation
	// A more sophisticated approach would simulate many random decks
	// and calculate the distribution, but that's computationally expensive

	// Simple heuristic: distribute proportionally based on card elixir costs
	totalCombinations := dsc.calculateConstrainedCombinations()

	// Calculate average elixir of all cards
	totalElixir := 0
	for _, card := range dsc.cards {
		totalElixir += card.Elixir
	}
	avgElixir := float64(totalElixir) / float64(len(dsc.cards))

	// For each range, estimate based on proximity to collection average
	for _, elixirRange := range StandardElixirRanges {
		rangeMid := (elixirRange.Min + elixirRange.Max) / 2
		distance := math.Abs(avgElixir - rangeMid)

		// Simple bell curve approximation
		proportion := math.Exp(-distance / 0.5)

		estimate := new(big.Int).Set(totalCombinations)
		// Multiply by proportion (convert to integer percentage)
		percentage := int64(proportion * 100)
		estimate.Div(estimate, big.NewInt(100))
		estimate.Mul(estimate, big.NewInt(percentage))

		estimates[elixirRange.Label] = estimate
	}

	return estimates
}

// estimateCombinationsByArchetype provides rough estimates for archetype-based combinations
func (dsc *DeckSpaceCalculator) estimateCombinationsByArchetype() map[string]*big.Int {
	estimates := make(map[string]*big.Int)

	// Common archetypes based on win condition types
	archetypes := []string{
		"Beatdown",
		"Control",
		"Cycle",
		"Siege",
		"Bridge Spam",
		"Bait",
	}

	// Rough estimate: divide total by number of major archetypes
	totalValid := dsc.calculateConstrainedCombinations()
	perArchetype := new(big.Int).Div(totalValid, big.NewInt(int64(len(archetypes))))

	for _, archetype := range archetypes {
		estimates[archetype] = new(big.Int).Set(perArchetype)
	}

	return estimates
}

// FormatLargeNumber formats a large number with K/M/B/T suffixes
// Examples: 1234 → "1.2K", 1234567 → "1.2M", 1234567890 → "1.2B"
func FormatLargeNumber(n *big.Int) string {
	if n == nil {
		return "0"
	}

	// Convert to float for easier manipulation
	f := new(big.Float).SetInt(n)
	value, _ := f.Float64()

	// Handle numbers that are too large for float64
	if n.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		// Use string length to approximate magnitude
		str := n.String()
		digits := len(str)

		if digits >= 13 { // Trillions
			return fmt.Sprintf("%.1fT", value/1e12)
		} else if digits >= 10 { // Billions
			return fmt.Sprintf("%.1fB", value/1e9)
		} else if digits >= 7 { // Millions
			return fmt.Sprintf("%.1fM", value/1e6)
		}
	}

	// Standard formatting for smaller numbers
	switch {
	case value >= 1e12:
		return fmt.Sprintf("%.1fT", value/1e12)
	case value >= 1e9:
		return fmt.Sprintf("%.1fB", value/1e9)
	case value >= 1e6:
		return fmt.Sprintf("%.1fM", value/1e6)
	case value >= 1e3:
		return fmt.Sprintf("%.1fK", value/1e3)
	default:
		return n.String()
	}
}

// combinations calculates C(n, k) = n! / (k! * (n-k)!)
// Uses big.Int to handle very large numbers
func combinations(n, k int) *big.Int {
	if k > n || k < 0 {
		return big.NewInt(0)
	}

	if k == 0 || k == n {
		return big.NewInt(1)
	}

	// Optimize by using the smaller of k and n-k
	if k > n-k {
		k = n - k
	}

	result := big.NewInt(1)

	// Calculate C(n, k) = (n * (n-1) * ... * (n-k+1)) / (k * (k-1) * ... * 1)
	for i := 0; i < k; i++ {
		result.Mul(result, big.NewInt(int64(n-i)))
		result.Div(result, big.NewInt(int64(i+1)))
	}

	return result
}
