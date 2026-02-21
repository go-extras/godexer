// Package versioncmd implements the `godexer version` command.
package versioncmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-extras/godexer/cmd/godexer/shared"
)

// AppVersion is the binary version, set at build time via:
//
//	-ldflags "-X github.com/go-extras/godexer/cmd/godexer/version.AppVersion=v1.2.3"
var AppVersion = "dev"

// Command implements `godexer version`.
type Command struct {
	ctx *shared.Context
	cmd *cobra.Command
}

// New creates the version command.
func New(ctx *shared.Context) *Command {
	c := &Command{ctx: ctx}
	c.cmd = &cobra.Command{
		Use:   "version",
		Short: "Print the godexer CLI version",
		RunE:  c.run,
	}
	return c
}

// Cmd returns the cobra command.
func (c *Command) Cmd() *cobra.Command { return c.cmd }

func (*Command) run(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "godexer version %s\n", AppVersion)
	return nil
}
