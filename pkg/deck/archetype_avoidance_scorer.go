package deck

import (
	"strings"
)

// archetypeCardAssociations maps archetype names to their strongly associated cards
// This is derived from archetype constraints but kept here to avoid import cycles
var archetypeCardAssociations = map[string][]string{
	"beatdown": {
		"Golem", "Giant", "Lava Hound", "Electro Giant",
		"Baby Dragon", "Night Witch", "Mega Minion", "Lumberjack",
		"Lightning", "Tornado", "Arrows",
	},
	"cycle": {
		"Hog Rider", "Miner", "Skeletons", "Ice Spirit",
		"Ice Golem", "Cannon", "Musketeer", "Log",
		"Fireball", "Electro Spirit", "Bats",
	},
	"control": {
		"Inferno Tower", "Cannon", "Bomb Tower", "Tesla",
		"Valkyrie", "Wizard", "Musketeer", "Archers",
		"Fireball", "Poison", "Log", "Arrows",
	},
	"siege": {
		"X-Bow", "Mortar", "Tesla", "Cannon", "Archers",
		"Ice Wizard", "Rocket", "Log", "Fireball",
		"Skeletons", "Ice Spirit",
	},
	"bridge_spam": {
		"Battle Ram", "Bandit", "Royal Ghost", "P.E.K.K.A",
		"Mega Knight", "Electro Wizard", "Magic Archer",
		"Poison", "Zap",
	},
	"midrange": {
		"Hog Rider", "Royal Giant", "Valkyrie", "Musketeer",
		"Knight", "Fireball", "Zap", "Cannon",
	},
	"spawndeck": {
		"Goblin Hut", "Barbarian Hut", "Furnace", "Tombstone",
		"Poison", "Fireball", "Graveyard", "Giant",
	},
	"bait": {
		"Goblin Barrel", "Princess", "Goblin Gang", "Skeleton Army",
		"Rocket", "Log", "Zap", "Arrows",
		"Knight", "Ice Spirit", "Inferno Tower",
	},
}

// ArchetypeAvoidanceScorer calculates penalties for cards associated with avoided archetypes
type ArchetypeAvoidanceScorer struct {
	avoidArchetypes map[string]bool
}

// NewArchetypeAvoidanceScorer creates a scorer that penalizes cards from avoided archetypes
func NewArchetypeAvoidanceScorer(avoidArchetypes []string) *ArchetypeAvoidanceScorer {
	scorer := &ArchetypeAvoidanceScorer{
		avoidArchetypes: make(map[string]bool),
	}

	// Normalize and store archetypes to avoid
	for _, arch := range avoidArchetypes {
		normalized := normalizeArchetype(arch)
		if normalized != "" {
			scorer.avoidArchetypes[normalized] = true
		}
	}

	return scorer
}

// normalizeArchetype converts user input to normalized archetype key
func normalizeArchetype(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))

	switch normalized {
	case "beatdown":
		return "beatdown"
	case "cycle":
		return "cycle"
	case "control":
		return "control"
	case "siege":
		return "siege"
	case "bridge_spam", "bridgespam", "bridge spam":
		return "bridge_spam"
	case "midrange", "mid-range", "mid range":
		return "midrange"
	case "spawndeck", "spawn deck", "spawn":
		return "spawndeck"
	case "bait", "spell bait", "spell_bait":
		return "bait"
	default:
		return "" // Invalid archetype
	}
}

// ScoreCard returns a penalty (negative score) for cards associated with avoided archetypes
// Returns 0.0 if card is not in avoided archetypes, or negative penalty if it is
func (s *ArchetypeAvoidanceScorer) ScoreCard(cardName string) float64 {
	if len(s.avoidArchetypes) == 0 {
		return 0.0
	}

	penalty := 0.0

	// Check each avoided archetype
	for archetype := range s.avoidArchetypes {
		preferredCards, exists := archetypeCardAssociations[archetype]
		if !exists {
			continue
		}

		// Check if card is in preferred cards (strong association)
		for _, preferredCard := range preferredCards {
			if strings.EqualFold(cardName, preferredCard) {
				// Strong penalty for preferred cards of avoided archetypes
				penalty -= 0.3
				break
			}
		}
	}

	return penalty
}

// ScoreDeck returns the average penalty for all cards in a deck
func (s *ArchetypeAvoidanceScorer) ScoreDeck(cardNames []string) float64 {
	if len(cardNames) == 0 || len(s.avoidArchetypes) == 0 {
		return 0.0
	}

	totalPenalty := 0.0
	for _, cardName := range cardNames {
		totalPenalty += s.ScoreCard(cardName)
	}

	return totalPenalty / float64(len(cardNames))
}

// IsEnabled returns whether any archetypes are being avoided
func (s *ArchetypeAvoidanceScorer) IsEnabled() bool {
	return len(s.avoidArchetypes) > 0
}
