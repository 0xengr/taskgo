package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxTitleRunes bounds a normalized title.
const MaxTitleRunes = 500

// NormalizeTitle trims a title, collapses every run of Unicode whitespace to a
// single ASCII space, rejects remaining control characters, and enforces the
// rune limit after normalization.
func NormalizeTitle(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", NewValidationError("title is not valid UTF-8")
	}

	var b strings.Builder
	pendingSpace := false
	for _, r := range raw {
		if unicode.IsSpace(r) {
			if b.Len() > 0 {
				pendingSpace = true
			}
			continue
		}
		if unicode.IsControl(r) {
			return "", NewValidationError("title contains control characters")
		}
		if pendingSpace {
			b.WriteByte(' ')
			pendingSpace = false
		}
		b.WriteRune(r)
	}

	title := b.String()
	if title == "" {
		return "", NewValidationError("title is required")
	}
	if n := utf8.RuneCountInString(title); n > MaxTitleRunes {
		return "", NewValidationError("title must be at most %d characters (got %d)", MaxTitleRunes, n)
	}
	return title, nil
}
