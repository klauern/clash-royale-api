package mulligan

import (
	"strings"
	"time"
)

// Generator creates mulligan guides for decks
type Generator struct {
	cardDatabase map[string]CardInfo
}

// CardInfo contains information about a card for mulligan analysis
type CardInfo struct {
	Name         string
	Elixir       int
	Type         string // "troop", "spell", "building"
	Rarity       string
	Role         CardRole
	OpeningScore float64
}

// NewGenerator creates a new mulligan guide generator
func NewGenerator() *Generator {
	return &Generator{
		cardDatabase: initializeCardDatabase(),
	}
}

// GenerateGuide creates a mulligan guide for the given deck
func (g *Generator) GenerateGuide(deckCards []string, deckName string) (*MulliganGuide, error) {
	// Analyze deck composition
	analysis := g.analyzeDeck(deckCards)

	// Generate general principles based on archetype
	principles := g.generatePrinciples(analysis)

	// Generate matchup-specific strategies
	matchups := g.generateMatchups(analysis)

	// Identify cards to never open with
	neverOpenWith := g.identifyBadOpenings(analysis)

	// Identify ideal opening cards
	idealOpenings := g.identifyIdealOpenings(analysis)

	return &MulliganGuide{
		DeckName:          deckName,
		DeckCards:         deckCards,
		Archetype:         analysis.Archetype,
		GeneralPrinciples: principles,
		Matchups:          matchups,
		NeverOpenWith:     neverOpenWith,
		IdealOpenings:     idealOpenings,
		GeneratedAt:       time.Now(),
	}, nil
}

// analyzeDeck performs strategic analysis of the deck
func (g *Generator) analyzeDeck(deckCards []string) DeckAnalysis {
	var openingCards []OpeningCard
	var winConditions, defensiveCards, cycleCards, spells, buildings []string
	totalElixir := 0

	for _, cardName := range deckCards {
		cardInfo, exists := g.cardDatabase[cardName]
		if !exists {
			// Default handling for unknown cards
			cardInfo = CardInfo{
				Name:         cardName,
				Elixir:       4,
				Type:         "troop",
				Role:         RoleSupport,
				OpeningScore: 0.5,
			}
		}

		totalElixir += cardInfo.Elixir

		openingCard := OpeningCard{
			Name:         cardName,
			Elixir:       cardInfo.Elixir,
			Role:         cardInfo.Role,
			OpeningScore: cardInfo.OpeningScore,
			Reasons:      g.getOpeningReasons(cardInfo),
		}
		openingCards = append(openingCards, openingCard)

		// Categorize cards
		switch cardInfo.Role {
		case RoleWinCondition:
			winConditions = append(winConditions, cardName)
		case RoleDefensive, RoleBuilding:
			defensiveCards = append(defensiveCards, cardName)
			if cardInfo.Type == "building" {
				buildings = append(buildings, cardName)
			}
		case RoleCycle:
			cycleCards = append(cycleCards, cardName)
		case RoleSpell:
			spells = append(spells, cardName)
		}
	}

	avgElixir := float64(totalElixir) / float64(len(deckCards))
	archetype := g.determineArchetype(winConditions, defensiveCards, cycleCards, spells, avgElixir)

	return DeckAnalysis{
		DeckCards:      deckCards,
		Archetype:      archetype,
		AvgElixir:      avgElixir,
		OpeningCards:   openingCards,
		WinConditions:  winConditions,
		DefensiveCards: defensiveCards,
		CycleCards:     cycleCards,
		Spells:         spells,
		Buildings:      buildings,
	}
}

