package recommend

import "fmt"

// SlotType classifies a card by which deck slots it may legally occupy.
type SlotType int

const (
	// RegularCard has no special slot — occupies one of the 5 ordinary slots.
	RegularCard SlotType = iota
	// RegularEvo is a standard evolved card that may occupy the evo slot or flex slot.
	RegularEvo
	// ChampionSlotOnlyEvo is an evolved card that may only occupy the champion slot
	// or flex slot — never the regular evo slot (e.g. Balloon evo).
	ChampionSlotOnlyEvo
	// Champion is a champion-rarity card that may occupy the champion slot or flex slot.
	Champion
)

// CardWithSlotType pairs a card name with its resolved slot type.
type CardWithSlotType struct {
	Name     string
	SlotType SlotType
}

// DeckSlotPolicy defines the slot constraints for a legal Clash Royale deck.
// Rule: 1 evo slot + 1 champion slot + 1 flex slot = 3 total special cards max.
type DeckSlotPolicy struct {
	MaxEvoSlots      int // 1: RegularEvo only
	MaxChampionSlots int // 1: Champion or ChampionSlotOnlyEvo
	MaxFlexSlots     int // 1: RegularEvo or Champion
}

// DefaultPolicy returns the standard Clash Royale slot policy.
func DefaultPolicy() DeckSlotPolicy {
	return DeckSlotPolicy{
		MaxEvoSlots:      1,
		MaxChampionSlots: 1,
		MaxFlexSlots:     1,
	}
}

// Validate checks that deck satisfies the slot policy.
// Returns the first violation found, or nil if the deck is legal.
//
// Slot assignment rules:
//   - EvoSlot   (1): RegularEvo only
//   - ChampSlot (1): Champion or ChampionSlotOnlyEvo (only these go here)
//   - FlexSlot  (1): RegularEvo or Champion (NOT ChampionSlotOnlyEvo)
func (p DeckSlotPolicy) Validate(deck []CardWithSlotType) error {
	var evos, champions, champSlotEvos int
	for _, c := range deck {
		switch c.SlotType {
		case RegularEvo:
			evos++
		case Champion:
			champions++
		case ChampionSlotOnlyEvo:
			champSlotEvos++
		}
	}

	// Rule 1: max 1 Champion card per deck (game limit).
	if champions > p.MaxChampionSlots {
		return fmt.Errorf(
			"slot policy violation: %d champion(s) in deck — max %d allowed",
			champions, p.MaxChampionSlots,
		)
	}

	// Rule 2: ChampionSlotOnlyEvo may only occupy the champion slot.
	// Flex is restricted to RegularEvo or Champion, so if two ChampionSlotOnlyEvo
	// cards are present, both compete for the single champion slot.
	if champSlotEvos > p.MaxChampionSlots {
		return fmt.Errorf(
			"slot policy violation: %d ChampionSlotOnlyEvo card(s) — only %d champion slot(s) available",
			champSlotEvos, p.MaxChampionSlots,
		)
	}

	// Rule 3: champion slot shared between Champion and ChampionSlotOnlyEvo.
	// If both are present, one must go in flex — only Champion may do so.
	champSlotUsed := champions + champSlotEvos
	if champSlotUsed > p.MaxChampionSlots {
		champOverflow := champSlotUsed - p.MaxChampionSlots
		// Only Champion (not ChampionSlotOnlyEvo) may overflow to flex.
		// Since Rule 2 ensures champSlotEvos <= 1, overflow must come from Champion.
		if champions < champOverflow {
			return fmt.Errorf(
				"slot policy violation: cannot route %d ChampionSlotOnlyEvo overflow to flex",
				champSlotEvos,
			)
		}
	}

	// Rule 4: evo capacity — evo slot + flex slot accommodate RegularEvo cards;
	// flex is also shared with any Champion overflow.
	champOverflow := max(0, champSlotUsed-p.MaxChampionSlots)
	evoSlotFills := min(evos, p.MaxEvoSlots)
	evoOverflow := evos - evoSlotFills
	flexDemand := evoOverflow + champOverflow
	if flexDemand > p.MaxFlexSlots {
		return fmt.Errorf(
			"slot policy violation: deck needs %d flex slot(s) (evo overflow %d + champ overflow %d) but only %d available",
			flexDemand, evoOverflow, champOverflow, p.MaxFlexSlots,
		)
	}

	return nil
}

// SlotViolations returns all slot rule violations in the deck (not just the first).
func (p DeckSlotPolicy) SlotViolations(deck []CardWithSlotType) []string {
	if err := p.Validate(deck); err != nil {
		return []string{err.Error()}
	}
	return nil
}
