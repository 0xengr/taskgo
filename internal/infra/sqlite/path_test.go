package sqlite

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPathHonorsTaskgoDB(t *testing.T) {
	t.Setenv("TASKGO_DB", "/custom/location/taskgo.db")

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath returned error: %v", err)
	}
	if got != "/custom/location/taskgo.db" {
		t.Fatalf("DefaultPath = %q, want the TASKGO_DB value", got)
	}
}

func TestDefaultPathUsesXDGDataHome(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG_DATA_HOME applies to Linux only")
	}
	t.Setenv("TASKGO_DB", "")
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath returned error: %v", err)
	}
	want := filepath.Join(dataHome, "taskgo", "taskgo.db")
	if got != want {
		t.Fatalf("DefaultPath = %q, want %q", got, want)
	}
}
