package deck

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// BenchmarkDeckSpaceCalculator benchmarks the deck space calculator
func BenchmarkDeckSpaceCalculator(b *testing.B) {
	player := &clashroyale.Player{
		Tag:   "#BENCH",
		Name:  "Benchmark Player",
		Cards: createLargeCardCollection(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc, _ := NewDeckSpaceCalculator(player)
		_ = calc.CalculateStats()
	}
}

// BenchmarkCalculateConstrainedCombinations benchmarks the constrained combinations calculation
func BenchmarkCalculateConstrainedCombinations(b *testing.B) {
	player := &clashroyale.Player{
		Tag:   "#BENCH",
		Name:  "Benchmark Player",
		Cards: createLargeCardCollection(),
	}

	calc, _ := NewDeckSpaceCalculator(player)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.calculateConstrainedCombinations()
	}
}

// BenchmarkCombinations benchmarks the combinations calculation
func BenchmarkCombinations(b *testing.B) {
	tests := []struct {
		name string
		n    int
		k    int
	}{
		{"C(8,8)", 8, 8},
		{"C(20,8)", 20, 8},
		{"C(50,8)", 50, 8},
		{"C(100,8)", 100, 8},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = combinations(tt.n, tt.k)
			}
		})
	}
}

// BenchmarkNewDeckGenerator benchmarks generator creation
func BenchmarkNewDeckGenerator(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategySmartSample,
		Candidates: candidates,
		SampleSize: 1000,
		Constraints: &GeneratorConstraints{
			MinAvgElixir:        2.0,
			MaxAvgElixir:        5.0,
			RequireWinCondition: true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewDeckGenerator(config)
	}
}

// BenchmarkDeckGenerator_GenerateOne benchmarks single deck generation
func BenchmarkDeckGenerator_GenerateOne(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateOne(ctx)
	}
}

// BenchmarkDeckGenerator_Generate benchmarks batch deck generation
func BenchmarkDeckGenerator_Generate(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(ctx, 100)
	}
}

// BenchmarkDeckGenerator_GenerateByStrategy benchmarks different strategies
func BenchmarkDeckGenerator_GenerateByStrategy(b *testing.B) {
	candidates := createLargeCandidateSet(50)

	strategies := []struct {
		name     string
		strategy GeneratorStrategy
	}{
		{"RandomSample", StrategyRandomSample},
		{"SmartSample", StrategySmartSample},
	}

	for _, tc := range strategies {
		b.Run(tc.name, func(b *testing.B) {
			config := GeneratorConfig{
				Strategy:   tc.strategy,
				Candidates: candidates,
				SampleSize: 1000,
				Seed:       12345,
			}

			gen, _ := NewDeckGenerator(config)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = gen.Generate(ctx, 10)
			}
		})
	}
}

// BenchmarkDeckIterator_Checkpoint benchmarks checkpoint creation
func BenchmarkDeckIterator_Checkpoint(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	iterator, _ := gen.Iterator()
	ctx := context.Background()

	// Generate some decks first
	for i := 0; i < 10; i++ {
		iterator.Next(ctx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = iterator.Checkpoint()
	}
}

// BenchmarkDeckIterator_Resume benchmarks checkpoint resume
func BenchmarkDeckIterator_Resume(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	iterator1, _ := gen.Iterator()
	ctx := context.Background()

	// Generate some decks and create checkpoint
	for i := 0; i < 10; i++ {
		iterator1.Next(ctx)
	}
	checkpoint := iterator1.Checkpoint()

	// Benchmark resume
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iterator2, _ := gen.Iterator()
		_ = iterator2.Resume(checkpoint)
	}
}

// BenchmarkDeckGenerator_Allocation benchmarks memory allocation
func BenchmarkDeckGenerator_Allocation(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, _ := gen.GenerateOne(ctx)
		_ = deck
	}
}

// BenchmarkValidateDeck benchmarks deck validation
func BenchmarkValidateDeck(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategyRandomSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	ctx := context.Background()
	deck, _ := gen.GenerateOne(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.validateDeck(deck)
	}
}

// Helper functions for benchmarks

