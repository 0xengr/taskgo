package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newShowCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show one task in full",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.GetTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			r.printTask(task)
			return nil
		},
	}
}

func newDoneCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a task done",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.CompleteTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Completed #%d\n", task.ID)
			return nil
		},
	}
}

func newReopenCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen a done task",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.ReopenTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Reopened #%d\n", task.ID)
			return nil
		},
	}
}

func newArchiveCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Hide a task without changing its completion",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.ArchiveTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Archived #%d\n", task.ID)
			return nil
		},
	}
}

func newUnarchiveCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <id>",
		Short: "Restore an archived task",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.UnarchiveTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Unarchived #%d\n", task.ID)
			return nil
		},
	}
}

func newEditCmd(r *runner) *cobra.Command {
	var (
		title       string
		description string
		priorityRaw string
		projectName string
		inbox       bool
	)

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a task",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			titleSet := cmd.Flags().Changed("title")
			descriptionSet := cmd.Flags().Changed("description")
			prioritySet := cmd.Flags().Changed("priority")
			projectSet := cmd.Flags().Changed("project")

			if projectSet && inbox {
				return &usageError{fmt.Errorf("--project and --inbox cannot be used together")}
			}
			if !titleSet && !descriptionSet && !prioritySet && !projectSet && !inbox {
				return &usageError{fmt.Errorf("nothing to change")}
			}

			var priority domain.Priority
			if prioritySet {
				if strings.TrimSpace(priorityRaw) == "" {
					return &usageError{fmt.Errorf("priority cannot be empty")}
				}
				priority, err = domain.ParsePriority(priorityRaw)
				if err != nil {
					return err
				}
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.EditTask(cmd.Context(), app.EditTaskInput{
					ID:             id,
					Title:          title,
					SetTitle:       titleSet,
					Description:    description,
					SetDescription: descriptionSet,
					Priority:       priority,
					SetPriority:    prioritySet,
					ProjectName:    projectName,
					SetProject:     projectSet,
					Inbox:          inbox,
				})
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Updated #%d\n", task.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "replace the title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "replace the description; an empty value clears it")
	cmd.Flags().StringVar(&priorityRaw, "priority", "", "replace the priority: low, medium, or high")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "move to a project by name")
	cmd.Flags().BoolVar(&inbox, "inbox", false, "move to the inbox")
	return cmd
}

func newDeleteCmd(r *runner) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a task",
		Args:  exactlyOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			if !force {
				if !r.stdinIsTerminal() {
					return errors.New("refusing to delete without confirmation; pass --force in non-interactive sessions")
				}

				var existing domain.Task
				if err := r.withApp(cmd.Context(), func(a *app.App) error {
					var err error
					existing, err = a.GetTask(cmd.Context(), id)
					return err
				}); err != nil {
					return err
				}

				ok, err := confirm(r.in, r.errOut, fmt.Sprintf("Delete #%d %q? [y/N] ", existing.ID, existing.Title))
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("deletion cancelled")
				}
			}

			var task domain.Task
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				task, err = a.DeleteTask(cmd.Context(), id)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newTaskView(task))
			}
			fmt.Fprintf(r.out, "Deleted #%d\n", task.ID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "delete without confirmation")
	return cmd
}

func exactlyOneArg(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return &usageError{fmt.Errorf("expected 1 argument, got %d", len(args))}
	}
	return nil
}

// parseTaskID accepts both "42" and "#42".
func parseTaskID(raw string) (int64, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || id <= 0 {
		return 0, &usageError{errors.New("task id must be a positive number (got " + strconv.Quote(raw) + ")")}
	}
	return id, nil
}

func (r *runner) printTask(task domain.Task) {
	fmt.Fprintf(r.out, "#%d  %s%s\n", task.ID, r.priorityLabel(task.Priority), task.Title)

	project := "Inbox"
	if task.Project != nil {
		project = task.Project.Name
	}
	fmt.Fprintf(r.out, "\nProject:    %s\n", project)
	fmt.Fprintf(r.out, "Status:     %s\n", openStatus(task.Open()))
	fmt.Fprintf(r.out, "Archived:   %s\n", yesNo(task.Archived()))
	fmt.Fprintf(r.out, "Created:    %s\n", localTime(task.CreatedAt))
	if task.CompletedAt != nil {
		fmt.Fprintf(r.out, "Completed:  %s\n", localTime(*task.CompletedAt))
	}
	if task.ArchivedAt != nil {
		fmt.Fprintf(r.out, "Archived at: %s\n", localTime(*task.ArchivedAt))
	}

	if task.Description != nil && *task.Description != "" {
		fmt.Fprintf(r.out, "\n%s\n", *task.Description)
	}
}

func openStatus(open bool) string {
	if open {
		return "open"
	}
	return "done"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func localTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
}
