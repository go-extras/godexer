package godexer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/go-extras/errors"
	"github.com/spf13/afero"
	"gopkg.in/Knetic/govaluate.v2"

	"github.com/go-extras/godexer/internal/logger"
)

type BeforeCommandExecuteCallback func(command Command, variables map[string]any)

type HooksAfter map[string]func(variables map[string]any) error

var registeredCommands = make(map[string]func(ectx *ExecutorContext) Command)

var registeredValueFuncs = make(map[string]any)

func MaybeEvalValue(val any, variables map[string]any) any {
	// we can only eval strings
	v1, ok := val.(string)
	if !ok {
		return val
	}

	fnMap := template.FuncMap{
		"shell_escape": ShellEscape,
	}
	for k, v := range registeredValueFuncs {
		fnMap[k] = v
	}

	// check if the value is a valid template
	tmpl, err := template.
		New("tpl").
		Funcs(fnMap).
		Parse(v1)
	if err != nil {
		return val
	}

	// execute
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		return val
	}

	return buf.String()
}

// RegisterValueFunc registers template value functions.
// Not safe for concurrent usage.
func RegisterValueFunc(name string, fn any) {
	registeredValueFuncs[name] = fn
}

// UnregisterValueFunc unregisters template value functions.
// Not safe for concurrent usage.
func UnregisterValueFunc(name string) {
	delete(registeredValueFuncs, name)
}

func RegisterCommand(name string, cmd func(ectx *ExecutorContext) Command) {
	registeredCommands[name] = cmd
}

func GetRegisteredCommands() map[string]func(ectx *ExecutorContext) Command {
	m := make(map[string]func(ectx *ExecutorContext) Command)
	for k, v := range registeredCommands {
		m[k] = v
	}
	return m
}

type CommandDebugInfo struct {
	ID       int
	Contents []byte
}

type Command interface {
	GetRequires() string
	GetStepName() string
	GetHookAfter() string
	GetDescription(variables map[string]any) string
	Execute(variables map[string]any) error
}

type DebugInfoer interface {
	DebugInfo() *CommandDebugInfo
	SetDebugInfo(*CommandDebugInfo)
}

type BaseCommand struct {
	Type        string
	StepName    string
	Description string
	Requires    string
	CallsAfter  string
	Ectx        *ExecutorContext

	debugInfo *CommandDebugInfo
}

func (r *BaseCommand) DebugInfo() *CommandDebugInfo {
	return r.debugInfo
}

func (r *BaseCommand) SetDebugInfo(di *CommandDebugInfo) {
	r.debugInfo = di
}

func (r *BaseCommand) GetRequires() string {
	return r.Requires
}

func (r *BaseCommand) GetStepName() string {
	if r.StepName == "" && r.debugInfo != nil && r.debugInfo.ID > 0 {
		return fmt.Sprintf("__step_no_%03d", r.debugInfo.ID)
	}

	return r.StepName
}

func (r *BaseCommand) GetHookAfter() string {
	return r.CallsAfter
}

func (r *BaseCommand) GetDescription(variables map[string]any) string {
	if desc, ok := MaybeEvalValue(r.Description, variables).(string); ok {
		return desc
	}
	return ""
}

type ExecutorContext struct {
	Fs       afero.Fs
	Stdout   io.Writer
	Stderr   io.Writer
	Executor *Executor
	Logger   Logger
}

type RawScenario struct {
	Commands []json.RawMessage `json:"commands"`
}

type Executor struct {
	ectx                         *ExecutorContext
	commands                     []Command
	hooksAfter                   HooksAfter
	evaluatorFunctions           map[string]govaluate.ExpressionFunction
	commandTypes                 map[string]func(ectx *ExecutorContext) Command
	beforeCommandExecuteCallback BeforeCommandExecuteCallback
	stepNameSuffix               string
}

type Option func(*Executor)

