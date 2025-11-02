package ssh_test

import (
	"bytes"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
	"github.com/go-extras/godexer/internal/testutils"
	sshexec "github.com/go-extras/godexer/ssh"
)

func TestScpWriteFile(t *testing.T) {
	t.Run("Execute_EmptyFilename", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, nil, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create command
		var stdout, stderr bytes.Buffer
		cmd := sshexec.NewScpWriterFileCommand(client)(&godexer.ExecutorContext{
			Stdout: &stdout,
			Stderr: &stderr,
			Logger: &logger.Logger{},
		})

		writeCmd := cmd.(*sshexec.ScpWriteFileCommand)
		writeCmd.File = ""
		writeCmd.Contents = "test"
		writeCmd.Permissions = "0644"
		writeCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = writeCmd.Execute(vars)
		c.Assert(err, qt.ErrorMatches, `filename in "test_step" is empty`)
	})

	t.Run("Execute_EmptyPermissions", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, nil, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create command
		var stdout, stderr bytes.Buffer
		cmd := sshexec.NewScpWriterFileCommand(client)(&godexer.ExecutorContext{
			Stdout: &stdout,
			Stderr: &stderr,
			Logger: &logger.Logger{},
		})

		writeCmd := cmd.(*sshexec.ScpWriteFileCommand)
		writeCmd.File = "/tmp/test.txt"
		writeCmd.Contents = "test"
		writeCmd.Permissions = ""
		writeCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = writeCmd.Execute(vars)
		c.Assert(err, qt.ErrorMatches, `filemode permissions in "test_step" are empty`)
	})
}
