package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// SynergyCategory defines common synergy patterns between cards
type SynergyCategory string

const (
	SynergyTankSupport  SynergyCategory = "tank_support"  // Tank + Support troops
	SynergyBait         SynergyCategory = "bait"          // Bait combos (log bait, zap bait)
	SynergySpellCombo   SynergyCategory = "spell_combo"   // Spell combinations
	SynergyWinCondition SynergyCategory = "win_condition" // Win condition combos
	SynergyDefensive    SynergyCategory = "defensive"     // Defensive combinations
	SynergyCycle        SynergyCategory = "cycle"         // Cycle card combinations
	SynergyBridgeSpam   SynergyCategory = "bridge_spam"   // Bridge spam combinations
)

// SynergyPair represents synergy between two cards
type SynergyPair struct {
	Card1       string          `json:"card1"`
	Card2       string          `json:"card2"`
	SynergyType SynergyCategory `json:"synergy_type"`
	Score       float64         `json:"score"` // 0.0 to 1.0
	Description string          `json:"description"`
}

// SynergyDatabase holds known card synergies
type SynergyDatabase struct {
	Pairs      []SynergyPair                     `json:"pairs"`
	Categories map[SynergyCategory][]SynergyPair `json:"categories"`
}

// DeckSynergyAnalysis represents the synergy analysis of a deck
type DeckSynergyAnalysis struct {
	TotalScore       float64                 `json:"total_score"`       // 0-100
	AverageScore     float64                 `json:"average_score"`     // Average synergy per pair
	TopSynergies     []SynergyPair           `json:"top_synergies"`     // Best synergies in deck
	MissingSynergies []string                `json:"missing_synergies"` // Cards that don't synergize well
	CategoryScores   map[SynergyCategory]int `json:"category_scores"`   // Count by category
}

// SynergyRecommendation suggests a card to add based on synergies
type SynergyRecommendation struct {
	CardName     string        `json:"card_name"`
	SynergyScore float64       `json:"synergy_score"`
	Synergies    []SynergyPair `json:"synergies"` // Synergies with current deck
	Reason       string        `json:"reason"`
}

// synergyDataFile represents the JSON data structure for loading synergy pairs
type synergyDataFile struct {
	Version     int           `json:"version"`
	Description string        `json:"description"`
	LastUpdated string        `json:"last_updated"`
	Pairs       []SynergyPair `json:"pairs"`
}

// LoadSynergyDatabase loads synergy pairs from a JSON file
// If the file cannot be found or read, falls back to NewSynergyDatabase()
func LoadSynergyDatabase(dataDir, filename string) *SynergyDatabase {
	if dataDir == "" {
		dataDir = "data"
	}
	if filename == "" {
		filename = "synergy_pairs.json"
	}

	path := filepath.Join(dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		// Fall back to hardcoded database if file not found
		return NewSynergyDatabase()
	}

	var synergyData synergyDataFile
	if err := json.Unmarshal(data, &synergyData); err != nil {
		// Fall back to hardcoded database on parse error
		return NewSynergyDatabase()
	}

	// Organize by category
	categories := make(map[SynergyCategory][]SynergyPair)
	for _, pair := range synergyData.Pairs {
		categories[pair.SynergyType] = append(categories[pair.SynergyType], pair)
	}

	return buildSynergyDatabase(synergyData.Pairs)
}

// buildCategoryMap organizes synergy pairs by category type
func buildSynergyDatabase(pairs []SynergyPair) *SynergyDatabase {
	categories := make(map[SynergyCategory][]SynergyPair)
	for _, pair := range pairs {
		categories[pair.SynergyType] = append(categories[pair.SynergyType], pair)
	}

	return &SynergyDatabase{
		Pairs:      pairs,
		Categories: categories,
	}
}

