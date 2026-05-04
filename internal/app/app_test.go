package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, Version) {
		t.Fatalf("expected version output to contain %q, got %q", Version, got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunHelpOnEmptyArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run(nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "Usage:") {
		t.Fatalf("expected help output, got %q", got)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"bogus"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}

	if got := stderr.String(); !strings.Contains(got, "unknown command") {
		t.Fatalf("expected unknown command message, got %q", got)
	}
}
