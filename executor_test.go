package godexer_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/go-extras/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

const runnerResultStdout = "foo!"

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecRunnerHelper", "--", command}
	cs = append(cs, args...)
	//nolint:gosec // This is a test helper that intentionally uses os.Args[0]
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestExecRunnerHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("DUMMY") != "1" {
		panic("missing dummy var")
	}
	// some code here to check arguments perhaps?
	_, _ = fmt.Fprint(os.Stdout, runnerResultStdout)
	time.Sleep(100 * time.Millisecond) // this is to make sure it's flushed by the system
	_, _ = fmt.Fprintf(os.Stderr, "%s", os.Args[len(os.Args)-1])
	time.Sleep(100 * time.Millisecond) // this is to make sure it's flushed by the system
	if os.Args[len(os.Args)-1] == "error" {
		//nolint:revive // This is a test helper that simulates process exit
		os.Exit(1)
	}
	//nolint:revive // This is a test helper that simulates process exit
	os.Exit(0)
}

type testcmd struct {
	godexer.MessageCommand
}

type testcmdfail struct {
	godexer.BaseCommand
}

func (*testcmdfail) Execute(_ map[string]any) error {
	return errors.New("test error")
}

const executorExecuteScript = `commands:
  - type: message
    stepName: step_one
    description: Test call {{ index . "var1" }}
    callsAfter: f1
  - type: message
    stepName: step_skip_one
    description: Test call skip one
    requires: 'strlen(var1) == 1'
  - type: variable
    stepName: step_two
    description: Test call two
    variable: dummy
    value: 'value: {{ index . "var2" }}'
  - type: message
    stepName: step_two_and_a_half
    description: Test call two and a half
    requires: 'dummy == "value: val2"'
  - type: sleep
    stepName: step_three
    description: Test call three
    seconds: 10
  - type: password
    stepName: step_four
    description: Test call four
    variable: pwd
    charset: abcd
  - type: writefile
    stepName: step_five
    description: Test call five
    file: /some/file
    contents: 'value: {{ index . "var2" }}'
  - type: message
    stepName: step_skip_two
    description: Test call skip two
    requires: '!file_exists("/some/file")'
  - type: exec
    stepName: step_six
    description: Test call six
    cmd: ["xxx"]
    env:
      - DUMMY=1
`

