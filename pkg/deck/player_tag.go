package deck

import (
	"github.com/klauer/clash-royale-api/go/internal/playertag"
)

// SanitizePlayerTag validates a player tag for safe filesystem usage and returns
// a canonical form without the leading #.
func SanitizePlayerTag(playerTag string) (string, error) {
	return playertag.Sanitize(playerTag)
}

// DisplayPlayerTag validates a player tag and returns the canonical display form
// with the leading # preserved.
func DisplayPlayerTag(playerTag string) (string, error) {
	return playertag.Display(playerTag)
}
