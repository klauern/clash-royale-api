// Package storage provides path construction utilities for Clash Royale data storage.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Default data directory structure
const (
	DefaultDataDir     = "./data"
	StaticDir          = "static"
	PlayersDir         = "players"
	AnalysisDir        = "analysis"
	DecksDir           = "decks"
	EventDecksDir      = "event_decks"
	CSVDir             = "csv"
	CSVPlayersSubdir   = "players"
	CSVReferenceSubdir = "reference"
	CSVEventsSubdir    = "events"
	CSVAnalysisSubdir  = "analysis"
)

// PathBuilder constructs standardized file paths for data storage
type PathBuilder struct {
	BaseDir string
}

// NewPathBuilder creates a PathBuilder with the specified base directory
func NewPathBuilder(baseDir string) *PathBuilder {
	if baseDir == "" {
		baseDir = DefaultDataDir
	}
	return &PathBuilder{BaseDir: baseDir}
}

// GetStaticDir returns the static data directory path
func (pb *PathBuilder) GetStaticDir() string {
	return filepath.Join(pb.BaseDir, StaticDir)
}

// GetPlayersDir returns the players data directory path
func (pb *PathBuilder) GetPlayersDir() string {
	return filepath.Join(pb.BaseDir, PlayersDir)
}

// GetAnalysisDir returns the analysis data directory path
func (pb *PathBuilder) GetAnalysisDir() string {
	return filepath.Join(pb.BaseDir, AnalysisDir)
}

// GetDecksDir returns the deck recommendations directory path
func (pb *PathBuilder) GetDecksDir() string {
	return filepath.Join(pb.BaseDir, DecksDir)
}

// GetEventDecksDir returns the event decks directory path
func (pb *PathBuilder) GetEventDecksDir() string {
	return filepath.Join(pb.BaseDir, EventDecksDir)
}

// GetCSVDir returns the CSV export directory path
func (pb *PathBuilder) GetCSVDir() string {
	return filepath.Join(pb.BaseDir, CSVDir)
}

// GetCSVPlayersDir returns the CSV players subdirectory path
func (pb *PathBuilder) GetCSVPlayersDir() string {
	return filepath.Join(pb.GetCSVDir(), CSVPlayersSubdir)
}

// GetCSVReferenceDir returns the CSV reference subdirectory path
func (pb *PathBuilder) GetCSVReferenceDir() string {
	return filepath.Join(pb.GetCSVDir(), CSVReferenceSubdir)
}

// GetCSVEventsDir returns the CSV events subdirectory path
func (pb *PathBuilder) GetCSVEventsDir() string {
	return filepath.Join(pb.GetCSVDir(), CSVEventsSubdir)
}

// GetCSVAnalysisDir returns the CSV analysis subdirectory path
func (pb *PathBuilder) GetCSVAnalysisDir() string {
	return filepath.Join(pb.GetCSVDir(), CSVAnalysisSubdir)
}

// GetPlayerFilePath returns the file path for a player profile JSON
// Format: data/players/{playerTag}.json
func (pb *PathBuilder) GetPlayerFilePath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	return filepath.Join(pb.GetPlayersDir(), fmt.Sprintf("%s.json", sanitized))
}

// GetAnalysisFilePath returns the timestamped file path for an analysis result
// Format: data/analysis/YYYYMMDD_HHMMSS_analysis_{playerTag}.json
func (pb *PathBuilder) GetAnalysisFilePath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_analysis_%s.json", timestamp, sanitized)
	return filepath.Join(pb.GetAnalysisDir(), filename)
}

// GetDeckFilePath returns the timestamped file path for a deck recommendation
// Format: data/decks/YYYYMMDD_HHMMSS_deck_{playerTag}.json
func (pb *PathBuilder) GetDeckFilePath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_deck_%s.json", timestamp, sanitized)
	return filepath.Join(pb.GetDecksDir(), filename)
}

