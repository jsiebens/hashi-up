package operator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

type CommandRes struct {
	StdOut []byte
	StdErr []byte
}

type CommandOperator interface {
	Execute(command string) (CommandRes, error)
	Upload(src io.Reader, remotePath string, mode string) error
	UploadFile(path string, remotePath string, mode string) error
}

type Callback func(CommandOperator) error

func ExecuteLocal(callback Callback) error {
	return callback(NewLocalOperator())
}

func ExecuteRemote(host string, user string, privateKey string, password string, callback Callback) error {
	var method ssh.AuthMethod

	if password != "" {
		method = ssh.Password(password)
	} else if privateKey == "" {
		sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))

		if err != nil {
			return SshAgentError
		}

		defer sshAgentConn.Close()

		client := agent.NewClient(sshAgentConn)
		list, err := client.List()

		if err != nil || len(list) == 0 {
			return SshAgentError
		}

		method = ssh.PublicKeysCallback(client.Signers)
	} else {
		buffer, err := ioutil.ReadFile(expandPath(privateKey))
		if err != nil {
			return errors.Wrapf(err, "unable to parse private key: %s", privateKey)
		}

		key, err := ssh.ParsePrivateKey(buffer)

		if err != nil {
			if err.Error() != "ssh: this private key is passphrase protected" {
				return errors.Wrapf(err, "unable to parse private key: %s", privateKey)
			}

			sshAgent, closeAgent := privateKeyUsingSSHAgent(privateKey + ".pub")
			defer closeAgent()

			if sshAgent != nil {
				method = sshAgent
			} else {
				fmt.Printf("Enter passphrase for '%s': ", privateKey)
				STDIN := int(os.Stdin.Fd())
				bytePassword, _ := terminal.ReadPassword(STDIN)
				fmt.Println()

				key, err = ssh.ParsePrivateKeyWithPassphrase(buffer, bytePassword)
				if err != nil {
					return errors.Wrapf(err, "parse private key with passphrase failed: %s", privateKey)
				}
				method = ssh.PublicKeys(key)
			}
		} else {
			method = ssh.PublicKeys(key)
		}
	}

	return executeRemote(host, user, method, callback)
}

func privateKeyUsingSSHAgent(publicKeyPath string) (ssh.AuthMethod, func() error) {
	if sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		sshAgent := agent.NewClient(sshAgentConn)

		signers, err := sshAgent.Signers()
		if err != nil || len(signers) == 0 {
			return nil, sshAgentConn.Close
		}

		pubkey, err := ioutil.ReadFile(expandPath(publicKeyPath))
		if err != nil {
			return nil, sshAgentConn.Close
		}

		authkey, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
		if err != nil {
			return nil, sshAgentConn.Close
		}
		parsedkey := authkey.Marshal()

		for _, signer := range signers {
			if bytes.Equal(signer.PublicKey().Marshal(), parsedkey) {
				return ssh.PublicKeys(signer), sshAgentConn.Close
			}
		}
	}
	return nil, func() error { return nil }
}

func executeRemote(address string, user string, authMethod ssh.AuthMethod, callback Callback) error {

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			host = address
			port = "22"
		} else {
			return fmt.Errorf("error splitting host/port: %w", err)
		}
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	operator, err := NewSSHOperator(net.JoinHostPort(host, port), config)

	if err != nil {
		return TargetConnectError
	}

	defer operator.Close()

	return callback(operator)
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}


