package integration

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "cannot locate the integration test source")
		os.Exit(1)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	dir, err := os.MkdirTemp("", "taskgo-binary-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create binary directory: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(dir, "taskgo")

	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/taskgo")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build taskgo: %v\n%s", err, output)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

type result struct {
	stdout string
	stderr string
	exit   int
}

func run(t *testing.T, dbPath string, args ...string) result {
	t.Helper()
	return runIn(t, t.TempDir(), dbPath, args...)
}

func runIn(t *testing.T, dir, dbPath string, args ...string) result {
	t.Helper()

	got, err := execute(dir, dbPath, args...)
	if err != nil {
		t.Fatalf("run taskgo %v: %v", args, err)
	}
	return got
}

// execute is the goroutine-safe form used by concurrency tests.
func execute(dir, dbPath string, args ...string) (result, error) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TASKGO_DB="+dbPath, "NO_COLOR=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exit := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exit = exitErr.ExitCode()
		} else {
			return result{}, err
		}
	}

	return result{stdout: stdout.String(), stderr: stderr.String(), exit: exit}, nil
}

func newDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "data", "taskgo.db")
}
