// Package runcmd implements the `godexer run` command.
package runcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/cmd/godexer/shared"
	internallogger "github.com/go-extras/godexer/internal/logger"
	godexerversion "github.com/go-extras/godexer/version"
)

// Command implements `godexer run`.
type Command struct {
	ctx *shared.Context
	cmd *cobra.Command

	vars            []string
	varFiles        []string
	varFromEnv      bool
	varJSON         string
	logLevel        string
	timeout         time.Duration
	quiet           bool
	verbose         bool
	includeBasePath string
}

// New creates the run command.
func New(ctx *shared.Context) *Command {
	c := &Command{ctx: ctx}
	c.cmd = &cobra.Command{
		Use:   "run <scenario>",
		Short: "Execute a godexer scenario",
		Long: `Execute a godexer scenario from a YAML or JSON file.

	Use '-' as the scenario argument to read from stdin.

	Use --log-level to choose trace, debug, info, warn (or warning), or error. When set,
	--log-level overrides the legacy -q/--quiet and -v/--verbose flags.`,
		Args: cobra.ExactArgs(1),
		RunE: c.run,
	}

	f := c.cmd.Flags()
	f.StringArrayVar(&c.vars, "var", nil, "Set a variable (name=value, repeatable)")
	f.StringArrayVar(&c.varFiles, "var-file", nil, "Load variables from a YAML/JSON file (repeatable)")
	f.BoolVar(&c.varFromEnv, "var-from-env", false, "Load variables from environment variables")
	f.StringVar(&c.varJSON, "var-json", "", "Load variables from a JSON object string")
	f.StringVar(&c.logLevel, "log-level", "", "Set log level (trace, debug, info, warn/warning, error). Overrides legacy -q/--quiet and -v/--verbose")
	f.DurationVar(&c.timeout, "timeout", 0, "Execution timeout, e.g. 30m (0 = no timeout)")
	f.BoolVarP(&c.quiet, "quiet", "q", false, "Quiet mode: suppress info output")
	f.BoolVarP(&c.verbose, "verbose", "v", false, "Verbose mode: show debug/trace output (wins over --quiet when both are set)")
	f.StringVar(&c.includeBasePath, "include-base-path", "",
		"Base path for include commands (default: directory of the scenario file)")

	return c
}

// Cmd returns the cobra command.
func (c *Command) Cmd() *cobra.Command { return c.cmd }

func (c *Command) run(cmd *cobra.Command, args []string) error {
	level, err := resolveLogLevel(
		logLevelInput{value: c.logLevel, set: cmd.Flags().Changed("log-level")},
		legacyLogFlags{quiet: c.quiet, verbose: c.verbose},
	)
	if err != nil {
		return shared.NewExitError(3, err)
	}

	scenarioPath := args[0]

	// Read scenario content and determine base directory for includes.
	content, baseDir, err := readScenario(scenarioPath)
	if err != nil {
		return shared.NewExitError(3, fmt.Errorf("failed to read scenario: %w", err))
	}

	// Override include base path if flag is set.
	if c.includeBasePath != "" {
		absPath, absErr := filepath.Abs(c.includeBasePath)
		if absErr != nil {
			return shared.NewExitError(3, fmt.Errorf("invalid --include-base-path: %w", absErr))
		}
		baseDir = absPath
	}

	// Load variables in merge order: var-file → env → var-json → --var.
	variables, err := c.loadVariables()
	if err != nil {
		return shared.NewExitError(3, err)
	}

	// Register all built-in commands plus include (rooted at baseDir).
	cmds := godexer.GetRegisteredCommands()
	cmds["include"] = godexer.NewIncludeCommandWithBasePath(newRootFS(), baseDir)

	logger := newCLILogger(level, cmd.ErrOrStderr())

	ex, err := godexer.NewWithScenario(
		string(content),
		godexer.WithCommandTypes(cmds),
		godexer.WithDefaultEvaluatorFunctions(),
		godexerversion.WithVersionFuncs(),
		godexer.WithLogger(logger),
	)
	if err != nil {
		return shared.NewExitError(2, fmt.Errorf("failed to parse scenario: %w", err))
	}

	execErr := c.execute(cmd.Context(), ex, variables)
	if execErr != nil {
		return shared.NewExitError(1, fmt.Errorf("execution failed: %w", execErr))
	}

	return nil
}

