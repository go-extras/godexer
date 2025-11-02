package godexer

import (
	"github.com/go-extras/errors"
)

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("variable", NewVariableCommand)
}

func NewVariableCommand(ectx *ExecutorContext) Command {
	return &VariableCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

type VariableCommand struct {
	BaseCommand
	Variable string
	Value    any
}

func (s *VariableCommand) Execute(variables map[string]any) error {
	if s.Variable == "" {
		return errors.New("variable: variable name cannot be empty")
	}

	variables[s.Variable] = MaybeEvalValue(s.Value, variables)

	return nil
}
