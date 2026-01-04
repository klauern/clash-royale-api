// Package evaluation provides comprehensive deck evaluation functionality
package evaluation

import (
	"fmt"
	"net/url"
	"strings"
)

// cardIDMap maps card names to their Clash Royale API card IDs
var cardIDMap = map[string]string{
	// Troops
	"Knight":               "26000000",
	"Archers":              "26000001",
	"Goblins":              "26000002",
	"Giant":                "26000003",
	"P.E.K.K.A":            "26000004",
	"Minions":              "26000005",
	"Balloon":              "26000006",
	"Witch":                "26000007",
	"Barbarians":           "26000008",
	"Golem":                "26000009",
	"Skeletons":            "26000010",
	"Valkyrie":             "26000011",
	"Skeleton Army":        "26000012",
	"Bomber":               "26000013",
	"Musketeer":            "26000014",
	"Baby Dragon":          "26000015",
	"Prince":               "26000016",
	"Wizard":               "26000017",
	"Mini P.E.K.K.A":       "26000018",
	"Spear Goblins":        "26000019",
	"Giant Skeleton":       "26000020",
	"Hog Rider":            "26000021",
	"Minion Horde":         "26000022",
	"Ice Wizard":           "26000023",
	"Royal Giant":          "26000024",
	"Guards":               "26000025",
	"Princess":             "26000026",
	"Dark Prince":          "26000027",
	"Three Musketeers":     "26000028",
	"Lava Hound":           "26000029",
	"Ice Spirit":           "26000030",
	"Fire Spirit":          "26000031",
	"Miner":                "26000032",
	"Sparky":               "26000033",
	"Bowler":               "26000034",
	"Lumberjack":           "26000035",
	"Battle Ram":           "26000036",
	"Inferno Dragon":       "26000037",
	"Ice Golem":            "26000038",
	"Mega Minion":          "26000039",
	"Dart Goblin":          "26000040",
	"Goblin Gang":          "26000041",
	"Electro Wizard":       "26000042",
	"Elite Barbarians":     "26000043",
	"Hunter":               "26000044",
	"Executioner":          "26000045",
	"Bandit":               "26000046",
	"Royal Recruits":       "26000047",
	"Night Witch":          "26000048",
	"Bats":                 "26000049",
	"Royal Ghost":          "26000050",
	"Ram Rider":            "26000051",
	"Zappies":              "26000052",
	"Rascals":              "26000053",
	"Cannon Cart":          "26000054",
	"Mega Knight":          "26000055",
	"Skeleton Barrel":      "26000056",
	"Flying Machine":       "26000057",
	"Wall Breakers":        "26000058",
	"Royal Hogs":           "26000059",
	"Goblin Giant":         "26000060",
	"Fisherman":            "26000061",
	"Magic Archer":         "26000062",
	"Electro Dragon":       "26000063",
	"Firecracker":          "26000064",
	"Mighty Miner":         "26000065",
	"Super Witch":          "26000066",
	"Elixir Golem":         "26000067",
	"Battle Healer":        "26000068",
	"Skeleton King":        "26000069",
	"Super Lava Hound":     "26000070",
	"Super Magic Archer":   "26000071",
	"Archer Queen":         "26000072",
	"Santa Hog Rider":      "26000073",
	"Golden Knight":        "26000074",
	"Super Ice Golem":      "26000075",
	"Monk":                 "26000077",
	"Super Archers":        "26000078",
	"Skeleton Dragons":     "26000080",
	"Terry":                "26000081",
	"Super Mini P.E.K.K.A": "26000082",
	"Mother Witch":         "26000083",
	"Electro Spirit":       "26000084",
	"Electro Giant":        "26000085",
	"Raging Prince":        "26000086",
	"Phoenix":              "26000087",

	// Buildings
	"Cannon":           "27000000",
	"Goblin Hut":       "27000001",
	"Mortar":           "27000002",
	"Inferno Tower":    "27000003",
	"Bomb Tower":       "27000004",
	"Barbarian Hut":    "27000005",
	"Tesla":            "27000006",
	"Elixir Collector": "27000007",
	"X-Bow":            "27000008",
	"Tombstone":        "27000009",
	"Furnace":          "27000010",
	"Goblin Cage":      "27000012",
	"Goblin Drill":     "27000013",
	"Party Hut":        "27000014",

	// Spells
	"Fireball":         "28000000",
	"Arrows":           "28000001",
	"Rage":             "28000002",
	"Rocket":           "28000003",
	"Goblin Barrel":    "28000004",
	"Freeze":           "28000005",
	"Mirror":           "28000006",
	"Lightning":        "28000007",
	"Zap":              "28000008",
	"Poison":           "28000009",
	"Graveyard":        "28000010",
	"The Log":          "28000011",
	"Tornado":          "28000012",
	"Clone":            "28000013",
	"Earthquake":       "28000014",
	"Barbarian Barrel": "28000015",
	"Heal Spirit":      "28000016",
	"Giant Snowball":   "28000017",
	"Royal Delivery":   "28000018",
	"Party Rocket":     "28000020",
}

