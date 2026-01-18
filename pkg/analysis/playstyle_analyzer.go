package analysis

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

const balancedLabel = "Balanced"

// PlaystyleAnalysis represents the analysis of a player's playstyle
type PlaystyleAnalysis struct {
	PlayerTag    string    `json:"player_tag"`
	PlayerName   string    `json:"player_name"`
	AnalysisTime time.Time `json:"analysis_time"`

	// Battle Statistics
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	TotalBattles   int     `json:"total_battles"`
	WinRate        float64 `json:"win_rate"`
	ThreeCrownWins int     `json:"three_crown_wins"`
	ThreeCrownRate float64 `json:"three_crown_rate"`

	// Playstyle Characteristics
	AggressionLevel      string   `json:"aggression_level"`
	Consistency          string   `json:"consistency"`
	CurrentDeckAvgElixir float64  `json:"current_deck_avg_elixir"`
	CurrentWinCondition  string   `json:"current_win_condition"`
	DeckStyle            string   `json:"deck_style"`
	PlaystyleTraits      []string `json:"playstyle_traits"`

	// Current Deck Analysis
	CurrentDeckCards       []string `json:"current_deck_cards"`
	DeckElixirDistribution string   `json:"deck_elixir_distribution"`
}

// DeckRecommendation represents a deck recommendation with scoring
type DeckRecommendation struct {
	Deck          *DeckAnalysis `json:"deck"`
	Score         int           `json:"score"`
	Reasons       []string      `json:"reasons"`
	Compatibility string        `json:"compatibility"`
}

// DeckAnalysis represents a deck with its analysis
type DeckAnalysis struct {
	DeckName      string             `json:"deck_name"`
	WinCondition  string             `json:"win_condition"`
	AverageElixir float64            `json:"average_elixir"`
	Cards         []clashroyale.Card `json:"cards"`
	Strategy      string             `json:"strategy"`
	DeckDetail    []CardDeckDetail   `json:"deck_detail"`
}

// CardDeckDetail represents detailed card information in a deck
type CardDeckDetail struct {
	Name              string `json:"name"`
	Level             int    `json:"level"`
	MaxLevel          int    `json:"max_level"`
	Elixir            int    `json:"elixir"`
	EvolutionLevel    int    `json:"evolution_level,omitempty"`
	MaxEvolutionLevel int    `json:"max_evolution_level,omitempty"`
}

// AnalyzePlaystyle performs comprehensive playstyle analysis
func AnalyzePlaystyle(player *clashroyale.Player) (*PlaystyleAnalysis, error) {
	if player == nil {
		return nil, fmt.Errorf("player cannot be nil")
	}

	// Calculate basic statistics
	totalBattles := player.Wins + player.Losses
	winRate := 0.0
	if totalBattles > 0 {
		winRate = float64(player.Wins) / float64(totalBattles) * 100
	}

	threeCrownRate := 0.0
	if player.Wins > 0 {
		threeCrownRate = float64(player.ThreeCrownWins) / float64(player.Wins) * 100
	}

	// Analyze current deck
	currentDeckCards := make([]string, 0, len(player.CurrentDeck))
	for _, card := range player.CurrentDeck {
		currentDeckCards = append(currentDeckCards, card.Name)
	}

	deckAvgElixir := calculateDeckAverageElixir(player.CurrentDeck)
	winCondition := findWinCondition(player.CurrentDeck)
	deckStyle := determineDeckStyle(deckAvgElixir)

	// Determine playstyle characteristics
	aggressionLevel := determineAggressionLevel(threeCrownRate)
	consistency := determineConsistency(winRate)
	playstyleTraits := generatePlaystyleTraits(threeCrownRate, winRate, deckAvgElixir)

	// Create analysis
	analysis := &PlaystyleAnalysis{
		PlayerTag:              player.Tag,
		PlayerName:             player.Name,
		AnalysisTime:           time.Now(),
		Wins:                   player.Wins,
		Losses:                 player.Losses,
		TotalBattles:           totalBattles,
		WinRate:                roundToTwo(winRate),
		ThreeCrownWins:         player.ThreeCrownWins,
		ThreeCrownRate:         roundToTwo(threeCrownRate),
		AggressionLevel:        aggressionLevel,
		Consistency:            consistency,
		CurrentDeckAvgElixir:   roundToTwo(deckAvgElixir),
		CurrentWinCondition:    winCondition,
		DeckStyle:              deckStyle,
		PlaystyleTraits:        playstyleTraits,
		CurrentDeckCards:       currentDeckCards,
		DeckElixirDistribution: analyzeElixirDistribution(player.CurrentDeck),
	}

	return analysis, nil
}

