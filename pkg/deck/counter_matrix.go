// Package deck provides counter relationship analysis for Clash Royale decks.
// This file implements the counter matrix that tracks which cards counter which threats.
package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// CounterCategory defines the type of counter capability a card provides
type CounterCategory string

const (
	CounterSplashDefense    CounterCategory = "splash_defense"    // Handles swarm pushes
	CounterTankKillers      CounterCategory = "tank_killers"      // High DPS or % damage
	CounterBuildingCounters CounterCategory = "building_counters" // Counters buildings
	CounterAirDefense       CounterCategory = "air_defense"       // Targets air
	CounterSwarmClear       CounterCategory = "swarm_clear"       // Clears swarms
	CounterBuildings        CounterCategory = "buildings"         // Defensive buildings

	defaultDataDir = "data"
)

// Counter represents a card that counters a specific threat
type Counter struct {
	Card          string  `json:"card"`
	Effectiveness float64 `json:"effectiveness"` // 0.0 to 1.0
	Reason        string  `json:"reason"`
}

// Threat represents a card or archetype that poses a threat
type Threat struct {
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Counters []Counter `json:"counters"`
}

// CounterCategoryGroup defines a group of cards that share a counter capability
type CounterCategoryGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Cards       []string `json:"cards"`
}

// counterDataFile represents the JSON data structure for loading counter relationships
type counterDataFile struct {
	Version           int                    `json:"version"`
	Description       string                 `json:"description"`
	LastUpdated       string                 `json:"last_updated"`
	Threats           []Threat               `json:"threats"`
	CounterCategories []CounterCategoryGroup `json:"counter_categories"`
}

// CounterMatrix holds counter relationships between cards
type CounterMatrix struct {
	// threatCounters maps threat names to their counters
	threatCounters map[string][]Counter

	// counterCategories maps category names to their card groups
	counterCategories map[CounterCategory][]string

	// cardCapabilities maps card names to the counter categories they provide
	cardCapabilities map[string][]CounterCategory

	mu sync.RWMutex
}

// NewCounterMatrix creates a new empty counter matrix
func NewCounterMatrix() *CounterMatrix {
	return &CounterMatrix{
		threatCounters:    make(map[string][]Counter),
		counterCategories: make(map[CounterCategory][]string),
		cardCapabilities:  make(map[string][]CounterCategory),
	}
}

// LoadCounterMatrix loads counter relationships from a JSON file
// If the file cannot be found or read, returns a matrix with default data
func LoadCounterMatrix(dataDir, filename string) *CounterMatrix {
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	if filename == "" {
		filename = "counter_relationships.json"
	}

	path := filepath.Join(dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		// Fall back to hardcoded defaults if file not found
		return NewCounterMatrixWithDefaults()
	}

	var counterData counterDataFile
	if err := json.Unmarshal(data, &counterData); err != nil {
		// Fall back to hardcoded defaults on parse error
		return NewCounterMatrixWithDefaults()
	}

	matrix := NewCounterMatrix()

	// Load threat counters
	for _, threat := range counterData.Threats {
		matrix.threatCounters[threat.Name] = threat.Counters

		// Build card capabilities from threat counters
		for _, counter := range threat.Counters {
			category := CounterCategory(threat.Category)
			matrix.cardCapabilities[counter.Card] = append(matrix.cardCapabilities[counter.Card], category)
		}
	}

	// Load counter categories
	for _, cat := range counterData.CounterCategories {
		category := CounterCategory(cat.Name)
		matrix.counterCategories[category] = cat.Cards

		// Update card capabilities from categories
		for _, card := range cat.Cards {
			matrix.cardCapabilities[card] = append(matrix.cardCapabilities[card], category)
		}
	}

	return matrix
}

// NewCounterMatrixWithDefaults creates a counter matrix with default data
// This provides basic counter relationships if no data file is available
func NewCounterMatrixWithDefaults() *CounterMatrix {
	matrix := NewCounterMatrix()

	// Default threat counters - Mega Knight
	matrix.threatCounters["Mega Knight"] = []Counter{
		{Card: "Inferno Tower", Effectiveness: 1.0, Reason: "Percentage damage destroys MK quickly"},
		{Card: "Inferno Dragon", Effectiveness: 0.95, Reason: "Percentage damage, needs protection"},
		{Card: "P.E.K.K.A", Effectiveness: 0.9, Reason: "High damage tanks through MK"},
	}

	// Default threat counters - Balloon
	matrix.threatCounters["Balloon"] = []Counter{
		{Card: "Inferno Tower", Effectiveness: 1.0, Reason: "Melts balloon instantly"},
		{Card: "Electro Wizard", Effectiveness: 0.9, Reason: "Targets air, resets abilities"},
		{Card: "Musketeer", Effectiveness: 0.85, Reason: "High DPS air targeting"},
	}

	// Default threat counters - Graveyard
	matrix.threatCounters["Graveyard"] = []Counter{
		{Card: "Tornado", Effectiveness: 0.95, Reason: "Groups skeletons for splash"},
		{Card: "Baby Dragon", Effectiveness: 0.9, Reason: "Splash clears skeletons efficiently"},
		{Card: "Valkyrie", Effectiveness: 0.9, Reason: "Spin attack clears all skeletons"},
	}

	// Default threat counters - Hog Rider
	matrix.threatCounters["Hog Rider"] = []Counter{
		{Card: "Tornado", Effectiveness: 0.9, Reason: "Pulls to King Tower activation"},
		{Card: "Cannon", Effectiveness: 0.85, Reason: "Cheap, reliable distraction"},
		{Card: "Tesla", Effectiveness: 0.85, Reason: "Hidden until Hog arrives"},
	}

	// Default counter categories
	matrix.counterCategories[CounterAirDefense] = []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Tesla"}
	matrix.counterCategories[CounterSplashDefense] = []string{"Valkyrie", "Baby Dragon", "Wizard"}
	matrix.counterCategories[CounterTankKillers] = []string{"Inferno Tower", "P.E.K.K.A", "Mini P.E.K.K.A"}
	matrix.counterCategories[CounterSwarmClear] = []string{"The Log", "Zap", "Arrows", "Valkyrie"}
	matrix.counterCategories[CounterBuildings] = []string{"Cannon", "Tesla", "Inferno Tower", "Bomb Tower"}

	// Build cardCapabilities from categories (required for HasCapability to work)
	for category, cards := range matrix.counterCategories {
		for _, card := range cards {
			matrix.cardCapabilities[card] = append(matrix.cardCapabilities[card], category)
		}
	}

	return matrix
}

