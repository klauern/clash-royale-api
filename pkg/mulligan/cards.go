package mulligan

import (
	_ "embed"
	"encoding/json"
)

//go:embed config/cards.json
var defaultCardsJSON []byte

// CardConfig represents the JSON configuration structure for card data
type CardConfig struct {
	Version int                    `json:"version"`
	Cards   map[string]CardInfoJSON `json:"cards"`
}

// CardInfoJSON represents a card in the JSON configuration
type CardInfoJSON struct {
	Name         string  `json:"name"`
	Elixir       int     `json:"elixir"`
	Type         string  `json:"type"`
	Rarity       string  `json:"rarity,omitempty"`
	Role         string  `json:"role"`
	OpeningScore float64 `json:"opening_score"`
}

// LoadCardDatabase loads card data from embedded JSON
func LoadCardDatabase() map[string]CardInfo {
	var config CardConfig
	if err := json.Unmarshal(defaultCardsJSON, &config); err != nil {
		// Return empty map on error (should not happen with embedded data)
		return make(map[string]CardInfo)
	}

	cards := make(map[string]CardInfo, len(config.Cards))
	for name, cardJSON := range config.Cards {
		cards[name] = CardInfo{
			Name:         cardJSON.Name,
			Elixir:       cardJSON.Elixir,
			Type:         cardJSON.Type,
			Rarity:       cardJSON.Rarity,
			Role:         CardRole(cardJSON.Role),
			OpeningScore: cardJSON.OpeningScore,
		}
	}
	return cards
}

// Deprecated: initializeCardDatabase creates a database of cards with their mulligan properties
// Use LoadCardDatabase instead. This function is kept for backward compatibility.
func initializeCardDatabase() map[string]CardInfo {
	return LoadCardDatabase()
}
