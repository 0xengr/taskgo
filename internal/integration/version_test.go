package integration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "version")
	if got.exit != 0 {
		t.Fatalf("exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "taskgo") || !strings.Contains(got.stdout, "dev") {
		t.Fatalf("version stdout = %q", got.stdout)
	}

	asJSON := run(t, db, "version", "--json")
	var payload struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.Unmarshal([]byte(asJSON.stdout), &payload); err != nil {
		t.Fatalf("decode version JSON %q: %v", asJSON.stdout, err)
	}
	if payload.Version == "" || payload.Commit == "" {
		t.Fatalf("payload = %+v, want version and commit", payload)
	}

	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Fatalf("version created a database (%v)", err)
	}
}
