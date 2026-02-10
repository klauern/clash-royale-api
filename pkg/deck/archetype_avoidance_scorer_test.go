package deck

import (
	"testing"
)

func TestArchetypeAvoidanceScorer_ScoreCard(t *testing.T) {
	tests := []struct {
		name            string
		avoidArchetypes []string
		cardName        string
		expectedPenalty float64
		description     string
	}{
		{
			name:            "no archetypes to avoid",
			avoidArchetypes: []string{},
			cardName:        "Golem",
			expectedPenalty: 0.0,
			description:     "Should return 0.0 when no archetypes are avoided",
		},
		{
			name:            "avoid beatdown - golem gets penalty",
			avoidArchetypes: []string{"beatdown"},
			cardName:        "Golem",
			expectedPenalty: -0.3,
			description:     "Golem is preferred in beatdown, should get -0.3 penalty",
		},
		{
			name:            "avoid beatdown - hog rider no penalty",
			avoidArchetypes: []string{"beatdown"},
			cardName:        "Hog Rider",
			expectedPenalty: 0.0,
			description:     "Hog Rider not in beatdown preferred cards, no penalty",
		},
		{
			name:            "avoid multiple archetypes - golem penalty",
			avoidArchetypes: []string{"beatdown", "cycle"},
			cardName:        "Golem",
			expectedPenalty: -0.3,
			description:     "Golem gets penalty for being in beatdown (one archetype match)",
		},
		{
			name:            "avoid cycle - ice spirit gets penalty",
			avoidArchetypes: []string{"cycle"},
			cardName:        "Ice Spirit",
			expectedPenalty: -0.3,
			description:     "Ice Spirit is preferred in cycle, should get -0.3 penalty",
		},
		{
			name:            "avoid siege - xbow gets penalty",
			avoidArchetypes: []string{"siege"},
			cardName:        "X-Bow",
			expectedPenalty: -0.3,
			description:     "X-Bow is preferred in siege, should get -0.3 penalty",
		},
		{
			name:            "case insensitive archetype",
			avoidArchetypes: []string{"BeAtDoWn"},
			cardName:        "Golem",
			expectedPenalty: -0.3,
			description:     "Should handle case-insensitive archetype names",
		},
		{
			name:            "bridge_spam variants",
			avoidArchetypes: []string{"bridge spam"},
			cardName:        "Battle Ram",
			expectedPenalty: -0.3,
			description:     "Should handle bridge spam with space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewArchetypeAvoidanceScorer(tt.avoidArchetypes)
			penalty := scorer.ScoreCard(tt.cardName)

			if penalty != tt.expectedPenalty {
				t.Errorf("%s: expected penalty %v, got %v", tt.description, tt.expectedPenalty, penalty)
			}
		})
	}
}

func TestArchetypeAvoidanceScorer_ScoreDeck(t *testing.T) {
	tests := []struct {
		name            string
		avoidArchetypes []string
		deckCards       []string
		minPenalty      float64
		description     string
	}{
		{
			name:            "empty deck",
			avoidArchetypes: []string{"beatdown"},
			deckCards:       []string{},
			minPenalty:      0.0,
			description:     "Empty deck should return 0.0",
		},
		{
			name:            "beatdown deck avoid beatdown",
			avoidArchetypes: []string{"beatdown"},
			deckCards:       []string{"Golem", "Giant", "Baby Dragon", "Lightning", "Tornado", "Arrows", "Mega Minion", "Lumberjack"},
			minPenalty:      -0.3,
			description:     "Beatdown deck with many beatdown cards should have significant penalty",
		},
		{
			name:            "cycle deck avoid beatdown",
			avoidArchetypes: []string{"beatdown"},
			deckCards:       []string{"Hog Rider", "Ice Spirit", "Cannon", "Fireball", "The Log", "Skeletons", "Musketeer", "Ice Golem"},
			minPenalty:      0.0,
			description:     "Cycle deck should have no penalty when avoiding beatdown (no beatdown cards)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewArchetypeAvoidanceScorer(tt.avoidArchetypes)
			avgPenalty := scorer.ScoreDeck(tt.deckCards)

			if avgPenalty > tt.minPenalty {
				t.Errorf("%s: expected penalty <= %v, got %v", tt.description, tt.minPenalty, avgPenalty)
			}
		})
	}
}

func TestArchetypeAvoidanceScorer_IsEnabled(t *testing.T) {
	tests := []struct {
		name            string
		avoidArchetypes []string
		expected        bool
	}{
		{
			name:            "no archetypes",
			avoidArchetypes: []string{},
			expected:        false,
		},
		{
			name:            "one archetype",
			avoidArchetypes: []string{"beatdown"},
			expected:        true,
		},
		{
			name:            "multiple archetypes",
			avoidArchetypes: []string{"beatdown", "cycle"},
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewArchetypeAvoidanceScorer(tt.avoidArchetypes)
			if scorer.IsEnabled() != tt.expected {
				t.Errorf("expected IsEnabled() = %v, got %v", tt.expected, scorer.IsEnabled())
			}
		})
	}
}
