package evaluation

import (
	"strings"
	"testing"
)

func TestGenerateDeckLink(t *testing.T) {
	tests := []struct {
		name        string
		cards       []string
		wantValid   bool
		wantError   string
		checkURL    bool
		urlContains string
	}{
		{
			name: "valid 8-card deck",
			cards: []string{
				"Hog Rider",
				"Musketeer",
				"Knight",
				"Fireball",
				"Zap",
				"Tesla",
				"Ice Spirit",
				"Cannon",
			},
			wantValid:   true,
			checkURL:    true,
			urlContains: "link.clashroyale.com/deck/en?deck=",
		},
		{
			name: "deck with 7 cards - invalid",
			cards: []string{
				"Hog Rider",
				"Musketeer",
				"Knight",
				"Fireball",
				"Zap",
				"Tesla",
				"Ice Spirit",
			},
			wantValid: false,
			wantError: "deck must contain exactly 8 cards",
		},
		{
			name: "deck with unknown card",
			cards: []string{
				"Hog Rider",
				"Musketeer",
				"Knight",
				"Fireball",
				"Zap",
				"Tesla",
				"Ice Spirit",
				"Unknown Card",
			},
			wantValid: false,
			wantError: "unknown card",
		},
		{
			name: "case insensitive card matching",
			cards: []string{
				"hog rider",
				"MUSKETEER",
				"KnIgHt",
				"Fireball",
				"Zap",
				"Tesla",
				"Ice Spirit",
				"Cannon",
			},
			wantValid:   true,
			checkURL:    true,
			urlContains: "link.clashroyale.com/deck/en?deck=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := GenerateDeckLink(tt.cards)

			if link.Valid != tt.wantValid {
				t.Errorf("GenerateDeckLink() Valid = %v, want %v", link.Valid, tt.wantValid)
			}

			if !tt.wantValid {
				if tt.wantError != "" && !strings.Contains(link.Error, tt.wantError) {
					t.Errorf("GenerateDeckLink() Error = %v, want error containing %v", link.Error, tt.wantError)
				}
			} else {
				if tt.checkURL && !strings.Contains(link.URL, tt.urlContains) {
					t.Errorf("GenerateDeckLink() URL = %v, want URL containing %v", link.URL, tt.urlContains)
				}

				// Verify card IDs are present
				if len(link.CardIDs) != 8 {
					t.Errorf("GenerateDeckLink() CardIDs length = %d, want 8", len(link.CardIDs))
				}
			}
		})
	}
}

func TestValidateDeckLink(t *testing.T) {
	tests := []struct {
		name    string
		link    *DeckLink
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid link",
			link: &DeckLink{
				URL:     "https://link.clashroyale.com/deck/en?deck=26000021%3B26000014%3B26000000%3B28000000%3B28000008%3B27000006%3B26000030%3B27000000",
				Valid:   true,
				CardIDs: []string{"26000021", "26000014", "26000000", "28000000", "28000008", "27000006", "26000030", "27000000"},
			},
			wantErr: false,
		},
		{
			name: "invalid link - not valid",
			link: &DeckLink{
				Valid: false,
				Error: "test error",
			},
			wantErr: true,
			errMsg:  "invalid link",
		},
		{
			name: "invalid link - empty URL",
			link: &DeckLink{
				Valid: true,
				URL:   "",
			},
			wantErr: true,
			errMsg:  "empty URL",
		},
		{
			name: "invalid link - wrong host",
			link: &DeckLink{
				Valid: true,
				URL:   "https://example.com/deck?deck=123",
			},
			wantErr: true,
			errMsg:  "invalid host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeckLink(tt.link)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeckLink() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateDeckLink() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestGetCardName(t *testing.T) {
	tests := []struct {
		cardID   string
		wantName string
	}{
		{
			cardID:   "26000021",
			wantName: "Hog Rider",
		},
		{
			cardID:   "26000014",
			wantName: "Musketeer",
		},
		{
			cardID:   "28000000",
			wantName: "Fireball",
		},
		{
			cardID:   "27000000",
			wantName: "Cannon",
		},
		{
			cardID:   "99999999",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.cardID, func(t *testing.T) {
			gotName := GetCardName(tt.cardID)
			if gotName != tt.wantName {
				t.Errorf("GetCardName(%s) = %v, want %v", tt.cardID, gotName, tt.wantName)
			}
		})
	}
}
