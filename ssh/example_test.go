package ssh_test

import (
	"io"

	"golang.org/x/crypto/ssh"

	"github.com/go-extras/godexer"
	sshexec "github.com/go-extras/godexer/ssh"
)

// Example_registerSSHCommands shows how to wire SSH-based commands into the
// executor. This illustrative example does not perform any network calls.
func Example_registerSSHCommands() {
	// In real code, obtain an *ssh.Client via ssh.Dial.
	var client *ssh.Client

	cmds := godexer.GetRegisteredCommands()
	cmds["scp_writefile"] = sshexec.NewScpWriterFileCommand(client)
	cmds["ssh_exec"] = sshexec.NewSSHExecCommand(client, io.Discard, io.Discard)

	// scenario := "..."
	// ex, _ := godexer.NewWithScenario(scenario, godexer.WithCommandTypes(cmds))
	// _ = ex.Execute(map[string]any{"name": "world"})
}
