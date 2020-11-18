package operator

import (
	"bytes"
	"github.com/bramvdbogaerde/go-scp"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHOperator struct {
	conn *ssh.Client
}

func NewSSHOperator(address string, config *ssh.ClientConfig) (*SSHOperator, error) {
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	operator := SSHOperator{
		conn: conn,
	}

	return &operator, nil
}

func (s SSHOperator) Close() error {
	return s.conn.Close()
}

func (s SSHOperator) Execute(command string) (CommandRes, error) {
	sess, err := s.conn.NewSession()
	if err != nil {
		return CommandRes{}, err
	}

	defer sess.Close()

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return CommandRes{}, err
	}

	output := bytes.Buffer{}

	wg := sync.WaitGroup{}

	stdOutWriter := io.MultiWriter(os.Stdout, &output)
	wg.Add(1)
	go func() {
		io.Copy(stdOutWriter, sessStdOut)
		wg.Done()
	}()
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return CommandRes{}, err
	}

	errorOutput := bytes.Buffer{}
	stdErrWriter := io.MultiWriter(os.Stderr, &errorOutput)
	wg.Add(1)
	go func() {
		io.Copy(stdErrWriter, sessStderr)
		wg.Done()
	}()

	err = sess.Run(command)

	wg.Wait()

	if err != nil {
		return CommandRes{}, err
	}

	return CommandRes{
		StdErr: errorOutput.Bytes(),
		StdOut: output.Bytes(),
	}, nil
}

func (s SSHOperator) Upload(source io.Reader, remotePath string, mode string) error {
	sess, err := s.conn.NewSession()
	if err != nil {
		return err
	}

	defer sess.Close()

	client := scp.Client{
		Session:      sess,
		Conn:         s.conn,
		Timeout:      time.Minute,
		RemoteBinary: "scp",
	}

	err = client.CopyFile(source, remotePath, mode)

	return err
}

func (s SSHOperator) UploadFile(path string, remotePath string, mode string) error {
	source, err := os.Open(expandPath(path))
	if err != nil {
		return err
	}
	defer source.Close()

	return s.Upload(source, remotePath, mode)
}
