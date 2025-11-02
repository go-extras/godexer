package executor_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

func TestForeach(t *testing.T) {
	t.Run("ExecuteWithIterableWithSlice", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.RawCommands = []json.RawMessage{
			[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
			[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
		}
		ex.Iterable = []int{1, 2, 3}
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "k=0\nv=1\nk=1\nv=2\nk=2\nv=3\n")
	})

	t.Run("ExecuteWithIterableWithMap", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.RawCommands = []json.RawMessage{
			[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
			[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
		}
		ex.Iterable = map[string]string{"dummy1": "yummy1", "dummy2": "yummy2", "dummy3": "yummy3"}
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Contains, "k=dummy1\nv=yummy1\n")
		c.Assert(memlog.String(), qt.Contains, "k=dummy2\nv=yummy2\n")
		c.Assert(memlog.String(), qt.Contains, "k=dummy3\nv=yummy3\n")
	})

	t.Run("ExecuteWithSlice", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.RawCommands = []json.RawMessage{
			[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
			[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
		}
		ex.Variable = "slice"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		variables["slice"] = []int{1, 2, 3}
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "k=0\nv=1\nk=1\nv=2\nk=2\nv=3\n")
	})

	t.Run("ExecuteWithMap", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.RawCommands = []json.RawMessage{
			[]byte(`{"type": "message","stepName": "test","description": "k={{.key}}"}`),
			[]byte(`{"type": "message","stepName": "test2","description": "v={{.value}}"}`),
		}
		ex.Variable = "map"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		variables["map"] = map[string]string{"dummy1": "yummy1", "dummy2": "yummy2", "dummy3": "yummy3"}
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Contains, "k=dummy1\nv=yummy1\n")
		c.Assert(memlog.String(), qt.Contains, "k=dummy2\nv=yummy2\n")
		c.Assert(memlog.String(), qt.Contains, "k=dummy3\nv=yummy3\n")
	})

	t.Run("Execute_MissingExecutor", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		variables := make(map[string]any)
		err := ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "this command must be run from the executor")
	})

	t.Run("Execute_MissingVariable", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "either iterable or variable must be set")
	})

	t.Run("Execute_MissingVariable2", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.Variable = "dummy"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "variable \"dummy\" does not exist")
	})

	t.Run("Execute_MissingVariable3", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.Variable = "dummy"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		variables["dummy"] = nil
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "variable \"dummy\" does not exist")
	})

	t.Run("Execute_WrongVariableType", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.Variable = "dummy"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		variables["dummy"] = struct{}{}
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "foreach: invalid variable type \"struct\" (expected slice or map)")
	})

	t.Run("Execute_WrongVariableType2", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := executor.NewForeachCommand(&executor.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*executor.ForeachCommand)
		ex.Variable = "dummy"
		commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
		commands["message"] = executor.NewMessageCommand

		exc, err := executor.NewWithScenario(
			"",
			executor.WithStdout(os.Stdout),
			executor.WithStderr(os.Stderr),
			executor.WithFS(fs),
			executor.WithCommandTypes(commands),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		variables["dummy"] = map[int]string{1: ""}
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "foreach: invalid map key type \"int\" (expected string)")
	})
}
