package dictation

import (
	"regexp"
	"strings"
)

var pasteCommandRe = regexp.MustCompile(`(?i)(?:^|[\s,.-])(?:вставить|вставляй|вставьте|вставь|paste)[\s.!,-]*$`)

const previewLimit = 60

// SplitPasteCommand strips a trailing paste command from text and reports
// whether one was found.
func SplitPasteCommand(text string) (string, bool) {
	loc := pasteCommandRe.FindStringIndex(text)
	if loc == nil {
		return text, false
	}
	trimmed := strings.TrimSpace(strings.TrimRight(text[:loc[0]], ",.-! "))
	return trimmed, true
}

func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "..."
}
