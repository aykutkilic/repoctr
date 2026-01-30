package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"repoctr/internal/version"
)

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current version of repo-ctr",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("repo-ctr %s\n", version.Version)
		},
	}
}
