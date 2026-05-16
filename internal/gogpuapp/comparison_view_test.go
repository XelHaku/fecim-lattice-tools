//go:build !cgo

package gogpuapp

import (
	"testing"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
)

func TestBuildComparisonView_ReturnsNonNilWidget(t *testing.T) {
	snap := comparisonvm.New().Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	got := buildComparisonView(snap, theme)
	if got == nil {
		t.Fatal("buildComparisonView returned nil")
	}
}

func TestBuildComparisonView_HandlesEmptySnapshot(t *testing.T) {
	snap := viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:          viewmodel.ModuleComparison,
			Title:       "FeCIM Comparison",
			Description: "Evidence-first technology comparison.",
			Status:      viewmodel.StatusFunctional,
		},
	}
	theme := material3.New(widget.Hex(0x2F5D50))
	got := buildComparisonView(snap, theme)
	if got == nil {
		t.Fatal("buildComparisonView returned nil for empty snapshot")
	}
}

func TestBuildComparisonView_DoesNotPanicOnRealComparisonData(t *testing.T) {
	snap := comparisonvm.New().Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("buildComparisonView panicked: %v", r)
		}
	}()

	for i := 0; i < 5; i++ {
		_ = buildComparisonView(snap, theme)
	}
}
