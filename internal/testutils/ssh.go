package testutils

import (
	"encoding/pem"
	"fmt"
	"net"
	"strings"

	"github.com/go-extras/errors"
	"golang.org/x/crypto/ssh"
)

func createKeyring(key *Key) (authMethod ssh.AuthMethod, err error) {
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

	signer, err := MakeSigner(key)
	if err == nil {
		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		return authMethod, errors.New("no usable key found")
	}

	return ssh.PublicKeys(signers...), nil
}

func GetClientConfig(user string, key *Key) (*ssh.ClientConfig, error) {
	auth := make([]ssh.AuthMethod, 1)

	authMethod, err := createKeyring(key)
	if err != nil {
		return nil, errors.Wrap(err, "unable to choose authentication method")
	}
	auth[0] = authMethod

	config := &ssh.ClientConfig{
		User: user,
		Auth: auth,
		//nolint:gosec // This is test utility code; production code should use proper host key verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

func CreateConn(host, port string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := ssh.Dial("tcp", net.JoinHostPort(host, port), config)
	return conn, err
}

func isEncryptedKey(pemBytes []byte) bool {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return false
	}

	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED")
}

type Key struct {
	PrivateKey string
	Password   string
}

func MakeSigner(key *Key) (signer ssh.Signer, err error) {
	buf := []byte(key.PrivateKey)

	if isEncryptedKey(buf) {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buf, []byte(key.Password))
	} else {
		signer, err = ssh.ParsePrivateKey(buf)
	}

	return signer, err
}

type MsgExit struct {
	Status uint32
}

type MsgString struct {
	Data string
}

type HandlerFunc func(string) ([]byte, uint32, bool)
type HandlerChFunc func(string, ssh.Channel) error

func NewServer(signer ssh.Signer, handlerFunc HandlerFunc, handlerChFunc HandlerChFunc) *Server {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
		PasswordCallback: func(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("Failed to listen (%+v)", err))
	}

	return &Server{
		listener:      listener,
		config:        config,
		handlerFunc:   handlerFunc,
		handlerChFunc: handlerChFunc,
	}
}

type Server struct {
	listener      net.Listener
	config        *ssh.ServerConfig
	handlerFunc   HandlerFunc
	handlerChFunc HandlerChFunc
}

func (s *Server) Start() {
	for {
		tcpConn, err := s.listener.Accept()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return
		}

		if err != nil {
			panic(fmt.Sprintf("Failed to accept incoming connection (%s)", err))
		}

		_, chans, reqs, err := ssh.NewServerConn(tcpConn, s.config)
		if err != nil {
			panic(fmt.Sprintf("Failed to handshake (%s)", err))
		}

		// log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go s.handleChannels(chans)
	}
}

func (s *Server) Stop() error {
	return s.listener.Close()
}

func (s *Server) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

func (s *Server) handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	ch, requests, err := newChannel.Accept()
	if err != nil {
		panic(fmt.Sprintf("Could not accept channel (%s)", err))
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "exec"
	// We just want to handle "exec".
	go func() {
		for req := range requests {
			switch req.Type {
			case "exec":
				var msg MsgString
				if err := ssh.Unmarshal(req.Payload, &msg); err != nil {
					panic(fmt.Sprintf("Could Unmarshal command (%+v, %+v)", req.Payload, err))
				}
				// log.Printf("got command: %q", msg.Data)
				req.Reply(true, nil)

				if s.handlerChFunc != nil {
					err = s.handlerChFunc(msg.Data, ch)
					if err != nil {
						panic(err)
					}
					continue
				}

				data, exitStatus, wantClose := s.handlerFunc(msg.Data)
				ch.Write(data)
				ch.SendRequest("exit-status", false, ssh.Marshal(&MsgExit{Status: exitStatus}))
				if wantClose {
					ch.Close()
				}
			case "pty-req":
				req.Reply(true, nil)
			default:
				panic(req.Type)
			}
		}
	}()
}

func (s *Server) Addr() *net.TCPAddr {
	addr := s.listener.Addr()
	a, ok := addr.(*net.TCPAddr)
	if !ok {
		panic("invalid dummy server addr type")
	}

	return a
}
