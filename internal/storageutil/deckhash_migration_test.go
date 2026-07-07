package storageutil

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deckhash"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func TestParseNamedDeckHashMigrationRow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rows := queryMigrationRows(t, `["archers","giant"]`)
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("expected migration row")
		}

		row, err := ParseNamedDeckHashMigrationRow(rows, "deck row")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if !row.Valid {
			t.Fatal("expected parsed row to be marked valid")
		}
		if row.Canonical == "" {
			t.Fatal("expected canonical hash to be populated")
		}
	})

	t.Run("invalid cards json", func(t *testing.T) {
		rows := queryMigrationRows(t, `{"invalid":true}`)
		defer rows.Close()

		if !rows.Next() {
			t.Fatal("expected migration row")
		}

		row, err := ParseNamedDeckHashMigrationRow(rows, "deck row")
		if err == nil {
			t.Fatal("expected parse error")
		}
		if row.ID != 1 {
			t.Fatalf("expected scanned row id 1, got %d", row.ID)
		}
		if !strings.Contains(err.Error(), "invalid cards JSON for deck row 1") {
			t.Fatalf("unexpected error text: %v", err)
		}

		var typeErr *json.UnmarshalTypeError
		if !errors.As(err, &typeErr) {
			t.Fatalf("expected wrapped json unmarshal type error, got: %T", err)
		}
	})
}

func TestMaybeRunDeckHashMigration(t *testing.T) {
	db := openMigrationTestDB(t)

	if _, err := db.Exec(`
		CREATE TABLE decks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_hash TEXT NOT NULL UNIQUE,
			cards TEXT NOT NULL,
			overall_score REAL NOT NULL
		);
		CREATE TABLE migrations (
			name TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`); err != nil {
		t.Fatalf("failed to create migration test schema: %v", err)
	}

	cards := []string{"archers", "giant"}
	cardsJSON := `["archers","giant"]`
	canonicalHash := deckhash.DeckHash(cards)
	if _, err := db.Exec(
		`INSERT INTO decks (deck_hash, cards, overall_score) VALUES (?, ?, ?), (?, ?, ?)`,
		"legacy-hash", cardsJSON, 7.0,
		"already-canonical", cardsJSON, 9.0,
	); err != nil {
		t.Fatalf("failed to seed migration rows: %v", err)
	}

	var afterCommitCalls int
	err := MaybeRunDeckHashMigration(db, DeckHashMigrationConfig{
		MigrationName: "deck_hash_canonical_v1",
		TableName:     "decks",
		LoadRows: func() ([]DeckHashMigrationRow, map[string]DeckHashMigrationRow, error) {
			rows, err := db.Query(`SELECT id, deck_hash, cards, overall_score FROM decks ORDER BY id`)
			if err != nil {
				return nil, nil, err
			}
			defer rows.Close()

			records := make([]DeckHashMigrationRow, 0)
			for rows.Next() {
				row, err := ParseNamedDeckHashMigrationRow(rows, "deck row")
				if err != nil {
					return nil, nil, err
				}
				records = append(records, row)
			}
			if err := rows.Err(); err != nil {
				return nil, nil, err
			}

			return records, SelectDeckHashMigrationWinners(records), nil
		},
		BeginTxError:  "failed to start deck hash migration transaction",
		RollbackError: "failed to rollback deck hash migration",
		CommitError:   "failed to commit deck hash migration",
		AfterCommit: func() error {
			afterCommitCalls++
			return nil
		},
		AfterCommitError: "failed to run post-migration hook",
	})
	if err != nil {
		t.Fatalf("expected migration to succeed, got error: %v", err)
	}
	if afterCommitCalls != 1 {
		t.Fatalf("expected after-commit hook to run once, got %d", afterCommitCalls)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM decks`).Scan(&count); err != nil {
		t.Fatalf("failed to count migrated rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one deduplicated row, got %d", count)
	}

	var migratedHash string
	if err := db.QueryRow(`SELECT deck_hash FROM decks LIMIT 1`).Scan(&migratedHash); err != nil {
		t.Fatalf("failed to read migrated hash: %v", err)
	}
	if migratedHash != canonicalHash {
		t.Fatalf("expected canonical hash %q, got %q", canonicalHash, migratedHash)
	}

	var migrationCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, "deck_hash_canonical_v1").Scan(&migrationCount); err != nil {
		t.Fatalf("failed to read migration marker: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected migration marker to be recorded once, got %d", migrationCount)
	}

	if err := MaybeRunDeckHashMigration(db, DeckHashMigrationConfig{
		MigrationName: "deck_hash_canonical_v1",
		TableName:     "decks",
		LoadRows: func() ([]DeckHashMigrationRow, map[string]DeckHashMigrationRow, error) {
			t.Fatal("expected applied migration to short-circuit before reloading rows")
			return nil, nil, nil
		},
		BeginTxError:  "failed to start deck hash migration transaction",
		RollbackError: "failed to rollback deck hash migration",
		CommitError:   "failed to commit deck hash migration",
		AfterCommit: func() error {
			t.Fatal("expected applied migration to skip after-commit hook")
			return nil
		},
		AfterCommitError: "failed to run post-migration hook",
	}); err != nil {
		t.Fatalf("expected already-applied migration to no-op, got error: %v", err)
	}
}

func queryMigrationRows(t *testing.T, cardsJSON string) *sql.Rows {
	t.Helper()

	db := openMigrationTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE migration_rows (
			id INTEGER PRIMARY KEY,
			deck_hash TEXT NOT NULL,
			cards TEXT NOT NULL,
			overall_score REAL NOT NULL
		)
	`); err != nil {
		t.Fatalf("failed to create migration_rows table: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO migration_rows (id, deck_hash, cards, overall_score) VALUES (1, 'legacy', ?, 5.0)`,
		cardsJSON,
	); err != nil {
		t.Fatalf("failed to seed migration_rows table: %v", err)
	}

	rows, err := db.Query(`SELECT id, deck_hash, cards, overall_score FROM migration_rows`)
	if err != nil {
		t.Fatalf("failed to query migration rows: %v", err)
	}

	return rows
}

func openMigrationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close sqlite database: %v", err)
		}
	})

	return db
}
