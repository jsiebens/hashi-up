package operator

type Error string

func (e Error) Error() string {
	return string(e)
}

const SshAgentError = Error("SshAgentError")
const TargetConnectError = Error("TargetConnectError")
