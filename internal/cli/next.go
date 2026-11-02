package cli

import (
	"errors"
	"fmt"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newNextCmd(r *runner) *cobra.Command {
	var (
		count       int
		projectName string
		here        bool
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "Recommend the next task",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if count < 1 {
				return &usageError{errors.New("--count must be at least 1")}
			}
			if here && cmd.Flags().Changed("project") {
				return &usageError{errors.New("--project and --here cannot be used together")}
			}

			scope := app.ScopeAll
			switch {
			case here:
				scope = app.ScopeHere
			case cmd.Flags().Changed("project"):
				scope = app.ScopeProject
			}

			var tasks []domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				tasks, err = a.NextTasks(cmd.Context(), app.NextInput{
					Scope:       scope,
					ProjectName: projectName,
					Count:       count,
				})
				return err
			}); err != nil {
				return err
			}

			if r.json {
				if count == 1 {
					if len(tasks) == 0 {
						return r.writeJSON(nil)
					}
					return r.writeJSON(newTaskView(tasks[0]))
				}
				views := make([]taskView, 0, len(tasks))
				for _, task := range tasks {
					views = append(views, newTaskView(task))
				}
				return r.writeJSON(views)
			}

			if len(tasks) == 0 {
				fmt.Fprintln(r.out, "Nothing open.")
				return nil
			}
			r.printTasks(tasks)
			return nil
		},
	}

	cmd.Flags().IntVar(&count, "count", 1, "number of tasks to return")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "scope to a project by name")
	cmd.Flags().BoolVar(&here, "here", false, "scope to the detected project")
	return cmd
}
