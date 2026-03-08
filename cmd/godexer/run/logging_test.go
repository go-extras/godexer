package runcmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	internallogger "github.com/go-extras/godexer/internal/logger"
)

func TestLegacyLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		quiet   bool
		verbose bool
		want    internallogger.Level
	}{
		{name: "default", want: internallogger.InfoLevel},
		{name: "quiet", quiet: true, want: internallogger.WarnLevel},
		{name: "verbose", verbose: true, want: internallogger.TraceLevel},
		{name: "verbose wins when both set", quiet: true, verbose: true, want: internallogger.TraceLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)

			got := legacyLogLevel(legacyLogFlags{quiet: tt.quiet, verbose: tt.verbose})

			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestResolveLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		explicit    string
		explicitSet bool
		quiet       bool
		verbose     bool
		want        internallogger.Level
		wantErr     string
	}{
		{name: "default", want: internallogger.InfoLevel},
		{name: "quiet", quiet: true, want: internallogger.WarnLevel},
		{name: "verbose", verbose: true, want: internallogger.TraceLevel},
		{name: "verbose wins when both legacy flags are set", quiet: true, verbose: true, want: internallogger.TraceLevel},
		{name: "warning alias", explicit: "warning", explicitSet: true, want: internallogger.WarnLevel},
		{name: "explicit log level overrides legacy flags", explicit: "error", explicitSet: true, quiet: true, verbose: true, want: internallogger.ErrorLevel},
		{name: "invalid explicit log level", explicit: "loud", explicitSet: true, wantErr: `--log-level: invalid log level "loud" \(expected one of: trace, debug, info, warn \(warning\), error\)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)

			got, err := resolveLogLevel(
				logLevelInput{value: tt.explicit, set: tt.explicitSet},
				legacyLogFlags{quiet: tt.quiet, verbose: tt.verbose},
			)

			if tt.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tt.wantErr)
				return
			}

			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestCLILoggerFiltersByLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    internallogger.Level
		contains []string
		omits    []string
	}{
		{
			name:     "warn suppresses info and debug",
			level:    internallogger.WarnLevel,
			contains: []string{"Warning: warn", "Error: err"},
			omits:    []string{"info", "debug", "trace"},
		},
		{
			name:     "trace includes all CLI levels",
			level:    internallogger.TraceLevel,
			contains: []string{"trace", "debug", "info", "Warning: warn", "Error: err"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			var stderr bytes.Buffer

			logger := newCLILogger(tt.level, &stderr)
			logger.Trace("trace")
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("err")

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

func TestCLILoggerFatalUsesConfiguredStderr(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{name: "Fatal", method: "Fatal", want: "Fatal: boom\n"},
		{name: "Fatalf", method: "Fatalf", want: "Fatal: boom 7\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			stderrPath := filepath.Join(t.TempDir(), "stderr.txt")

			//nolint:gosec // Test helper intentionally re-executes the current test binary.
			cmd := exec.Command(os.Args[0], "-test.run=TestCLILoggerFatalHelper")
			cmd.Env = append(os.Environ(),
				"GO_WANT_CLI_LOGGER_FATAL_HELPER=1",
				"CLI_LOGGER_FATAL_FILE="+stderrPath,
				"CLI_LOGGER_FATAL_METHOD="+tt.method,
				"GOCOVERDIR="+t.TempDir(),
			)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			var exitErr *exec.ExitError
			c.Assert(errors.As(err, &exitErr), qt.IsTrue)
			c.Assert(exitErr.ExitCode(), qt.Equals, 1)

			data, readErr := os.ReadFile(stderrPath)
			c.Assert(readErr, qt.IsNil)
			c.Assert(string(data), qt.Equals, tt.want)
			c.Assert(stderr.String(), qt.Equals, "")
		})
	}
}

func TestCLILoggerFatalHelper(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_LOGGER_FATAL_HELPER") != "1" {
		return
	}

	//nolint:gosec // Test controls this temp file path when launching the helper subprocess.
	stderrFile, err := os.Create(os.Getenv("CLI_LOGGER_FATAL_FILE"))
	if err != nil {
		t.Fatalf("failed to create stderr file: %v", err)
	}

	logger := newCLILogger(internallogger.InfoLevel, stderrFile)
	switch os.Getenv("CLI_LOGGER_FATAL_METHOD") {
	case "Fatal":
		logger.Fatal("boom")
	case "Fatalf":
		logger.Fatalf("boom %d", 7)
	default:
		t.Fatalf("unknown fatal method %q", os.Getenv("CLI_LOGGER_FATAL_METHOD"))
	}
}
