package runcmd_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
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

func TestRunCmd_HelpIncludesLogLevelFlag(t *testing.T) {
	c := qt.New(t)
	cmd := newRunCmd()
	var out bytes.Buffer
	cmd.Cmd().SetOut(&out)
	cmd.Cmd().SetErr(&out)
	cmd.Cmd().SetArgs([]string{"--help"})

	err := cmd.Cmd().Execute()

	c.Assert(err, qt.IsNil)
	help := out.String()
	c.Assert(strings.Contains(help, "--log-level string"), qt.IsTrue)
	c.Assert(strings.Contains(help, "warn/warning"), qt.IsTrue)
	c.Assert(strings.Contains(help, "Overrides legacy -q/--quiet and -v/--verbose"), qt.IsTrue)
	c.Assert(strings.Contains(help, "wins over --quiet when both are set"), qt.IsTrue)
}

func TestRunCmd_LogLevelControlsRuntimeLogger(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
		omits    []string
	}{
		{
			name:     "explicit info logs scenario descriptions",
			args:     []string{"--log-level", "info"},
			contains: []string{"hello"},
		},
		{
			name:  "explicit error suppresses info logs",
			args:  []string{"--log-level", "error"},
			omits: []string{"hello"},
		},
		{
			name:  "warning alias suppresses info logs",
			args:  []string{"--log-level", "warning"},
			omits: []string{"hello"},
		},
		{
			name:  "explicit log level overrides verbose compatibility flag",
			args:  []string{"--log-level", "warn", "--verbose"},
			omits: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			f := writeTempFile(t, msgScenario)
			cmd := newRunCmd()
			var stderr bytes.Buffer
			cmd.Cmd().SetErr(&stderr)
			cmd.Cmd().SetArgs(append(tt.args, f))

			err := cmd.Cmd().Execute()

			c.Assert(err, qt.IsNil)
			output := stderr.String()
			for _, expected := range tt.contains {
				c.Assert(strings.Contains(output, expected), qt.IsTrue)
			}
			for _, omitted := range tt.omits {
				c.Assert(strings.Contains(output, omitted), qt.IsFalse)
			}
		})
	}
}

func TestRunCmd_VarFlag(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, varScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--quiet", "--var", "name=world", f})

	err := cmd.Cmd().Execute()
	c.Assert(err, qt.IsNil)
}

func TestRunCmd_InvalidLogLevel(t *testing.T) {
	c := qt.New(t)

	f := writeTempFile(t, msgScenario)
	cmd := newRunCmd()
	cmd.Cmd().SetArgs([]string{"--log-level", "loud", f})

	err := cmd.Cmd().Execute()

	c.Assert(err, qt.IsNotNil)
	var exitErr *shared.ExitError
	c.Assert(errors.As(err, &exitErr), qt.IsTrue)
	c.Assert(exitErr.Code, qt.Equals, 3)
	c.Assert(exitErr.Error(), qt.Matches, `--log-level: invalid log level "loud" \(expected one of: trace, debug, info, warn \(warning\), error\)`)
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
