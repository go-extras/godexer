package executor_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
)

func TestMessageExecute(t *testing.T) {
	c := qt.New(t)
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
	c.Assert(err, qt.IsNil)
	desc := ex.GetDescription(map[string]any{
		"var1": "val1",
		"var2": "val2",
	})
	c.Assert(desc, qt.Equals, "data val1val2")
}
