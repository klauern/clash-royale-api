package clashroyale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GetPlayer retrieves player information for the given tag
func (c *Client) GetPlayer(tag string) (*Player, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s", url.PathEscape(normalizedTag))

	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create player request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Failed to get player %s", tag),
		}
	}

	var player Player
	if err := json.NewDecoder(resp.Body).Decode(&player); err != nil {
		return nil, fmt.Errorf("failed to decode player response: %w", err)
	}

	return &player, nil
}

// GetPlayerUpcomingChests retrieves the upcoming chest cycle for a player
func (c *Client) GetPlayerUpcomingChests(tag string) (*ChestCycle, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s/upcomingchests", url.PathEscape(normalizedTag))

	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create upcoming chests request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Failed to get upcoming chests for player %s", tag),
		}
	}

	var chestCycle ChestCycle
	if err := json.NewDecoder(resp.Body).Decode(&chestCycle); err != nil {
		return nil, fmt.Errorf("failed to decode chest cycle response: %w", err)
	}

	return &chestCycle, nil
}

// GetPlayerBattleLog retrieves the battle log for a player
func (c *Client) GetPlayerBattleLog(tag string) (*BattleLogResponse, error) {
	normalizedTag := NormalizeTag(tag)
	endpoint := fmt.Sprintf("/players/%s/battlelog", url.PathEscape(normalizedTag))

	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create battle log request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Failed to get battle log for player %s", tag),
		}
	}

	var battleLog BattleLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&battleLog); err != nil {
		return nil, fmt.Errorf("failed to decode battle log response: %w", err)
	}

	return &battleLog, nil
}

// GetCards retrieves the full list of cards
func (c *Client) GetCards() (*CardList, error) {
	endpoint := "/cards"

	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create cards request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get cards",
		}
	}

	var cards CardList
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return nil, fmt.Errorf("failed to decode cards response: %w", err)
	}

	return &cards, nil
}

// GetLocations retrieves the list of locations
func (c *Client) GetLocations() (*LocationList, error) {
	endpoint := "/locations"

	req, err := c.NewRequest(context.Background(), "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create locations request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeWithLog(resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get locations",
		}
	}

	var locations LocationList
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("failed to decode locations response: %w", err)
	}

	return &locations, nil
}
