package ssh

type CommandOperator interface {

	Execute(command string) (CommandRes, error)

	Upload(content string, remotePath string, permissions string) error

	UploadFile(path string, remotePath string, permissions string) error

}
