package engine

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	installer      []byte
	clusterToken   string
	serverURL      string
	cleanupPending bool

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

	return nil
}

// ConfigureNode uploads the installer and the configuration
// to a node prior to running the installation script.
func (e *Engine) ConfigureNode(node *Node) error {
	e.cleanupPending = true

	node.Logger.Info().Msg("Configuring node")

	installer, err := e.fetchInstallationScript()
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

	// TODO: Make the engine smarter by checking if the node has multiple interfaces
	//       and configuring the "node-ip" if HA is enabled.

	// Create the node configuration.
	config := e.Spec.Cluster.Merge(&node.Config)

	// Remove configuration keys that are server-specific.
	// TODO: Evaluate option to have separate server and agent configuration structs.
	if node.Role == RoleAgent {
		config.WriteKubeconfigMode = ""
		config.TLSSAN = nil
	}

	if node.Role == RoleServer {
		// This ensures that agents can connect to the servers in Vagrant. For reference, see:
		// https://github.com/alexellis/k3sup/issues/306#issuecomment-1059986048
		config.AdvertiseAddress = node.SSH.Host
	}

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
	e.Logger.Info().Str("server_url", e.serverURL).Msg("Detecting server URL")

	if err := e.installControlPlanes(); err != nil {
		return err
	}

	return e.installWorkers()
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

		node.Logger.Info().Msg("Running uninstallation script")
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
		if e.cleanupPending {
			node.Logger.Info().Msg("Cleaning up temporary files")
			if err := node.Do(sshx.Cmd{
				Cmd: "rm -rf /tmp/k3se",
			}); err != nil {
				return err
			}
		}

		if err := node.Disconnect(); err != nil {
			return err
		}
	}

	return nil
}

// KubeConfig writes the kubeconfig of the cluster to the specified location.
func (e *Engine) KubeConfig(outputPath string) error {
	server := e.FilterNodes(RoleServer)[0]

	// Download kubeconfig.
	newConfigBuffer := new(bytes.Buffer)
	server.Logger.Info().Msg("Downloading kubeconfig")
	if err := server.Do(sshx.Cmd{
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

	// Rename cluster, context and auth info for humans. If k3se is running as part of a
	// CI pipeline we will not adjust the names to allow for further processing downstream.
	if os.Getenv("CI") == "" {
		// Fetch hostname from kubeconfig.
		serverURL, err := url.Parse(e.serverURL)
		if err != nil {
			return err
		}

		// Use the FQDN of the API server as the cluster name.
		cluster := serverURL.Hostname()
		context := "admin@" + cluster

		newConfig.Clusters[cluster] = newConfig.Clusters["default"]
		delete(newConfig.Clusters, "default")
		newConfig.AuthInfos[context] = newConfig.AuthInfos["default"]
		delete(newConfig.AuthInfos, "default")
		newConfig.Contexts[context] = newConfig.Contexts["default"]
		delete(newConfig.Contexts, "default")
		newConfig.Contexts[context].Cluster = cluster
		newConfig.Contexts[context].AuthInfo = context

		newConfig.CurrentContext = context
	}

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

// fetchInstallationScript returns the downloaded the k3s installer.
func (e *Engine) fetchInstallationScript() ([]byte, error) {
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

// fetchClusterToken downloads the node token used to build a cluster.
func (e *Engine) fetchClusterToken(server *Node) error {
	tokenBuffer := new(bytes.Buffer)
	if err := server.Do(sshx.Cmd{
		Cmd:    "sudo cat /var/lib/rancher/k3s/server/token",
		Stdout: tokenBuffer,
	}); err != nil {
		return err
	}

	e.clusterToken = strings.TrimSpace(tokenBuffer.String())

	return nil
}

// installControlPlanes installs the k3s servers.
func (e *Engine) installControlPlanes() error {
	// These installation options are universal to HA and non-HA clusters.
	env := map[string]string{
		"INSTALL_K3S_FORCE_RESTART": "true",
		"INSTALL_K3S_EXEC":          "server",
		"INSTALL_K3s_CHANNEL":       e.Spec.Version,
	}

	servers := e.FilterNodes(RoleServer)

	// Enable HA mode if we have more than a single control-plane.
	if len(servers) > 1 {
		env["INSTALL_K3S_EXEC"] = "server --cluster-init"
	}

	for i := 0; i < len(servers); i++ {
		server := servers[i]

		if err := e.ConfigureNode(server); err != nil {
			return err
		}

		if i > 0 {
			env["K3S_URL"] = e.serverURL
			env["K3S_TOKEN"] = e.clusterToken
		}

		server.Logger.Info().Msg("Running installation script")
		if err := server.Do(sshx.Cmd{
			Cmd:    "/tmp/k3se/install.sh",
			Env:    env,
			Stdout: server,
		}); err != nil {
			return err
		}

		if err := e.fetchClusterToken(server); err != nil {
			return err
		}
	}

	return nil
}

// installWorkers installs the k3s worker nodes.
// This function is a no-op if there are no workers.
func (e *Engine) installWorkers() error {
	agents := e.FilterNodes(RoleAgent)

	if len(agents) > 0 {
		wg := sync.WaitGroup{}

		for _, agent := range agents {
			wg.Add(1)

			go func(agent *Node) {
				defer wg.Done()

				if err := e.ConfigureNode(agent); err != nil {
					agent.Logger.Error().Err(err).Msg("Failed to configure node")
					return
				}

				agent.Logger.Info().Msg("Running installation script")
				if err := agent.Do(sshx.Cmd{
					Cmd: "/tmp/k3se/install.sh",
					Env: map[string]string{
						"INSTALL_K3S_FORCE_RESTART": "true",
						"INSTALL_K3S_EXEC":          "agent",
						"INSTALL_K3s_CHANNEL":       e.Spec.Version,
						"K3S_TOKEN":                 e.clusterToken,
						"K3S_URL":                   e.serverURL,
					},
					Stdout: agent,
				}); err != nil {
					agent.Logger.Error().Err(err).Msg("Failed to run installation script")
					return
				}

			}(agent)
		}

		wg.Wait()
	}

	return nil
}
