package executor_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
)

func TestInclude(t *testing.T) {
	t.Run("Execute", func(t *testing.T) {
		c := qt.New(t)
		scripts := make(fstest.MapFS)

		scripts["script/include.yaml"] = &fstest.MapFile{
			Data: []byte(`commands:
  - type: message
    stepName: test0
    description: 'Base kind of test: {{ .var0 }}'
  - type: message
    stepName: test1
    description: 'Some kind of test: {{ .var1 }}'
  - type: message
    stepName: test2
    description: 'Another kind of test: {{ .var2 }}'
  - type: variable
    stepName: test_write
    variable: test_write_var
    value: 'final write'
`),
		}

		script := `commands:
  - type: include
    file: /script/include.yaml
    variables:
      var1: value1
      var2: value2
`

		commands := executor.GetRegisteredCommands()
		commands["include"] = executor.NewIncludeCommand(scripts)

		// load the executor instance
		memfs := afero.NewMemMapFs() // will not be used, but is required by the executor
		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		exc, err := executor.NewWithScenario(
			script,
			executor.WithStdout(stdout),
			executor.WithStderr(stderr),
			executor.WithFS(memfs),
			executor.WithCommandTypes(commands),
			executor.WithDefaultEvaluatorFunctions(),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)

		variables := make(map[string]any)
		variables["var0"] = "value0"
		err = exc.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "Base kind of test: value0\nSome kind of test: value1\nAnother kind of test: value2\n")
		c.Assert(variables, qt.DeepEquals, map[string]any{
			"__step:__step_no_001:skipped": false,
			"var0":                         "value0",
			"var1":                         "value1",
			"var2":                         "value2",
			"__step:test0:skipped":         false,
			"__step:test1:skipped":         false,
			"__step:test2:skipped":         false,
			"__step:test_write:skipped":    false,
			"test_write_var":               "final write",
		})
	})

	t.Run("Execute_NoMergeVars", func(t *testing.T) {
		c := qt.New(t)
		scripts := make(fstest.MapFS)

		scripts["script/include.yaml"] = &fstest.MapFile{
			Data: []byte(`commands:
  - type: message
    stepName: test0
    description: 'Base kind of test: {{ ._parent.var0 }}'
  - type: message
    stepName: test1
    description: 'Some kind of test: {{ .var1 }}'
  - type: message
    stepName: test2
    description: 'Another kind of test: {{ .var2 }}'
  - type: variable
    stepName: test_write
    variable: test_write_var
    value: 'final write'
`),
		}
		script := `commands:
  - type: include
    stepName: include_script
    file: /script/include.yaml
    noMergeVars: true
    variables:
      var1: value1
      var2: value2
`

		commands := executor.GetRegisteredCommands()
		commands["include"] = executor.NewIncludeCommand(scripts)

		// load the executor instance
		memfs := afero.NewMemMapFs() // will not be used, but is required by the executor
		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		exc, err := executor.NewWithScenario(
			script,
			executor.WithStdout(stdout),
			executor.WithStderr(stderr),
			executor.WithFS(memfs),
			executor.WithCommandTypes(commands),
			executor.WithDefaultEvaluatorFunctions(),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)

		variables := make(map[string]any)
		variables["var0"] = "value0"
		err = exc.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "Base kind of test: value0\nSome kind of test: value1\nAnother kind of test: value2\n")
		c.Assert(variables, qt.DeepEquals, map[string]any{
			"__step:include_script:skipped": false,
			"var0":                          "value0",
			"include_script_variables": map[string]any{
				"var1":                      "value1",
				"var2":                      "value2",
				"__step:test0:skipped":      false,
				"__step:test1:skipped":      false,
				"__step:test2:skipped":      false,
				"__step:test_write:skipped": false,
				"test_write_var":            "final write",
			},
		})
	})

	t.Run("Execute_NoMergeVars_WithBasePath", func(t *testing.T) {
		c := qt.New(t)
		scripts := make(fstest.MapFS)

		scripts["script/include.yaml"] = &fstest.MapFile{
			Data: []byte(`commands:
  - type: message
    stepName: test0
    description: 'Base kind of test: {{ ._parent.var0 }}'
  - type: message
    stepName: test1
    description: 'Some kind of test: {{ .var1 }}'
  - type: message
    stepName: test2
    description: 'Another kind of test: {{ .var2 }}'
  - type: variable
    stepName: test_write
    variable: test_write_var
    value: 'final write'
`),
		}
		script := `commands:
  - type: include
    stepName: include_script
    file: include.yaml
    noMergeVars: true
    variables:
      var1: value1
      var2: value2
`

		commands := executor.GetRegisteredCommands()
		commands["include"] = executor.NewIncludeCommandWithBasePath(scripts, "script")

		// load the executor instance
		memfs := afero.NewMemMapFs() // will not be used, but is required by the executor
		logger := logrus.New()
		memlog := &bytes.Buffer{}
		logger.SetOutput(memlog)
		logger.SetFormatter(&testutils.SimpleFormatter{})

		logout := logrus.StandardLogger().Out
		logformatter := logrus.StandardLogger().Formatter
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logrus.SetOutput(stdout)
		logrus.SetFormatter(&testutils.SimpleFormatter{})
		defer func() {
			logrus.SetOutput(logout)
			logrus.SetFormatter(logformatter)
		}()

		exc, err := executor.NewWithScenario(
			script,
			executor.WithStdout(stdout),
			executor.WithStderr(stderr),
			executor.WithFS(memfs),
			executor.WithCommandTypes(commands),
			executor.WithDefaultEvaluatorFunctions(),
			executor.WithLogger(logger),
		)
		c.Assert(err, qt.IsNil)

		variables := make(map[string]any)
		variables["var0"] = "value0"
		err = exc.Execute(variables)
		c.Assert(err, qt.IsNil)
		c.Assert(memlog.String(), qt.Equals, "Base kind of test: value0\nSome kind of test: value1\nAnother kind of test: value2\n")
		c.Assert(variables, qt.DeepEquals, map[string]any{
			"__step:include_script:skipped": false,
			"var0":                          "value0",
			"include_script_variables": map[string]any{
				"var1":                      "value1",
				"var2":                      "value2",
				"__step:test0:skipped":      false,
				"__step:test1:skipped":      false,
				"__step:test2:skipped":      false,
				"__step:test_write:skipped": false,
				"test_write_var":            "final write",
			},
		})
	})
}
