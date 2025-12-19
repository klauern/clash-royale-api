package deck

import (
	"fmt"
	"sort"
)

// SynergyCategory defines common synergy patterns between cards
type SynergyCategory string

const (
	SynergyTankSupport  SynergyCategory = "tank_support"   // Tank + Support troops
	SynergyBait         SynergyCategory = "bait"           // Bait combos (log bait, zap bait)
	SynergySpellCombo   SynergyCategory = "spell_combo"    // Spell combinations
	SynergyWinCondition SynergyCategory = "win_condition"  // Win condition combos
	SynergyDefensive    SynergyCategory = "defensive"      // Defensive combinations
	SynergyCycle        SynergyCategory = "cycle"          // Cycle card combinations
	SynergyBridgeSpam   SynergyCategory = "bridge_spam"    // Bridge spam combinations
)

// SynergyPair represents synergy between two cards
type SynergyPair struct {
	Card1       string          `json:"card1"`
	Card2       string          `json:"card2"`
	SynergyType SynergyCategory `json:"synergy_type"`
	Score       float64         `json:"score"`        // 0.0 to 1.0
	Description string          `json:"description"`
}

// SynergyDatabase holds known card synergies
type SynergyDatabase struct {
	Pairs      []SynergyPair             `json:"pairs"`
	Categories map[SynergyCategory][]SynergyPair `json:"categories"`
}

// DeckSynergyAnalysis represents the synergy analysis of a deck
type DeckSynergyAnalysis struct {
	TotalScore      float64                          `json:"total_score"`       // 0-100
	AverageScore    float64                          `json:"average_score"`     // Average synergy per pair
	TopSynergies    []SynergyPair                    `json:"top_synergies"`     // Best synergies in deck
	MissingSynergies []string                        `json:"missing_synergies"` // Cards that don't synergize well
	CategoryScores  map[SynergyCategory]int          `json:"category_scores"`   // Count by category
}

// SynergyRecommendation suggests a card to add based on synergies
type SynergyRecommendation struct {
	CardName     string        `json:"card_name"`
	SynergyScore float64       `json:"synergy_score"`
	Synergies    []SynergyPair `json:"synergies"` // Synergies with current deck
	Reason       string        `json:"reason"`
}

