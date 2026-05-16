//go:build !cgo

package gogpuapp

import (
	"testing"

	edavm "fecim-lattice-tools/shared/viewmodel/eda"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildEDAView(t *testing.T) {
	vm := edavm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildEDAView(snapshot, theme)
	if w == nil {
		t.Fatal("buildEDAView returned nil")
	}
}

func TestEDABoundaryNotice(t *testing.T) {
	vm := edavm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in EDA descriptor")
	}
}

func TestEDAExportFormats(t *testing.T) {
	vm := edavm.New()
	snapshot := vm.Snapshot()
	formats := map[string]bool{}
	for _, m := range snapshot.Metrics {
		if m.ID == "spice" || m.ID == "verilog" || m.ID == "liberty" || m.ID == "def" || m.ID == "lef" {
			formats[m.ID] = true
		}
	}
	if len(formats) < 5 {
		t.Error("Expected all 5 EDA export format metrics")
	}
}
