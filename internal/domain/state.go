package domain

import "strings"

// State selects Task completion independently of archive state.
type State string

const (
	StateOpen State = "open"
	StateDone State = "done"
	StateAll  State = "all"
)

// ParseState accepts the three canonical values case-insensitively. An empty
// value means the default.
func ParseState(raw string) (State, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(StateOpen):
		return StateOpen, nil
	case string(StateDone):
		return StateDone, nil
	case string(StateAll):
		return StateAll, nil
	default:
		return "", NewValidationError("state must be one of open, done, or all (got %q)", raw)
	}
}

// TaskQuery selects which Tasks a listing returns. The two dimensions are
// independent: completion and archive.
type TaskQuery struct {
	State      State
	Visibility Visibility
	ProjectID  *int64
	Priority   *Priority
	Limit      int
}

// Normalized fills in defaults for an unset query.
func (q TaskQuery) Normalized() TaskQuery {
	if q.State == "" {
		q.State = StateOpen
	}
	if q.Visibility == "" {
		q.Visibility = VisibilityActive
	}
	return q
}
