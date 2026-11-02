package cli

import (
	"context"
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
