package operator

import (
	goexecute "github.com/alexellis/go-execute/pkg/v1"
	"io"
	"os"
	"strconv"
)

type LocalOperator struct {
}

func NewLocalOperator() *LocalOperator {
	return &LocalOperator{}
}

func (e LocalOperator) Execute(command string) (CommandRes, error) {
	task := goexecute.ExecTask{
		Command:     command,
		Shell:       true,
		StreamStdio: true,
	}

	res, err := task.Execute()
	if err != nil {
		return CommandRes{}, err
	}

	return CommandRes{
		StdErr: []byte(res.Stderr),
		StdOut: []byte(res.Stdout),
	}, nil
}

func (e LocalOperator) UploadFile(path string, remotePath string, mode string) error {
	source, err := os.Open(expandPath(path))
	if err != nil {
		return err
	}
	defer source.Close()

	return e.Upload(source, remotePath, mode)
}

func (e LocalOperator) Upload(source io.Reader, remotePath string, mode string) error {
	permissions, err := strconv.ParseInt(mode, 10, 32)
	if err != nil {
		return err
	}

	destination, err := os.OpenFile(remotePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(permissions))
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return err
}
