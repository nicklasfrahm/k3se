package engine

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
)

// Options contains the configuration for an operation.
type Options struct {
	Logger   *zerolog.Logger
	SSHProxy *sshx.Client
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

// WithLogger allows to use a custom logger.
func WithLogger(logger *zerolog.Logger) Option {
	return func(options *Options) error {
		options.Logger = logger
		return nil
	}
}

// WithSSHProxy allows to use an existing SSH
// connection as an SSH bastion host.
func WithSSHProxy(proxy *sshx.Client) Option {
	return func(options *Options) error {
		options.SSHProxy = proxy
		return nil
	}
}

// WithTimeout allows to set a custom timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(options *Options) error {
		options.Timeout = timeout
		return nil
	}
}
