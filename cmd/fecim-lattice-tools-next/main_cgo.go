//go:build cgo

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "fecim-lattice-tools-next is a compatibility wrapper for the zero-CGO gogpu/ui shell.")
	fmt.Fprintln(os.Stderr, "Run with: CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools")
	os.Exit(2)
}
