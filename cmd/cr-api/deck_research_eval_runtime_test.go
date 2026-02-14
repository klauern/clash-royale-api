package main

import (
	"reflect"
	"testing"
)

func TestParseMethodsList(t *testing.T) {
	methods, err := parseMethodsList("baseline,genetic,constraint,role-first")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"baseline", "genetic", "constraint", "role-first"}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("methods mismatch: got %v want %v", methods, want)
	}
}

func TestParseTagsDefaults(t *testing.T) {
	tags := parseTags(nil)
	if len(tags) != len(defaultResearchTags) {
		t.Fatalf("default tags mismatch: got %d want %d", len(tags), len(defaultResearchTags))
	}
}
