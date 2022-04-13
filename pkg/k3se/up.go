package k3se

import (
	"github.com/nicklasfrahm/k3se/pkg/rexec"
)

func Up(options ...Option) error {
	// Fetch the options for this operation.
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return err
	}

	// Load the configuration file.
	config, err := LoadConfig(opts.ConfigPath)
	if err != nil {
		return err
	}

	// Verify the configuration file.
	if err := config.Verify(); err != nil {
		return err
	}

	// Create a runner for each node.
	// TODO: Switch type below to []*rexec.Runner once interface is fully implemented.
	runners := make([]*rexec.SSH, 0)
	for _, node := range config.Nodes {
		runner, err := rexec.NewSSH(&node.SSH, rexec.WithSSHProxy(&config.SSHProxy))
		if err != nil {
			return err
		}
		runners = append(runners, runner)
	}

	// Run pre-flight SSH connectivity check to prevent inconsistent deployment.
	for _, runner := range runners {
		if err := runner.Connect(); err != nil {
			return err
		}
		defer runner.Disconnect()
	}

	// TODO: Run pre-flight SSH connectivity check to prevent inconsistent deployment.

	// TODO: Create configuration file at /etc/rancher/k3s/config.yaml

	// TODO: Copy installation script to /tmp/k3s-install.sh

	// TODO: Run installation script via sudo.

	// TODO: Copy kubeconfig to /etc/rancher/k3s/k3s.yaml.

	// TODO: Store state on server nodes to allow for configuration diffing later on.

	return nil
}
