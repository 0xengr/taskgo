package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/0xengr/taskgo/internal/infra/sqlite"
	"github.com/spf13/cobra"
)

const (
	exitOK     = 0
	exitDomain = 1
	exitUsage  = 2
)

// usageError marks problems with how the command was invoked: unknown flags,
// missing arguments, or values that fail validation.
type usageError struct {
	err error
}

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

type runner struct {
	out     io.Writer
	errOut  io.Writer
	json    bool
	noColor bool
	openApp func(context.Context) (*app.App, func() error, error)
}

// Execute runs the command surface and returns the process exit code.
func Execute(args []string, stdout, stderr io.Writer) int {
	r := &runner{
		out:     stdout,
		errOut:  stderr,
		openApp: openDefaultApp,
	}

	root := newRoot(r)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(stderr, "taskgo: %v\n", err)
		return codeFor(err)
	}
	return exitOK
}

func codeFor(err error) int {
	var usage *usageError
	if errors.As(err, &usage) {
		return exitUsage
	}
	var validation *domain.ValidationError
	if errors.As(err, &validation) {
		return exitUsage
	}
	return exitDomain
}

func openDefaultApp(ctx context.Context) (*app.App, func() error, error) {
	path, err := sqlite.DefaultPath()
	if err != nil {
		return nil, nil, err
	}
	db, err := sqlite.Open(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	return app.New(db), db.Close, nil
}

func newRoot(r *runner) *cobra.Command {
	root := &cobra.Command{
		Use:           "taskgo",
		Short:         "Local-first task and project tracker for developers",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          noUnknownCommand,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &usageError{fmt.Errorf("unknown command %q", args[0])}
			}
			return cmd.Help()
		},
	}

	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{err: err}
	})

	root.PersistentFlags().BoolVar(&r.json, "json", false, "emit JSON")
	root.PersistentFlags().BoolVar(&r.noColor, "no-color", false, "disable color output")

	root.AddCommand(
		newAddCmd(r),
		newListCmd(r),
		newShowCmd(r),
		newDoneCmd(r),
		newReopenCmd(r),
		newArchiveCmd(r),
		newUnarchiveCmd(r),
		newEditCmd(r),
		newProjectCmd(r),
	)
	return root
}

func atLeastOneArg(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return &usageError{errors.New("a title is required")}
	}
	return nil
}

func noUnknownCommand(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return &usageError{fmt.Errorf("unknown command %q", args[0])}
	}
	return nil
}

func exactlyTwoArgs(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return &usageError{fmt.Errorf("expected 2 arguments, got %d", len(args))}
	}
	return nil
}

// withApp opens the store for one command and guarantees it is closed.
func (r *runner) withApp(ctx context.Context, fn func(*app.App) error) error {
	a, closeApp, err := r.openApp(ctx)
	if err != nil {
		return err
	}
	defer closeApp()
	return fn(a)
}

func noArgs(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return &usageError{fmt.Errorf("unexpected arguments: %v", args)}
	}
	return nil
}

// writeJSON emits one compact JSON value, deliberately without HTML escaping
// so titles survive intact.
func (r *runner) writeJSON(value any) error {
	encoder := json.NewEncoder(r.out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

// colorEnabled reports whether human output may use ANSI styling.
func (r *runner) colorEnabled() bool {
	if r.noColor || r.json || os.Getenv("NO_COLOR") != "" {
		return false
	}
	file, ok := r.out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
