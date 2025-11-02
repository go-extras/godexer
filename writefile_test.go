package executor_test

import (
	"bytes"
	"io"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

func TestWriteFile(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewWriterFileCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.WriteFileCommand)
		ex.File = "dummy"
		ex.Contents = "data {{ index .  \"var1\" }}{{ index .  \"var2\" }}"
		ex.Ectx.Logger = &logger.Logger{}

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)

		f, _ := fs.Open("dummy")
		d, _ := io.ReadAll(f)
		c.Assert(string(d), qt.Equals, "data val1val2")
	})

	t.Run("Execute_MissingFile", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewWriterFileCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.WriteFileCommand)
		ex.Contents = "data {{ index .  \"var1\" }}{{ index .  \"var2\" }}"
		ex.StepName = "step"

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.ErrorMatches, "filename in \"step\" is empty")
	})
}
