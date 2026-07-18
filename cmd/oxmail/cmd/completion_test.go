package cmd

import "testing"

func TestCompletionCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("completion", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Generate the autocompletion script")
	assertContains(t, output, "bash")
	assertContains(t, output, "zsh")
	assertContains(t, output, "fish")
}

func TestCompletionCmd_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil || cmd.Name() != "completion" {
		t.Fatalf("expected completion command, got %#v", cmd)
	}
}

func TestCompletionCmd_InvalidShellArgs(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Args == nil {
		t.Fatal("expected completion args validator")
	}
}
