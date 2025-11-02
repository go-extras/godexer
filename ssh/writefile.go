package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-extras/errors"
	"golang.org/x/crypto/ssh"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/scp"
)

func NewScpWriterFileCommand(sshClient *ssh.Client) func(ectx *executor.ExecutorContext) executor.Command {
	return func(ectx *executor.ExecutorContext) executor.Command {
		return &ScpWriteFileCommand{
			sshClient: sshClient,
			BaseCommand: executor.BaseCommand{
				Ectx: ectx,
			},
		}
	}
}

type ScpWriteFileCommand struct {
	executor.BaseCommand
	sshClient            *ssh.Client
	File                 string
	Contents             string
	ContentsFromFile     string
	ContentsFromVariable string
	Permissions          string
	Timeout              int
}

func (r *ScpWriteFileCommand) Execute(variables map[string]any) error {
	if len(r.File) == 0 {
		return errors.Errorf("filename in %q is empty", r.StepName)
	}
	if r.Permissions == "" {
		return errors.Errorf("filemode permissions in %q are empty", r.StepName)
	}

	remoteFileName := executor.MaybeEvalValue(r.File, variables).(string)

	var reader io.Reader

	switch {
	case r.ContentsFromVariable != "":
		variable := executor.MaybeEvalValue(r.ContentsFromVariable, variables).(string)
		switch v := variables[variable].(type) {
		case string:
			reader = strings.NewReader(v)
		case []byte:
			reader = bytes.NewReader(v)
		case fmt.Stringer:
			reader = strings.NewReader(v.String())
		}
	case r.ContentsFromFile != "":
		fileName := executor.MaybeEvalValue(r.ContentsFromFile, variables)
		f, err := os.Open(fileName.(string))
		if err != nil {
			return errors.Wrap(err, "can't open local file")
		}
		defer f.Close()
		reader = f
	default:
		contents := executor.MaybeEvalValue(r.Contents, variables)
		reader = strings.NewReader(contents.(string))
	}

	session, err := r.sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "unable to get ssh session")
	}
	defer session.Close()
	client := scp.NewClient(r.sshClient.Conn, session)

	if r.Timeout > 0 {
		client.Timeout = time.Duration(r.Timeout) * time.Second
	}

	r.Ectx.Logger.Debugf("Writing to %s", remoteFileName)
	err = client.CopyFile(reader, remoteFileName, r.Permissions)
	if err != nil {
		return err
	}

	return nil
}
