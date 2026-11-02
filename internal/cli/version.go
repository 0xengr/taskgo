package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set by the linker at build time.
var (
	version = "dev"
	commit  = "none"
)

func newVersionCmd(r *runner) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version and commit",
		Args:  noArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if r.json {
				return r.writeJSON(struct {
					Version string `json:"version"`
					Commit  string `json:"commit"`
				}{Version: version, Commit: commit})
			}
			fmt.Fprintf(r.out, "taskgo %s (%s)\n", version, commit)
			return nil
		},
	}
}
