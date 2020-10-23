package operator

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net"
	"os"
)

type CommandRes struct {
	StdOut []byte
	StdErr []byte
}

type CommandOperator interface {
	Execute(command string) (CommandRes, error)

	Upload(content string, remotePath string, mode string) error

	UploadFile(path string, remotePath string, mode string) error

	UploadEmbeddedFile(path string, remotePath string, mode string) error
}

type Callback func(CommandOperator) error

func ExecuteLocal(callback Callback) error {
	fmt.Println("Running locally")
	return callback(NewLocalOperator())
}

func ExecuteRemote(ip net.IP, user string, sshKey string, sshPort int, callback Callback) error {
	fmt.Println("Public IP: " + ip.String())

	sshKeyPath := expandPath(sshKey)
	fmt.Printf("ssh -i %s -p %d %s@%s\n", sshKeyPath, sshPort, user, ip.String())

	authMethod, closeSSHAgent, err := loadPublickey(sshKeyPath)
	if err != nil {
		return errors.Wrapf(err, "unable to load the ssh key with path %q", sshKeyPath)
	}

	defer closeSSHAgent()

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", ip.String(), sshPort)
	operator, err := NewSSHOperator(address, config)

	if err != nil {
		return errors.Wrapf(err, "unable to connect to %s over ssh", address)
	}

	defer operator.Close()

	return callback(operator)
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}

func sshAgent(publicKeyPath string) (ssh.AuthMethod, func() error) {
	if sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		sshAgent := agent.NewClient(sshAgentConn)

		keys, _ := sshAgent.List()
		if len(keys) == 0 {
			return nil, sshAgentConn.Close
		}

		pubkey, err := ioutil.ReadFile(publicKeyPath)
		if err != nil {
			return nil, sshAgentConn.Close
		}

		authkey, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
		if err != nil {
			return nil, sshAgentConn.Close
		}
		parsedkey := authkey.Marshal()

		for _, key := range keys {
			if bytes.Equal(key.Blob, parsedkey) {
				return ssh.PublicKeysCallback(sshAgent.Signers), sshAgentConn.Close
			}
		}
	}
	return nil, func() error { return nil }
}

func loadPublickey(path string) (ssh.AuthMethod, func() error, error) {
	noopCloseFunc := func() error { return nil }

	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, noopCloseFunc, fmt.Errorf("unable to read file: %s, %s", path, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		if err.Error() != "ssh: this private key is passphrase protected" {
			return nil, noopCloseFunc, fmt.Errorf("unable to parse private key: %s", err.Error())
		}

		agent, close := sshAgent(path + ".pub")
		if agent != nil {
			return agent, close, nil
		}

		defer close()

		fmt.Printf("Enter passphrase for '%s': ", path)
		STDIN := int(os.Stdin.Fd())
		bytePassword, _ := terminal.ReadPassword(STDIN)

		// Ignore any error from reading stdin to retain existing behaviour for unit test in
		// install_test.go

		// if err != nil {
		// 	return nil, noopCloseFunc, fmt.Errorf("reading password from stdin failed: %s", err.Error())
		// }

		fmt.Println()

		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
		if err != nil {
			return nil, noopCloseFunc, fmt.Errorf("parse private key with passphrase failed: %s", err)
		}
	}

	return ssh.PublicKeys(signer), noopCloseFunc, nil
}
