package cli

import (
	"fmt"

	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newListCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List open tasks",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, closeApp, err := r.openApp(cmd.Context())
			if err != nil {
				return err
			}
			defer closeApp()

			tasks, err := a.ListTasks(cmd.Context())
			if err != nil {
				return err
			}

			if r.json {
				views := make([]taskView, 0, len(tasks))
				for _, task := range tasks {
					views = append(views, newTaskView(task))
				}
				return r.writeJSON(views)
			}

			r.printTasks(tasks)
			return nil
		},
	}
}

func (r *runner) printTasks(tasks []domain.Task) {
	for _, task := range tasks {
		fmt.Fprintf(r.out, "#%-4d %s %s\n", task.ID, r.priorityLabel(task.Priority), task.Title)
	}
}

// priorityLabel pads to a stable visible width before adding color, so
// columns line up whether or not ANSI codes are present.
func (r *runner) priorityLabel(priority domain.Priority) string {
	label := fmt.Sprintf("%-6s", priority)
	if !r.colorEnabled() {
		return label
	}
	switch priority {
	case domain.PriorityHigh:
		return "\x1b[31m" + label + "\x1b[0m"
	case domain.PriorityMedium:
		return "\x1b[33m" + label + "\x1b[0m"
	case domain.PriorityLow:
		return "\x1b[32m" + label + "\x1b[0m"
	default:
		return label
	}
}
