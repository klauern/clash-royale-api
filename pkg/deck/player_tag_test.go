package deck

import "testing"

func TestSanitizePlayerTag(t *testing.T) {
	tag, err := SanitizePlayerTag(" #abc123 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "ABC123" {
		t.Fatalf("expected ABC123, got %s", tag)
	}
}

func TestSanitizePlayerTagInvalid(t *testing.T) {
	_, err := SanitizePlayerTag("../bad")
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
}
