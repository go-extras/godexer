package executor

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
	result := &CommandAwareError{
		err:       err,
		cmd:       cmd,
		variables: variables,
	}

	for k, v := range variables {
		result.variables[k] = v
	}

	return result
}

func (e *CommandAwareError) getType(myvar any) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
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
	commandId := 0

	if dicmd, ok := e.cmd.(DebugInfoer); ok {
		di := dicmd.DebugInfo()
		if di != nil {
			commandId = di.Id
		}
	}

	if commandId == 0 {
		return errors.Wrapf(e.err, "command failed (stepName=%s, commandType=%s)", stepName, e.getType(e.cmd)).Error()
	}

	return errors.Wrapf(e.err, "command failed (stepName=%s, commandId=%d, commandType=%s)", stepName, commandId, e.getType(e.cmd)).Error()
}

func (e *CommandAwareError) Unwrap() error {
	return e.err
}