func createLargeCardCollection() []clashroyale.Card {
	cards := make([]clashroyale.Card, 0, 50)

	// Add cards with various roles
	winConditions := []struct {
		name   string
		elixir int
	}{
		{"Hog Rider", 4},
		{"Giant", 5},
		{"Golem", 8},
		{"Miner", 3},
		{"Balloon", 5},
		{"Royal Giant", 6},
		{"Goblin Barrel", 3},
		{"Graveyard", 5},
		{"P.E.K.K.A", 7},
		{"Mega Knight", 7},
	}

	buildings := []struct {
		name   string
		elixir int
	}{
		{"Cannon", 3},
		{"Tesla", 4},
		{"Inferno Tower", 5},
		{"Bomb Tower", 4},
		{"X-Bow", 6},
		{"Mortar", 4},
		{"Elixir Collector", 6},
	}

	bigSpells := []struct {
		name   string
		elixir int
	}{
		{"Fireball", 4},
		{"Poison", 4},
		{"Rocket", 6},
		{"Lightning", 6},
	}

	smallSpells := []struct {
		name   string
		elixir int
	}{
		{"Zap", 2},
		{"Log", 2},
		{"Arrows", 3},
		{"Barbarian Barrel", 2},
		{"Snowball", 2},
	}

	support := []struct {
		name   string
		elixir int
	}{
		{"Musketeer", 4},
		{"Archers", 3},
		{"Baby Dragon", 4},
		{"Valkyrie", 4},
		{"Wizard", 5},
		{"Electro Wizard", 4},
		{"Hunter", 4},
		{"Witch", 5},
		{"Night Witch", 4},
	}

	cycle := []struct {
		name   string
		elixir int
	}{
		{"Skeletons", 1},
		{"Ice Spirit", 1},
		{"Bats", 2},
		{"Goblin Gang", 3},
		{"Knight", 3},
		{"Ice Golem", 2},
		{"Rascals", 5},
	}

	// Add all cards
	for _, card := range winConditions {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Rare", ElixirCost: card.elixir,
		})
	}
	for _, card := range buildings {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Common", ElixirCost: card.elixir,
		})
	}
	for _, card := range bigSpells {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Epic", ElixirCost: card.elixir,
		})
	}
	for _, card := range smallSpells {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Common", ElixirCost: card.elixir,
		})
	}
	for _, card := range support {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Rare", ElixirCost: card.elixir,
		})
	}
	for _, card := range cycle {
		cards = append(cards, clashroyale.Card{
			Name: card.name, Level: 11, MaxLevel: 14,
			Rarity: "Common", ElixirCost: card.elixir,
		})
	}

	return cards
}

func createLargeCandidateSet(count int) []*CardCandidate {
	candidates := make([]*CardCandidate, 0, count)

	winConditionCards := []struct {
		name   string
		elixir int
	}{
		{"Hog Rider", 4},
		{"Giant", 5},
		{"Golem", 8},
		{"Miner", 3},
		{"Balloon", 5},
		{"Royal Giant", 6},
		{"Goblin Barrel", 3},
		{"Graveyard", 5},
		{"P.E.K.K.A", 7},
		{"Mega Knight", 7},
	}

	buildingCards := []struct {
		name   string
		elixir int
	}{
		{"Cannon", 3},
		{"Tesla", 4},
		{"Inferno Tower", 5},
		{"Bomb Tower", 4},
		{"X-Bow", 6},
		{"Mortar", 4},
	}

	bigSpellCards := []struct {
		name   string
		elixir int
	}{
		{"Fireball", 4},
		{"Poison", 4},
		{"Rocket", 6},
		{"Lightning", 6},
	}

	smallSpellCards := []struct {
		name   string
		elixir int
	}{
		{"Zap", 2},
		{"Log", 2},
		{"Arrows", 3},
		{"Barbarian Barrel", 2},
		{"Snowball", 2},
	}

	supportCards := []struct {
		name   string
		elixir int
	}{
		{"Musketeer", 4},
		{"Archers", 3},
		{"Baby Dragon", 4},
		{"Valkyrie", 4},
		{"Wizard", 5},
		{"Electro Wizard", 4},
	}

	cycleCards := []struct {
		name   string
		elixir int
	}{
		{"Skeletons", 1},
		{"Ice Spirit", 1},
		{"Bats", 2},
		{"Goblin Gang", 3},
		{"Knight", 3},
		{"Ice Golem", 2},
	}

	// Add win conditions
	for i, card := range winConditionCards {
		if len(candidates) >= count {
			break
		}
		role := RoleWinCondition
		candidates = append(candidates, &CardCandidate{
			Name:              card.name,
			Level:             11,
			MaxLevel:          14,
			Elixir:            card.elixir,
			Role:              &role,
			Score:             1.0 + float64(i)*0.01,
			HasEvolution:      i%2 == 0,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 1,
		})
	}

	// Add buildings
	for i, card := range buildingCards {
		if len(candidates) >= count {
			break
		}
		role := RoleBuilding
		candidates = append(candidates, &CardCandidate{
			Name:     card.name,
			Level:    11,
			MaxLevel: 14,
			Elixir:   card.elixir,
			Role:     &role,
			Score:    1.0 + float64(i)*0.01,
		})
	}

	// Add big spells
	for i, card := range bigSpellCards {
		if len(candidates) >= count {
			break
		}
		role := RoleSpellBig
		candidates = append(candidates, &CardCandidate{
			Name:     card.name,
			Level:    11,
			MaxLevel: 14,
			Elixir:   card.elixir,
			Role:     &role,
			Score:    1.0 + float64(i)*0.01,
		})
	}

	// Add small spells
	for i, card := range smallSpellCards {
		if len(candidates) >= count {
			break
		}
		role := RoleSpellSmall
		candidates = append(candidates, &CardCandidate{
			Name:     card.name,
			Level:    11,
			MaxLevel: 14,
			Elixir:   card.elixir,
			Role:     &role,
			Score:    1.0 + float64(i)*0.01,
		})
	}

	// Add support cards
	for i, card := range supportCards {
		if len(candidates) >= count {
			break
		}
		role := RoleSupport
		candidates = append(candidates, &CardCandidate{
			Name:     card.name,
			Level:    11,
			MaxLevel: 14,
			Elixir:   card.elixir,
			Role:     &role,
			Score:    1.0 + float64(i)*0.01,
		})
	}

	// Add cycle cards
	for len(candidates) < count {
		role := RoleCycle
		for _, card := range cycleCards {
			if len(candidates) >= count {
				break
			}
			candidates = append(candidates, &CardCandidate{
				Name:     card.name,
				Level:    11,
				MaxLevel: 14,
				Elixir:   card.elixir,
				Role:     &role,
				Score:    0.8,
			})
		}
	}

	return candidates
}

