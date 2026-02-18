package mnistgui

import (
	"testing"
)

// TestRunSignature verifies the Run entrypoint has the canonical
// func([]string) error signature expected by the top-level CLI dispatcher.
// This is a compile-time contract check: if the signature changes the test
// file will not compile, catching breaking changes early.
func TestRunSignature(t *testing.T) {
	var _ func([]string) error = Run
}
