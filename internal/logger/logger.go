package logger

import (
	"log"
)

type Logger struct{}

func (*Logger) Debugf(format string, args ...any) {
	log.Printf("DEBUG: "+format, args...)
}

func (*Logger) Infof(format string, args ...any) {
	log.Printf("INFO: "+format, args...)
}

func (*Logger) Printf(format string, args ...any) {
	log.Printf(format, args...)
}

func (*Logger) Warnf(format string, args ...any) {
	log.Printf("WARN: "+format, args...)
}

func (*Logger) Warningf(format string, args ...any) {
	log.Printf("WARN: "+format, args...)
}

func (*Logger) Errorf(format string, args ...any) {
	log.Printf("ERR: "+format, args...)
}

func (*Logger) Fatalf(format string, args ...any) {
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Fatalf("FATAL: "+format, args...)
}

func (*Logger) Panicf(format string, args ...any) {
	//nolint:revive // Fatal/Panic methods are part of the logger interface
	log.Panicf(format, args...)
}

func (*Logger) Tracef(format string, args ...any) {
	log.Printf("TRACE: "+format, args...)
}

func (*Logger) Debug(args ...any) {
	prefix := []any{"DEBUG: "}
	log.Print(append(prefix, args...)...)
}

func (*Logger) Info(args ...any) {
	prefix := []any{"INFO: "}
	log.Print(append(prefix, args...)...)
}

func (*Logger) Print(args ...any) {
	log.Print(args...)
}

func (*Logger) Warn(args ...any) {
	prefix := []any{"WARN: "}
	log.Print(append(prefix, args...)...)
}

func (*Logger) Warning(args ...any) {
	prefix := []any{"WARN: "}
	log.Print(append(prefix, args...)...)
}

func (*Logger) Error(args ...any) {
	prefix := []any{"ERR: "}
	log.Print(append(prefix, args...)...)
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

func (*Logger) Trace(args ...any) {
	prefix := []any{"TRACE: "}
	log.Print(append(prefix, args...)...)
}
