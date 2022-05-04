package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/nicklasfrahm/k3se/pkg/ops"
)

var kubeConfigPath string
var skipInstall bool

var upCmd = &cobra.Command{
	Use:   "up [config]",
	Short: "Deploy or upgrade cluster",
	Long: `Deploy a new cluster or upgrade an existing one.

By default the command expects a "k3se.yml" config
file in the current directory. You may override this
by passing a path to the configuration file as a CLI
argument.

This command will also download the kubeconfig and
merge the new context to the kubeconfig located at
"~/.kube/config". Alternatively, you may use the
--kubeconfig flag to specify a custom location for
the new context to be written to.`,
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

		// Use manual override for kubeconfig path if provided.
		if kubeConfigPath != "" {
			opts = append(opts, ops.WithKubeConfigPath(kubeConfigPath))
		}

		if !skipInstall {
			if err := ops.Up(opts...); err != nil {
				return err
			}
		}

		return ops.KubeConfig(opts...)
	},
}

func init() {
	upCmd.Flags().StringVarP(&kubeConfigPath, "kubeconfig", "k", "~/.kube/config", "location to write the kubeconfig")
	upCmd.Flags().BoolVarP(&skipInstall, "skip-install", "s", false, "only download the kubeconfig")

	rootCmd.AddCommand(upCmd)
}
