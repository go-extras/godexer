package ssh_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/go-extras/godexer"
	"github.com/go-extras/godexer/internal/logger"
	"github.com/go-extras/godexer/internal/testutils"
	sshexec "github.com/go-extras/godexer/ssh"
)

var key = &testutils.Key{
	PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAlOXclIZrBuVPpMYUjRh8I+WRCn9rxAJdrTkb+yyqAzE0OWj9
yYSEvGlNziVvvDHWgJNQ+9mKHeu+1r8p1BTu5FsSF0XiFkFh2D/8MuiPPv+LNeJa
b+RshXp/AXwtFq6n721jf5/o4u4ukUEViLst6e+e0HFvxyI9hJMh9j3lAyoJ+xrq
25PyPcrj6D/NKp+ZwNJDHBto7R4GrVOcuDnEMgnmpudfe0Hb3UmGCuIs4Dfr2koq
98OHXCzF1sZM2INyU8Vr0zeYAVV6clW6kXaWJfe7dzWVcG3ZoUodMGyDWY3XqxHL
Z/H3N69mkrdlpfrd9/JuWJRvjWX0qaQi8dfuZQIDAQABAoIBAQCCGFhjGRL4QnEU
6dDY+tS0VIcmofBZoSuSBzzwd7TP9zTHGHntkbCcInHNtR3sU6s0SgLPGeI4hFsI
rJvyZpvXv86NsQx6H4RK+pTzMgi+pW5PlUcpTm6XLVE8ze9jSxUF+BCgWOqVJEBh
v3j+L3VNWYTsYMCmP796T0e0K54l5T6h5VYTUkbDrRUCJzwBc+VemNXu+N/t2KwT
qjrPr3Bi8xPWPnCZGzf7rrX27jBrPac1xoCIkWfkQvq/L729RPKfKoKj+/+Jwt+h
GOLIPpVK/zX8cyaR3EItzUrtTMY6RjCslSyverBdovP6MvUsJAYH+jsa4OlnGwOH
hYmRcYV9AoGBAPOllTmNz5rwpghXIW1zxAjTFeTTsedURTJRCpw5Y7/t/tZSDqTb
BD9Fwwck7Jhmyv/4rEbHEKzJmAKB/atXmRzMJwG8WEtA7uTLsrQMeoGwtwxrqPbJ
zDCiu8Pv0b5ieGnenqzlKDYhYCNcQ8hSKQG3JKZ6Ha+HE4LZP3NYrpEDAoGBAJxy
e+NrA1jVxDpokDZJAdUGZ49lHYD6FCuUN6mGtTgH7EG7Wveiy1a4CJrwwoyYTNnE
s4GPgFYvnxlv7Y9Fk4/JpZYcVSoc+9kVXuVptiDIA1hWGBW74t/4knMU7TZWwGZZ
8ybJreE4s6RG0JvunmxtFL6kD/LBGXZCklhqBYJ3AoGBANEcAOnnixFojqdD2J2u
qMYGHJlLEzn+OnFH2rpgCvtz0K6yuHzGuGtxfUQJbcITHxD3pSwNt4MEdiFY3ZUL
1o4/rQ6xTnov3ZiiNtqOhyn9t+zCDb7ZTRVE5a/xiOtEaiI6/aZX+t4SYQeYLVil
Iyqku6Dh186JOLapq+pcZ15vAoGBAJxpUTdTXCtKvT7wH45Ge5BxMMSKgW7bl6Li
MqxIw5FbSneFSzNeDRGMOP4/SyKpedwW7qjPwa1pOxWBc+7Tzu3o2qYzeWn7REgL
N68Be1dW4RFGMho4mGD38eMgvvCe1wj9UT4sUK1ltSS+r/3WGYmpnR3khRVcvYog
kJPYm92NAoGAX8scLZ767EJKz9c6NFHA/bcPlPL+QjgP0Gcc5OiaWAxpxwR7bQK6
7BKt+GU4dxsy3lUW6eelG9CSBi7i6J4bqXO3AA85BwaNmYjXUDPx/jwZ/ReqTQ3w
iwoaWqGXRXFVqnc+bqyGTyDthKcg5lCXhDK0vOOhTqF2ev6lVEkxwoE=
-----END RSA PRIVATE KEY-----
`}

func TestSSHExec(t *testing.T) {
	t.Run("Execute_Success", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("hello world"), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"echo", "test"}
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(stdout.String(), qt.Contains, "echo test")
	})

	t.Run("Execute_WithVariable", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("captured output"), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"echo", "test"}
		execCmd.Variable = "output"
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(vars["output"], qt.Contains, "captured output")
	})

	t.Run("Execute_AllowFail", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server that returns error
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("error output"), 1, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"false"}
		execCmd.AllowFail = true
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(vars["test_step_exit_status"], qt.Equals, 1)
	})

	t.Run("Execute_EmptyCommand", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte(""), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = nil
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.ErrorMatches, `command "test_step" is empty`)
	})

	t.Run("Execute_WithRetry", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server that fails first time, succeeds second time
		attemptCount := 0
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			attemptCount++
			if attemptCount == 1 {
				return []byte("error"), 1, true
			}
			return []byte("success"), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Mock TimeSleep to avoid actual delay
		originalTimeSleep := sshexec.TimeSleep
		sshexec.TimeSleep = func(d time.Duration) {
			// Do nothing
		}
		defer func() { sshexec.TimeSleep = originalTimeSleep }()

		// Create command
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"test"}
		execCmd.Attempts = 2
		execCmd.Delay = 1
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(attemptCount, qt.Equals, 2)
	})

	t.Run("Execute_WithCmdRedact", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("output"), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command with redacted output
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"echo", "secret"}
		execCmd.CmdRedact = "-"
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		c.Assert(stdout.String(), qt.Contains, "[command redacted]")
		c.Assert(stdout.String(), qt.Not(qt.Contains), "secret")
	})

	t.Run("Execute_WithOnEachFailure", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server that always fails
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("error"), 1, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command with OnEachFailure
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"false"}
		execCmd.StepName = "test_step"

		// Add a simple message command as OnEachFailure
		msgCmd := map[string]any{
			"type":        "message",
			"stepName":    "failure_msg",
			"description": "Command failed",
		}
		msgJSON, _ := json.Marshal(msgCmd)
		execCmd.OnEachFailure = []json.RawMessage{msgJSON}

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.Not(qt.IsNil))
	})

	t.Run("Execute_NoExecutor", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte("output"), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create command without executor
		var stdout, stderr bytes.Buffer
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Stdout: &stdout,
			Stderr: &stderr,
			Logger: &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"echo", "test"}
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.ErrorMatches, "this command must be run from the executor")
	})
}

func TestSSHExec_EscapeArgs(t *testing.T) {
	t.Run("EscapeArgs_Empty", func(t *testing.T) {
		c := qt.New(t)

		// Create SSH server
		signer, err := testutils.MakeSigner(key)
		c.Assert(err, qt.IsNil)

		server := testutils.NewServer(signer, func(cmd string) ([]byte, uint32, bool) {
			return []byte(""), 0, true
		}, nil)
		go server.Start()
		defer server.Stop()

		// Create SSH client
		config, err := testutils.GetClientConfig("testuser", key)
		c.Assert(err, qt.IsNil)

		port := fmt.Sprintf("%d", server.Addr().Port)
		client, err := testutils.CreateConn("127.0.0.1", port, config)
		c.Assert(err, qt.IsNil)
		defer client.Close()

		// Create executor
		var stdout, stderr bytes.Buffer
		ex := godexer.New(
			godexer.WithStdout(&stdout),
			godexer.WithStderr(&stderr),
			godexer.WithLogger(&logger.Logger{}),
		)

		// Create command with special characters
		cmd := sshexec.NewSSHExecCommand(client, &stdout, &stderr)(&godexer.ExecutorContext{
			Executor: ex,
			Stdout:   &stdout,
			Stderr:   &stderr,
			Logger:   &logger.Logger{},
		})

		execCmd := cmd.(*sshexec.ExecCommand)
		execCmd.Cmd = []string{"echo", "hello world", "test$var"}
		execCmd.StepName = "test_step"

		vars := make(map[string]any)
		err = execCmd.Execute(vars)
		c.Assert(err, qt.IsNil)
		// The command should be properly escaped
		output := stdout.String()
		c.Assert(output, qt.Contains, "echo")
	})
}