// execute runs the executor, optionally under a timeout.
func (c *Command) execute(ctx context.Context, ex *godexer.Executor, variables map[string]any) error {
	if c.timeout <= 0 {
		return ex.Execute(variables)
	}

	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- ex.Execute(variables) }()

	select {
	case err := <-done:
		return err
	case <-execCtx.Done():
		return fmt.Errorf("timed out after %s", c.timeout)
	}
}

// readScenario reads a scenario from a file path or stdin ("-").
// It returns the file contents and the absolute base directory for includes.
func readScenario(path string) (content []byte, baseDir string, err error) {
	if path == "-" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, "", err
		}
		baseDir, err = filepath.Abs(".")
		return content, baseDir, err
	}

	content, err = os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	baseDir, err = filepath.Abs(filepath.Dir(path))
	return content, baseDir, err
}

// loadVariables builds the variable map from all --var* flags in merge order.
func (c *Command) loadVariables() (map[string]any, error) {
	variables := make(map[string]any)

	// 1. --var-file files (in order given).
	for _, f := range c.varFiles {
		if err := mergeVarFile(f, variables); err != nil {
			return nil, fmt.Errorf("--var-file %q: %w", f, err)
		}
	}

	// 2. --var-from-env.
	if c.varFromEnv {
		for _, env := range os.Environ() {
			key, val, _ := strings.Cut(env, "=")
			variables[key] = val
		}
	}

	// 3. --var-json.
	if c.varJSON != "" {
		var jsonVars map[string]any
		if err := json.Unmarshal([]byte(c.varJSON), &jsonVars); err != nil {
			return nil, fmt.Errorf("--var-json: %w", err)
		}
		for k, v := range jsonVars {
			variables[k] = v
		}
	}

	// 4. --var name=value flags (in order given).
	for _, v := range c.vars {
		key, val, ok := strings.Cut(v, "=")
		if !ok {
			return nil, fmt.Errorf("--var %q: expected name=value format", v)
		}
		variables[key] = val
	}

	return variables, nil
}

// mergeVarFile reads a YAML or JSON file and merges its top-level keys into vars.
func mergeVarFile(path string, vars map[string]any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var fileVars map[string]any
	if err := yaml.Unmarshal(data, &fileVars); err != nil {
		return err
	}
	for k, v := range fileVars {
		vars[k] = v
	}
	return nil
}

// rootFS is a minimal fs.ReadFileFS rooted at the filesystem root ("/").
// It is used to satisfy godexer.NewIncludeCommandWithBasePath, which receives
// absolute paths via the basepath argument.
type rootFS struct{}

func newRootFS() fs.ReadFileFS { return &rootFS{} }

func (*rootFS) Open(name string) (fs.File, error) {
	return os.Open(string(os.PathSeparator) + name)
}

func (*rootFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(string(os.PathSeparator) + name)
}

// ---------------------------------------------------------------------------
// cliLogger implements godexer.Logger for the CLI.
// ---------------------------------------------------------------------------

type cliLogger struct {
	level  internallogger.Level
	stderr io.Writer
}

type legacyLogFlags struct {
	quiet   bool
	verbose bool
}

type logLevelInput struct {
	value string
	set   bool
}

func legacyLogLevel(flags legacyLogFlags) internallogger.Level {
	if flags.verbose {
		return internallogger.TraceLevel
	}
	if flags.quiet {
		return internallogger.WarnLevel
	}
	return internallogger.InfoLevel
}

