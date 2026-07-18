package cmd

import (
	"strings"
	"testing"
)

func TestVersionCommand_NotRegistered(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("version")
	requireCmdErr(t, err, "version command is not registered in current CLI source")
}

func TestVersionCommand_NotInHelp(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if containsCommandName(stdout+stderr, "version") {
		t.Fatal("unexpected version command in root help")
	}
}

func containsCommandName(output, commandName string) bool {
	return strings.Contains(output, "  "+commandName+" ") || strings.Contains(output, "  "+commandName+"\t")
}
