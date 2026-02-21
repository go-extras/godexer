package validatecmd_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/go-extras/godexer/cmd/godexer/shared"
	validatecmd "github.com/go-extras/godexer/cmd/godexer/validate"
)

const validScenario = `commands:
  - type: message
    description: hello
`

const invalidScenario = `commands:
  - type: unknown_type
    description: oops
`

func TestValidateCmd_ValidScenario(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, validScenario)

	cmd := validatecmd.New(&shared.Context{})
	var out bytes.Buffer
	cmd.Cmd().SetOut(&out)
	cmd.Cmd().SetArgs([]string{f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
	c.Assert(out.String(), qt.Equals, "Scenario is valid.\n")
}

func TestValidateCmd_InvalidScenario(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, invalidScenario)

	cmd := validatecmd.New(&shared.Context{})
	cmd.Cmd().SetArgs([]string{f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)

	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 2)
}

func TestValidateCmd_MissingFile(t *testing.T) {
	c := qt.New(t)

	cmd := validatecmd.New(&shared.Context{})
	cmd.Cmd().SetArgs([]string{"/nonexistent/path/scenario.yaml"})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)

	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestValidateCmd_ValidJSON(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, `{"commands":[{"type":"message","description":"hello"}]}`)

	cmd := validatecmd.New(&shared.Context{})
	var out bytes.Buffer
	cmd.Cmd().SetOut(&out)
	cmd.Cmd().SetArgs([]string{f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

// writeTempFile creates a temporary file with the given content and returns its path.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "scenario*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) }) //nolint:gosec // test-controlled path
	if _, err = f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}
