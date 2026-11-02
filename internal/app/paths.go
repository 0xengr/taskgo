package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xengr/taskgo/internal/domain"
)

// CanonicalRepoPath validates that raw names an existing directory and returns
// its canonical, symlink-resolved absolute path (ADR-0005).
func CanonicalRepoPath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", domain.NewValidationError("repository path is required")
	}

	absolute, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}

	info, err := os.Stat(absolute)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("repository path does not exist: %s", absolute)
		}
		return "", fmt.Errorf("inspect repository path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", absolute)
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}
	return resolved, nil
}
