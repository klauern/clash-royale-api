// Package events provides types and utilities for tracking Clash Royale event decks
// (challenges, tournaments, special events) with performance metrics.
package events

import (
	"fmt"
	"time"
)

// EventAnalysis represents comprehensive analysis of event decks
type EventAnalysis struct {
	PlayerTag      string                `json:"player_tag"`
	AnalysisTime   time.Time             `json:"analysis_time"`
	TotalDecks     int                   `json:"total_decks"`
	Summary        EventSummary          `json:"summary"`
	CardAnalysis   EventCardAnalysis     `json:"card_analysis"`
	ElixirAnalysis EventElixirAnalysis   `json:"elixir_analysis"`
	EventBreakdown map[string]EventStats `json:"event_breakdown"`
	TopDecks       []TopPerformingDeck   `json:"top_decks"`
}

// EventSummary provides high-level statistics about event deck performance
type EventSummary struct {
	TotalBattles       int     `json:"total_battles"`
	TotalWins          int     `json:"total_wins"`
	TotalLosses        int     `json:"total_losses"`
	OverallWinRate     float64 `json:"overall_win_rate"`
	AvgCrownsPerBattle float64 `json:"avg_crowns_per_battle"`
	AvgDeckElixir      float64 `json:"avg_deck_elixir"`
}

// EventCardAnalysis provides card usage and performance statistics
type EventCardAnalysis struct {
	MostUsedCards       []CardUsage   `json:"most_used_cards"`
	HighestWinRateCards []CardWinRate `json:"highest_win_rate_cards"`
	TotalUniqueCards    int           `json:"total_unique_cards"`
}

// CardUsage tracks how frequently a card appears across event decks
type CardUsage struct {
	CardName string `json:"card_name"`
	Count    int    `json:"count"`
}

// CardWinRate tracks a card's win rate across all event decks
type CardWinRate struct {
	CardName string  `json:"card_name"`
	WinRate  float64 `json:"win_rate"`
}

// CardConstraintSuggestion represents a suggested card constraint based on frequency analysis
type CardConstraintSuggestion struct {
	CardName    string  `json:"card_name"`
	Appearances int     `json:"appearances"`
	Percentage  float64 `json:"percentage"`
	TotalDecks  int     `json:"total_decks"`
}

// EventElixirAnalysis provides deck elixir cost performance analysis
type EventElixirAnalysis struct {
	LowElixir  ElixirRangeStats `json:"low_elixir"`
	MidElixir  ElixirRangeStats `json:"mid_elixir"`
	HighElixir ElixirRangeStats `json:"high_elixir"`
}

// ElixirRangeStats provides statistics for a specific elixir cost range
type ElixirRangeStats struct {
	Range      string  `json:"range"`
	DeckCount  int     `json:"deck_count"`
	AvgWinRate float64 `json:"avg_win_rate"`
}

// EventStats provides statistics for a specific event type
type EventStats struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
	Wins      int    `json:"wins"`
	Losses    int    `json:"losses"`
}

// TopPerformingDeck represents a high-performing event deck
type TopPerformingDeck struct {
	EventName string   `json:"event_name"`
	EventType string   `json:"event_type"`
	WinRate   float64  `json:"win_rate"`
	Record    string   `json:"record"` // e.g., "12W-2L"
	Deck      []string `json:"deck"`
	AvgElixir float64  `json:"avg_elixir"`
}

// AnalysisOptions configures event deck analysis behavior
type AnalysisOptions struct {
	MinBattlesForTopDecks int      `json:"min_battles_for_top_decks"` // Minimum battles to qualify for top decks
	LimitTopDecks         int      `json:"limit_top_decks"`           // Maximum number of top decks to include
	EventTypes            []string `json:"event_types"`               // Filter by event types (empty = all)
}

// DefaultAnalysisOptions returns sensible defaults for event analysis
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		MinBattlesForTopDecks: 3,
		LimitTopDecks:         5,
		EventTypes:            []string{},
	}
}

// AnalyzeEventDecks performs comprehensive analysis on a collection of event decks
// and returns detailed performance insights.
func AnalyzeEventDecks(decks []EventDeck, options AnalysisOptions) *EventAnalysis {
	// Filter decks by event type if specified
	filteredDecks := filterDecksByEventType(decks, options.EventTypes)

	if len(filteredDecks) == 0 {
		return &EventAnalysis{
			AnalysisTime:   time.Now(),
			Summary:        EventSummary{},
			CardAnalysis:   EventCardAnalysis{},
			ElixirAnalysis: EventElixirAnalysis{},
			EventBreakdown: map[string]EventStats{},
			TopDecks:       []TopPerformingDeck{},
		}
	}

	// Calculate basic statistics
	summary := calculateSummary(filteredDecks)
	cardAnalysis := analyzeCardUsage(filteredDecks)
	elixirAnalysis := analyzeElixirPerformance(filteredDecks)
	eventBreakdown := calculateEventBreakdown(filteredDecks)
	topDecks := identifyTopPerformingDecks(filteredDecks, options)

	// Determine player tag from first deck (all decks should have same player)
	playerTag := ""
	if len(filteredDecks) > 0 {
		playerTag = filteredDecks[0].PlayerTag
	}

	return &EventAnalysis{
		PlayerTag:      playerTag,
		AnalysisTime:   time.Now(),
		TotalDecks:     len(filteredDecks),
		Summary:        summary,
		CardAnalysis:   cardAnalysis,
		ElixirAnalysis: elixirAnalysis,
		EventBreakdown: eventBreakdown,
		TopDecks:       topDecks,
	}
}

