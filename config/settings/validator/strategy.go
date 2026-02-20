package validator

import (
	"fmt"
	"sort"
	"strings"
)

var validStrategies = map[string]struct{}{
	"token_bucket":           {},
	"leaky_bucket":           {},
	"fixed_window_counter":   {},
	"sliding_window_counter": {},
	"sliding_window_log":     {},
}

func ValidateStrategy(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("rateLimiter.strategy: not configured. allowed: %s", strings.Join(sortedKeys(validStrategies), ", "))
	}
	if _, ok := validStrategies[s]; ok {
		return s, nil
	}

	suggestion := closestMatch(s, sortedKeys(validStrategies))
	if suggestion != "" {
		return "", fmt.Errorf("rateLimiter.strategy: unknown value %q. did you mean %q? allowed: %s", s, suggestion, strings.Join(sortedKeys(validStrategies), ", "))
	}
	return "", fmt.Errorf("rateLimiter.strategy: unknown value %q. allowed: %s\n", s, strings.Join(sortedKeys(validStrategies), ", "))
}

// closestMatch returns the most similar string from candidates.
// Returns empty string if no candidate is close enough.
func closestMatch(s string, candidates []string) string {
	minDist := -1
	best := ""
	for _, c := range candidates {
		d := levenshtein(s, c)
		if minDist == -1 || d < minDist {
			minDist = d
			best = c
		}
	}

	if minDist == 0 {
		return ""
	}

	maxLen := max(len(s), len(best))
	if maxLen == 0 {
		return ""
	}
	ratio := float64(minDist) / float64(maxLen)
	if ratio <= 0.4 {
		return best
	}
	return ""
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= la; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			insertion := curr[j-1] + 1
			deletion := prev[j] + 1
			substitution := prev[j-1] + cost

			curr[j] = min(insertion, deletion, substitution)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// sortedKeys returns the keys of a map sorted alphabetically.
func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
