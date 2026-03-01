package deck

import (
	"testing"
)

// Helper function to create pointer to CardRole
//
//go:fix inline
func rolePtr(r CardRole) *CardRole {
	return new(r)
}

// TestClassifyCard tests role classification for known cards
func TestClassifyCard(t *testing.T) {
	tests := []struct {
		name       string
		cardName   string
		elixirCost int
		wantRole   *CardRole
	}{
		// Win Conditions
		{
			name:       "Hog Rider is win condition",
			cardName:   "Hog Rider",
			elixirCost: 4,
			wantRole:   new(RoleWinCondition),
		},
		{
			name:       "Royal Giant is win condition",
			cardName:   "Royal Giant",
			elixirCost: 6,
			wantRole:   new(RoleWinCondition),
		},
		{
			name:       "Battle Ram is win condition",
			cardName:   "Battle Ram",
			elixirCost: 4,
			wantRole:   new(RoleWinCondition),
		},
		{
			name:       "Goblin Barrel is win condition",
			cardName:   "Goblin Barrel",
			elixirCost: 3,
			wantRole:   new(RoleWinCondition),
		},
		// Buildings
		{
			name:       "Cannon is building",
			cardName:   "Cannon",
			elixirCost: 3,
			wantRole:   new(RoleBuilding),
		},
		{
			name:       "Goblin Cage is building",
			cardName:   "Goblin Cage",
			elixirCost: 4,
			wantRole:   new(RoleBuilding),
		},
		{
			name:       "Tesla is building",
			cardName:   "Tesla",
			elixirCost: 4,
			wantRole:   new(RoleBuilding),
		},
		// Big Spells
		{
			name:       "Fireball is big spell",
			cardName:   "Fireball",
			elixirCost: 4,
			wantRole:   new(RoleSpellBig),
		},
		{
			name:       "Poison is big spell",
			cardName:   "Poison",
			elixirCost: 4,
			wantRole:   new(RoleSpellBig),
		},
		{
			name:       "Lightning is big spell",
			cardName:   "Lightning",
			elixirCost: 6,
			wantRole:   new(RoleSpellBig),
		},
		// Small Spells
		{
			name:       "Zap is small spell",
			cardName:   "Zap",
			elixirCost: 2,
			wantRole:   new(RoleSpellSmall),
		},
		{
			name:       "Log is small spell",
			cardName:   "Log",
			elixirCost: 2,
			wantRole:   new(RoleSpellSmall),
		},
		{
			name:       "Arrows is small spell",
			cardName:   "Arrows",
			elixirCost: 3,
			wantRole:   new(RoleSpellSmall),
		},
		// Cycle Cards
		{
			name:       "Skeletons is cycle",
			cardName:   "Skeletons",
			elixirCost: 1,
			wantRole:   new(RoleCycle),
		},
		{
			name:       "Fire Spirit is cycle",
			cardName:   "Fire Spirit",
			elixirCost: 1,
			wantRole:   new(RoleCycle),
		},
		{
			name:       "Ice Spirit is cycle",
			cardName:   "Ice Spirit",
			elixirCost: 1,
			wantRole:   new(RoleCycle),
		},
		{
			name:       "Bats is cycle",
			cardName:   "Bats",
			elixirCost: 2,
			wantRole:   new(RoleCycle),
		},
		// Support Troops
		{
			name:       "Musketeer is support",
			cardName:   "Musketeer",
			elixirCost: 4,
			wantRole:   new(RoleSupport),
		},
		{
			name:       "Knight is support",
			cardName:   "Knight",
			elixirCost: 3,
			wantRole:   new(RoleSupport),
		},
		{
			name:       "Goblin Gang is support",
			cardName:   "Goblin Gang",
			elixirCost: 3,
			wantRole:   new(RoleSupport),
		},
		{
			name:       "Skeleton Dragons is support",
			cardName:   "Skeleton Dragons",
			elixirCost: 4,
			wantRole:   new(RoleSupport),
		},
		{
			name:       "Archers is support",
			cardName:   "Archers",
			elixirCost: 3,
			wantRole:   new(RoleSupport),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := ClassifyCard(tt.cardName, tt.elixirCost)

			if role == nil {
				if tt.wantRole != nil {
					t.Errorf("ClassifyCard(%v) = nil, want %v", tt.cardName, *tt.wantRole)
				}
				return
			}

			if tt.wantRole == nil {
				t.Errorf("ClassifyCard(%v) = %v, want nil", tt.cardName, *role)
				return
			}

			if *role != *tt.wantRole {
				t.Errorf("ClassifyCard(%v) = %v, want %v", tt.cardName, *role, *tt.wantRole)
			}
		})
	}
}

