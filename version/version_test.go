package version_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/testutils"
	"github.com/go-extras/godexer/version"
)

func TestWithDefaultEvaluatorFunctions(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testcases := []struct {
			fn string
			v1 string
			v2 string
		}{
			{
				fn: "version_gt",
				v1: "1.4",
				v2: "1.1",
			},
			{
				fn: "version_gte",
				v1: "1.4",
				v2: "1.1",
			},
			{
				fn: "version_gte",
				v1: "1.4",
				v2: "1.4",
			},
			{
				fn: "version_lt",
				v1: "1.1",
				v2: "1.4",
			},
			{
				fn: "version_lte",
				v1: "1.1",
				v2: "1.4",
			},
			{
				fn: "version_lte",
				v1: "1.4",
				v2: "1.4",
			},
			{
				fn: "version_eq",
				v1: "1.4",
				v2: "1.4",
			},
		}

		for _, tc := range testcases {
			scenario := fmt.Sprintf(`commands:
  - type: message
    stepName: test0
    description: Command must be called
    requires: '%s(v1, "%s")'
`, tc.fn, tc.v2)
			fs := afero.NewMemMapFs()

			var memlog bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&memlog)
			logger.SetFormatter(&testutils.SimpleFormatter{})

			commands := make(map[string]func(ctx *executor.ExecutorContext) executor.Command)
			commands["message"] = executor.NewMessageCommand

			c := qt.New(t)
			ex, err := executor.NewWithScenario(
				scenario,
				version.WithVersionFuncs(),
				executor.WithStdout(os.Stdout),
				executor.WithStderr(os.Stderr),
				executor.WithFS(fs),
				executor.WithCommandTypes(commands),
				executor.WithLogger(logger),
			)
			c.Assert(err, qt.IsNil)

			vars := make(map[string]any)
			vars["v1"] = tc.v1
			err = ex.Execute(vars)
			c.Assert(err, qt.IsNil)
			c.Assert(strings.TrimSpace(memlog.String()), qt.Equals, "Command must be called")
		}
	})

	t.Run("negative", func(t *testing.T) {
		testcases := []struct {
			fn string
			v1 string
			v2 string
		}{
			{
				fn: "version_gt",
				v1: "1.1",
				v2: "1.4",
			},
			{
				fn: "version_gte",
				v1: "1.1",
				v2: "1.4",
			},
			{
				fn: "version_lt",
				v1: "1.4",
				v2: "1.1",
			},
			{
				fn: "version_lte",
				v1: "1.4",
				v2: "1.1",
			},
			{
				fn: "version_eq",
				v1: "1.1",
				v2: "1.4",
			},
		}

		for _, tc := range testcases {
			scenario := fmt.Sprintf(`commands:
  - type: message
    stepName: test0
    description: Command must be called
    requires: '%s(v1, "%s")'
`, tc.fn, tc.v2)
			fs := afero.NewMemMapFs()

			var memlog bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&memlog)
			logger.SetFormatter(&testutils.SimpleFormatter{})

			commands := make(map[string]func(ctx *executor.ExecutorContext) executor.Command)
			commands["message"] = executor.NewMessageCommand

			c := qt.New(t)
			ex, err := executor.NewWithScenario(
				scenario,
				version.WithVersionFuncs(),
				executor.WithStdout(os.Stdout),
				executor.WithStderr(os.Stderr),
				executor.WithFS(fs),
				executor.WithCommandTypes(commands),
				executor.WithLogger(logger),
			)
			c.Assert(err, qt.IsNil)

			vars := make(map[string]any)
			vars["v1"] = tc.v1
			err = ex.Execute(vars)
			c.Assert(err, qt.IsNil)
			c.Assert(strings.TrimSpace(memlog.String()), qt.Not(qt.Equals), "Command must be called")
		}
	})

	t.Run("error", func(t *testing.T) {
		testcases := []struct {
			fn string
			v1 string
			v2 string
		}{
			{
				fn: "version_gt",
				v1: "invalid1",
				v2: "1.2",
			},
			{
				fn: "version_gt",
				v1: "1.1",
				v2: "invalid2",
			},
			{
				fn: "version_gt",
				v1: "invalid1",
				v2: "invalid2",
			},
		}

		for _, tc := range testcases {
			scenario := fmt.Sprintf(`commands:
  - type: message
    stepName: test0
    description: Command must be called
    requires: '%s(v1, "%s")'
`, tc.fn, tc.v2)
			fs := afero.NewMemMapFs()

			var memlog bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&memlog)
			logger.SetFormatter(&testutils.SimpleFormatter{})

			commands := make(map[string]func(ctx *executor.ExecutorContext) executor.Command)
			commands["message"] = executor.NewMessageCommand

			c := qt.New(t)
			ex, err := executor.NewWithScenario(
				scenario,
				version.WithVersionFuncs(),
				executor.WithStdout(os.Stdout),
				executor.WithStderr(os.Stderr),
				executor.WithFS(fs),
				executor.WithCommandTypes(commands),
				executor.WithLogger(logger),
			)
			c.Assert(err, qt.IsNil)

			vars := make(map[string]any)
			vars["v1"] = tc.v1
			err = ex.Execute(vars)
			c.Assert(err, qt.ErrorIs, version.ErrInvalidVersion)
		}
	})
}
