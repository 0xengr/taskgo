package domain

import "strings"

// Visibility selects archived state independently of completion.
type Visibility string

const (
	VisibilityActive   Visibility = "active"
	VisibilityArchived Visibility = "archived"
	VisibilityAll      Visibility = "all"
)

// ParseVisibility accepts the three canonical values case-insensitively. An
// empty value means the default.
func ParseVisibility(raw string) (Visibility, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(VisibilityActive):
		return VisibilityActive, nil
	case string(VisibilityArchived):
		return VisibilityArchived, nil
	case string(VisibilityAll):
		return VisibilityAll, nil
	default:
		return "", NewValidationError("visibility must be one of active, archived, or all (got %q)", raw)
	}
}
