//go:build !cgo

package gogpuapp

import (
	"testing"

	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildCircuitsView(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildCircuitsView(snapshot, theme)
	if w == nil {
		t.Fatal("buildCircuitsView returned nil")
	}
}

func TestCircuitsBoundaryNotice(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in circuits descriptor")
	}
}

func TestCircuitsMetrics(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Metrics) == 0 {
		t.Error("Expected metrics in circuits snapshot")
	}
	foundADC := false
	for _, m := range snapshot.Metrics {
		if m.ID == "adc" {
			foundADC = true
		}
	}
	if !foundADC {
		t.Error("Expected ADC metric in circuits snapshot")
	}
}