// TestClassifyCardUnknownReturnsNil verifies unknown cards are not heuristically classified.
func TestClassifyCardUnknownReturnsNil(t *testing.T) {
	tests := []struct {
		name       string
		cardName   string
		elixirCost int
	}{
		{
			name:       "Unknown 1 elixir",
			cardName:   "Unknown Card",
			elixirCost: 1,
		},
		{
			name:       "Unknown 2 elixir",
			cardName:   "Unknown Card",
			elixirCost: 2,
		},
		{
			name:       "Unknown 3 elixir",
			cardName:   "Unknown Card",
			elixirCost: 3,
		},
		{
			name:       "Unknown 4 elixir",
			cardName:   "Unknown Card",
			elixirCost: 4,
		},
		{
			name:       "Unknown 6 elixir",
			cardName:   "Unknown Card",
			elixirCost: 6,
		},
		{
			name:       "Unknown 8 elixir",
			cardName:   "Unknown Card",
			elixirCost: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := ClassifyCard(tt.cardName, tt.elixirCost)
			if role != nil {
				t.Errorf("ClassifyCard(%v, %v) = %v, want nil", tt.cardName, tt.elixirCost, *role)
			}
		})
	}
}

func TestClassifyCardAliases(t *testing.T) {
	role := ClassifyCard("The Log", 2)
	if role == nil {
		t.Fatal("ClassifyCard(The Log) returned nil")
	}

	if *role != RoleSpellSmall {
		t.Fatalf("ClassifyCard(The Log) = %v, want %v", *role, RoleSpellSmall)
	}
}

// TestClassifyCardCandidate tests classifying a CardCandidate
func TestClassifyCardCandidate(t *testing.T) {
	candidate := CardCandidate{
		Name:   "Hog Rider",
		Elixir: 4,
	}

	role := ClassifyCardCandidate(&candidate)

	if role == nil {
		t.Fatal("ClassifyCardCandidate returned nil")
	}

	if *role != RoleWinCondition {
		t.Errorf("ClassifyCardCandidate(Hog Rider) = %v, want WinCondition", *role)
	}

	// Verify role was set on candidate
	if candidate.Role == nil {
		t.Error("ClassifyCardCandidate did not set candidate.Role")
	} else if *candidate.Role != RoleWinCondition {
		t.Errorf("candidate.Role = %v, want WinCondition", *candidate.Role)
	}
}

// TestClassifyAllCandidates tests classifying multiple candidates
func TestClassifyAllCandidates(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "Hog Rider", Elixir: 4},
		{Name: "Cannon", Elixir: 3},
		{Name: "Fireball", Elixir: 4},
		{Name: "Zap", Elixir: 2},
		{Name: "Skeletons", Elixir: 1},
	}

	ClassifyAllCandidates(candidates)

	// Check all were classified
	expectedRoles := []CardRole{
		RoleWinCondition,
		RoleBuilding,
		RoleSpellBig,
		RoleSpellSmall,
		RoleCycle,
	}

	for i, candidate := range candidates {
		if candidate.Role == nil {
			t.Errorf("Candidate %v was not classified", candidate.Name)
			continue
		}

		if *candidate.Role != expectedRoles[i] {
			t.Errorf("Candidate %v role = %v, want %v",
				candidate.Name, *candidate.Role, expectedRoles[i])
		}
	}
}

// TestIsWinCondition tests win condition check
func TestIsWinCondition(t *testing.T) {
	tests := []struct {
		cardName string
		want     bool
	}{
		{"Hog Rider", true},
		{"Royal Giant", true},
		{"Battle Ram", true},
		{"Goblin Barrel", true},
		{"Cannon", false},
		{"Fireball", false},
		{"Skeletons", false},
	}

	for _, tt := range tests {
		t.Run(tt.cardName, func(t *testing.T) {
			got := IsWinCondition(tt.cardName)
			if got != tt.want {
				t.Errorf("IsWinCondition(%v) = %v, want %v", tt.cardName, got, tt.want)
			}
		})
	}
}

