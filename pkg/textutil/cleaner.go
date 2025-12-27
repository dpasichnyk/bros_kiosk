package textutil

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9\s]`)
	whitespaceRegex      = regexp.MustCompile(`\s+`)
	leadingCleanupRegex  = regexp.MustCompile(`^[\s,.\-–—:]+`)
)

// NormalizeText prepares text for comparison by lowercasing and removing special chars.
// Logic ported from app.js:normalizeText
func NormalizeText(text string) string {
	s := strings.ToLower(text)
	s = nonAlphanumericRegex.ReplaceAllString(s, "")
	s = whitespaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// StripHTML removes HTML tags from a string.
// Logic ported from app.js:stripHtml
func StripHTML(html string) string {
	if html == "" {
		return ""
	}

	// Replace block tags with space to prevent words merging
	// Simple regex approach as per original JS implementation
	// Note: A proper HTML parser would be better but this maintains behavior parity
	// for the "nonsense" removal refactor.
	blockTags := regexp.MustCompile(`(?i)<\/?(a|li|p|div|br|h[1-6])[^>]*>`)
	s := blockTags.ReplaceAllString(html, " ")

	// Remove all other tags
	allTags := regexp.MustCompile(`<[^>]+>`)
	s = allTags.ReplaceAllString(s, "")

	// Replace nbsp
	s = strings.ReplaceAll(s, "&nbsp;", " ")

	// Normalize whitespace
	s = whitespaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// CleanSummary removes the title from the beginning of the summary if present,
// and performs general cleanup.
func CleanSummary(summary, title string) string {
	summary = StripHTML(summary)
	if summary == "" {
		return ""
	}

	normalizedSummary := NormalizeText(summary)
	normalizedTitle := NormalizeText(title)

	if strings.HasPrefix(normalizedSummary, normalizedTitle) {
		matchLen := findMatchLength(summary, title)
		if matchLen > 0 && matchLen < len(summary) {
			summary = summary[matchLen:]
		} else if matchLen >= len(summary) {
			// Title covers the whole summary
			return ""
		}
	}

	// Clean up leading punctuation
	return strings.TrimSpace(leadingCleanupRegex.ReplaceAllString(summary, ""))
}

// findMatchLength finds the length of the prefix in summary that matches title (fuzzily).
// Logic ported from app.js:findMatchLength
func findMatchLength(summary, title string) int {
	normTitle := NormalizeText(title)
	if normTitle == "" {
		return 0
	}

	normTitleChars := strings.ReplaceAll(normTitle, " ", "")

	summaryRunes := []rune(summary)
	matchedNormChars := 0
	summaryIndex := 0

	for summaryIndex < len(summaryRunes) && matchedNormChars < len(normTitleChars) {
		char := summaryRunes[summaryIndex]
		lowerChar := strings.ToLower(string(char))

		isAlphanumeric := false
		if (lowerChar >= "a" && lowerChar <= "z") || (lowerChar >= "0" && lowerChar <= "9") {
			isAlphanumeric = true
		}

		if isAlphanumeric {
			if lowerChar == string(normTitleChars[matchedNormChars]) {
				matchedNormChars++
			}
		}
		summaryIndex++
	}

	return summaryIndex
}
