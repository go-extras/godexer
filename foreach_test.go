package executor_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	. "github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

type ForeachTestSuite struct {
	suite.Suite
}

func (t *ForeachTestSuite) TestExecuteWithIterableWithSlice() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.RawCommands = []json.RawMessage{
		[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
		[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
	}
	ex.Iterable = []int{1, 2, 3}
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	err = ex.Execute(variables)
	t.Require().NoError(err)
	t.Equal("k=0\nv=1\nk=1\nv=2\nk=2\nv=3\n", memlog.String())
}

func (t *ForeachTestSuite) TestExecuteWithIterableWithMap() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.RawCommands = []json.RawMessage{
		[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
		[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
	}
	ex.Iterable = map[string]string{"dummy1": "yummy1", "dummy2": "yummy2", "dummy3": "yummy3"}
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	err = ex.Execute(variables)
	t.Require().NoError(err)
	t.Contains(memlog.String(), "k=dummy1\nv=yummy1\n")
	t.Contains(memlog.String(), "k=dummy2\nv=yummy2\n")
	t.Contains(memlog.String(), "k=dummy3\nv=yummy3\n")
}

func (t *ForeachTestSuite) TestExecuteWithSlice() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.RawCommands = []json.RawMessage{
		[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
		[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
	}
	ex.Variable = "slice"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	variables["slice"] = []int{1, 2, 3}
	err = ex.Execute(variables)
	t.Require().NoError(err)
	t.Equal("k=0\nv=1\nk=1\nv=2\nk=2\nv=3\n", memlog.String())
}

func (t *ForeachTestSuite) TestExecuteWithMap() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.RawCommands = []json.RawMessage{
		[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
		[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
	}
	ex.Variable = "map"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	variables["map"] = map[string]string{"dummy1": "yummy1", "dummy2": "yummy2", "dummy3": "yummy3"}
	err = ex.Execute(variables)
	t.Require().NoError(err)
	t.Contains(memlog.String(), "k=dummy1\nv=yummy1\n")
	t.Contains(memlog.String(), "k=dummy2\nv=yummy2\n")
	t.Contains(memlog.String(), "k=dummy3\nv=yummy3\n")
}

func (t *ForeachTestSuite) TestExecute_MissingExecutor() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	variables := make(map[string]any)
	err := ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("this command must be run from the executor", err.Error())
}

func (t *ForeachTestSuite) TestExecute_MissingVariable() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	err = ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("either iterable or variable must be set", err.Error())
}

func (t *ForeachTestSuite) TestExecute_MissingVariable2() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.Variable = "dummy"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	err = ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("variable \"dummy\" does not exist", err.Error())
}

func (t *ForeachTestSuite) TestExecute_MissingVariable3() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.Variable = "dummy"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	variables["dummy"] = nil
	err = ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("variable \"dummy\" does not exist", err.Error())
}

func (t *ForeachTestSuite) TestExecute_WrongVariableType() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.Variable = "dummy"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	variables["dummy"] = struct{}{}
	err = ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("foreach: invalid variable type \"struct\" (expected slice or map)", err.Error())
}

func (t *ForeachTestSuite) TestExecute_WrongVariableType2() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := NewForeachCommand(&ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*ForeachCommand)
	ex.Variable = "dummy"
	commands := make(map[string]func(*ExecutorContext) Command)
	commands["message"] = NewMessageCommand

	exc, err := NewWithScenario(
		"",
		WithStdout(os.Stdout),
		WithStderr(os.Stderr),
		WithFS(fs),
		WithCommandTypes(commands),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	variables["dummy"] = map[int]string{1: ""}
	err = ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("foreach: invalid map key type \"int\" (expected string)", err.Error())
}

func TestForeach(t *testing.T) {
	suite.Run(t, new(ForeachTestSuite))
}
