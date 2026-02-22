package diagnostic

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"kitten", "sitting", 3},
		{"hello", "hello", 0},
		{"abc", "abd", 1},
		{"palette", "palatte", 1},
		{"sprite", "sprit", 1},
		{"noise", "noize", 1},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestSuggestMatch(t *testing.T) {
	candidates := []string{"red", "green", "blue", "yellow"}

	tests := []struct {
		input    string
		maxDist  int
		expected string
	}{
		{"red", 2, `did you mean "red"?`},    // exact match (distance 0 <= maxDist)
		{"ree", 2, `did you mean "red"?`},    // distance 1
		{"gren", 2, `did you mean "green"?`}, // distance 1
		{"bue", 2, `did you mean "blue"?`},   // distance 1
		{"xyz", 2, ""},                       // too far
		{"yellw", 2, `did you mean "yellow"?`},
	}

	for _, tt := range tests {
		got := SuggestMatch(tt.input, candidates, tt.maxDist)
		if got != tt.expected {
			t.Errorf("SuggestMatch(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSuggestMatch_ExactMatch(t *testing.T) {
	// Exact match has distance 0, which is <= maxDistance, so it returns suggestion.
	// But in practice callers only call SuggestMatch for unknown keys.
	got := SuggestMatch("red", []string{"red", "blue"}, 2)
	if got != `did you mean "red"?` {
		t.Errorf("exact match should still suggest: got %q", got)
	}
}

func TestDiagnostic_Format(t *testing.T) {
	tests := []struct {
		diag     Diagnostic
		expected string
	}{
		{
			Diagnostic{File: "test.sprite", Line: 15, Column: 8, Severity: Error, Message: "unknown key"},
			"test.sprite:15:8: error: unknown key",
		},
		{
			Diagnostic{File: "test.sprite", Line: 15, Column: 8, Severity: Error, Message: "unknown key 'bue'", Suggestion: `did you mean "blue"?`},
			`test.sprite:15:8: error: unknown key 'bue' (did you mean "blue"?)`,
		},
		{
			Diagnostic{Severity: Warning, Message: "ragged row"},
			"warning: ragged row",
		},
	}

	for _, tt := range tests {
		got := tt.diag.Format()
		if got != tt.expected {
			t.Errorf("Format() = %q, want %q", got, tt.expected)
		}
	}
}
