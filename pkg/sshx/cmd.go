package sshx

import (
	"fmt"
	"io"
)

// Cmd describes a command to be executed on the remote host.
type Cmd struct {
	Cmd    string
	Env    map[string]string
	Shell  bool
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// String compiles the command to be executed. I am aware that
// this is not the most efficient way to do this because it does
// a lot of reallocations.
func (c *Cmd) String() string {
	cmd := c.Cmd

	// Note that we also need to wrap the command in a
	// shell if we want to inject environment variables.
	if c.Shell || c.Env != nil {
		cmd = fmt.Sprintf("sh -c '%s'", c.Cmd)
	}

	if c.Env != nil {
		for k, v := range c.Env {
			cmd = fmt.Sprintf("%s='%s' %s", k, v, cmd)
		}

		cmd = fmt.Sprintf("env %s", cmd)
	}

	fmt.Println(cmd)

	return cmd
}
