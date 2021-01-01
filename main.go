package main

import (
	"fmt"
	"github.com/jsiebens/hashi-up/cmd"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {

		switch err {
		case operator.SshAgentError:
			fmt.Print(sshAgentErrorMessage)
		case operator.TargetConnectError:
			fmt.Print(targetConnectError)
		default:
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

const sshAgentErrorMessage = `
There was an issue finding a private key. 
This could happen when hashi-up can not reach an authentication agent or when no private key is loaded.

How to fix this?

- check if an authentication agent is running and add a private key, e.g. 'ssh-add ~/.ssh/id_rsa'
- or add the '--ssh-target-key' flag to use a specific key, e.g. '--ssh-target-key ~/.ssh/id_rsa'

`

const targetConnectError = `
There was an issue connecting to your target host. 
This could happen when hashi-up can not reach the target host or when the private key authentication is invalid.

How to fix this?

- check if the target host is reachable and an SSH server is running
- check if the user and the private key are valid

`
