package storageutil

import (
	"database/sql"
	"encoding/json"
	"errors"
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
		if errors.As(err, &typeErr) {
			return
		}
		t.Fatalf("expected wrapped json unmarshal type error, got: %T", err)
	})
}

func queryMigrationRows(t *testing.T, cardsJSON string) *sql.Rows {
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

	if _, err := db.Exec(`CREATE TABLE migration_rows (id INTEGER, deck_hash TEXT, cards TEXT, overall_score REAL)`); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO migration_rows (id, deck_hash, cards, overall_score) VALUES (1, 'legacy', ?, 9.5)`, cardsJSON); err != nil {
		t.Fatalf("failed to insert row: %v", err)
	}

	rows, err := db.Query(`SELECT id, deck_hash, cards, overall_score FROM migration_rows`)
	if err != nil {
		t.Fatalf("failed to query migration rows: %v", err)
	}
	return rows
}
