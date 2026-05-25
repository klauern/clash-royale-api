package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func renderReviewHuman(report *reviewReport) {
	p := report.Player
	printf("═══════════════════════════════════════════════════════\n")
	printf("  PLAYER REVIEW: %s (%s)\n", p.Name, p.Tag)
	printf("═══════════════════════════════════════════════════════\n\n")

	renderReviewProfile(report)
	renderReviewPlaystyle(report)
	renderReviewArchetype(report)
	renderReviewUpgrades(report)
	renderReviewBudget(report)
}

func renderReviewProfile(report *reviewReport) {
	p := report.Player
	printf("── Profile ─────────────────────────────────────────────\n")
	printf("  Level:    %d   Trophies: %d   Best: %d\n", p.ExpLevel, p.Trophies, p.BestTrophies)
	printf("  Arena:    %s\n", p.Arena.Name)
	if p.League.ID > 0 {
		printf("  League:   %s\n", p.League.Name)
	}
	winRate := 0.0
	if p.BattleCount > 0 {
		winRate = float64(p.Wins) / float64(p.BattleCount) * 100
	}
	printf("  Record:   %d W / %d L  (%.1f%% win rate)\n\n", p.Wins, p.Losses, winRate)
}

func renderReviewPlaystyle(report *reviewReport) {
	printf("── Playstyle ───────────────────────────────────────────\n")
	if ps := report.Playstyle; ps != nil {
		printf("  Style:      %s\n", ps.DeckStyle)
		printf("  Aggression: %s\n", ps.AggressionLevel)
		if len(ps.PlaystyleTraits) > 0 {
			printf("  Traits:     %s\n", strings.Join(ps.PlaystyleTraits, ", "))
		}
		printf("  Avg Elixir: %.1f\n", ps.CurrentDeckAvgElixir)
	}
	printf("\n")
}

func renderReviewArchetype(report *reviewReport) {
	printf("── Top Archetype ───────────────────────────────────────\n")
	if arch := report.TopArchetype; arch != nil {
		printf("  Name:       %s\n", arch.Name)
		printf("  Win Con:    %s\n", arch.WinCondition)
		printf("  Viability:  %.1f  (%s)\n", arch.ViabilityScore, arch.ViabilityTier)
		if arch.GoldToCompetitive > 0 {
			printf("  Gold to competitive: %s\n", formatGoldCompact(arch.GoldToCompetitive))
		}
	} else {
		printf("  No archetypes detected.\n")
	}
	printf("\n")
}

func renderReviewUpgrades(report *reviewReport) {
	printf("── Top Cross-Archetype Upgrade Priorities ──────────────\n")
	if len(report.CrossArchUpgrades) == 0 {
		printf("  No cross-archetype upgrade data.\n")
	}
	for i, up := range report.CrossArchUpgrades {
		printf("  %d. %-22s  Lv %d  cost %s gold\n",
			i+1, up.CardName, up.CurrentLevel, formatGoldCompact(up.GoldCost))
		printf("     Unlocks %d archetype(s)  (+%.1f viability)\n",
			up.ArchetypesUnlocked, up.TotalViabilityGain)
	}
	printf("\n")
}

func renderReviewBudget(report *reviewReport) {
	printf("── Budget: Next 20k Gold ───────────────────────────────\n")
	if len(report.BudgetDecks) == 0 {
		printf("  No decks within 20,000 gold.\n")
	}
	for i, bd := range report.BudgetDecks {
		if i >= 3 {
			break
		}
		deckName := "—"
		if bd.Deck != nil {
			deckName = strings.Join(bd.Deck.Deck, ", ")
			if len(deckName) > 50 {
				deckName = deckName[:50] + "…"
			}
		}
		printf("  %d. Score %.1f → %.1f  cost %s gold\n",
			i+1, bd.CurrentScore, bd.ProjectedScore, formatGoldCompact(bd.TotalGoldNeeded))
		printf("     %s\n", deckName)
	}
	printf("\n")
}

func renderReviewJSON(report *reviewReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal review report: %w", err)
	}
	printf("%s\n", data)
	return nil
}

func renderReviewMarkdown(_ *reviewReport) error {
	return fmt.Errorf("markdown output not yet implemented (coming in C4)")
}
