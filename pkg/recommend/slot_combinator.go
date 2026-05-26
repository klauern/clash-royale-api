package recommend

import "sort"

// SlotAssignment describes one valid assignment of special cards to slots.
type SlotAssignment struct {
	// EvoSlot holds the RegularEvo assigned to the dedicated evo slot (nil if unused).
	EvoSlot *CardWithSlotType
	// ChampionSlot holds the Champion or ChampionSlotOnlyEvo in the champion slot (nil if unused).
	ChampionSlot *CardWithSlotType
	// FlexSlot holds the RegularEvo or Champion in the flex slot (nil if unused).
	FlexSlot *CardWithSlotType
	// Score is the evaluator score for this assignment (higher is better).
	Score float64
}

// SlotCandidates collects the special-card candidates from a classified deck.
type SlotCandidates struct {
	RegularEvos          []CardWithSlotType
	ChampionSlotOnlyEvos []CardWithSlotType
	Champions            []CardWithSlotType
}

// CollectCandidates partitions classified deck cards into slot candidate groups.
func CollectCandidates(deck []CardWithSlotType) SlotCandidates {
	var sc SlotCandidates
	for _, c := range deck {
		switch c.SlotType {
		case RegularEvo:
			sc.RegularEvos = append(sc.RegularEvos, c)
		case ChampionSlotOnlyEvo:
			sc.ChampionSlotOnlyEvos = append(sc.ChampionSlotOnlyEvos, c)
		case Champion:
			sc.Champions = append(sc.Champions, c)
		}
	}
	return sc
}

// ptr returns a pointer to a copy of v.
func ptr(v CardWithSlotType) *CardWithSlotType { return &v }

// EnumerateAssignments returns all valid slot assignments for the given
// candidates, scored by the provided scorer function, sorted descending by
// score. topN limits results (0 = return all).
//
// Valid assignment rules (DefaultPolicy):
//   - EvoSlot: one RegularEvo (optional)
//   - ChampionSlot: one Champion or one ChampionSlotOnlyEvo (optional)
//   - FlexSlot: one RegularEvo or one Champion (optional)
//   - No card may appear in more than one slot.
func EnumerateAssignments(
	candidates SlotCandidates,
	policy DeckSlotPolicy,
	scorer func(SlotAssignment) float64,
	topN int,
) []SlotAssignment {
	evoOptions := withEmpty(candidates.RegularEvos)
	champSlotOptions := withEmpty(append(candidates.Champions, candidates.ChampionSlotOnlyEvos...))

	var assignments []SlotAssignment
	for _, evoCard := range evoOptions {
		for _, champCard := range champSlotOptions {
			if sameCard(evoCard, champCard) {
				continue
			}
			for _, flexCard := range flexOptions(candidates, evoCard, champCard) {
				if a, ok := buildAssignment(evoCard, champCard, flexCard, policy, scorer); ok {
					assignments = append(assignments, a)
				}
			}
		}
	}

	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].Score > assignments[j].Score
	})
	if topN > 0 && len(assignments) > topN {
		assignments = assignments[:topN]
	}
	return assignments
}

// withEmpty prepends an empty sentinel card to options (represents "slot unused").
func withEmpty(cards []CardWithSlotType) []CardWithSlotType {
	return append([]CardWithSlotType{{}}, cards...)
}

func sameCard(a, b CardWithSlotType) bool {
	return a.Name != "" && b.Name != "" && a.Name == b.Name
}

// flexOptions builds the valid flex slot choices given the already-assigned evo and champ cards.
func flexOptions(sc SlotCandidates, evoCard, champCard CardWithSlotType) []CardWithSlotType {
	opts := []CardWithSlotType{{}} // empty = slot unused
	for _, c := range sc.RegularEvos {
		if c.Name != evoCard.Name {
			opts = append(opts, c)
		}
	}
	for _, c := range sc.Champions {
		if c.Name != champCard.Name {
			opts = append(opts, c)
		}
	}
	return opts
}

// buildAssignment constructs and validates a SlotAssignment, returning (a, true) if valid.
func buildAssignment(
	evoCard, champCard, flexCard CardWithSlotType,
	policy DeckSlotPolicy,
	scorer func(SlotAssignment) float64,
) (SlotAssignment, bool) {
	a := SlotAssignment{}
	if evoCard.Name != "" {
		a.EvoSlot = ptr(evoCard)
	}
	if champCard.Name != "" {
		a.ChampionSlot = ptr(champCard)
	}
	if flexCard.Name != "" {
		a.FlexSlot = ptr(flexCard)
	}
	if policy.Validate(assignmentToDeck(a)) != nil {
		return SlotAssignment{}, false
	}
	a.Score = scorer(a)
	return a, true
}

func assignmentToDeck(a SlotAssignment) []CardWithSlotType {
	var deck []CardWithSlotType
	if a.EvoSlot != nil {
		deck = append(deck, *a.EvoSlot)
	}
	if a.ChampionSlot != nil {
		deck = append(deck, *a.ChampionSlot)
	}
	if a.FlexSlot != nil {
		deck = append(deck, *a.FlexSlot)
	}
	return deck
}
