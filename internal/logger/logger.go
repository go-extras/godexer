package logger

import (
	"log"
)

type Logger struct{}

func (l *Logger) Debugf(format string, args ...any) {
	log.Printf("DEBUG: "+format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	log.Printf("INFO: "+format, args...)
}

func (l *Logger) Printf(format string, args ...any) {
	log.Printf(format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	log.Printf("WARN: "+format, args...)
}

func (l *Logger) Warningf(format string, args ...any) {
	log.Printf("WARN: "+format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	log.Printf("ERR: "+format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	log.Fatalf("FATAL: "+format, args...)
}

func (l *Logger) Panicf(format string, args ...any) {
	log.Panicf(format, args...)
}

func (l *Logger) Tracef(format string, args ...any) {
	log.Printf("TRACE: "+format, args...)
}

func (l *Logger) Debug(args ...any) {
	prefix := []any{"DEBUG: "}
	log.Print(append(prefix, args...)...)
}

func (l *Logger) Info(args ...any) {
	prefix := []any{"INFO: "}
	log.Print(append(prefix, args...)...)
}

func (l *Logger) Print(args ...any) {
	log.Print(args...)
}

func (l *Logger) Warn(args ...any) {
	prefix := []any{"WARN: "}
	log.Print(append(prefix, args...)...)
}

func (l *Logger) Warning(args ...any) {
	prefix := []any{"WARN: "}
	log.Print(append(prefix, args...)...)
}

func (l *Logger) Error(args ...any) {
	prefix := []any{"ERR: "}
	log.Print(append(prefix, args...)...)
}

func (l *Logger) Fatal(args ...any) {
	prefix := []any{"FATAL: "}
	log.Fatal(append(prefix, args...)...)
}

func (l *Logger) Panic(args ...any) {
	log.Panic(args...)
}

func (l *Logger) Trace(args ...any) {
	prefix := []any{"TRACE: "}
	log.Print(append(prefix, args...)...)
}
