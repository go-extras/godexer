package executor_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/go-extras/godexer"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func (t *ErrorsTestSuite) TestCommandAwareError() {
	fs := afero.NewMemMapFs()

	cmd := executor.NewMessageCommand(&executor.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*executor.MessageCommand)
	ex.SetDebugInfo(&executor.CommandDebugInfo{
		ID:       4242,
		Contents: []byte("{}"),
	})

	variables := map[string]any{
		"var1": "value1",
	}

	cmderr := executor.NewCommandAwareError(errors.New("test error"), ex, variables)
	t.Equal("command failed (stepName=__step_no_4242, commandId=4242, commandType=MessageCommand): test error", cmderr.Error())
	t.Equal(map[string]any{
		"var1": "value1",
	}, cmderr.Variables())
	t.Exactly(ex, cmderr.Command())
}

func TestErrors(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}
