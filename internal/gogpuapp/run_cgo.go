//go:build cgo

package gogpuapp

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

type Options struct {
	ActiveModuleID viewmodel.ModuleID
}

func Run(options Options) error {
	return fmt.Errorf("gogpu/ui app requires CGO_ENABLED=0; rebuild with CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools")
}
