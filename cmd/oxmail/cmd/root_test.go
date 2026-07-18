package cmd

import (
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Oxmail")
	assertContains(t, output, "--json")
	assertContains(t, output, "--api-url")
	assertContains(t, output, "domain")
	assertContains(t, output, "user")
	assertContains(t, output, "alias")
	assertContains(t, output, "status")
	assertContains(t, output, "logs")
	assertContains(t, output, "send-test")
}

func TestRootCmd_JSONFlag(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("status", "--json", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "--json")
}

func TestRootCmd_APIURLFlag(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC("--api-url", "http://test:9999", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "--api-url")
}

func TestRootCmd_NoArgs(t *testing.T) {
	resetFlags()
	stdout, stderr, err := executeCommandC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout + stderr
	assertContains(t, output, "Oxmail")
	assertContains(t, output, "Usage:")
}

func TestRootCmd_UnknownFlag(t *testing.T) {
	resetFlags()
	_, _, err := executeCommandC("--unknown-flag")
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
