package engine

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
)

const (
	InstallerURL = "https://get.k3s.io"
)

// Engine is a type that encapsulates parts of the installation logic.
type Engine struct {
	Logger *zerolog.Logger

	sync.Mutex
	installer    []byte
	clusterToken string
	serverURL    string

	Spec *Config
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
// and RoleServer to match all control-plane nodes.
func (e *Engine) FilterNodes(selector Role) []*Node {
	var nodes []*Node

	// We are NOT using range here because we need pointers to the
	// actual nodes inside the Spec holding the connection state.
	// This would not work with range as it does "call-by-value",
	// meaning that the value of the iterator is a copy of the value.
	for i := 0; i < len(e.Spec.Nodes); i++ {
		node := &e.Spec.Nodes[i]

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

	// If TLS SANs are configured, the first one will be used as the server URL.
	// If not, the host address of the first controlplane will be used.
	firstControlplane := e.FilterNodes(RoleServer)[0]
	e.serverURL = fmt.Sprintf("https://%s:6443", firstControlplane.SSH.Host)
	if len(e.Spec.Cluster.TLSSAN) > 0 {
		e.serverURL = fmt.Sprintf("https://%s:6443", e.Spec.Cluster.TLSSAN[0])
	}
	e.Logger.Info().Str("server_url", e.serverURL).Msgf("Detecting server URL")

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

	// Remove configuration keys that are server-specific.
	// TODO: Evaluate option to have separate server and agent configuration structs.
	if node.Role == RoleAgent {
		config.WriteKubeconfigMode = ""
		config.TLSSAN = nil
	}

	if node.Role == RoleServer {
		// This ensures that agent can connect to the servers in Vagrant. For reference, see:
		// https://github.com/alexellis/k3sup/issues/306#issuecomment-1059986048
		config.AdvertiseAddress = node.SSH.Host
	}

	// TODO: Configure the "advertise address" based on the first
	//       SAN and modify the kubeconfig accordingly.

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

	return nil
}

// Install runs the installation script on the node.
func (e *Engine) Install() error {
	firstControlplane := e.FilterNodes(RoleServer)[0]

	if err := e.ConfigureNode(firstControlplane); err != nil {
		return err
	}

	if err := firstControlplane.Do(sshx.Cmd{
		Cmd: "/tmp/k3se/install.sh",
		Env: map[string]string{
			"INSTALL_K3S_FORCE_RESTART": "true",
			"INSTALL_K3s_CHANNEL":       e.Spec.Version,
		},
		Stdout: firstControlplane,
	}); err != nil {
		return err
	}

	agents := e.FilterNodes(RoleAgent)
	if len(agents) > 0 {
		// Download cluster token.
		token := new(bytes.Buffer)
		if err := firstControlplane.Do(sshx.Cmd{
			Cmd:    "sudo cat /var/lib/rancher/k3s/server/node-token",
			Stdout: token,
		}); err != nil {
			return err
		}
		e.clusterToken = strings.TrimSpace(token.String())

		// Configure the agents.
		wg := sync.WaitGroup{}
		for _, agent := range agents {
			wg.Add(1)
			go func(agent *Node) {
				defer wg.Done()

				if err := e.installAgent(agent); err != nil {
					return
				}
			}(agent)
		}

		wg.Wait()
	}

	return nil
}

// Uninstall runs the uninstallation script on all nodes.
func (e *Engine) Uninstall() error {
	// Get a list of all nodes.
	nodes := e.FilterNodes(RoleAny)
	for _, node := range nodes {
		// TODO: Check if k3s is installed and if not skip the uninstallation.

		uninstallScript := "k3s-uninstall.sh"
		if node.Role == RoleAgent {
			uninstallScript = "k3s-agent-uninstall.sh"
		}

		if err := node.Do(sshx.Cmd{
			Cmd:    uninstallScript,
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
	// Establish connection to proxy if host is specified.
	var sshProxy *sshx.Client
	if e.Spec.SSHProxy.Host != "" {
		var err error
		if sshProxy, err = sshx.NewClient(&e.Spec.SSHProxy); err != nil {
			return err
		}
	}

	// Get a list of all nodes and connect to them.
	for i := 0; i < len(e.Spec.Nodes); i++ {
		// We need to create a proper handle here as the nodes in the Spec
		// will hold the connection state and range only does "call-by-value".
		node := &e.Spec.Nodes[i]

		// Inject logger into node.
		node.Logger = e.Logger.With().Str("host", node.SSH.Host).Logger()

		if err := node.Connect(WithSSHProxy(sshProxy), WithLogger(&node.Logger)); err != nil {
			return err
		}
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

// installAgent installs a k3s worker node.
func (e *Engine) installAgent(node *Node) error {
	if err := e.ConfigureNode(node); err != nil {
		node.Logger.Error().Err(err).Msg("Failed to configure node")
		return err
	}

	if err := node.Do(sshx.Cmd{
		Cmd: "/tmp/k3se/install.sh",
		Env: map[string]string{
			"INSTALL_K3S_FORCE_RESTART": "true",
			"INSTALL_K3s_CHANNEL":       e.Spec.Version,
			"K3S_TOKEN":                 e.clusterToken,
			"K3S_URL":                   e.serverURL,
		},
		Stdout: node,
	}); err != nil {
		node.Logger.Error().Err(err).Msg("Failed to run installation script")
		return err
	}

	return nil
}

// KubeConfig writes the kubeconfig of the cluster to the specified location.
func (e *Engine) KubeConfig(outputPath string) error {
	firstControlPlane := e.FilterNodes(RoleServer)[0]

	// Download kubeconfig.
	newConfigBuffer := new(bytes.Buffer)
	if err := firstControlPlane.Do(sshx.Cmd{
		Cmd:    "sudo cat /etc/rancher/k3s/k3s.yaml",
		Stdout: newConfigBuffer,
	}); err != nil {
		return err
	}

	// Fix API server URL.
	newConfig, err := clientcmd.Load(newConfigBuffer.Bytes())
	if err != nil {
		e.Logger.Error().Err(err).Msg("Failed to parse kubeconfig")
		return err
	}
	// To my knowledge k3s always names its cluster, auth info and context "default".
	newConfig.Clusters["default"].Server = e.serverURL

	// TODO: Rename cluster, context and auth info for humans if env["CI"] is unset.

	// Resolve the home directory in the output path.
	if outputPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		outputPath = filepath.Join(home, outputPath[1:])
	}

	// Read existing local config.
	oldConfigBytes, err := ioutil.ReadFile(outputPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// If the file does not exist, we can just write the new config.
		if err := clientcmd.WriteToFile(*newConfig, outputPath); err != nil {
			return err
		}
		return nil
	}

	// Parse existing local config.
	oldConfig, err := clientcmd.Load(oldConfigBytes)
	if err != nil {
		return err
	}

	// Merge the new config with the existing one.
	for name, cluster := range newConfig.Clusters {
		oldConfig.Clusters[name] = cluster
	}
	for name, authInfo := range newConfig.AuthInfos {
		oldConfig.AuthInfos[name] = authInfo
	}
	for name, context := range newConfig.Contexts {
		oldConfig.Contexts[name] = context
	}

	return clientcmd.WriteToFile(*oldConfig, outputPath)
}
