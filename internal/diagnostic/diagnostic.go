package diagnostic

import (
	"fmt"
	"strings"
)

// Severity indicates the importance of a diagnostic.
type Severity int

const (
	Error   Severity = iota
	Warning
	Hint
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	case Hint:
		return "hint"
	default:
		return "unknown"
	}
}

// Diagnostic represents a located message about a source file.
type Diagnostic struct {
	File       string
	Line       int // 1-indexed
	Column     int // 1-indexed
	Severity   Severity
	Message    string
	Suggestion string // "did you mean X?"
}

// Format returns the diagnostic as a human-readable string in file:line:col format.
func (d Diagnostic) Format() string {
	var b strings.Builder
	if d.File != "" {
		b.WriteString(d.File)
		if d.Line > 0 {
			fmt.Fprintf(&b, ":%d", d.Line)
			if d.Column > 0 {
				fmt.Fprintf(&b, ":%d", d.Column)
			}
		}
		b.WriteString(": ")
	}
	fmt.Fprintf(&b, "%s: %s", d.Severity, d.Message)
	if d.Suggestion != "" {
		fmt.Fprintf(&b, " (%s)", d.Suggestion)
	}
	return b.String()
}

// LevenshteinDistance returns the edit distance between two strings.
func LevenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// SuggestMatch finds the closest match from candidates within the given threshold.
func SuggestMatch(input string, candidates []string, maxDistance int) string {
	best := ""
	bestDist := maxDistance + 1

	for _, c := range candidates {
		d := LevenshteinDistance(input, c)
		if d < bestDist {
			bestDist = d
			best = c
		}
	}

	if best != "" && bestDist <= maxDistance {
		return fmt.Sprintf("did you mean %q?", best)
	}
	return ""
}
