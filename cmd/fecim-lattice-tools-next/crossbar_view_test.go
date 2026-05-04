//go:build !cgo

package main

import (
	"testing"

	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildCrossbarView(t *testing.T) {
	vm := crossbarvm.New(4, 4)
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildCrossbarView(snapshot, theme)
	if w == nil {
		t.Fatal("buildCrossbarView returned nil")
	}
}