// GetEventDeckPlayerDir returns the player-specific event deck directory
// Format: data/event_decks/{playerTag}/
func (pb *PathBuilder) GetEventDeckPlayerDir(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	return filepath.Join(pb.GetEventDecksDir(), sanitized)
}

// GetEventDeckCollectionPath returns the path to a player's event deck collection file
// Format: data/event_decks/{playerTag}/collection.json
func (pb *PathBuilder) GetEventDeckCollectionPath(playerTag string) string {
	return filepath.Join(pb.GetEventDeckPlayerDir(playerTag), "collection.json")
}

// GetEventDeckTypeDir returns the directory for a specific event type
// Format: data/event_decks/{playerTag}/{eventType}/
func (pb *PathBuilder) GetEventDeckTypeDir(playerTag, eventType string) string {
	sanitized := SanitizePlayerTag(playerTag)
	return filepath.Join(pb.GetEventDecksDir(), sanitized, eventType)
}

// GetCSVPlayerExportPath returns the path for a player CSV export
// Format: data/csv/players/{playerTag}_player.csv
func (pb *PathBuilder) GetCSVPlayerExportPath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	filename := fmt.Sprintf("%s_player.csv", sanitized)
	return filepath.Join(pb.GetCSVPlayersDir(), filename)
}

// GetCSVCardsExportPath returns the path for a cards CSV export
// Format: data/csv/players/{playerTag}_cards.csv
func (pb *PathBuilder) GetCSVCardsExportPath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	filename := fmt.Sprintf("%s_cards.csv", sanitized)
	return filepath.Join(pb.GetCSVPlayersDir(), filename)
}

// GetCSVEventsExportPath returns the path for an events CSV export
// Format: data/csv/events/{playerTag}_events.csv
func (pb *PathBuilder) GetCSVEventsExportPath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	filename := fmt.Sprintf("%s_events.csv", sanitized)
	return filepath.Join(pb.GetCSVEventsDir(), filename)
}

// GetCSVAnalysisExportPath returns the path for an analysis CSV export
// Format: data/csv/analysis/{playerTag}_analysis.csv
func (pb *PathBuilder) GetCSVAnalysisExportPath(playerTag string) string {
	sanitized := SanitizePlayerTag(playerTag)
	filename := fmt.Sprintf("%s_analysis.csv", sanitized)
	return filepath.Join(pb.GetCSVAnalysisDir(), filename)
}

// SanitizePlayerTag removes # prefix and converts to safe filename
// Example: #R8QGUQRCV -> R8QGUQRCV
func SanitizePlayerTag(playerTag string) string {
	return strings.TrimPrefix(playerTag, "#")
}

// EnsureDataDirectories creates all standard data directories if they don't exist
func (pb *PathBuilder) EnsureDataDirectories() error {
	dirs := []string{
		pb.GetStaticDir(),
		pb.GetPlayersDir(),
		pb.GetAnalysisDir(),
		pb.GetDecksDir(),
		pb.GetEventDecksDir(),
		pb.GetCSVPlayersDir(),
		pb.GetCSVReferenceDir(),
		pb.GetCSVEventsDir(),
		pb.GetCSVAnalysisDir(),
	}

	for _, dir := range dirs {
		if err := EnsureDirectory(dir); err != nil {
			return err
		}
	}

	return nil
}

// GetLatestFile returns the most recently modified file matching a pattern in a directory
// Returns empty string if no matching files found
func (pb *PathBuilder) GetLatestFile(dirPath, pattern string) (string, error) {
	if !DirectoryExists(dirPath) {
		return "", nil
	}

	files, err := filepath.Glob(filepath.Join(dirPath, pattern))
	if err != nil {
		return "", fmt.Errorf("failed to glob files: %w", err)
	}

	if len(files) == 0 {
		return "", nil
	}

	// Find most recent file
	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		modTime := info.ModTime()

		// Compare modification times
		if latestFile == "" || modTime.After(latestTime) {
			latestFile = file
			latestTime = modTime
		}
	}

	return latestFile, nil
}