// NewSynergyDatabase creates a synergy database with known card combinations
func NewSynergyDatabase() *SynergyDatabase {
	pairs := []SynergyPair{
		// Tank + Support synergies
		{Card1: "Giant", Card2: "Witch", SynergyType: SynergyTankSupport, Score: 0.9, Description: "Witch supports Giant with splash damage and spawns"},
		{Card1: "Giant", Card2: "Sparky", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Giant tanks while Sparky deals massive damage"},
		{Card1: "Giant", Card2: "Musketeer", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Musketeer provides ranged support behind Giant"},
		{Card1: "Golem", Card2: "Night Witch", SynergyType: SynergyTankSupport, Score: 0.95, Description: "Classic Golem beatdown synergy"},
		{Card1: "Golem", Card2: "Baby Dragon", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Baby Dragon provides splash support"},
		{Card1: "Lava Hound", Card2: "Balloon", SynergyType: SynergyWinCondition, Score: 0.95, Description: "LavaLoon: overwhelming air pressure"},
		{Card1: "Lava Hound", Card2: "Miner", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Miner supports Lava Hound pups"},

		// Bait synergies
		{Card1: "Goblin Barrel", Card2: "Princess", SynergyType: SynergyBait, Score: 0.95, Description: "Log bait: Princess baits log for Goblin Barrel"},
		{Card1: "Goblin Barrel", Card2: "Goblin Gang", SynergyType: SynergyBait, Score: 0.9, Description: "Multiple goblin threats overwhelm spells"},
		{Card1: "Goblin Barrel", Card2: "Dart Goblin", SynergyType: SynergyBait, Score: 0.85, Description: "Dart Goblin baits small spells"},
		{Card1: "Skeleton Barrel", Card2: "Goblin Barrel", SynergyType: SynergyBait, Score: 0.8, Description: "Double barrel pressure"},
		{Card1: "Princess", Card2: "Goblin Gang", SynergyType: SynergyBait, Score: 0.85, Description: "Log bait pressure"},

		// Spell combos
		{Card1: "Hog Rider", Card2: "Fireball", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Fireball clears defenders for Hog"},
		{Card1: "Hog Rider", Card2: "Earthquake", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Earthquake destroys buildings for Hog"},
		{Card1: "Tornado", Card2: "Fireball", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Tornado groups troops for Fireball"},
		{Card1: "Tornado", Card2: "Rocket", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Tornado + Rocket tower finish"},
		{Card1: "Graveyard", Card2: "Freeze", SynergyType: SynergySpellCombo, Score: 0.9, Description: "Freeze allows Graveyard skeletons to connect"},
		{Card1: "Graveyard", Card2: "Poison", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Poison clears small troops from Graveyard"},

		// Bridge spam
		{Card1: "P.E.K.K.A", Card2: "Battle Ram", SynergyType: SynergyBridgeSpam, Score: 0.85, Description: "PEKKA Bridge Spam pressure"},
		{Card1: "P.E.K.K.A", Card2: "Bandit", SynergyType: SynergyBridgeSpam, Score: 0.8, Description: "Bandit supports PEKKA counterpush"},
		{Card1: "Bandit", Card2: "Battle Ram", SynergyType: SynergyBridgeSpam, Score: 0.8, Description: "Fast dual-lane pressure"},
		{Card1: "Royal Ghost", Card2: "Bandit", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "Invisible bridge spam"},

		// Defensive synergies
		{Card1: "Cannon", Card2: "Ice Spirit", SynergyType: SynergyDefensive, Score: 0.8, Description: "Cheap defensive combo"},
		{Card1: "Tesla", Card2: "Ice Spirit", SynergyType: SynergyDefensive, Score: 0.75, Description: "Tesla + Ice Spirit kiting"},
		{Card1: "Inferno Tower", Card2: "Zap", SynergyType: SynergyDefensive, Score: 0.85, Description: "Zap resets for Inferno Tower"},
		{Card1: "Inferno Dragon", Card2: "Zap", SynergyType: SynergyDefensive, Score: 0.8, Description: "Zap protects Inferno Dragon beam"},

		// Cycle synergies
		{Card1: "Ice Spirit", Card2: "Skeletons", SynergyType: SynergyCycle, Score: 0.85, Description: "Ultra-cheap cycle combo"},
		{Card1: "Ice Spirit", Card2: "Log", SynergyType: SynergyCycle, Score: 0.8, Description: "Cheap cycle and control"},
		{Card1: "Skeletons", Card2: "Log", SynergyType: SynergyCycle, Score: 0.75, Description: "Cycle and clear combo"},

		// Win condition synergies
		{Card1: "Miner", Card2: "Balloon", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Miner tanks for Balloon"},
		{Card1: "Miner", Card2: "Goblin Barrel", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Dual win condition pressure"},
		{Card1: "X-Bow", Card2: "Tesla", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Double building lock"},
		{Card1: "Mortar", Card2: "Cannon", SynergyType: SynergyWinCondition, Score: 0.85, Description: "Mortar + defensive building"},
	}

	// Organize by category
	categories := make(map[SynergyCategory][]SynergyPair)
	for _, pair := range pairs {
		categories[pair.SynergyType] = append(categories[pair.SynergyType], pair)
	}

	return &SynergyDatabase{
		Pairs:      pairs,
		Categories: categories,
	}
}

// GetSynergy returns the synergy score between two cards (0.0 to 1.0)
// Returns 0 if no known synergy exists
func (db *SynergyDatabase) GetSynergy(card1, card2 string) float64 {
	for _, pair := range db.Pairs {
		if (pair.Card1 == card1 && pair.Card2 == card2) ||
			(pair.Card1 == card2 && pair.Card2 == card1) {
			return pair.Score
		}
	}
	return 0.0
}

// GetSynergyPair returns the synergy pair details if it exists
func (db *SynergyDatabase) GetSynergyPair(card1, card2 string) *SynergyPair {
	for _, pair := range db.Pairs {
		if (pair.Card1 == card1 && pair.Card2 == card2) ||
			(pair.Card1 == card2 && pair.Card2 == card1) {
			return &pair
		}
	}
	return nil
}

// AnalyzeDeckSynergy scores overall deck synergy
func (db *SynergyDatabase) AnalyzeDeckSynergy(deck []string) *DeckSynergyAnalysis {
	if len(deck) == 0 {
		return &DeckSynergyAnalysis{}
	}

	topSynergies := make([]SynergyPair, 0)
	totalScore := 0.0
	pairCount := 0
	categoryScores := make(map[SynergyCategory]int)
	cardSynergyCounts := make(map[string]int)

	// Check all pairs
	for i := 0; i < len(deck); i++ {
		for j := i + 1; j < len(deck); j++ {
			if pair := db.GetSynergyPair(deck[i], deck[j]); pair != nil {
				topSynergies = append(topSynergies, *pair)
				totalScore += pair.Score
				pairCount++
				categoryScores[pair.SynergyType]++
				cardSynergyCounts[deck[i]]++
				cardSynergyCounts[deck[j]]++
			}
		}
	}

	// Sort top synergies by score
	sort.Slice(topSynergies, func(i, j int) bool {
		return topSynergies[i].Score > topSynergies[j].Score
	})

	// Limit to top 5
	if len(topSynergies) > 5 {
		topSynergies = topSynergies[:5]
	}

	// Find cards with no synergies
	missingSynergies := make([]string, 0)
	for _, card := range deck {
		if cardSynergyCounts[card] == 0 {
			missingSynergies = append(missingSynergies, card)
		}
	}

	// Calculate average
	avgScore := 0.0
	if pairCount > 0 {
		avgScore = totalScore / float64(pairCount)
	}

	// Normalize total score to 0-100
	// Maximum possible score with 8 cards = 28 pairs * 1.0 = 28
	// Scale to 100
	normalizedTotal := (totalScore / 28.0) * 100.0

	return &DeckSynergyAnalysis{
		TotalScore:       normalizedTotal,
		AverageScore:     avgScore,
		TopSynergies:     topSynergies,
		MissingSynergies: missingSynergies,
		CategoryScores:   categoryScores,
	}
}

// SuggestSynergyCards recommends cards that synergize with the current deck
func (db *SynergyDatabase) SuggestSynergyCards(currentDeck []string, available []*CardCandidate) []*SynergyRecommendation {
	if len(currentDeck) == 0 || len(available) == 0 {
		return nil
	}

	// Create map for quick lookup
	inDeck := make(map[string]bool)
	for _, card := range currentDeck {
		inDeck[card] = true
	}

	// Score each available card by its synergies with current deck
	recommendations := make(map[string]*SynergyRecommendation)

	for _, candidate := range available {
		// Skip cards already in deck
		if inDeck[candidate.Name] {
			continue
		}

		synergies := make([]SynergyPair, 0)
		totalSynergy := 0.0

		// Check synergies with each card in deck
		for _, deckCard := range currentDeck {
			if pair := db.GetSynergyPair(candidate.Name, deckCard); pair != nil {
				synergies = append(synergies, *pair)
				totalSynergy += pair.Score
			}
		}

		// Only recommend if has synergies
		if len(synergies) > 0 {
			avgSynergy := totalSynergy / float64(len(synergies))

			reason := fmt.Sprintf("Synergizes with %d cards in your deck", len(synergies))
			if len(synergies) >= 3 {
				reason = fmt.Sprintf("Strong synergies with %d cards: %s", len(synergies), synergies[0].Card2)
			}

			recommendations[candidate.Name] = &SynergyRecommendation{
				CardName:     candidate.Name,
				SynergyScore: avgSynergy,
				Synergies:    synergies,
				Reason:       reason,
			}
		}
	}

	// Convert to slice and sort by score
	result := make([]*SynergyRecommendation, 0, len(recommendations))
	for _, rec := range recommendations {
		result = append(result, rec)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SynergyScore > result[j].SynergyScore
	})

	// Return top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// GetCategoryDescription returns a human-readable description of a synergy category
func GetCategoryDescription(category SynergyCategory) string {
	descriptions := map[SynergyCategory]string{
		SynergyTankSupport:  "Tank + Support",
		SynergyBait:         "Spell Bait",
		SynergySpellCombo:   "Spell Combo",
		SynergyWinCondition: "Win Condition",
		SynergyDefensive:    "Defensive",
		SynergyCycle:        "Cycle",
		SynergyBridgeSpam:   "Bridge Spam",
	}
	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return string(category)
}
