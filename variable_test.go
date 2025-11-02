package executor_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
)

func TestVariable(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewVariableCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.VariableCommand)
		ex.Variable = "foo"
		ex.Value = "{{ index .  \"var1\" }}{{ index .  \"var2\" }}"

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)
		c.Assert(m["foo"], qt.Equals, "val1val2")
	})

	t.Run("Execute_Int", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewVariableCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.VariableCommand)
		ex.Variable = "foo"
		ex.Value = 1

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)
		c.Assert(m["foo"], qt.Equals, 1)
	})

	t.Run("Execute_Empty", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewVariableCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.VariableCommand)
		ex.Variable = "foo"

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)
		c.Assert(m["foo"], qt.IsNil)
	})

	t.Run("Execute_NoVariable", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := executor.NewVariableCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*executor.VariableCommand)
		ex.Value = "dummy"

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.ErrorMatches, "variable: variable name cannot be empty")
	})
}