// determineArchetype identifies the deck's playstyle
func (g *Generator) determineArchetype(winConditions, defensive, cycle, spells []string, avgElixir float64) Archetype {
	winConCount := len(winConditions)
	defCount := len(defensive)
	cycleCount := len(cycle)
	spellCount := len(spells)

	// High elixir with big win conditions = beatdown
	if avgElixir > 4.0 && winConCount >= 1 {
		for _, winCon := range winConditions {
			if strings.Contains(strings.ToLower(winCon), "golem") ||
				strings.Contains(strings.ToLower(winCon), "giant") ||
				strings.Contains(strings.ToLower(winCon), "lava hound") {
				return ArchetypeBeatdown
			}
		}
	}

	// Low elixir with fast cycle = cycle deck
	if avgElixir < 3.5 && cycleCount >= 3 {
		return ArchetypeCycle
	}

	// Many spells = control/bait
	if spellCount >= 3 {
		return ArchetypeBait
	}

	// Many buildings = siege or defensive
	if defCount >= 3 {
		// Check for siege buildings
		for _, def := range defensive {
			if strings.Contains(strings.ToLower(def), "xbow") ||
				strings.Contains(strings.ToLower(def), "mortar") {
				return ArchetypeSiege
			}
		}
		return ArchetypeControl
	}

	// Check for bridge spam characteristics
	if avgElixir >= 3.0 && avgElixir <= 4.0 && winConCount >= 1 {
		for _, winCon := range winConditions {
			if strings.Contains(strings.ToLower(winCon), "battle ram") ||
				strings.Contains(strings.ToLower(winCon), "hog rider") {
				return ArchetypeBridgeSpam
			}
		}
	}

	// Default to midrange
	return ArchetypeMidrange
}

// generatePrinciples creates general opening principles based on archetype
func (g *Generator) generatePrinciples(analysis DeckAnalysis) []string {
	var principles []string

	switch analysis.Archetype {
	case ArchetypeBeatdown:
		principles = []string{
			"Start with cheap cycle cards to find your tank",
			"Never commit your main win condition without elixir advantage",
			"Play defensive buildings reactively, not proactively",
			"Build big pushes in single lane when you have elixir lead",
		}
	case ArchetypeCycle:
		principles = []string{
			"Open with your cheapest cycle card",
			"Apply constant pressure to prevent opponent building",
			"Keep cycling back to your win condition quickly",
			"Use defensive building only when needed for specific threat",
		}
	case ArchetypeControl:
		principles = []string{
			"Open with cheap cycle or defensive card",
			"Scout opponent's deck before committing major resources",
			"Save key spells for high-value targets",
			"Use buildings to control both lanes when possible",
		}
	case ArchetypeSiege:
		principles = []string{
			"Protect siege building at all costs",
			"Apply pressure in opposite lane when placing siege",
			"Have cheap cycle ready to defend siege building",
			"Never open with siege building against unknown deck",
		}
	case ArchetypeBridgeSpam:
		principles = []string{
			"Apply early pressure with medium-cost units",
			"Support win conditions with cheap troops",
			"Use defensive building to counter their pressure",
			"Cycle quickly to maintain pressure",
		}
	default:
		principles = []string{
			"Start with safe, low-cost cycle cards",
			"Scout opponent's strategy before committing",
			"Maintain elixir advantage in early game",
			"Adapt opening strategy based on opponent's first play",
		}
	}

	return principles
}

// generateMatchups creates matchup-specific opening strategies
func (g *Generator) generateMatchups(analysis DeckAnalysis) []Matchup {
	matchups := []Matchup{
		{
			OpponentType: "Beatdown (Giant, Golem, Lava Hound)",
			OpeningPlay:  g.getBeatdownOpening(analysis),
			Reason:       "Apply early pressure, force them to defend instead of building",
			Backup:       "If they ignore pressure, build stronger defense and counter-push",
			KeyCards:     analysis.DefensiveCards,
			DangerLevel:  "medium",
		},
		{
			OpponentType: "Hog Cycle / Fast Cycle",
			OpeningPlay:  g.getCycleOpening(analysis),
			Reason:       "Have defense ready for immediate pressure",
			Backup:       "Start cycling if they play defensive, maintain pressure",
			KeyCards:     append(analysis.DefensiveCards, analysis.CycleCards...),
			DangerLevel:  "high",
		},
		{
			OpponentType: "Bridge Spam (Battle Ram, Bandit)",
			OpeningPlay:  g.getBridgeSpamOpening(analysis),
			Reason:       "Defensive positioning ready for sudden pressure",
			Backup:       "Use cheap troops to swarm their units",
			KeyCards:     analysis.CycleCards,
			DangerLevel:  "high",
		},
		{
			OpponentType: "Siege (X-Bow, Mortar)",
			OpeningPlay:  g.getSiegeOpening(analysis),
			Reason:       "Prevent siege lock, force defensive play",
			Backup:       "Apply pressure in opposite lane",
			KeyCards:     analysis.WinConditions,
			DangerLevel:  "high",
		},
		{
			OpponentType: "Control / Spell Bait",
			OpeningPlay:  g.getControlOpening(analysis),
			Reason:       "Cheap cycle to scout their strategy",
			Backup:       "Mirror their play style, punish overextensions",
			KeyCards:     analysis.Spells,
			DangerLevel:  "medium",
		},
		{
			OpponentType: "Unknown (Scout First)",
			OpeningPlay:  g.getSafeOpening(analysis),
			Reason:       "Safe play that provides information",
			Backup:       "React defensively to their first move",
			KeyCards:     analysis.CycleCards,
			DangerLevel:  "low",
		},
	}

	return matchups
}

