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
	renderReviewDelta(report)
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

func renderReviewDelta(report *reviewReport) {
	printf("── Current Deck vs Best Recommended ────────────────────\n")
	d := report.DeckDelta
	if d == nil {
		printf("  Deck delta unavailable.\n\n")
		return
	}
	printf("  Overall:    current %.2f  →  recommended %.2f  (Δ %+.2f)\n",
		d.CurrentScore, d.RecommendedScore, d.ScoreDelta)
	printf("  Archetype:  current %-14s  →  recommended %s\n",
		d.CurrentArchetype, d.RecommendedArchetype)
	printf("  Level fit:  current %.0f%%  →  recommended %.0f%%\n",
		d.CurrentLevelFit*100, d.RecommendedLevelFit*100)
	if len(d.SharedCards) > 0 {
		printf("  Kept (%d):   %s\n", len(d.SharedCards), strings.Join(d.SharedCards, ", "))
	}
	if len(d.ReplacedCards) > 0 {
		printf("  New  (%d):   %s\n", len(d.ReplacedCards), strings.Join(d.ReplacedCards, ", "))
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

func renderReviewMarkdown(report *reviewReport) error {
	var sb strings.Builder
	p := report.Player

	fmt.Fprintf(&sb, "# Player Review: %s (%s)\n\n", p.Name, p.Tag)

	reviewMarkdownProfile(&sb, report)
	reviewMarkdownPlaystyle(&sb, report)
	reviewMarkdownArchetype(&sb, report)
	reviewMarkdownUpgrades(&sb, report)
	reviewMarkdownDelta(&sb, report)
	reviewMarkdownBudget(&sb, report)

	printf("%s", sb.String())
	return nil
}

func reviewMarkdownProfile(sb *strings.Builder, report *reviewReport) {
	p := report.Player
	winRate := 0.0
	if p.BattleCount > 0 {
		winRate = float64(p.Wins) / float64(p.BattleCount) * 100
	}
	fmt.Fprintf(sb, "## Profile\n\n")
	fmt.Fprintf(sb, "| Field | Value |\n|-------|-------|\n")
	fmt.Fprintf(sb, "| Level | %d |\n", p.ExpLevel)
	fmt.Fprintf(sb, "| Trophies | %d (best: %d) |\n", p.Trophies, p.BestTrophies)
	fmt.Fprintf(sb, "| Arena | %s |\n", p.Arena.Name)
	if p.League.ID > 0 {
		fmt.Fprintf(sb, "| League | %s |\n", p.League.Name)
	}
	fmt.Fprintf(sb, "| Record | %d W / %d L (%.1f%%) |\n\n", p.Wins, p.Losses, winRate)
}

func reviewMarkdownPlaystyle(sb *strings.Builder, report *reviewReport) {
	fmt.Fprintf(sb, "## Playstyle\n\n")
	ps := report.Playstyle
	if ps == nil {
		fmt.Fprintf(sb, "_No playstyle data._\n\n")
		return
	}
	fmt.Fprintf(sb, "| Field | Value |\n|-------|-------|\n")
	fmt.Fprintf(sb, "| Style | %s |\n", ps.DeckStyle)
	fmt.Fprintf(sb, "| Aggression | %s |\n", ps.AggressionLevel)
	fmt.Fprintf(sb, "| Avg Elixir | %.1f |\n", ps.CurrentDeckAvgElixir)
	if len(ps.PlaystyleTraits) > 0 {
		fmt.Fprintf(sb, "| Traits | %s |\n", strings.Join(ps.PlaystyleTraits, ", "))
	}
	fmt.Fprintf(sb, "\n")
}

func reviewMarkdownArchetype(sb *strings.Builder, report *reviewReport) {
	fmt.Fprintf(sb, "## Top Archetype\n\n")
	arch := report.TopArchetype
	if arch == nil {
		fmt.Fprintf(sb, "_No archetypes detected._\n\n")
		return
	}
	fmt.Fprintf(sb, "**%s** (%s) — %.1f viability (%s)\n\n", arch.Name, arch.WinCondition, arch.ViabilityScore, arch.ViabilityTier)
	if arch.GoldToCompetitive > 0 {
		fmt.Fprintf(sb, "Gold to competitive tier: **%s**\n\n", formatGoldCompact(arch.GoldToCompetitive))
	}
}

func reviewMarkdownUpgrades(sb *strings.Builder, report *reviewReport) {
	fmt.Fprintf(sb, "## Cross-Archetype Upgrade Priorities\n\n")
	if len(report.CrossArchUpgrades) == 0 {
		fmt.Fprintf(sb, "_No cross-archetype upgrade data._\n\n")
		return
	}
	fmt.Fprintf(sb, "| # | Card | Level | Gold Cost | Archetypes Unlocked | Viability Gain |\n")
	fmt.Fprintf(sb, "|---|------|-------|-----------|--------------------|-----------------|\n")
	for i, up := range report.CrossArchUpgrades {
		fmt.Fprintf(sb, "| %d | %s | %d | %s | %d | +%.1f |\n",
			i+1, up.CardName, up.CurrentLevel, formatGoldCompact(up.GoldCost),
			up.ArchetypesUnlocked, up.TotalViabilityGain)
	}
	fmt.Fprintf(sb, "\n")
}

func reviewMarkdownDelta(sb *strings.Builder, report *reviewReport) {
	fmt.Fprintf(sb, "## Current Deck vs Best Recommended\n\n")
	d := report.DeckDelta
	if d == nil {
		fmt.Fprintf(sb, "_Deck delta unavailable._\n\n")
		return
	}
	fmt.Fprintf(sb, "| | Current | Recommended | Delta |\n")
	fmt.Fprintf(sb, "|-|---------|-------------|-------|\n")
	fmt.Fprintf(sb, "| Overall Score | %.2f | %.2f | **%+.2f** |\n", d.CurrentScore, d.RecommendedScore, d.ScoreDelta)
	fmt.Fprintf(sb, "| Archetype | %s | %s | — |\n", d.CurrentArchetype, d.RecommendedArchetype)
	fmt.Fprintf(sb, "| Level Fit | %.0f%% | %.0f%% | %+.0f%% |\n\n",
		d.CurrentLevelFit*100, d.RecommendedLevelFit*100, (d.RecommendedLevelFit-d.CurrentLevelFit)*100)
	if len(d.SharedCards) > 0 {
		fmt.Fprintf(sb, "**Kept cards (%d):** %s\n\n", len(d.SharedCards), strings.Join(d.SharedCards, ", "))
	}
	if len(d.ReplacedCards) > 0 {
		fmt.Fprintf(sb, "**New cards (%d):** %s\n\n", len(d.ReplacedCards), strings.Join(d.ReplacedCards, ", "))
	}
}

func reviewMarkdownBudget(sb *strings.Builder, report *reviewReport) {
	fmt.Fprintf(sb, "## Budget: Next 20k Gold\n\n")
	if len(report.BudgetDecks) == 0 {
		fmt.Fprintf(sb, "_No decks within 20,000 gold._\n\n")
		return
	}
	fmt.Fprintf(sb, "| # | Current Score | Projected Score | Gold Cost | Cards |\n")
	fmt.Fprintf(sb, "|---|--------------|----------------|-----------|-------|\n")
	for i, bd := range report.BudgetDecks {
		if i >= 3 {
			break
		}
		deckCards := "—"
		if bd.Deck != nil {
			deckCards = strings.Join(bd.Deck.Deck, ", ")
			if len(deckCards) > 60 {
				deckCards = deckCards[:60] + "…"
			}
		}
		fmt.Fprintf(sb, "| %d | %.1f | %.1f | %s | %s |\n",
			i+1, bd.CurrentScore, bd.ProjectedScore, formatGoldCompact(bd.TotalGoldNeeded), deckCards)
	}
	fmt.Fprintf(sb, "\n")
}
