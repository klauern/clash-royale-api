package playertag

import "testing"

func TestSanitize(t *testing.T) {
	t.Run("normalizes and uppercases", func(t *testing.T) {
		got, err := Sanitize(" #abc123 ")
		if err != nil {
			t.Fatalf("Sanitize returned error: %v", err)
		}
		if got != "ABC123" {
			t.Fatalf("Sanitize = %q, want %q", got, "ABC123")
		}
	})

	t.Run("rejects invalid characters", func(t *testing.T) {
		if _, err := Sanitize("../bad"); err == nil {
			t.Fatal("expected error for invalid player tag")
		}
	})
}

func TestDisplay(t *testing.T) {
	got, err := Display(" abc123 ")
	if err != nil {
		t.Fatalf("Display returned error: %v", err)
	}
	if got != "#ABC123" {
		t.Fatalf("Display = %q, want %q", got, "#ABC123")
	}
}
