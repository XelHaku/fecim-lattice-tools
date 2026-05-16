//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestSidebarBuildsForAllModules(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	w := buildSidebar(descriptors, 0)
	if w == nil {
		t.Fatal("buildSidebar returned nil")
	}
}

func TestSidebarActiveIndex(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	w := buildSidebar(descriptors, 2)
	if w == nil {
		t.Fatal("buildSidebar with activeIndex=2 returned nil")
	}
}

func TestSidebarMaterialBuildsForAllModules(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildSidebarMaterial(descriptors, 0, theme)
	if w == nil {
		t.Fatal("buildSidebarMaterial returned nil")
	}
}
