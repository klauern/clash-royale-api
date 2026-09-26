// Package storage provides path construction utilities for Clash Royale data storage.
package storage

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/playertag"
)

// Default data directory structure
const (
	DefaultDataDir      = "./data"
	StaticDir           = "static"
	AnalysisDir         = "analysis"
	CSVDir              = "csv"
	CSVPlayersSubdir    = "players"
	CSVReferenceSubdir  = "reference"
	CSVEventsSubdir     = "events"
	CSVAnalysisSubdir   = "analysis"
	CSVBattlesSubdir    = "battles"
	CSVArchetypesSubdir = "archetypes"
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

// GetStaticCardsPath returns the path to the cached card database JSON.
func (pb *PathBuilder) GetStaticCardsPath() string {
	return filepath.Join(pb.GetStaticDir(), "cards.json")
}

// GetAnalysisDir returns the analysis data directory path
func (pb *PathBuilder) GetAnalysisDir() string {
	return filepath.Join(pb.BaseDir, AnalysisDir)
}

// GetEvolutionShardsPath returns the path to the evolution shard inventory file.
func (pb *PathBuilder) GetEvolutionShardsPath() string {
	return filepath.Join(pb.BaseDir, "evolution_shards.json")
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

// GetAnalysisFilePath returns the timestamped file path for an analysis result
// Format: data/analysis/YYYYMMDD_HHMMSS_analysis_{playerTag}.json
func (pb *PathBuilder) GetAnalysisFilePath(playerTag string) (string, error) {
	sanitized, err := playertag.Sanitize(playerTag)
	if err != nil {
		return "", err
	}
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_analysis_%s.json", timestamp, sanitized)
	return filepath.Join(pb.GetAnalysisDir(), filename), nil
}
