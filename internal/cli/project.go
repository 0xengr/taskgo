package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/0xengr/taskgo/internal/app"
	"github.com/0xengr/taskgo/internal/domain"
	"github.com/spf13/cobra"
)

func newProjectCmd(r *runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create and manage projects",
		Args:  noUnknownCommand,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newProjectAddCmd(r),
		newProjectListCmd(r),
		newProjectRenameCmd(r),
		newProjectEditCmd(r),
		newProjectArchiveCmd(r),
		newProjectUnarchiveCmd(r),
		newProjectDeleteCmd(r),
	)
	return cmd
}

func newProjectAddCmd(r *runner) *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a project, optionally bound to a repository",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := r.addProject(cmd.Context(), app.AddProjectInput{
				Name:     strings.Join(args, " "),
				RepoPath: repo,
				HasRepo:  cmd.Flags().Changed("repo"),
			})
			if err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newProjectView(project))
			}
			if project.RepoPath != nil {
				fmt.Fprintf(r.out, "Created project %q (repository: %s)\n", project.Name, *project.RepoPath)
			} else {
				fmt.Fprintf(r.out, "Created project %q\n", project.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "bind the project to a local repository path")
	return cmd
}

func newProjectListCmd(r *runner) *cobra.Command {
	var visibilityRaw string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			visibility, err := domain.ParseVisibility(visibilityRaw)
			if err != nil {
				return err
			}

			var projects []domain.Project
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				projects, err = a.ListProjects(cmd.Context(), visibility)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				views := make([]projectView, 0, len(projects))
				for _, project := range projects {
					views = append(views, newProjectView(project))
				}
				return r.writeJSON(views)
			}

			writer := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
			for _, project := range projects {
				repo := "-"
				if project.RepoPath != nil {
					repo = *project.RepoPath
				}
				marker := ""
				if project.Archived() {
					marker = "  (archived)"
				}
				fmt.Fprintf(writer, "%s\t%s%s\n", project.Name, repo, marker)
			}
			return writer.Flush()
		},
	}

	cmd.Flags().StringVar(&visibilityRaw, "visibility", string(domain.VisibilityActive), "active, archived, or all")
	return cmd
}

func newProjectRenameCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <current> <new>",
		Short: "Rename a project",
		Args:  exactlyTwoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var project domain.Project
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				project, err = a.RenameProject(cmd.Context(), args[0], args[1])
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newProjectView(project))
			}
			fmt.Fprintf(r.out, "Renamed %q to %q\n", args[0], project.Name)
			return nil
		},
	}
}

func newProjectEditCmd(r *runner) *cobra.Command {
	var (
		repo      string
		clearRepo bool
	)

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Change a project's repository binding",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoSet := cmd.Flags().Changed("repo")
			if repoSet && clearRepo {
				return &usageError{fmt.Errorf("--repo and --clear-repo cannot be used together")}
			}
			if !repoSet && !clearRepo {
				return &usageError{fmt.Errorf("nothing to change")}
			}

			var project domain.Project
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				project, err = a.EditProject(cmd.Context(), strings.Join(args, " "), repo, repoSet, clearRepo)
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newProjectView(project))
			}
			if project.RepoPath != nil {
				fmt.Fprintf(r.out, "Updated project %q (repository: %s)\n", project.Name, *project.RepoPath)
			} else {
				fmt.Fprintf(r.out, "Updated project %q (repository cleared)\n", project.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "bind the project to a local repository path")
	cmd.Flags().BoolVar(&clearRepo, "clear-repo", false, "remove the repository binding")
	return cmd
}

func newProjectArchiveCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <name>",
		Short: "Archive a project and hide its tasks",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			var project domain.Project
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				project, err = a.ArchiveProject(cmd.Context(), strings.Join(args, " "))
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newProjectView(project))
			}
			fmt.Fprintf(r.out, "Archived project %q\n", project.Name)
			return nil
		},
	}
}

func newProjectUnarchiveCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <name>",
		Short: "Restore an archived project",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			var project domain.Project
			if err := r.withApp(cmd.Context(), func(a *app.App) error {
				var err error
				project, err = a.UnarchiveProject(cmd.Context(), strings.Join(args, " "))
				return err
			}); err != nil {
				return err
			}

			if r.json {
				return r.writeJSON(newProjectView(project))
			}
			fmt.Fprintf(r.out, "Unarchived project %q\n", project.Name)
			return nil
		},
	}
}

func newProjectDeleteCmd(r *runner) *cobra.Command {
	var (
		toInbox bool
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a project, choosing what happens to its tasks",
		Args:  atLeastOneArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			if toInbox && force {
				return &usageError{fmt.Errorf("--to-inbox and --force cannot be used together")}
			}

			name := strings.Join(args, " ")
			project, count, err := r.projectTaskCount(cmd.Context(), name)
			if err != nil {
				return err
			}

			strategy := app.DeleteProjectWithTasks
			if toInbox {
				strategy = app.DeleteProjectToInbox
			} else {
				if count > 0 && !force {
					return fmt.Errorf(
						"project %q has %s; use --to-inbox to preserve them or --force to delete them",
						project.Name, taskCountLabel(count),
					)
				}
				if !r.stdinIsTerminal() {
					return errors.New("refusing to delete a project without an interactive confirmation; run this command in a terminal")
				}

				prompt := fmt.Sprintf("Delete project %q? [y/N] ", project.Name)
				if count > 0 {
					prompt = fmt.Sprintf("Delete project %q and %s? [y/N] ", project.Name, taskCountLabel(count))
				}
				ok, err := confirm(r.in, r.errOut, prompt)
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("deletion cancelled")
				}
			}

			result, err := r.deleteProject(cmd.Context(), name, strategy)
			if err != nil {
				return err
			}

			if r.json {
				view := projectDeletionView{Project: newProjectView(result.Project)}
				if toInbox {
					moved := result.MovedTasks
					view.MovedTasks = &moved
				}
				return r.writeJSON(view)
			}

			if toInbox && result.MovedTasks > 0 {
				fmt.Fprintf(r.out, "Deleted project %q (moved %s to Inbox)\n", result.Project.Name, taskCountLabel(result.MovedTasks))
			} else {
				fmt.Fprintf(r.out, "Deleted project %q\n", result.Project.Name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&toInbox, "to-inbox", false, "move the project's tasks to the inbox")
	cmd.Flags().BoolVar(&force, "force", false, "delete the project and its tasks (still requires confirmation)")
	return cmd
}

func taskCountLabel(count int64) string {
	if count == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", count)
}

type projectDeletionView struct {
	Project    projectView `json:"project"`
	MovedTasks *int64      `json:"moved_tasks,omitempty"`
}

func (r *runner) projectTaskCount(ctx context.Context, name string) (domain.Project, int64, error) {
	var (
		project domain.Project
		count   int64
	)
	err := r.withApp(ctx, func(a *app.App) error {
		var err error
		project, count, err = a.ProjectTaskCount(ctx, name)
		return err
	})
	return project, count, err
}

func (r *runner) deleteProject(ctx context.Context, name string, strategy app.ProjectDeletionStrategy) (app.DeletedProject, error) {
	var result app.DeletedProject
	err := r.withApp(ctx, func(a *app.App) error {
		var err error
		result, err = a.DeleteProject(ctx, name, strategy)
		return err
	})
	return result, err
}

type projectView struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	RepositoryPath *string `json:"repository_path"`
	CreatedAt      string  `json:"created_at"`
	ArchivedAt     *string `json:"archived_at"`
}

func newProjectView(project domain.Project) projectView {
	return projectView{
		ID:             project.ID,
		Name:           project.Name,
		RepositoryPath: project.RepoPath,
		CreatedAt:      jsonTime(project.CreatedAt),
		ArchivedAt:     nullableJSONTime(project.ArchivedAt),
	}
}

func (r *runner) addProject(ctx context.Context, in app.AddProjectInput) (domain.Project, error) {
	var project domain.Project
	err := r.withApp(ctx, func(a *app.App) error {
		var err error
		project, err = a.AddProject(ctx, in)
		return err
	})
	return project, err
}
