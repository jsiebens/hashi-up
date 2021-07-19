package operator

import (
	"io"
	"os"
	"strconv"

	goexecute "github.com/alexellis/go-execute/pkg/v1"
)

type LocalOperator struct {
}

func NewLocalOperator() *LocalOperator {
	return &LocalOperator{}
}

func (e LocalOperator) Execute(command string) error {
	task := goexecute.ExecTask{
		Command:     command,
		Shell:       true,
		StreamStdio: true,
	}

	_, err := task.Execute()
	if err != nil {
		return err
	}

	return nil
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
	permissions, err := strconv.ParseInt(mode, 8, 32)
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
