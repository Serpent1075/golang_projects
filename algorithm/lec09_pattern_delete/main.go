package main

import (
	"fmt"
	"strings"
)

func removeMostFrequentPattern(text string) (int, string, string) {
	if len(text) == 0 {
		return 0, "", ""
	}

	patterns := make(map[string]int)
	for i := 0; i < len(text); i++ {
		for j := i + 1; j <= len(text); j++ {
			pattern := text[i:j]
			patterns[pattern]++
		}
	}

	maxCount := 0
	maxPattern := ""
	for pattern, count := range patterns {
		if count > maxCount || (count == maxCount && pattern > maxPattern) {
			maxCount = count
			maxPattern = pattern
		}
	}

	return maxCount, maxPattern, strings.ReplaceAll(text, maxPattern, "")
}

func main() {
	text := "abcabcwedabc"
	maxCount, maxPattern, result := removeMostFrequentPattern(text)
	fmt.Printf("원본 문자열: %s\n 패턴: %s\n 카운트: %d\n", text, maxPattern, maxCount)
	fmt.Printf("패턴 제거 후 문자열: %s\n", result)

}
