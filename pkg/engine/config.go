package engine

import (
	"errors"
	"io/ioutil"

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

// K3sConfig describes the configuration of a k3s node.
type K3sConfig struct {
	WriteKubeconfigMode string   `yaml:"write-kubeconfig-mode"`
	TLSSAN              []string `yaml:"tls-san"`
	NodeLabel           []string `yaml:"node-label"`
}

// Config describes the state of a k3s cluster. For general
// reference, please refer to the k3s installation options:
// https://rancher.com/docs/k3s/latest/en/installation/install-options
type Config struct {
	// Version is the version of k3s to use. It may also be a
	// channel as specified in the k3s installation options.
	Version string `yaml:"version"`

	// Config is the desired content of the k3s configuration file
	// that is shared among all nodes.
	Cluster K3sConfig `yaml:"config"`

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

	if controlPlanes > 1 {
		return errors.New("unimplemented: multiple control-plane nodes")

		// TODO: Check that backend is not SQLite if HA is enabled.
	}

	return nil
}

// NodesByRole returns a list of nodes based on the specified selector.
// Use RoleAny to match all nodes, RoleAgent to match all worker nodes,
// and RoleServer to match all control-plane nodes.
func (c *Config) NodesByRole(selector Role) []*Node {
	var nodes []*Node

	for _, node := range c.Nodes {
		if node.Role == selector || selector == RoleAny {
			nodes = append(nodes, &node)
		}
	}

	return nodes
}

// LoadConfig sets up the configuration parser and loads
// the configuration file.
func LoadConfig(configFile string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(configFile)
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