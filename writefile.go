package godexer

import (
	"os"
	"strconv"

	"github.com/go-extras/errors"
	"github.com/spf13/afero"
)

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("writefile", NewWriterFileCommand)
}

func NewWriterFileCommand(ectx *ExecutorContext) Command {
	return &WriteFileCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

type WriteFileCommand struct {
	BaseCommand
	Contents    string
	File        string
	Permissions string
}

func (r *WriteFileCommand) Execute(variables map[string]any) error {
	if len(r.File) == 0 {
		return errors.Errorf("filename in %q is empty", r.StepName)
	}

	contents := MaybeEvalValue(r.Contents, variables)

	var mode os.FileMode = 0644

	// convert string permissions to octal mode by parsing from oct string
	if r.Permissions != "" {
		v, err := strconv.ParseUint(r.Permissions, 8, 32)
		if err == nil {
			mode = os.FileMode(v)
		}
	}

	r.Ectx.Logger.Debugf("Writing to %s", r.File)
	err := afero.WriteFile(r.Ectx.Fs, r.File, []byte(contents.(string)), mode)
	if err != nil {
		return err
	}

	return nil
}
