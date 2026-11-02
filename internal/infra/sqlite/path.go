package sqlite

import (
	"os"
	"path/filepath"
	"runtime"
)

const databaseFileName = "taskgo.db"

// DefaultPath resolves the database location. TASKGO_DB wins; otherwise the
// platform application-data directory is used (ADR-0002).
func DefaultPath() (string, error) {
	if path := os.Getenv("TASKGO_DB"); path != "" {
		return path, nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "taskgo", databaseFileName), nil
}

func dataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return dir, nil
		}
		return os.UserConfigDir()
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return dir, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}
