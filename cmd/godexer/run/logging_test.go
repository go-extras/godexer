package runcmd

import (
	"bytes"
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
		{name: "invalid explicit log level", explicit: "loud", explicitSet: true, wantErr: `--log-level: invalid log level "loud" \(expected one of: trace, debug, info, warn, error\)`},
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
