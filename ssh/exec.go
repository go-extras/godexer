package ssh

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/go-extras/errors"
	"golang.org/x/crypto/ssh"

	"github.com/go-extras/godexer"
)

var TimeSleep = time.Sleep

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
	executor.BaseCommand
	sshClient      *ssh.Client
	stdout         io.Writer
	stderr         io.Writer
	Cmd            []string
	CmdRedact      string
	Variable       string
	AllowFail      bool
	OnEachFailure  []json.RawMessage
	OnFinalFailure []json.RawMessage
	Env            map[string]string

	// the following parameters will allow retrying the command
	Attempts int // if 0 or 1, no retry
	Delay    int // seconds
}

func NewSshExecCommand(sshClient *ssh.Client, stdout, stderr io.Writer) func(ectx *executor.ExecutorContext) executor.Command {
	return func(ectx *executor.ExecutorContext) executor.Command {
		return &ExecCommand{
			sshClient: sshClient,
			stdout:    stdout,
			stderr:    stderr,
			BaseCommand: executor.BaseCommand{
				Ectx: ectx,
			},
		}
	}
}

func (r *ExecCommand) Execute(variables map[string]any) error {
	if r.Ectx.Executor == nil {
		return errors.Errorf("this command must be run from the executor")
	}

	var cmds []string
	for _, v := range r.Cmd {
		cmds = append(cmds, executor.MaybeEvalValue(v, variables).(string))
	}

	cmd := escapeArgs(cmds)

	if len(cmd) == 0 {
		return errors.Errorf("command %q is empty", r.StepName)
	}

	switch r.CmdRedact {
	case "":
		fmt.Fprintf(r.stdout, "%s$ %s\n", r.sshClient.RemoteAddr().String(), cmd)
	case "-":
		fmt.Fprintf(r.stdout, "%s$ %s\n", r.sshClient.RemoteAddr().String(), "[command redacted]")
	default:
		cmdRedact := executor.MaybeEvalValue(r.CmdRedact, variables).(string)
		fmt.Fprintf(r.stdout, "%s$ %s\n", r.sshClient.RemoteAddr().String(), cmdRedact)
	}

	session, err := r.sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "unable to get ssh session")
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return errors.Wrap(err, "failed to request pty")
	}

	var buf executor.Buffer
	if r.Variable == "" {
		session.Stdout = r.Ectx.Stdout
		session.Stderr = r.Ectx.Stderr
	} else {
		session.Stdout = executor.NewCombinedWriter([]io.Writer{
			r.Ectx.Stdout,
			&buf,
		})
		session.Stderr = executor.NewCombinedWriter([]io.Writer{
			r.Ectx.Stderr,
			&buf,
		})
	}

	if r.Env != nil {
		for k, v := range r.Env {
			if err := session.Setenv(k, executor.MaybeEvalValue(v, variables).(string)); err != nil {
				return errors.Wrap(err, "failed to set ssh environment variable")
			}
		}
	}

	err = session.Start(cmd)
	if err != nil {
		return err
	}
	err = session.Wait()

	if r.Variable != "" {
		variables[r.Variable] = buf.String()
	}

	if r.AllowFail {
		if exitError, ok := err.(*ssh.ExitError); ok {
			err = nil
			variables[r.StepName+"_exit_status"] = exitError.ExitStatus()
		} else {
			variables[r.StepName+"_exit_status"] = 0
		}
	}

	if err == nil {
		return nil
	}

	r.Ectx.Logger.Infof("Got an error and attepts = %d", r.Attempts)

	if r.OnEachFailure != nil {
		innerErr := r.onFailure(r.OnEachFailure, variables)
		r.Ectx.Logger.Errorf("Got an error when running OnEachFailure: %+v", innerErr)
	}

	if r.Attempts <= 1 {
		if r.OnFinalFailure != nil {
			innerErr := r.onFailure(r.OnFinalFailure, variables)
			r.Ectx.Logger.Errorf("Got an error when running OnFinalFailure: %+v", innerErr)
		}
		return err
	}

	if _, ok := err.(*ssh.ExitError); ok {
		r.Attempts--
		r.Ectx.Logger.Infof("Got execution failure, will retry (attempts left %d)", r.Attempts)
		TimeSleep(time.Duration(r.Delay) * time.Second)
		err = r.Execute(variables)
	}

	return err
}

func (r *ExecCommand) onFailure(commands []json.RawMessage, variables map[string]any) error {
	cmdScriptObj := struct {
		Commands any `json:"commands"`
	}{}
	cmdScriptObj.Commands = commands
	script, err := json.Marshal(cmdScriptObj)
	if err != nil {
		return errors.Wrap(err, "cannot marshal commands script")
	}

	e, err := r.Ectx.Executor.WithScenario(string(script))
	if err != nil {
		return errors.Wrap(err, "cannot load child executor")
	}

	return e.Execute(variables)
}
