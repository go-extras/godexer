package godexer_test

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

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
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

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
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

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
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

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
		ex.Value = "dummy"

		m := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}
		err := ex.Execute(m)
		c.Assert(err, qt.ErrorMatches, "variable: variable name cannot be empty")
	})

	t.Run("Execute_BoolValueFunc_False", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		// Register a boolean value function that returns false
		godexer.RegisterValueFunc("some_bool", func(path string) bool {
			return false
		})
		defer godexer.UnregisterValueFunc("some_bool")

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
		ex.Variable = "result"
		ex.Value = `{{ some_bool "test_path" }}`

		m := make(map[string]any)
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)
		c.Assert(m["result"], qt.Equals, "false")
	})

	t.Run("Execute_BoolValueFunc_True", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		// Register a boolean value function that returns true
		godexer.RegisterValueFunc("some_bool", func(path string) bool {
			return true
		})
		defer godexer.UnregisterValueFunc("some_bool")

		cmd := godexer.NewVariableCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.VariableCommand)
		ex.Variable = "result"
		ex.Value = `{{ some_bool "test_path" }}`

		m := make(map[string]any)
		err := ex.Execute(m)
		c.Assert(err, qt.IsNil)
		c.Assert(m["result"], qt.Equals, "true")
	})
}
