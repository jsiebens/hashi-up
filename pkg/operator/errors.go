package operator

import "fmt"

type TargetConnectError struct {
	reason error
}

func NewTargetConnectError(message error) *TargetConnectError {
	return &TargetConnectError{
		reason: message,
	}
}
func (e *TargetConnectError) Error() string {
	return fmt.Sprintf("%s", e.reason)
}

type SshAgentError struct {
	reason error
}

func NewSshAgentError(message error) *SshAgentError {
	return &SshAgentError{
		reason: message,
	}
}
func (e *SshAgentError) Error() string {
	return fmt.Sprintf("%s", e.reason)
}
