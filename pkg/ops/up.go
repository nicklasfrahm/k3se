package ops

import (
	"github.com/nicklasfrahm/k3se/pkg/engine"
	"github.com/nicklasfrahm/k3se/pkg/sshx"
)

func Up(options ...Option) error {
	// Fetch the options for this operation.
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return err
	}

	// Load the configuration file.
	config, err := engine.LoadConfig(opts.ConfigPath)
	if err != nil {
		return err
	}

	eng, err := engine.New(engine.WithLogger(opts.Logger))
	if err != nil {
		return err
	}

	if err := eng.SetSpec(config); err != nil {
		return err
	}

	// Establish connection to proxy if host is specified.
	var sshProxy *sshx.Client
	if config.SSHProxy.Host != "" {
		if sshProxy, err = sshx.NewClient(&config.SSHProxy); err != nil {
			return err
		}
	}

	// Get a list of all nodes and connect to them.
	nodes := config.NodesByRole(engine.RoleAny)
	for _, node := range nodes {
		if err := node.Connect(engine.WithSSHProxy(sshProxy)); err != nil {
			return err
		}

		if err := eng.Configure(node); err != nil {
			return err
		}

		if err := eng.Install(node); err != nil {
			return err
		}
	}

	// TODO: Copy kubeconfig to /etc/rancher/k3s/k3s.yaml.

	// TODO: Store state on server nodes to allow for configuration diffing later on.
	// TODO: Fetch state from Git history.

	// Clean up and disconnect from all nodes.
	for _, node := range nodes {
		if err := eng.Cleanup(node); err != nil {
			return err
		}

		if err := node.Disconnect(); err != nil {
			return err
		}
	}

	return nil
}
