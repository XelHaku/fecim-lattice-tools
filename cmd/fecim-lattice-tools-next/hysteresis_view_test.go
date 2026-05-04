//go:build !cgo

package main

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
	points := defaultHysteresisLoop()
	if len(points) != 200 {
		t.Errorf("defaultHysteresisLoop len = %d, want 200", len(points))
	}
}
