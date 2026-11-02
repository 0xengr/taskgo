package domain

import "strings"

// Priority is an ordinal classification of importance. It is deliberately not
// a numeric score: see ADR-0003 and the glossary.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// ParsePriority accepts the three canonical values case-insensitively. An
// empty value means "use the default".
func ParsePriority(raw string) (Priority, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(PriorityMedium):
		return PriorityMedium, nil
	case string(PriorityLow):
		return PriorityLow, nil
	case string(PriorityHigh):
		return PriorityHigh, nil
	default:
		return "", NewValidationError("priority must be one of low, medium, or high (got %q)", raw)
	}
}

func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// Rank orders priorities for selection. Lower is more important, so a
// ascending sort puts high first.
func (p Priority) Rank() int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	default:
		return 3
	}
}
