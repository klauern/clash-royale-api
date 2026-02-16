// Package events provides types and utilities for tracking Clash Royale event decks
// (challenges, tournaments, special events) with performance metrics.
package events

import (
	"fmt"
	"sort"
	"time"
)

// EventAnalysis represents comprehensive analysis of event decks
type EventAnalysis struct {
	PlayerTag       string                `json:"player_tag"`
	AnalysisTime    time.Time             `json:"analysis_time"`
	TotalDecks      int                   `json:"total_decks"`
	Summary         EventSummary          `json:"summary"`
	CardAnalysis    EventCardAnalysis     `json:"card_analysis"`
	ElixirAnalysis  EventElixirAnalysis   `json:"elixir_analysis"`
	EventBreakdown  map[string]EventStats `json:"event_breakdown"`
	TopDecks        []TopPerformingDeck   `json:"top_decks"`
	MatchupAnalysis EventMatchupAnalysis  `json:"matchup_analysis"`
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

// EventMatchupAnalysis provides deck-vs-deck and archetype matchup performance.
type EventMatchupAnalysis struct {
	TotalTrackedBattles int                     `json:"total_tracked_battles"`
	UniqueDeckMatchups  int                     `json:"unique_deck_matchups"`
	TopWinningMatchups  []DeckMatchupStats      `json:"top_winning_matchups"`
	TopLosingMatchups   []DeckMatchupStats      `json:"top_losing_matchups"`
	MostPlayedMatchups  []DeckMatchupStats      `json:"most_played_matchups"`
	ArchetypeMatchups   []ArchetypeMatchupStats `json:"archetype_matchups"`
}

// DeckMatchupStats tracks win/loss performance for a deck pair.
type DeckMatchupStats struct {
	PlayerDeckHash   string   `json:"player_deck_hash"`
	OpponentDeckHash string   `json:"opponent_deck_hash"`
	PlayerDeck       []string `json:"player_deck"`
	OpponentDeck     []string `json:"opponent_deck"`
	Battles          int      `json:"battles"`
	Wins             int      `json:"wins"`
	Losses           int      `json:"losses"`
	Draws            int      `json:"draws"`
	WinRate          float64  `json:"win_rate"`
}

// ArchetypeMatchupStats tracks win/loss performance for archetype pairs.
type ArchetypeMatchupStats struct {
	PlayerArchetype   string  `json:"player_archetype"`
	OpponentArchetype string  `json:"opponent_archetype"`
	Battles           int     `json:"battles"`
	Wins              int     `json:"wins"`
	Losses            int     `json:"losses"`
	Draws             int     `json:"draws"`
	WinRate           float64 `json:"win_rate"`
}

// AnalysisOptions configures event deck analysis behavior
type AnalysisOptions struct {
	MinBattlesForTopDecks int      `json:"min_battles_for_top_decks"` // Minimum battles to qualify for top decks
	LimitTopDecks         int      `json:"limit_top_decks"`           // Maximum number of top decks to include
	EventTypes            []string `json:"event_types"`               // Filter by event types (empty = all)
	MinBattlesForMatchups int      `json:"min_battles_for_matchups"`  // Minimum battles for matchup reporting
	LimitTopMatchups      int      `json:"limit_top_matchups"`        // Max number of matchup rows per category
}

// DefaultAnalysisOptions returns sensible defaults for event analysis
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		MinBattlesForTopDecks: 3,
		LimitTopDecks:         5,
		EventTypes:            []string{},
		MinBattlesForMatchups: 2,
		LimitTopMatchups:      10,
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
			MatchupAnalysis: EventMatchupAnalysis{
				TopWinningMatchups: []DeckMatchupStats{},
				TopLosingMatchups:  []DeckMatchupStats{},
				MostPlayedMatchups: []DeckMatchupStats{},
				ArchetypeMatchups:  []ArchetypeMatchupStats{},
			},
		}
	}

	// Calculate basic statistics
	summary := calculateSummary(filteredDecks)
	cardAnalysis := analyzeCardUsage(filteredDecks)
	elixirAnalysis := analyzeElixirPerformance(filteredDecks)
	eventBreakdown := calculateEventBreakdown(filteredDecks)
	topDecks := identifyTopPerformingDecks(filteredDecks, options)
	matchupAnalysis := analyzeDeckMatchups(filteredDecks, options)

	// Determine player tag from first deck (all decks should have same player)
	playerTag := ""
	if len(filteredDecks) > 0 {
		playerTag = filteredDecks[0].PlayerTag
	}

	return &EventAnalysis{
		PlayerTag:       playerTag,
		AnalysisTime:    time.Now(),
		TotalDecks:      len(filteredDecks),
		Summary:         summary,
		CardAnalysis:    cardAnalysis,
		ElixirAnalysis:  elixirAnalysis,
		EventBreakdown:  eventBreakdown,
		TopDecks:        topDecks,
		MatchupAnalysis: matchupAnalysis,
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

//nolint:gocyclo,funlen // Analysis flow is kept explicit for readability of ranking categories.
func analyzeDeckMatchups(decks []EventDeck, options AnalysisOptions) EventMatchupAnalysis {
	deckAgg := make(map[string]*DeckMatchupStats)
	archetypeAgg := make(map[string]*ArchetypeMatchupStats)
	totalTracked := 0

	for _, eventDeck := range decks {
		for _, battle := range eventDeck.Battles {
			if updateMatchupAggregates(eventDeck, battle, deckAgg, archetypeAgg) {
				totalTracked++
			}
		}
	}

	deckStats := make([]DeckMatchupStats, 0, len(deckAgg))
	for _, stats := range deckAgg {
		if stats.Battles < options.MinBattlesForMatchups {
			continue
		}
		stats.WinRate = float64(stats.Wins) / float64(stats.Battles)
		deckStats = append(deckStats, *stats)
	}

	archetypeStats := make([]ArchetypeMatchupStats, 0, len(archetypeAgg))
	for _, stats := range archetypeAgg {
		if stats.Battles < options.MinBattlesForMatchups {
			continue
		}
		stats.WinRate = float64(stats.Wins) / float64(stats.Battles)
		archetypeStats = append(archetypeStats, *stats)
	}

	topWinning := append([]DeckMatchupStats(nil), deckStats...)
	sort.Slice(topWinning, func(i, j int) bool {
		if topWinning[i].WinRate == topWinning[j].WinRate {
			return topWinning[i].Battles > topWinning[j].Battles
		}
		return topWinning[i].WinRate > topWinning[j].WinRate
	})

	topLosing := append([]DeckMatchupStats(nil), deckStats...)
	sort.Slice(topLosing, func(i, j int) bool {
		if topLosing[i].WinRate == topLosing[j].WinRate {
			return topLosing[i].Battles > topLosing[j].Battles
		}
		return topLosing[i].WinRate < topLosing[j].WinRate
	})

	mostPlayed := append([]DeckMatchupStats(nil), deckStats...)
	sort.Slice(mostPlayed, func(i, j int) bool {
		if mostPlayed[i].Battles == mostPlayed[j].Battles {
			return mostPlayed[i].WinRate > mostPlayed[j].WinRate
		}
		return mostPlayed[i].Battles > mostPlayed[j].Battles
	})

	sort.Slice(archetypeStats, func(i, j int) bool {
		if archetypeStats[i].Battles == archetypeStats[j].Battles {
			return archetypeStats[i].WinRate > archetypeStats[j].WinRate
		}
		return archetypeStats[i].Battles > archetypeStats[j].Battles
	})

	return EventMatchupAnalysis{
		TotalTrackedBattles: totalTracked,
		UniqueDeckMatchups:  len(deckAgg),
		TopWinningMatchups:  limitDeckMatchupStats(topWinning, options.LimitTopMatchups),
		TopLosingMatchups:   limitDeckMatchupStats(topLosing, options.LimitTopMatchups),
		MostPlayedMatchups:  limitDeckMatchupStats(mostPlayed, options.LimitTopMatchups),
		ArchetypeMatchups:   limitArchetypeMatchupStats(archetypeStats, options.LimitTopMatchups),
	}
}

//nolint:gocyclo // Matchup extraction handles fallback/deck-hash/archetype branching.
func updateMatchupAggregates(
	eventDeck EventDeck,
	battle BattleRecord,
	deckAgg map[string]*DeckMatchupStats,
	archetypeAgg map[string]*ArchetypeMatchupStats,
) bool {
	playerDeck := battle.PlayerDeck
	if len(playerDeck) == 0 {
		playerDeck = deckNamesFromEventDeck(eventDeck)
	}
	if len(playerDeck) == 0 || len(battle.OpponentDeck) == 0 {
		return false
	}

	playerHash := battle.PlayerDeckHash
	if playerHash == "" {
		playerHash = deckHash(playerDeck)
	}
	opponentHash := battle.OpponentDeckHash
	if opponentHash == "" {
		opponentHash = deckHash(battle.OpponentDeck)
	}
	if playerHash == "" || opponentHash == "" {
		return false
	}

	deckStats := getOrCreateDeckMatchup(deckAgg, playerHash, opponentHash, playerDeck, battle.OpponentDeck)
	deckStats.Battles++
	incrementResultCounters(battle.Result, &deckStats.Wins, &deckStats.Losses, &deckStats.Draws)

	playerArchetype := battle.PlayerDeckArchetype
	if playerArchetype == "" {
		playerArchetype = inferDeckArchetype(playerDeck)
	}
	opponentArchetype := battle.OpponentDeckArchetype
	if opponentArchetype == "" {
		opponentArchetype = inferDeckArchetype(battle.OpponentDeck)
	}
	if playerArchetype != "" && opponentArchetype != "" {
		archetypeStats := getOrCreateArchetypeMatchup(archetypeAgg, playerArchetype, opponentArchetype)
		archetypeStats.Battles++
		incrementResultCounters(battle.Result, &archetypeStats.Wins, &archetypeStats.Losses, &archetypeStats.Draws)
	}

	return true
}

func getOrCreateDeckMatchup(
	deckAgg map[string]*DeckMatchupStats,
	playerHash, opponentHash string,
	playerDeck, opponentDeck []string,
) *DeckMatchupStats {
	key := playerHash + "::" + opponentHash
	stats, exists := deckAgg[key]
	if exists {
		return stats
	}

	stats = &DeckMatchupStats{
		PlayerDeckHash:   playerHash,
		OpponentDeckHash: opponentHash,
		PlayerDeck:       append([]string(nil), playerDeck...),
		OpponentDeck:     append([]string(nil), opponentDeck...),
	}
	deckAgg[key] = stats
	return stats
}

func getOrCreateArchetypeMatchup(
	archetypeAgg map[string]*ArchetypeMatchupStats,
	playerArchetype, opponentArchetype string,
) *ArchetypeMatchupStats {
	key := playerArchetype + "::" + opponentArchetype
	stats, exists := archetypeAgg[key]
	if exists {
		return stats
	}

	stats = &ArchetypeMatchupStats{
		PlayerArchetype:   playerArchetype,
		OpponentArchetype: opponentArchetype,
	}
	archetypeAgg[key] = stats
	return stats
}

func incrementResultCounters(result string, wins, losses, draws *int) {
	switch result {
	case BattleResultWin:
		(*wins)++
	case BattleResultLoss:
		(*losses)++
	default:
		(*draws)++
	}
}

func deckNamesFromEventDeck(eventDeck EventDeck) []string {
	names := make([]string, 0, len(eventDeck.Deck.Cards))
	for _, card := range eventDeck.Deck.Cards {
		if card.Name == "" {
			continue
		}
		names = append(names, card.Name)
	}
	return names
}

func limitDeckMatchupStats(stats []DeckMatchupStats, limit int) []DeckMatchupStats {
	if limit <= 0 || len(stats) <= limit {
		return stats
	}
	return stats[:limit]
}

func limitArchetypeMatchupStats(stats []ArchetypeMatchupStats, limit int) []ArchetypeMatchupStats {
	if limit <= 0 || len(stats) <= limit {
		return stats
	}
	return stats[:limit]
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
