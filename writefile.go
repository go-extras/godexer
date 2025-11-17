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

	contents, ok := MaybeEvalValue(r.Contents, variables).(string)
	if !ok {
		contents = r.Contents
	}
	fileName, ok := MaybeEvalValue(r.File, variables).(string)
	if !ok {
		fileName = r.File
	}

	var mode os.FileMode = 0644

	// convert string permissions to octal mode by parsing from oct string
	if r.Permissions != "" {
		v, err := strconv.ParseUint(r.Permissions, 8, 32)
		if err == nil {
			mode = os.FileMode(v)
		}
	}

	r.Ectx.Logger.Debugf("Writing to %s", fileName)
	err := afero.WriteFile(r.Ectx.Fs, fileName, []byte(contents), mode)
	if err != nil {
		return err
	}

	return nil
}
