package clashroyale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/ratelimit"
)

// Client represents a Clash Royale API client
type Client struct {
	httpClient  *http.Client
	apiToken    string
	rateLimiter ratelimit.Limiter
	baseURL     string
}

// NewClient creates a new Clash Royale API client
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		rateLimiter: ratelimit.New(1, // 1 request per second
			ratelimit.Per(time.Second),
		),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.clashroyale.com/v1",
	}
}

// APIError represents an error response from the Clash Royale API
type APIError struct {
	StatusCode int
	Message    string
	Reason     string
}

func (e APIError) Error() string {
	return fmt.Sprintf("API error %d: %s - %s", e.StatusCode, e.Reason, e.Message)
}

// NewRequest creates a new HTTP request with proper headers
func (c *Client) NewRequest(ctx context.Context, method, endpoint string) (*http.Request, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// Do performs an HTTP request with retry logic and rate limiting
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Rate limit the request
	c.rateLimiter.Take()

	var resp *http.Response
	var err error

	// Simple retry loop
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(time.Second * time.Duration(attempt)):
			}
		}

		// Clone the request for each attempt
		reqClone := req.Clone(req.Context())
		resp, err = c.httpClient.Do(reqClone)
		if err != nil {
			continue // Network error, retry
		}

		// Check for rate limit (429) or server errors (5xx) - retry these
		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
			if resp.StatusCode == 429 {
				delay := retryAfterDelay(resp, attempt)
				closeWithLog(resp.Body, "response body")
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(delay):
				}
				continue
			}
			closeWithLog(resp.Body, "response body")
			continue
		}

		// Check for client errors (4xx except 429) - don't retry these
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return nil, parseAPIError(resp)
		}

		// Success or other status code
		return resp, nil
	}

	// All retries exhausted
	if resp != nil {
		closeWithLog(resp.Body, "response body")
	}
	return nil, fmt.Errorf("max retries exceeded: %w", err)
}

func retryAfterDelay(resp *http.Response, attempt int) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return time.Second * time.Duration(max(attempt, 1))
	}
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil || seconds <= 0 {
		return time.Second * time.Duration(max(attempt, 1))
	}
	return time.Second * time.Duration(seconds)
}

func parseAPIError(resp *http.Response) APIError {
	defer closeWithLog(resp.Body, "response body")
	payload := struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Client error: %d", resp.StatusCode),
			Reason:     resp.Status,
		}
	}
	if payload.Message == "" {
		payload.Message = fmt.Sprintf("Client error: %d", resp.StatusCode)
	}
	if payload.Reason == "" {
		payload.Reason = resp.Status
	}
	return APIError{
		StatusCode: resp.StatusCode,
		Message:    payload.Message,
		Reason:     payload.Reason,
	}
}
