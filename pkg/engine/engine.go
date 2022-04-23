package engine

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
	"github.com/rs/zerolog"
)

const (
	InstallerURL = "https://get.k3s.io"
)

// Engine is a type that encapsulates parts of the installation logic.
type Engine struct {
	Logger *zerolog.Logger

	sync.Mutex
	installer []byte
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

// Configure uploads the installer and the configuration
// prior to a node prior to running the installation.
func (e *Engine) Configure(node *Node) error {
	installer, err := e.Installer()
	if err != nil {
		return err
	}

	if err := node.Upload("/tmp/k3se/install.sh", bytes.NewReader(installer)); err != nil {
		return err
	}

	// TODO: Configure the "advertise address" based on the first SAN and modify the
	// kubeconfig accordingly.

	// TODO: Upload configuration and move it to appropriate location using "sudo".

	return nil
}

// Cleanup removes all temporary files from the node.
func (e *Engine) Cleanup(node *Node) error {
	if err := node.Do(sshx.Cmd{
		Cmd: "echo $MYVAR > ~/test.txt",
		Env: map[string]string{
			"MYVAR": "hello",
		},
	}); err != nil {
		return err
	}

	return node.Do(sshx.Cmd{
		Cmd: "rm -rf /tmp/k3se",
	})
}
