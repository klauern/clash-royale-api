package main

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestHumanReadableBytes(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{999, "999 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
	}

	for _, tt := range tests {
		got := humanReadableBytes(tt.size)
		if got != tt.want {
			t.Errorf("humanReadableBytes(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestBuildCleanupOptions(t *testing.T) {
	makeCmd := func(args []string) *cli.Command {
		cmd := &cli.Command{
			Flags: []cli.Flag{
				&cli.Float64Flag{Name: "min-score"},
				&cli.IntFlag{Name: "older-than-days"},
				&cli.StringFlag{Name: "archetype"},
			},
		}
		if err := cmd.Run(context.Background(), append([]string{"storage-cleanup-test"}, args...)); err != nil {
			t.Fatalf("failed to run command for test setup: %v", err)
		}
		return cmd
	}

	t.Run("requires at least one filter", func(t *testing.T) {
		cmd := makeCmd(nil)
		_, err := buildCleanupOptions(cmd)
		if err == nil {
			t.Fatalf("expected error for missing filters")
		}
	})

	t.Run("valid min score", func(t *testing.T) {
		cmd := makeCmd([]string{"--min-score", "7.2"})
		opts, err := buildCleanupOptions(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.MinScore != 7.2 {
			t.Fatalf("expected min score 7.2, got %.2f", opts.MinScore)
		}
	})

	t.Run("invalid min score range", func(t *testing.T) {
		cmd := makeCmd([]string{"--min-score", "12.0"})
		if _, err := buildCleanupOptions(cmd); err == nil {
			t.Fatalf("expected error for invalid min score")
		}
	})

	t.Run("invalid older-than-days range", func(t *testing.T) {
		cmd := makeCmd([]string{"--older-than-days", "-1"})
		if _, err := buildCleanupOptions(cmd); err == nil {
			t.Fatalf("expected error for invalid older-than-days")
		}
	})
}
