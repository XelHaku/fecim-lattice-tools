//go:build !cgo

package design

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestDesignTokens_Colors(t *testing.T) {
	if Primary != widget.Hex(0x2F5D50) {
		t.Errorf("Primary = %v, want 0x2F5D50", Primary)
	}
	if Surface != widget.Hex(0xF4F5F3) {
		t.Errorf("Surface = %v, want 0xF4F5F3", Surface)
	}
	if OnSurface != widget.Hex(0x1A1C1A) {
		t.Errorf("OnSurface = %v, want 0x1A1C1A", OnSurface)
	}
}

func TestDesignTokens_Spacing(t *testing.T) {
	if SidebarWidth != 240 {
		t.Errorf("SidebarWidth = %v, want 240", SidebarWidth)
	}
	if ContentPad != 24 {
		t.Errorf("ContentPad = %v, want 24", ContentPad)
	}
}
