package executor_test

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	. "github.com/go-extras/godexer"
)

type VariableTestSuite struct {
	suite.Suite
}

func (t *VariableTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	cmd := NewVariableCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*VariableCommand)
	ex.Variable = "foo"
	ex.Value = "{{ index .  \"var1\" }}{{ index .  \"var2\" }}"

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.NoError(err)
	t.Equal("val1val2", m["foo"])
}

func (t *VariableTestSuite) TestExecute_Int() {
	fs := afero.NewMemMapFs()

	cmd := NewVariableCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*VariableCommand)
	ex.Variable = "foo"
	ex.Value = 1

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.NoError(err)
	t.Equal(1, m["foo"])
}

func (t *VariableTestSuite) TestExecute_Empty() {
	fs := afero.NewMemMapFs()

	cmd := NewVariableCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*VariableCommand)
	ex.Variable = "foo"

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.NoError(err)
	t.Equal(nil, m["foo"])
}

func (t *VariableTestSuite) TestExecute_NoVariable() {
	fs := afero.NewMemMapFs()

	cmd := NewVariableCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*VariableCommand)
	ex.Value = "dummy"

	m := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}
	err := ex.Execute(m)
	t.EqualError(err, "variable: variable name cannot be empty")
}

func TestVariable(t *testing.T) {
	suite.Run(t, new(VariableTestSuite))
}
