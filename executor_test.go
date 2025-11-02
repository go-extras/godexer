package executor_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/go-extras/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	. "github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

const runnerResultStdout = "foo!"

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecRunnerHelper", "--", command}
	cs = append(cs, args...)
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
		os.Exit(1)
	}
	os.Exit(0)
}

type ExecutorTestSuite struct {
	suite.Suite
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

func (t *ExecutorTestSuite) TestExecute() {
	// init stuff to load the executor
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)
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
	TimeSleep = func(d time.Duration) {
		t.Equal(float64(10), d.Seconds())
	}
	defer func() { TimeSleep = time.Sleep }()

	// fake exec command to avoid running the real shell commands
	ExecCommandFn = fakeExecCommand
	defer func() { ExecCommandFn = exec.Command }()

	// executor vars
	vars := map[string]any{
		"var1": "val1",
		"var2": "val2",
	}

	// ... test starts here ...

	// load executor
	exc, err := NewWithScenario(
		executorExecuteScript,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
		WithDefaultEvaluatorFunctions(),
		WithLogger(logger),
	)
	t.Require().NoError(err)
	t.Require().NotNil(exc)

	err = exc.Execute(vars)
	t.Require().NoError(err)

	t.Equal("hookvalue", vars["hookvar"], "hook defined var doesn't match")
	t.Equal("value: val2", vars["dummy"])
	t.Regexp(regexp.MustCompile(`^[a-d]{8}$`), vars["pwd"])
	f, err := fs.Open("/some/file")
	t.Require().NoError(err)

	d, _ := io.ReadAll(f)
	t.Equal("value: val2", string(d))

	d, _ = io.ReadAll(stdout)
	t.Equal(`Test call val1
Test call two
Test call two and a half
Test call three
Sleeping for 10 seconds
Test call four
Test call five
Test call six
Executing: xxx
foo!`, string(d))
}

func (t *ExecutorTestSuite) TestWithEvaluatorFunction() {
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)

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

	vars := map[string]any{}

	exc, err := NewWithScenario(
		cmds,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
		WithEvaluatorFunction("test", func(args ...any) (any, error) {
			if args[0].(string) == "foo" {
				return "bar", nil
			}

			return "", nil
		}),
		WithLogger(logrus.StandardLogger()),
	)

	t.Require().NoError(err)
	t.Require().NotNil(exc)

	err = exc.Execute(vars)
	d, _ := io.ReadAll(stdout)
	t.Equal("Test call one\n", string(d))
}

func (t *ExecutorTestSuite) TestExecute_MissingFunction() {
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)

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

	vars := map[string]any{}

	exc, err := NewWithScenario(
		cmds,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
	)
	t.Require().NoError(err)
	t.Require().NotNil(exc)

	err = exc.Execute(vars)
	t.EqualError(errors.Cause(err), "Cannot transition token types from VARIABLE [strlen] to CLAUSE [40]")
}

func (t *ExecutorTestSuite) TestExecute_InvalidCommandType() {
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)

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

	_, err := NewWithScenario(
		cmds,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
	)
	t.EqualError(err, "invalid command type: \"brambora\"")
}

type testcmd struct {
	MessageCommand
}

func (t *ExecutorTestSuite) TestRegisterCommand() {
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)

	RegisterCommand("testcmd", func(ectx *ExecutorContext) Command {
		return &testcmd{}
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

	vars := map[string]any{}

	exc, err := NewWithScenario(
		cmds,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
		WithLogger(logrus.StandardLogger()),
	)
	t.Require().NoError(err)
	t.Require().NotNil(exc)

	err = exc.Execute(vars)
	d, _ := io.ReadAll(stdout)
	t.Equal("Test call one\n", string(d))
}

type testcmdfail struct {
	MessageCommand
}

func (c *testcmdfail) Execute(_ map[string]any) error {
	return errors.New("test error")
}

func (t *ExecutorTestSuite) TestExecute_FailCommand() {
	fs := afero.NewMemMapFs()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	hooksAfter := make(HooksAfter)

	RegisterCommand("testcmdfail", func(ectx *ExecutorContext) Command {
		return &testcmdfail{}
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

	vars := map[string]any{}

	exc, err := NewWithScenario(
		cmds,
		WithHooksAfter(hooksAfter),
		WithStdout(stdout),
		WithStderr(stderr),
		WithFS(fs),
		WithLogger(logrus.StandardLogger()),
	)
	t.Require().NoError(err)
	t.Require().NotNil(exc)

	err = exc.Execute(vars)
	t.EqualError(errors.Cause(err), "test error")
}

func (t *ExecutorTestSuite) TestMaybeEvalValue() {
	// valid parsable template
	val := MaybeEvalValue(`{{ index . "foo" }}`, map[string]any{"foo": "bar"})
	t.Equal("bar", val)

	// invalid template
	val = MaybeEvalValue(`{{ index . "foo" }`, map[string]any{"foo": "bar"})
	t.Equal(`{{ index . "foo" }`, val)

	// non-string value
	val = MaybeEvalValue(42, map[string]any{"foo": "bar"})
	t.Equal(42, val)
}

func TestExecutor(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
