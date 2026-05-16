//go:build !cgo

package main

import (
	"log"

	"fecim-lattice-tools/internal/gogpuapp"
)

func main() {
	if err := gogpuapp.Run(gogpuapp.Options{}); err != nil {
		log.Fatal(err)
	}
}
