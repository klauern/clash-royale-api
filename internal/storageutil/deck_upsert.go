package storageutil

import (
	"encoding/json"
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deckhash"
)

// ExistingDeckRecord captures the existing row returned by a deck-hash lookup.
type ExistingDeckRecord struct {
	ID    int
	Score float64
}

// DeckUpsertHooks provides package-specific storage operations for shared deck upserts.
type DeckUpsertHooks struct {
	LookupExisting func(deckHash string) (*ExistingDeckRecord, error)
	Insert         func(deckHash, cardsJSON string) (int, error)
	UpdateExisting func(existing ExistingDeckRecord, deckHash, cardsJSON string) error
}

// DeckUpsertResult reports the canonical hash and selected row identity.
type DeckUpsertResult struct {
	DeckHash  string
	CardsJSON string
	ID        int
	IsNew     bool
}

// UpsertDeck centralizes canonical hash generation, card serialization, and
// existing-row lookup before storage-specific insert/update callbacks run.
func UpsertDeck(cards []string, hooks DeckUpsertHooks) (DeckUpsertResult, error) {
	deckHash := deckhash.DeckHash(cards)
	cardsJSONBytes, err := json.Marshal(cards)
	if err != nil {
		return DeckUpsertResult{}, fmt.Errorf("failed to marshal cards: %w", err)
	}
	cardsJSON := string(cardsJSONBytes)

	existing, err := hooks.LookupExisting(deckHash)
	if err != nil {
		return DeckUpsertResult{}, err
	}
	if existing == nil {
		id, err := hooks.Insert(deckHash, cardsJSON)
		if err != nil {
			return DeckUpsertResult{}, err
		}
		return DeckUpsertResult{DeckHash: deckHash, CardsJSON: cardsJSON, ID: id, IsNew: true}, nil
	}

	if hooks.UpdateExisting != nil {
		if err := hooks.UpdateExisting(*existing, deckHash, cardsJSON); err != nil {
			return DeckUpsertResult{}, err
		}
	}

	return DeckUpsertResult{DeckHash: deckHash, CardsJSON: cardsJSON, ID: existing.ID, IsNew: false}, nil
}
