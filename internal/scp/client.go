package scp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	// stores the SSH session while the connection is running
	session *ssh.Session

	// stores the SSH connection itself in order to close it after transfer
	conn ssh.Conn

	// the clients wait for the given timeout until given up the connection
	Timeout time.Duration

	// the absolute path to the remote SCP binary
	RemoteBinary string
}

func NewClient(conn ssh.Conn, session *ssh.Session) *Client {
	return &Client{
		session:      session,
		conn:         conn,
		Timeout:      time.Minute,
		RemoteBinary: "scp",
	}
}

// CopyFromFile copies the contents of an os.File to a remote location, it will get the length of the file by looking it up from the filesystem
func (a *Client) CopyFromFile(file os.File, remotePath, permissions string) error {
	stat, _ := file.Stat()
	return a.Copy(&file, remotePath, permissions, stat.Size())
}

// CopyFile copies the contents of an io.Reader to a remote location, the length is determined by reading the io.Reader until EOF
// if the file length in know in advance please use "Copy" instead
func (a *Client) CopyFile(fileReader io.Reader, remotePath, permissions string) error {
	contentsBytes, _ := io.ReadAll(fileReader)
	bytesReader := bytes.NewReader(contentsBytes)

	return a.Copy(bytesReader, remotePath, permissions, int64(len(contentsBytes)))
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// checkResponse checks the response it reads from the remote, and will return a single error in case
// of failure
func checkResponse(r io.Reader) error {
	response, err := ParseResponse(r)
	if err != nil {
		return fmt.Errorf("ParseResponse: %w", err)
	}

	if response.IsFailure() {
		return errors.New(response.GetMessage())
	}

	return nil
}

// Copy copies the contents of an io.Reader to a remote location
func (a *Client) Copy(r io.Reader, remotePath, permissions string, size int64) error {
	filename := path.Base(remotePath)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)

	w, err := a.session.StdinPipe()
	if err != nil {
		return fmt.Errorf("StdinPipe: %w", err)
	}

	stdout, err := a.session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe: %w", err)
	}

	go func() {
		defer wg.Done()
		defer w.Close()
		_, err = fmt.Fprintln(w, "C"+permissions, size, filename)
		if err != nil {
			errCh <- fmt.Errorf("perms/size/fname: %w", err)
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- fmt.Errorf("checkResponse 1: %w", err)
			return
		}

		_, err = io.Copy(w, r)
		if err != nil {
			errCh <- fmt.Errorf("io.Copy: %w", err)
			return
		}

		_, err = fmt.Fprint(w, "\x00")
		if err != nil {
			errCh <- fmt.Errorf("OK (\\x00): %w", err)
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- fmt.Errorf("checkResponse 2: %w", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		err := a.session.Run(fmt.Sprintf("%s -qt %s", a.RemoteBinary, remotePath))
		if err != nil {
			errCh <- fmt.Errorf("session run scp: %w", err)
			return
		}
	}()

	if waitTimeout(&wg, a.Timeout) {
		return errors.New("timeout when upload files")
	}

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Client) Close() error {
	return errors.Join(
		a.session.Close(),
		a.conn.Close(),
	)
}
