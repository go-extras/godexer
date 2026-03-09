package godexer

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"
	"github.com/sirupsen/logrus"

	"github.com/go-extras/godexer/internal/testutils"
)

type experimentProbeCommand struct {
	BaseCommand
	Variable   string `json:"variable"`
	Experiment string `json:"experiment"`
}

func (c *experimentProbeCommand) Execute(variables map[string]any) error {
	variables[c.Variable] = c.Ectx.Executor.experimentEnabled(c.Experiment)
	return nil
}

func newExperimentProbeCommand(ectx *ExecutorContext) Command {
	return &experimentProbeCommand{BaseCommand: BaseCommand{Ectx: ectx}}
}

func newScenarioMetadataLogger() (*logrus.Logger, *bytes.Buffer) {
	logger := logrus.New()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	logger.SetFormatter(&testutils.SimpleFormatter{})
	return logger, buf
}

func TestScenarioMetadataParsing(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		wantExpr bool
		wantCmds int
	}{
		{
			name: "yaml with meta experiments",
			scenario: `meta:
  experiments:
    - expr
commands:
  - type: message
    description: hello
`,
			wantExpr: true,
			wantCmds: 1,
		},
		{
			name:     "json with meta experiments",
			scenario: `{"meta":{"experiments":["expr"]},"commands":[{"type":"message","description":"hello"}]}`,
			wantExpr: true,
			wantCmds: 1,
		},
		{
			name: "backward compatible without meta",
			scenario: `commands:
  - type: message
    description: hello
`,
			wantExpr: false,
			wantCmds: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)

			ex, err := NewWithScenario(tt.scenario, WithCommandTypes(map[string]func(*ExecutorContext) Command{
				"message": NewMessageCommand,
			}))

			c.Assert(err, qt.IsNil)
			c.Assert(ex.experimentEnabled(experimentExpr), qt.Equals, tt.wantExpr)
			c.Assert(len(ex.commands), qt.Equals, tt.wantCmds)
		})
	}
}

func TestScenarioMetadataWarnsOnUnknownExperiments(t *testing.T) {
	c := qt.New(t)
	logger, logbuf := newScenarioMetadataLogger()

	ex, err := NewWithScenario(`meta:
  experiments:
    - expr
    - typo
    - "-future"
commands: []
`, WithLogger(logger))

	c.Assert(err, qt.IsNil)
	c.Assert(ex.experimentEnabled(experimentExpr), qt.Equals, true)
	c.Assert(strings.Contains(logbuf.String(), `unknown experiment "typo"`), qt.IsTrue)
	c.Assert(strings.Contains(logbuf.String(), `unknown experiment "future"`), qt.IsTrue)
}

func TestScenarioMetadataWithScenarioInheritance(t *testing.T) {
	t.Run("inherits enabled experiment by default", func(t *testing.T) {
		c := qt.New(t)

		parent, err := NewWithScenario(`meta:
  experiments:
    - expr
commands: []
`)
		c.Assert(err, qt.IsNil)

		child, err := parent.WithScenario(`commands: []`)
		c.Assert(err, qt.IsNil)
		c.Assert(child.experimentEnabled(experimentExpr), qt.Equals, true)
	})

	t.Run("child can explicitly opt out", func(t *testing.T) {
		c := qt.New(t)

		parent, err := NewWithScenario(`meta:
  experiments:
    - expr
commands: []
`)
		c.Assert(err, qt.IsNil)

		child, err := parent.WithScenario(`meta:
  experiments:
    - "-expr"
commands: []
`)
		c.Assert(err, qt.IsNil)
		c.Assert(child.experimentEnabled(experimentExpr), qt.Equals, false)
	})

	t.Run("child can enable experiment when parent does not", func(t *testing.T) {
		c := qt.New(t)

		parent, err := NewWithScenario(`commands: []`)
		c.Assert(err, qt.IsNil)

		child, err := parent.WithScenario(`meta:
  experiments:
    - expr
commands: []
`)
		c.Assert(err, qt.IsNil)
		c.Assert(child.experimentEnabled(experimentExpr), qt.Equals, true)
	})
}

