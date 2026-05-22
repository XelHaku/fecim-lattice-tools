package edalattice

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsInvalidDimensions(t *testing.T) {
	err := Run([]string{"-rows", "0", "-cols", "4"})
	if err == nil {
		t.Fatal("expected error for invalid dimensions")
	}
	if !strings.Contains(err.Error(), "invalid lattice dimensions") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLatticeGenReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := runLatticeGen([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runLatticeGen error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM Lattice Generator") {
		t.Fatalf("stderr = %q, want usage", text)
	}
	if _, err := os.Stat(filepath.Join(home, ".fecim")); !os.IsNotExist(err) {
		t.Fatalf("invalid flag initialized logging directory; stat error = %v", err)
	}
}
