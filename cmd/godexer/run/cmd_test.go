package runcmd_test

import (
	"errors"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	runcmd "github.com/go-extras/godexer/cmd/godexer/run"
	"github.com/go-extras/godexer/cmd/godexer/shared"
)

const msgScenario = `commands:
  - type: message
    description: hello
`

const execScenario = `commands:
  - type: exec
    cmd: ["echo", "run-ok"]
`

const badTypeScenario = `commands:
  - type: unknown_type
`

const varScenario = `commands:
  - type: variable
    variable: out
    value: "{{ index . \"name\" }}"
`

// writeTempFile creates a temp file with the given content, cleaned up after the test.
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
	_ = f.Close()
	return f.Name()
}

// newRunCmd creates a fresh run command for each test.
func newRunCmd() *runcmd.Command {
	return runcmd.New(&shared.Context{})
}

func TestRunCmd_SimpleScenario(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_ExecCommand(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, execScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_VerboseFlag(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--verbose", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_VarFlag(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, varScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var", "name=world", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_VarFlagInvalidFormat(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var", "noequals", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestRunCmd_VarJSONFlag(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", `--var-json={"env":"prod"}`, f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_VarJSONInvalid(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var-json", "{not json}", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestRunCmd_VarFileFlag(t *testing.T) {
	c := qt.New(t)

	varsFile := writeTempFile(t, "name: world\nenv: prod\n")
	scenarioFile := writeTempFile(t, msgScenario)

	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var-file", varsFile, scenarioFile})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_VarFileMissing(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var-file", "/nonexistent/vars.yaml", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestRunCmd_VarFileInvalidYAML(t *testing.T) {
	c := qt.New(t)

	varsFile := writeTempFile(t, ":\n  bad: [yaml")
	scenarioFile := writeTempFile(t, msgScenario)

	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var-file", varsFile, scenarioFile})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestRunCmd_VarFromEnv(t *testing.T) {
	c := qt.New(t)

	t.Setenv("GODEXER_TEST_VAR", "testvalue")

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var-from-env", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_ParseError(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, badTypeScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 2)
}

func TestRunCmd_MissingFile(t *testing.T) {
	c := qt.New(t)

	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "/nonexistent/scenario.yaml"})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
}

func TestRunCmd_IncludeBasePath(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--include-base-path", os.TempDir(), f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_InvalidIncludeBasePath(t *testing.T) {
	// filepath.Abs never fails on valid string inputs; just confirm the flag is accepted.
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--include-base-path", ".", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_Timeout(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--timeout", "30s", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_MultipleVars(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var", "a=1", "--var", "b=2", "--var", "c=3", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}
