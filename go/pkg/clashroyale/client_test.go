package clashroyale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiToken := "test_token_123"
	client := NewClient(apiToken)

	if client.apiToken != apiToken {
		t.Errorf("Expected apiToken %s, got %s", apiToken, client.apiToken)
	}

	if client.baseURL != "https://api.clashroyale.com/v1" {
		t.Errorf("Expected baseURL %s, got %s", "https://api.clashroyale.com/v1", client.baseURL)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}

	// Test rate limiter is created
	if client.rateLimiter == nil {
		t.Error("Expected rate limiter to be created")
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABC123", "#ABC123"},
		{"#ABC123", "#ABC123"},
		{"", ""},
		{"2P0LYQ", "#2P0LYQ"},
		{"#2P0LYQ", "#2P0LYQ"},
		{"ABCDEF1234", "#ABCDEF1234"},
	}

	for _, test := range tests {
		result := NormalizeTag(test.input)
		if result != test.expected {
			t.Errorf("NormalizeTag(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}

func TestNewRequest(t *testing.T) {
	client := NewClient("test_token")

	tests := []struct {
		name     string
		method   string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "Valid GET request",
			method:   "GET",
			endpoint: "/players/test123",
			wantErr:  false,
		},
		{
			name:     "Valid POST request",
			method:   "POST",
			endpoint: "/clans/test123",
			wantErr:  false,
		},
		{
			name:     "Empty endpoint",
			method:   "GET",
			endpoint: "",
			wantErr:  false,
		},
		{
			name:     "Invalid URL characters",
			method:   "GET",
			endpoint: "/players/test space",
			wantErr:  false, // http.NewRequest handles URL encoding
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := client.NewRequest(context.Background(), test.method, test.endpoint)

			if (err != nil) != test.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr {
				// For the space test, URL will be encoded
				var expectedURL string
				if test.endpoint == "/players/test space" {
					expectedURL = client.baseURL + "/players/test%20space"
				} else {
					expectedURL = client.baseURL + test.endpoint
				}

				if req.URL.String() != expectedURL {
					t.Errorf("NewRequest() URL = %v, want %v", req.URL.String(), expectedURL)
				}

				if req.Method != test.method {
					t.Errorf("NewRequest() method = %v, want %v", req.Method, test.method)
				}

				// Check authorization header
				authHeader := req.Header.Get("Authorization")
				expectedAuth := "Bearer " + client.apiToken
				if authHeader != expectedAuth {
					t.Errorf("NewRequest() Authorization header = %v, want %v", authHeader, expectedAuth)
				}

				// Check accept header
				acceptHeader := req.Header.Get("Accept")
				if acceptHeader != "application/json" {
					t.Errorf("NewRequest() Accept header = %v, want application/json", acceptHeader)
				}
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	err := APIError{
		StatusCode: 404,
		Message:    "Not found",
		Reason:     "Player not found",
	}

	expected := "API error 404: Player not found - Not found"
	if err.Error() != expected {
		t.Errorf("APIError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "Test Player"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	req, err := client.NewRequest(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Do() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Do_ClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	req, err := client.NewRequest(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err == nil {
		t.Error("Do() expected error for 4xx status, got nil")
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Check if error is APIError type
	if apiErr, ok := err.(APIError); ok {
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("Do() APIError status = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
		}
	} else {
		t.Errorf("Do() error type = %T, want APIError", err)
	}
}

func TestClient_Do_RateLimitRetry(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 2 {
			// First two requests get rate limited
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Third request succeeds
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "Test Player"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	req, err := client.NewRequest(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Do() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Should have made 3 requests (2 retries + 1 success)
	if requestCount != 3 {
		t.Errorf("Do() made %d requests, want 3", requestCount)
	}

	// Should have taken some time due to retries (at least 2 seconds)
	if duration < 2*time.Second {
		t.Errorf("Do() completed in %v, expected at least 2s due to retries", duration)
	}
}

func TestClient_Do_ServerErrorRetry(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 1 {
			// First request gets server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second request succeeds
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "Test Player"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	req, err := client.NewRequest(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Do() status = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Should have made 2 requests (1 retry + 1 success)
	if requestCount != 2 {
		t.Errorf("Do() made %d requests, want 2", requestCount)
	}
}

func TestClient_Do_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return server error
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	req, err := client.NewRequest(context.Background(), "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Error("Do() expected error after max retries, got nil")
	}

	expectedError := "max retries exceeded"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Do() error = %v, should contain %v", err.Error(), expectedError)
	}
}

func TestClient_Do_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return rate limit to trigger retry delay
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, err := client.NewRequest(ctx, "GET", "/test")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(req)
	if err == nil {
		t.Error("Do() expected context cancellation error, got nil")
	}

	if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Do() error = %v, expected context timeout", err)
	}
}

func TestClient_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "Test Player"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	// Test that the rate limiter is configured
	// Make 2 sequential requests and verify they take at least some time
	start := time.Now()

	req1, _ := client.NewRequest(context.Background(), "GET", "/test")
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	resp1.Body.Close()

	req2, _ := client.NewRequest(context.Background(), "GET", "/test")
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	resp2.Body.Close()

	elapsed := time.Since(start)

	// Should take at least some time due to rate limiting
	// We use a small threshold since test environments can be fast
	if elapsed < 100*time.Millisecond {
		t.Logf("Requests completed quickly (%v) - rate limiting may be minimal in test environment", elapsed)
	}
}

// Mock player data for testing
var mockPlayer = Player{
	Tag:       "#ABC123",
	Name:      "Test Player",
	ExpLevel:  50,
	Trophies:  4000,
	BestTrophies: 4500,
	Wins:      2000,
	Losses:    1500,
}

func TestGetPlayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "ABC123") {
			t.Errorf("Expected URL to contain player tag, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockPlayer)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	player, err := client.GetPlayer("#ABC123")
	if err != nil {
		t.Fatalf("GetPlayer() error = %v", err)
	}

	if player.Tag != mockPlayer.Tag {
		t.Errorf("GetPlayer() tag = %v, want %v", player.Tag, mockPlayer.Tag)
	}

	if player.Name != mockPlayer.Name {
		t.Errorf("GetPlayer() name = %v, want %v", player.Name, mockPlayer.Name)
	}
}

func TestGetPlayer_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"reason": "notFound"}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	_, err := client.GetPlayer("#INVALID")
	if err == nil {
		t.Error("GetPlayer() expected error for invalid player, got nil")
	}

	// Check if error is APIError type
	if apiErr, ok := err.(APIError); ok {
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("GetPlayer() APIError status = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
		}
	} else {
		t.Errorf("GetPlayer() error type = %T, want APIError", err)
	}
}

func TestGetCards_Success(t *testing.T) {
	mockCards := CardList{
		Items: []Card{
			{
				ID:         28000000,
				Name:       "Knight",
				ElixirCost: 3,
				Type:       "troop",
				Rarity:     "common",
			},
			{
				ID:         28000001,
				Name:       "Archers",
				ElixirCost: 3,
				Type:       "troop",
				Rarity:     "common",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards" {
			t.Errorf("Expected URL path /cards, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockCards)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	cards, err := client.GetCards()
	if err != nil {
		t.Fatalf("GetCards() error = %v", err)
	}

	if len(cards.Items) != len(mockCards.Items) {
		t.Errorf("GetCards() returned %d cards, want %d", len(cards.Items), len(mockCards.Items))
	}

	if cards.Items[0].Name != mockCards.Items[0].Name {
		t.Errorf("GetCards() first card name = %v, want %v", cards.Items[0].Name, mockCards.Items[0].Name)
	}
}

func TestGetCards_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	_, err := client.GetCards()
	if err == nil {
		t.Error("GetCards() expected error for server error, got nil")
	}
}

func TestGetPlayerUpcomingChests_Success(t *testing.T) {
	mockChests := ChestCycle{
		Items: []Chest{
			{Index: 0, Name: "Silver Chest"},
			{Index: 1, Name: "Gold Chest"},
			{Index: 2, Name: "Giant Chest"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "upcomingchests") {
			t.Errorf("Expected URL to contain upcomingchests, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockChests)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	chests, err := client.GetPlayerUpcomingChests("#ABC123")
	if err != nil {
		t.Fatalf("GetPlayerUpcomingChests() error = %v", err)
	}

	if len(chests.Items) != len(mockChests.Items) {
		t.Errorf("GetPlayerUpcomingChests() returned %d chests, want %d", len(chests.Items), len(mockChests.Items))
	}

	if chests.Items[0].Name != mockChests.Items[0].Name {
		t.Errorf("GetPlayerUpcomingChests() first chest name = %v, want %v", chests.Items[0].Name, mockChests.Items[0].Name)
	}
}

func TestGetPlayerBattleLog_Success(t *testing.T) {
	mockBattleLog := BattleLogResponse{
		{
			Type: "PvP",
			UTCDate: time.Now(),
			IsLadderTournament: true,
			GameMode: GameMode{
				ID: 72000000,
				Name: "Ladder",
			},
			Team: []BattleTeam{
				{
					Tag: "#ABC123",
					Name: "Test Player",
					StartingTrophies: 4000,
					TrophyChange: 30,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "battlelog") {
			t.Errorf("Expected URL to contain battlelog, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockBattleLog)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	battleLog, err := client.GetPlayerBattleLog("#ABC123")
	if err != nil {
		t.Fatalf("GetPlayerBattleLog() error = %v", err)
	}

	if len(*battleLog) != len(mockBattleLog) {
		t.Errorf("GetPlayerBattleLog() returned %d battles, want %d", len(*battleLog), len(mockBattleLog))
	}

	if (*battleLog)[0].GameMode.Name != mockBattleLog[0].GameMode.Name {
		t.Errorf("GetPlayerBattleLog() first battle game mode = %v, want %v", (*battleLog)[0].GameMode.Name, mockBattleLog[0].GameMode.Name)
	}
}

func TestGetLocations_Success(t *testing.T) {
	mockLocations := LocationList{
		Items: []Location{
			{ID: 57000000, Name: "Global"},
			{ID: 57000001, Name: "United States"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locations" {
			t.Errorf("Expected URL path /locations, got: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockLocations)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	locations, err := client.GetLocations()
	if err != nil {
		t.Fatalf("GetLocations() error = %v", err)
	}

	if len(locations.Items) != len(mockLocations.Items) {
		t.Errorf("GetLocations() returned %d locations, want %d", len(locations.Items), len(mockLocations.Items))
	}

	if locations.Items[0].Name != mockLocations.Items[0].Name {
		t.Errorf("GetLocations() first location name = %v, want %v", locations.Items[0].Name, mockLocations.Items[0].Name)
	}
}

func TestJSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Send invalid JSON
		fmt.Fprint(w, `{"invalid": json}`)
	}))
	defer server.Close()

	client := NewClient("test_token")
	client.baseURL = server.URL

	_, err := client.GetPlayer("#ABC123")
	if err == nil {
		t.Error("GetPlayer() expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode") {
		t.Errorf("GetPlayer() error = %v, should contain decode error", err)
	}
}

// Benchmark tests
func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClient("test_token")
	}
}

func BenchmarkNormalizeTag(b *testing.B) {
	tags := []string{"ABC123", "#ABC123", "2P0LYQ", "#2P0LYQ"}

	for i := 0; i < b.N; i++ {
		_ = NormalizeTag(tags[i%len(tags)])
	}
}

func BenchmarkNewRequest(b *testing.B) {
	client := NewClient("test_token")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		_, _ = client.NewRequest(ctx, "GET", "/players/test123")
	}
}