// calculateDeckAverageElixir calculates the average elixir cost of a deck
func calculateDeckAverageElixir(deck []clashroyale.Card) float64 {
	if len(deck) == 0 {
		return 0
	}

	totalElixir := 0
	for _, card := range deck {
		totalElixir += card.ElixirCost
	}

	return float64(totalElixir) / float64(len(deck))
}

// findWinCondition identifies the primary win condition in a deck
func findWinCondition(deck []clashroyale.Card) string {
	if len(deck) == 0 {
		return ""
	}

	// Priority order for win conditions
	priorityOrder := []string{
		"Royal Giant", "Hog Rider", "Giant", "Battle Ram", "Goblin Barrel",
		"Miner", "Lava Hound", "Golem", "P.E.K.K.A", "Sparky", "Bowler",
		"Electro Giant", "Phoenix", "Monk",
	}

	for _, wc := range priorityOrder {
		for _, card := range deck {
			if card.Name == wc {
				return wc
			}
		}
	}

	// If no standard win condition found, check for any high elixir building/spell
	for _, card := range deck {
		if card.ElixirCost >= 5 && (strings.Contains(card.Type, "Building") ||
			strings.Contains(card.Type, "Spell")) {
			return card.Name
		}
	}

	return ""
}

// determineDeckStyle categorizes deck style based on average elixir
func determineDeckStyle(avgElixir float64) string {
	if avgElixir < 3.0 {
		return "Ultra-fast cycle"
	} else if avgElixir < 3.5 {
		return "Fast cycle"
	} else if avgElixir < 4.0 {
		return balancedLabel
	} else {
		return "Beatdown/Heavy"
	}
}

// determineAggressionLevel determines aggression based on three-crown rate
func determineAggressionLevel(threeCrownRate float64) string {
	if threeCrownRate > 75 {
		return "VERY AGGRESSIVE"
	} else if threeCrownRate > 60 {
		return "Aggressive"
	} else {
		return "Defensive/Reactive"
	}
}

// determineConsistency determines consistency based on win rate
func determineConsistency(winRate float64) string {
	if winRate > 55 {
		return "High"
	} else if winRate > 48 {
		return balancedLabel
	} else {
		return "Learning"
	}
}

// generatePlaystyleTraits creates a list of playstyle traits
func generatePlaystyleTraits(threeCrownRate, winRate, avgElixir float64) []string {
	traits := make([]string, 0)

	// Aggression traits
	if threeCrownRate > 75 {
		traits = append(traits, "Goes for tower damage aggressively")
		traits = append(traits, "Prefers offensive pressure over defensive play")
	} else if threeCrownRate > 60 {
		traits = append(traits, "Balanced offense with strong finishing")
	} else {
		traits = append(traits, "Prefers defensive counterplay")
	}

	// Consistency traits
	if winRate > 55 {
		traits = append(traits, "Consistent execution and matchup knowledge")
	} else if winRate > 48 {
		traits = append(traits, "Adapting and learning matchups")
	} else {
		traits = append(traits, "Building skills and adapting strategy")
	}

	// Tempo traits
	if avgElixir > 0 {
		if avgElixir < 3.0 {
			traits = append(traits, "Prefers constant pressure with fast cycle")
		} else if avgElixir < 3.5 {
			traits = append(traits, "Comfortable with aggressive tempo")
		}
	}

	return traits
}

// analyzeElixirDistribution provides a description of elixir distribution
func analyzeElixirDistribution(deck []clashroyale.Card) string {
	if len(deck) == 0 {
		return "No cards"
	}

	lowElixir := 0  // 1-2 elixir
	medElixir := 0  // 3-4 elixir
	highElixir := 0 // 5+ elixir

	for _, card := range deck {
		switch {
		case card.ElixirCost <= 2:
			lowElixir++
		case card.ElixirCost <= 4:
			medElixir++
		default:
			highElixir++
		}
	}

	return fmt.Sprintf("Low: %d, Med: %d, High: %d", lowElixir, medElixir, highElixir)
}

// roundToTwo rounds a float64 to two decimal places
func roundToTwo(num float64) float64 {
	return math.Round(num*100) / 100
}