func TestScenarioMetadataSubExecuteInheritanceAndOverride(t *testing.T) {
	c := qt.New(t)
	logger, _ := newScenarioMetadataLogger()
	commands := map[string]func(*ExecutorContext) Command{
		"commands":         NewSubExecuteCommand,
		"experiment_probe": newExperimentProbeCommand,
	}

	ex, err := NewWithScenario(`meta:
  experiments:
    - expr
commands:
  - type: commands
    commands:
      - type: experiment_probe
        variable: inherited_expr
        experiment: expr
  - type: commands
    meta:
      experiments:
        - "-expr"
    commands:
      - type: experiment_probe
        variable: opted_out_expr
        experiment: expr
`, WithCommandTypes(commands), WithLogger(logger))

	c.Assert(err, qt.IsNil)

	variables := make(map[string]any)
	err = ex.Execute(variables)

	c.Assert(err, qt.IsNil)
	c.Assert(variables["inherited_expr"], qt.Equals, true)
	c.Assert(variables["opted_out_expr"], qt.Equals, false)
}

func TestScenarioMetadataIncludeOverride(t *testing.T) {
	c := qt.New(t)
	logger, _ := newScenarioMetadataLogger()
	scripts := fstest.MapFS{
		"script/child.yaml": &fstest.MapFile{Data: []byte(`meta:
  experiments:
    - "-expr"
commands:
  - type: experiment_probe
    variable: child_expr
    experiment: expr
`)},
	}
	commands := map[string]func(*ExecutorContext) Command{
		"include":          NewIncludeCommand(scripts),
		"experiment_probe": newExperimentProbeCommand,
	}

	ex, err := NewWithScenario(`meta:
  experiments:
    - expr
commands:
  - type: include
    file: /script/child.yaml
`, WithCommandTypes(commands), WithLogger(logger))

	c.Assert(err, qt.IsNil)

	variables := make(map[string]any)
	err = ex.Execute(variables)

	c.Assert(err, qt.IsNil)
	c.Assert(variables["child_expr"], qt.Equals, false)
}

func TestScenarioMetadataRequiresEvaluationInheritanceAndOverride(t *testing.T) {
	c := qt.New(t)
	logger, _ := newScenarioMetadataLogger()

	ex, err := NewWithScenario(`meta:
  experiments:
    - expr
commands:
  - type: commands
    commands:
      - type: variable
        variable: inherited_expr
        value: inherited
        requires: 'name matches "^wor"'
  - type: commands
    meta:
      experiments:
        - "-expr"
    commands:
      - type: variable
        variable: opted_out_expr
        value: legacy
        requires: 'name =~ "^wor"'
`, WithLogger(logger))

	c.Assert(err, qt.IsNil)

	variables := map[string]any{"name": "world"}
	err = ex.Execute(variables)

	c.Assert(err, qt.IsNil)
	c.Assert(variables["inherited_expr"], qt.Equals, "inherited")
	c.Assert(variables["opted_out_expr"], qt.Equals, "legacy")
}

func TestScenarioMetadataWithCommandsPreservesExperimentsAndEvaluators(t *testing.T) {
	c := qt.New(t)
	logger, _ := newScenarioMetadataLogger()

	parent, err := NewWithScenario(`meta:
  experiments:
    - expr
commands:
  - type: variable
    variable: cloned_result
    value: ok
    requires: 'echo("world") matches "^wor"'
`, WithLogger(logger), WithRegisteredEvaluatorFunction("echo", func(args ...any) (any, error) {
		return args[0], nil
	}))

	c.Assert(err, qt.IsNil)

	child := parent.WithCommands(parent.commands)
	variables := make(map[string]any)
	err = child.Execute(variables)

	c.Assert(child.experimentEnabled(experimentExpr), qt.Equals, true)
	c.Assert(err, qt.IsNil)
	c.Assert(variables["cloned_result"], qt.Equals, "ok")
}
