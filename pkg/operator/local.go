package operator

import (
	goexecute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/markbates/pkger"
	"io"
	"io/ioutil"
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

func (e LocalOperator) Upload(content string, remotePath string, mode string) error {
	permissions, err := strconv.ParseInt(mode, 10, 32)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(remotePath, []byte(content), os.FileMode(permissions))
}

func (e LocalOperator) UploadFile(path string, remotePath string, mode string) error {
	permissions, err := strconv.ParseInt(mode, 10, 32)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(remotePath, content, os.FileMode(permissions))
}

func (e LocalOperator) UploadEmbeddedFile(path string, remotePath string, mode string) error {
	permissions, err := strconv.ParseInt(mode, 10, 32)
	if err != nil {
		return err
	}

	source, err := pkger.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(remotePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(permissions))
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return err
}
