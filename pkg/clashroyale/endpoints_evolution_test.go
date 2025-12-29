package clashroyale

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPlayer_WithEvolutionData(t *testing.T) {
	tests := []struct {
		name         string
		responseJSON string
		checkFunc    func(t *testing.T, player *Player)
		wantErr      bool
	}{
		{
			name: "player with no evolution cards",
			responseJSON: `{
				"tag": "#TEST001",
				"name": "NoEvoPlayer",
				"expLevel": 14,
				"trophies": 5000,
				"bestTrophies": 5500,
				"wins": 100,
				"losses": 50,
				"battleCount": 150,
				"threeCrownWins": 25,
				"cards": [
					{
						"id": 26000001,
						"name": "P.E.K.K.A",
						"level": 12,
						"maxLevel": 14,
						"count": 50,
						"iconUrls": {
							"medium": "https://example.com/pekka.png"
						},
						"elixirCost": 7,
						"type": "Troop",
						"rarity": "Epic"
					}
				],
				"arena": {"id": 54000015, "name": "Legendary Arena"}
			}`,
			checkFunc: func(t *testing.T, player *Player) {
				if len(player.Cards) == 0 {
					t.Error("Expected cards to be present")
					return
				}
				card := player.Cards[0]
				if card.Name != "P.E.K.K.A" {
					t.Errorf("Expected card name P.E.K.K.A, got %s", card.Name)
				}
				if card.EvolutionLevel != 0 {
					t.Errorf("Expected evolutionLevel 0, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 0 {
					t.Errorf("Expected maxEvolutionLevel 0, got %d", card.MaxEvolutionLevel)
				}
			},
			wantErr: false,
		},
		{
			name: "player with single evolution card - unlocked",
			responseJSON: `{
				"tag": "#TEST002",
				"name": "SingleEvoPlayer",
				"expLevel": 14,
				"trophies": 6000,
				"bestTrophies": 6500,
				"wins": 200,
				"losses": 100,
				"battleCount": 300,
				"threeCrownWins": 50,
				"cards": [
					{
						"id": 26000002,
						"name": "Archers",
						"level": 14,
						"maxLevel": 14,
						"count": 100,
						"iconUrls": {
							"medium": "https://example.com/archers.png",
							"evolutionMedium": "https://example.com/archers_evo.png"
						},
						"elixirCost": 3,
						"type": "Troop",
						"rarity": "Common",
						"evolutionLevel": 1,
						"maxEvolutionLevel": 1,
						"starLevel": 0
					}
				],
				"arena": {"id": 54000015, "name": "Legendary Arena"}
			}`,
			checkFunc: func(t *testing.T, player *Player) {
				if len(player.Cards) == 0 {
					t.Error("Expected cards to be present")
					return
				}
				card := player.Cards[0]
				if card.Name != "Archers" {
					t.Errorf("Expected card name Archers, got %s", card.Name)
				}
				if card.EvolutionLevel != 1 {
					t.Errorf("Expected evolutionLevel 1, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 1 {
					t.Errorf("Expected maxEvolutionLevel 1, got %d", card.MaxEvolutionLevel)
				}
				if card.IconUrls.EvolutionMedium == "" {
					t.Error("Expected evolution icon URL to be present")
				}
			},
			wantErr: false,
		},
		{
			name: "player with multi-evolution card - partially unlocked",
			responseJSON: `{
				"tag": "#TEST003",
				"name": "MultiEvoPlayer",
				"expLevel": 14,
				"trophies": 7000,
				"bestTrophies": 7500,
				"wins": 300,
				"losses": 150,
				"battleCount": 450,
				"threeCrownWins": 75,
				"cards": [
					{
						"id": 26000000,
						"name": "Knight",
						"level": 14,
						"maxLevel": 14,
						"count": 200,
						"iconUrls": {
							"medium": "https://example.com/knight.png",
							"evolutionMedium": "https://example.com/knight_evo.png"
						},
						"elixirCost": 3,
						"type": "Troop",
						"rarity": "Common",
						"evolutionLevel": 2,
						"maxEvolutionLevel": 3,
						"starLevel": 3
					}
				],
				"arena": {"id": 54000015, "name": "Legendary Arena"}
			}`,
			checkFunc: func(t *testing.T, player *Player) {
				if len(player.Cards) == 0 {
					t.Error("Expected cards to be present")
					return
				}
				card := player.Cards[0]
				if card.Name != "Knight" {
					t.Errorf("Expected card name Knight, got %s", card.Name)
				}
				if card.EvolutionLevel != 2 {
					t.Errorf("Expected evolutionLevel 2, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 3 {
					t.Errorf("Expected maxEvolutionLevel 3, got %d", card.MaxEvolutionLevel)
				}
				if card.StarLevel != 3 {
					t.Errorf("Expected starLevel 3, got %d", card.StarLevel)
				}
			},
			wantErr: false,
		},
		{
			name: "player with evolution supported but not unlocked",
			responseJSON: `{
				"tag": "#TEST004",
				"name": "LockedEvoPlayer",
				"expLevel": 14,
				"trophies": 5500,
				"bestTrophies": 6000,
				"wins": 150,
				"losses": 75,
				"battleCount": 225,
				"threeCrownWins": 40,
				"cards": [
					{
						"id": 26000003,
						"name": "Bomber",
						"level": 11,
						"maxLevel": 14,
						"count": 500,
						"iconUrls": {
							"medium": "https://example.com/bomber.png"
						},
						"elixirCost": 2,
						"type": "Troop",
						"rarity": "Common",
						"maxEvolutionLevel": 1
					}
				],
				"arena": {"id": 54000015, "name": "Legendary Arena"}
			}`,
			checkFunc: func(t *testing.T, player *Player) {
				if len(player.Cards) == 0 {
					t.Error("Expected cards to be present")
					return
				}
				card := player.Cards[0]
				if card.Name != "Bomber" {
					t.Errorf("Expected card name Bomber, got %s", card.Name)
				}
				if card.EvolutionLevel != 0 {
					t.Errorf("Expected evolutionLevel 0 (not unlocked), got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 1 {
					t.Errorf("Expected maxEvolutionLevel 1, got %d", card.MaxEvolutionLevel)
				}
			},
			wantErr: false,
		},
		{
			name: "player with multiple cards - mixed evolution states",
			responseJSON: `{
				"tag": "#TEST005",
				"name": "MixedEvoPlayer",
				"expLevel": 14,
				"trophies": 8000,
				"bestTrophies": 8500,
				"wins": 500,
				"losses": 250,
				"battleCount": 750,
				"threeCrownWins": 125,
				"cards": [
					{
						"id": 26000000,
						"name": "Knight",
						"level": 14,
						"maxLevel": 14,
						"evolutionLevel": 3,
						"maxEvolutionLevel": 3,
						"iconUrls": {"medium": "url", "evolutionMedium": "evo_url"},
						"elixirCost": 3,
						"type": "Troop",
						"rarity": "Common"
					},
					{
						"id": 26000001,
						"name": "Archers",
						"level": 13,
						"maxLevel": 14,
						"maxEvolutionLevel": 1,
						"iconUrls": {"medium": "url"},
						"elixirCost": 3,
						"type": "Troop",
						"rarity": "Common"
					},
					{
						"id": 26000002,
						"name": "P.E.K.K.A",
						"level": 12,
						"maxLevel": 14,
						"iconUrls": {"medium": "url"},
						"elixirCost": 7,
						"type": "Troop",
						"rarity": "Epic"
					}
				],
				"arena": {"id": 54000015, "name": "Legendary Arena"}
			}`,
			checkFunc: func(t *testing.T, player *Player) {
				if len(player.Cards) != 3 {
					t.Errorf("Expected 3 cards, got %d", len(player.Cards))
					return
				}

				// Check Knight (fully evolved)
				knight := player.Cards[0]
				if knight.EvolutionLevel != 3 || knight.MaxEvolutionLevel != 3 {
					t.Errorf("Knight: expected evo 3/3, got %d/%d",
						knight.EvolutionLevel, knight.MaxEvolutionLevel)
				}

				// Check Archers (evolution supported but not unlocked)
				archers := player.Cards[1]
				if archers.EvolutionLevel != 0 || archers.MaxEvolutionLevel != 1 {
					t.Errorf("Archers: expected evo 0/1, got %d/%d",
						archers.EvolutionLevel, archers.MaxEvolutionLevel)
				}

				// Check P.E.K.K.A (no evolution support)
				pekka := player.Cards[2]
				if pekka.EvolutionLevel != 0 || pekka.MaxEvolutionLevel != 0 {
					t.Errorf("P.E.K.K.A: expected evo 0/0, got %d/%d",
						pekka.EvolutionLevel, pekka.MaxEvolutionLevel)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, tt.responseJSON)
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewClient("test_token")
			client.baseURL = server.URL

			// Make request
			player, err := client.GetPlayer("#TEST")

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlayer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Run custom checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, player)
			}

			// Validate all cards
			for _, card := range player.Cards {
				if err := card.Validate(); err != nil {
					t.Errorf("Card %s failed validation: %v", card.Name, err)
				}
			}
		})
	}
}

func TestGetCards_WithEvolutionData(t *testing.T) {
	responseJSON := `{
		"items": [
			{
				"id": 26000000,
				"name": "Knight",
				"maxLevel": 14,
				"iconUrls": {
					"medium": "https://example.com/knight.png",
					"evolutionMedium": "https://example.com/knight_evo.png"
				},
				"elixirCost": 3,
				"type": "Troop",
				"rarity": "Common",
				"maxEvolutionLevel": 3
			},
			{
				"id": 26000001,
				"name": "Archers",
				"maxLevel": 14,
				"iconUrls": {
					"medium": "https://example.com/archers.png",
					"evolutionMedium": "https://example.com/archers_evo.png"
				},
				"elixirCost": 3,
				"type": "Troop",
				"rarity": "Common",
				"maxEvolutionLevel": 1
			},
			{
				"id": 26000002,
				"name": "P.E.K.K.A",
				"maxLevel": 14,
				"iconUrls": {
					"medium": "https://example.com/pekka.png"
				},
				"elixirCost": 7,
				"type": "Troop",
				"rarity": "Epic"
			}
		]
	}`

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseJSON)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient("test_token")
	client.baseURL = server.URL

	// Make request
	cardList, err := client.GetCards()
	if err != nil {
		t.Fatalf("GetCards() error = %v", err)
	}

	if len(cardList.Items) != 3 {
		t.Fatalf("Expected 3 cards, got %d", len(cardList.Items))
	}

	// Check Knight
	knight := cardList.Items[0]
	if knight.Name != "Knight" {
		t.Errorf("Expected Knight, got %s", knight.Name)
	}
	if knight.MaxEvolutionLevel != 3 {
		t.Errorf("Knight: expected maxEvolutionLevel 3, got %d", knight.MaxEvolutionLevel)
	}
	if knight.IconUrls.EvolutionMedium == "" {
		t.Error("Knight: expected evolution icon URL")
	}

	// Check Archers
	archers := cardList.Items[1]
	if archers.Name != "Archers" {
		t.Errorf("Expected Archers, got %s", archers.Name)
	}
	if archers.MaxEvolutionLevel != 1 {
		t.Errorf("Archers: expected maxEvolutionLevel 1, got %d", archers.MaxEvolutionLevel)
	}

	// Check P.E.K.K.A (no evolution)
	pekka := cardList.Items[2]
	if pekka.Name != "P.E.K.K.A" {
		t.Errorf("Expected P.E.K.K.A, got %s", pekka.Name)
	}
	if pekka.MaxEvolutionLevel != 0 {
		t.Errorf("P.E.K.K.A: expected maxEvolutionLevel 0, got %d", pekka.MaxEvolutionLevel)
	}
	if pekka.IconUrls.EvolutionMedium != "" {
		t.Error("P.E.K.K.A: expected no evolution icon URL")
	}
}

func TestEvolutionJSON_Unmarshaling(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		checkFunc func(t *testing.T, card Card)
		wantErr   bool
	}{
		{
			name: "all evolution fields present",
			jsonData: `{
				"id": 26000000,
				"name": "Knight",
				"level": 14,
				"maxLevel": 14,
				"count": 1000,
				"iconUrls": {
					"medium": "url1",
					"evolutionMedium": "url2"
				},
				"elixirCost": 3,
				"type": "Troop",
				"rarity": "Common",
				"evolutionLevel": 2,
				"maxEvolutionLevel": 3,
				"starLevel": 1
			}`,
			checkFunc: func(t *testing.T, card Card) {
				if card.EvolutionLevel != 2 {
					t.Errorf("Expected evolutionLevel 2, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 3 {
					t.Errorf("Expected maxEvolutionLevel 3, got %d", card.MaxEvolutionLevel)
				}
				if card.StarLevel != 1 {
					t.Errorf("Expected starLevel 1, got %d", card.StarLevel)
				}
				if card.IconUrls.EvolutionMedium != "url2" {
					t.Errorf("Expected evolution icon 'url2', got '%s'", card.IconUrls.EvolutionMedium)
				}
			},
			wantErr: false,
		},
		{
			name: "evolution fields omitted (defaults to 0)",
			jsonData: `{
				"id": 26000001,
				"name": "P.E.K.K.A",
				"level": 12,
				"maxLevel": 14,
				"count": 50,
				"iconUrls": {
					"medium": "url1"
				},
				"elixirCost": 7,
				"type": "Troop",
				"rarity": "Epic"
			}`,
			checkFunc: func(t *testing.T, card Card) {
				if card.EvolutionLevel != 0 {
					t.Errorf("Expected evolutionLevel 0, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 0 {
					t.Errorf("Expected maxEvolutionLevel 0, got %d", card.MaxEvolutionLevel)
				}
				if card.StarLevel != 0 {
					t.Errorf("Expected starLevel 0, got %d", card.StarLevel)
				}
				if card.IconUrls.EvolutionMedium != "" {
					t.Errorf("Expected no evolution icon, got '%s'", card.IconUrls.EvolutionMedium)
				}
			},
			wantErr: false,
		},
		{
			name: "partial evolution data (maxEvolutionLevel only)",
			jsonData: `{
				"id": 26000002,
				"name": "Bomber",
				"level": 11,
				"maxLevel": 14,
				"count": 500,
				"iconUrls": {
					"medium": "url1"
				},
				"elixirCost": 2,
				"type": "Troop",
				"rarity": "Common",
				"maxEvolutionLevel": 1
			}`,
			checkFunc: func(t *testing.T, card Card) {
				if card.EvolutionLevel != 0 {
					t.Errorf("Expected evolutionLevel 0, got %d", card.EvolutionLevel)
				}
				if card.MaxEvolutionLevel != 1 {
					t.Errorf("Expected maxEvolutionLevel 1, got %d", card.MaxEvolutionLevel)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var card Card
			err := json.Unmarshal([]byte(tt.jsonData), &card)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, card)
			}
		})
	}
}
