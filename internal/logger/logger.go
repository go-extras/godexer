package logger

import (
	"fmt"
	"log"
	"strings"
)

type Level string

const (
	TraceLevel Level = "trace"
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(TraceLevel):
		return TraceLevel, nil
	case string(DebugLevel):
		return DebugLevel, nil
	case string(InfoLevel):
		return InfoLevel, nil
	case string(WarnLevel), "warning":
		return WarnLevel, nil
	case string(ErrorLevel):
		return ErrorLevel, nil
	default:
		return "", fmt.Errorf("invalid log level %q (expected one of: trace, debug, info, warn (warning), error)", value)
	}
}

func (l Level) Allows(message Level) bool {
	return levelOrder(message.normalized()) >= levelOrder(l.normalized())
}

func (l Level) normalized() Level {
	switch l {
	case TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		return l
	default:
		return InfoLevel
	}
}

func levelOrder(level Level) int {
	switch level {
	case TraceLevel:
		return 0
	case DebugLevel:
		return 1
	case InfoLevel:
		return 2
	case WarnLevel:
		return 3
	case ErrorLevel:
		return 4
	default:
		return 2
	}
}

type Logger struct {
	Level Level
}

func (l *Logger) enabled(level Level) bool {
	if l == nil {
		return InfoLevel.Allows(level)
	}
	return l.Level.Allows(level)
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.enabled(DebugLevel) {
		log.Printf("DEBUG: "+format, args...)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	if l.enabled(InfoLevel) {
		log.Printf("INFO: "+format, args...)
	}
}

func (l *Logger) Printf(format string, args ...any) {
	if l.enabled(InfoLevel) {
		log.Printf(format, args...)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.enabled(WarnLevel) {
		log.Printf("WARN: "+format, args...)
	}
}

func (l *Logger) Warningf(format string, args ...any) {
	if l.enabled(WarnLevel) {
		log.Printf("WARN: "+format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if l.enabled(ErrorLevel) {
		log.Printf("ERR: "+format, args...)
	}
}

func (*Logger) Fatalf(format string, args ...any) {
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Fatalf("FATAL: "+format, args...)
}

func (*Logger) Panicf(format string, args ...any) {
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Panicf(format, args...)
}

func (l *Logger) Tracef(format string, args ...any) {
	if l.enabled(TraceLevel) {
		log.Printf("TRACE: "+format, args...)
	}
}

func (l *Logger) Debug(args ...any) {
	if l.enabled(DebugLevel) {
		prefix := []any{"DEBUG: "}
		log.Print(append(prefix, args...)...)
	}
}

func (l *Logger) Info(args ...any) {
	if l.enabled(InfoLevel) {
		prefix := []any{"INFO: "}
		log.Print(append(prefix, args...)...)
	}
}

func (l *Logger) Print(args ...any) {
	if l.enabled(InfoLevel) {
		log.Print(args...)
	}
}

func (l *Logger) Warn(args ...any) {
	if l.enabled(WarnLevel) {
		prefix := []any{"WARN: "}
		log.Print(append(prefix, args...)...)
	}
}

func (l *Logger) Warning(args ...any) {
	if l.enabled(WarnLevel) {
		prefix := []any{"WARN: "}
		log.Print(append(prefix, args...)...)
	}
}

func (l *Logger) Error(args ...any) {
	if l.enabled(ErrorLevel) {
		prefix := []any{"ERR: "}
		log.Print(append(prefix, args...)...)
	}
}

func (*Logger) Fatal(args ...any) {
	prefix := []any{"FATAL: "}
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Fatal(append(prefix, args...)...)
}

func (*Logger) Panic(args ...any) {
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Panic(args...)
}

func (l *Logger) Trace(args ...any) {
	if l.enabled(TraceLevel) {
		prefix := []any{"TRACE: "}
		log.Print(append(prefix, args...)...)
	}
}