// TestIsBuilding tests building check
func TestIsBuilding(t *testing.T) {
	tests := []struct {
		cardName string
		want     bool
	}{
		{"Cannon", true},
		{"Tesla", true},
		{"Goblin Cage", true},
		{"Hog Rider", false},
		{"Fireball", false},
	}

	for _, tt := range tests {
		t.Run(tt.cardName, func(t *testing.T) {
			got := IsBuilding(tt.cardName)
			if got != tt.want {
				t.Errorf("IsBuilding(%v) = %v, want %v", tt.cardName, got, tt.want)
			}
		})
	}
}

// TestIsSpell tests spell check
func TestIsSpell(t *testing.T) {
	tests := []struct {
		cardName   string
		elixirCost int
		want       bool
	}{
		{"Fireball", 4, true},
		{"Zap", 2, true},
		{"Lightning", 6, true},
		{"Arrows", 3, true},
		{"Hog Rider", 4, false},
		{"Cannon", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.cardName, func(t *testing.T) {
			got := IsSpell(tt.cardName, tt.elixirCost)
			if got != tt.want {
				t.Errorf("IsSpell(%v) = %v, want %v", tt.cardName, got, tt.want)
			}
		})
	}
}

// TestCountRoles tests counting roles in a deck
func TestCountRoles(t *testing.T) {
	winCon := RoleWinCondition
	building := RoleBuilding
	spellBig := RoleSpellBig
	spellSmall := RoleSpellSmall
	support := RoleSupport
	cycle := RoleCycle

	candidates := []CardCandidate{
		{Name: "Hog", Role: &winCon},
		{Name: "Cannon", Role: &building},
		{Name: "Fireball", Role: &spellBig},
		{Name: "Zap", Role: &spellSmall},
		{Name: "Musketeer", Role: &support},
		{Name: "Knight", Role: &support},
		{Name: "Skeletons", Role: &cycle},
		{Name: "Fire Spirit", Role: &cycle},
	}

	counts := CountRoles(candidates)

	expected := map[CardRole]int{
		RoleWinCondition: 1,
		RoleBuilding:     1,
		RoleSpellBig:     1,
		RoleSpellSmall:   1,
		RoleSupport:      2,
		RoleCycle:        2,
	}

	for role, expectedCount := range expected {
		if counts[role] != expectedCount {
			t.Errorf("CountRoles[%v] = %v, want %v", role, counts[role], expectedCount)
		}
	}
}