func New(opts ...Option) *Executor {
	ectx := &ExecutorContext{
		Fs:     afero.NewOsFs(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Logger: &logger.Logger{},
	}
	ex := &Executor{
		ectx:                         ectx,
		evaluatorFunctions:           make(map[string]govaluate.ExpressionFunction),
		beforeCommandExecuteCallback: func(Command, map[string]any) {},
		hooksAfter:                   make(HooksAfter),
		commandTypes:                 registeredCommands,
	}
	ex.ectx.Executor = ex
	for _, opt := range opts {
		opt(ex)
	}
	return ex
}

func NewWithScenario(scenario string, opts ...Option) (*Executor, error) {
	ex := New(opts...)
	err := ex.AppendScenario(scenario)
	if err != nil {
		return nil, err
	}
	return ex, nil
}

func WithCommands(cmds []Command) func(ex *Executor) {
	return func(ex *Executor) {
		ex.commands = cmds
	}
}

func WithStepNameSuffix(suffix string) func(ex *Executor) {
	return func(ex *Executor) {
		ex.stepNameSuffix = suffix
	}
}

func WithHookAfter(name string, hook func(variables map[string]any) error) func(ex *Executor) {
	return func(ex *Executor) {
		ex.hooksAfter[name] = hook
	}
}

func WithHooksAfter(hooks HooksAfter) func(ex *Executor) {
	return func(ex *Executor) {
		ex.hooksAfter = hooks
	}
}

func WithStdout(stdout io.Writer) func(ex *Executor) {
	return func(ex *Executor) {
		ex.ectx.Stdout = stdout
	}
}

func WithStderr(stderr io.Writer) func(ex *Executor) {
	return func(ex *Executor) {
		ex.ectx.Stderr = stderr
	}
}

func WithFS(fs afero.Fs) func(ex *Executor) {
	return func(ex *Executor) {
		ex.ectx.Fs = fs
	}
}

func WithCommandTypes(types map[string]func(ectx *ExecutorContext) Command) func(ex *Executor) {
	return func(ex *Executor) {
		ex.commandTypes = types
	}
}

func WithRegisteredCommandTypes() func(ex *Executor) {
	return WithCommandTypes(GetRegisteredCommands())
}

func WithLogger(log Logger) func(ex *Executor) {
	return func(ex *Executor) {
		ex.ectx.Logger = log
	}
}

func WithEvaluatorFunction(name string, fn govaluate.ExpressionFunction) func(ex *Executor) {
	return func(ex *Executor) {
		ex.RegisterEvaluatorFunction(name, fn)
	}
}

func WithEvaluatorFunctions(funcs map[string]govaluate.ExpressionFunction) func(ex *Executor) {
	return func(ex *Executor) {
		for name, fn := range funcs {
			ex.RegisterEvaluatorFunction(name, fn)
		}
	}
}

// WithDefaultEvaluatorFunctions registers value evaluator functions.
//
// There are 3 of them available:
// - `file_exists(filename string) (bool, error)` - returns true if file exists
// - `strlen(str string) (float64, error)` - returns the length of the string as a number
// - `shell_escape(str string) (string, error)` - returns the shell-escaped version of the string
func WithDefaultEvaluatorFunctions() func(ex *Executor) {
	return func(ex *Executor) {
		ex.RegisterEvaluatorFunction("file_exists", func(args ...any) (any, error) {
			if len(args) != 1 {
				return false, errors.New("invalid number of arguments")
			}
			fname, ok := args[0].(string)
			if !ok {
				return false, errors.New("filename must be a string")
			}
			exists, err := fileExists(ex.ectx.Fs, fname)
			return exists, err
		})

		ex.RegisterEvaluatorFunction("strlen", func(args ...any) (any, error) {
			if len(args) != 1 {
				return false, errors.New("invalid number of arguments")
			}
			length := len(args[0].(string))
			return (float64)(length), nil
		})

		ex.RegisterEvaluatorFunction("shell_escape", func(args ...any) (any, error) {
			if len(args) != 1 {
				return false, errors.New("invalid number of arguments")
			}
			result := escapeArgs([]string{args[0].(string)})
			return result, nil
		})
	}
}

func (ex *Executor) AppendScenario(scenario string) error {
	data, err := toJSON([]byte(scenario))
	if err != nil {
		return err
	}
	cmds := &RawScenario{}

	err = json.Unmarshal(data, cmds)
	if err != nil {
		return err
	}

	for id, rawCmd := range cmds.Commands {
		var tq struct{ Type string }
		err := json.Unmarshal(rawCmd, &tq)
		if err != nil {
			return err
		}
		fn := ex.commandTypes[tq.Type]
		if fn == nil {
			return errors.Errorf("invalid command type: %q", tq.Type)
		}
		cmd := fn(ex.ectx)
		err = json.Unmarshal(rawCmd, cmd)
		if err != nil {
			return err
		}

		if dicmd, ok := cmd.(DebugInfoer); ok {
			dicmd.SetDebugInfo(&CommandDebugInfo{
				ID:       id + 1,
				Contents: rawCmd,
			})
		}
		ex.commands = append(ex.commands, cmd)
	}

	return nil
}

// Execute runs installation script commands/actions according to the
// provided params map.
func (ex *Executor) Execute(variables map[string]any) (err error) {
	for _, cmd := range ex.commands {
		ex.beforeCommandExecuteCallback(cmd, variables)

		skip, err := ex.checkRequires(cmd, variables)
		if err != nil {
			return NewCommandAwareError(err, cmd, variables)
		}
		stepName := cmd.GetStepName() + ex.stepNameSuffix
		variables["__step:"+stepName+":skipped"] = skip
		if skip {
			continue
		}

		desc := cmd.GetDescription(variables)
		if desc != "" {
			ex.ectx.Logger.Info(desc)
		}
		err = cmd.Execute(variables)
		if err != nil {
			return NewCommandAwareError(err, cmd, variables)
		}

		hookName := cmd.GetHookAfter()
		if hookName == "" {
			continue
		}

		if hook := ex.hooksAfter[hookName]; hook != nil {
			err := hook(variables)
			if err != nil {
				return NewCommandAwareError(err, cmd, variables)
			}
		}
	}

	return nil
}

func (ex *Executor) SetBeforeCommandExecuteCallback(cb BeforeCommandExecuteCallback) *Executor {
	ex.beforeCommandExecuteCallback = cb
	return ex
}

func (ex *Executor) WithScenario(scenario string, opts ...Option) (*Executor, error) {
	newOpts := []Option{
		WithHooksAfter(ex.hooksAfter),
		WithStdout(ex.ectx.Stdout),
		WithStderr(ex.ectx.Stderr),
		WithFS(ex.ectx.Fs),
		WithCommandTypes(ex.commandTypes),
		WithLogger(ex.ectx.Logger),
		WithEvaluatorFunctions(ex.evaluatorFunctions),
	}
	newOpts = append(newOpts, opts...)
	result, err := NewWithScenario(scenario, newOpts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ex *Executor) WithCommands(cmds []Command, opts ...Option) *Executor {
	newOpts := []Option{
		WithCommands(cmds),
		WithHooksAfter(ex.hooksAfter),
		WithStdout(ex.ectx.Stdout),
		WithStderr(ex.ectx.Stderr),
		WithFS(ex.ectx.Fs),
		WithCommandTypes(ex.commandTypes),
		WithLogger(ex.ectx.Logger),
		WithEvaluatorFunctions(ex.evaluatorFunctions),
	}
	newOpts = append(newOpts, opts...)
	result := New(newOpts...)
	return result
}

func (ex *Executor) CommandTypes() map[string]func(ectx *ExecutorContext) Command {
	result := make(map[string]func(ectx *ExecutorContext) Command, len(ex.commandTypes))
	for k, v := range ex.commandTypes {
		result[k] = v
	}
	return result
}

func (ex *Executor) CommandTypeFn(typ string) (func(ectx *ExecutorContext) Command, bool) {
	cmd, ok := ex.commandTypes[typ]
	return cmd, ok
}

// RegisterEvaluatorFunction registers an evaluator function.
// The function must follow the signature: `func(args ...any) (any, error)`.
func (ex *Executor) RegisterEvaluatorFunction(name string, fn func(args ...any) (any, error)) *Executor {
	ex.evaluatorFunctions[name] = fn
	return ex
}

func (ex *Executor) checkRequires(cmd Command, variables map[string]any) (skip bool, err error) {
	reqs := cmd.GetRequires()
	if reqs == "" {
		return false, nil
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(reqs, ex.evaluatorFunctions)
	if err != nil {
		return false, err
	}

	reqresi, err := expression.Evaluate(variables)
	if err != nil {
		return false, err
	}

	reqres, ok := reqresi.(bool)
	if !ok {
		return false, fmt.Errorf("Requires return type must be bool, got %T", reqresi)
	}

	if !reqres {
		ex.ectx.Logger.Tracef("requirements not met, skipping %q", cmd.GetStepName())
		variables[cmd.GetStepName()] = nil
		return true, err
	}

	return false, nil
}
