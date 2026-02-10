package clashroyale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// makeAPIRequest is a generic helper to reduce duplication across API endpoints.
// It handles the common pattern of: create request, execute, check status, decode JSON.
func makeAPIRequest[T any](c *Client, endpoint, errorMsg string) (*T, error) {
	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    errorMsg,
		}
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetPlayer retrieves player information for the given tag
func (c *Client) GetPlayer(tag string) (*Player, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s", url.PathEscape(normalizedTag))
	return makeAPIRequest[Player](c, endpoint, fmt.Sprintf("Failed to get player %s", tag))
}

// GetPlayerUpcomingChests retrieves the upcoming chest cycle for a player
func (c *Client) GetPlayerUpcomingChests(tag string) (*ChestCycle, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s/upcomingchests", url.PathEscape(normalizedTag))
	return makeAPIRequest[ChestCycle](c, endpoint, fmt.Sprintf("Failed to get upcoming chests for player %s", tag))
}

// GetPlayerBattleLog retrieves the battle log for a player
func (c *Client) GetPlayerBattleLog(tag string) (*BattleLogResponse, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s/battlelog", url.PathEscape(normalizedTag))
	return makeAPIRequest[BattleLogResponse](c, endpoint, fmt.Sprintf("Failed to get battle log for player %s", tag))
}

// GetCards retrieves the full list of cards
func (c *Client) GetCards() (*CardList, error) {
	return makeAPIRequest[CardList](c, "/cards", "Failed to get cards")
}

// GetLocations retrieves the list of locations
func (c *Client) GetLocations() (*LocationList, error) {
	return makeAPIRequest[LocationList](c, "/locations", "Failed to get locations")
}
