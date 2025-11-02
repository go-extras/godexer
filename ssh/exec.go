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
	godexer.BaseCommand
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

func NewSSHExecCommand(sshClient *ssh.Client, stdout, stderr io.Writer) func(ectx *godexer.ExecutorContext) godexer.Command {
	return func(ectx *godexer.ExecutorContext) godexer.Command {
		return &ExecCommand{
			sshClient: sshClient,
			stdout:    stdout,
			stderr:    stderr,
			BaseCommand: godexer.BaseCommand{
				Ectx: ectx,
			},
		}
	}
}

func (r *ExecCommand) Execute(variables map[string]any) error {
	if r.Ectx.Executor == nil {
		return errors.Errorf("this command must be run from the executor")
	}

	cmd, err := r.prepareCommand(variables)
	if err != nil {
		return err
	}

	r.printCommand(cmd, variables)

	session, err := r.createSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var buf godexer.Buffer
	r.setupSessionIO(session, &buf)

	if err := r.setEnvironment(session, variables); err != nil {
		return err
	}

	err = r.runCommand(session, cmd, &buf, variables)
	if err == nil {
		return nil
	}

	return r.handleError(err, variables)
}

func (r *ExecCommand) prepareCommand(variables map[string]any) (string, error) {
	var cmds []string
	for _, v := range r.Cmd {
		cmds = append(cmds, godexer.MaybeEvalValue(v, variables).(string))
	}

	cmd := escapeArgs(cmds)
	if len(cmd) == 0 {
		return "", errors.Errorf("command %q is empty", r.StepName)
	}

	return cmd, nil
}

func (r *ExecCommand) printCommand(cmd string, variables map[string]any) {
	addr := r.sshClient.RemoteAddr().String()
	switch r.CmdRedact {
	case "":
		fmt.Fprintf(r.stdout, "%s$ %s\n", addr, cmd)
	case "-":
		fmt.Fprintf(r.stdout, "%s$ %s\n", addr, "[command redacted]")
	default:
		cmdRedact, ok := godexer.MaybeEvalValue(r.CmdRedact, variables).(string)
		if !ok {
			cmdRedact = "[invalid redact value]"
		}
		fmt.Fprintf(r.stdout, "%s$ %s\n", addr, cmdRedact)
	}
}

func (r *ExecCommand) createSession() (*ssh.Session, error) {
	session, err := r.sshClient.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ssh session")
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, errors.Wrap(err, "failed to request pty")
	}

	return session, nil
}

func (r *ExecCommand) setupSessionIO(session *ssh.Session, buf *godexer.Buffer) {
	if r.Variable == "" {
		session.Stdout = r.Ectx.Stdout
		session.Stderr = r.Ectx.Stderr
	} else {
		session.Stdout = godexer.NewCombinedWriter([]io.Writer{r.Ectx.Stdout, buf})
		session.Stderr = godexer.NewCombinedWriter([]io.Writer{r.Ectx.Stderr, buf})
	}
}

func (r *ExecCommand) setEnvironment(session *ssh.Session, variables map[string]any) error {
	if r.Env == nil {
		return nil
	}

	for k, v := range r.Env {
		if err := session.Setenv(k, godexer.MaybeEvalValue(v, variables).(string)); err != nil {
			return errors.Wrap(err, "failed to set ssh environment variable")
		}
	}
	return nil
}

func (r *ExecCommand) runCommand(session *ssh.Session, cmd string, buf *godexer.Buffer, variables map[string]any) error {
	if err := session.Start(cmd); err != nil {
		return err
	}

	err := session.Wait()

	if r.Variable != "" {
		variables[r.Variable] = buf.String()
	}

	if r.AllowFail {
		r.handleAllowFail(err, variables)
		return nil
	}

	return err
}

func (r *ExecCommand) handleAllowFail(err error, variables map[string]any) {
	if exitError, ok := err.(*ssh.ExitError); ok {
		variables[r.StepName+"_exit_status"] = exitError.ExitStatus()
	} else {
		variables[r.StepName+"_exit_status"] = 0
	}
}

func (r *ExecCommand) handleError(err error, variables map[string]any) error {
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
		return r.Execute(variables)
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
