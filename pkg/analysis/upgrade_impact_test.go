package analysis

import "testing"

func TestInferRoleUsesSharedConfigRoleMapping(t *testing.T) {
	analyzer := &UpgradeImpactAnalyzer{}

	tests := []struct {
		name     string
		cardName string
		wantRole string
	}{
		{name: "win condition", cardName: "Hog Rider", wantRole: "win_conditions"},
		{name: "building", cardName: "Cannon", wantRole: "buildings"},
		{name: "big spell", cardName: "Fireball", wantRole: "spells_big"},
		{name: "small spell", cardName: "Zap", wantRole: "spells_small"},
		{name: "alias spell", cardName: "The Log", wantRole: "spells_small"},
		{name: "dual-role precedence", cardName: "Heal Spirit", wantRole: "spells_small"},
		{name: "legacy win-condition override", cardName: "Goblin Drill", wantRole: "win_conditions"},
		{name: "support", cardName: "Musketeer", wantRole: "support"},
		{name: "unknown defaults to support", cardName: "Unknown Card", wantRole: "support"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.inferRole(tt.cardName)
			if got != tt.wantRole {
				t.Errorf("inferRole(%q) = %q, want %q", tt.cardName, got, tt.wantRole)
			}
		})
	}
}

func TestGetRoleImportanceByInferredRole(t *testing.T) {
	analyzer := &UpgradeImpactAnalyzer{}

	tests := []struct {
		name       string
		cardName   string
		importance float64
	}{
		{name: "win condition", cardName: "Hog Rider", importance: 1.0},
		{name: "building", cardName: "Cannon", importance: 0.7},
		{name: "big spell", cardName: "Fireball", importance: 0.6},
		{name: "small spell", cardName: "Zap", importance: 0.5},
		{name: "cycle", cardName: "Skeletons", importance: 0.4},
		{name: "support", cardName: "Musketeer", importance: 0.4},
		{name: "unknown defaults to support", cardName: "Unknown Card", importance: 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.getRoleImportance(tt.cardName)
			if got != tt.importance {
				t.Errorf("getRoleImportance(%q) = %v, want %v", tt.cardName, got, tt.importance)
			}
		})
	}
}
