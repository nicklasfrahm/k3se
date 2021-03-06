package engine

import (
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
	"github.com/rs/zerolog"
)

var loglevel = regexp.MustCompile(`\[([^\]]*)\]\s*`)

const (
	// Program is used to configure the name of the configuration file.
	Program = "k3se"
)

// Node describes the configuration of a node.
type Node struct {
	Role   Role        `yaml:"role"`
	SSH    sshx.Config `yaml:"ssh"`
	Server Server      `yaml:"server,omitempty"`
	Agent  Agent       `yaml:"agent,omitempty"`

	Client *sshx.Client   `yaml:"-"`
	Logger zerolog.Logger `yaml:"-"`
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

	// Restrict permissions.
	if err := node.Client.SFTP.Chmod(dst, 0644); err != nil {
		return err
	}

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

// Write writes a log information for the node.
// TODO: Make this more efficient by reducing allocations.
func (node *Node) Write(raw []byte) (int, error) {
	// Remove the log level supplied by the k3s install script.
	trimmed := loglevel.ReplaceAll(raw, []byte(""))

	lines := strings.Split(string(trimmed), "\n")
	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			node.Logger.Debug().Msg(line)
		}
	}

	return len(raw), nil
}