func resolveLogLevel(explicit logLevelInput, legacy legacyLogFlags) (internallogger.Level, error) {
	if explicit.set {
		level, err := internallogger.ParseLevel(explicit.value)
		if err != nil {
			return "", fmt.Errorf("--log-level: %w", err)
		}
		return level, nil
	}

	return legacyLogLevel(legacy), nil
}

func newCLILogger(level internallogger.Level, stderr io.Writer) *cliLogger {
	if stderr == nil {
		stderr = os.Stderr
	}
	return &cliLogger{level: level, stderr: stderr}
}

func (l *cliLogger) stderrWriter() io.Writer {
	if l == nil || l.stderr == nil {
		return os.Stderr
	}
	return l.stderr
}

func (l *cliLogger) enabled(level internallogger.Level) bool {
	if l == nil {
		return internallogger.InfoLevel.Allows(level)
	}
	return l.level.Allows(level)
}

func (l *cliLogger) writef(level internallogger.Level, prefix, format string, args ...any) {
	if !l.enabled(level) {
		return
	}
	fmt.Fprintf(l.stderrWriter(), prefix+format+"\n", args...)
}

func (l *cliLogger) write(level internallogger.Level, prefix string, args ...any) {
	if !l.enabled(level) {
		return
	}
	writer := l.stderrWriter()
	fmt.Fprint(writer, prefix)
	fmt.Fprintln(writer, args...)
}

func (l *cliLogger) Debugf(format string, args ...any) {
	l.writef(internallogger.DebugLevel, "", format, args...)
}

func (l *cliLogger) Infof(format string, args ...any) {
	l.writef(internallogger.InfoLevel, "", format, args...)
}

func (l *cliLogger) Printf(format string, args ...any) {
	l.writef(internallogger.InfoLevel, "", format, args...)
}

func (l *cliLogger) Warnf(format string, args ...any) {
	l.writef(internallogger.WarnLevel, "Warning: ", format, args...)
}

func (l *cliLogger) Warningf(format string, args ...any) {
	l.writef(internallogger.WarnLevel, "Warning: ", format, args...)
}

func (l *cliLogger) Errorf(format string, args ...any) {
	l.writef(internallogger.ErrorLevel, "Error: ", format, args...)
}

func (l *cliLogger) Fatalf(format string, args ...any) {
	fmt.Fprintf(l.stderrWriter(), "Fatal: "+format+"\n", args...)
	os.Exit(1) //nolint:revive // Fatal is expected to exit
}

func (*cliLogger) Panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func (l *cliLogger) Tracef(format string, args ...any) {
	l.writef(internallogger.TraceLevel, "", format, args...)
}

func (l *cliLogger) Debug(args ...any) {
	l.write(internallogger.DebugLevel, "", args...)
}

func (l *cliLogger) Info(args ...any) {
	l.write(internallogger.InfoLevel, "", args...)
}

func (l *cliLogger) Print(args ...any) {
	l.write(internallogger.InfoLevel, "", args...)
}

func (l *cliLogger) Warn(args ...any) {
	l.write(internallogger.WarnLevel, "Warning: ", args...)
}

func (l *cliLogger) Warning(args ...any) {
	l.write(internallogger.WarnLevel, "Warning: ", args...)
}

func (l *cliLogger) Error(args ...any) {
	l.write(internallogger.ErrorLevel, "Error: ", args...)
}

func (l *cliLogger) Fatal(args ...any) {
	writer := l.stderrWriter()
	fmt.Fprint(writer, "Fatal: ")
	fmt.Fprintln(writer, args...)
	os.Exit(1) //nolint:revive // Fatal is expected to exit
}

func (*cliLogger) Panic(args ...any) {
	panic(fmt.Sprint(args...))
}

func (l *cliLogger) Trace(args ...any) {
	l.write(internallogger.TraceLevel, "", args...)
}
