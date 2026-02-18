package edagui

import (
	"testing"
)

// TestRunSignature verifies the Run entrypoint has the canonical
// func([]string) error signature expected by the top-level CLI dispatcher.
func TestRunSignature(t *testing.T) {
	var _ func([]string) error = Run
}
