package ops

import "github.com/rs/zerolog"

const (
	// Program is used to configure the name of the configuration file.
	Program = "k3se"
	// DefaultKubeConfigPath is the default path to the kubeconfig file.
	DefaultKubeConfigPath = "~/.kube/config"
)

// Options contains the configuration for an operation.
type Options struct {
	ConfigPath     string
	KubeConfigPath string
	Logger         *zerolog.Logger
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

	return &Options{
		ConfigPath:     Program + ".yml",
		KubeConfigPath: DefaultKubeConfigPath,
	}
}

// WithConfigPath overrides the default configuration path.
func WithConfigPath(configPath string) Option {
	return func(options *Options) error {
		options.ConfigPath = configPath
		return nil
	}
}

// WithLogger overrides the default logger.
func WithLogger(logger *zerolog.Logger) Option {
	return func(options *Options) error {
		options.Logger = logger
		return nil
	}
}

// WithKubeConfigPath overrides the default kubeconfig path.
func WithKubeConfigPath(kubeConfigPath string) Option {
	return func(options *Options) error {
		options.KubeConfigPath = kubeConfigPath
		return nil
	}
}
