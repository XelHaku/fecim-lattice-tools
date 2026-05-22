package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRunSubcommandDispatchReportsUnknownWithoutExiting(t *testing.T) {
	var stderr bytes.Buffer

	handled, code := runSubcommandDispatch([]string{"not-a-subcommand"}, io.Discard, &stderr)

	if !handled {
		t.Fatal("runSubcommandDispatch handled = false, want true")
	}
	if code != 1 {
		t.Fatalf("runSubcommandDispatch code = %d, want 1", code)
	}
	text := stderr.String()
	if !strings.Contains(text, `unknown subcommand "not-a-subcommand"`) {
		t.Fatalf("stderr = %q, want unknown subcommand context", text)
	}
	if !strings.Contains(text, "Usage:") {
		t.Fatalf("stderr = %q, want root usage", text)
	}
}

func TestRunSubcommandDispatchReportsHysteresisFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	handled, code := runSubcommandDispatch([]string{"hysteresis", "-definitely-not-a-flag"}, &stdout, &stderr)

	if !handled {
		t.Fatal("runSubcommandDispatch handled = false, want true")
	}
	if code != 1 {
		t.Fatalf("runSubcommandDispatch code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want flag error context", text)
	}
	if !strings.Contains(text, "FeCIM Hysteresis") {
		t.Fatalf("stderr = %q, want hysteresis usage", text)
	}
	if !strings.Contains(text, "FeCIM Lattice Tools") {
		t.Fatalf("stderr = %q, want root usage", text)
	}
}

func TestRunMainReportsRootFlagErrorWithoutExiting(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runMain([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("runMain code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want flag error context", text)
	}
	if !strings.Contains(text, "Usage:") {
		t.Fatalf("stderr = %q, want root usage", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want top-level error prefix", text)
	}
}

func TestNormalizeEngine_TableDriven(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "preisach"},
		{"preisach", "preisach"},
		{"P", "preisach"},
		{" lk ", "lk"},
		{"landau", "lk"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		if got := normalizeEngine(tt.in); got != tt.want {
			t.Fatalf("normalizeEngine(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestCmdSkip_TableDriven(t *testing.T) {
	tests := []struct {
		args []string
		want int
	}{
		{nil, 0},
		{[]string{}, 0},
		{[]string{"gui"}, 1},
		{[]string{"cli"}, 0},
		{[]string{"gui", "-x"}, 1},
	}

	for _, tt := range tests {
		if got := cmdSkip(tt.args); got != tt.want {
			t.Fatalf("cmdSkip(%v)=%d want %d", tt.args, got, tt.want)
		}
	}
}
