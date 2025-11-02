package godexer

import (
	"encoding/json"

	"github.com/go-extras/errors"
)

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("commands", NewSubExecuteCommand)
}

// SubExecuteCommand is designed to be used by other command types
// one of the possible usages is an implementation of an include command
type SubExecuteCommand struct {
	BaseCommand
	RawCommands []json.RawMessage `json:"commands"`
}

func NewSubExecuteCommand(ectx *ExecutorContext) Command {
	return &SubExecuteCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

func (r *SubExecuteCommand) Execute(variables map[string]any) error {
	if r.Ectx.Executor == nil {
		return errors.Errorf("this command must be run from the executor")
	}

	cmdScriptObj := struct {
		Commands any `json:"commands"`
	}{}
	cmdScriptObj.Commands = r.RawCommands
	script, err := json.Marshal(cmdScriptObj)
	if err != nil {
		return errors.Wrap(err, "cannot marshal commands script")
	}

	executor, err := r.Ectx.Executor.WithScenario(string(script))
	if err != nil {
		return errors.Wrap(err, "cannot load child executor")
	}

	return executor.Execute(variables)
}
