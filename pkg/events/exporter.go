// Package events provides export functionality for event decks in multiple formats
package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormat represents the supported export formats
type ExportFormat string

const (
	FormatCSV       ExportFormat = "csv"
	FormatJSON      ExportFormat = "json"
	FormatDeckList  ExportFormat = "decklist"  // Simple deck list format
	FormatRoyaleAPI ExportFormat = "royaleapi" // RoyaleAPI deck format
)

// ExportOptions configures the export behavior
type ExportOptions struct {
	Format         ExportFormat `json:"format"`
	OutputDir      string       `json:"output_dir"`
	EventTypes     []EventType  `json:"event_types,omitempty"`  // Filter by event types
	StartDate      *time.Time   `json:"start_date,omitempty"`   // Filter by start date
	EndDate        *time.Time   `json:"end_date,omitempty"`     // Filter by end date
	MinBattles     int          `json:"min_battles,omitempty"`  // Minimum battles per deck
	MinWinRate     float64      `json:"min_win_rate,omitempty"` // Minimum win rate (0.0-1.0)
	IncludeBattles bool         `json:"include_battles"`        // Include detailed battle logs
	GroupByEvent   bool         `json:"group_by_event"`         // Group decks by event type
}

// DefaultExportOptions returns sensible defaults for exporting
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Format:         FormatCSV,
		OutputDir:      "./exports",
		MinBattles:     0,
		MinWinRate:     0.0,
		IncludeBattles: true,
		GroupByEvent:   false,
	}
}

// Exporter handles exporting event deck collections to various formats
type Exporter struct {
	options ExportOptions
}

// NewExporter creates a new event deck exporter
func NewExporter(options ExportOptions) *Exporter {
	return &Exporter{
		options: options,
	}
}

// Export performs the export operation using the configured options
func (e *Exporter) Export(collection *EventDeckCollection) error {
	// Apply filters to the collection
	filtered := e.applyFilters(collection)

	if len(filtered.Decks) == 0 {
		return fmt.Errorf("no decks match the specified filters")
	}

	// Group by event if requested
	if e.options.GroupByEvent {
		filtered = e.groupByEventType(filtered)
	}

	// Export based on format
	switch e.options.Format {
	case FormatCSV:
		return e.exportCSV(filtered)
	case FormatJSON:
		return e.exportJSON(filtered)
	case FormatDeckList:
		return e.exportDeckList(filtered)
	case FormatRoyaleAPI:
		return e.exportRoyaleAPI(filtered)
	default:
		return fmt.Errorf("unsupported export format: %s", e.options.Format)
	}
}

