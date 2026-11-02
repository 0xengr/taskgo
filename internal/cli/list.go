package cli

import (
	"fmt"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newListCmd(r *runner) *cobra.Command {
	var (
		stateRaw      string
		visibilityRaw string
		all           bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			state, visibility, err := parseSelection(cmd, stateRaw, visibilityRaw, all)
			if err != nil {
				return err
			}

			var tasks []domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				tasks, err = a.ListTasks(cmd.Context(), domain.TaskQuery{
					State:      state,
					Visibility: visibility,
				})
				return err
			}); err != nil {
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

	cmd.Flags().StringVar(&stateRaw, "state", string(domain.StateOpen), "open, done, or all")
	cmd.Flags().StringVar(&visibilityRaw, "visibility", string(domain.VisibilityActive), "active, archived, or all")
	cmd.Flags().BoolVar(&all, "all", false, "include done and archived tasks")
	return cmd
}

// parseSelection resolves the two independent list dimensions. --all is a
// shortcut for both, but never silently overrides an explicit dimension.
func parseSelection(cmd *cobra.Command, stateRaw, visibilityRaw string, all bool) (domain.State, domain.Visibility, error) {
	if all {
		if cmd.Flags().Changed("state") || cmd.Flags().Changed("visibility") {
			return "", "", &usageError{fmt.Errorf("--all cannot be combined with --state or --visibility")}
		}
		return domain.StateAll, domain.VisibilityAll, nil
	}

	state, err := domain.ParseState(stateRaw)
	if err != nil {
		return "", "", err
	}
	visibility, err := domain.ParseVisibility(visibilityRaw)
	if err != nil {
		return "", "", err
	}
	return state, visibility, nil
}

func (r *runner) printTasks(tasks []domain.Task) {
	for _, task := range tasks {
		suffix := ""
		if task.Project != nil {
			suffix += "  [" + task.Project.Name + "]"
		}
		if task.Archived() {
			suffix += "  (archived)"
		}
		fmt.Fprintf(r.out, "#%-4d %s %s%s\n", task.ID, r.priorityLabel(task.Priority), task.Title, suffix)
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
