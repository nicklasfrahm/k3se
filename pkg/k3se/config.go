package k3se

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/nicklasfrahm/k3se/pkg/rexec"
	"gopkg.in/yaml.v3"
)

const (
	// RoleServer is the role of a control-plane node in k3s.
	RoleServer = "server"
	// RoleAgent is the role of a worker node in k3s.
	RoleAgent = "agent"
	// Program is used to configure the name of the configuration file.
	Program = "k3se"
)

// Node describes the configuration of a node.
type Node struct {
	Role      string          `yaml:"role"`
	SSH       rexec.SSHConfig `yaml:"ssh"`
	NodeLabel []string        `yaml:"node-label"`
}

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

	// Config is the desired content of the k3s configuration file.
	Cluster K3sConfig `yaml:"config"`

	// Nodes is a list of nodes to deploy the cluster on.
	Nodes []Node `yaml:"nodes"`

	// SSHProxy describes the SSH connection configuration
	// for an SSH proxy, often also referred to as bastion
	// host or jumpbox.
	SSHProxy rexec.SSHConfig `yaml:"ssh-proxy"`
}

// Verify verifies the configuration file.
func (c *Config) Verify() error {
	if c == nil {
		return errors.New("configuration empty")
	}

	checks := []func() error{
		c.VerifyServerCount,
	}

	var configInvalid bool
	for _, check := range checks {
		if err := check(); err != nil {
			configInvalid = true

			// TODO: Improve how errors are displayed and handled.
			fmt.Println(err)
		}
	}

	if configInvalid {
		return errors.New("config invalid")
	}

	// TODO: Check that backend is not SQLite if HA is enabled.

	return nil
}

// VerifyServerCount verifies that at least one API server is
// specified.
func (c *Config) VerifyServerCount() error {
	var controlPlanes = 0

	if c.Nodes == nil || len(c.Nodes) == 0 {
		return errors.New("no nodes specified")
	}

	for _, node := range c.Nodes {
		if node.Role == RoleServer {
			controlPlanes += 1
		}
	}

	if controlPlanes == 0 {
		return errors.New("no control-plane nodes specified")
	}

	return nil
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
