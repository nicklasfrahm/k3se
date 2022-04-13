// Package rexec provides APIs to execute commands on remote machines.
package rexec

// Cmd represents an external command being prepared
// or run. The API is similar to os/exec.Cmd.
type Cmd interface {
	// Name is the command to run.
	Path() string
}

// Runner is the interface for running commands. This
// can be for example via an SSH session, inside a pod
// or on the local machine.
type Runner interface {
	// Connect establishes a connection to the execution
	// environment.
	Connect() error
	// Command prepares a command to be run.
	Command(name string, arg ...string) *Cmd
	// Disconnect closes the connection to the execution
	// environment.
	Disconnect() error
}
