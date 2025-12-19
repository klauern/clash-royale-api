package clashroyale

import (
	"testing"
)

func TestCard_Validate_Evolution(t *testing.T) {
	tests := []struct {
		name    string
		card    Card
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid - no evolution support",
			card: Card{
				Name:              "P.E.K.K.A",
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
			},
			wantErr: false,
		},
		{
			name: "valid - evolution supported but not unlocked",
			card: Card{
				Name:              "Knight",
				EvolutionLevel:    0,
				MaxEvolutionLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "valid - single evolution unlocked",
			card: Card{
				Name:              "Archers",
				EvolutionLevel:    1,
				MaxEvolutionLevel: 1,
			},
			wantErr: false,
		},
		{
			name: "valid - multi-evolution partially unlocked",
			card: Card{
				Name:              "Knight",
				EvolutionLevel:    2,
				MaxEvolutionLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "valid - multi-evolution fully unlocked",
			card: Card{
				Name:              "Musketeer",
				EvolutionLevel:    3,
				MaxEvolutionLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "invalid - evolution level exceeds max",
			card: Card{
				Name:              "Knight",
				EvolutionLevel:    4,
				MaxEvolutionLevel: 3,
			},
			wantErr: true,
			errMsg:  "evolution level 4 cannot be greater than max evolution level 3",
		},
		{
			name: "invalid - evolution level exceeds max by 1",
			card: Card{
				Name:              "Bomber",
				EvolutionLevel:    2,
				MaxEvolutionLevel: 1,
			},
			wantErr: true,
			errMsg:  "evolution level 2 cannot be greater than max evolution level 1",
		},
		{
			name: "valid - both evolution fields at zero",
			card: Card{
				Name:              "Goblin Barrel",
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
			},
			wantErr: false,
		},
		{
			name: "valid - realistic scenario from API",
			card: Card{
				ID:                27000012,
				Name:              "Goblin Cage",
				Level:             7,
				MaxLevel:          14,
				Count:             22,
				ElixirCost:        4,
				Rarity:            "rare",
				MaxEvolutionLevel: 1,
				StarLevel:         1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Card.Validate() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Card.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Card.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCard_Validate_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name    string
		card    Card
		wantErr bool
	}{
		{
			name: "max evolution level at upper bound",
			card: Card{
				Name:              "Knight",
				EvolutionLevel:    3,
				MaxEvolutionLevel: 3,
			},
			wantErr: false,
		},
		{
			name: "evolution level at max",
			card: Card{
				Name:              "Test Card",
				EvolutionLevel:    10,
				MaxEvolutionLevel: 10,
			},
			wantErr: false,
		},
		{
			name: "single evolution level",
			card: Card{
				Name:              "Cannon",
				EvolutionLevel:    1,
				MaxEvolutionLevel: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Card.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCard_Validate_StarLevel(t *testing.T) {
	tests := []struct {
		name     string
		starLevel int
		wantErr  bool
	}{
		{"star level 0", 0, false},
		{"star level 1", 1, false},
		{"star level 2", 2, false},
		{"star level 3", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Card{
				Name:      "Test Card",
				StarLevel: tt.starLevel,
			}

			err := card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Card.Validate() with starLevel=%d error = %v, wantErr %v",
					tt.starLevel, err, tt.wantErr)
			}
		})
	}
}

func TestCard_Validate_MultipleCards(t *testing.T) {
	// Test a collection of cards similar to what would be returned from the API
	cards := []Card{
		{Name: "Knight", EvolutionLevel: 0, MaxEvolutionLevel: 3},
		{Name: "Archers", EvolutionLevel: 1, MaxEvolutionLevel: 1},
		{Name: "P.E.K.K.A", EvolutionLevel: 0, MaxEvolutionLevel: 0},
		{Name: "Musketeer", EvolutionLevel: 3, MaxEvolutionLevel: 3},
		{Name: "Giant", EvolutionLevel: 1, MaxEvolutionLevel: 2},
	}

	for _, card := range cards {
		t.Run(card.Name, func(t *testing.T) {
			err := card.Validate()
			if err != nil {
				t.Errorf("Card %s failed validation: %v", card.Name, err)
			}
		})
	}
}

func TestCard_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		card    Card
		wantErr bool
		errMsg  string
	}{
		{
			name: "evolution level far exceeds max",
			card: Card{
				Name:              "Test",
				EvolutionLevel:    100,
				MaxEvolutionLevel: 1,
			},
			wantErr: true,
			errMsg:  "evolution level 100 cannot be greater than max evolution level 1",
		},
		{
			name: "evolution level just above max",
			card: Card{
				Name:              "Test",
				EvolutionLevel:    1,
				MaxEvolutionLevel: 0,
			},
			wantErr: true,
			errMsg:  "evolution level 1 cannot be greater than max evolution level 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Card.Validate() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Card.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Card.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCard_Validate_NegativeValues(t *testing.T) {
	tests := []struct {
		name    string
		card    Card
		wantErr bool
		errMsg  string
	}{
		{
			name: "negative evolution level",
			card: Card{
				Name:              "Test",
				EvolutionLevel:    -1,
				MaxEvolutionLevel: 3,
			},
			wantErr: true,
			errMsg:  "evolution level cannot be negative: -1",
		},
		{
			name: "negative max evolution level",
			card: Card{
				Name:              "Test",
				EvolutionLevel:    0,
				MaxEvolutionLevel: -2,
			},
			wantErr: true,
			errMsg:  "max evolution level cannot be negative: -2",
		},
		{
			name: "negative star level",
			card: Card{
				Name:      "Test",
				StarLevel: -3,
			},
			wantErr: true,
			errMsg:  "star level cannot be negative: -3",
		},
		{
			name: "multiple negative values - evolution level checked first",
			card: Card{
				Name:              "Test",
				EvolutionLevel:    -1,
				MaxEvolutionLevel: -2,
				StarLevel:         -3,
			},
			wantErr: true,
			errMsg:  "evolution level cannot be negative: -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Card.Validate() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Card.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Card.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

