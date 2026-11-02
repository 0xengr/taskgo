// Command taskgo-bench measures process-level latency of taskgo against a
// seeded database. It is intentionally not a Go test: timing belongs on a
// quiet machine, not in CI.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
	"github.com/0xengr/taskgo/internal/infra/sqlite"
)

func main() {
	binary := flag.String("binary", "./bin/taskgo", "path to the taskgo binary")
	tasks := flag.Int("tasks", 10000, "number of tasks to seed")
	runs := flag.Int("runs", 30, "measured runs per command")
	maxP95 := flag.Duration("max-p95", 100*time.Millisecond, "fail when any command exceeds this p95")
	dbPath := flag.String("db", "", "database path (default: a temporary file)")
	flag.Parse()

	if err := run(*binary, *tasks, *runs, *maxP95, *dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "taskgo-bench: %v\n", err)
		os.Exit(1)
	}
}

type scenario struct {
	name string
	args []string
}

func run(binary string, taskCount, runs int, maxP95 time.Duration, dbPath string) error {
	absolute, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	if info, err := os.Stat(absolute); err != nil || info.IsDir() {
		return fmt.Errorf("binary not found at %s (run `make build` first)", absolute)
	}
	if runs < 1 {
		return fmt.Errorf("runs must be at least 1")
	}

	cleanup := func() {}
	if dbPath == "" {
		dir, err := os.MkdirTemp("", "taskgo-bench-")
		if err != nil {
			return fmt.Errorf("create temp directory: %w", err)
		}
		dbPath = filepath.Join(dir, "taskgo.db")
		cleanup = func() { os.RemoveAll(dir) }
	}
	defer cleanup()

	fmt.Printf("Preparing %s\n", dbPath)
	seeded, err := seed(context.Background(), dbPath, taskCount)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	if seeded {
		fmt.Printf("Seeded %d tasks\n", taskCount)
	} else {
		fmt.Printf("Using the existing seeded database\n")
	}

	scenarios := []scenario{
		{name: "list", args: []string{"list"}},
		{name: "show", args: []string{"show", "1"}},
		{name: "next", args: []string{"next"}},
		{name: "project list", args: []string{"project", "list"}},
	}

	fmt.Printf("\nEnvironment\n")
	fmt.Printf("  OS/arch:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  CPUs:     %d\n", runtime.NumCPU())
	fmt.Printf("  Go:       %s\n", runtime.Version())
	fmt.Printf("  Tasks:    %d\n", taskCount)
	fmt.Printf("  Runs:     %d per command\n", runs)

	fmt.Printf("\nResults (p95, includes process startup and database open)\n")

	failed := false
	for _, item := range scenarios {
		if _, err := measure(absolute, dbPath, item.args); err != nil {
			return fmt.Errorf("%s warmup: %w", item.name, err)
		}

		samples := make([]time.Duration, 0, runs)
		for i := 0; i < runs; i++ {
			elapsed, err := measure(absolute, dbPath, item.args)
			if err != nil {
				return fmt.Errorf("%s run %d: %w", item.name, i+1, err)
			}
			samples = append(samples, elapsed)
		}

		p95 := percentile(samples, 0.95)
		status := "ok"
		if p95 > maxP95 {
			status = "SLOW"
			failed = true
		}
		fmt.Printf("  %-14s %8.1f ms  %s\n", item.name, float64(p95)/float64(time.Millisecond), status)
	}

	fmt.Printf("\nTarget: p95 <= %s\n", maxP95)
	if failed {
		return fmt.Errorf("at least one command exceeded the p95 target")
	}
	return nil
}

func measure(binary, dbPath string, args []string) (time.Duration, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "TASKGO_DB="+dbPath, "NO_COLOR=1")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func percentile(samples []time.Duration, fraction float64) time.Duration {
	ordered := append([]time.Duration(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })

	index := int(math.Ceil(fraction*float64(len(ordered)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(ordered) {
		index = len(ordered) - 1
	}
	return ordered[index]
}

func seed(ctx context.Context, dbPath string, count int) (bool, error) {
	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		return false, err
	}
	defer db.Close()

	existing, err := db.ListProjects(ctx, domain.VisibilityAll)
	if err != nil {
		return false, err
	}
	if len(existing) > 0 {
		return false, nil
	}

	now := time.Now().UTC()
	projects := make([]int64, 0, 5)
	for i := 0; i < 5; i++ {
		name, key, err := domain.NormalizeProjectName(fmt.Sprintf("Bench %d", i+1))
		if err != nil {
			return false, err
		}
		project, err := db.InsertProject(ctx, domain.NewProject{Name: name, NameKey: key}, now)
		if err != nil {
			return false, err
		}
		projects = append(projects, project.ID)
	}

	priorities := []domain.Priority{
		domain.PriorityLow,
		domain.PriorityMedium,
		domain.PriorityHigh,
	}
	for i := 0; i < count; i++ {
		var projectID *int64
		if i%5 != 0 {
			id := projects[i%len(projects)]
			projectID = &id
		}
		if _, err := db.InsertTask(ctx, domain.NewTask{
			Title:     fmt.Sprintf("benchmark task %d", i),
			Priority:  priorities[i%len(priorities)],
			ProjectID: projectID,
		}, now.Add(-time.Duration(i)*time.Minute)); err != nil {
			return false, err
		}
	}
	return true, nil
}
