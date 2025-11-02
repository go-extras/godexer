package main

// running:
//   ./ssh -script script.yaml -vars vars.yaml -key privatersa.key -sshuser root -sshhost 200.201.202.203

import (
	"encoding/pem"
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/go-extras/errors"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"

	"github.com/go-extras/godexer"
	sshexec "github.com/go-extras/godexer/ssh"
)

func fileExists(fs afero.Fs, fname string) (bool, error) {
	if _, err := fs.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

type Key struct {
	PrivateKey []byte
	Password   []byte
}

func isEncryptedKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED")
}

func makeSigner(key *Key) (signer ssh.Signer, err error) {
	buf := key.PrivateKey

	if isEncryptedKey(buf) {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buf, key.Password)
	} else {
		signer, err = ssh.ParsePrivateKey(buf)
	}

	return signer, err
}

func createKeyring(keys []*Key) (authMethod ssh.AuthMethod, err error) {
	signers := make([]ssh.Signer, 0)

	// if host.Password == "" && host.Key == nil {
	//	keys := []string{
	//		os.Getenv("HOME") + "/.ssh/id_ecdsa",
	//		os.Getenv("HOME") + "/.ssh/id_rsa",
	//		//os.Getenv("HOME") + "/.ssh/id_rsa_nopass",
	//		os.Getenv("HOME") + "/.ssh/id_dsa",
	//	}
	//
	//	for _, keyname := range keys {
	//		signer, err := MakeSigner(keyname)
	//		if err == nil {
	//			signers = append(signers, signer)
	//		}
	//	}
	//}

	if keys == nil {
		panic("host keys is unexpectedly nil")
	}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}

	if len(signers) == 0 {
		return authMethod, errors.New("no usable key found")
	}

	return ssh.PublicKeys(signers...), nil
}

func main() {
	scriptPtr := flag.String("script", "", "script filename")
	varsPtr := flag.String("vars", "", "variables filename")
	keyPtr := flag.String("key", "", "key filename")
	keyPasswordPtr := flag.String("key-password", "", "key password")
	sshHostPtr := flag.String("sshhost", "", "ssh host")
	sshUserPtr := flag.String("sshuser", "", "ssh user")
	flag.Parse()

	fs := afero.NewOsFs()

	// script
	if exists, _ := fileExists(fs, *scriptPtr); !exists {
		panic("script file not found")
	}
	scriptFile, err := fs.Open(*scriptPtr)
	if err != nil {
		panic(err)
	}
	defer scriptFile.Close()
	scriptCmds, err := afero.ReadAll(scriptFile)
	if err != nil {
		panic(err)
	}

	// vars
	if exists, _ := fileExists(fs, *varsPtr); !exists {
		panic("vars file not found")
	}
	varsFile, err := fs.Open(*varsPtr)
	if err != nil {
		panic(err)
	}
	defer varsFile.Close()
	varsData, err := afero.ReadAll(varsFile)
	if err != nil {
		panic(err)
	}

	// key file
	if exists, _ := fileExists(fs, *keyPtr); !exists {
		panic("key file not found")
	}
	keyFile, err := fs.Open(*keyPtr)
	if err != nil {
		panic(err)
	}
	defer keyFile.Close()
	keydata, err := afero.ReadAll(keyFile)
	if err != nil {
		panic(err)
	}
	keys := make([]*Key, 1)
	keys[0] = &Key{
		PrivateKey: keydata,
		Password:   []byte(*keyPasswordPtr),
	}

	auth := make([]ssh.AuthMethod, 1)
	authMethod, err := createKeyring(keys)
	if err != nil {
		panic(errors.Errorf("unable to choose authentication method: %v", err))
	}
	auth[0] = authMethod
	if *sshUserPtr == "" {
		panic(errors.New("missing user name"))
	}
	config := &ssh.ClientConfig{
		User: *sshUserPtr,
		Auth: auth,
		//nolint:gosec // This is example code; production code should use proper host key verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(*sshHostPtr, "22"), config)
	if err != nil {
		panic(err)
	}
	commands := godexer.GetRegisteredCommands()
	commands["scp_writefile"] = sshexec.NewScpWriterFileCommand(conn)
	commands["ssh_exec"] = sshexec.NewSSHExecCommand(conn, os.Stdout, os.Stderr)

	exc, err := godexer.NewWithScenario(
		string(scriptCmds),
		godexer.WithCommandTypes(commands),
		godexer.WithDefaultEvaluatorFunctions(),
	)
	if err != nil {
		panic(err)
	}

	var vars map[string]any
	err = yaml.Unmarshal(varsData, &vars)
	if err != nil {
		panic(err)
	}

	err = exc.Execute(vars)
	if err != nil {
		panic(err)
	}

	log.Println("No errors occured")
	// log.Printf("Variable %q has a value %q\n", "output", vars["output"])
}
