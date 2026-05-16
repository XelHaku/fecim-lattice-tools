package main

import (
	"fmt"
	"os"

	"fecim-lattice-tools/internal/gogpuscreenshot"
)

func main() {
	if err := gogpuscreenshot.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
