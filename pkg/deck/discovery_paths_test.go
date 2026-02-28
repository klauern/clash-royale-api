package deck

import (
	"path/filepath"
	"testing"
)

func TestDiscoverySessionBaseDir(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	got := DiscoverySessionBaseDir()
	want := filepath.Join(homeDir, ".cr-api", "discover")
	if got != want {
		t.Fatalf("DiscoverySessionBaseDir() = %q, want %q", got, want)
	}
}

func TestDiscoveryArtifactPathsForTag(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	const tag = "P2LQGRJCU"
	baseDir := filepath.Join(homeDir, ".cr-api", "discover")

	if got, want := DiscoveryCheckpointPathForTag(tag), filepath.Join(baseDir, tag+".json"); got != want {
		t.Fatalf("DiscoveryCheckpointPathForTag() = %q, want %q", got, want)
	}
	if got, want := DiscoveryPIDPathForTag(tag), filepath.Join(baseDir, tag+".pid"); got != want {
		t.Fatalf("DiscoveryPIDPathForTag() = %q, want %q", got, want)
	}
	if got, want := DiscoveryLogPathForTag(tag), filepath.Join(baseDir, tag+".log"); got != want {
		t.Fatalf("DiscoveryLogPathForTag() = %q, want %q", got, want)
	}
}
