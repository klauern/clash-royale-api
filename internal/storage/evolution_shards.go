// Package storage provides file I/O utilities for persisting Clash Royale data.
package storage

import "time"

// EvolutionShardInventory tracks shard counts per card with a last-updated timestamp.
type EvolutionShardInventory struct {
	LastUpdated time.Time      `json:"last_updated"`
	Shards      map[string]int `json:"shards"`
}

// NewEvolutionShardInventory returns an initialized inventory with an empty shard map.
func NewEvolutionShardInventory() EvolutionShardInventory {
	return EvolutionShardInventory{Shards: make(map[string]int)}
}

// LoadEvolutionShardInventory reads the inventory from disk, returning an empty inventory if missing.
func LoadEvolutionShardInventory(filePath string) (EvolutionShardInventory, error) {
	if !FileExists(filePath) {
		return NewEvolutionShardInventory(), nil
	}

	var inventory EvolutionShardInventory
	if err := ReadJSON(filePath, &inventory); err != nil {
		return EvolutionShardInventory{}, err
	}
	if inventory.Shards == nil {
		inventory.Shards = make(map[string]int)
	}

	return inventory, nil
}

// SaveEvolutionShardInventory writes the inventory to disk.
func SaveEvolutionShardInventory(filePath string, inventory EvolutionShardInventory) error {
	if inventory.Shards == nil {
		inventory.Shards = make(map[string]int)
	}
	return WriteJSON(filePath, inventory)
}
