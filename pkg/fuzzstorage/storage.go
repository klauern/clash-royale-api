// Package fuzzstorage provides persistent storage for deck fuzzing results.
// Unlike the leaderboard storage which is per-player, this stores top decks
// from fuzzing runs for reuse in subsequent runs.
package fuzzstorage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/closeutil"
	"github.com/klauer/clash-royale-api/go/pkg/deckhash"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const defaultDBName = "fuzz_top_decks.db"

// Storage provides persistent storage for top decks from fuzzing runs
type Storage struct {
	db     *sql.DB
	dbPath string
}

// NewStorage creates a new Storage instance for fuzzing results
// The database file is stored at ~/.cr-api/fuzz_top_decks.db by default
func NewStorage(dbPath string) (*Storage, error) {
	if dbPath == "" {
		// Use default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		fuzzDir := filepath.Join(homeDir, ".cr-api")
		dbPath = filepath.Join(fuzzDir, defaultDBName)

		// Ensure directory exists
		if err := os.MkdirAll(fuzzDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create .cr-api directory: %w", err)
		}
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{
		db:     db,
		dbPath: dbPath,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		closeutil.CloseWithLog("fuzzstorage", db, "fuzz storage database")
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// GetDBPath returns the path to the SQLite database file
func (s *Storage) GetDBPath() string {
	return s.dbPath
}

// initSchema creates the database schema if it doesn't exist
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS top_decks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		deck_hash TEXT NOT NULL UNIQUE,
		cards TEXT NOT NULL,
		overall_score REAL NOT NULL,
		attack_score REAL NOT NULL,
		defense_score REAL NOT NULL,
		synergy_score REAL NOT NULL,
		versatility_score REAL NOT NULL,
		avg_elixir REAL NOT NULL,
		archetype TEXT NOT NULL,
		archetype_conf REAL NOT NULL,
		evaluated_at DATETIME NOT NULL,
		run_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_overall_score ON top_decks(overall_score DESC);
	CREATE INDEX IF NOT EXISTS idx_archetype ON top_decks(archetype);
	CREATE INDEX IF NOT EXISTS idx_evaluated_at ON top_decks(evaluated_at DESC);
	`

	_, err := s.db.Exec(schema)
	return err
}

// DeckEntry represents a stored deck from fuzzing
type DeckEntry struct {
	ID               int
	Cards            []string
	OverallScore     float64
	AttackScore      float64
	DefenseScore     float64
	SynergyScore     float64
	VersatilityScore float64
	AvgElixir        float64
	Archetype        string
	ArchetypeConf    float64
	EvaluatedAt      time.Time
	RunID            string
}

// SaveTopDecks saves the top N decks from a fuzzing run
// Returns the number of decks saved
func (s *Storage) SaveTopDecks(decks []DeckEntry) (int, error) {
	saved := 0

	for _, deck := range decks {
		_, isNew, err := s.InsertDeck(&deck)
		if err != nil {
			return saved, fmt.Errorf("failed to save deck: %w", err)
		}
		if isNew {
			saved++
		}
	}

	return saved, nil
}

// InsertDeck inserts or updates a deck entry
// If a deck with the same cards exists (same hash), it updates if the new score is better
// Returns the deck ID and whether it was a new insert (true) or update (false)
func (s *Storage) InsertDeck(entry *DeckEntry) (int, bool, error) {
	// Compute deck hash for deduplication
	deckHash := deckhash.DeckHash(entry.Cards)

	// Serialize cards to JSON
	cardsJSON, err := json.Marshal(entry.Cards)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal cards: %w", err)
	}

	// Check if deck already exists
	var existingID int
	var existingScore float64
	err = s.db.QueryRow("SELECT id, overall_score FROM top_decks WHERE deck_hash = ?", deckHash).Scan(&existingID, &existingScore)

	if err == sql.ErrNoRows {
		// Insert new deck
		result, err := s.db.Exec(`
			INSERT INTO top_decks (
				deck_hash, cards, overall_score, attack_score, defense_score,
				synergy_score, versatility_score, avg_elixir,
				archetype, archetype_conf, evaluated_at, run_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			deckHash, string(cardsJSON), entry.OverallScore, entry.AttackScore,
			entry.DefenseScore, entry.SynergyScore, entry.VersatilityScore,
			entry.AvgElixir, entry.Archetype, entry.ArchetypeConf,
			entry.EvaluatedAt, entry.RunID,
		)
		if err != nil {
			return 0, false, fmt.Errorf("failed to insert deck: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return 0, false, fmt.Errorf("failed to get insert id: %w", err)
		}

		entry.ID = int(id)
		return int(id), true, nil
	} else if err != nil {
		return 0, false, fmt.Errorf("failed to check for existing deck: %w", err)
	}

	// Deck exists - update only if new score is better
	if entry.OverallScore > existingScore {
		_, err = s.db.Exec(`
			UPDATE top_decks SET
				overall_score = ?, attack_score = ?, defense_score = ?,
				synergy_score = ?, versatility_score = ?, avg_elixir = ?,
				archetype = ?, archetype_conf = ?, evaluated_at = ?, run_id = ?
			WHERE id = ?
		`,
			entry.OverallScore, entry.AttackScore, entry.DefenseScore,
			entry.SynergyScore, entry.VersatilityScore, entry.AvgElixir,
			entry.Archetype, entry.ArchetypeConf, entry.EvaluatedAt,
			entry.RunID, existingID,
		)
		if err != nil {
			return 0, false, fmt.Errorf("failed to update deck: %w", err)
		}

		entry.ID = existingID
		return existingID, false, nil
	}

	// Score not better, no update
	return existingID, false, nil
}

// UpdateDeck updates an existing deck entry by ID with new evaluation data.
func (s *Storage) UpdateDeck(entry *DeckEntry) error {
	if entry.ID <= 0 {
		return fmt.Errorf("invalid deck ID: %d", entry.ID)
	}

	_, err := s.db.Exec(`
		UPDATE top_decks SET
			overall_score = ?, attack_score = ?, defense_score = ?,
			synergy_score = ?, versatility_score = ?, avg_elixir = ?,
			archetype = ?, archetype_conf = ?, evaluated_at = ?, run_id = ?
		WHERE id = ?
	`,
		entry.OverallScore, entry.AttackScore, entry.DefenseScore,
		entry.SynergyScore, entry.VersatilityScore, entry.AvgElixir,
		entry.Archetype, entry.ArchetypeConf, entry.EvaluatedAt,
		entry.RunID, entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deck: %w", err)
	}

	return nil
}

// GetTopN retrieves the top N decks by overall score
func (s *Storage) GetTopN(n int) ([]DeckEntry, error) {
	query := `
		SELECT id, deck_hash, cards, overall_score, attack_score, defense_score,
		       synergy_score, versatility_score, avg_elixir, archetype, archetype_conf, evaluated_at, run_id
		FROM top_decks
		ORDER BY overall_score DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query top decks: %w", err)
	}
	defer closeutil.CloseWithLog("fuzzstorage", rows, "top decks rows")

	return s.scanRows(rows)
}

// GetByArchetype retrieves decks of a specific archetype
func (s *Storage) GetByArchetype(archetype string, limit int) ([]DeckEntry, error) {
	query := `
		SELECT id, deck_hash, cards, overall_score, attack_score, defense_score,
		       synergy_score, versatility_score, avg_elixir, archetype, archetype_conf, evaluated_at, run_id
		FROM top_decks
		WHERE archetype = ?
		ORDER BY overall_score DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, archetype, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query by archetype: %w", err)
	}
	defer closeutil.CloseWithLog("fuzzstorage", rows, "deck rows by archetype")

	return s.scanRows(rows)
}

// QueryOptions defines filtering options for querying decks
type QueryOptions struct {
	MinScore        float64
	MaxScore        float64
	Archetype       string
	MinAvgElixir    float64
	MaxAvgElixir    float64
	RequireAllCards []string
	RequireAnyCards []string
	ExcludeCards    []string
	Limit           int
	Offset          int
}

// Query retrieves deck entries based on the provided options
func (s *Storage) Query(opts QueryOptions) ([]DeckEntry, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT id, deck_hash, cards, overall_score, attack_score, defense_score,
		       synergy_score, versatility_score, avg_elixir, archetype, archetype_conf, evaluated_at, run_id
		FROM top_decks
		WHERE 1=1
	`)
	args := []any{}

	// Apply filters
	if opts.MinScore > 0 {
		query.WriteString(" AND overall_score >= ?")
		args = append(args, opts.MinScore)
	}
	if opts.MaxScore > 0 {
		query.WriteString(" AND overall_score <= ?")
		args = append(args, opts.MaxScore)
	}
	if opts.Archetype != "" {
		query.WriteString(" AND archetype = ?")
		args = append(args, opts.Archetype)
	}
	if opts.MinAvgElixir > 0 {
		query.WriteString(" AND avg_elixir >= ?")
		args = append(args, opts.MinAvgElixir)
	}
	if opts.MaxAvgElixir > 0 {
		query.WriteString(" AND avg_elixir <= ?")
		args = append(args, opts.MaxAvgElixir)
	}

	// Card filters
	if len(opts.RequireAllCards) > 0 {
		for _, card := range opts.RequireAllCards {
			query.WriteString(" AND cards LIKE ?")
			args = append(args, "%"+card+"%")
		}
	}
	if len(opts.RequireAnyCards) > 0 {
		var subQuery strings.Builder
		subQuery.WriteString(" AND (")
		for i, card := range opts.RequireAnyCards {
			if i > 0 {
				subQuery.WriteString(" OR ")
			}
			subQuery.WriteString("cards LIKE ?")
			args = append(args, "%"+card+"%")
		}
		subQuery.WriteString(")")
		query.WriteString(subQuery.String())
	}
	if len(opts.ExcludeCards) > 0 {
		for _, card := range opts.ExcludeCards {
			query.WriteString(" AND cards NOT LIKE ?")
			args = append(args, "%"+card+"%")
		}
	}

	query.WriteString(" ORDER BY overall_score DESC")

	// Apply limit and offset
	if opts.Limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query.WriteString(" OFFSET ?")
		args = append(args, opts.Offset)
	}

	rows, err := s.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query decks: %w", err)
	}
	defer closeutil.CloseWithLog("fuzzstorage", rows, "stats rows")

	return s.scanRows(rows)
}

// ArchetypeHistogram returns deck counts grouped by archetype for the given filters.
// Limit and offset are intentionally ignored so the histogram represents the full matching set.
//
//nolint:funlen,gocognit,gocyclo // Query assembly contains explicit filter combinations.
func (s *Storage) ArchetypeHistogram(opts QueryOptions) (map[string]int, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT archetype, COUNT(*) AS deck_count
		FROM top_decks
		WHERE 1=1
	`)
	args := []any{}

	if opts.MinScore > 0 {
		query.WriteString(" AND overall_score >= ?")
		args = append(args, opts.MinScore)
	}
	if opts.MaxScore > 0 {
		query.WriteString(" AND overall_score <= ?")
		args = append(args, opts.MaxScore)
	}
	if opts.Archetype != "" {
		query.WriteString(" AND archetype = ?")
		args = append(args, opts.Archetype)
	}
	if opts.MinAvgElixir > 0 {
		query.WriteString(" AND avg_elixir >= ?")
		args = append(args, opts.MinAvgElixir)
	}
	if opts.MaxAvgElixir > 0 {
		query.WriteString(" AND avg_elixir <= ?")
		args = append(args, opts.MaxAvgElixir)
	}

	if len(opts.RequireAllCards) > 0 {
		for _, card := range opts.RequireAllCards {
			query.WriteString(" AND cards LIKE ?")
			args = append(args, "%"+card+"%")
		}
	}
	if len(opts.RequireAnyCards) > 0 {
		var subQuery strings.Builder
		subQuery.WriteString(" AND (")
		for i, card := range opts.RequireAnyCards {
			if i > 0 {
				subQuery.WriteString(" OR ")
			}
			subQuery.WriteString("cards LIKE ?")
			args = append(args, "%"+card+"%")
		}
		subQuery.WriteString(")")
		query.WriteString(subQuery.String())
	}
	if len(opts.ExcludeCards) > 0 {
		for _, card := range opts.ExcludeCards {
			query.WriteString(" AND cards NOT LIKE ?")
			args = append(args, "%"+card+"%")
		}
	}

	query.WriteString(" GROUP BY archetype ORDER BY deck_count DESC, archetype ASC")

	rows, err := s.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query archetype histogram: %w", err)
	}
	defer closeutil.CloseWithLog("fuzzstorage", rows, "archetype histogram rows")

	histogram := make(map[string]int)
	for rows.Next() {
		var archetype string
		var count int
		if err := rows.Scan(&archetype, &count); err != nil {
			return nil, fmt.Errorf("failed to scan archetype histogram row: %w", err)
		}
		histogram[archetype] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating archetype histogram rows: %w", err)
	}

	return histogram, nil
}

