// Package rootcmd wires the root cobra.Command for the godexer CLI binary.
package rootcmd

import (
	"github.com/spf13/cobra"

	runcmd "github.com/go-extras/godexer/cmd/godexer/run"
	"github.com/go-extras/godexer/cmd/godexer/shared"
	validatecmd "github.com/go-extras/godexer/cmd/godexer/validate"
	versioncmd "github.com/go-extras/godexer/cmd/godexer/version"
)

// New creates and returns the root cobra.Command for the godexer CLI.
func New() *cobra.Command {
	ctx := &shared.Context{}

	root := &cobra.Command{
		Use:           "godexer",
		Short:         "A CLI tool for running godexer scenarios",
		Long:          `godexer runs and validates declarative YAML/JSON scenarios.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	root.AddCommand(
		runcmd.New(ctx).Cmd(),
		validatecmd.New(ctx).Cmd(),
		versioncmd.New(ctx).Cmd(),
	)

	return root
}
