package deck

import (
	"math/big"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewDeckSpaceCalculator(t *testing.T) {
	tests := []struct {
		name    string
		player  *clashroyale.Player
		wantErr bool
	}{
		{
			name:    "nil player returns error",
			player:  nil,
			wantErr: true,
		},
		{
			name: "valid player with cards",
			player: &clashroyale.Player{
				Tag:  "#TEST123",
				Name: "Test Player",
				Cards: []clashroyale.Card{
					{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4, EvolutionLevel: 0},
					{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3, EvolutionLevel: 0},
					{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4, EvolutionLevel: 0},
					{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2, EvolutionLevel: 0},
					{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4, EvolutionLevel: 0},
					{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3, EvolutionLevel: 0},
					{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1, EvolutionLevel: 0},
					{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1, EvolutionLevel: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "player with evolution cards",
			player: &clashroyale.Player{
				Tag:  "#TEST456",
				Name: "Test Player 2",
				Cards: []clashroyale.Card{
					{Name: "Valkyrie", Level: 14, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4, EvolutionLevel: 1, MaxEvolutionLevel: 1},
					{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", ElixirCost: 3, EvolutionLevel: 1, MaxEvolutionLevel: 1},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc, err := NewDeckSpaceCalculator(tt.player)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeckSpaceCalculator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if calc == nil {
					t.Error("NewDeckSpaceCalculator() returned nil calculator")
					return
				}
				if len(calc.cards) != len(tt.player.Cards) {
					t.Errorf("NewDeckSpaceCalculator() cards count = %d, want %d", len(calc.cards), len(tt.player.Cards))
				}
			}
		})
	}
}

func TestDeckSpaceCalculator_CardCategorization(t *testing.T) {
	player := &clashroyale.Player{
		Tag:  "#TEST",
		Name: "Test",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},    // Win Condition
			{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},     // Building
			{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},     // Big Spell
			{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},        // Small Spell
			{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},    // Support
			{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},    // Support
			{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},  // Cycle
			{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1}, // Cycle
		},
	}

	calc, err := NewDeckSpaceCalculator(player)
	if err != nil {
		t.Fatalf("NewDeckSpaceCalculator() error = %v", err)
	}

	// Check role categorization
	expectedRoleCounts := map[CardRole]int{
		RoleWinCondition: 1,
		RoleBuilding:     1,
		RoleSpellBig:     1,
		RoleSpellSmall:   1,
		RoleSupport:      2,
		RoleCycle:        2,
	}

	for role, expectedCount := range expectedRoleCounts {
		actualCount := len(calc.cardsByRole[role])
		if actualCount != expectedCount {
			t.Errorf("Role %s: got %d cards, want %d", role, actualCount, expectedCount)
		}
	}
}

func TestCombinations(t *testing.T) {
	tests := []struct {
		name string
		n    int
		k    int
		want string // Use string to avoid overflow in test expectations
	}{
		{
			name: "C(8, 8) = 1",
			n:    8,
			k:    8,
			want: "1",
		},
		{
			name: "C(8, 0) = 1",
			n:    8,
			k:    0,
			want: "1",
		},
		{
			name: "C(8, 1) = 8",
			n:    8,
			k:    1,
			want: "8",
		},
		{
			name: "C(10, 3) = 120",
			n:    10,
			k:    3,
			want: "120",
		},
		{
			name: "C(52, 5) = 2598960 (poker hands)",
			n:    52,
			k:    5,
			want: "2598960",
		},
		{
			name: "C(100, 8) - large number",
			n:    100,
			k:    8,
			want: "186087894300",
		},
		{
			name: "C(5, 10) = 0 (k > n)",
			n:    5,
			k:    10,
			want: "0",
		},
		{
			name: "C(10, -1) = 0 (negative k)",
			n:    10,
			k:    -1,
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := combinations(tt.n, tt.k)
			want := new(big.Int)
			want.SetString(tt.want, 10)

			if got.Cmp(want) != 0 {
				t.Errorf("combinations(%d, %d) = %s, want %s", tt.n, tt.k, got.String(), tt.want)
			}
		})
	}
}

func TestFormatLargeNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string // Input as string to create big.Int
		want  string
	}{
		{
			name:  "nil returns 0",
			input: "",
			want:  "0",
		},
		{
			name:  "123 stays as is",
			input: "123",
			want:  "123",
		},
		{
			name:  "999 stays as is",
			input: "999",
			want:  "999",
		},
		{
			name:  "1234 becomes 1.2K",
			input: "1234",
			want:  "1.2K",
		},
		{
			name:  "1500 becomes 1.5K",
			input: "1500",
			want:  "1.5K",
		},
		{
			name:  "1234567 becomes 1.2M",
			input: "1234567",
			want:  "1.2M",
		},
		{
			name:  "1000000 becomes 1.0M",
			input: "1000000",
			want:  "1.0M",
		},
		{
			name:  "1234567890 becomes 1.2B",
			input: "1234567890",
			want:  "1.2B",
		},
		{
			name:  "1234567890123 becomes 1.2T",
			input: "1234567890123",
			want:  "1.2T",
		},
		{
			name:  "very large number",
			input: "186087894300",
			want:  "186.1B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input *big.Int
			if tt.input == "" {
				input = nil
			} else {
				input = new(big.Int)
				input.SetString(tt.input, 10)
			}

			got := FormatLargeNumber(input)
			if got != tt.want {
				t.Errorf("FormatLargeNumber(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestCalculateConstrainedCombinations(t *testing.T) {
	tests := []struct {
		name         string
		player       *clashroyale.Player
		wantZero     bool
		wantPositive bool
	}{
		{
			name: "insufficient cards returns zero",
			player: &clashroyale.Player{
				Tag:  "#TEST",
				Name: "Test",
				Cards: []clashroyale.Card{
					{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4}, // Only 1 card total
				},
			},
			wantZero:     true,
			wantPositive: false,
		},
		{
			name: "exactly 8 cards with perfect composition",
			player: &clashroyale.Player{
				Tag:  "#TEST",
				Name: "Test",
				Cards: []clashroyale.Card{
					{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
					{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
					{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
					{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
					{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
				},
			},
			wantZero:     false,
			wantPositive: true,
		},
		{
			name: "many cards in each role",
			player: &clashroyale.Player{
				Tag:  "#TEST",
				Name: "Test",
				Cards: []clashroyale.Card{
					// Win Conditions (3)
					{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Giant", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 5},
					{Name: "Goblin Barrel", Level: 11, MaxLevel: 14, Rarity: "Epic", ElixirCost: 3},
					// Buildings (2)
					{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
					{Name: "Tesla", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 4},
					// Big Spells (2)
					{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Poison", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					// Small Spells (3)
					{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
					{Name: "Log", Level: 13, MaxLevel: 14, Rarity: "Legendary", ElixirCost: 2},
					{Name: "Arrows", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
					// Support (4)
					{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
					{Name: "Baby Dragon", Level: 11, MaxLevel: 14, Rarity: "Epic", ElixirCost: 4},
					{Name: "Valkyrie", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
					// Cycle (4)
					{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
					{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
					{Name: "Bats", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
					{Name: "Electro Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
				},
			},
			wantZero:     false,
			wantPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc, err := NewDeckSpaceCalculator(tt.player)
			if err != nil {
				t.Fatalf("NewDeckSpaceCalculator() error = %v", err)
			}

			result := calc.calculateConstrainedCombinations()

			if tt.wantZero && result.Cmp(big.NewInt(0)) != 0 {
				t.Errorf("calculateConstrainedCombinations() = %s, want 0", result.String())
			}

			if tt.wantPositive && result.Cmp(big.NewInt(0)) <= 0 {
				t.Errorf("calculateConstrainedCombinations() = %s, want > 0", result.String())
			}
		})
	}
}

func TestCalculateStats(t *testing.T) {
	player := &clashroyale.Player{
		Tag:  "#TEST",
		Name: "Test Player",
		Cards: []clashroyale.Card{
			// Perfect 8-card composition
			{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
			{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
			{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
			{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
		},
	}

	calc, err := NewDeckSpaceCalculator(player)
	if err != nil {
		t.Fatalf("NewDeckSpaceCalculator() error = %v", err)
	}

	stats := calc.CalculateStats()

	// Basic checks
	if stats == nil {
		t.Fatal("CalculateStats() returned nil")
	}

	if stats.TotalCards != 8 {
		t.Errorf("TotalCards = %d, want 8", stats.TotalCards)
	}

	if stats.TotalCombinations == nil {
		t.Error("TotalCombinations is nil")
	}

	if stats.ValidCombinations == nil {
		t.Error("ValidCombinations is nil")
	}

	// With exactly 1 card per role (perfect composition), should be exactly 1 combination
	if stats.ValidCombinations.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("ValidCombinations = %s, want 1 (perfect 8-card composition)", stats.ValidCombinations.String())
	}

	// Total combinations for C(8, 8) = 1
	if stats.TotalCombinations.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("TotalCombinations = %s, want 1", stats.TotalCombinations.String())
	}

	// Check that role counts are correct
	expectedRoles := map[CardRole]int{
		RoleWinCondition: 1,
		RoleBuilding:     1,
		RoleSpellBig:     1,
		RoleSpellSmall:   1,
		RoleSupport:      2,
		RoleCycle:        2,
	}

	for role, expectedCount := range expectedRoles {
		actualCount, exists := stats.CardsByRole[role]
		if !exists {
			t.Errorf("Role %s not found in CardsByRole", role)
			continue
		}
		if actualCount != expectedCount {
			t.Errorf("CardsByRole[%s] = %d, want %d", role, actualCount, expectedCount)
		}
	}

	// Check that elixir range map is initialized
	if stats.ByElixirRange == nil {
		t.Error("ByElixirRange is nil")
	}

	// Check that archetype map is initialized
	if stats.ByArchetype == nil {
		t.Error("ByArchetype is nil")
	}
}

func TestCalculateStats_LargeCollection(t *testing.T) {
	// Create a player with many cards to test large combination calculations
	player := &clashroyale.Player{
		Tag:  "#LARGE",
		Name: "Large Collection",
		Cards: []clashroyale.Card{
			// 3 Win Conditions
			{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Giant", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 5},
			{Name: "Goblin Barrel", Level: 11, MaxLevel: 14, Rarity: "Epic", ElixirCost: 3},
			// 3 Buildings
			{Name: "Cannon", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
			{Name: "Tesla", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 4},
			{Name: "Inferno Tower", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 5},
			// 3 Big Spells
			{Name: "Fireball", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Poison", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Lightning", Level: 11, MaxLevel: 14, Rarity: "Epic", ElixirCost: 6},
			// 3 Small Spells
			{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
			{Name: "Log", Level: 13, MaxLevel: 14, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Arrows", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
			// 4 Support
			{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			{Name: "Archers", Level: 12, MaxLevel: 14, Rarity: "Common", ElixirCost: 3},
			{Name: "Baby Dragon", Level: 11, MaxLevel: 14, Rarity: "Epic", ElixirCost: 4},
			{Name: "Valkyrie", Level: 11, MaxLevel: 14, Rarity: "Rare", ElixirCost: 4},
			// 4 Cycle
			{Name: "Skeletons", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
			{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
			{Name: "Bats", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 2},
			{Name: "Electro Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", ElixirCost: 1},
		},
	}

	calc, err := NewDeckSpaceCalculator(player)
	if err != nil {
		t.Fatalf("NewDeckSpaceCalculator() error = %v", err)
	}

	stats := calc.CalculateStats()

	// Expected: C(3,1) * C(3,1) * C(3,1) * C(3,1) * C(4,2) * C(4,2)
	// = 3 * 3 * 3 * 3 * 6 * 6 = 2916
	expected := big.NewInt(2916)

	if stats.ValidCombinations.Cmp(expected) != 0 {
		t.Errorf("ValidCombinations = %s, want %s (3*3*3*3*6*6)", stats.ValidCombinations.String(), expected.String())
	}

	// Total cards should be 20
	if stats.TotalCards != 20 {
		t.Errorf("TotalCards = %d, want 20", stats.TotalCards)
	}
}
