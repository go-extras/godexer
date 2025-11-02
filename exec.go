package godexer

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/go-extras/errors"
)

var ExecCommandFn = exec.Command

//nolint:gochecknoinits // init is used for automatic command registration
func init() {
	RegisterCommand("", NewExecCommand) // default command
	RegisterCommand("exec", NewExecCommand)
}

func escapeArgs(args []string) (result string) {
	if len(args) == 0 {
		return result
	}

	var escaped []string
	for _, s := range args {
		escaped = append(escaped, shellescape.Quote(s))
	}
	return strings.Join(escaped, " ")
}

type ExecCommand struct {
	BaseCommand
	Cmd       []string
	Variable  string
	AllowFail bool
	Env       []string

	// the following parameters will allow retrying the command
	Attempts int // if 0 or 1, no retry
	Delay    int // seconds
}

func NewExecCommand(ectx *ExecutorContext) Command {
	return &ExecCommand{
		BaseCommand: BaseCommand{
			Ectx: ectx,
		},
	}
}

func (r *ExecCommand) Execute(variables map[string]any) error {
	if len(r.Cmd) == 0 {
		return errors.Errorf("command %q is empty", r.StepName)
	}

	var cmds []string
	for _, v := range r.Cmd {
		cmds = append(cmds, MaybeEvalValue(v, variables).(string))
	}

	if len(cmds) == 0 {
		// should never actually happen
		return errors.Errorf("command %q is empty", r.StepName)
	}

	r.Ectx.Logger.Info(strings.TrimSpace(fmt.Sprintf("Executing: %s %s", cmds[0], escapeArgs(cmds[1:]))))

	cmd := ExecCommandFn(cmds[0], cmds[1:]...)

	var buf Buffer
	if r.Variable == "" {
		cmd.Stdout = r.Ectx.Stdout
		cmd.Stderr = r.Ectx.Stderr
	} else {
		cmd.Stdout = NewCombinedWriter([]io.Writer{
			r.Ectx.Stdout,
			&buf,
		})
		cmd.Stderr = NewCombinedWriter([]io.Writer{
			r.Ectx.Stderr,
			&buf,
		})
	}

	cmd.Env = append(cmd.Env, r.Env...)

	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()

	if r.Variable != "" {
		variables[r.Variable] = buf.String()
	}

	if r.AllowFail {
		if exitError, ok := err.(*exec.ExitError); ok {
			err = nil
			variables[r.StepName+"_exit_status"] = exitError.ExitCode()
		} else {
			variables[r.StepName+"_exit_status"] = 0
		}
	}

	if err != nil {
		r.Ectx.Logger.Infof("Got an error and attempts = %d", r.Attempts)
	}

	if r.Attempts > 1 {
		if _, ok := err.(*exec.ExitError); ok {
			r.Attempts--
			r.Ectx.Logger.Infof("Got execution failure, will retry (attempts left %d)", r.Attempts)
			TimeSleep(time.Duration(r.Delay) * time.Second)
			err = r.Execute(variables)
		}
	}

	return err
}
