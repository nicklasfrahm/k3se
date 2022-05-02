package ops

import (
	"github.com/nicklasfrahm/k3se/pkg/engine"
)

// TODO: Reduce amount of network traffic. The current setup is
//       not optimal as it will connect to all nodes despite only
//       needing to download the config from a single node.
func KubeConfig(options ...Option) error {
	// Fetch the options for this operation.
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return err
	}

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

	if err := eng.Connect(); err != nil {
		return err
	}

	if err := eng.KubeConfig(opts.KubeConfigPath); err != nil {
		return err
	}

	if err := eng.Disconnect(); err != nil {
		return err
	}

	return nil
}
