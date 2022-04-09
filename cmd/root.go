package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"
var help bool

var rootCmd = &cobra.Command{
	Use:   "k3se",
	Short: "A lightweight and declarative k3s engine",
	Long: `  _    _____
 | | _|___ / ___  ___
 | |/ / |_ \/ __|/ _ \
 |   < ___) \__ \  __/
 |_|\_\____/|___/\___|

A lightweight Kubernetes engine that deploys k3s
clusters declaratively based on a cluster config
file.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if help {
			cmd.Help()
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
	Version:      version,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&help, "help", "h", false, "display help for command")
}

// Execute starts the invocation of the command line interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
