package engine

import (
	"io"
	"path/filepath"

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

// Upload writes the specified content to the remote file on the node.
func (node *Node) Upload(dst string, src io.Reader) error {
	// Get base directory for the file.
	dir := filepath.Dir(dst)

	// Create directory if it does not exist.
	if err := node.Client.SFTP.MkdirAll(dir); err != nil {
		return err
	}

	// Upload file.
	file, err := node.Client.SFTP.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	// Empty existing file.
	if err := file.Truncate(0); err != nil {
		return err
	}

	// Overwrite file content.
	_, err = io.Copy(file, src)
	return err
}

// Do executes a command on the node.
func (node *Node) Do(cmd sshx.Cmd) error {
	return node.Client.Do(cmd)
}