// filterDecksByEventType filters decks to only include specified event types
func filterDecksByEventType(decks []EventDeck, eventTypes []string) []EventDeck {
	if len(eventTypes) == 0 {
		return decks
	}

	// Convert event types to a map for fast lookup
	allowedTypes := make(map[string]bool)
	for _, et := range eventTypes {
		allowedTypes[et] = true
	}

	var filtered []EventDeck
	for _, deck := range decks {
		if allowedTypes[string(deck.EventType)] {
			filtered = append(filtered, deck)
		}
	}
	return filtered
}

// calculateSummary computes overall performance statistics across all event decks
func calculateSummary(decks []EventDeck) EventSummary {
	summary := EventSummary{}

	for _, deck := range decks {
		perf := deck.Performance
		summary.TotalBattles += perf.TotalBattles()
		summary.TotalWins += perf.Wins
		summary.TotalLosses += perf.Losses
		summary.AvgCrownsPerBattle += float64(perf.CrownsEarned-perf.CrownsLost) / float64(perf.TotalBattles())
		summary.AvgDeckElixir += deck.Deck.AvgElixir
	}

	// Calculate win rate
	if summary.TotalBattles > 0 {
		summary.OverallWinRate = float64(summary.TotalWins) / float64(summary.TotalBattles)
		summary.AvgCrownsPerBattle /= float64(len(decks))
		summary.AvgDeckElixir /= float64(len(decks))
	}

	return summary
}

// analyzeCardUsage tracks card frequency and performance across all event decks
func analyzeCardUsage(decks []EventDeck) EventCardAnalysis {
	cardUsage := make(map[string]int)   // card name -> usage count
	cardWins := make(map[string]int)    // card name -> total wins with this card
	cardBattles := make(map[string]int) // card name -> total battles with this card

	// Track all unique cards seen
	uniqueCards := make(map[string]bool)

	for _, deck := range decks {
		for _, card := range deck.Deck.Cards {
			cardName := card.Name
			cardUsage[cardName]++
			uniqueCards[cardName] = true

			// Add battle stats for this card
			cardWins[cardName] += deck.Performance.Wins
			cardBattles[cardName] += deck.Performance.TotalBattles()
		}
	}

	// Convert to slices and sort
	var mostUsed []CardUsage
	for name, count := range cardUsage {
		mostUsed = append(mostUsed, CardUsage{CardName: name, Count: count})
	}

	// Sort by count descending
	for i := 0; i < len(mostUsed)-1; i++ {
		for j := i + 1; j < len(mostUsed); j++ {
			if mostUsed[j].Count > mostUsed[i].Count {
				mostUsed[i], mostUsed[j] = mostUsed[j], mostUsed[i]
			}
		}
	}

	// Limit top 10 most used cards
	if len(mostUsed) > 10 {
		mostUsed = mostUsed[:10]
	}

	// Calculate win rates
	var highestWinRate []CardWinRate
	for name, wins := range cardWins {
		battles := cardBattles[name]
		if battles >= 3 { // Only include cards with sufficient sample size
			winRate := float64(wins) / float64(battles)
			highestWinRate = append(highestWinRate, CardWinRate{
				CardName: name,
				WinRate:  winRate,
			})
		}
	}

	// Sort by win rate descending
	for i := 0; i < len(highestWinRate)-1; i++ {
		for j := i + 1; j < len(highestWinRate); j++ {
			if highestWinRate[j].WinRate > highestWinRate[i].WinRate {
				highestWinRate[i], highestWinRate[j] = highestWinRate[j], highestWinRate[i]
			}
		}
	}

	// Limit top 10 highest win rate cards
	if len(highestWinRate) > 10 {
		highestWinRate = highestWinRate[:10]
	}

	return EventCardAnalysis{
		MostUsedCards:       mostUsed,
		HighestWinRateCards: highestWinRate,
		TotalUniqueCards:    len(uniqueCards),
	}
}

