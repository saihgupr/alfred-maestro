package main

import (
	"sort"
	"strings"
	"unicode"
)

// ScoredMacro associates a KmMacro with its computed match score.
type ScoredMacro struct {
	Macro KmMacro
	Score float64
}

// tokenize splits a string into lowercase alphanumeric words, splitting on
// whitespace, punctuation, symbols, and camelCase boundaries.
func tokenize(s string) []string {
	var words []string
	var cur []rune

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Check for camelCase boundary (e.g. "HomeAssistant" -> "Home", "Assistant")
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) && len(cur) > 0 {
			words = append(words, string(cur))
			cur = nil
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur = append(cur, unicode.ToLower(r))
		} else {
			if len(cur) > 0 {
				words = append(words, string(cur))
				cur = nil
			}
		}
	}
	if len(cur) > 0 {
		words = append(words, string(cur))
	}
	return words
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min3(a, b, c int) int {
	return min(min(a, b), c)
}

// minSubstringDistance computes the minimum edit distance between query token q
// and any contiguous substring of target word w.
// This allows matching words even with prefix/suffix differences or internal typos
// (e.g. "ssinta" against "assistant" has distance 1).
func minSubstringDistance(q, w string) int {
	qRunes, wRunes := []rune(q), []rune(w)
	qn, wn := len(qRunes), len(wRunes)
	if qn == 0 {
		return 0
	}
	if wn == 0 {
		return qn
	}

	dp := make([][]int, qn+1)
	for i := range dp {
		dp[i] = make([]int, wn+1)
		dp[i][0] = i
	}
	for j := 0; j <= wn; j++ {
		dp[0][j] = 0 // Free start anywhere in w
	}

	for i := 1; i <= qn; i++ {
		for j := 1; j <= wn; j++ {
			cost := 1
			if qRunes[i-1] == wRunes[j-1] {
				cost = 0
			}
			dp[i][j] = min3(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}

	minDist := qn
	for j := 1; j <= wn; j++ {
		if dp[qn][j] < minDist {
			minDist = dp[qn][j]
		}
	}
	return minDist
}

// isSubsequence checks if all runes of q appear in s in sequential order.
func isSubsequence(q, s string) bool {
	qRunes, sRunes := []rune(q), []rune(s)
	qi := 0
	for _, r := range sRunes {
		if qi < len(qRunes) && qRunes[qi] == r {
			qi++
		}
	}
	return qi == len(qRunes)
}

// matchAcronym checks if query characters match the initials of consecutive words
// in wordList (e.g. "ha" matching "Home Assistant", "km" matching "Keyboard Maestro").
func matchAcronym(q string, wordList []string) bool {
	if len(q) < 2 || len(wordList) < len(q) {
		return false
	}
	qRunes := []rune(strings.ToLower(q))

	// Search for any starting word index where consecutive words match initials
	for start := 0; start <= len(wordList)-len(qRunes); start++ {
		matched := true
		for i := 0; i < len(qRunes); i++ {
			word := wordList[start+i]
			if len(word) == 0 || rune(word[0]) != qRunes[i] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

// ScoreMacro evaluates how well a macro matches a search query.
// Returns (score, matched). Higher score means better match.
func ScoreMacro(m KmMacro, rawQuery string) (float64, bool) {
	query := strings.TrimSpace(strings.ToLower(rawQuery))
	if query == "" {
		return 0, true
	}

	nameLower := strings.ToLower(m.Name)
	catLower := strings.ToLower(m.Category)
	nameWords := tokenize(nameLower)
	catWords := tokenize(catLower)
	qTokens := tokenize(query)

	lenPenalty := float64(len(nameLower)) * 2.0

	// 1. Exact full name match
	if nameLower == query {
		return 100000.0 - lenPenalty, true
	}

	// 2. Exact prefix match on name
	if strings.HasPrefix(nameLower, query) {
		return 50000.0 - float64(len(nameLower)-len(query))*5, true
	}

	// 3. Exact word match on name (e.g. query "ram" matching word "ram" in "Snapshot Home Assistant RAM")
	for _, w := range nameWords {
		if w == query {
			return 35000.0 - lenPenalty, true
		}
	}

	// 4. Exact contiguous substring in name
	if idx := strings.Index(nameLower, query); idx >= 0 {
		isWordBoundary := (idx == 0 || !unicode.IsLetter(rune(nameLower[idx-1])))
		base := 15000.0
		if isWordBoundary {
			base = 25000.0
		}
		score := base - float64(idx)*15.0 - lenPenalty
		return score, true
	}

	// 5. Acronym / Initials match for single token (e.g. "ha" -> "Home Assistant", "km" -> "Keyboard Maestro")
	if len(qTokens) == 1 && matchAcronym(query, nameWords) {
		return 18000.0 - lenPenalty, true
	}

	// 6. Multi-token scoring (handles multiple words, out-of-order words, typos, and token initials)
	if len(qTokens) > 0 {
		totalScore := 0.0
		matchedTokens := 0
		lastFoundPos := -1
		inOrder := true

		for _, token := range qTokens {
			bestTokenScore := 0.0
			tokenPos := -1

			// Check against name words
			for wIdx, w := range nameWords {
				var s float64
				if w == token {
					s = 2000.0
				} else if strings.HasPrefix(w, token) {
					s = 1300.0 + 200.0*(float64(len(token))/float64(len(w)))
				} else if strings.Contains(w, token) {
					s = 900.0 + 100.0*(float64(len(token))/float64(len(w)))
				} else {
					// Typo tolerance via substring edit distance
					tLen := len(token)
					allowedDist := 0
					if tLen >= 4 && tLen <= 5 {
						allowedDist = 1
					} else if tLen >= 6 {
						allowedDist = 2
					}

					if allowedDist > 0 {
						dist := minSubstringDistance(token, w)
						if dist <= allowedDist {
							s = (750.0 - float64(dist)*200.0) * (1.0 - float64(dist)/float64(tLen))
						}
					}
				}

				if s > bestTokenScore {
					bestTokenScore = s
					tokenPos = wIdx
				}
			}

			// Check acronym match for this token (e.g. query "restart ha", token "ha" matches "Home Assistant")
			if bestTokenScore < 1500.0 && len(token) >= 2 && matchAcronym(token, nameWords) {
				bestTokenScore = 1500.0
			}

			// Check category if token wasn't strongly matched in macro name
			if bestTokenScore < 800.0 && len(catWords) > 0 {
				for _, cw := range catWords {
					if cw == token {
						if 700.0 > bestTokenScore {
							bestTokenScore = 700.0
						}
					} else if strings.HasPrefix(cw, token) {
						if 500.0 > bestTokenScore {
							bestTokenScore = 500.0
						}
					}
				}
			}

			if bestTokenScore > 0 {
				matchedTokens++
				totalScore += bestTokenScore
				if tokenPos >= 0 {
					if tokenPos < lastFoundPos {
						inOrder = false
					}
					lastFoundPos = tokenPos
				}
			}
		}

		// Required matches:
		// 1 token query: must match
		// 2 token query: both must match
		// 3+ token query: at least (N - 1) tokens must match
		minRequired := len(qTokens)
		if len(qTokens) >= 3 {
			minRequired = len(qTokens) - 1
		}

		if matchedTokens >= minRequired && matchedTokens > 0 {
			bonus := float64(matchedTokens) * 1500.0
			if matchedTokens == len(qTokens) {
				bonus += 5000.0 // All tokens matched bonus
			}
			if inOrder && matchedTokens > 1 {
				bonus += 1500.0 // Preserved word order bonus
			}
			score := totalScore + bonus - lenPenalty
			return score, true
		}
	}

	// 7. Category match (when query is a single term matching category name)
	if catLower != "" {
		if catLower == query {
			return 14000.0 - lenPenalty, true
		}
		if strings.HasPrefix(catLower, query) {
			return 10000.0 - lenPenalty, true
		}
		for _, cw := range catWords {
			if cw == query {
				return 8000.0 - lenPenalty, true
			}
		}
	}

	// 8. Hotkey or TriggerString match
	if m.Hotkey != "" && strings.Contains(strings.ToLower(m.Hotkey), query) {
		return 6000.0 - lenPenalty, true
	}
	if m.TriggerString != "" && strings.Contains(strings.ToLower(m.TriggerString), query) {
		return 6000.0 - lenPenalty, true
	}

	// 9. Fallback loose subsequence match ONLY for queries > 3 chars
	// (Prevents short words like "ram" from matching arbitrary non-contiguous letters)
	if len(query) > 3 && isSubsequence(query, nameLower) {
		return 600.0 - lenPenalty, true
	}

	return 0, false
}

// SearchMacros filters and ranks a slice of KmMacro based on query.
// If query is empty, all macros are returned in their original order.
func SearchMacros(macros []KmMacro, query string) []KmMacro {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return macros
	}

	var scored []ScoredMacro
	for _, m := range macros {
		score, matched := ScoreMacro(m, trimmed)
		if matched && score > 0 {
			scored = append(scored, ScoredMacro{
				Macro: m,
				Score: score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		if len(scored[i].Macro.Name) != len(scored[j].Macro.Name) {
			return len(scored[i].Macro.Name) < len(scored[j].Macro.Name)
		}
		return scored[i].Macro.Name < scored[j].Macro.Name
	})

	results := make([]KmMacro, len(scored))
	for i, sm := range scored {
		results[i] = sm.Macro
	}
	return results
}
