package domain

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// MaxProjectNameRunes bounds a Project name.
const MaxProjectNameRunes = 200

// foldCaser is stateless and safe for concurrent use.
var foldCaser = cases.Fold()

// Project is a logical workspace that groups Tasks and may be bound to one
// local repository path (ADR-0001).
type Project struct {
	ID         int64
	Name       string
	RepoPath   *string
	CreatedAt  time.Time
	ArchivedAt *time.Time
}

func (p Project) Archived() bool { return p.ArchivedAt != nil }

// NewProject is a validated Project that has not been persisted yet.
type NewProject struct {
	Name     string
	NameKey  string
	RepoPath *string
}

// NormalizeProjectName trims a name, validates it, and computes the unique
// case-folded key described in ADR-0006.
func NormalizeProjectName(raw string) (name, key string, err error) {
	if !utf8.ValidString(raw) {
		return "", "", NewValidationError("project name is not valid UTF-8")
	}

	name = strings.TrimSpace(raw)
	if name == "" {
		return "", "", NewValidationError("project name is required")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", "", NewValidationError("project name contains control characters")
		}
	}
	if n := utf8.RuneCountInString(name); n > MaxProjectNameRunes {
		return "", "", NewValidationError("project name must be at most %d characters (got %d)", MaxProjectNameRunes, n)
	}

	key = foldCaser.String(norm.NFC.String(name))
	return name, key, nil
}