// analyzeElixirPerformance analyzes performance by deck elixir cost ranges
func analyzeElixirPerformance(decks []EventDeck) EventElixirAnalysis {
	var lowDecks, midDecks, highDecks []EventDeck

	// Categorize decks by average elixir cost
	for _, deck := range decks {
		avgElixir := deck.Deck.AvgElixir
		switch {
		case avgElixir < 3.5:
			lowDecks = append(lowDecks, deck)
		case avgElixir < 4.5:
			midDecks = append(midDecks, deck)
		default:
			highDecks = append(highDecks, deck)
		}
	}

	return EventElixirAnalysis{
		LowElixir:  calculateElixirRangeStats(lowDecks, "Low (0.0-3.4)"),
		MidElixir:  calculateElixirRangeStats(midDecks, "Mid (3.5-4.4)"),
		HighElixir: calculateElixirRangeStats(highDecks, "High (4.5+)"),
	}
}

// calculateElixirRangeStats computes statistics for decks within an elixir range
func calculateElixirRangeStats(decks []EventDeck, rangeLabel string) ElixirRangeStats {
	if len(decks) == 0 {
		return ElixirRangeStats{
			Range:      rangeLabel,
			DeckCount:  0,
			AvgWinRate: 0,
		}
	}

	totalWins := 0
	totalBattles := 0

	for _, deck := range decks {
		perf := deck.Performance
		totalWins += perf.Wins
		totalBattles += perf.TotalBattles()
	}

	var avgWinRate float64
	if totalBattles > 0 {
		avgWinRate = float64(totalWins) / float64(totalBattles)
	}

	return ElixirRangeStats{
		Range:      rangeLabel,
		DeckCount:  len(decks),
		AvgWinRate: avgWinRate,
	}
}

// calculateEventBreakdown computes performance statistics by event type
func calculateEventBreakdown(decks []EventDeck) map[string]EventStats {
	eventStats := make(map[string]EventStats)

	for _, deck := range decks {
		eventType := string(deck.EventType)
		stats, exists := eventStats[eventType]
		if !exists {
			stats = EventStats{
				EventType: eventType,
				Count:     0,
				Wins:      0,
				Losses:    0,
			}
		}

		stats.Count++
		stats.Wins += deck.Performance.Wins
		stats.Losses += deck.Performance.Losses

		eventStats[eventType] = stats
	}

	return eventStats
}

// identifyTopPerformingDecks finds the best performing decks based on win rate and battle count
func identifyTopPerformingDecks(decks []EventDeck, options AnalysisOptions) []TopPerformingDeck {
	var topDecks []TopPerformingDeck

	for _, deck := range decks {
		// Skip decks that don't meet minimum battle requirement
		if deck.Performance.TotalBattles() < options.MinBattlesForTopDecks {
			continue
		}

		// Extract card names from the deck
		cardNames := make([]string, len(deck.Deck.Cards))
		for i, card := range deck.Deck.Cards {
			cardNames[i] = card.Name
		}

		// Create record string (e.g., "12W-2L")
		record := fmt.Sprintf("%dW-%dL", deck.Performance.Wins, deck.Performance.Losses)

		topDeck := TopPerformingDeck{
			EventName: deck.EventName,
			EventType: string(deck.EventType),
			WinRate:   deck.Performance.WinRate,
			Record:    record,
			Deck:      cardNames,
			AvgElixir: deck.Deck.AvgElixir,
		}

		topDecks = append(topDecks, topDeck)
	}

	// Sort by win rate descending
	for i := 0; i < len(topDecks)-1; i++ {
		for j := i + 1; j < len(topDecks); j++ {
			if topDecks[j].WinRate > topDecks[i].WinRate {
				topDecks[i], topDecks[j] = topDecks[j], topDecks[i]
			}
		}
	}

	// Limit to requested number
	if len(topDecks) > options.LimitTopDecks {
		topDecks = topDecks[:options.LimitTopDecks]
	}

	return topDecks
}

// SuggestCardConstraints analyzes top decks and suggests cards that appear frequently
// based on a percentage threshold. Returns suggested cards with their frequency stats.
func SuggestCardConstraints(decks []EventDeck, threshold float64) []CardConstraintSuggestion {
	if len(decks) == 0 {
		return []CardConstraintSuggestion{}
	}

	// Analyze card usage across the provided decks
	cardAnalysis := analyzeCardUsage(decks)

	// Calculate percentage for each card and filter by threshold
	var suggestions []CardConstraintSuggestion
	totalDecks := len(decks)

	for _, cardUsage := range cardAnalysis.MostUsedCards {
		percentage := (float64(cardUsage.Count) / float64(totalDecks)) * 100.0

		// Only include cards that meet or exceed the threshold
		if percentage >= threshold {
			suggestions = append(suggestions, CardConstraintSuggestion{
				CardName:    cardUsage.CardName,
				Appearances: cardUsage.Count,
				Percentage:  percentage,
				TotalDecks:  totalDecks,
			})
		}
	}

	// Sort by percentage descending (highest frequency first)
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Percentage > suggestions[i].Percentage {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions
}
