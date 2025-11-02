package godexer_test

import (
	"bytes"
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
)

func TestCommandAwareError(t *testing.T) {
	c := qt.New(t)
	fs := afero.NewMemMapFs()

	cmd := godexer.NewMessageCommand(&godexer.ExecutorContext{
		Fs:     fs,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	ex := cmd.(*godexer.MessageCommand)
	ex.SetDebugInfo(&godexer.CommandDebugInfo{
		ID:       4242,
		Contents: []byte("{}"),
	})

	variables := map[string]any{
		"var1": "value1",
	}

	cmderr := godexer.NewCommandAwareError(errors.New("test error"), ex, variables)
	c.Assert(cmderr.Error(), qt.Equals, "command failed (stepName=__step_no_4242, commandId=4242, commandType=MessageCommand): test error")
	c.Assert(cmderr.Variables(), qt.DeepEquals, map[string]any{
		"var1": "value1",
	})
	c.Assert(cmderr.Command(), qt.Equals, ex)
}
