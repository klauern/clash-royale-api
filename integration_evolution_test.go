//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// TestEvolutionAPIResponseStructure verifies that the Clash Royale API
// returns evolution data in the expected format.
//
// This test requires a valid CLASH_ROYALE_API_TOKEN and DEFAULT_PLAYER_TAG
// environment variables. It fetches real player data and verifies:
// - Field names match our struct tags (camelCase)
// - Evolution fields are present in the response
// - Data types are correct
// - Documents which cards support evolution
func TestEvolutionAPIResponseStructure(t *testing.T) {
	// Get API token
	token := os.Getenv("CLASH_ROYALE_API_TOKEN")
	if token == "" {
		t.Skip("CLASH_ROYALE_API_TOKEN not set, skipping integration test")
	}

	// Get player tag
	playerTag := os.Getenv("DEFAULT_PLAYER_TAG")
	if playerTag == "" {
		t.Skip("DEFAULT_PLAYER_TAG not set, skipping integration test")
	}

	// Create client
	client := clashroyale.NewClient(token)

	// Fetch player data
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		t.Fatalf("Failed to fetch player data: %v", err)
	}

	if len(player.Cards) == 0 {
		t.Fatal("Player has no cards - cannot verify evolution data")
	}

	// Track evolution statistics
	cardsWithEvolution := []string{}
	cardsWithMultiEvolution := []string{}
	maxEvolutionLevelSeen := 0
	hasEvolutionLevel := false
	hasMaxEvolutionLevel := false
	hasEvolutionIcon := false

	// Examine each card
	for _, card := range player.Cards {
		// Check if evolution fields are present
		if card.MaxEvolutionLevel > 0 {
			hasMaxEvolutionLevel = true
			cardsWithEvolution = append(cardsWithEvolution, card.Name)

			if card.MaxEvolutionLevel > maxEvolutionLevelSeen {
				maxEvolutionLevelSeen = card.MaxEvolutionLevel
			}

			if card.MaxEvolutionLevel > 1 {
				cardsWithMultiEvolution = append(cardsWithMultiEvolution,
					fmt.Sprintf("%s (maxEvo=%d)", card.Name, card.MaxEvolutionLevel))
			}

			if card.EvolutionLevel > 0 {
				hasEvolutionLevel = true
			}

			if card.IconUrls.EvolutionMedium != "" {
				hasEvolutionIcon = true
			}

			// Validate the data
			if err := card.Validate(); err != nil {
				t.Errorf("Card validation failed for %s: %v", card.Name, err)
			}

			// Check evolution level constraints
			if card.EvolutionLevel > card.MaxEvolutionLevel {
				t.Errorf("Card %s has evolutionLevel (%d) > maxEvolutionLevel (%d)",
					card.Name, card.EvolutionLevel, card.MaxEvolutionLevel)
			}

			// Check for negative values
			if card.EvolutionLevel < 0 {
				t.Errorf("Card %s has negative evolutionLevel: %d", card.Name, card.EvolutionLevel)
			}
			if card.MaxEvolutionLevel < 0 {
				t.Errorf("Card %s has negative maxEvolutionLevel: %d", card.Name, card.MaxEvolutionLevel)
			}
		}

		// Check StarLevel (separate from evolution)
		if card.StarLevel < 0 {
			t.Errorf("Card %s has negative starLevel: %d", card.Name, card.StarLevel)
		}
	}

	// Print findings
	t.Logf("\n=== Evolution API Response Verification ===")
	t.Logf("Total cards: %d", len(player.Cards))
	t.Logf("Cards with evolution support: %d", len(cardsWithEvolution))
	t.Logf("Max evolution level seen: %d", maxEvolutionLevelSeen)
	t.Logf("Has evolutionLevel field: %v", hasEvolutionLevel)
	t.Logf("Has maxEvolutionLevel field: %v", hasMaxEvolutionLevel)
	t.Logf("Has evolutionMedium icon: %v", hasEvolutionIcon)

	if len(cardsWithEvolution) > 0 {
		t.Logf("\nCards with evolution support (showing first 10):")
		for i, name := range cardsWithEvolution {
			if i >= 10 {
				t.Logf("  ... and %d more", len(cardsWithEvolution)-10)
				break
			}
			t.Logf("  - %s", name)
		}
	}

	if len(cardsWithMultiEvolution) > 0 {
		t.Logf("\nCards with multi-evolution support:")
		for _, name := range cardsWithMultiEvolution {
			t.Logf("  - %s", name)
		}
	}

	// Verify at least some evolution data was found
	if !hasMaxEvolutionLevel {
		t.Error("No cards with maxEvolutionLevel found - API structure may have changed")
	}

	// Print sample card JSON for documentation
	if len(cardsWithEvolution) > 0 {
		// Find the first evolved card
		for _, card := range player.Cards {
			if card.MaxEvolutionLevel > 0 {
				cardJSON, err := json.MarshalIndent(card, "", "  ")
				if err != nil {
					t.Errorf("Failed to marshal card JSON: %v", err)
				} else {
					t.Logf("\n=== Sample Evolution Card JSON ===\n%s\n", string(cardJSON))
				}
				break
			}
		}
	}

	t.Logf("\n=== Verification Complete ===")
}

