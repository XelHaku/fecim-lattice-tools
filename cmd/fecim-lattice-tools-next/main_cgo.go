//go:build cgo

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "fecim-lattice-tools-next uses gogpu/ui through the zero-CGO WebGPU stack.")
	fmt.Fprintln(os.Stderr, "Run with: CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next")
	os.Exit(2)
}
