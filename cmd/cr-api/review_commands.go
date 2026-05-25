package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

const reviewFormatFlag = "format"

func addReviewCommand() *cli.Command {
	return &cli.Command{
		Name:  "review",
		Usage: "Full player review: profile, playstyle, top archetype, upgrade priorities, and budget",
		Description: "Runs all major analyzers for a player in a single call and outputs a " +
			"consolidated report covering profile snapshot, playstyle summary, top detected " +
			"archetype, top-3 cross-archetype upgrade priorities, and a budget-aware " +
			"'next 20k gold' suggestion. Use --format=json or --format=markdown for " +
			"structured output.",
		Flags: []cli.Flag{
			playerTagFlag(true),
			&cli.StringFlag{
				Name:  reviewFormatFlag,
				Usage: "Output format: human, json, markdown",
				Value: batchFormatHuman,
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					if s != "" && s != batchFormatHuman && s != batchFormatJSON && s != batchFormatMarkdown {
						return fmt.Errorf("invalid format: %q (must be: human, json, or markdown)", s)
					}
					return nil
				},
			},
		},
		Action: reviewCommandAction,
	}
}

func reviewCommandAction(ctx context.Context, cmd *cli.Command) error {
	format := cmd.String(reviewFormatFlag)
	if format == "" {
		format = batchFormatHuman
	}

	report, err := runReview(ctx, cmd)
	if err != nil {
		return err
	}

	switch format {
	case batchFormatJSON:
		return renderReviewJSON(report)
	case batchFormatMarkdown:
		return renderReviewMarkdown(report)
	default:
		renderReviewHuman(report)
		return nil
	}
}
