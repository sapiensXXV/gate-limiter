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
		return "", fmt.Errorf("rateLimiter.strategy가 설정되어 있지 않습니다. 유효한 값: %s", strings.Join(sortedKeys(validStrategies), ", "))
	}
	if _, ok := validStrategies[s]; ok {
		return s, nil
	}

	// 가장 가까운 후보를 찾아서 제안
	suggestion := closestMatch(s, sortedKeys(validStrategies))
	if suggestion != "" {
		return "", fmt.Errorf("알 수 없는 rateLimiter.strategy 값 %q입니다. 유사한 값: %q. 사용 가능한 값: %s", s, suggestion, strings.Join(sortedKeys(validStrategies), ", "))
	}
	return "", fmt.Errorf("알 수 없는 rateLimiter.strategy 값 %q 입니다. 사용 가능한 값: %s\n", s, strings.Join(sortedKeys(validStrategies), ", "))
}

// closestMatch: candidates 중 s와 가장 비슷한 문자열을 반환.
// 너무 차이가 크면 빈 문자열을 반환해서 제안하지 않음.
func closestMatch(s string, candidates []string) string {
	// 낮을수록 더 비슷
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
		return "" // 정확히 일치하는 경우 제안 필요 없음
	}

	// 유사도 비율 기준: edit distance / max(len) 이 일정 이하인 경우만 제안
	maxLen := max(len(s), len(best))
	if maxLen == 0 {
		return ""
	}
	ratio := float64(minDist) / float64(maxLen)
	if ratio <= 0.4 { // 40% 이하 차이만 제안 (경험적으로 괜찮은 임계값)
		return best
	}
	return ""
}

// levenshtein 두 문자열 사이의 Levenshtein distance 계산
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// dp테이블을 1차원으로 최적화
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

// sortedKeys 맵의 키를 정렬된 슬라이스로 반환
func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
