package godexer_test

import (
	"bytes"
	"io"
	"os"
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

		cmd := godexer.NewWriterFileCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.WriteFileCommand)
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

		// Verify default permissions (0644)
		info, err := fs.Stat("dummy")
		c.Assert(err, qt.IsNil)
		c.Assert(info.Mode().Perm(), qt.Equals, os.FileMode(0644))
	})

	t.Run("Execute_CustomPermissions", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewWriterFileCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.WriteFileCommand)
		ex.File = "dummy"
		ex.Contents = "test content"
		ex.Permissions = "0755"
		ex.Ectx.Logger = &logger.Logger{}

		m := map[string]any{}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)

		f, _ := fs.Open("dummy")
		d, _ := io.ReadAll(f)
		c.Assert(string(d), qt.Equals, "test content")

		// Verify custom permissions (0755)
		info, err := fs.Stat("dummy")
		c.Assert(err, qt.IsNil)
		c.Assert(info.Mode().Perm(), qt.Equals, os.FileMode(0755))
	})

	t.Run("Execute_InvalidPermissions", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewWriterFileCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.WriteFileCommand)
		ex.File = "dummy"
		ex.Contents = "test content"
		ex.Permissions = "invalid"
		ex.Ectx.Logger = &logger.Logger{}

		m := map[string]any{}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)

		// Verify it falls back to default permissions (0644) on invalid input
		info, err := fs.Stat("dummy")
		c.Assert(err, qt.IsNil)
		c.Assert(info.Mode().Perm(), qt.Equals, os.FileMode(0644))
	})

	t.Run("Execute_MissingFile", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewWriterFileCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.WriteFileCommand)
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
