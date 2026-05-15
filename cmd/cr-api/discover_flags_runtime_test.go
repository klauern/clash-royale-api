package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestDiscoverRunFlagsParity(t *testing.T) {
	t.Parallel()

	withoutResume := discoverRunFlags(false)
	withResume := discoverRunFlags(true)

	withoutNames := flagNames(withoutResume)
	withNames := flagNames(withResume)

	if len(withNames) != len(withoutNames)+1 {
		t.Fatalf("expected one additional flag, got without=%d with=%d", len(withoutNames), len(withNames))
	}

	if _, ok := withNames["resume"]; !ok {
		t.Fatalf("expected resume flag when includeResume=true")
	}
	if _, ok := withoutNames["resume"]; ok {
		t.Fatalf("did not expect resume flag when includeResume=false")
	}

	delete(withNames, "resume")
	if !reflect.DeepEqual(withNames, withoutNames) {
		t.Fatalf("discover start/run flag sets differ:\nwithout=%v\nwith=%v", withoutNames, withNames)
	}
}

func TestBuildDiscoverRunArgs(t *testing.T) {
	t.Parallel()

	var got []string
	cmd := &cli.Command{
		Flags: discoverRunFlags(true),
		Action: func(_ context.Context, cmd *cli.Command) error {
			got = buildDiscoverRunArgs(cmd, "P2ABC", false)
			return nil
		},
	}

	err := cmd.Run(context.Background(), []string{
		"discover-test",
		"--tag=P2ABC",
		"--strategy=genetic",
		"--sample-size=777",
		"--generations=12",
		"--population=64",
		"--mutation-rate=0.25",
		"--crossover-rate=0.9",
		"--island-model",
		"--island-count=6",
		"--migration-interval=8",
		"--migration-size=3",
		"--limit=123",
		"--verbose",
	})
	if err != nil {
		t.Fatalf("command run failed: %v", err)
	}

	want := []string{
		"deck", "discover", "run",
		"--tag=#P2ABC",
		"--strategy=genetic",
		"--sample-size=777",
		"--generations=12",
		"--population=64",
		"--mutation-rate=0.25",
		"--crossover-rate=0.90",
		"--island-count=6",
		"--migration-interval=8",
		"--migration-size=3",
		"--limit=123",
		"--verbose",
		"--island-model",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildDiscoverRunArgs mismatch:\n got: %v\nwant: %v", got, want)
	}
}

func TestBuildDiscoverRunArgsResumeBackground(t *testing.T) {
	t.Parallel()

	var got []string
	cmd := &cli.Command{
		Flags: discoverRunFlags(false),
		Action: func(_ context.Context, cmd *cli.Command) error {
			got = buildDiscoverRunArgs(cmd, "P2ABC", true)
			return nil
		},
	}

	err := cmd.Run(context.Background(), []string{
		"discover-test",
		"--tag=P2ABC",
		"--background",
		"--verbose",
	})
	if err != nil {
		t.Fatalf("command run failed: %v", err)
	}

	want := []string{
		"deck", "discover", "run",
		"--resume",
		"--tag=#P2ABC",
		"--verbose",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildDiscoverRunArgs resume mismatch:\n got: %v\nwant: %v", got, want)
	}
}

func flagNames(flags []cli.Flag) map[string]struct{} {
	names := make(map[string]struct{}, len(flags))
	for _, flag := range flags {
		names[flag.Names()[0]] = struct{}{}
	}
	return names
}
