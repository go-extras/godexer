// Package validatecmd implements the `godexer validate` command.
package validatecmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/cmd/godexer/shared"
)

// Command implements `godexer validate`.
type Command struct {
	ctx *shared.Context
	cmd *cobra.Command

	// varFiles is accepted for future template-level validation but currently unused.
	varFiles []string
}

// New creates the validate command.
func New(ctx *shared.Context) *Command {
	c := &Command{ctx: ctx}
	c.cmd = &cobra.Command{
		Use:   "validate <scenario>",
		Short: "Validate a godexer scenario without executing it",
		Long: `Parse and validate a scenario file, checking for syntax errors and
unknown command types. Use '-' as the scenario argument to read from stdin.`,
		Args: cobra.ExactArgs(1),
		RunE: c.run,
	}

	f := c.cmd.Flags()
	f.StringArrayVar(&c.varFiles, "var-file", nil, "Load variables from a YAML/JSON file (reserved for future use, repeatable)")

	return c
}

// Cmd returns the cobra command.
func (c *Command) Cmd() *cobra.Command { return c.cmd }

func (c *Command) run(cmd *cobra.Command, args []string) error {
	scenarioPath := args[0]

	var (
		content []byte
		err     error
	)

	if scenarioPath == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(scenarioPath)
	}
	if err != nil {
		return shared.NewExitError(3, fmt.Errorf("failed to read scenario: %w", err))
	}

	_, err = godexer.NewWithScenario(
		string(content),
		godexer.WithRegisteredCommandTypes(),
		godexer.WithDefaultEvaluatorFunctions(),
	)
	if err != nil {
		return shared.NewExitError(2, fmt.Errorf("validation failed: %w", err))
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Scenario is valid.")
	return nil
}
