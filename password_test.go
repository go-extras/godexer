package godexer_test

import (
	"bytes"
	"regexp"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
)

func TestPassword(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewPassword(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		pwd := cmd.(*godexer.PasswordCommand)
		pwd.Variable = "foo"
		vars := make(map[string]any)
		err := pwd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(len(vars["foo"].(string)), qt.Equals, 8)
	})

	t.Run("Execute_CustomLen", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewPassword(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		pwd := cmd.(*godexer.PasswordCommand)
		pwd.Variable = "foo"
		pwd.Length = 10
		vars := make(map[string]any)
		err := pwd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(len(vars["foo"].(string)), qt.Equals, 10)
	})

	t.Run("Execute_MinLen", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewPassword(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		pwd := cmd.(*godexer.PasswordCommand)
		pwd.Variable = "foo"
		pwd.Length = 5
		vars := make(map[string]any)
		err := pwd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(len(vars["foo"].(string)), qt.Equals, 8)
	})

	t.Run("Execute_MissingVar", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewPassword(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		pwd := cmd.(*godexer.PasswordCommand)
		vars := make(map[string]any)
		err := pwd.Execute(vars)
		c.Assert(err, qt.ErrorMatches, "password: variable name cannot be empty")
	})

	t.Run("Execute_Charset", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewPassword(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		pwd := cmd.(*godexer.PasswordCommand)
		pwd.Variable = "foo"
		pwd.Charset = "abcd"
		vars := make(map[string]any)
		err := pwd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(vars["foo"], qt.Matches, regexp.MustCompile(`^[a-d]{8}$`))
	})
}
