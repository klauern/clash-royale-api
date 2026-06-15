package deck

import (
	"fmt"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/internal/datapath"
)

// DiscoverySessionBaseDir returns the base directory for discovery session artifacts.
func DiscoverySessionBaseDir() string {
	return datapath.DiscoveryDirOrFallback()
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
