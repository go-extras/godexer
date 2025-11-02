package executor_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

type SubExecuteTestSuite struct {
	suite.Suite
}

func (t *SubExecuteTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := executor.NewSubExecuteCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*executor.SubExecuteCommand)
	ex.RawCommands = []json.RawMessage{
		[]byte(`{"type": "message","stepName": "test","description": "Some kind of test"}`),
		[]byte(`{"type": "message","stepName": "test2","description": "Another kind of test"}`),
	}
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
	t.Require().NoError(err)
	ex.Ectx.Executor = exc

	variables := make(map[string]any)
	err = ex.Execute(variables)
	t.Require().NoError(err)
	t.Equal("Some kind of test\nAnother kind of test\n", memlog.String())
}

func (t *SubExecuteTestSuite) TestExecute_MissingExecutor() {
	fs := afero.NewMemMapFs()

	logger := logrus.New()
	memlog := &bytes.Buffer{}
	logger.SetOutput(memlog)
	logger.SetFormatter(&testutils.SimpleFormatter{})

	cmd := executor.NewSubExecuteCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Logger: logger,
	})
	ex := cmd.(*executor.SubExecuteCommand)
	commands := make(map[string]func(*executor.ExecutorContext) executor.Command)
	commands["message"] = executor.NewMessageCommand

	variables := make(map[string]any)
	err := ex.Execute(variables)
	t.Require().Error(err)
	t.Equal("this command must be run from the executor", err.Error())
}

func TestSubExecute(t *testing.T) {
	suite.Run(t, new(SubExecuteTestSuite))
}
