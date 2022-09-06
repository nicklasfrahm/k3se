package engine

import (
	"errors"
	"os"
	"strings"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
	"gopkg.in/yaml.v3"
)

// Role is the type of a node in the cluster.
type Role string

const (
	// RoleAny is the role selector that matches any node.
	RoleAny Role = "any"
	// RoleServer is the role of a control-plane node in k3s.
	RoleServer Role = "server"
	// RoleAgent is the role of a worker node in k3s.
	RoleAgent Role = "agent"
)

var (
	// Channels is a list of the available release channels.
	Channels = []string{"stable", "latest", "testing"}
)

// Cluster defines share settings across all servers and agents.
type Cluster struct {
	Server Server `yaml:"server,omitempty"`
	Agent  Agent  `yaml:"agent,omitempty"`
}

// Config describes the state of a k3s cluster. For general
// reference, please refer to the k3s installation options:
// https://rancher.com/docs/k3s/latest/en/installation/install-options
type Config struct {
	// Version is the version of k3s to use. It may also be a
	// channel as specified in the k3s installation options.
	Version string `yaml:"version"`

	// Cluster defines shared configuration settings across all
	// servers and agents.
	Cluster Cluster `yaml:"cluster"`

	// Nodes is a list of nodes to deploy the cluster on. It stores
	// both, connection information and node-specific configuration.
	Nodes []Node `yaml:"nodes"`

	// SSHProxy describes the SSH connection configuration
	// for an SSH proxy, often also referred to as bastion
	// host or jumpbox.
	SSHProxy sshx.Config `yaml:"ssh-proxy"`
}

// Verify verifies the configuration file.
// TODO: How do we pass a logger to this function?
// TODO: Use logger to display configuration errors.
func (c *Config) Verify() error {
	if c == nil {
		return errors.New("configuration empty")
	}

	// TODO: Also allow versions in the format of `v1.24.4+k3s1`.
	channelValid := false
	for _, channel := range Channels {
		if channel == c.Version {
			channelValid = true
			break
		}
	}
	if !channelValid {
		return errors.New("unsupported version must be one of: " + strings.Join(Channels, ", "))
	}

	if c.Nodes == nil || len(c.Nodes) == 0 {
		return errors.New("no nodes specified")
	}

	var controlPlanes = 0
	for _, node := range c.Nodes {
		if node.Role == RoleServer {
			controlPlanes += 1
		}
	}

	if controlPlanes == 0 {
		return errors.New("no control-plane nodes specified")
	}

	if controlPlanes%2 == 0 {
		return errors.New("number of control-plane nodes must be odd")
	}

	return nil
}

// LoadConfig sets up the configuration parser and loads
// the configuration file.
func LoadConfig(configFile string) (*Config, error) {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Parse YAML config into struct.
	config := new(Config)
	if err := yaml.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}

	return config, nil
}
