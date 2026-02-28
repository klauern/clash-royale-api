package playertag

import (
	"fmt"
	"regexp"
	"strings"
)

var pattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Sanitize validates and canonicalizes a player tag for storage and display.
// It trims whitespace, removes leading '#', enforces alnum chars, and uppercases.
func Sanitize(playerTag string) (string, error) {
	tag := strings.TrimPrefix(strings.TrimSpace(playerTag), "#")
	if tag == "" {
		return "", fmt.Errorf("player tag is required")
	}
	if !pattern.MatchString(tag) {
		return "", fmt.Errorf("invalid player tag: must contain only letters and digits")
	}
	return strings.ToUpper(tag), nil
}

// Display validates and canonicalizes a player tag for user-facing output.
func Display(playerTag string) (string, error) {
	tag, err := Sanitize(playerTag)
	if err != nil {
		return "", err
	}
	return "#" + tag, nil
}