// NewSynergyDatabase creates a synergy database with known card combinations
func NewSynergyDatabase() *SynergyDatabase {
	pairs := []SynergyPair{
		// Tank + Support synergies
		{Card1: "Giant", Card2: "Witch", SynergyType: SynergyTankSupport, Score: 0.9, Description: "Witch supports Giant with splash damage and spawns"},
		{Card1: "Giant", Card2: "Sparky", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Giant tanks while Sparky deals massive damage"},
		{Card1: "Giant", Card2: "Musketeer", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Musketeer provides ranged support behind Giant"},
		{Card1: "Giant", Card2: "Dark Prince", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Dark Prince provides splash support and charging damage"},
		{Card1: "Giant", Card2: "Mini P.E.K.K.A", SynergyType: SynergyTankSupport, Score: 0.75, Description: "Mini PEKKA defends then supports Giant counter-push"},
		{Card1: "Golem", Card2: "Night Witch", SynergyType: SynergyTankSupport, Score: 0.95, Description: "Classic Golem beatdown synergy"},
		{Card1: "Golem", Card2: "Baby Dragon", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Baby Dragon provides splash support"},
		{Card1: "Golem", Card2: "Lumberjack", SynergyType: SynergyTankSupport, Score: 0.9, Description: "Lumberjack provides rage and fast clearing"},
		{Card1: "Lava Hound", Card2: "Balloon", SynergyType: SynergyWinCondition, Score: 0.95, Description: "LavaLoon: overwhelming air pressure"},
		{Card1: "Lava Hound", Card2: "Miner", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Miner supports Lava Hound pups"},
		{Card1: "Lava Hound", Card2: "Mega Minion", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Mega Minion provides strong air support"},
		{Card1: "Lava Hound", Card2: "Skeleton Dragons", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Skeleton Dragons provide splash air support"},
		{Card1: "Mega Knight", Card2: "Bats", SynergyType: SynergyTankSupport, Score: 0.75, Description: "Bats provide fast swarm defense"},
		{Card1: "Mega Knight", Card2: "Inferno Dragon", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Inferno Dragon handles tanks while MK defends"},
		{Card1: "Mega Knight", Card2: "Minions", SynergyType: SynergyTankSupport, Score: 0.75, Description: "Minions provide air support for MK"},
		{Card1: "Mega Knight", Card2: "Electro Wizard", SynergyType: SynergyTankSupport, Score: 0.85, Description: "E-Wiz provides reset and ranged support"},
		{Card1: "Mega Knight", Card2: "Goblin Gang", SynergyType: SynergyTankSupport, Score: 0.7, Description: "Goblin Gang provides defensive bait value"},
		{Card1: "Electro Giant", Card2: "Tornado", SynergyType: SynergyTankSupport, Score: 0.9, Description: "Tornado groups enemies for E-Giant zaps"},
		{Card1: "Electro Giant", Card2: "Heal Spirit", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Heal Spirit sustains E-Giant push"},
		{Card1: "Electro Giant", Card2: "Mother Witch", SynergyType: SynergyTankSupport, Score: 0.85, Description: "Mother Witch converts swarms to hogs"},
		{Card1: "Electro Giant", Card2: "Dark Prince", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Dark Prince provides splash and charging support"},
		{Card1: "P.E.K.K.A", Card2: "Electro Wizard", SynergyType: SynergyTankSupport, Score: 0.85, Description: "E-Wiz provides reset and support for PEKKA"},
		{Card1: "P.E.K.K.A", Card2: "Magic Archer", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Magic Archer provides ranged piercing support"},
		{Card1: "P.E.K.K.A", Card2: "Dark Prince", SynergyType: SynergyTankSupport, Score: 0.8, Description: "Dark Prince provides splash support"},

		// Bait synergies
		{Card1: "Goblin Barrel", Card2: "Princess", SynergyType: SynergyBait, Score: 0.95, Description: "Log bait: Princess baits log for Goblin Barrel"},
		{Card1: "Goblin Barrel", Card2: "Goblin Gang", SynergyType: SynergyBait, Score: 0.9, Description: "Multiple goblin threats overwhelm spells"},
		{Card1: "Goblin Barrel", Card2: "Dart Goblin", SynergyType: SynergyBait, Score: 0.85, Description: "Dart Goblin baits small spells"},
		{Card1: "Goblin Barrel", Card2: "Skeleton Army", SynergyType: SynergyBait, Score: 0.85, Description: "Swarm bait forces spell usage"},
		{Card1: "Goblin Barrel", Card2: "Inferno Tower", SynergyType: SynergyBait, Score: 0.75, Description: "Building bait punishes spell usage"},
		{Card1: "Skeleton Barrel", Card2: "Goblin Barrel", SynergyType: SynergyBait, Score: 0.8, Description: "Double barrel pressure"},
		{Card1: "Princess", Card2: "Goblin Gang", SynergyType: SynergyBait, Score: 0.85, Description: "Log bait pressure"},
		{Card1: "Princess", Card2: "Dart Goblin", SynergyType: SynergyBait, Score: 0.85, Description: "Dual log bait threats"},
		{Card1: "Graveyard", Card2: "Skeleton Army", SynergyType: SynergyBait, Score: 0.75, Description: "Skeleton flood overwhelms single spells"},
		{Card1: "Graveyard", Card2: "Tombstone", SynergyType: SynergyBait, Score: 0.8, Description: "Continuous skeleton pressure"},
		{Card1: "Skeleton Army", Card2: "Goblin Gang", SynergyType: SynergyBait, Score: 0.8, Description: "Dual swarm bait"},
		{Card1: "Bats", Card2: "Minions", SynergyType: SynergyBait, Score: 0.75, Description: "Zap bait flying swarms"},
		{Card1: "Spear Goblins", Card2: "Goblins", SynergyType: SynergyBait, Score: 0.7, Description: "Small spell bait pressure"},
		{Card1: "Goblin Hut", Card2: "Furnace", SynergyType: SynergyBait, Score: 0.75, Description: "Building spam forces spell usage"},
		{Card1: "X-Bow", Card2: "Tesla", SynergyType: SynergyBait, Score: 0.9, Description: "Double building bait and defense"},

		// Spell combos
		{Card1: "Hog Rider", Card2: "Fireball", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Fireball clears defenders for Hog"},
		{Card1: "Hog Rider", Card2: "Earthquake", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Earthquake destroys buildings for Hog"},
		{Card1: "Hog Rider", Card2: "Freeze", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Freeze guarantees Hog tower damage"},
		{Card1: "Tornado", Card2: "Fireball", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Tornado groups troops for Fireball"},
		{Card1: "Tornado", Card2: "Rocket", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Tornado + Rocket tower finish"},
		{Card1: "Tornado", Card2: "Executioner", SynergyType: SynergySpellCombo, Score: 0.9, Description: "Tornado pulls troops into Executioner's axe"},
		{Card1: "Tornado", Card2: "Bowler", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Tornado + Bowler knockback combo"},
		{Card1: "Tornado", Card2: "Ice Wizard", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Tornado groups for Ice Wizard slow"},
		{Card1: "Tornado", Card2: "Baby Dragon", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Tornado pulls troops for Baby Dragon splash"},
		{Card1: "Graveyard", Card2: "Freeze", SynergyType: SynergySpellCombo, Score: 0.9, Description: "Freeze allows Graveyard skeletons to connect"},
		{Card1: "Graveyard", Card2: "Poison", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Poison clears small troops from Graveyard"},
		{Card1: "Poison", Card2: "Miner", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Poison + Miner chip damage combo"},
		{Card1: "Earthquake", Card2: "Royal Giant", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Earthquake removes buildings for RG"},
		{Card1: "Earthquake", Card2: "Miner", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Earthquake clears buildings for Miner"},
		{Card1: "Freeze", Card2: "Balloon", SynergyType: SynergySpellCombo, Score: 0.9, Description: "Freeze guarantees Balloon connection"},
		{Card1: "Rage", Card2: "Lumberjack", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Double rage acceleration"},
		{Card1: "Rage", Card2: "Balloon", SynergyType: SynergySpellCombo, Score: 0.85, Description: "Rage accelerates Balloon to tower"},
		{Card1: "Rage", Card2: "Elite Barbarians", SynergyType: SynergySpellCombo, Score: 0.8, Description: "Rage boosts E-Barbs speed and DPS"},

		// Bridge spam
		{Card1: "P.E.K.K.A", Card2: "Battle Ram", SynergyType: SynergyBridgeSpam, Score: 0.85, Description: "PEKKA Bridge Spam pressure"},
		{Card1: "P.E.K.K.A", Card2: "Bandit", SynergyType: SynergyBridgeSpam, Score: 0.8, Description: "Bandit supports PEKKA counterpush"},
		{Card1: "Battle Ram", Card2: "Bandit", SynergyType: SynergyBridgeSpam, Score: 0.8, Description: "Fast dual-lane pressure"},
		{Card1: "Battle Ram", Card2: "Minions", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "Air support for Battle Ram push"},
		{Card1: "Battle Ram", Card2: "Dark Prince", SynergyType: SynergyBridgeSpam, Score: 0.85, Description: "Dual charge pressure"},
		{Card1: "Bandit", Card2: "Royal Ghost", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "Invisible bridge spam"},
		{Card1: "Bandit", Card2: "Magic Archer", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "Bandit dash with Magic Archer support"},
		{Card1: "Bandit", Card2: "Electro Wizard", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "E-Wiz support for Bandit"},
		{Card1: "Royal Ghost", Card2: "Dark Prince", SynergyType: SynergyBridgeSpam, Score: 0.75, Description: "Dual invisible pressure"},
		{Card1: "Royal Ghost", Card2: "Minions", SynergyType: SynergyBridgeSpam, Score: 0.7, Description: "Air support for Ghost push"},
		{Card1: "Lumberjack", Card2: "Balloon", SynergyType: SynergyBridgeSpam, Score: 0.95, Description: "LumberLoon: Rage boost for Balloon"},

		// Defensive synergies
		{Card1: "Cannon", Card2: "Ice Spirit", SynergyType: SynergyDefensive, Score: 0.8, Description: "Cheap defensive combo"},
		{Card1: "Cannon", Card2: "Knight", SynergyType: SynergyDefensive, Score: 0.8, Description: "Knight + Cannon cheap defense"},
		{Card1: "Tesla", Card2: "Ice Spirit", SynergyType: SynergyDefensive, Score: 0.75, Description: "Tesla + Ice Spirit kiting"},
		{Card1: "Tesla", Card2: "Tornado", SynergyType: SynergyDefensive, Score: 0.85, Description: "Tornado pulls troops to Tesla"},
		{Card1: "Inferno Tower", Card2: "Zap", SynergyType: SynergyDefensive, Score: 0.85, Description: "Zap resets for Inferno Tower"},
		{Card1: "Inferno Tower", Card2: "Tornado", SynergyType: SynergyDefensive, Score: 0.9, Description: "Tornado pulls tanks to Inferno"},
		{Card1: "Inferno Dragon", Card2: "Zap", SynergyType: SynergyDefensive, Score: 0.8, Description: "Zap protects Inferno Dragon beam"},
		{Card1: "Bomb Tower", Card2: "Valkyrie", SynergyType: SynergyDefensive, Score: 0.75, Description: "Dual splash defensive combo"},
		{Card1: "Goblin Cage", Card2: "Guards", SynergyType: SynergyDefensive, Score: 0.7, Description: "Defensive troops chain"},
		{Card1: "Mega Minion", Card2: "Bats", SynergyType: SynergyDefensive, Score: 0.75, Description: "Air defense combo"},
		{Card1: "Musketeer", Card2: "Ice Spirit", SynergyType: SynergyDefensive, Score: 0.75, Description: "Musketeer + freeze for air defense"},
		{Card1: "Hunter", Card2: "Tornado", SynergyType: SynergyDefensive, Score: 0.85, Description: "Tornado groups for Hunter burst"},
		{Card1: "Electro Wizard", Card2: "Mega Minion", SynergyType: SynergyDefensive, Score: 0.75, Description: "E-Wiz reset + air defense"},

		// Cycle synergies
		{Card1: "Ice Spirit", Card2: "Skeletons", SynergyType: SynergyCycle, Score: 0.85, Description: "Ultra-cheap cycle combo"},
		{Card1: "Ice Spirit", Card2: "Fire Spirit", SynergyType: SynergyCycle, Score: 0.8, Description: "Cheap spirit cycle"},
		{Card1: "Ice Spirit", Card2: "Spear Goblins", SynergyType: SynergyCycle, Score: 0.75, Description: "Fast cycle defensive combo"},
		{Card1: "Ice Spirit", Card2: "Bats", SynergyType: SynergyCycle, Score: 0.75, Description: "Ultra-cheap air cycle"},
		{Card1: "Ice Spirit", Card2: "Log", SynergyType: SynergyCycle, Score: 0.8, Description: "Cheap cycle and control"},
		{Card1: "Skeletons", Card2: "Goblins", SynergyType: SynergyCycle, Score: 0.8, Description: "Fast cycle swarm combo"},
		{Card1: "Skeletons", Card2: "Ice Golem", SynergyType: SynergyCycle, Score: 0.8, Description: "Cheap cycle tank"},
		{Card1: "Skeletons", Card2: "Log", SynergyType: SynergyCycle, Score: 0.75, Description: "Cycle and clear combo"},
		{Card1: "Fire Spirit", Card2: "Heal Spirit", SynergyType: SynergyCycle, Score: 0.75, Description: "Dual spirit cycle"},
		{Card1: "Fire Spirit", Card2: "Goblins", SynergyType: SynergyCycle, Score: 0.7, Description: "Fast rotation combo"},
		{Card1: "Heal Spirit", Card2: "Skeletons", SynergyType: SynergyCycle, Score: 0.75, Description: "Ultra-fast cycle"},

		// Win condition synergies
		{Card1: "Hog Rider", Card2: "Valkyrie", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Valkyrie tanks and clears for Hog"},
		{Card1: "Hog Rider", Card2: "Ice Golem", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Ice Golem kites and tanks for Hog"},
		{Card1: "Hog Rider", Card2: "Musketeer", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Musketeer supports Hog push"},
		{Card1: "Royal Giant", Card2: "Fisherman", SynergyType: SynergyWinCondition, Score: 0.85, Description: "Fisherman activates King Tower for RG"},
		{Card1: "Royal Giant", Card2: "Lightning", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Lightning clears defensive buildings"},
		{Card1: "Royal Giant", Card2: "Hunter", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Hunter provides defensive synergy"},
		{Card1: "X-Bow", Card2: "Tesla", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Double building lock"},
		{Card1: "X-Bow", Card2: "Archers", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Archers defend X-Bow"},
		{Card1: "X-Bow", Card2: "Ice Golem", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Ice Golem kites for X-Bow defense"},
		{Card1: "Mortar", Card2: "Cannon", SynergyType: SynergyWinCondition, Score: 0.85, Description: "Mortar + defensive building"},
		{Card1: "Mortar", Card2: "Knight", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Knight tanks and defends for Mortar"},
		{Card1: "Mortar", Card2: "Archers", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Archers support Mortar defense"},
		{Card1: "Mortar", Card2: "Skeletons", SynergyType: SynergyWinCondition, Score: 0.7, Description: "Skeletons cycle and defend"},
		{Card1: "Miner", Card2: "Balloon", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Miner tanks for Balloon"},
		{Card1: "Miner", Card2: "Goblin Barrel", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Dual win condition pressure"},
		{Card1: "Miner", Card2: "Wall Breakers", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Dual tower pressure"},
		{Card1: "Miner", Card2: "Skeleton Barrel", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Dual air pressure"},
		{Card1: "Ram Rider", Card2: "P.E.K.K.A", SynergyType: SynergyWinCondition, Score: 0.8, Description: "PEKKA supports Ram Rider push"},
		{Card1: "Ram Rider", Card2: "Mega Knight", SynergyType: SynergyWinCondition, Score: 0.75, Description: "MK defends then Ram counterpush"},
		{Card1: "Royal Hogs", Card2: "Earthquake", SynergyType: SynergyWinCondition, Score: 0.85, Description: "Earthquake clears buildings for Royal Hogs"},
		{Card1: "Royal Hogs", Card2: "Fisherman", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Fisherman pulls defenders away"},
		{Card1: "Wall Breakers", Card2: "Giant", SynergyType: SynergyWinCondition, Score: 0.75, Description: "Dual tower threat pressure"},
		{Card1: "Sparky", Card2: "Giant", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Giant tanks while Sparky charges"},
		{Card1: "Sparky", Card2: "Goblin Giant", SynergyType: SynergyWinCondition, Score: 0.9, Description: "Goblin Giant tanks with spear support"},
		{Card1: "Sparky", Card2: "Tornado", SynergyType: SynergyWinCondition, Score: 0.85, Description: "Tornado groups enemies for Sparky"},
		{Card1: "Three Musketeers", Card2: "Battle Ram", SynergyType: SynergyWinCondition, Score: 0.9, Description: "3M split with Battle Ram pressure"},
		{Card1: "Three Musketeers", Card2: "Ice Golem", SynergyType: SynergyWinCondition, Score: 0.8, Description: "Ice Golem tanks for 3M split"},
	}

	return buildSynergyDatabase(pairs)
}

// GetSynergy returns the synergy score between two cards (0.0 to 1.0)
// Returns 0 if no known synergy exists
func (db *SynergyDatabase) GetSynergy(card1, card2 string) float64 {
	for _, pair := range db.Pairs {
		if (pair.Card1 == card1 && pair.Card2 == card2) ||
			(pair.Card1 == card2 && pair.Card2 == card1) {
			return pair.Score
		}
	}
	return 0.0
}

// GetSynergyPair returns the synergy pair details if it exists
func (db *SynergyDatabase) GetSynergyPair(card1, card2 string) *SynergyPair {
	for _, pair := range db.Pairs {
		if (pair.Card1 == card1 && pair.Card2 == card2) ||
			(pair.Card1 == card2 && pair.Card2 == card1) {
			return &pair
		}
	}
	return nil
}

// AnalyzeDeckSynergy scores overall deck synergy
func (db *SynergyDatabase) AnalyzeDeckSynergy(deck []string) *DeckSynergyAnalysis {
	if len(deck) == 0 {
		return &DeckSynergyAnalysis{}
	}

	topSynergies := make([]SynergyPair, 0)
	totalScore := 0.0
	pairCount := 0
	categoryScores := make(map[SynergyCategory]int)
	cardSynergyCounts := make(map[string]int)

	// Check all pairs
	for i := 0; i < len(deck); i++ {
		for j := i + 1; j < len(deck); j++ {
			if pair := db.GetSynergyPair(deck[i], deck[j]); pair != nil {
				topSynergies = append(topSynergies, *pair)
				totalScore += pair.Score
				pairCount++
				categoryScores[pair.SynergyType]++
				cardSynergyCounts[deck[i]]++
				cardSynergyCounts[deck[j]]++
			}
		}
	}

	// Sort top synergies by score
	sort.Slice(topSynergies, func(i, j int) bool {
		return topSynergies[i].Score > topSynergies[j].Score
	})

	// Limit to top 5
	if len(topSynergies) > 5 {
		topSynergies = topSynergies[:5]
	}

	// Find cards with no synergies
	missingSynergies := make([]string, 0)
	for _, card := range deck {
		if cardSynergyCounts[card] == 0 {
			missingSynergies = append(missingSynergies, card)
		}
	}

	// Calculate average
	avgScore := 0.0
	if pairCount > 0 {
		avgScore = totalScore / float64(pairCount)
	}

	// Normalize total score to 0-100
	// Maximum possible score with 8 cards = 28 pairs * 1.0 = 28
	// Scale to 100
	normalizedTotal := (totalScore / 28.0) * 100.0

	return &DeckSynergyAnalysis{
		TotalScore:       normalizedTotal,
		AverageScore:     avgScore,
		TopSynergies:     topSynergies,
		MissingSynergies: missingSynergies,
		CategoryScores:   categoryScores,
	}
}

// SuggestSynergyCards recommends cards that synergize with the current deck
func (db *SynergyDatabase) SuggestSynergyCards(currentDeck []string, available []*CardCandidate) []*SynergyRecommendation {
	if len(currentDeck) == 0 || len(available) == 0 {
		return nil
	}

	// Create map for quick lookup
	inDeck := make(map[string]bool)
	for _, card := range currentDeck {
		inDeck[card] = true
	}

	// Score each available card by its synergies with current deck
	recommendations := make(map[string]*SynergyRecommendation)

	for _, candidate := range available {
		// Skip cards already in deck
		if inDeck[candidate.Name] {
			continue
		}

		synergies := make([]SynergyPair, 0)
		totalSynergy := 0.0

		// Check synergies with each card in deck
		for _, deckCard := range currentDeck {
			if pair := db.GetSynergyPair(candidate.Name, deckCard); pair != nil {
				synergies = append(synergies, *pair)
				totalSynergy += pair.Score
			}
		}

		// Only recommend if has synergies
		if len(synergies) > 0 {
			avgSynergy := totalSynergy / float64(len(synergies))

			reason := fmt.Sprintf("Synergizes with %d cards in your deck", len(synergies))
			if len(synergies) >= 3 {
				reason = fmt.Sprintf("Strong synergies with %d cards: %s", len(synergies), synergies[0].Card2)
			}

			recommendations[candidate.Name] = &SynergyRecommendation{
				CardName:     candidate.Name,
				SynergyScore: avgSynergy,
				Synergies:    synergies,
				Reason:       reason,
			}
		}
	}

	// Convert to slice and sort by score
	result := make([]*SynergyRecommendation, 0, len(recommendations))
	for _, rec := range recommendations {
		result = append(result, rec)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SynergyScore > result[j].SynergyScore
	})

	// Return top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// GetCategoryDescription returns a human-readable description of a synergy category
func GetCategoryDescription(category SynergyCategory) string {
	descriptions := map[SynergyCategory]string{
		SynergyTankSupport:  "Tank + Support",
		SynergyBait:         "Spell Bait",
		SynergySpellCombo:   "Spell Combo",
		SynergyWinCondition: "Win Condition",
		SynergyDefensive:    "Defensive",
		SynergyCycle:        "Cycle",
		SynergyBridgeSpam:   "Bridge Spam",
	}
	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return string(category)
}

// CalculateDeckSynergy calculates the synergy score for a deck.
// This is a convenience wrapper for AnalyzeDeckSynergy that matches
// the naming convention specified in the improved scoring design.
func (db *SynergyDatabase) CalculateDeckSynergy(deck []string) *DeckSynergyAnalysis {
	return db.AnalyzeDeckSynergy(deck)
}
