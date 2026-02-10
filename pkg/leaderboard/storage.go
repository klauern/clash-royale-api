package leaderboard

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Storage provides persistent storage for deck leaderboards using SQLite
type Storage struct {
	db        *sql.DB
	playerTag string
	dbPath    string
}

// NewStorage creates a new Storage instance for the given player tag
// The database file is stored at ~/.cr-api/leaderboards/<player_tag>.db
func NewStorage(playerTag string) (*Storage, error) {
	// Construct path: ~/.cr-api/leaderboards/<player_tag>.db
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Remove # prefix from player tag for filename safety
	sanitizedTag := strings.TrimPrefix(playerTag, "#")
	leaderboardDir := filepath.Join(homeDir, ".cr-api", "leaderboards")
	dbPath := filepath.Join(leaderboardDir, fmt.Sprintf("%s.db", sanitizedTag))

	// Ensure directory exists
	if err := os.MkdirAll(leaderboardDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create leaderboards directory: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{
		db:        db,
		playerTag: playerTag,
		dbPath:    dbPath,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		closeWithLog(db, "leaderboard database")
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
	CREATE TABLE IF NOT EXISTS decks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		deck_hash TEXT NOT NULL UNIQUE,
		cards TEXT NOT NULL,
		overall_score REAL NOT NULL,
		attack_score REAL NOT NULL,
		defense_score REAL NOT NULL,
		synergy_score REAL NOT NULL,
		versatility_score REAL NOT NULL,
		f2p_score REAL NOT NULL,
		playability_score REAL NOT NULL,
		archetype TEXT NOT NULL,
		archetype_conf REAL NOT NULL,
		strategy TEXT,
		avg_elixir REAL NOT NULL,
		evaluated_at DATETIME NOT NULL,
		player_tag TEXT NOT NULL,
		evaluation_version TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_overall_score ON decks(overall_score DESC);
	CREATE INDEX IF NOT EXISTS idx_archetype ON decks(archetype);
	CREATE INDEX IF NOT EXISTS idx_strategy ON decks(strategy);
	CREATE INDEX IF NOT EXISTS idx_archetype_score ON decks(archetype, overall_score DESC);
	CREATE INDEX IF NOT EXISTS idx_strategy_score ON decks(strategy, overall_score DESC);
	CREATE INDEX IF NOT EXISTS idx_evaluated_at ON decks(evaluated_at DESC);

	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		player_tag TEXT NOT NULL,
		total_decks_evaluated INTEGER NOT NULL,
		total_unique_decks INTEGER NOT NULL,
		last_updated DATETIME NOT NULL,
		avg_eval_time_ms REAL NOT NULL,
		top_score REAL NOT NULL,
		avg_score REAL NOT NULL
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

// computeDeckHash computes a SHA256 hash of the sorted card names for deduplication
func computeDeckHash(cards []string) string {
	// Sort cards to ensure consistent hash regardless of order
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)

	// Compute SHA256 hash
	data := strings.Join(sorted, "|")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// InsertDeck inserts or updates a deck entry in the leaderboard
// If a deck with the same cards exists (same hash), it updates the existing entry
// Returns the deck ID and whether it was a new insert (true) or update (false)
func (s *Storage) InsertDeck(entry *DeckEntry) (int, bool, error) {
	// Compute deck hash for deduplication
	entry.DeckHash = computeDeckHash(entry.Cards)

	// Serialize cards to JSON
	cardsJSON, err := json.Marshal(entry.Cards)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal cards: %w", err)
	}

	// Check if deck already exists
	var existingID int
	err = s.db.QueryRow("SELECT id FROM decks WHERE deck_hash = ?", entry.DeckHash).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new deck
		result, err := s.db.Exec(`
			INSERT INTO decks (
				deck_hash, cards, overall_score, attack_score, defense_score,
				synergy_score, versatility_score, f2p_score, playability_score,
				archetype, archetype_conf, strategy, avg_elixir,
				evaluated_at, player_tag, evaluation_version
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			entry.DeckHash, string(cardsJSON), entry.OverallScore, entry.AttackScore,
			entry.DefenseScore, entry.SynergyScore, entry.VersatilityScore,
			entry.F2PScore, entry.PlayabilityScore, entry.Archetype,
			entry.ArchetypeConf, entry.Strategy, entry.AvgElixir,
			entry.EvaluatedAt, entry.PlayerTag, entry.EvaluationVersion,
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

	// Update existing deck
	_, err = s.db.Exec(`
		UPDATE decks SET
			overall_score = ?, attack_score = ?, defense_score = ?,
			synergy_score = ?, versatility_score = ?, f2p_score = ?,
			playability_score = ?, archetype = ?, archetype_conf = ?,
			strategy = ?, avg_elixir = ?, evaluated_at = ?,
			evaluation_version = ?
		WHERE id = ?
	`,
		entry.OverallScore, entry.AttackScore, entry.DefenseScore,
		entry.SynergyScore, entry.VersatilityScore, entry.F2PScore,
		entry.PlayabilityScore, entry.Archetype, entry.ArchetypeConf,
		entry.Strategy, entry.AvgElixir, entry.EvaluatedAt,
		entry.EvaluationVersion, existingID,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to update deck: %w", err)
	}

	entry.ID = existingID
	return existingID, false, nil
}

// Query retrieves deck entries based on the provided options
func (s *Storage) Query(opts QueryOptions) ([]DeckEntry, error) {
	query, args := buildDeckQuery(opts)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query decks: %w", err)
	}
	defer closeWithLog(rows, "deck rows")

	entries, err := scanDeckEntries(rows)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// buildDeckQuery constructs the SQL query and arguments from query options
func buildDeckQuery(opts QueryOptions) (string, []interface{}) {
	query := "SELECT id, deck_hash, cards, overall_score, attack_score, defense_score, synergy_score, versatility_score, f2p_score, playability_score, archetype, archetype_conf, strategy, avg_elixir, evaluated_at, player_tag, evaluation_version FROM decks WHERE 1=1"
	args := []interface{}{}

	query, args = applyScoreFilters(query, args, opts)
	query, args = applyMetadataFilters(query, args, opts)
	query, args = applyCardFilters(query, args, opts)
	query = applySortingAndPagination(query, &args, opts)

	return query, args
}

// applyScoreFilters adds score-based filters to the query
func applyScoreFilters(query string, args []interface{}, opts QueryOptions) (string, []interface{}) {
	if opts.MinScore > 0 {
		query += " AND overall_score >= ?"
		args = append(args, opts.MinScore)
	}
	if opts.MaxScore > 0 {
		query += " AND overall_score <= ?"
		args = append(args, opts.MaxScore)
	}
	return query, args
}

// applyMetadataFilters adds archetype, strategy, and elixir filters
func applyMetadataFilters(query string, args []interface{}, opts QueryOptions) (string, []interface{}) {
	if opts.Archetype != "" {
		query += " AND archetype = ?"
		args = append(args, opts.Archetype)
	}
	if opts.Strategy != "" {
		query += " AND strategy = ?"
		args = append(args, opts.Strategy)
	}
	if opts.MinAvgElixir > 0 {
		query += " AND avg_elixir >= ?"
		args = append(args, opts.MinAvgElixir)
	}
	if opts.MaxAvgElixir > 0 {
		query += " AND avg_elixir <= ?"
		args = append(args, opts.MaxAvgElixir)
	}
	return query, args
}

// applyCardFilters adds card-based filters (require all, any, exclude)
func applyCardFilters(query string, args []interface{}, opts QueryOptions) (string, []interface{}) {
	query, args = applyRequireAllCards(query, args, opts.RequireAllCards)
	query, args = applyRequireAnyCards(query, args, opts.RequireAnyCards)
	query, args = applyExcludeCards(query, args, opts.ExcludeCards)
	return query, args
}

// applyRequireAllCards adds filters for cards that must all be present
func applyRequireAllCards(query string, args []interface{}, cards []string) (string, []interface{}) {
	for _, card := range cards {
		query += " AND cards LIKE ?"
		args = append(args, "%"+card+"%")
	}
	return query, args
}

// applyRequireAnyCards adds filters for cards where at least one must be present
func applyRequireAnyCards(query string, args []interface{}, cards []string) (string, []interface{}) {
	if len(cards) == 0 {
		return query, args
	}

	subQuery := " AND ("
	for i, card := range cards {
		if i > 0 {
			subQuery += " OR "
		}
		subQuery += "cards LIKE ?"
		args = append(args, "%"+card+"%")
	}
	subQuery += ")"
	query += subQuery
	return query, args
}

// applyExcludeCards adds filters for cards that must not be present
func applyExcludeCards(query string, args []interface{}, cards []string) (string, []interface{}) {
	for _, card := range cards {
		query += " AND cards NOT LIKE ?"
		args = append(args, "%"+card+"%")
	}
	return query, args
}

// applySortingAndPagination adds ORDER BY, LIMIT, and OFFSET clauses
func applySortingAndPagination(query string, args *[]interface{}, opts QueryOptions) string {
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "overall_score"
	}
	sortOrder := strings.ToUpper(opts.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if opts.Limit > 0 {
		query += " LIMIT ?"
		*args = append(*args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		*args = append(*args, opts.Offset)
	}

	return query
}

// scanDeckEntries scans database rows into DeckEntry structs
func scanDeckEntries(rows *sql.Rows) ([]DeckEntry, error) {
	entries := []DeckEntry{}
	for rows.Next() {
		entry, err := scanSingleDeckEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// scanSingleDeckEntry scans a single row into a DeckEntry
func scanSingleDeckEntry(rows *sql.Rows) (DeckEntry, error) {
	var entry DeckEntry
	var cardsJSON string
	var strategyNull sql.NullString

	err := rows.Scan(
		&entry.ID, &entry.DeckHash, &cardsJSON, &entry.OverallScore,
		&entry.AttackScore, &entry.DefenseScore, &entry.SynergyScore,
		&entry.VersatilityScore, &entry.F2PScore, &entry.PlayabilityScore,
		&entry.Archetype, &entry.ArchetypeConf, &strategyNull,
		&entry.AvgElixir, &entry.EvaluatedAt, &entry.PlayerTag,
		&entry.EvaluationVersion,
	)
	if err != nil {
		return DeckEntry{}, fmt.Errorf("failed to scan row: %w", err)
	}

	if err := json.Unmarshal([]byte(cardsJSON), &entry.Cards); err != nil {
		return DeckEntry{}, fmt.Errorf("failed to unmarshal cards: %w", err)
	}

	if strategyNull.Valid {
		entry.Strategy = strategyNull.String
	}

	return entry, nil
}

// GetTopN retrieves the top N decks by overall score
func (s *Storage) GetTopN(n int) ([]DeckEntry, error) {
	return s.Query(TopNQueryOptions(n))
}

// GetByArchetype retrieves decks of a specific archetype
func (s *Storage) GetByArchetype(archetype string, limit int) ([]DeckEntry, error) {
	return s.Query(ArchetypeQueryOptions(archetype, limit))
}

// DeleteDeck removes a deck from the leaderboard by ID
func (s *Storage) DeleteDeck(id int) error {
	_, err := s.db.Exec("DELETE FROM decks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}
	return nil
}

// Clear removes all decks from the leaderboard
func (s *Storage) Clear() error {
	_, err := s.db.Exec("DELETE FROM decks")
	if err != nil {
		return fmt.Errorf("failed to clear decks: %w", err)
	}
	return nil
}

// Vacuum compacts the SQLite database file.
func (s *Storage) Vacuum() error {
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}
	return nil
}

// CleanupOptions controls filtered deck deletion.
type CleanupOptions struct {
	MinScore  float64
	OlderThan time.Time
	Archetype string
}

// Cleanup deletes decks matching the provided filters and returns rows deleted.
// At least one filter must be set.
func (s *Storage) Cleanup(opts CleanupOptions) (int64, error) {
	query := "DELETE FROM decks WHERE 1=1"
	args := make([]any, 0, 3)
	filters := 0

	if opts.MinScore > 0 {
		query += " AND overall_score < ?"
		args = append(args, opts.MinScore)
		filters++
	}
	if !opts.OlderThan.IsZero() {
		query += " AND evaluated_at < ?"
		args = append(args, opts.OlderThan)
		filters++
	}
	if strings.TrimSpace(opts.Archetype) != "" {
		query += " AND archetype = ?"
		args = append(args, strings.TrimSpace(opts.Archetype))
		filters++
	}

	if filters == 0 {
		return 0, fmt.Errorf("at least one cleanup filter is required")
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup decks: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read cleanup row count: %w", err)
	}
	return deleted, nil
}

// PruneTopNPerArchetype keeps only the top N scored decks per archetype.
// Returns the number of decks deleted.
func (s *Storage) PruneTopNPerArchetype(n int) (int64, error) {
	if n < 1 {
		return 0, fmt.Errorf("n must be >= 1")
	}

	result, err := s.db.Exec(`
		DELETE FROM decks
		WHERE id IN (
			SELECT id
			FROM (
				SELECT id,
				       ROW_NUMBER() OVER (
				           PARTITION BY archetype
				           ORDER BY overall_score DESC, id ASC
				       ) AS rank_in_archetype
				FROM decks
			)
			WHERE rank_in_archetype > ?
		)
	`, n)
	if err != nil {
		return 0, fmt.Errorf("failed to prune decks: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read prune row count: %w", err)
	}
	return deleted, nil
}

// ExportJSON writes all stored decks as a JSON array to the given file path.
// Returns the number of exported decks.
func (s *Storage) ExportJSON(path string) (int, error) {
	decks, err := s.Query(QueryOptions{
		SortBy:    "overall_score",
		SortOrder: "desc",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to load decks for export: %w", err)
	}

	data, err := json.MarshalIndent(decks, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal export data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return 0, fmt.Errorf("failed to write export file: %w", err)
	}

	return len(decks), nil
}

// ImportJSON loads deck entries from a JSON array file.
// Returns inserted and updated counts.
func (s *Storage) ImportJSON(path string) (int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read import file: %w", err)
	}

	var decks []DeckEntry
	if err := json.Unmarshal(data, &decks); err != nil {
		return 0, 0, fmt.Errorf("failed to parse import file: %w", err)
	}

	inserted := 0
	updated := 0
	for i := range decks {
		entry := decks[i]
		entry.ID = 0
		entry.DeckHash = ""
		if entry.EvaluatedAt.IsZero() {
			entry.EvaluatedAt = time.Now()
		}
		if entry.PlayerTag == "" {
			entry.PlayerTag = s.playerTag
		}
		if entry.EvaluationVersion == "" {
			entry.EvaluationVersion = "imported"
		}

		_, isNew, err := s.InsertDeck(&entry)
		if err != nil {
			return inserted, updated, fmt.Errorf("failed importing deck %d: %w", i+1, err)
		}
		if isNew {
			inserted++
		} else {
			updated++
		}
	}

	return inserted, updated, nil
}

// GetStats retrieves the current leaderboard statistics
func (s *Storage) GetStats() (*LeaderboardStats, error) {
	var stats LeaderboardStats
	err := s.db.QueryRow(`
		SELECT player_tag, total_decks_evaluated, total_unique_decks,
		       last_updated, avg_eval_time_ms, top_score, avg_score
		FROM stats WHERE id = 1
	`).Scan(
		&stats.PlayerTag, &stats.TotalDecksEvaluated, &stats.TotalUniqueDecks,
		&stats.LastUpdated, &stats.AvgEvalTimeMs, &stats.TopScore, &stats.AvgScore,
	)

	if err == sql.ErrNoRows {
		// No stats yet, return empty stats
		return &LeaderboardStats{
			PlayerTag:   s.playerTag,
			LastUpdated: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}

// UpdateStats updates the leaderboard statistics
// This should be called after inserting/updating decks
func (s *Storage) UpdateStats(stats *LeaderboardStats) error {
	_, err := s.db.Exec(`
		INSERT INTO stats (id, player_tag, total_decks_evaluated, total_unique_decks, last_updated, avg_eval_time_ms, top_score, avg_score)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			player_tag = excluded.player_tag,
			total_decks_evaluated = excluded.total_decks_evaluated,
			total_unique_decks = excluded.total_unique_decks,
			last_updated = excluded.last_updated,
			avg_eval_time_ms = excluded.avg_eval_time_ms,
			top_score = excluded.top_score,
			avg_score = excluded.avg_score
	`,
		stats.PlayerTag, stats.TotalDecksEvaluated, stats.TotalUniqueDecks,
		stats.LastUpdated, stats.AvgEvalTimeMs, stats.TopScore, stats.AvgScore,
	)
	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}
	return nil
}

// RecalculateStats computes statistics from the current deck data
// Returns the updated stats
func (s *Storage) RecalculateStats() (*LeaderboardStats, error) {
	var stats LeaderboardStats
	stats.PlayerTag = s.playerTag
	stats.LastUpdated = time.Now()

	// Count unique decks
	err := s.db.QueryRow("SELECT COUNT(*) FROM decks").Scan(&stats.TotalUniqueDecks)
	if err != nil {
		return nil, fmt.Errorf("failed to count decks: %w", err)
	}

	// For now, assume total evaluated = total unique (will be updated by discovery system)
	stats.TotalDecksEvaluated = stats.TotalUniqueDecks

	// Calculate top and average scores
	if stats.TotalUniqueDecks > 0 {
		err = s.db.QueryRow("SELECT MAX(overall_score), AVG(overall_score) FROM decks").Scan(&stats.TopScore, &stats.AvgScore)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate score stats: %w", err)
		}
	}

	// Update stats in database
	if err := s.UpdateStats(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// Count returns the total number of decks in the leaderboard
func (s *Storage) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM decks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count decks: %w", err)
	}
	return count, nil
}

// GetArchetypeCounts returns deck counts grouped by archetype, ordered by count descending.
func (s *Storage) GetArchetypeCounts() ([]ArchetypeCount, error) {
	rows, err := s.db.Query(`
		SELECT archetype, COUNT(*) AS deck_count
		FROM decks
		GROUP BY archetype
		ORDER BY deck_count DESC, archetype ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query archetype counts: %w", err)
	}
	defer closeWithLog(rows, "archetype count rows")

	counts := make([]ArchetypeCount, 0)
	for rows.Next() {
		var c ArchetypeCount
		if err := rows.Scan(&c.Archetype, &c.Count); err != nil {
			return nil, fmt.Errorf("failed to scan archetype count row: %w", err)
		}
		counts = append(counts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating archetype count rows: %w", err)
	}
	return counts, nil
}
