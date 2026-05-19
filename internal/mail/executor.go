package mail

import "os/exec"

// CommandExecutor abstracts command execution for testability.
// Both PostfixManager and DovecotManager use this interface to run system commands.
type CommandExecutor interface {
	Run(name string, args ...string) error
}

// ExecCommandExecutor is the real implementation that executes system commands.
type ExecCommandExecutor struct{}

// Run executes a system command and returns any error.
func (e *ExecCommandExecutor) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
