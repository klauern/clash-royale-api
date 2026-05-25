package main

import "testing"

func TestRequireEventAPITokenUsesExplicitToken(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "")

	got, err := requireEventAPIToken("explicit-token")
	if err != nil {
		t.Fatalf("requireEventAPIToken() unexpected error: %v", err)
	}
	if got != "explicit-token" {
		t.Fatalf("requireEventAPIToken() = %q, want %q", got, "explicit-token")
	}
}

func TestRequireEventAPITokenFallsBackToEnvironment(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "env-token")

	got, err := requireEventAPIToken("")
	if err != nil {
		t.Fatalf("requireEventAPIToken() unexpected error: %v", err)
	}
	if got != "env-token" {
		t.Fatalf("requireEventAPIToken() = %q, want %q", got, "env-token")
	}
}

func TestRequireEventAPITokenReturnsRequiredMessageWhenMissing(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "")

	_, err := requireEventAPIToken("")
	if err == nil {
		t.Fatal("requireEventAPIToken() expected error, got nil")
	}
	if err.Error() != requiredAPITokenMessage {
		t.Fatalf("requireEventAPIToken() error = %q, want %q", err.Error(), requiredAPITokenMessage)
	}
}
