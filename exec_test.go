package godexer_test

import (
	"bytes"
	"io"
	"os/exec"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

func TestExec(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd := godexer.NewExecCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &stdout,
			Stderr: &stderr,
		})
		ex := cmd.(*godexer.ExecCommand)
		ex.Cmd = []string{"test"}
		ex.Ectx.Logger = &logger.Logger{}
		ex.Env = []string{"DUMMY=1"}

		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()
		vars := make(map[string]any)
		err := ex.Execute(vars)
		c.Assert(err, qt.IsNil)
		d, _ := io.ReadAll(&stdout)
		c.Assert(string(d), qt.Equals, runnerResultStdout)
		d, _ = io.ReadAll(&stderr)
		c.Assert(string(d), qt.Equals, "test")
	})

	t.Run("Execute_WithVariable", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd := godexer.NewExecCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &stdout,
			Stderr: &stderr,
		})
		ex := cmd.(*godexer.ExecCommand)
		ex.Cmd = []string{"test"}
		ex.Ectx.Logger = &logger.Logger{}
		ex.Variable = "myvar"
		ex.Env = []string{"DUMMY=1"}

		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()
		vars := make(map[string]any)
		err := ex.Execute(vars)
		c.Assert(err, qt.IsNil)
		d, _ := io.ReadAll(&stdout)
		c.Assert(string(d), qt.Equals, runnerResultStdout)
		d, _ = io.ReadAll(&stderr)
		c.Assert(string(d), qt.Equals, "test")
		c.Assert(vars["myvar"], qt.Equals, "foo!test")
	})

	t.Run("Execute_WithExitStatus", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd := godexer.NewExecCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &stdout,
			Stderr: &stderr,
		})
		ex := cmd.(*godexer.ExecCommand)
		ex.Cmd = []string{"error"}
		ex.Ectx.Logger = &logger.Logger{}
		ex.AllowFail = true
		ex.StepName = "test"
		ex.Env = []string{"DUMMY=1"}

		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()
		vars := make(map[string]any)
		err := ex.Execute(vars)
		c.Assert(err, qt.IsNil)
		d, _ := io.ReadAll(&stdout)
		c.Assert(string(d), qt.Equals, runnerResultStdout)
		d, _ = io.ReadAll(&stderr)
		c.Assert(string(d), qt.Equals, "error")
		c.Assert(vars["test_exit_status"], qt.Equals, 1)
	})

	t.Run("Execute_Empty", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		cmd := godexer.NewExecCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		})
		ex := cmd.(*godexer.ExecCommand)
		ex.StepName = "dummy"

		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()
		err := ex.Execute(nil)
		c.Assert(err, qt.ErrorMatches, "command \"dummy\" is empty")
	})

	t.Run("Execute_EvaluteVariables", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := godexer.NewExecCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &stdout,
			Stderr: &stderr,
		})
		ex := cmd.(*godexer.ExecCommand)
		ex.Cmd = []string{"{{ index .  \"var1\" }}{{ index .  \"var2\" }}"}
		ex.Ectx.Logger = &logger.Logger{}
		ex.Env = []string{"DUMMY=1"}

		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()
		err := ex.Execute(map[string]any{
			"var1": "val1",
			"var2": "val2",
		})
		c.Assert(err, qt.IsNil)
		d, _ := io.ReadAll(&stdout)
		c.Assert(string(d), qt.Equals, runnerResultStdout)
		d, _ = io.ReadAll(&stderr)
		c.Assert(string(d), qt.Equals, "val1val2")
	})
}
