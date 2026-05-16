//go:build !cgo

package gogpuapp

import (
	"testing"

	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildHysteresisView(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildHysteresisView(snapshot, theme)
	if w == nil {
		t.Fatal("buildHysteresisView returned nil")
	}
}

func TestBuildHysteresisViewContainsMaterialSections(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Sections) == 0 {
		t.Error("No material sections in snapshot")
	}
}

func TestDefaultHysteresisLoop(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Plots) == 0 {
		t.Error("No plots in hysteresis snapshot")
		return
	}
	plot := snapshot.Plots[0]
	if len(plot.Series) == 0 {
		t.Error("No series in hysteresis plot")
		return
	}
	points := plot.Series[0].Points
	if len(points) != 200 {
		t.Errorf("Loop points len = %d, want 200", len(points))
	}
}

func TestHysteresisBoundaryNotice(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in hysteresis descriptor")
	}
}

func TestHysteresisPlotData(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	found := false
	for _, plot := range snapshot.Plots {
		if plot.ID == "pe_loop" {
			found = true
			if plot.XLabel == "" || plot.YLabel == "" {
				t.Error("Plot axis labels missing")
			}
		}
	}
	if !found {
		t.Error("No pe_loop plot found in hysteresis snapshot")
	}
}

func TestHysteresisRetention(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	found := false
	for _, plot := range snapshot.Plots {
		if plot.ID == "retention" {
			found = true
			if len(plot.Series) == 0 || len(plot.Series[0].Points) == 0 {
				t.Error("Retention plot has no data points")
			}
		}
	}
	if !found {
		t.Error("No retention plot found in hysteresis snapshot")
	}
}

func TestHysteresisComputedMetrics(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	hasPr := false
	hasEc := false
	for _, m := range snapshot.Metrics {
		if m.ID == "pr" {
			hasPr = true
		}
		if m.ID == "ec_plus" {
			hasEc = true
		}
	}
	if !hasPr {
		t.Error("No Pr metric in hysteresis snapshot")
	}
	if !hasEc {
		t.Error("No Ec metric in hysteresis snapshot")
	}
}
