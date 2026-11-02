package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newAddCmd(r *runner) *cobra.Command {
	var (
		description string
		priorityRaw string
		projectName string
		useInbox    bool
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Capture a new task",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectSet := cmd.Flags().Changed("project")
			if useInbox && projectSet {
				return &usageError{fmt.Errorf("--project and --inbox cannot be used together")}
			}

			target := app.TargetAuto
			switch {
			case useInbox:
				target = app.TargetInbox
			case projectSet:
				target = app.TargetProject
			}

			priority, err := domain.ParsePriority(priorityRaw)
			if err != nil {
				return err
			}

			task, err := r.addTask(cmd.Context(), app.AddTaskInput{
				Title:       strings.Join(args, " "),
				Description: description,
				Priority:    priority,
				Target:      target,
				ProjectName: projectName,
			})
			if err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			destination := "Inbox"
			if task.Project != nil {
				destination = task.Project.Name
			}
			fmt.Fprintf(r.out, "Added #%d to %s\n", task.ID, destination)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "task description")
	cmd.Flags().StringVar(&priorityRaw, "priority", string(domain.PriorityMedium), "priority: low, medium, or high")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "assign to a project by name")
	cmd.Flags().BoolVar(&useInbox, "inbox", false, "capture to the inbox even inside a project repository")
	return cmd
}

func (r *runner) addTask(ctx context.Context, in app.AddTaskInput) (domain.Task, error) {
	var task domain.Task
	err := r.withApp(ctx, func(a *app.App) error {
		var err error
		task, err = a.AddTask(ctx, in)
		return err
	})
	return task, err
}

type projectRefView struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// taskView is the JSON shape of a Task. It is intentionally unstable until
// v1.0 (see the spec's output contract).
type taskView struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Priority    string          `json:"priority"`
	Project     *projectRefView `json:"project"`
	CreatedAt   string          `json:"created_at"`
	CompletedAt *string         `json:"completed_at"`
	ArchivedAt  *string         `json:"archived_at"`
}

func newTaskView(task domain.Task) taskView {
	var project *projectRefView
	if task.Project != nil {
		project = &projectRefView{ID: task.Project.ID, Name: task.Project.Name}
	}

	return taskView{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    string(task.Priority),
		Project:     project,
		CreatedAt:   jsonTime(task.CreatedAt),
		CompletedAt: nullableJSONTime(task.CompletedAt),
		ArchivedAt:  nullableJSONTime(task.ArchivedAt),
	}
}

func jsonTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func nullableJSONTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := jsonTime(*t)
	return &formatted
}
