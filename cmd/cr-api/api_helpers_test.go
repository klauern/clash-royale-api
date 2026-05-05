package main

import (
	"strings"
	"testing"
)

func TestResolveAPIToken_PrefersExplicitArg(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "from-env")
	if got := resolveAPIToken("from-flag"); got != "from-flag" {
		t.Errorf("got %q, want %q", got, "from-flag")
	}
}

func TestResolveAPIToken_FallsBackToEnv(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "from-env")
	if got := resolveAPIToken(""); got != "from-env" {
		t.Errorf("got %q, want %q", got, "from-env")
	}
}

func TestResolveAPIToken_EmptyWhenNeitherSet(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "")
	if got := resolveAPIToken(""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestRequireAPITokenValue_FallsBackToEnv(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "from-env")
	got, err := requireAPITokenValue("", apiClientOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from-env" {
		t.Errorf("got %q, want %q", got, "from-env")
	}
}

func TestRequireAPITokenValue_ErrorsWhenMissing(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "")
	_, err := requireAPITokenValue("", apiClientOptions{})
	if err == nil {
		t.Fatal("expected error when both flag and env are empty")
	}
	if !strings.Contains(err.Error(), "API token is required") {
		t.Errorf("error message %q lacks expected guidance", err.Error())
	}
}

func TestRequireAPITokenValue_CustomMissingMessage(t *testing.T) {
	t.Setenv(apiTokenEnvVar, "")
	_, err := requireAPITokenValue("", apiClientOptions{missingToken: "custom hint"})
	if err == nil || err.Error() != "custom hint" {
		t.Errorf("got %v, want exactly %q", err, "custom hint")
	}
}
