package executor_test

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	. "github.com/go-extras/godexer"
)

type PasswordTestSuite struct {
	suite.Suite
}

func (t *PasswordTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	cmd := NewPassword(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	pwd := cmd.(*PasswordCommand)
	pwd.Variable = "foo"
	vars := make(map[string]any)
	err := pwd.Execute(vars)
	t.NoError(err)
	t.Len(vars["foo"], 8)
}

func (t *PasswordTestSuite) TestExecute_CustomLen() {
	fs := afero.NewMemMapFs()

	cmd := NewPassword(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	pwd := cmd.(*PasswordCommand)
	pwd.Variable = "foo"
	pwd.Length = 10
	vars := make(map[string]any)
	err := pwd.Execute(vars)
	t.NoError(err)
	t.Len(vars["foo"], 10)
}

func (t *PasswordTestSuite) TestExecute_MinLen() {
	fs := afero.NewMemMapFs()

	cmd := NewPassword(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	pwd := cmd.(*PasswordCommand)
	pwd.Variable = "foo"
	pwd.Length = 5
	vars := make(map[string]any)
	err := pwd.Execute(vars)
	t.NoError(err)
	t.Len(vars["foo"], 8)
}

func (t *PasswordTestSuite) TestExecute_MissingVar() {
	fs := afero.NewMemMapFs()

	cmd := NewPassword(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	pwd := cmd.(*PasswordCommand)
	vars := make(map[string]any)
	err := pwd.Execute(vars)
	t.EqualError(err, "password: variable name cannot be empty")
}

func (t *PasswordTestSuite) TestExecute_Charset() {
	fs := afero.NewMemMapFs()

	cmd := NewPassword(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	pwd := cmd.(*PasswordCommand)
	pwd.Variable = "foo"
	pwd.Charset = "abcd"
	vars := make(map[string]any)
	err := pwd.Execute(vars)
	t.NoError(err)
	t.Regexp(regexp.MustCompile(`^[a-d]{8}$`), vars["foo"])
}

func TestPassword(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}