// TestEvolutionFieldNaming verifies that JSON field names match our expectations
func TestEvolutionFieldNaming(t *testing.T) {
	// Test JSON unmarshaling with sample data
	sampleJSON := `{
		"id": 26000000,
		"name": "Knight",
		"level": 14,
		"maxLevel": 14,
		"count": 1000,
		"iconUrls": {
			"medium": "https://example.com/knight.png",
			"evolutionMedium": "https://example.com/knight_evo.png"
		},
		"elixirCost": 3,
		"type": "Troop",
		"rarity": "Common",
		"evolutionLevel": 2,
		"maxEvolutionLevel": 3,
		"starLevel": 0
	}`

	var card clashroyale.Card
	err := json.Unmarshal([]byte(sampleJSON), &card)
	if err != nil {
		t.Fatalf("Failed to unmarshal sample JSON: %v", err)
	}

	// Verify fields were populated
	if card.Name != "Knight" {
		t.Errorf("Expected name 'Knight', got '%s'", card.Name)
	}
	if card.EvolutionLevel != 2 {
		t.Errorf("Expected evolutionLevel 2, got %d", card.EvolutionLevel)
	}
	if card.MaxEvolutionLevel != 3 {
		t.Errorf("Expected maxEvolutionLevel 3, got %d", card.MaxEvolutionLevel)
	}
	if card.StarLevel != 0 {
		t.Errorf("Expected starLevel 0, got %d", card.StarLevel)
	}
	if card.IconUrls.EvolutionMedium != "https://example.com/knight_evo.png" {
		t.Errorf("Expected evolution icon URL, got '%s'", card.IconUrls.EvolutionMedium)
	}

	// Test round-trip (marshal and unmarshal)
	roundTripJSON, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
	}

	var card2 clashroyale.Card
	err = json.Unmarshal(roundTripJSON, &card2)
	if err != nil {
		t.Fatalf("Failed to unmarshal round-trip JSON: %v", err)
	}

	// Verify round-trip preserved evolution data
	if card2.EvolutionLevel != card.EvolutionLevel {
		t.Errorf("Round-trip lost evolutionLevel: %d != %d", card2.EvolutionLevel, card.EvolutionLevel)
	}
	if card2.MaxEvolutionLevel != card.MaxEvolutionLevel {
		t.Errorf("Round-trip lost maxEvolutionLevel: %d != %d", card2.MaxEvolutionLevel, card.MaxEvolutionLevel)
	}
}

// TestEvolutionOmitempty verifies that omitempty works correctly for evolution fields
func TestEvolutionOmitempty(t *testing.T) {
	// Card with no evolution should not include evolution fields in JSON
	cardNoEvo := clashroyale.Card{
		ID:         26000001,
		Name:       "P.E.K.K.A",
		Level:      12,
		MaxLevel:   14,
		ElixirCost: 7,
		Type:       "Troop",
		Rarity:     "Epic",
	}

	jsonData, err := json.Marshal(cardNoEvo)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
	}

	jsonStr := string(jsonData)

	// Verify evolution fields are omitted when zero
	if containsField(jsonStr, "evolutionLevel") {
		t.Error("Expected evolutionLevel to be omitted when zero")
	}
	if containsField(jsonStr, "maxEvolutionLevel") {
		t.Error("Expected maxEvolutionLevel to be omitted when zero")
	}
	if containsField(jsonStr, "starLevel") {
		t.Error("Expected starLevel to be omitted when zero")
	}

	// Card with evolution should include fields
	cardWithEvo := clashroyale.Card{
		ID:                26000000,
		Name:              "Knight",
		Level:             14,
		MaxLevel:          14,
		ElixirCost:        3,
		Type:              "Troop",
		Rarity:            "Common",
		EvolutionLevel:    1,
		MaxEvolutionLevel: 3,
	}

	jsonData, err = json.Marshal(cardWithEvo)
	if err != nil {
		t.Fatalf("Failed to marshal card with evolution: %v", err)
	}

	jsonStr = string(jsonData)

	if !containsField(jsonStr, "evolutionLevel") {
		t.Error("Expected evolutionLevel to be present when non-zero")
	}
	if !containsField(jsonStr, "maxEvolutionLevel") {
		t.Error("Expected maxEvolutionLevel to be present when non-zero")
	}
}

// Helper function to check if JSON string contains a field
func containsField(jsonStr, fieldName string) bool {
	return json.Valid([]byte(jsonStr)) &&
		(len(jsonStr) > 0 && fieldName != "" &&
			contains(jsonStr, fmt.Sprintf(`"%s":`, fieldName)))
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
