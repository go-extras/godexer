package executor_test

import (
	"bytes"
	"io"
	"os/exec"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
)

type ExecTestSuite struct {
	suite.Suite
}

func (t *ExecTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := executor.NewExecCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	ex := cmd.(*executor.ExecCommand)
	ex.Cmd = []string{"test"}
	ex.Ectx.Logger = &logger.Logger{}
	ex.Env = []string{"DUMMY=1"}

	executor.ExecCommandFn = fakeExecCommand
	defer func() { executor.ExecCommandFn = exec.Command }()
	vars := make(map[string]any)
	err := ex.Execute(vars)
	t.NoError(err)
	d, _ := io.ReadAll(&stdout)
	t.Equal(runnerResultStdout, string(d))
	d, _ = io.ReadAll(&stderr)
	t.Equal("test", string(d))
}

func (t *ExecTestSuite) TestExecute_WithVariable() {
	fs := afero.NewMemMapFs()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := executor.NewExecCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	ex := cmd.(*executor.ExecCommand)
	ex.Cmd = []string{"test"}
	ex.Ectx.Logger = &logger.Logger{}
	ex.Variable = "myvar"
	ex.Env = []string{"DUMMY=1"}

	executor.ExecCommandFn = fakeExecCommand
	defer func() { executor.ExecCommandFn = exec.Command }()
	vars := make(map[string]any)
	err := ex.Execute(vars)
	t.Require().NoError(err)
	d, _ := io.ReadAll(&stdout)
	t.Equal(runnerResultStdout, string(d))
	d, _ = io.ReadAll(&stderr)
	t.Equal("test", string(d))
	t.Equal("foo!test", vars["myvar"])
}

func (t *ExecTestSuite) TestExecute_WithExitStatus() {
	fs := afero.NewMemMapFs()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := executor.NewExecCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	ex := cmd.(*executor.ExecCommand)
	ex.Cmd = []string{"error"}
	ex.Ectx.Logger = &logger.Logger{}
	ex.AllowFail = true
	ex.StepName = "test"
	ex.Env = []string{"DUMMY=1"}

	executor.ExecCommandFn = fakeExecCommand
	defer func() { executor.ExecCommandFn = exec.Command }()
	vars := make(map[string]any)
	err := ex.Execute(vars)
	t.Require().NoError(err)
	d, _ := io.ReadAll(&stdout)
	t.Equal(runnerResultStdout, string(d))
	d, _ = io.ReadAll(&stderr)
	t.Equal("error", string(d))
	t.Equal(1, vars["test_exit_status"])
}

func (t *ExecTestSuite) TestExecute_Empty() {
	fs := afero.NewMemMapFs()

	cmd := executor.NewExecCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*executor.ExecCommand)
	ex.StepName = "dummy"

	executor.ExecCommandFn = fakeExecCommand
	defer func() { executor.ExecCommandFn = exec.Command }()
	err := ex.Execute(nil)
	t.EqualError(err, "command \"dummy\" is empty")
}

func (t *ExecTestSuite) TestExecute_EvaluteVariables() {
	fs := afero.NewMemMapFs()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := executor.NewExecCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	ex := cmd.(*executor.ExecCommand)
	ex.Cmd = []string{"{{ index .  \"var1\" }}{{ index .  \"var2\" }}"}
	ex.Ectx.Logger = &logger.Logger{}
	ex.Env = []string{"DUMMY=1"}

	executor.ExecCommandFn = fakeExecCommand
	defer func() { executor.ExecCommandFn = exec.Command }()
	err := ex.Execute(map[string]any{
		"var1": "val1",
		"var2": "val2",
	})
	t.NoError(err)
	d, _ := io.ReadAll(&stdout)
	t.Equal(runnerResultStdout, string(d))
	d, _ = io.ReadAll(&stderr)
	t.Equal("val1val2", string(d))
}

func TestExec(t *testing.T) {
	suite.Run(t, new(ExecTestSuite))
}
