package storage

import "testing"

func TestSanitizePlayerTag(t *testing.T) {
	t.Run("normalizes and uppercases", func(t *testing.T) {
		got, err := SanitizePlayerTag(" #abc123 ")
		if err != nil {
			t.Fatalf("SanitizePlayerTag returned error: %v", err)
		}
		if got != "ABC123" {
			t.Fatalf("SanitizePlayerTag = %q, want %q", got, "ABC123")
		}
	})

	t.Run("rejects invalid characters", func(t *testing.T) {
		if _, err := SanitizePlayerTag("../bad"); err == nil {
			t.Fatal("expected error for invalid player tag")
		}
	})
}
