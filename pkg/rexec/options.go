package rexec

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options contains the configuration for an operation.
type Options struct {
	Logger   *zerolog.Logger
	SSHProxy *SSHConfig
	Timeout  time.Duration
}

// Option applies a configuration option
// for the execution of an operation.
type Option func(options *Options) error

// Apply applies the option functions to the current set of options.
func (o *Options) Apply(options ...Option) (*Options, error) {
	for _, option := range options {
		if err := option(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// GetDefaultOptions returns the default options
// for all operations of this library.
func GetDefaultOptions() *Options {
	logger := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	return &Options{
		SSHProxy: nil,
		Timeout:  time.Second * 5,
		Logger:   &logger,
	}
}

// WithSSHProxy configures an SSH bastion host.
func WithSSHProxy(sshProxy *SSHConfig) Option {
	return func(options *Options) error {
		options.SSHProxy = sshProxy
		return nil
	}
}
