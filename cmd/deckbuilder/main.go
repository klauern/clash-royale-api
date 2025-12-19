package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Version information (set via ldflags during build)
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// Version flag
	showVersion := flag.Bool("version", false, "Show version information")

	tag := flag.String("tag", "", "Player tag (with or without #)")
	dataDir := flag.String("data-dir", "data", "Base data directory")
	analysisDir := flag.String("analysis-dir", "data/analysis", "Directory containing card analysis JSON files")
	analysisFile := flag.String("analysis-file", "", "Explicit analysis JSON file to load (overrides --analysis-dir lookup)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("deckbuilder version %s (commit: %s, built: %s)\n", version, commit, buildTime)
		os.Exit(0)
	}

	if *tag == "" {
		log.Fatal("missing required --tag")
	}

	builder := deck.NewBuilder(*dataDir)

	var (
		analysis *deck.CardAnalysis
		err      error
	)

	if *analysisFile != "" {
		analysis, err = builder.LoadAnalysis(*analysisFile)
		if err != nil {
			log.Fatalf("failed to load analysis file: %v", err)
		}
	} else {
		analysis, err = builder.LoadLatestAnalysis(*tag, *analysisDir)
		if err != nil {
			log.Fatalf("failed to load latest analysis: %v", err)
		}
	}

	deckRec, err := builder.BuildDeckFromAnalysis(*analysis)
	if err != nil {
		log.Fatalf("failed to build deck: %v", err)
	}

	if path, err := builder.SaveDeck(deckRec, "", *tag); err == nil {
		fmt.Printf("Saved deck to %s\n\n", path)
	} else {
		log.Printf("warning: could not save deck: %v", err)
	}

	fmt.Printf("Recommended 1v1 deck for %s\n", *tag)
	fmt.Println("================================")
	for i, card := range deckRec.DeckDetail {
		fmt.Printf("%d. %-20s lv.%d/%d (%s) - %d elixir - role: %s - score: %.3f\n",
			i+1, card.Name, card.Level, card.MaxLevel, card.Rarity, card.Elixir, card.Role, card.Score)
	}

	fmt.Printf("\nAverage elixir: %.2f\n", deckRec.AvgElixir)
	if len(deckRec.Notes) > 0 {
		fmt.Println("\nStrategic notes:")
		for _, note := range deckRec.Notes {
			fmt.Printf("- %s\n", note)
		}
	}
}
