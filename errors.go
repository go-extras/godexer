package godexer

import (
	"reflect"

	"github.com/go-extras/errors"
)

type CommandAwareError struct {
	err       error
	cmd       Command
	variables map[string]any
}

func NewCommandAwareError(err error, cmd Command, variables map[string]any) *CommandAwareError {
	varsCopy := make(map[string]any, len(variables))
	for k, v := range variables {
		varsCopy[k] = v
	}

	result := &CommandAwareError{
		err:       err,
		cmd:       cmd,
		variables: varsCopy,
	}

	return result
}

func (*CommandAwareError) getType(myvar any) string {
	t := reflect.TypeOf(myvar)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

func (e *CommandAwareError) Cause() error {
	return e.err
}

func (e *CommandAwareError) Variables() map[string]any {
	return e.variables
}

func (e *CommandAwareError) Command() Command {
	return e.cmd
}

func (e *CommandAwareError) Error() string {
	stepName := e.cmd.GetStepName()
	commandID := 0

	if dicmd, ok := e.cmd.(DebugInfoer); ok {
		di := dicmd.DebugInfo()
		if di != nil {
			commandID = di.ID
		}
	}

	if commandID == 0 {
		return errors.Wrapf(e.err, "command failed (stepName=%s, commandType=%s)", stepName, e.getType(e.cmd)).Error()
	}

	return errors.Wrapf(e.err, "command failed (stepName=%s, commandId=%d, commandType=%s)", stepName, commandID, e.getType(e.cmd)).Error()
}

func (e *CommandAwareError) Unwrap() error {
	return e.err
}