func TestExecutor(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		// init stuff to load the executor
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)
		hooksAfter["f1"] = func(variables map[string]any) error {
			variables["hookvar"] = "hookvalue"
			return nil
		}
		logger := logrus.StandardLogger()
		logout := logger.Out
		logformatter := logger.Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		// a way to mock the sleep function
		godexer.TimeSleep = func(d time.Duration) {
			c.Assert(d.Seconds(), qt.Equals, float64(10))
		}
		defer func() { godexer.TimeSleep = time.Sleep }()

		// fake exec command to avoid running the real shell commands
		godexer.ExecCommandFn = fakeExecCommand
		defer func() { godexer.ExecCommandFn = exec.Command }()

		// executor vars
		vars := map[string]any{
			"var1": "val1",
			"var2": "val2",
		}

		// ... test starts here ...

		// load executor
		exc, err := godexer.NewWithScenario(
			executorExecuteScript,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
			godexer.WithDefaultEvaluatorFunctions(),
			godexer.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)
		c.Assert(exc, qt.IsNotNil)

		err = exc.Execute(vars)
		c.Assert(err, qt.IsNil)

		c.Assert(vars["hookvar"], qt.Equals, "hookvalue", qt.Commentf("hook defined var doesn't match"))
		c.Assert(vars["dummy"], qt.Equals, "value: val2")
		c.Assert(vars["pwd"], qt.Matches, regexp.MustCompile(`^[a-d]{8}$`))
		f, err := fs.Open("/some/file")
		c.Assert(err, qt.IsNil)

		d, _ := io.ReadAll(f)
		c.Assert(string(d), qt.Equals, "value: val2")

		d, _ = io.ReadAll(stdout)
		c.Assert(string(d), qt.Equals, `Test call val1
Test call two
Test call two and a half
Test call three
Sleeping for 10 seconds
Test call four
Test call five
Test call six
Executing: xxx
foo!`)
	})

	t.Run("WithEvaluatorFunction", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)

		cmds := `
commands:
  - type: message
    stepName: step_skip_one
    description: Test call one
    requires: 'test("foo") == "bar"'
`

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		vars := make(map[string]any)

		exc, err := godexer.NewWithScenario(
			cmds,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
			godexer.WithEvaluatorFunction("test", func(args ...any) (any, error) {
				if args[0].(string) == "foo" {
					return "bar", nil
				}

				return "", nil
			}),
			godexer.WithLogger(logrus.StandardLogger()),
		)

		c.Assert(err, qt.IsNil)
		c.Assert(exc, qt.IsNotNil)

		_ = exc.Execute(vars)
		d, _ := io.ReadAll(stdout)
		c.Assert(string(d), qt.Equals, "Test call one\n")
	})

	t.Run("Execute_MissingFunction", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)

		cmds := `
commands:
  - type: message
    stepName: step_skip_one
    description: Test call skip one
    requires: 'strlen(var1) == 1'
`

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		vars := make(map[string]any)

		exc, err := godexer.NewWithScenario(
			cmds,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
		)
		c.Assert(err, qt.IsNil)
		c.Assert(exc, qt.IsNotNil)

		err = exc.Execute(vars)
		c.Assert(errors.Cause(err), qt.ErrorMatches, "Cannot transition token types from VARIABLE \\[strlen\\] to CLAUSE \\[40\\]")
	})

	t.Run("Execute_InvalidCommandType", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)

		cmds := `
commands:
  - type: brambora
    stepName: step_skip_one
    description: Test call skip one
`

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		_, err := godexer.NewWithScenario(
			cmds,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
		)
		c.Assert(err, qt.ErrorMatches, "invalid command type: \"brambora\"")
	})

	t.Run("RegisterCommand", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)

		godexer.RegisterCommand("testcmd", func(ectx *godexer.ExecutorContext) godexer.Command {
			return &testcmd{
				MessageCommand: godexer.MessageCommand{
					BaseCommand: godexer.BaseCommand{
						Ectx: ectx,
					},
				},
			}
		})

		cmds := `
commands:
  - type: testcmd
    stepName: step_skip_one
    description: Test call one
`

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		vars := make(map[string]any)

		exc, err := godexer.NewWithScenario(
			cmds,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
			godexer.WithLogger(logrus.StandardLogger()),
		)
		c.Assert(err, qt.IsNil)
		c.Assert(exc, qt.IsNotNil)

		_ = exc.Execute(vars)
		d, _ := io.ReadAll(stdout)
		c.Assert(string(d), qt.Equals, "Test call one\n")
	})

	t.Run("Execute_FailCommand", func(t *testing.T) {
		c := qt.New(t)
		fs := afero.NewMemMapFs()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		hooksAfter := make(godexer.HooksAfter)

		godexer.RegisterCommand("testcmdfail", func(ectx *godexer.ExecutorContext) godexer.Command {
			return &testcmdfail{
				BaseCommand: godexer.BaseCommand{
					Ectx: ectx,
				},
			}
		})

		cmds := `
commands:
  - type: testcmdfail
    stepName: step_skip_one
    description: Test call one
`

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		vars := make(map[string]any)

		exc, err := godexer.NewWithScenario(
			cmds,
			godexer.WithHooksAfter(hooksAfter),
			godexer.WithStdout(stdout),
			godexer.WithStderr(stderr),
			godexer.WithFS(fs),
			godexer.WithLogger(logrus.StandardLogger()),
		)
		c.Assert(err, qt.IsNil)
		c.Assert(exc, qt.IsNotNil)

		err = exc.Execute(vars)
		c.Assert(errors.Cause(err), qt.ErrorMatches, "test error")
	})

	t.Run("MaybeEvalValue", func(t *testing.T) {
		c := qt.New(t)
		// valid parsable template
		val := godexer.MaybeEvalValue(`{{ index . "foo" }}`, map[string]any{"foo": "bar"})
		c.Assert(val, qt.Equals, "bar")

		// invalid template
		val = godexer.MaybeEvalValue(`{{ index . "foo" }`, map[string]any{"foo": "bar"})
		c.Assert(val, qt.Equals, `{{ index . "foo" }`)

		// non-string value
		val = godexer.MaybeEvalValue(42, map[string]any{"foo": "bar"})
		c.Assert(val, qt.Equals, 42)
	})
}
