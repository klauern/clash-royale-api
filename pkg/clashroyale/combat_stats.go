package clashroyale

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// CombatStats represents the combat statistics of a card at a specific level (Standard Level 11)
type CombatStats struct {
	Hitpoints       int     `json:"hitpoints,omitempty"`
	Damage          int     `json:"damage,omitempty"`
	DamagePerSecond int     `json:"damagePerSecond,omitempty"`
	HitSpeed        float64 `json:"hitSpeed,omitempty"`
	Speed           string  `json:"speed,omitempty"`   // Slow, Medium, Fast, Very Fast
	Targets         string  `json:"targets,omitempty"` // Ground, Air, Air & Ground, Buildings
	Range           float64 `json:"range,omitempty"`   // Range in tiles
	Radius          float64 `json:"radius,omitempty"`  // Area of effect radius
	Lifetime        float64 `json:"lifetime,omitempty"`
	SpawnCount      int     `json:"spawnCount,omitempty"`
	DeathDamage     int     `json:"deathDamage,omitempty"`
	DashDamage      int     `json:"dashDamage,omitempty"`
}

// CardStatsRegistry represents the collection of all card stats
type CardStatsRegistry struct {
	// Map of Card Name -> CombatStats
	Stats map[string]CombatStats `json:"stats"`
}

// LoadStats loads combat stats from a JSON file
func LoadStats(filepath string) (*CardStatsRegistry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stats file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats file: %w", err)
	}

	var registry CardStatsRegistry
	if err := json.Unmarshal(bytes, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse stats file: %w", err)
	}

	return &registry, nil
}

// GetStats returns the stats for a given card name
func (r *CardStatsRegistry) GetStats(cardName string) *CombatStats {
	if stats, ok := r.Stats[cardName]; ok {
		return &stats
	}
	return nil
}

// DPSPerElixir calculates damage efficiency for elixir cost
func (cs *CombatStats) DPSPerElixir(elixir int) float64 {
	if elixir <= 0 {
		return 0
	}
	return float64(cs.DamagePerSecond) / float64(elixir)
}

// HPPerElixir calculates survivability efficiency for elixir cost
func (cs *CombatStats) HPPerElixir(elixir int) float64 {
	if elixir <= 0 {
		return 0
	}
	return float64(cs.Hitpoints) / float64(elixir)
}

// RangeEffectiveness scores range advantage (0-1 normalized)
// Higher range is generally better, with diminishing returns beyond 7 tiles
func (cs *CombatStats) RangeEffectiveness() float64 {
	if cs.Range <= 0 {
		return 0 // Melee units get 0 range effectiveness
	}
	// Normalize range effectiveness: 0-1 scale, with 7 tiles as excellent
	return math.Min(cs.Range/7.0, 1.0)
}

// TargetCoverage scores target flexibility (Air&Ground > single target)
func (cs *CombatStats) TargetCoverage() float64 {
	switch strings.ToLower(cs.Targets) {
	case "air & ground":
		return 1.0 // Best coverage
	case "ground":
		return 0.7 // Ground-only is decent
	case "air":
		return 0.6 // Air-only is more limited
	case "buildings":
		return 0.4 // Building-only is situational
	default:
		return 0.5 // Unknown/other targets
	}
}

// SpeedEffectiveness scores movement speed advantage (0-1 normalized)
func (cs *CombatStats) SpeedEffectiveness() float64 {
	switch strings.ToLower(cs.Speed) {
	case "very fast":
		return 1.0
	case "fast":
		return 0.8
	case "medium":
		return 0.6
	case "slow":
		return 0.4
	default:
		return 0.5 // Unknown speed
	}
}

// RoleSpecificEffectiveness evaluates combat effectiveness by card role
// Accepts role as string to avoid circular import with deck package
func (cs *CombatStats) RoleSpecificEffectiveness(role string) float64 {
	if role == "" {
		return 0.5 // Neutral if no role defined
	}

	var effectiveness float64

	switch strings.ToLower(role) {
	case "wincondition":
		// Win conditions value high HP and good damage
		hpScore := math.Min(float64(cs.Hitpoints)/3000.0, 1.0) // Normalize to ~3000 HP max
		damageScore := math.Min(float64(cs.Damage)/500.0, 1.0) // Normalize to ~500 damage max
		effectiveness = (hpScore*0.6 + damageScore*0.4)

	case "building":
		// Buildings value high HP and lifetime
		hpScore := math.Min(float64(cs.Hitpoints)/2000.0, 1.0)
		lifetimeScore := math.Min(cs.Lifetime/60.0, 1.0) // 60 seconds as reference
		effectiveness = (hpScore*0.7 + lifetimeScore*0.3)

	case "support":
		// Support values range and target flexibility
		rangeScore := cs.RangeEffectiveness()
		targetScore := cs.TargetCoverage()
		effectiveness = (rangeScore*0.6 + targetScore*0.4)

	case "spell":
		// Spells value damage and radius
		damageScore := math.Min(float64(cs.Damage)/600.0, 1.0)
		radiusScore := math.Min(cs.Radius/5.0, 1.0) // 5 tiles as reference radius
		effectiveness = (damageScore*0.7 + radiusScore*0.3)

	case "cycle":
		// Cycle cards value cost efficiency and speed
		// Note: we don't have elixir here, so focus on damage/speed ratio
		damageSpeedRatio := float64(cs.Damage) * cs.SpeedEffectiveness()
		effectiveness = math.Min(damageSpeedRatio/300.0, 1.0)

	default:
		effectiveness = 0.5
	}

	return math.Max(0, math.Min(1, effectiveness)) // Clamp to 0-1 range
}

// CombatEffectiveness returns normalized 0-1 score combining multiple combat factors
func (cs *CombatStats) CombatEffectiveness(elixir int) float64 {
	if elixir <= 0 {
		return 0
	}

	// Weight different factors based on card type
	dpsEfficiency := cs.DPSPerElixir(elixir)
	hpEfficiency := cs.HPPerElixir(elixir)
	rangeScore := cs.RangeEffectiveness()
	targetScore := cs.TargetCoverage()
	speedScore := cs.SpeedEffectiveness()

	// Normalize efficiency scores (rough estimates for balance)
	dpsNormalized := math.Min(dpsEfficiency/50.0, 1.0) // ~50 DPS/elixir as excellent
	hpNormalized := math.Min(hpEfficiency/400.0, 1.0)  // ~400 HP/elixir as excellent

	// Combine factors with weights
	composite := (dpsNormalized * 0.3) + // Damage output
		(hpNormalized * 0.25) + // Survivability
		(rangeScore * 0.15) + // Range advantage
		(targetScore * 0.15) + // Target flexibility
		(speedScore * 0.15) // Mobility

	return math.Max(0, math.Min(1, composite)) // Clamp to 0-1 range
}

// StatsSummary returns a human-readable summary of key stats
func (cs *CombatStats) StatsSummary() string {
	var parts []string

	if cs.Hitpoints > 0 {
		parts = append(parts, fmt.Sprintf("HP: %d", cs.Hitpoints))
	}
	if cs.Damage > 0 {
		parts = append(parts, fmt.Sprintf("DMG: %d", cs.Damage))
	}
	if cs.DamagePerSecond > 0 {
		parts = append(parts, fmt.Sprintf("DPS: %d", cs.DamagePerSecond))
	}
	if cs.Range > 0 {
		parts = append(parts, fmt.Sprintf("Range: %.1f", cs.Range))
	}
	if cs.Targets != "" {
		parts = append(parts, fmt.Sprintf("Targets: %s", cs.Targets))
	}

	return strings.Join(parts, ", ")
}
