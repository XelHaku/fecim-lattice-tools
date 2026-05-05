//go:build !cgo

package main

import (
	"testing"

	docsvm "fecim-lattice-tools/shared/viewmodel/docs"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildDocsView(t *testing.T) {
	vm := docsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildDocsView(snapshot, theme)
	if w == nil {
		t.Fatal("buildDocsView returned nil")
	}
}

func TestDocsBoundaryNotice(t *testing.T) {
	vm := docsvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in docs descriptor")
	}
}

func TestDocsSections(t *testing.T) {
	vm := docsvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Sections) == 0 {
		t.Error("Expected sections in docs snapshot")
	}
	foundHonesty := false
	for _, s := range snapshot.Sections {
		if s.ID == "honesty" {
			foundHonesty = true
		}
	}
	if !foundHonesty {
		t.Error("Expected honesty audit section in docs snapshot")
	}
}
