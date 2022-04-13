package rexec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/user"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

// SSHConfig describes the SSH connection configuration.
type SSHConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	KeyFile     string `yaml:"key-file"`
	Key         string `yaml:"key"`
	Passphrase  string `yaml:"passphrase"`
	Fingerprint string `yaml:"fingerprint"`
}

// SSH is a runner that executes commands on a remote host via SSH.
type SSH struct {
	Logger  *zerolog.Logger
	Proxy   *SSHConfig
	Target  *SSHConfig
	Timeout time.Duration

	proxyClient  *ssh.Client
	targetClient *ssh.Client
}

// NewSSH returns a new SSH-based runner.
func NewSSH(target *SSHConfig, options ...Option) (*SSH, error) {
	opts, err := GetDefaultOptions().Apply(options...)
	if err != nil {
		return nil, err
	}

	if target.Port == 0 {
		target.Port = 22
	}

	proxy := opts.SSHProxy
	if proxy != nil {
		if proxy.Port == 0 {
			proxy.Port = 22
		}
	}

	return &SSH{
		Logger:  opts.Logger,
		Proxy:   proxy,
		Target:  target,
		Timeout: opts.Timeout,
	}, nil
}

// NewClientConfig creates a new client config that is compatible with
// the `golang.org/x/crypto/ssh` package.
func (runner *SSH) NewClientConfig(config *SSHConfig) (*ssh.ClientConfig, error) {
	// Set default values.
	username := config.User
	if username == "" {
		username = "root"
	}

	// Load the private key. A key that is specified directly takes
	// precedence over a key file.
	key := config.Key
	if key == "" && config.KeyFile != "" {
		// Resolve the home directory if necessary.
		if config.KeyFile[0] == '~' {
			userInfo, err := user.Current()
			if err != nil {
				return nil, err
			}
			config.KeyFile = userInfo.HomeDir + config.KeyFile[1:]
		}

		keyBytes, err := ioutil.ReadFile(config.KeyFile)
		if err != nil {
			return nil, err
		}
		key = string(keyBytes)
	}

	var authMethod ssh.AuthMethod
	if key != "" {
		// Use passphrase to decrypt the private key.
		if config.Passphrase != "" {
			signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(config.Passphrase))
			if err != nil {
				return nil, err
			}
			authMethod = ssh.PublicKeys(signer)
		} else {
			signer, err := ssh.ParsePrivateKey([]byte(key))
			if err != nil {
				return nil, err
			}
			authMethod = ssh.PublicKeys(signer)
		}
	} else if config.Password != "" {
		// Fall back to password authentication.
		authMethod = ssh.Password(config.Password)
		runner.Logger.Warn().Msg("Using password authentication is insecure!")
		runner.Logger.Warn().Msg("Please consider using public key authentication!")
	} else {
		return nil, errors.New("no authentication method specified")
	}

	var hostKeyCallback ssh.HostKeyCallback
	if config.Fingerprint != "" {
		// Configure host key verification.
		hostKeyCallback = func(hostname string, remote net.Addr, pubKey ssh.PublicKey) error {
			fingerprint := ssh.FingerprintSHA256(pubKey)
			if config.Fingerprint != fingerprint {
				return fmt.Errorf("fingerprint mismatch: server fingerprint: %s", fingerprint)
			}
			return nil
		}
	} else {
		runner.Logger.Warn().Msg("Skipping host key verification is insecure!")
		runner.Logger.Warn().Msg("This allows for person-in-the-middle attacks!")
		runner.Logger.Warn().Msg("Please consider using fingerprint verification!")
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return &ssh.ClientConfig{
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: hostKeyCallback,
		User:            username,
		Timeout:         runner.Timeout,
	}, nil
}

// Connect establishes a connection to the SSH host.
func (runner *SSH) Connect() error {
	targetAddress := fmt.Sprintf("%s:%d", runner.Target.Host, runner.Target.Port)
	targetConfig, err := runner.NewClientConfig(runner.Target)
	if err != nil {
		return err
	}

	if runner.Proxy != nil {
		proxyAddress := fmt.Sprintf("%s:%d", runner.Proxy.Host, runner.Proxy.Port)
		proxyConfig, err := runner.NewClientConfig(runner.Proxy)
		if err != nil {
			return err
		}

		runner.proxyClient, err = ssh.Dial("tcp", proxyAddress, proxyConfig)
		if err != nil {
			return err
		}

		// Create a TCP connection to from the proxy host to the target.
		netConn, err := runner.proxyClient.Dial("tcp", targetAddress)
		if err != nil {
			return err
		}

		targetConn, channel, req, err := ssh.NewClientConn(netConn, targetAddress, targetConfig)
		if err != nil {
			return err
		}

		runner.targetClient = ssh.NewClient(targetConn, channel, req)
	} else {
		if runner.targetClient, err = ssh.Dial("tcp", targetAddress, targetConfig); err != nil {
			return err
		}
	}

	return nil
}

// Disconnect closes the SSH connections in reverse order to how they were opened.
func (runner *SSH) Disconnect() error {
	if runner.targetClient != nil {
		if err := runner.targetClient.Close(); err != nil {
			return err
		}
		runner.targetClient = nil
	}

	if runner.proxyClient != nil {
		if err := runner.proxyClient.Close(); err != nil {
			return err
		}
		runner.proxyClient = nil
	}

	return nil
}
