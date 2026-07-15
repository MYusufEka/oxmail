package mail

import (
	"fmt"
	"os/exec"
)

// CommandExecutor abstracts command execution for testability.
// Both PostfixManager and DovecotManager use this interface to run system commands.
type CommandExecutor interface {
	Run(name string, args ...string) error
	// RunWithOutput runs a command and returns stdout.
	RunWithOutput(name string, args ...string) (string, error)
}

// ExecCommandExecutor is the real implementation that executes system commands.
type ExecCommandExecutor struct{}

// Run executes a system command and returns any error.
func (e *ExecCommandExecutor) Run(name string, args ...string) error {
	return e.exec(name, args...).Run()
}

// RunWithOutput runs a command and returns stdout.
func (e *ExecCommandExecutor) RunWithOutput(name string, args ...string) (string, error) {
	cmd := e.exec(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w\n%s", name, err, string(out))
	}
	return string(out), nil
}

func (e *ExecCommandExecutor) exec(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// DockerExecExecutor runs commands inside another container via `docker exec`.
// Used because postfix/dovecot daemons run in separate containers (different PID namespaces).
type DockerExecExecutor struct {
	ContainerName string
}

// Run executes a command inside the container via `docker exec`.
func (e *DockerExecExecutor) Run(name string, args ...string) error {
	_, err := e.RunWithOutput(name, args...)
	return err
}

// RunWithOutput executes a command inside the container and returns stdout.
func (e *DockerExecExecutor) RunWithOutput(name string, args ...string) (string, error) {
	dockerArgs := append([]string{"exec", e.ContainerName, name}, args...)
	cmd := exec.Command("docker", dockerArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker exec %s %s: %w\n%s", e.ContainerName, name, err, string(out))
	}
	return string(out), nil
}
