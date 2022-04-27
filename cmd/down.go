package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/nicklasfrahm/k3se/pkg/ops"
)

var downCmd = &cobra.Command{
	Use:   "down [config]",
	Short: "Destroy a cluster",
	Long: `Destroy an existing cluster by removing k3s entirely
from the node. Use with caution as this will destroy
all data stored in your cluster and cannot be undone.

By default the command expects a "k3se.yaml" config
file in the current directory. You may override this
by passing a path to the configuration file as a CLI
argument.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})

		opts := []ops.Option{
			ops.WithLogger(&logger),
		}

		// Use manual override for config path if provided.
		if len(args) == 1 {
			opts = append(opts, ops.WithConfigPath(args[0]))
		}

		return ops.Down(opts...)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
