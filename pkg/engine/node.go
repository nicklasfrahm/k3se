package engine

import (
	"github.com/nicklasfrahm/k3se/pkg/sshx"
)

const (
	// Program is used to configure the name of the configuration file.
	Program = "k3se"
)

// Node describes the configuration of a node.
type Node struct {
	Role   Role        `yaml:"role"`
	SSH    sshx.Config `yaml:"ssh"`
	Config K3sConfig   `yaml:"config"`

	Client *sshx.Client `yaml:"-"`
}

// Connect establishes a connection to the node.
func (node *Node) Connect(options ...Option) error {
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return err
	}

	node.Client, err = sshx.NewClient(&node.SSH,
		sshx.WithProxy(opts.SSHProxy),
		sshx.WithLogger(opts.Logger),
		sshx.WithTimeout(opts.Timeout),
	)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect closes the connection to the node.
func (node *Node) Disconnect() error {
	if node.Client != nil {
		return node.Client.Close()
	}

	return nil
}
