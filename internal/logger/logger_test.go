package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Level
	}{
		{name: "trace", input: "trace", want: TraceLevel},
		{name: "debug uppercase", input: "DEBUG", want: DebugLevel},
		{name: "warn alias", input: "warning", want: WarnLevel},
		{name: "trimmed error", input: "  error ", want: ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)

			got, err := ParseLevel(tt.input)

			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestParseLevelInvalid(t *testing.T) {
	c := qt.New(t)

	_, err := ParseLevel("loud")

	c.Assert(err, qt.ErrorMatches, `invalid log level "loud" \(expected one of: trace, debug, info, warn, error\)`)
}

func TestLoggerFiltersByLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		contains []string
		omits    []string
	}{
		{
			name:     "default info suppresses debug and trace",
			level:    "",
			contains: []string{"INFO: info", "WARN: warn", "ERR: err"},
			omits:    []string{"DEBUG: debug", "TRACE: trace"},
		},
		{
			name:     "warn suppresses info",
			level:    WarnLevel,
			contains: []string{"WARN: warn", "ERR: err"},
			omits:    []string{"INFO: info", "DEBUG: debug", "TRACE: trace"},
		},
		{
			name:     "trace includes everything",
			level:    TraceLevel,
			contains: []string{"TRACE: trace", "DEBUG: debug", "INFO: info", "WARN: warn", "ERR: err"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			var buf bytes.Buffer

			origWriter := log.Writer()
			origFlags := log.Flags()
			origPrefix := log.Prefix()
			log.SetOutput(&buf)
			log.SetFlags(0)
			log.SetPrefix("")
			defer func() {
				log.SetOutput(origWriter)
				log.SetFlags(origFlags)
				log.SetPrefix(origPrefix)
			}()

			logger := &Logger{Level: tt.level}
			logger.Trace("trace")
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("err")

			output := buf.String()
			for _, expected := range tt.contains {
				c.Assert(strings.Contains(output, expected), qt.IsTrue)
			}
			for _, omitted := range tt.omits {
				c.Assert(strings.Contains(output, omitted), qt.IsFalse)
			}
		})
	}
}
