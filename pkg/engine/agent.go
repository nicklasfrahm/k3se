package engine

// Agent describes the configuration of a k3s agent. For more information, please refer to the k3s documentation:
// https://rancher.com/docs/k3s/latest/en/installation/install-options/agent-config/#k3s-agent-cli-help
type Agent struct {
	Debug                         bool     `yaml:"debug,omitempty"`
	V                             int      `yaml:"v,omitempty"`
	VModule                       string   `yaml:"vmodule,omitempty"`
	Log                           string   `yaml:"log,omitempty"`
	AlsoLogToStderr               bool     `yaml:"also-log-to-stderr,omitempty"`
	Server                        string   `yaml:"server,omitempty"`
	DataDir                       string   `yaml:"data-dir,omitempty"`
	NodeName                      string   `yaml:"node-name,omitempty"`
	WithNodeID                    bool     `yaml:"with-node-id,omitempty"`
	NodeLabel                     []string `yaml:"node-label,omitempty"`
	NodeTaint                     []string `yaml:"node-taint,omitempty"`
	ImageCredentialProviderBinDir string   `yaml:"image-credential-provider-bin-dir,omitempty"`
	ImageCredentialProviderConfig string   `yaml:"image-credential-provider-config,omitempty"`
	Docker                        bool     `yaml:"docker,omitempty"`
	ContainerRuntimeEndpoint      string   `yaml:"container-runtime-endpoint,omitempty"`
	PauseImage                    string   `yaml:"pause-image,omitempty"`
	Snapshotter                   string   `yaml:"snapshotter,omitempty"`
	PrivateRegistry               string   `yaml:"private-registry,omitempty"`
	NodeIP                        []string `yaml:"node-ip,omitempty"`
	NodeExternalIP                []string `yaml:"node-external-ip,omitempty"`
	ResolvConf                    string   `yaml:"resolv-conf,omitempty"`
	FlannelIface                  string   `yaml:"flannel-iface,omitempty"`
	FlannelConf                   string   `yaml:"flannel-conf,omitempty"`
	KubeletArg                    []string `yaml:"kubelet-arg,omitempty"`
	KubeProxyArg                  []string `yaml:"kube-proxy-arg,omitempty"`
	ProtectKernelDefaults         bool     `yaml:"protect-kernel-defaults,omitempty"`
	Rootless                      bool     `yaml:"rootless,omitempty"`
	LBServerPort                  int      `yaml:"lb-server-port,omitempty"`
	// Note: Deprecated options, such as "--no-flannel", and all token-related flags
	// are ommitted because k3se handles tokens automatically for you.
}
