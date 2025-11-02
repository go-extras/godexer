package godexer

import (
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
	File     string
	Contents string
}

func (r *WriteFileCommand) Execute(variables map[string]any) error {
	if len(r.File) == 0 {
		return errors.Errorf("filename in %q is empty", r.StepName)
	}

	contents := MaybeEvalValue(r.Contents, variables)

	r.Ectx.Logger.Debugf("Writing to %s", r.File)
	err := afero.WriteFile(r.Ectx.Fs, r.File, []byte(contents.(string)), 0644)
	if err != nil {
		return err
	}

	return nil
}
