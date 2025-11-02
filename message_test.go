package executor_test

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/go-extras/godexer"
)

type MessageTestSuite struct {
	suite.Suite
}

func (t *MessageTestSuite) TestExecute() {
	fs := afero.NewMemMapFs()

	cmd := executor.NewMessageCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*executor.MessageCommand)
	ex.Description = "data {{ index .  \"var1\" }}{{ index .  \"var2\" }}"

	err := ex.Execute(map[string]any{
		"var1": "val1",
		"var2": "val2",
	})
	t.NoError(err)
	desc := ex.GetDescription(map[string]any{
		"var1": "val1",
		"var2": "val2",
	})
	t.Equal("data val1val2", desc)
}

func TestMessage(t *testing.T) {
	suite.Run(t, new(MessageTestSuite))
}
