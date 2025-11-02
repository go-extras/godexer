package godexer_test

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

func TestSubExecute(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := godexer.NewSubExecuteCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*godexer.SubExecuteCommand)
		ex.RawCommands = []json.RawMessage{
			[]byte(`{"type": "message","stepName": "test","description": "Some kind of test"}`),
			[]byte(`{"type": "message","stepName": "test2","description": "Another kind of test"}`),
		}
		commands := make(map[string]func(*godexer.ExecutorContext) godexer.Command)
		commands["message"] = godexer.NewMessageCommand

		exc, err := godexer.NewWithScenario(
			"",
			godexer.WithStdout(os.Stdout),
			godexer.WithStderr(os.Stderr),
			godexer.WithFS(fs),
			godexer.WithCommandTypes(commands),
			godexer.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		ex.Ectx.Executor = exc

		variables := make(map[string]any)
		err = ex.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "Some kind of test\nAnother kind of test\n")
	})

	t.Run("Execute_MissingExecutor", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()

		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		cmd := godexer.NewSubExecuteCommand(&godexer.ExecutorContext{
			Fs:     fs,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
			Logger: logger,
		})
		ex := cmd.(*godexer.SubExecuteCommand)
		commands := make(map[string]func(*godexer.ExecutorContext) godexer.Command)
		commands["message"] = godexer.NewMessageCommand

		variables := make(map[string]any)
		err := ex.Execute(variables)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err.Error(), qt.Equals, "this command must be run from the executor")
	})
}
