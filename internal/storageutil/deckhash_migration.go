package storageutil

import (
	"database/sql"
	"encoding/json"
	"errors"
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

// EnsureMigration runs the migration once and records it on success.
func EnsureMigration(db *sql.DB, name string, run func() error) error {
	applied, err := IsMigrationApplied(db, name)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	if err := run(); err != nil {
		return err
	}
	return RecordMigration(db, name)
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

// ParseNamedDeckHashMigrationRow parses a migration row and wraps errors with a row label.
func ParseNamedDeckHashMigrationRow(rows *sql.Rows, rowLabel string) (DeckHashMigrationRow, error) {
	row, err := ParseDeckHashMigrationRow(rows)
	if err != nil {
		return row, WrapDeckHashMigrationRowError(row, err, rowLabel)
	}
	return row, nil
}

// WrapDeckHashMigrationRowError normalizes row parsing errors across storage packages.
func WrapDeckHashMigrationRowError(row DeckHashMigrationRow, err error, rowLabel string) error {
	if isDeckHashMigrationJSONError(err) {
		return fmt.Errorf("invalid cards JSON for %s %d: %w", rowLabel, row.ID, err)
	}

	return fmt.Errorf("failed to scan %s %d: %w", rowLabel, row.ID, err)
}

func isDeckHashMigrationJSONError(err error) bool {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}

	var unmarshalTypeErr *json.UnmarshalTypeError
	return errors.As(err, &unmarshalTypeErr)
}

// LoadDeckHashMigrationRows parses migration rows and selects canonical winners.
func LoadDeckHashMigrationRows(rows *sql.Rows, rowLabel string, onInvalidJSON func(DeckHashMigrationRow, error)) ([]DeckHashMigrationRow, map[string]DeckHashMigrationRow, error) {
	records := make([]DeckHashMigrationRow, 0)
	for rows.Next() {
		row, err := ParseNamedDeckHashMigrationRow(rows, rowLabel)
		if err != nil {
			if onInvalidJSON != nil && isDeckHashMigrationJSONError(err) {
				onInvalidJSON(row, err)
				records = append(records, row)
				continue
			}
			return nil, nil, err
		}
		records = append(records, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate deck hash migration rows: %w", err)
	}
	return records, SelectDeckHashMigrationWinners(records), nil
}

// PreferDeckHashMigrationWinner determines which row should be kept for a canonical hash.
func PreferDeckHashMigrationWinner(candidate, current DeckHashMigrationRow) bool {
	if candidate.OverallScore != current.OverallScore {
		return candidate.OverallScore > current.OverallScore
	}
	return candidate.ID < current.ID
}

// SelectDeckHashMigrationWinners picks one canonical winner per canonical hash.
func SelectDeckHashMigrationWinners(records []DeckHashMigrationRow) map[string]DeckHashMigrationRow {
	winners := make(map[string]DeckHashMigrationRow)
	for _, row := range records {
		if !row.Valid {
			continue
		}
		current, exists := winners[row.Canonical]
		if !exists || PreferDeckHashMigrationWinner(row, current) {
			winners[row.Canonical] = row
		}
	}
	return winners
}

// ApplyDeckHashMigration deletes duplicate rows and updates winner hashes to canonical values.
func ApplyDeckHashMigration(tx *sql.Tx, tableName string, records []DeckHashMigrationRow, winners map[string]DeckHashMigrationRow) error {
	for _, row := range records {
		if !row.Valid {
			continue
		}
		winner := winners[row.Canonical]
		if row.ID == winner.ID {
			continue
		}
		query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
		if _, err := tx.Exec(query, row.ID); err != nil {
			return fmt.Errorf("failed to delete duplicate %s row %d: %w", tableName, row.ID, err)
		}
	}

	for _, winner := range winners {
		if winner.DeckHash == winner.Canonical {
			continue
		}
		query := fmt.Sprintf("UPDATE %s SET deck_hash = ? WHERE id = ?", tableName)
		if _, err := tx.Exec(query, winner.Canonical, winner.ID); err != nil {
			return fmt.Errorf("failed to update %s hash for row %d: %w", tableName, winner.ID, err)
		}
	}
	return nil
}

// ApplyDeckHashMigrationInTx runs canonical deck-hash updates in one transaction.
func ApplyDeckHashMigrationInTx(db *sql.DB, migrationLabel, tableName string, records []DeckHashMigrationRow, winners map[string]DeckHashMigrationRow, afterCommit func() error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start %s deck hash migration transaction: %w", migrationLabel, err)
	}
	if err := ApplyDeckHashMigration(tx, tableName, records, winners); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed rollback %s deck hash migration after error %v: %w", migrationLabel, err, rollbackErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed commit %s deck hash migration: %w", migrationLabel, err)
	}
	if afterCommit != nil {
		if err := afterCommit(); err != nil {
			return err
		}
	}
	return nil
}
