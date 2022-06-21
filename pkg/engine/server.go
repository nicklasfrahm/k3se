package engine

// Server describes the configuration of a k3s server. For more information, please refer to the k3s documentation:
// https://rancher.com/docs/k3s/latest/en/installation/install-options/server-config/#k3s-server-cli-help
type Server struct {
	V                             int      `yaml:"v,omitempty"`
	VModule                       string   `yaml:"vmodule,omitempty"`
	Log                           string   `yaml:"log,omitempty"`
	AlsoLogToStderr               bool     `yaml:"also-log-to-stderr,omitempty"`
	BindAddress                   string   `yaml:"bind-address,omitempty"`
	HTTPSListenPort               int      `yaml:"https-listen-port,omitempty"`
	AdvertiseAddress              string   `yaml:"advertise-address,omitempty"`
	AdvertisePort                 int      `yaml:"advertise-port,omitempty"`
	TLSSAN                        []string `yaml:"tls-san,omitempty"`
	DataDir                       string   `yaml:"data-dir,omitempty"`
	ClusterCIDR                   string   `yaml:"cluster-cidr,omitempty"`
	ServiceCIDR                   []string `yaml:"service-cidr,omitempty"`
	ServiceNodePortRange          string   `yaml:"service-node-port-range,omitempty"`
	ClusterDNS                    []string `yaml:"cluster-dns,omitempty"`
	ClusterDomain                 string   `yaml:"cluster-domain,omitempty"`
	FlannelBackend                string   `yaml:"flannel-backend,omitempty"`
	WriteKubeconfig               string   `yaml:"write-kubeconfig,omitempty"`
	WriteKubeconfigMode           string   `yaml:"write-kubeconfig-mode,omitempty"`
	EtcdArg                       []string `yaml:"etcd-arg,omitempty"`
	KubeAPIServerArg              []string `yaml:"kube-apiserver-arg,omitempty"`
	KubeSchedulerArg              []string `yaml:"kube-scheduler-arg,omitempty"`
	KubeControllerManagerArg      []string `yaml:"kube-controller-manager-arg,omitempty"`
	KubeCloudControllerManagerArg []string `yaml:"kube-cloud-controller-manager-arg,omitempty"`
	DatastoreEndpoint             string   `yaml:"datastore-endpoint,omitempty"`
	DatastoreCAFile               string   `yaml:"datastore-cafile,omitempty"`
	DatastoreCertFile             string   `yaml:"datastore-certfile,omitempty"`
	DatastoreKeyFile              string   `yaml:"datastore-keyfile,omitempty"`
	EtcdExposeMetrics             bool     `yaml:"etcd-expose-metrics,omitempty"`
	EtcdDisableSnapshots          bool     `yaml:"etcd-disable-snapshots,omitempty"`
	EtcdSnapshotName              string   `yaml:"etcd-snapshot-name,omitempty"`
	EtcdSnapshotScheduleCron      string   `yaml:"etcd-snapshot-schedule-cron,omitempty"`
	EtcdSnapshotRetention         int      `yaml:"etcd-snapshot-retention,omitempty"`
	EtcdSnapshotDir               string   `yaml:"etcd-snapshot-dir,omitempty"`
	EtcdS3                        bool     `yaml:"etcd-s3,omitempty"`
	EtcdS3Endpoint                string   `yaml:"etcd-s3-endpoint,omitempty"`
	EtcdS3EndpointCA              string   `yaml:"etcd-s3-endpoint-ca,omitempty"`
	EtcdS3SkipSSLVerify           bool     `yaml:"etcd-s3-skip-ssl-verify,omitempty"`
	EtcdS3AccessKey               string   `yaml:"etcd-s3-access-key,omitempty"`
	EtcdS3SecretKey               string   `yaml:"etcd-s3-secret-key,omitempty"`
	EtcdS3Bucket                  string   `yaml:"etcd-s3-bucket,omitempty"`
	EtcdS3Region                  string   `yaml:"etcd-s3-region,omitempty"`
	EtcdS3Folder                  string   `yaml:"etcd-s3-folder,omitempty"`
	DefaultLocalStoragePath       string   `yaml:"default-local-storage-path,omitempty"`
	Disable                       []string `yaml:"disable,omitempty"`
	DisableScheduler              bool     `yaml:"disable-scheduler,omitempty"`
	DisableCloudController        bool     `yaml:"disable-cloud-controller,omitempty"`
	DisableKubeProxy              bool     `yaml:"disable-kube-proxy,omitempty"`
	DisableNetworkPolicy          bool     `yaml:"disable-network-policy,omitempty"`
	NodeName                      string   `yaml:"node-name,omitempty"`
	WithNodeID                    bool     `yaml:"with-node-id,omitempty"`
	NodeLabel                     []string `yaml:"node-label,omitempty"`
	NodeTaint                     []string `yaml:"node-taint,omitempty"`
	ImageCredentialProviderBinDir string   `yaml:"image-credential-provider-bin-dir,omitempty"`
	ImageCredentialProviderConfig string   `yaml:"image-credential-provider-config,omitempty"`
	Docker                        string   `yaml:"docker,omitempty"`
	ContainerRuntimeEndpoint      string   `yaml:"container-runtime-endpoint,omitempty"`
	PauseImage                    string   `yaml:"pause-image,omitempty"`
	Snapshotter                   string   `yaml:"snapshotter,omitempty"`
	PrivateRegistry               string   `yaml:"private-registry,omitempty"`
	NodeIP                        []string `yaml:"node-ip,omitempty"`
	NodeExternalIP                []string `yaml:"node-external-ip,omitempty"`
	ResolvConf                    string   `yaml:"resolv-conf,omitempty"`
	KubeletArg                    []string `yaml:"kubelet-arg,omitempty"`
	KubeProxyArg                  []string `yaml:"kube-proxy-arg,omitempty"`
	ProtectKernelDefaults         bool     `yaml:"protect-kernel-defaults,omitempty"`
	Rootless                      bool     `yaml:"rootless,omitempty"`
	Server                        string   `yaml:"server,omitempty"`
	// Options to manage the clustering, such as "--cluster-init", are omitted as this
	// is handled automatically by the engine.
	ClusterResetRestorePath bool   `yaml:"cluster-reset-restore-path,omitempty"`
	SecretsEncryption       bool   `yaml:"secrets-encryption,omitempty"`
	SystemDefaultRegistry   string `yaml:"system-default-registry,omitempty"`
	SELinux                 bool   `yaml:"selinux,omitempty"`
	LBServerPort            int    `yaml:"lb-server-port,omitempty"`
	// Deprecated options, such as "--no-flannel", are omitted.
}