// applyFilters filters the event deck collection based on export options
func (e *Exporter) applyFilters(collection *EventDeckCollection) *EventDeckCollection {
	var filteredDecks []EventDeck

	for _, deck := range collection.Decks {
		// Filter by event type
		if len(e.options.EventTypes) > 0 {
			allowed := false
			for _, allowedType := range e.options.EventTypes {
				if deck.EventType == allowedType {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// Filter by date range
		if e.options.StartDate != nil && deck.StartTime.Before(*e.options.StartDate) {
			continue
		}
		if e.options.EndDate != nil && deck.StartTime.After(*e.options.EndDate) {
			continue
		}

		// Filter by minimum battles
		if e.options.MinBattles > 0 && deck.Performance.TotalBattles() < e.options.MinBattles {
			continue
		}

		// Filter by minimum win rate
		if e.options.MinWinRate > 0 && deck.Performance.WinRate < e.options.MinWinRate {
			continue
		}

		// Create a copy with optional battle exclusion
		deckCopy := deck
		if !e.options.IncludeBattles {
			deckCopy.Battles = []BattleRecord{}
		}

		filteredDecks = append(filteredDecks, deckCopy)
	}

	return &EventDeckCollection{
		PlayerTag:   collection.PlayerTag,
		Decks:       filteredDecks,
		LastUpdated: time.Now(),
	}
}

// groupByEventType creates separate collections for each event type
func (e *Exporter) groupByEventType(collection *EventDeckCollection) *EventDeckCollection {
	// For simplicity, we'll sort the decks by event type
	// The actual grouping will be handled in the export functions
	sortedDecks := make([]EventDeck, len(collection.Decks))
	copy(sortedDecks, collection.Decks)

	// Sort by event type
	for i := 0; i < len(sortedDecks)-1; i++ {
		for j := i + 1; j < len(sortedDecks); j++ {
			if string(sortedDecks[j].EventType) < string(sortedDecks[i].EventType) {
				sortedDecks[i], sortedDecks[j] = sortedDecks[j], sortedDecks[i]
			}
		}
	}

	return &EventDeckCollection{
		PlayerTag:   collection.PlayerTag,
		Decks:       sortedDecks,
		LastUpdated: time.Now(),
	}
}

// exportCSV exports the collection to CSV format
func (e *Exporter) exportCSV(collection *EventDeckCollection) (returnErr error) {
	if err := os.MkdirAll(e.options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("event_decks_%s.csv", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(e.options.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("failed to close CSV file: %w", err)
		}
	}()

	// Write headers
	headers := []string{
		"Event ID", "Event Name", "Event Type", "Start Time", "End Time",
		"Deck Cards", "Average Elixir", "Total Battles", "Wins", "Losses",
		"Win Rate", "Current Streak", "Best Streak", "Crowns Earned",
		"Crowns Lost", "Progress", "Notes",
	}

	if _, err := file.WriteString(strings.Join(headers, ",") + "\n"); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write deck data
	currentEventType := ""
	for _, deck := range collection.Decks {
		// Add event type separator if grouping
		if e.options.GroupByEvent && string(deck.EventType) != currentEventType {
			currentEventType = string(deck.EventType)
			if _, err := file.WriteString(fmt.Sprintf("# Event Type: %s\n", currentEventType)); err != nil {
				return fmt.Errorf("failed to write event separator: %w", err)
			}
		}

		// Format deck cards
		cardNames := make([]string, len(deck.Deck.Cards))
		for i, card := range deck.Deck.Cards {
			cardNames[i] = fmt.Sprintf("%s(L%d)", card.Name, card.Level)
		}

		endTime := ""
		if deck.EndTime != nil {
			endTime = deck.EndTime.Format("2006-01-02 15:04:05")
		}

		row := []string{
			deck.EventID,
			deck.EventName,
			string(deck.EventType),
			deck.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			strings.Join(cardNames, "|"),
			fmt.Sprintf("%.1f", deck.Deck.AvgElixir),
			fmt.Sprintf("%d", deck.Performance.TotalBattles()),
			fmt.Sprintf("%d", deck.Performance.Wins),
			fmt.Sprintf("%d", deck.Performance.Losses),
			fmt.Sprintf("%.2f", deck.Performance.WinRate),
			fmt.Sprintf("%d", deck.Performance.CurrentStreak),
			fmt.Sprintf("%d", deck.Performance.BestStreak),
			fmt.Sprintf("%d", deck.Performance.CrownsEarned),
			fmt.Sprintf("%d", deck.Performance.CrownsLost),
			string(deck.Performance.Progress),
			deck.Notes,
		}

		csvRow := make([]string, len(row))
		for i, value := range row {
			if strings.Contains(value, ",") || strings.Contains(value, " ") {
				csvRow[i] = fmt.Sprintf(`"%s"`, value)
			} else {
				csvRow[i] = value
			}
		}

		if _, err := file.WriteString(strings.Join(csvRow, ",") + "\n"); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// exportJSON exports the collection to JSON format
func (e *Exporter) exportJSON(collection *EventDeckCollection) (returnErr error) {
	if err := os.MkdirAll(e.options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("event_decks_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(e.options.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("failed to close JSON file: %w", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(collection); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// exportDeckList exports decks in a simple deck list format
func (e *Exporter) exportDeckList(collection *EventDeckCollection) (returnErr error) {
	if err := os.MkdirAll(e.options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("decks_%s.txt", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(e.options.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create deck list file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("failed to close deck list file: %w", err)
		}
	}()

	currentEventType := ""
	for _, deck := range collection.Decks {
		// Add event header if grouping
		if e.options.GroupByEvent && string(deck.EventType) != currentEventType {
			currentEventType = string(deck.EventType)
			if _, err := file.WriteString(fmt.Sprintf("\n=== %s ===\n", currentEventType)); err != nil {
				return fmt.Errorf("failed to write event header: %w", err)
			}
		}

		// Write deck info
		if _, err := file.WriteString(fmt.Sprintf("\n%s - %s\n", deck.EventName, deck.StartTime.Format("2006-01-02"))); err != nil {
			return fmt.Errorf("failed to write deck header: %w", err)
		}

		if _, err := file.WriteString(fmt.Sprintf("Record: %dW-%dL (%.1f%% WR, %.1f avg elixir)\n",
			deck.Performance.Wins, deck.Performance.Losses,
			deck.Performance.WinRate*100, deck.Deck.AvgElixir)); err != nil {
			return fmt.Errorf("failed to write deck stats: %w", err)
		}

		// Write cards
		for i, card := range deck.Deck.Cards {
			if _, err := file.WriteString(fmt.Sprintf("%d. %s (Level %d, %d elixir)\n",
				i+1, card.Name, card.Level, card.ElixirCost)); err != nil {
				return fmt.Errorf("failed to write card: %w", err)
			}
		}
	}

	return nil
}

// exportRoyaleAPI exports decks in RoyaleAPI deck link format
func (e *Exporter) exportRoyaleAPI(collection *EventDeckCollection) (returnErr error) {
	if err := os.MkdirAll(e.options.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("deck_links_%s.txt", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(e.options.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create deck links file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("failed to close deck links file: %w", err)
		}
	}()

	currentEventType := ""
	for i, deck := range collection.Decks {
		// Add event header if grouping
		if e.options.GroupByEvent && string(deck.EventType) != currentEventType {
			currentEventType = string(deck.EventType)
			if _, err := file.WriteString(fmt.Sprintf("\n=== %s ===\n", currentEventType)); err != nil {
				return fmt.Errorf("failed to write event header: %w", err)
			}
		}

		// Generate deck link (this would require card IDs)
		cardIds := make([]string, len(deck.Deck.Cards))
		for i, card := range deck.Deck.Cards {
			cardIds[i] = fmt.Sprintf("%d", card.ID)
		}

		deckLink := fmt.Sprintf("https://royaleapi.com/decks/%s", strings.Join(cardIds, ";"))

		if _, err := file.WriteString(fmt.Sprintf("%d. %s\n", i+1, deck.EventName)); err != nil {
			return fmt.Errorf("failed to write deck name: %w", err)
		}

		if _, err := file.WriteString(fmt.Sprintf("   %s\n", deckLink)); err != nil {
			return fmt.Errorf("failed to write deck link: %w", err)
		}

		if _, err := file.WriteString(fmt.Sprintf("   Record: %dW-%dL (%.1f%% WR)\n",
			deck.Performance.Wins, deck.Performance.Losses,
			deck.Performance.WinRate*100)); err != nil {
			return fmt.Errorf("failed to write deck record: %w", err)
		}

		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}

	return nil
}
