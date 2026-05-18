package storageutil

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deckhash"
)

// IsMigrationApplied returns true if the migration exists in the migrations table.
func IsMigrationApplied(db *sql.DB, name string) (bool, error) {
	var exists int
	err := db.QueryRow("SELECT 1 FROM migrations WHERE name = ?", name).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query migrations table: %w", err)
	}
	return true, nil
}

// RecordMigration inserts a migration marker into the migrations table.
func RecordMigration(db *sql.DB, name string) error {
	if _, err := db.Exec(
		"INSERT INTO migrations (name, applied_at) VALUES (?, CURRENT_TIMESTAMP)",
		name,
	); err != nil {
		return fmt.Errorf("failed to record deck hash migration: %w", err)
	}
	return nil
}

// DeckHashMigrationRow is the shared row shape used during deck hash canonicalization.
type DeckHashMigrationRow struct {
	ID           int
	DeckHash     string
	CardsJSON    string
	OverallScore float64
	Canonical    string
	Valid        bool
}

// ParseDeckHashMigrationRow scans migration fields and computes the canonical deck hash.
func ParseDeckHashMigrationRow(rows *sql.Rows) (DeckHashMigrationRow, error) {
	var row DeckHashMigrationRow
	if err := rows.Scan(&row.ID, &row.DeckHash, &row.CardsJSON, &row.OverallScore); err != nil {
		return row, err
	}

	var cards []string
	if err := json.Unmarshal([]byte(row.CardsJSON), &cards); err != nil {
		return row, err
	}

	row.Canonical = deckhash.DeckHash(cards)
	row.Valid = true
	return row, nil
}

// PreferDeckHashMigrationWinner determines which row should be kept for a canonical hash.
func PreferDeckHashMigrationWinner(candidate, current DeckHashMigrationRow) bool {
	if candidate.OverallScore != current.OverallScore {
		return candidate.OverallScore > current.OverallScore
	}
	return candidate.ID < current.ID
}
