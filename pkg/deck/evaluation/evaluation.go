//nolint:funlen,goconst,gocognit,gocyclo // Existing orchestration is unchanged by the structural split.
package evaluation

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	overallWeightAttack      = 0.23
	overallWeightDefense     = 0.22
	overallWeightSynergy     = 0.21
	overallWeightVersatility = 0.14
	overallWeightF2P         = 0.10
	overallWeightPlayability = 0.10
)

// applyCriticalFlawPenalties applies penalties for compositional flaws that
// make a deck fundamentally unviable beyond what category scores capture.
//
//nolint:gocognit,gocyclo // Domain penalty matrix is explicit to keep balancing transparent.
func applyCriticalFlawPenalties(baseScore float64, deckCards []deck.CardCandidate) float64 {
	score := baseScore

	// Check for critical attack flaws
	winConditionCount := 0
	spellCount := 0
	bigSpellCount := 0
	for _, card := range deckCards {
		if card.Role != nil {
			if *card.Role == deck.RoleWinCondition {
				winConditionCount++
			}
			if *card.Role == deck.RoleSpellBig {
				bigSpellCount++
			}
			if *card.Role == deck.RoleSpellBig || *card.Role == deck.RoleSpellSmall {
				spellCount++
			}
		}
	}

	// Penalty for no win condition: -2.0 points (critical flaw)
	if winConditionCount == 0 {
		score -= 2.0
	}

	// Penalty for no spells: -1.5 points (severe limitation)
	if spellCount == 0 {
		score -= 1.5
	}

	// Check for critical defense flaws
	antiAirCount := 0
	for _, card := range deckCards {
		if card.Stats != nil && (card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground") {
			antiAirCount++
		}
	}

	// Penalty for no anti-air: -2.0 points (critical vulnerability)
	if antiAirCount == 0 {
		score -= 2.0
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

// Evaluate performs comprehensive deck evaluation with all scoring and analysis
// If playerContext is provided, evaluation will include player-specific context such as:
// - Card levels from player's collection
// - Arena-specific card availability
// - Evolution unlock status
func Evaluate(deckCards []deck.CardCandidate, synergyDB *deck.SynergyDatabase, playerContext *PlayerContext) EvaluationResult {
	// Extract deck card names
	deckNames := make([]string, len(deckCards))
	for i, card := range deckCards {
		deckNames[i] = card.Name
	}

	// Calculate average elixir
	avgElixir := calculateAvgElixir(deckCards)

	// Phase 1: Category Scoring
	attackScore := ScoreAttack(deckCards)
	defenseScore := ScoreDefense(deckCards)
	synergyScore := ScoreSynergy(deckCards, synergyDB)
	versatilityScore := ScoreVersatility(deckCards)
	f2pScore := ScoreF2P(deckCards)
	playabilityScore := ScorePlayability(deckCards, playerContext)

	// Phase 2: Archetype Detection
	archetypeResult := DetectArchetype(deckCards)

	// Phase 3: Build Analysis Sections
	defenseAnalysis := BuildDefenseAnalysis(deckCards)
	attackAnalysis := BuildAttackAnalysis(deckCards)
	baitAnalysis := BuildBaitAnalysis(deckCards)
	cycleAnalysis := BuildCycleAnalysis(deckCards)
	ladderAnalysis := BuildLadderAnalysis(deckCards, playerContext)
	evolutionAnalysis := BuildEvolutionAnalysis(deckCards, playerContext)

	// Phase 4: Calculate Overall Score (weighted average)
	// Weights: Attack 23%, Defense 22%, Synergy 21%, Versatility 14%, F2P 10%, Playability 10%
	// Balanced emphasis on attack/defense/synergy fundamentals
	// Critical flaws are separately penalized via applyCriticalFlawPenalties
	baseOverallScore := (attackScore.Score * overallWeightAttack) +
		(defenseScore.Score * overallWeightDefense) +
		(synergyScore.Score * overallWeightSynergy) +
		(versatilityScore.Score * overallWeightVersatility) +
		(f2pScore.Score * overallWeightF2P) +
		(playabilityScore.Score * overallWeightPlayability)

	// When player context is available, replace Playability with ladder viability at the same weight.
	contextualScore := baseOverallScore
	if playerContext != nil {
		contextualScore = baseOverallScore - (playabilityScore.Score * overallWeightPlayability) + (ladderAnalysis.Score * overallWeightPlayability)
	}

	levelRatio := 1.0
	normalizationFactor := 1.0
	normalizedScore := contextualScore
	overallScore := contextualScore

	if playerContext != nil {
		levelRatio = calculateDeckLevelRatio(deckCards, playerContext)
		normalizationFactor = calculateLevelNormalizationFactor(levelRatio)
		normalizedScore = clampScoreToTen(contextualScore * normalizationFactor)

		// Prototype blend: preserve base signal, then fold in ladder and level-normalized context.
		overallScore = (contextualScore * 0.75) + (ladderAnalysis.Score * 0.15) + (normalizedScore * 0.10)
	}

	overallScore = clampScoreToTen(overallScore)

	// Apply penalties for critical compositional flaws
	// These are severe enough to warrant direct overall score penalties
	overallScore = applyCriticalFlawPenalties(overallScore, deckCards)

	overallRating := ScoreToRating(overallScore)

	// Build synergy matrix (if database provided)
	var synergyMatrix SynergyMatrix
	if synergyDB != nil {
		synergyAnalysis := synergyDB.AnalyzeDeckSynergy(deckNames)
		if synergyAnalysis != nil {
			maxPairs := (len(deckNames) * (len(deckNames) - 1)) / 2
			pairCount := 0
			for _, count := range synergyAnalysis.CategoryScores {
				pairCount += count
			}
			coverage := 0.0
			if maxPairs > 0 {
				coverage = float64(pairCount) / float64(maxPairs) * 100.0
			}

			synergyMatrix = SynergyMatrix{
				Pairs:            synergyAnalysis.TopSynergies,
				TotalScore:       synergyScore.Score,
				AverageSynergy:   synergyAnalysis.AverageScore,
				PairCount:        pairCount,
				MaxPossiblePairs: maxPairs,
				SynergyCoverage:  coverage,
			}
		}
	}

	// Analyze missing cards if player context provided
	var missingCardsAnalysis *MissingCardsAnalysis
	if playerContext != nil {
		missingCardsAnalysis = IdentifyMissingCardsWithContext(deckCards, playerContext)

		// Apply arena-based score penalties for locked cards
		if missingCardsAnalysis != nil && missingCardsAnalysis.MissingCount > 0 {
			lockedCount := 0
			for _, card := range missingCardsAnalysis.MissingCards {
				if card.IsLocked {
					lockedCount++
				}
			}

			// Penalty: -2 points per locked card, -1 point per unlocked but missing card
			penalty := float64(lockedCount)*2.0 + float64(missingCardsAnalysis.MissingCount-lockedCount)*1.0
			overallScore -= penalty

			// Ensure score doesn't go below 0
			overallScore = clampScoreToTen(overallScore)

			// Recalculate overall rating with penalty applied
			overallRating = ScoreToRating(overallScore)
		}
	}

	var overallBreakdown *OverallScoreBreakdown
	if playerContext != nil {
		overallBreakdown = &OverallScoreBreakdown{
			BaseScore:           baseOverallScore,
			ContextualScore:     contextualScore,
			LadderScore:         ladderAnalysis.Score,
			NormalizedScore:     normalizedScore,
			FinalScore:          overallScore,
			DeckLevelRatio:      levelRatio,
			NormalizationFactor: normalizationFactor,
		}
	}

	// Assemble complete result
	return EvaluationResult{
		Deck:      deckNames,
		AvgElixir: avgElixir,

		Attack:      attackScore,
		Defense:     defenseScore,
		Synergy:     synergyScore,
		Versatility: versatilityScore,
		F2PFriendly: f2pScore,
		Playability: playabilityScore,

		OverallScore:  overallScore,
		OverallRating: overallRating,

		DetectedArchetype:   archetypeResult.Primary,
		ArchetypeConfidence: archetypeResult.PrimaryConfidence,

		DefenseAnalysis:   defenseAnalysis,
		AttackAnalysis:    attackAnalysis,
		BaitAnalysis:      baitAnalysis,
		CycleAnalysis:     cycleAnalysis,
		LadderAnalysis:    ladderAnalysis,
		EvolutionAnalysis: evolutionAnalysis,

		SynergyMatrix:        synergyMatrix,
		MissingCardsAnalysis: missingCardsAnalysis,
		OverallBreakdown:     overallBreakdown,
	}
}
