package leaderboard

import (
	"io"
	"log"
)

func closeWithLog(closer io.Closer, resource string) {
	if err := closer.Close(); err != nil {
		log.Printf("leaderboard: failed to close %s: %v", resource, err)
	}
}
