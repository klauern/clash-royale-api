package config

import "testing"

func TestGetRoleDescription(t *testing.T) {
	tests := []struct {
		name string
		role CardRole
		want string
	}{
		{name: "win condition", role: RoleWinCondition, want: "Primary tower-damaging threat"},
		{name: "building", role: RoleBuilding, want: "Defensive building or siege structure"},
		{name: "big spell", role: RoleSpellBig, want: "High-damage spell (4+ elixir)"},
		{name: "small spell", role: RoleSpellSmall, want: "Utility spell (2-3 elixir)"},
		{name: "support", role: RoleSupport, want: "Mid-cost support troop"},
		{name: "cycle", role: RoleCycle, want: "Cheap cycle card (1-2 elixir)"},
		{name: "unknown", role: CardRole("invalid"), want: "Unknown role"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRoleDescription(tt.role); got != tt.want {
				t.Fatalf("GetRoleDescription(%q) = %q, want %q", tt.role, got, tt.want)
			}
		})
	}
}
