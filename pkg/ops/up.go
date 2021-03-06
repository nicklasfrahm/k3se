package ops

import (
	"github.com/nicklasfrahm/k3se/pkg/engine"
)

func Up(options ...Option) error {
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

	if err := eng.Install(); err != nil {
		return err
	}

	// TODO: Store state on server nodes to allow for configuration diffing later on.
	// TODO: Fetch state from Git history.

	if err := eng.Disconnect(); err != nil {
		return err
	}

	return nil
}
