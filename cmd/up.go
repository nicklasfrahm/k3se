package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nicklasfrahm/k3se/pkg/ops"
)

var upCmd = &cobra.Command{
	Use:   "up [config]",
	Short: "Deploy or upgrade cluster",
	Long: `Deploy a new cluster or upgrade an existing one.

By default the command expects a "k3se.yaml" config
file in the current directory. You may override this
by passing a path to the configuration file as a CLI
argument.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use manual override for config path if provided.
		if len(args) == 1 {
			return ops.Up(ops.WithConfigPath(args[0]))
		}

		return ops.Up()
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
