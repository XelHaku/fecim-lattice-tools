package edalattice

import (
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
