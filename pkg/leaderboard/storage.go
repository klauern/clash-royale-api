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
	// Build SQL query dynamically based on filters
	query := "SELECT id, deck_hash, cards, overall_score, attack_score, defense_score, synergy_score, versatility_score, f2p_score, playability_score, archetype, archetype_conf, strategy, avg_elixir, evaluated_at, player_tag, evaluation_version FROM decks WHERE 1=1"
	args := []interface{}{}

	// Apply filters
	if opts.MinScore > 0 {
		query += " AND overall_score >= ?"
		args = append(args, opts.MinScore)
	}
	if opts.MaxScore > 0 {
		query += " AND overall_score <= ?"
		args = append(args, opts.MaxScore)
	}
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

	// Card filters (require all, any, or exclude)
	if len(opts.RequireAllCards) > 0 {
		for _, card := range opts.RequireAllCards {
			query += " AND cards LIKE ?"
			args = append(args, "%"+card+"%")
		}
	}
	if len(opts.RequireAnyCards) > 0 {
		subQuery := " AND ("
		for i, card := range opts.RequireAnyCards {
			if i > 0 {
				subQuery += " OR "
			}
			subQuery += "cards LIKE ?"
			args = append(args, "%"+card+"%")
		}
		subQuery += ")"
		query += subQuery
	}
	if len(opts.ExcludeCards) > 0 {
		for _, card := range opts.ExcludeCards {
			query += " AND cards NOT LIKE ?"
			args = append(args, "%"+card+"%")
		}
	}

	// Apply sorting
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "overall_score"
	}
	sortOrder := strings.ToUpper(opts.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply limit and offset
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query decks: %w", err)
	}
	defer closeWithLog(rows, "deck rows")

	// Parse results
	entries := []DeckEntry{}
	for rows.Next() {
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
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Unmarshal cards JSON
		if err := json.Unmarshal([]byte(cardsJSON), &entry.Cards); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cards: %w", err)
		}

		if strategyNull.Valid {
			entry.Strategy = strategyNull.String
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
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
