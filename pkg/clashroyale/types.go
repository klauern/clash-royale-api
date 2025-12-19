package clashroyale

import (
	"fmt"
	"time"
)

// NormalizeTag ensures the tag has a leading # for API requests
func NormalizeTag(tag string) string {
	if len(tag) > 0 && tag[0] != '#' {
		return "#" + tag
	}
	return tag
}

// Player represents a player profile
type Player struct {
	Tag                   string    `json:"tag"`
	Name                  string    `json:"name"`
	NameSet               bool      `json:"nameSet"`
	ExpLevel              int       `json:"expLevel"`
	ExpPoints             int       `json:"expPoints"`
	Trophies              int       `json:"trophies"`
	BestTrophies          int       `json:"bestTrophies"`
	Wins                  int       `json:"wins"`
	Losses                int       `json:"losses"`
	BattleCount           int       `json:"battleCount"`
	ThreeCrownWins        int       `json:"threeCrownWins"`
	ChallengeWins         int       `json:"challengeWins"`
	ChallengeMaxWins      int       `json:"challengeMaxWins"`
	TournamentWins        int       `json:"tournamentWins"`
	TournamentBattleCount int       `json:"tournamentBattleCount"`
	Role                  string    `json:"role"`
	Clan                  *Clan     `json:"clan,omitempty"`
	Arena                 Arena     `json:"arena"`
	League                League    `json:"league"`
	CurrentDeck           []Card    `json:"currentDeck,omitempty"`
	Cards                 []Card    `json:"cards"`
	StarPoints            int       `json:"starPoints"`
	Donations             int       `json:"donations"`
	TotalDonations        int       `json:"totalDonations"`
	ChallengeCardsWon     int       `json:"challengeCardsWon"`
	Level                 int       `json:"level"`
	Experience            int       `json:"experience"`
	CreatedAt             time.Time `json:"createdAt"`
}

// Clan represents player's clan information
type Clan struct {
	Tag              string    `json:"tag"`
	Name             string    `json:"name"`
	ClanScore        int       `json:"clanScore"`
	ClanScoreData    ClanScore `json:"clanScoreData"`
	Donations        int       `json:"donations"`
	BadgeID          int       `json:"badgeId"`
	Type             string    `json:"type"`
	RequiredTrophies int       `json:"requiredTrophies"`
	ChatFrequency    string    `json:"chatFrequency"`
	Description      string    `json:"description"`
	Members          int       `json:"members"`
	MemberList       []Member  `json:"memberList,omitempty"`
}

// ClanScore represents clan score data
type ClanScore struct {
	Previous         int `json:"previous"`
	Current          int `json:"current"`
	PreviousSeasonID int `json:"previousSeasonId"`
}

// Member represents a clan member
type Member struct {
	Tag               string `json:"tag"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	ExpLevel          int    `json:"expLevel"`
	Trophies          int    `json:"trophies"`
	Arena             Arena  `json:"arena"`
	LastSeen          string `json:"lastSeen"`
	Donations         int    `json:"donations"`
	DonationsReceived int    `json:"donationsReceived"`
	ClanRank          int    `json:"clanRank"`
	PreviousClanRank  int    `json:"previousClanRank"`
}

// Arena represents an arena
type Arena struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Arena       string `json:"arena"`
	ArenaID     int    `json:"arenaId"`
	TrophyLimit int    `json:"trophyLimit"`
}

// League represents a league
type League struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrls,omitempty"`
}

// IconUrls represents the different icon URLs for a card or chest
type IconUrls struct {
	Medium          string `json:"medium,omitempty"`
	EvolutionMedium string `json:"evolutionMedium,omitempty"`
}

// Paging represents cursor-based pagination info
type Paging struct {
	Cursors PagingCursors `json:"cursors,omitempty"`
}

// PagingCursors represents pagination cursors
type PagingCursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

// Card represents a card
type Card struct {
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	Level             int      `json:"level"`
	MaxLevel          int      `json:"maxLevel"`
	Count             int      `json:"count"`
	IconUrls          IconUrls `json:"iconUrls"`
	ElixirCost        int      `json:"elixirCost"`
	Type              string   `json:"type"`
	Rarity            string   `json:"rarity"`
	Description       string   `json:"description,omitempty"`
	EvolutionLevel    int      `json:"evolutionLevel,omitempty"`
	MaxEvolutionLevel int      `json:"maxEvolutionLevel,omitempty"`
	StarLevel         int      `json:"starLevel,omitempty"`
}

// Validate checks if the card data is valid
func (c *Card) Validate() error {
	// Check for negative values first (basic data integrity)
	if c.EvolutionLevel < 0 {
		return fmt.Errorf("evolution level cannot be negative: %d", c.EvolutionLevel)
	}

	if c.MaxEvolutionLevel < 0 {
		return fmt.Errorf("max evolution level cannot be negative: %d", c.MaxEvolutionLevel)
	}

	if c.StarLevel < 0 {
		return fmt.Errorf("star level cannot be negative: %d", c.StarLevel)
	}

	// Check logical constraints
	if c.EvolutionLevel > c.MaxEvolutionLevel {
		return fmt.Errorf("evolution level %d cannot be greater than max evolution level %d",
			c.EvolutionLevel, c.MaxEvolutionLevel)
	}

	return nil
}

// Chest represents an upcoming chest
type Chest struct {
	Name     string   `json:"name"`
	Index    int      `json:"index"`
	IconUrls IconUrls `json:"iconUrls"`
}

// ChestCycle represents the upcoming chest cycle
type ChestCycle struct {
	Items []Chest `json:"items"`
}

// Battle represents a battle entry
type Battle struct {
	Type               string       `json:"type"`
	Team               []BattleTeam `json:"team"`
	Opponent           []BattleTeam `json:"opponent"`
	UTCDate            time.Time    `json:"utcDate"`
	IsLadderTournament bool         `json:"isLadderTournament"`
	GameMode           GameMode     `json:"gameMode"`
	Deck               []Card       `json:"deck,omitempty"`
	DeckAverage        int          `json:"deckAverage,omitempty"`
}

// BattleTeam represents a team in a battle
type BattleTeam struct {
	Tag              string `json:"tag"`
	Name             string `json:"name"`
	StartingTrophies int    `json:"startingTrophies"`
	TrophyChange     int    `json:"trophyChange"`
	Crowns           int    `json:"crowns"`
	Clan             *Clan  `json:"clan,omitempty"`
	Cards            []Card `json:"cards,omitempty"`
}

// GameMode represents a game mode
type GameMode struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	DeckLink   string `json:"deckLink,omitempty"`
	NotCounted bool   `json:"notCounted"`
}

// BattleLogResponse represents the battle log response
type BattleLogResponse []Battle

// CardList represents the response for cards endpoint
type CardList struct {
	Items  []Card `json:"items"`
	Paging Paging `json:"paging,omitempty"`
}

// Location represents a location
type Location struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	IsCountry   bool   `json:"isCountry"`
	CountryCode string `json:"countryCode"`
}

// LocationList represents the response for locations endpoint
type LocationList struct {
	Items  []Location `json:"items"`
	Paging Paging     `json:"paging,omitempty"`
}