// DeckLink represents a shareable Clash Royale deck link
type DeckLink struct {
	// URL is the full shareable link
	URL string `json:"url"`

	// ShortURL is a shortened version (if available)
	ShortURL string `json:"short_url,omitempty"`

	// Cards are the card names in the deck
	Cards []string `json:"cards"`

	// CardIDs are the API card IDs
	CardIDs []string `json:"card_ids"`

	// Valid indicates if the link was successfully generated
	Valid bool `json:"valid"`

	// Error contains error message if link generation failed
	Error string `json:"error,omitempty"`
}

// GenerateDeckLink creates a Clash Royale shareable link for the given deck
// The link format is: https://link.clashroyale.com/deck/en?deck=ID1;ID2;ID3;ID4;ID5;ID6;ID7;ID8
func GenerateDeckLink(cardNames []string) *DeckLink {
	link := &DeckLink{
		Cards: cardNames,
	}

	// Validate deck size
	if len(cardNames) != 8 {
		link.Valid = false
		link.Error = fmt.Sprintf("deck must contain exactly 8 cards, got %d", len(cardNames))
		return link
	}

	// Convert card names to IDs
	cardIDs := make([]string, 0, 8)
	for _, cardName := range cardNames {
		// Try exact match first
		cardID, found := cardIDMap[cardName]
		if !found {
			// Try case-insensitive match
			cardID, found = findCardIDCaseInsensitive(cardName)
			if !found {
				link.Valid = false
				link.Error = fmt.Sprintf("unknown card: %s", cardName)
				return link
			}
		}
		cardIDs = append(cardIDs, cardID)
	}

	link.CardIDs = cardIDs

	// Build the URL
	deckParam := strings.Join(cardIDs, ";")
	baseURL := "https://link.clashroyale.com/deck/en"

	// Use url.URL to properly encode parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		link.Valid = false
		link.Error = fmt.Sprintf("failed to parse URL: %v", err)
		return link
	}

	q := u.Query()
	q.Set("deck", deckParam)
	u.RawQuery = q.Encode()

	link.URL = u.String()
	link.Valid = true

	return link
}

// findCardIDCaseInsensitive searches for a card ID with case-insensitive matching
// Also handles partial matches (e.g., "Log" matches "The Log")
func findCardIDCaseInsensitive(cardName string) (string, bool) {
	lowerName := strings.ToLower(cardName)

	// First try exact case-insensitive match
	for name, id := range cardIDMap {
		if strings.ToLower(name) == lowerName {
			return id, true
		}
	}

	// Try partial match (e.g., "Log" should match "The Log")
	for name, id := range cardIDMap {
		lowerMapName := strings.ToLower(name)
		// Check if the search term is contained in the full name
		if strings.Contains(lowerMapName, lowerName) {
			return id, true
		}
	}

	return "", false
}

// ValidateDeckLink checks if a deck link is properly formed and accessible
func ValidateDeckLink(link *DeckLink) error {
	if !link.Valid {
		return fmt.Errorf("invalid link: %s", link.Error)
	}

	if link.URL == "" {
		return fmt.Errorf("empty URL")
	}

	// Parse URL to validate structure
	u, err := url.Parse(link.URL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	// Check host
	if u.Host != "link.clashroyale.com" {
		return fmt.Errorf("invalid host: %s (expected link.clashroyale.com)", u.Host)
	}

	// Check deck parameter
	deckParam := u.Query().Get("deck")
	if deckParam == "" {
		return fmt.Errorf("missing deck parameter")
	}

	// Validate deck parameter format
	parts := strings.Split(deckParam, ";")
	if len(parts) != 8 {
		return fmt.Errorf("deck parameter must contain 8 card IDs, got %d", len(parts))
	}

	return nil
}

// GetCardName returns the card name for a given card ID
func GetCardName(cardID string) string {
	for name, id := range cardIDMap {
		if id == cardID {
			return name
		}
	}
	return ""
}
