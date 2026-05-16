//go:build !cgo

// Package design provides Material 3 design tokens and layout constants
// for the FeCIM Lattice Tools gogpu/ui shell.
package design

import "github.com/gogpu/ui/widget"

// Material 3 color tokens. Seed: deep green (#2F5D50).
var (
	Primary          = widget.Hex(0x2F5D50)
	PrimaryDark      = widget.Hex(0x1F463C)
	PrimaryLight     = widget.Hex(0x6F9C8D)
	Surface          = widget.Hex(0xF4F5F3)
	SurfaceContainer = widget.Hex(0xE8EBE7)
	OnSurface        = widget.Hex(0x1A1C1A)
	OnSurfaceVariant = widget.Hex(0x444744)
	Secondary        = widget.Hex(0x58685E)
	Error            = widget.Hex(0xBA1A1A)
)

// Layout spacing constants (pixels).
const (
	SidebarWidth = 240
	ContentPad   = 24
	CardGap      = 14
	SectionGap   = 10
	TopBarHeight = 52
)
