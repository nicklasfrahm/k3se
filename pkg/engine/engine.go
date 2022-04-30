package engine

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

const (
	InstallerURL = "https://get.k3s.io"
)

// Engine is a type that encapsulates parts of the installation logic.
type Engine struct {
	Logger *zerolog.Logger

	sync.Mutex
	installer []byte

	Nodes []*Node
	Spec  *Config
}

// New creates a new Engine.
func New(options ...Option) (*Engine, error) {
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return nil, err
	}

	return &Engine{
		Logger: opts.Logger,
	}, nil
}

// Installer returns the downloaded the k3s installer.
func (e *Engine) Installer() ([]byte, error) {
	// Lock engine to prevent concurrent access to installer cache.
	e.Lock()

	if len(e.installer) == 0 {
		resp, err := http.Get(InstallerURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		e.installer, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	e.Unlock()

	return e.installer, nil
}

// FilterNodes returns a list of nodes based on the specified selector.
// Use RoleAny to match all nodes, RoleAgent to match all worker nodes,
// and RoleServer to match all control-plane nodes. Note that this will
// a nil slice if Connect has not been called yet.
func (e *Engine) FilterNodes(selector Role) []*Node {
	var nodes []*Node

	for _, node := range e.Nodes {
		if node.Role == selector || selector == RoleAny {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// SetSpec configures the desired state of the cluster. Note
// that the config will only be applied if the verification
// succeeds.
func (e *Engine) SetSpec(config *Config) error {
	if err := config.Verify(); err != nil {
		return err
	}

	e.Spec = config

	return nil
}

// ConfigureNode uploads the installer and the configuration
// to a node prior to running the installation script.
func (e *Engine) ConfigureNode(node *Node) error {
	// Upload the installer.
	installer, err := e.Installer()
	if err != nil {
		return err
	}

	if err := node.Upload("/tmp/k3se/install.sh", bytes.NewReader(installer)); err != nil {
		return err
	}
	if err := node.Do(sshx.Cmd{
		Cmd: "chmod +x /tmp/k3se/install.sh",
	}); err != nil {
		return err
	}

	// Create the node configuration.
	config := e.Spec.Cluster.Merge(&node.Config)

	// TODO: Configure the "advertise address" based on the first SAN and modify the
	// kubeconfig accordingly.

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := node.Upload("/tmp/k3se/config.yaml", bytes.NewReader(configBytes)); err != nil {
		return err
	}

	if err := node.Do(sshx.Cmd{
		Cmd: "sudo mkdir -m 755 -p /etc/rancher/k3s",
	}); err != nil {
		return err
	}

	if err := node.Do(sshx.Cmd{
		Cmd: "sudo chown root:root /tmp/k3se/config.yaml && sudo mv /tmp/k3se/config.yaml /etc/rancher/k3s",
	}); err != nil {
		return err
	}

	// TODO: Upload configuration and move it to appropriate location using "sudo".

	return nil
}

// Install runs the installation script on the node.
func (e *Engine) Install() error {
	firstControlplane := e.FilterNodes(RoleServer)[0]

	if err := e.ConfigureNode(firstControlplane); err != nil {
		return err
	}

	if err := firstControlplane.Do(sshx.Cmd{
		Cmd: "sudo /tmp/k3se/install.sh",
		Env: map[string]string{
			"INSTALL_K3s_CHANNEL": e.Spec.Version,
		},
		Stdout: firstControlplane,
	}); err != nil {
		return err
	}

	// Force restart as the installer may have changed the configuration.
	return firstControlplane.Do(sshx.Cmd{
		Cmd: "sudo systemctl restart k3s",
	})
}

// Uninstall runs the uninstallation script on all nodes.
func (e *Engine) Uninstall() error {
	// Get a list of all nodes.
	nodes := e.FilterNodes(RoleAny)
	for _, node := range nodes {
		// TODO: Check if k3s is installed and if not skip the uninstallation.

		if err := node.Do(sshx.Cmd{
			Cmd:    "k3s-uninstall.sh",
			Shell:  true,
			Stderr: node,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Connect establishes an SSH connection to all nodes.
func (e *Engine) Connect() error {
	e.Nodes = make([]*Node, 0)

	// Establish connection to proxy if host is specified.
	var sshProxy *sshx.Client
	if e.Spec.SSHProxy.Host != "" {
		var err error
		if sshProxy, err = sshx.NewClient(&e.Spec.SSHProxy); err != nil {
			return err
		}
	}

	// Get a list of all nodes and connect to them.
	for _, node := range e.Spec.Nodes {
		// Inject logger into node.
		node.Logger = e.Logger.With().Str("host", node.SSH.Host).Logger()

		if err := node.Connect(WithSSHProxy(sshProxy), WithLogger(&node.Logger)); err != nil {
			return err
		}

		// Nodes store the connection state so we want to maintain pointers to them.
		e.Nodes = append(e.Nodes, &node)
	}

	return nil
}

// Disconnect closes all SSH connections to all nodes.
func (e *Engine) Disconnect() error {
	nodes := e.FilterNodes(RoleAny)

	for _, node := range nodes {
		// Clean up temporary files before disconnecting.
		if err := node.Do(sshx.Cmd{
			Cmd: "rm -rf /tmp/k3se",
		}); err != nil {
			return err
		}

		if err := node.Disconnect(); err != nil {
			return err
		}
	}

	return nil
}
