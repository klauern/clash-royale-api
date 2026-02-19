package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestParseUnlockedEvolutionsCSV(t *testing.T) {
	t.Parallel()

	got := parseUnlockedEvolutionsCSV("  Skeletons , ,Archers,  Firecracker  ")
	want := []string{"Skeletons", "Archers", "Firecracker"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseUnlockedEvolutionsCSV mismatch: got %v, want %v", got, want)
	}
}

func TestUnlockedEvolutionsFromCommandUsesEnv(t *testing.T) {
	t.Setenv(unlockedEvolutionsEnvVar, "Knight, Archers")
	var got []string

	cmd := &cli.Command{
		Flags: []cli.Flag{unlockedEvolutionsFlag()},
		Action: func(_ context.Context, cmd *cli.Command) error {
			got = unlockedEvolutionsFromCommand(cmd)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"test-command"}); err != nil {
		t.Fatalf("command run failed: %v", err)
	}

	want := []string{"Knight", "Archers"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unlockedEvolutionsFromCommand env parse mismatch: got %v, want %v", got, want)
	}
}

func TestUnlockedEvolutionsFromCommandFlagOverridesEnv(t *testing.T) {
	t.Setenv(unlockedEvolutionsEnvVar, "Knight, Archers")
	var got []string

	cmd := &cli.Command{
		Flags: []cli.Flag{unlockedEvolutionsFlag()},
		Action: func(_ context.Context, cmd *cli.Command) error {
			got = unlockedEvolutionsFromCommand(cmd)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"test-command", "--unlocked-evolutions", "Goblin Drill,Firecracker"}); err != nil {
		t.Fatalf("command run failed: %v", err)
	}

	want := []string{"Goblin Drill", "Firecracker"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unlockedEvolutionsFromCommand flag parse mismatch: got %v, want %v", got, want)
	}
}
