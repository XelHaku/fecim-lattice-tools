//go:build !cgo

package main

import (
	"testing"

	mnistvm "fecim-lattice-tools/shared/viewmodel/mnist"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildMNISTView(t *testing.T) {
	vm := mnistvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildMNISTView(snapshot, theme)
	if w == nil {
		t.Fatal("buildMNISTView returned nil")
	}
}

func TestMNISTBoundaryNotice(t *testing.T) {
	vm := mnistvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in MNIST descriptor")
	}
}

func TestMNISTMetrics(t *testing.T) {
	vm := mnistvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Metrics) == 0 {
		t.Error("Expected metrics in MNIST snapshot")
	}
	foundAcc := false
	for _, m := range snapshot.Metrics {
		if m.ID == "accuracy" {
			foundAcc = true
		}
	}
	if !foundAcc {
		t.Error("Expected accuracy metric in MNIST snapshot")
	}
}