// scanRows scans rows from a query into DeckEntry slices
func (s *Storage) scanRows(rows *sql.Rows) ([]DeckEntry, error) {
	entries := []DeckEntry{}
	for rows.Next() {
		var entry DeckEntry
		var cardsJSON string
		var runIDNull sql.NullString

		err := rows.Scan(
			&entry.ID, new(string), &cardsJSON, &entry.OverallScore,
			&entry.AttackScore, &entry.DefenseScore, &entry.SynergyScore,
			&entry.VersatilityScore, &entry.AvgElixir, &entry.Archetype,
			&entry.ArchetypeConf, &entry.EvaluatedAt, &runIDNull,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Unmarshal cards JSON
		if err := json.Unmarshal([]byte(cardsJSON), &entry.Cards); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cards: %w", err)
		}

		if runIDNull.Valid {
			entry.RunID = runIDNull.String
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// DeleteDeck removes a deck from storage by ID
func (s *Storage) DeleteDeck(id int) error {
	_, err := s.db.Exec("DELETE FROM top_decks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}
	return nil
}

// Clear removes all decks from storage
func (s *Storage) Clear() error {
	_, err := s.db.Exec("DELETE FROM top_decks")
	if err != nil {
		return fmt.Errorf("failed to clear decks: %w", err)
	}
	return nil
}

// Count returns the total number of decks in storage
func (s *Storage) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM top_decks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count decks: %w", err)
	}
	return count, nil
}
