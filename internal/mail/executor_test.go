package mail

import (
	"strings"
	"testing"
)

func TestExecCommandExecutor_Run(t *testing.T) {
	executor := &ExecCommandExecutor{}

	t.Run("run echo command succeeds", func(t *testing.T) {
		err := executor.Run("echo", "hello")
		if err != nil {
			t.Fatalf("Run(echo) = %v, want nil", err)
		}
	})

	t.Run("run unknown command fails", func(t *testing.T) {
		err := executor.Run("nonexistent-command-xyz")
		if err == nil {
			t.Fatal("Run(nonexistent) = nil, want error")
		}
	})
}

func TestExecCommandExecutor_RunWithOutput(t *testing.T) {
	executor := &ExecCommandExecutor{}

	t.Run("echo returns output", func(t *testing.T) {
		out, err := executor.RunWithOutput("echo", "hello world")
		if err != nil {
			t.Fatalf("RunWithOutput(echo) = _, %v, want nil", err)
		}
		out = strings.TrimSpace(out)
		if out != "hello world" {
			t.Errorf("output = %q, want %q", out, "hello world")
		}
	})

	t.Run("unknown command returns error", func(t *testing.T) {
		_, err := executor.RunWithOutput("nonexistent-command-xyz")
		if err == nil {
			t.Fatal("RunWithOutput(nonexistent) = nil, want error")
		}
	})
}

func TestDockerExecExecutor(t *testing.T) {
	// DockerExecExecutor wraps docker exec. Test it fails when container doesn't exist
	// (instead of requiring a running container).
	t.Run("Run fails gracefully without container", func(t *testing.T) {
		executor := &DockerExecExecutor{ContainerName: "nonexistent-test-container"}
		err := executor.Run("echo", "hello")
		if err == nil {
			t.Skip("docker available and running - skip non-docker env test")
			return
		}
		if !strings.Contains(err.Error(), "No such container") &&
			!strings.Contains(err.Error(), "not found") &&
			!strings.Contains(err.Error(), "Cannot connect") &&
			!strings.Contains(err.Error(), "connection refused") &&
			!strings.Contains(err.Error(), "docker") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