// TestHasBalancedRoles tests deck balance validation
func TestHasBalancedRoles(t *testing.T) {
	winCon := RoleWinCondition
	building := RoleBuilding
	spellBig := RoleSpellBig
	spellSmall := RoleSpellSmall
	support := RoleSupport
	cycle := RoleCycle

	tests := []struct {
		name       string
		candidates []CardCandidate
		wantValid  bool
	}{
		{
			name: "Balanced deck",
			candidates: []CardCandidate{
				{Role: &winCon},
				{Role: &building},
				{Role: &spellBig},
				{Role: &spellSmall},
				{Role: &support},
				{Role: &support},
				{Role: &cycle},
				{Role: &cycle},
			},
			wantValid: true,
		},
		{
			name: "No win condition",
			candidates: []CardCandidate{
				{Role: &building},
				{Role: &spellBig},
				{Role: &spellSmall},
				{Role: &support},
				{Role: &support},
				{Role: &cycle},
				{Role: &cycle},
				{Role: &cycle},
			},
			wantValid: false,
		},
		{
			name: "No small spell",
			candidates: []CardCandidate{
				{Role: &winCon},
				{Role: &building},
				{Role: &spellBig},
				{Role: &spellBig},
				{Role: &support},
				{Role: &support},
				{Role: &cycle},
				{Role: &cycle},
			},
			wantValid: false,
		},
		{
			name: "No cycle cards",
			candidates: []CardCandidate{
				{Role: &winCon},
				{Role: &building},
				{Role: &spellBig},
				{Role: &spellSmall},
				{Role: &support},
				{Role: &support},
				{Role: &support},
				{Role: &support},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := HasBalancedRoles(tt.candidates)
			if valid != tt.wantValid {
				t.Errorf("HasBalancedRoles() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

// TestGetRoleDescription tests getting human-readable role descriptions
func TestGetRoleDescription(t *testing.T) {
	tests := []struct {
		role string
		want string
	}{
		{string(RoleWinCondition), "Primary tower-damaging threat"},
		{string(RoleBuilding), "Defensive building or siege structure"},
		{string(RoleSpellBig), "High-damage spell (4+ elixir)"},
		{string(RoleSpellSmall), "Utility spell (2-3 elixir)"},
		{string(RoleSupport), "Mid-cost support troop"},
		{string(RoleCycle), "Cheap cycle card (1-2 elixir)"},
		{"InvalidRole", "Unknown role"}, // Unknown role falls back
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			desc := GetRoleDescription(CardRole(tt.role))
			if desc != tt.want {
				t.Errorf("GetRoleDescription(%v) = %v, want %v", tt.role, desc, tt.want)
			}
		})
	}
}

// TestClassifyCardWithEvolution tests evolution-aware role classification
func TestClassifyCardWithEvolution(t *testing.T) {
	tests := []struct {
		name           string
		cardName       string
		elixirCost     int
		evolutionLevel int
		wantRole       *CardRole
	}{
		// Evolved Valkyrie - base is support, evolved stays support (override confirms this)
		{
			name:           "Valkyrie unevolved is support",
			cardName:       "Valkyrie",
			elixirCost:     4,
			evolutionLevel: 0,
			wantRole:       new(RoleSupport),
		},
		{
			name:           "Valkyrie evolved is support (override)",
			cardName:       "Valkyrie",
			elixirCost:     4,
			evolutionLevel: 1,
			wantRole:       new(RoleSupport),
		},
		// Evolved Knight - base is support, evolved stays support
		{
			name:           "Knight unevolved is support",
			cardName:       "Knight",
			elixirCost:     3,
			evolutionLevel: 0,
			wantRole:       new(RoleSupport),
		},
		{
			name:           "Knight evolved is support (override)",
			cardName:       "Knight",
			elixirCost:     3,
			evolutionLevel: 1,
			wantRole:       new(RoleSupport),
		},
		// Evolved Royal Giant - base is win condition, evolved stays win condition
		{
			name:           "Royal Giant unevolved is win condition",
			cardName:       "Royal Giant",
			elixirCost:     6,
			evolutionLevel: 0,
			wantRole:       new(RoleWinCondition),
		},
		{
			name:           "Royal Giant evolved is win condition (override)",
			cardName:       "Royal Giant",
			elixirCost:     6,
			evolutionLevel: 1,
			wantRole:       new(RoleWinCondition),
		},
		// Evolved Golem - base is win condition, evolved stays win condition
		{
			name:           "Golem unevolved is win condition",
			cardName:       "Golem",
			elixirCost:     8,
			evolutionLevel: 0,
			wantRole:       new(RoleWinCondition),
		},
		{
			name:           "Golem evolved is win condition (override)",
			cardName:       "Golem",
			elixirCost:     8,
			evolutionLevel: 1,
			wantRole:       new(RoleWinCondition),
		},
		// Card without evolution override - no change when evolved
		{
			name:           "Hog Rider unevolved is win condition",
			cardName:       "Hog Rider",
			elixirCost:     4,
			evolutionLevel: 0,
			wantRole:       new(RoleWinCondition),
		},
		{
			name:           "Hog Rider evolved is still win condition (no override)",
			cardName:       "Hog Rider",
			elixirCost:     4,
			evolutionLevel: 1,
			wantRole:       new(RoleWinCondition),
		},
		// Card without evolution capability
		{
			name:           "Musketeer with evolutionLevel 0 is support",
			cardName:       "Musketeer",
			elixirCost:     4,
			evolutionLevel: 0,
			wantRole:       new(RoleSupport),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := ClassifyCardWithEvolution(tt.cardName, tt.elixirCost, tt.evolutionLevel)

			if role == nil {
				if tt.wantRole != nil {
					t.Errorf("ClassifyCardWithEvolution(%v, evo=%v) = nil, want %v",
						tt.cardName, tt.evolutionLevel, *tt.wantRole)
				}
				return
			}

			if tt.wantRole == nil {
				t.Errorf("ClassifyCardWithEvolution(%v, evo=%v) = %v, want nil",
					tt.cardName, tt.evolutionLevel, *role)
				return
			}

			if *role != *tt.wantRole {
				t.Errorf("ClassifyCardWithEvolution(%v, evo=%v) = %v, want %v",
					tt.cardName, tt.evolutionLevel, *role, *tt.wantRole)
			}
		})
	}
}

// TestClassifyCardCandidateWithEvolution tests classifying a CardCandidate with evolution
func TestClassifyCardCandidateWithEvolution(t *testing.T) {
	tests := []struct {
		name         string
		candidate    CardCandidate
		expectedRole CardRole
	}{
		{
			name: "Evolved Valkyrie candidate",
			candidate: CardCandidate{
				Name:           "Valkyrie",
				Elixir:         4,
				EvolutionLevel: 1,
			},
			expectedRole: RoleSupport,
		},
		{
			name: "Unevolved Valkyrie candidate",
			candidate: CardCandidate{
				Name:           "Valkyrie",
				Elixir:         4,
				EvolutionLevel: 0,
			},
			expectedRole: RoleSupport,
		},
		{
			name: "Evolved Royal Giant candidate",
			candidate: CardCandidate{
				Name:           "Royal Giant",
				Elixir:         6,
				EvolutionLevel: 2,
			},
			expectedRole: RoleWinCondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := ClassifyCardCandidate(&tt.candidate)

			if role == nil {
				t.Fatalf("ClassifyCardCandidate returned nil")
			}

			if *role != tt.expectedRole {
				t.Errorf("ClassifyCardCandidate() role = %v, want %v", *role, tt.expectedRole)
			}

			// Verify role was set on candidate
			if tt.candidate.Role == nil {
				t.Error("ClassifyCardCandidate did not set candidate.Role")
			} else if *tt.candidate.Role != tt.expectedRole {
				t.Errorf("candidate.Role = %v, want %v", *tt.candidate.Role, tt.expectedRole)
			}
		})
	}
}

// TestHasEvolutionOverride tests checking if a card has an evolution override
func TestHasEvolutionOverride(t *testing.T) {
	tests := []struct {
		name           string
		cardName       string
		evolutionLevel int
		want           bool
	}{
		{"Valkyrie evolved has override", "Valkyrie", 1, true},
		{"Valkyrie unevolved no override", "Valkyrie", 0, false},
		{"Knight evolved has override", "Knight", 1, true},
		{"Hog Rider evolved no override", "Hog Rider", 1, false},
		{"Unknown card no override", "Unknown Card", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasEvolutionOverride(tt.cardName, tt.evolutionLevel)
			if got != tt.want {
				t.Errorf("HasEvolutionOverride(%v, evo=%v) = %v, want %v",
					tt.cardName, tt.evolutionLevel, got, tt.want)
			}
		})
	}
}

// TestGetEvolutionOverrideRole tests getting evolution override role
func TestGetEvolutionOverrideRole(t *testing.T) {
	tests := []struct {
		name           string
		cardName       string
		evolutionLevel int
		wantRole       *CardRole
	}{
		{
			name:           "Valkyrie evolved returns support override",
			cardName:       "Valkyrie",
			evolutionLevel: 1,
			wantRole:       new(RoleSupport),
		},
		{
			name:           "Valkyrie unevolved returns nil",
			cardName:       "Valkyrie",
			evolutionLevel: 0,
			wantRole:       nil,
		},
		{
			name:           "Hog Rider evolved returns nil (no override)",
			cardName:       "Hog Rider",
			evolutionLevel: 1,
			wantRole:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEvolutionOverrideRole(tt.cardName, tt.evolutionLevel)

			if got == nil {
				if tt.wantRole != nil {
					t.Errorf("GetEvolutionOverrideRole(%v, evo=%v) = nil, want %v",
						tt.cardName, tt.evolutionLevel, *tt.wantRole)
				}
				return
			}

			if tt.wantRole == nil {
				t.Errorf("GetEvolutionOverrideRole(%v, evo=%v) = %v, want nil",
					tt.cardName, tt.evolutionLevel, *got)
				return
			}

			if *got != *tt.wantRole {
				t.Errorf("GetEvolutionOverrideRole(%v, evo=%v) = %v, want %v",
					tt.cardName, tt.evolutionLevel, *got, *tt.wantRole)
			}
		})
	}
}
