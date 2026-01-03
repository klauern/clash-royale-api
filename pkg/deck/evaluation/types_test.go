package evaluation

import (
	"testing"
)

func TestArchetypeString(t *testing.T) {
	tests := []struct {
		archetype Archetype
		expected  string
	}{
		{ArchetypeBeatdown, "beatdown"},
		{ArchetypeControl, "control"},
		{ArchetypeCycle, "cycle"},
		{ArchetypeBridge, "bridge"},
		{ArchetypeSiege, "siege"},
		{ArchetypeBait, "bait"},
		{ArchetypeGraveyard, "graveyard"},
		{ArchetypeMiner, "miner"},
		{ArchetypeHybrid, "hybrid"},
		{ArchetypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.archetype.String()
			if result != tt.expected {
				t.Errorf("Archetype.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCategoryScoreStructure(t *testing.T) {
	score := CategoryScore{
		Score:      8.5,
		Rating:     RatingAmazing,
		Assessment: "Test assessment",
		Stars:      3,
	}

	if score.Score != 8.5 {
		t.Errorf("CategoryScore.Score = %v, want 8.5", score.Score)
	}
	if score.Rating != RatingAmazing {
		t.Errorf("CategoryScore.Rating = %v, want %v", score.Rating, RatingAmazing)
	}
	if score.Assessment != "Test assessment" {
		t.Errorf("CategoryScore.Assessment = %v, want 'Test assessment'", score.Assessment)
	}
	if score.Stars != 3 {
		t.Errorf("CategoryScore.Stars = %v, want 3", score.Stars)
	}
}

func TestEvaluationResultStructure(t *testing.T) {
	result := EvaluationResult{
		Deck:                []string{"Giant", "Musketeer", "Zap", "Fireball", "Mini P.E.K.K.A", "Valkyrie", "Mega Minion", "Ice Spirit"},
		AvgElixir:           3.5,
		DetectedArchetype:   ArchetypeBeatdown,
		ArchetypeConfidence: 0.85,
		OverallScore:        7.5,
		OverallRating:       RatingGreat,
	}

	if len(result.Deck) != 8 {
		t.Errorf("EvaluationResult.Deck length = %v, want 8", len(result.Deck))
	}
	if result.AvgElixir != 3.5 {
		t.Errorf("EvaluationResult.AvgElixir = %v, want 3.5", result.AvgElixir)
	}
	if result.DetectedArchetype != ArchetypeBeatdown {
		t.Errorf("EvaluationResult.DetectedArchetype = %v, want %v", result.DetectedArchetype, ArchetypeBeatdown)
	}
	if result.ArchetypeConfidence != 0.85 {
		t.Errorf("EvaluationResult.ArchetypeConfidence = %v, want 0.85", result.ArchetypeConfidence)
	}
}

func TestAnalysisSectionStructure(t *testing.T) {
	section := AnalysisSection{
		Title:   "Defense Analysis",
		Summary: "Strong defensive capabilities",
		Details: []string{"Building for defense", "Anti-air coverage"},
		Score:   8.0,
		Rating:  RatingAmazing,
	}

	if section.Title != "Defense Analysis" {
		t.Errorf("AnalysisSection.Title = %v, want 'Defense Analysis'", section.Title)
	}
	if len(section.Details) != 2 {
		t.Errorf("AnalysisSection.Details length = %v, want 2", len(section.Details))
	}
}

func TestSynergyMatrixStructure(t *testing.T) {
	matrix := SynergyMatrix{
		TotalScore:       8.5,
		AverageSynergy:   0.75,
		PairCount:        12,
		MaxPossiblePairs: 28,
		SynergyCoverage:  85.0,
	}

	if matrix.PairCount != 12 {
		t.Errorf("SynergyMatrix.PairCount = %v, want 12", matrix.PairCount)
	}
	if matrix.MaxPossiblePairs != 28 {
		t.Errorf("SynergyMatrix.MaxPossiblePairs = %v, want 28", matrix.MaxPossiblePairs)
	}
}
