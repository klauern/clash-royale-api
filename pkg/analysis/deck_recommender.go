package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// DeckRecommendationResult contains recommendation results
type DeckRecommendationResult struct {
	Recommended    *DeckRecommendation `json:"recommended"`
	AllScores      []*DeckRecommendation `json:"all_scores"`
	AnalysisTime   string               `json:"analysis_time"`
}

// RecommendDecks analyzes playstyle and recommends the best deck
func RecommendDecks(playstyle *PlaystyleAnalysis, dataDir string) (*DeckRecommendationResult, error) {
	if playstyle == nil {
		return nil, fmt.Errorf("playstyle analysis cannot be nil")
	}

	// Load available decks from data directory
	decks, err := loadDecksFromDirectory(dataDir)
	if err != nil {
		// If no decks found, create some example decks
		decks = createExampleDecks()
	}

	// Score each deck based on playstyle
	deckScores := make([]*DeckRecommendation, 0, len(decks))

	for _, deck := range decks {
		score, reasons := scoreDeckForPlaystyle(deck, playstyle)

		recommendation := &DeckRecommendation{
			Deck:          deck,
			Score:         score,
			Reasons:       reasons,
			Compatibility: determineCompatibility(score),
		}
		deckScores = append(deckScores, recommendation)
	}

	// Sort by score (highest first)
	sort.Slice(deckScores, func(i, j int) bool {
		return deckScores[i].Score > deckScores[j].Score
	})

	result := &DeckRecommendationResult{
		Recommended:  deckScores[0],
		AllScores:    deckScores,
		AnalysisTime: playstyle.AnalysisTime.Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// loadDecksFromDirectory loads deck files from the data directory
func loadDecksFromDirectory(dataDir string) ([]*DeckAnalysis, error) {
	decksDir := filepath.Join(dataDir, "decks")
	if _, err := os.Stat(decksDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("decks directory not found")
	}

	files, err := filepath.Glob(filepath.Join(decksDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read deck files: %w", err)
	}

	decks := make([]*DeckAnalysis, 0, len(files))
	for _, file := range files {
		deck, err := loadDeckFromFile(file)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// loadDeckFromFile loads a single deck from a JSON file
func loadDeckFromFile(filename string) (*DeckAnalysis, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read deck file: %w", err)
	}

	var deck DeckAnalysis
	if err := json.Unmarshal(data, &deck); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deck data: %w", err)
	}

	return &deck, nil
}

// createExampleDecks creates some example deck recommendations
func createExampleDecks() []*DeckAnalysis {
	decks := make([]*DeckAnalysis, 0)

	// Hog Cycle Deck
	hogCycle := &DeckAnalysis{
		DeckName:     "Hog Cycle Fast Aggression",
		WinCondition: "Hog Rider",
		AverageElixir: 2.9,
		Strategy: "Cycle through cheap cards while applying constant pressure with Hog Rider. Use Ice Spirit and Skeletons for cheap cycle. Use Zap for support and defense.",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", ElixirCost: 4, Level: 11, MaxLevel: 14},
			{Name: "Ice Spirit", ElixirCost: 1, Level: 11, MaxLevel: 14},
			{Name: "Skeletons", ElixirCost: 1, Level: 11, MaxLevel: 14},
			{Name: "Zap", ElixirCost: 2, Level: 12, MaxLevel: 14},
			{Name: "The Log", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Tombstone", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Cannon", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Tesla", ElixirCost: 4, Level: 11, MaxLevel: 14},
		},
	}
	hogCycle.DeckDetail = convertCardsToDeckDetail(hogCycle.Cards)
	decks = append(decks, hogCycle)

	// Battle Ram Cycle Deck
	battleRamCycle := &DeckAnalysis{
		DeckName:     "Battle Ram Cycle Balanced Aggression",
		WinCondition: "Battle Ram",
		AverageElixir: 3.3,
		Strategy: "Use Battle Ram as primary win condition with support from Bandit and Inferno Dragon. Control the bridge area and apply pressure.",
		Cards: []clashroyale.Card{
			{Name: "Battle Ram", ElixirCost: 4, Level: 11, MaxLevel: 14},
			{Name: "Bandit", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Inferno Dragon", ElixirCost: 4, Level: 11, MaxLevel: 14},
			{Name: "Bats", ElixirCost: 2, Level: 11, MaxLevel: 14},
			{Name: "Zap", ElixirCost: 2, Level: 12, MaxLevel: 14},
			{Name: "Rascals", ElixirCost: 5, Level: 11, MaxLevel: 14},
			{Name: "Knight", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Goblin Barrel", ElixirCost: 3, Level: 11, MaxLevel: 14},
		},
	}
	battleRamCycle.DeckDetail = convertCardsToDeckDetail(battleRamCycle.Cards)
	decks = append(decks, battleRamCycle)

	// Goblin Barrel Bait Deck
	goblinBarrelBait := &DeckAnalysis{
		DeckName:     "Goblin Barrel Bait Control",
		WinCondition: "Goblin Barrel",
		AverageElixir: 3.1,
		Strategy: "Bait out spells with Princess and Goblin Gang, then punish with Goblin Barrel. Control the game with cheap defensive units.",
		Cards: []clashroyale.Card{
			{Name: "Goblin Barrel", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Princess", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Goblin Gang", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Inferno Tower", ElixirCost: 5, Level: 11, MaxLevel: 14},
			{Name: "Knight", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Tesla", ElixirCost: 4, Level: 11, MaxLevel: 14},
			{Name: "Log", ElixirCost: 3, Level: 11, MaxLevel: 14},
			{Name: "Ice Spirit", ElixirCost: 1, Level: 11, MaxLevel: 14},
		},
	}
	goblinBarrelBait.DeckDetail = convertCardsToDeckDetail(goblinBarrelBait.Cards)
	decks = append(decks, goblinBarrelBait)

	return decks
}

// convertCardsToDeckDetail converts clashroyale.Card to CardDeckDetail
func convertCardsToDeckDetail(cards []clashroyale.Card) []CardDeckDetail {
	detail := make([]CardDeckDetail, len(cards))
	for i, card := range cards {
		detail[i] = CardDeckDetail{
			Name:              card.Name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Elixir:            card.ElixirCost,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
	}
	return detail
}

// scoreDeckForPlaystyle scores a deck based on how well it matches a playstyle
func scoreDeckForPlaystyle(deck *DeckAnalysis, playstyle *PlaystyleAnalysis) (int, []string) {
	score := 0
	reasons := make([]string, 0)

	// Factor 1: Aggression match
	if playstyle.ThreeCrownRate > 75 {
		// Very aggressive player
		if deck.AverageElixir < 2.9 {
			score += 40
			reasons = append(reasons, fmt.Sprintf("Ultra-low elixir (%.1f) matches your very aggressive playstyle", deck.AverageElixir))
		} else if deck.AverageElixir < 3.1 {
			score += 30
			reasons = append(reasons, fmt.Sprintf("Low elixir (%.1f) suits aggressive play", deck.AverageElixir))
		}
	} else if playstyle.ThreeCrownRate > 60 {
		// Moderately aggressive
		if deck.AverageElixir < 3.5 {
			score += 20
			reasons = append(reasons, fmt.Sprintf("Moderate elixir (%.1f) fits your playstyle", deck.AverageElixir))
		}
	}

	// Factor 2: Card level quality (simulated - would use actual player data)
	if len(deck.DeckDetail) > 0 {
		totalLevelRatio := 0.0
		for _, card := range deck.DeckDetail {
			totalLevelRatio += float64(card.Level) / float64(card.MaxLevel)
		}
		avgLevelRatio := totalLevelRatio / float64(len(deck.DeckDetail))
		score += int(avgLevelRatio * 30)
		reasons = append(reasons, fmt.Sprintf("Card level ratio: %.1f%%", avgLevelRatio*100))
	}

	// Factor 3: Win condition match
	if playstyle.CurrentWinCondition != "" && playstyle.CurrentWinCondition == deck.WinCondition {
		score += 15
		reasons = append(reasons, fmt.Sprintf("Already using %s - familiar playstyle", deck.WinCondition))
	}

	// Factor 4: Deck style compatibility
	switch playstyle.DeckStyle {
	case "Ultra-fast cycle":
		if deck.AverageElixir < 3.0 {
			score += 10
			reasons = append(reasons, "Matches your fast cycle preference")
		}
	case "Fast cycle":
		if deck.AverageElixir < 3.5 {
			score += 10
			reasons = append(reasons, "Fits your fast tempo style")
		}
	case "Balanced":
		if deck.AverageElixir >= 3.0 && deck.AverageElixir < 4.0 {
			score += 10
			reasons = append(reasons, "Balanced elixir cost for your style")
		}
	case "Beatdown/Heavy":
		if deck.AverageElixir >= 4.0 {
			score += 10
			reasons = append(reasons, "Heavy elixir deck for beatdown style")
		}
	}

	// Base score for having a viable deck
	score += 20

	return score, reasons
}

// determineCompatibility returns a compatibility rating based on score
func determineCompatibility(score int) string {
	switch {
	case score >= 80:
		return "Excellent Match"
	case score >= 60:
		return "Good Match"
	case score >= 40:
		return "Fair Match"
	default:
		return "Consider Alternatives"
	}
}