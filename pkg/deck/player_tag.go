package deck

import (
	"fmt"
	"regexp"
	"strings"
)

var playerTagPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// SanitizePlayerTag validates a player tag for safe filesystem usage and returns
// a canonical form without the leading #.
func SanitizePlayerTag(playerTag string) (string, error) {
	tag := strings.TrimPrefix(strings.TrimSpace(playerTag), "#")
	if tag == "" {
		return "", fmt.Errorf("player tag is required")
	}
	if !playerTagPattern.MatchString(tag) {
		return "", fmt.Errorf("invalid player tag: must contain only letters and digits")
	}
	return strings.ToUpper(tag), nil
}
