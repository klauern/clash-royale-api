package storageutil

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

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

func TestEnsureMigration(t *testing.T) {
	db := openMigrationTestDB(t)
	if _, err := db.Exec("CREATE TABLE migrations (name TEXT PRIMARY KEY, applied_at DATETIME NOT NULL)"); err != nil {
		t.Fatalf("failed to create migrations table: %v", err)
	}

	runs := 0
	run := func() error {
		runs++
		return nil
	}

	if err := EnsureMigration(db, "deck_hash_v1", run); err != nil {
		t.Fatalf("EnsureMigration returned error: %v", err)
	}
	if err := EnsureMigration(db, "deck_hash_v1", run); err != nil {
		t.Fatalf("EnsureMigration second run returned error: %v", err)
	}
	if runs != 1 {
		t.Fatalf("expected migration to run once, ran %d times", runs)
	}
}

func TestLoadDeckHashMigrationRows(t *testing.T) {
	rows := queryMigrationRowsFromValues(
		t,
		`["archers","giant"]`,
		`{"invalid":true}`,
		`["giant","archers"]`,
	)
	defer rows.Close()

	var warned []string
	records, winners, err := LoadDeckHashMigrationRows(rows, "deck row", func(row DeckHashMigrationRow, err error) {
		warned = append(warned, fmt.Sprintf("%d:%v", row.ID, err))
	})
	if err != nil {
		t.Fatalf("LoadDeckHashMigrationRows returned error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if len(warned) != 1 {
		t.Fatalf("expected 1 invalid-JSON warning, got %d", len(warned))
	}
	if records[1].Valid {
		t.Fatal("expected invalid JSON row to remain invalid")
	}
	if len(winners) != 1 {
		t.Fatalf("expected 1 canonical winner, got %d", len(winners))
	}
}

func queryMigrationRows(t *testing.T, cardsJSON string) *sql.Rows {
	t.Helper()
	return queryMigrationRowsFromValues(t, cardsJSON)
}

func queryMigrationRowsFromValues(t *testing.T, cardsJSON ...string) *sql.Rows {
	t.Helper()

	db := openMigrationTestDB(t)
	_, err := db.Exec(`CREATE TABLE decks (
		id INTEGER PRIMARY KEY,
		deck_hash TEXT NOT NULL,
		cards TEXT NOT NULL,
		overall_score REAL NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create decks table: %v", err)
	}

	for i, value := range cardsJSON {
		_, err = db.Exec(
			"INSERT INTO decks (id, deck_hash, cards, overall_score) VALUES (?, 'legacy', ?, ?)",
			i+1,
			value,
			42-i,
		)
		if err != nil {
			t.Fatalf("failed to insert deck row %d: %v", i+1, err)
		}
	}

	rows, err := db.Query("SELECT id, deck_hash, cards, overall_score FROM decks ORDER BY id ASC")
	if err != nil {
		t.Fatalf("failed to query deck rows: %v", err)
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
