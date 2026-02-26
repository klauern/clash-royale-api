package deck

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverySessionBaseDir returns the base directory for discovery session artifacts.
func DiscoverySessionBaseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cr-api", "discover")
}

// DiscoveryCheckpointPathForTag returns the discovery checkpoint path for a sanitized player tag.
func DiscoveryCheckpointPathForTag(playerTag string) string {
	return filepath.Join(DiscoverySessionBaseDir(), fmt.Sprintf("%s.json", playerTag))
}

// DiscoveryPIDPathForTag returns the discovery PID file path for a sanitized player tag.
func DiscoveryPIDPathForTag(playerTag string) string {
	return filepath.Join(DiscoverySessionBaseDir(), fmt.Sprintf("%s.pid", playerTag))
}

// DiscoveryLogPathForTag returns the discovery log file path for a sanitized player tag.
func DiscoveryLogPathForTag(playerTag string) string {
	return filepath.Join(DiscoverySessionBaseDir(), fmt.Sprintf("%s.log", playerTag))
}