// Helper methods for specific opening plays
func (g *Generator) getBeatdownOpening(analysis DeckAnalysis) string {
	if len(analysis.WinConditions) > 0 {
		return analysis.WinConditions[0] + " at bridge"
	}
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " at bridge"
	}
	return "Medium cost troop at bridge"
}

func (g *Generator) getCycleOpening(analysis DeckAnalysis) string {
	if len(analysis.Buildings) > 0 {
		return analysis.Buildings[0] + " in center (4 from river)"
	}
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " in back"
	}
	return "Defensive positioning in center"
}

func (g *Generator) getBridgeSpamOpening(analysis DeckAnalysis) string {
	if len(analysis.DefensiveCards) > 0 {
		return analysis.DefensiveCards[0] + " in back"
	}
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " in back"
	}
	return "Ranged troop in back"
}

func (g *Generator) getSiegeOpening(analysis DeckAnalysis) string {
	if len(analysis.WinConditions) > 0 {
		return analysis.WinConditions[0] + " opposite lane immediately"
	}
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " at bridge"
	}
	return "Immediate pressure at bridge"
}

func (g *Generator) getControlOpening(analysis DeckAnalysis) string {
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " in back"
	}
	return "Cheap card in back"
}

func (g *Generator) getSafeOpening(analysis DeckAnalysis) string {
	if len(analysis.CycleCards) > 0 {
		return analysis.CycleCards[0] + " in back"
	}
	if len(analysis.DefensiveCards) > 0 {
		return analysis.DefensiveCards[0] + " in back"
	}
	return "Lowest elixir card in back"
}

// identifyBadOpenings lists cards that should never be opened with
func (g *Generator) identifyBadOpenings(analysis DeckAnalysis) []string {
	var badOpenings []string

	for _, spell := range analysis.Spells {
		// High damage spells shouldn't be opened with
		cardInfo := g.cardDatabase[spell]
		if cardInfo.Elixir >= 4 {
			badOpenings = append(badOpenings, spell+" - waste of elixir without targets")
		}
	}

	for _, winCon := range analysis.WinConditions {
		// Expensive win conditions shouldn't be opened with
		cardInfo := g.cardDatabase[winCon]
		if cardInfo.Elixir >= 5 {
			badOpenings = append(badOpenings, winCon+" - too expensive for opening")
		}
	}

	// Defensive buildings should be reactive
	for _, building := range analysis.Buildings {
		badOpenings = append(badOpenings, building+" - play reactively, not proactively")
	}

	return badOpenings
}

// identifyIdealOpenings lists the best opening cards
func (g *Generator) identifyIdealOpenings(analysis DeckAnalysis) []string {
	var idealOpenings []string

	// Check all cards for good opening potential
	for _, card := range analysis.DeckCards {
		cardInfo, exists := g.cardDatabase[card]
		if !exists {
			continue
		}

		// Consider card ideal if:
		// 1. It's cheap (≤3 elixir) AND has decent opening score (≥0.5)
		// 2. OR it has high opening score (≥0.7) regardless of elixir cost
		if (cardInfo.Elixir <= 3 && cardInfo.OpeningScore >= 0.5) ||
			cardInfo.OpeningScore >= 0.7 {
			idealOpenings = append(idealOpenings, card)
		}
	}

	return idealOpenings
}

// getOpeningReasons provides reasons why a card is good/bad for opening
func (g *Generator) getOpeningReasons(cardInfo CardInfo) []string {
	var reasons []string

	if cardInfo.Elixir <= 2 {
		reasons = append(reasons, "Low cost allows fast cycling")
	}

	if cardInfo.Elixir >= 5 {
		reasons = append(reasons, "Too expensive for opening play")
	}

	if cardInfo.Role == RoleSpell {
		reasons = append(reasons, "Better saved for high-value targets")
	}

	if cardInfo.Type == "building" && cardInfo.Role == RoleDefensive {
		reasons = append(reasons, "Defensive building should be reactive")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Viable opening card")
	}

	return reasons
}