// BenchmarkParallelGeneration benchmarks parallel deck generation
func BenchmarkParallelGeneration(b *testing.B) {
	candidates := createLargeCandidateSet(50)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			config := GeneratorConfig{
				Strategy:   StrategyRandomSample,
				Candidates: candidates,
				SampleSize: 100,
				Seed:       time.Now().UnixNano(),
			}

			gen, _ := NewDeckGenerator(config)
			ctx := context.Background()
			_, _ = gen.GenerateOne(ctx)
		}
	})
}

// BenchmarkLargeCardCollection benchmarks with varying collection sizes
func BenchmarkLargeCardCollection(b *testing.B) {
	sizes := []int{20, 30, 40, 50, 60, 70, 80, 90, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Cards_%d", size), func(b *testing.B) {
			candidates := createLargeCandidateSet(size)
			config := GeneratorConfig{
				Strategy:   StrategySmartSample,
				Candidates: candidates,
				SampleSize: 1000,
				Seed:       12345,
			}

			gen, _ := NewDeckGenerator(config)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = gen.GenerateOne(ctx)
			}
		})
	}
}

// BenchmarkExhaustiveIterator benchmarks exhaustive iteration
func BenchmarkExhaustiveIterator(b *testing.B) {
	// Use smaller collection for exhaustive testing
	candidates := createLargeCandidateSet(12) // Just enough for valid decks
	config := GeneratorConfig{
		Strategy:   StrategyExhaustive,
		Candidates: candidates,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	iterator, _ := gen.Iterator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, _ := iterator.Next(ctx)
		if deck == nil {
			iterator.Reset()
		}
	}
}

// BenchmarkStringOperations benchmarks string operations in deck generation
func BenchmarkStringOperations(b *testing.B) {
	deck := []string{"Hog Rider", "Musketeer", "Fireball", "Zap", "Cannon", "Ice Spirit", "Skeletons", "Archers"}

	b.Run("Join", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = strings.Join(deck, ", ")
		}
	})

	b.Run("Sort", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sorted := make([]string, len(deck))
			copy(sorted, deck)
			sort.Strings(sorted)
			_ = sorted
		}
	})
}

// BenchmarkThroughput decks per second measurement
func BenchmarkThroughput(b *testing.B) {
	candidates := createLargeCandidateSet(50)
	config := GeneratorConfig{
		Strategy:   StrategySmartSample,
		Candidates: candidates,
		SampleSize: 1000,
		Seed:       12345,
	}

	gen, _ := NewDeckGenerator(config)
	ctx := context.Background()

	start := time.Now()
	count := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decks, _ := gen.Generate(ctx, 100)
		count += len(decks)
	}

	elapsed := time.Since(start)
	rate := float64(count) / elapsed.Seconds()

	b.ReportMetric(rate, "decks/sec")
}
