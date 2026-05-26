package recommend

import (
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	rarityChampion  = "Champion"
	cardNameBalloon = "Balloon"
)

// championSlotOnlyEvos lists evolved cards that may only occupy the champion slot.
// This is the only hand-curated exception; all other SlotTypes derive from rarity
// and evolution data available in the card API.
var championSlotOnlyEvos = map[string]bool{
	cardNameBalloon: true,
}

// CardSlotTyper resolves SlotType from card metadata.
// Populate it from the /cards API response or from a CardAnalysis.
type CardSlotTyper struct {
	cardRarity map[string]string
}

// NewCardSlotTyper builds a typeer from the full card catalog (GetCards response).
func NewCardSlotTyper(cards []clashroyale.Card) *CardSlotTyper {
	rarities := make(map[string]string, len(cards))
	for _, c := range cards {
		rarities[c.Name] = c.Rarity
	}
	return &CardSlotTyper{cardRarity: rarities}
}

// NewCardSlotTyperFromDeckAnalysis builds a typeer from deck.CardAnalysis.CardLevels,
// which carries rarity for every card in the player's collection.
func NewCardSlotTyperFromDeckAnalysis(a deck.CardAnalysis) *CardSlotTyper {
	rarities := make(map[string]string, len(a.CardLevels))
	for name, data := range a.CardLevels {
		rarities[name] = data.Rarity
	}
	return &CardSlotTyper{cardRarity: rarities}
}

// SlotTypeFor returns the SlotType for a card given its name and whether it
// is currently evolved in this deck (evolutionLevel > 0).
func (t *CardSlotTyper) SlotTypeFor(cardName string, evolved bool) SlotType {
	if t.cardRarity[cardName] == rarityChampion {
		return Champion
	}
	if evolved {
		if championSlotOnlyEvos[cardName] {
			return ChampionSlotOnlyEvo
		}
		return RegularEvo
	}
	return RegularCard
}

// ClassifyDeckDetail converts a DeckRecommendation's cards to CardWithSlotType.
func (t *CardSlotTyper) ClassifyDeckDetail(cards []deck.CardDetail) []CardWithSlotType {
	result := make([]CardWithSlotType, len(cards))
	for i, c := range cards {
		result[i] = CardWithSlotType{
			Name:     c.Name,
			SlotType: t.SlotTypeFor(c.Name, c.EvolutionLevel > 0),
		}
	}
	return result
}

// CardDetailLike is a minimal struct for slot classification when not using deck.CardDetail.
type CardDetailLike struct {
	CardName string
	Evolved  bool
}

// ClassifyCards converts CardDetailLike entries to CardWithSlotType.
func (t *CardSlotTyper) ClassifyCards(cards []CardDetailLike) []CardWithSlotType {
	result := make([]CardWithSlotType, len(cards))
	for i, c := range cards {
		result[i] = CardWithSlotType{
			Name:     c.CardName,
			SlotType: t.SlotTypeFor(c.CardName, c.Evolved),
		}
	}
	return result
}
