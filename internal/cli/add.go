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
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Capture a new task",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			priority, err := domain.ParsePriority(priorityRaw)
			if err != nil {
				return err
			}

			task, err := r.addTask(cmd.Context(), app.AddTaskInput{
				Title:       strings.Join(args, " "),
				Description: description,
				Priority:    priority,
			})
			if err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Added #%d to Inbox\n", task.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "task description")
	cmd.Flags().StringVar(&priorityRaw, "priority", string(domain.PriorityMedium), "priority: low, medium, or high")
	return cmd
}

func (r *runner) addTask(ctx context.Context, in app.AddTaskInput) (domain.Task, error) {
	a, closeApp, err := r.openApp(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer closeApp()
	return a.AddTask(ctx, in)
}

// taskView is the JSON shape of a Task. It is intentionally unstable until
// v1.0 (see the spec's output contract).
type taskView struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Priority    string  `json:"priority"`
	Project     any     `json:"project"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at"`
	ArchivedAt  *string `json:"archived_at"`
}

func newTaskView(task domain.Task) taskView {
	return taskView{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    string(task.Priority),
		Project:     nil,
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
