package scp_test

import (
	"bytes"
	"crypto/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-extras/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	. "github.com/go-extras/godexer/internal/scp"
	"github.com/go-extras/godexer/internal/testutils"
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

func genRandomBytes(size int) (blk []byte, err error) {
	blk = make([]byte, size)
	_, err = rand.Read(blk)
	return blk, err
}

func getScpHandler(t *testing.T, expectedFname string, expectedSize int, uploaded *int32) func(cmd string, ch ssh.Channel) error {
	return func(cmd string, ch ssh.Channel) error {
		if strings.HasPrefix(cmd, "scp -qt") {
			filePath := cmd[8:]
			if t != nil {
				require.Equal(t, expectedFname, filePath)
			}

			// log.Print("got scp request")

			//// send ok
			//ch.Write([]byte("\x00"))
			//// log.Print("sent ok1")

			var fileName [1024]byte
			// read file name
			n, err := ch.Read(fileName[:])
			if err != nil {
				return err
			}
			// log.Printf("read %d bytes (filename)", n)
			// C0644 1073741824 path
			fileNameStr := strings.Trim(string(fileName[:]), "\x00\n")
			parts := strings.Split(fileNameStr, " ")
			// mode := parts[0]                    // mode
			size, err := strconv.Atoi(parts[1]) // size
			if err != nil {
				return errors.Wrap(err, "can't convert size")
			}
			// name := parts[2] // name

			if t != nil {
				require.Equal(t, expectedSize, size)
			}

			// log.Printf("mode=%s, size=%d, name=%s", mode, size, name)

			// send ok
			ch.Write([]byte("\x00"))
			// log.Print("sent ok2")

			var fileData []byte

			// read file data
			for size > 0 {
				size -= 1024

				var fileDataChunk [1024]byte
				n, err = ch.Read(fileDataChunk[:])
				if err != nil {
					return err
				}
				// log.Printf("read %d bytes (filedata)", n)

				fileData = append(fileData, fileDataChunk[:n]...)
			}

			// log.Printf("read size %d", len(fileData))

			ch.Write([]byte("\x00"))
			// log.Print("sent ok3")

			atomic.StoreInt32(uploaded, 1)

			ch.SendRequest("exit-status", false, ssh.Marshal(&testutils.MsgExit{Status: 0}))
			// log.Print("sent exit-status")
			ch.Close()
			// log.Print("closed channel")

			return nil
		}

		return nil
	}
}

func TestClient_Copy(t *testing.T) {
	const DataSize = 5 * 1024 * 1024

	r := require.New(t)

	signer, err := testutils.MakeSigner(key)
	r.NoError(err)

	var uploaded int32

	srv := testutils.NewServer(signer, nil, getScpHandler(t, "/remote/path", DataSize, &uploaded))
	go srv.Start()
	defer srv.Stop()

	addr := srv.Addr()

	cfg, err := testutils.GetClientConfig("root", key)
	r.NoError(err)

	sshClient, err := testutils.CreateConn(addr.IP.String(), strconv.Itoa(addr.Port), cfg)
	r.NoError(err)

	session, err := sshClient.NewSession()
	r.NoError(err)

	defer session.Close()
	client := NewClient(sshClient.Conn, session)

	data, _ := genRandomBytes(DataSize)
	buf := bytes.NewReader(data)
	err = client.Copy(buf, "/remote/path", "0644", int64(len(data)))
	r.NoError(err)

	actualUploaded := atomic.LoadInt32(&uploaded)
	r.Equal(int32(1), actualUploaded)
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func BenchmarkClient_Copy(b *testing.B) {
	const DataSize = 5 * 1024 * 1024

	signer, err := testutils.MakeSigner(key)
	checkerr(err)

	var uploaded int32
	srv := testutils.NewServer(signer, nil, getScpHandler(nil, "/remote/path", DataSize, &uploaded))
	go srv.Start()
	defer srv.Stop()

	addr := srv.Addr()

	cfg, err := testutils.GetClientConfig("root", key)
	checkerr(err)

	sshClient, err := testutils.CreateConn(addr.IP.String(), strconv.Itoa(addr.Port), cfg)
	checkerr(err)

	data, _ := genRandomBytes(DataSize)

	// run the function b.N times
	for n := 0; n < b.N; n++ {
		session, err := sshClient.NewSession()
		checkerr(err)
		client := NewClient(sshClient.Conn, session)
		buf := bytes.NewReader(data)
		err = client.Copy(buf, "/remote/path", "0644", int64(len(data)))
		session.Close()
		checkerr(err)

		actualUploaded := atomic.LoadInt32(&uploaded)
		if actualUploaded != 1 {
			panic(actualUploaded)
		}
		atomic.StoreInt32(&uploaded, 0)
	}
}