// GetCountersForThreat returns all counters for a given threat
func (cm *CounterMatrix) GetCountersForThreat(threatName string) []Counter {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if counters, exists := cm.threatCounters[threatName]; exists {
		return counters
	}
	return nil
}

// GetCounterEffectiveness returns how effective a card is at countering a threat (0.0 to 1.0)
func (cm *CounterMatrix) GetCounterEffectiveness(threatName, cardName string) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if counters, exists := cm.threatCounters[threatName]; exists {
		for _, counter := range counters {
			if counter.Card == cardName {
				return counter.Effectiveness
			}
		}
	}
	return 0.0
}

// GetCardsInCategory returns all cards that belong to a counter category
func (cm *CounterMatrix) GetCardsInCategory(category CounterCategory) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cards, exists := cm.counterCategories[category]; exists {
		return cards
	}
	return nil
}

// GetCardCapabilities returns all counter categories a card belongs to
func (cm *CounterMatrix) GetCardCapabilities(cardName string) []CounterCategory {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if capabilities, exists := cm.cardCapabilities[cardName]; exists {
		return capabilities
	}
	return nil
}

// HasCapability returns true if a card has a specific counter capability
func (cm *CounterMatrix) HasCapability(cardName string, category CounterCategory) bool {
	capabilities := cm.GetCardCapabilities(cardName)
	for _, cap := range capabilities {
		if cap == category {
			return true
		}
	}
	return false
}

// CountCardsWithCapability counts how many cards in the deck have a specific capability
func (cm *CounterMatrix) CountCardsWithCapability(cardNames []string, category CounterCategory) int {
	count := 0
	for _, card := range cardNames {
		if cm.HasCapability(card, category) {
			count++
		}
	}
	return count
}

// AnalyzeThreatCoverage analyzes how well a deck can counter a specific threat
func (cm *CounterMatrix) AnalyzeThreatCoverage(deckCards []string, threatName string) ThreatCoverage {
	coverage := ThreatCoverage{
		ThreatName: threatName,
		CanCounter: false,
	}

	counters := cm.GetCountersForThreat(threatName)
	if counters == nil {
		coverage.Reason = fmt.Sprintf("No counter data for threat: %s", threatName)
		return coverage
	}

	var deckCounters []Counter
	var missingCounters []Counter

	for _, counter := range counters {
		// Check if this counter is in the deck
		for _, deckCard := range deckCards {
			if deckCard == counter.Card {
				deckCounters = append(deckCounters, counter)
				break
			}
		}
	}

	// Find missing high-effectiveness counters
	for _, counter := range counters {
		if counter.Effectiveness >= 0.8 {
			hasCounter := false
			for _, deckCounter := range deckCounters {
				if deckCounter.Card == counter.Card {
					hasCounter = true
					break
				}
			}
			if !hasCounter {
				missingCounters = append(missingCounters, counter)
			}
		}
	}

	coverage.CanCounter = len(deckCounters) > 0
	coverage.DeckCounters = deckCounters
	coverage.MissingCounters = missingCounters

	// Calculate overall effectiveness
	totalEffectiveness := 0.0
	for _, counter := range deckCounters {
		totalEffectiveness += counter.Effectiveness
	}
	if len(deckCounters) > 0 {
		coverage.Effectiveness = totalEffectiveness / float64(len(deckCounters))
	}

	// Generate reason
	if coverage.CanCounter {
		if coverage.Effectiveness >= 0.9 {
			coverage.Reason = fmt.Sprintf("Excellent counter: %v", formatCounterNames(deckCounters))
		} else if coverage.Effectiveness >= 0.7 {
			coverage.Reason = fmt.Sprintf("Good counter: %v", formatCounterNames(deckCounters))
		} else {
			coverage.Reason = fmt.Sprintf("Weak counter: %v", formatCounterNames(deckCounters))
		}
	} else {
		coverage.Reason = fmt.Sprintf("No counters to %s in deck", threatName)
		if len(missingCounters) > 0 {
			coverage.Suggestion = fmt.Sprintf("Consider adding: %v", formatCounterNames(missingCounters))
		}
	}

	return coverage
}

// ThreatCoverage represents how well a deck can counter a specific threat
type ThreatCoverage struct {
	ThreatName      string    `json:"threat_name"`
	CanCounter      bool      `json:"can_counter"`
	Effectiveness   float64   `json:"effectiveness"`
	DeckCounters    []Counter `json:"deck_counters"`
	MissingCounters []Counter `json:"missing_counters"`
	Reason          string    `json:"reason"`
	Suggestion      string    `json:"suggestion,omitempty"`
}

// formatCounterNames returns a comma-separated list of counter card names
func formatCounterNames(counters []Counter) string {
	if len(counters) == 0 {
		return "none"
	}
	names := make([]string, len(counters))
	for i, counter := range counters {
		names[i] = counter.Card
	}
	return joinString(names, ", ")
}

// joinString joins a slice of strings with a separator
func joinString(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
