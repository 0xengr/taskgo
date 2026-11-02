package domain

import "fmt"

// ValidationError reports input that is well-formed at the flag level but
// violates a domain rule. The CLI maps it to a usage exit code.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func NewValidationError(format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// NotFoundError reports a well-formed reference that does not resolve. It is
// a domain failure rather than a usage error.
type NotFoundError struct {
	Kind string
	Ref  string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Kind, e.Ref)
}
