// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Architecture constants for crossbar array types.
const (
	Architecture1T1R = "1T1R (Transistor)"
	Architecture0T1R = "0T1R (Passive)"
	Architecture2T1R = "2T1R (Dual Transistor)"
)

// ArchitectureSelector is a shared widget for selecting crossbar architecture.
// It provides a dropdown to choose between 1T1R (transistor-isolated) and
// 0T1R (passive crossbar) architectures, which affects physics calculations
// for sneak paths and IR drop.
type ArchitectureSelector struct {
	widget.BaseWidget

	// Current selection
	architecture string

	// UI components
	label  *widget.Label
	select_ *widget.Select

	// Callback when architecture changes
	OnChanged func(architecture string)
}

// NewArchitectureSelector creates a new architecture selector widget.
func NewArchitectureSelector(onChanged func(architecture string)) *ArchitectureSelector {
	as := &ArchitectureSelector{
		architecture: Architecture1T1R, // Default to 1T1R
		OnChanged:    onChanged,
	}

	as.label = widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	as.select_ = widget.NewSelect([]string{Architecture1T1R, Architecture0T1R, Architecture2T1R}, func(s string) {
		as.architecture = s
		if as.OnChanged != nil {
			as.OnChanged(s)
		}
	})
	as.select_.SetSelected(Architecture1T1R)

	as.ExtendBaseWidget(as)
	return as
}

// CreateRenderer implements fyne.Widget.
func (as *ArchitectureSelector) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		widget.NewSeparator(),
		as.label,
		as.select_,
	)
	return widget.NewSimpleRenderer(content)
}

// GetArchitecture returns the currently selected architecture.
func (as *ArchitectureSelector) GetArchitecture() string {
	return as.architecture
}

// SetArchitecture sets the current architecture selection.
func (as *ArchitectureSelector) SetArchitecture(arch string) {
	as.architecture = arch
	as.select_.SetSelected(arch)
}

// Is1T1R returns true if the current selection is 1T1R architecture.
// 1T1R provides ~1000:1 sneak path isolation compared to passive 0T1R.
func (as *ArchitectureSelector) Is1T1R() bool {
	return as.architecture == "" ||
		as.architecture == Architecture1T1R ||
		as.architecture == "1T1R"
}

// Is2T1R returns true if the current selection is 2T1R architecture.
// 2T1R provides individual cell addressing with dual transistor control.
func (as *ArchitectureSelector) Is2T1R() bool {
	return as.architecture == Architecture2T1R ||
		as.architecture == "2T1R"
}

// GetSelect returns the underlying Select widget for direct access if needed.
func (as *ArchitectureSelector) GetSelect() *widget.Select {
	return as.select_
}

// ArchitectureInfo returns educational content about the selected architecture.
func ArchitectureInfo(arch string) (title, content string) {
	if arch == Architecture1T1R || arch == "1T1R" || arch == "" {
		return "1T1R Architecture",
			"1T1R = One Transistor per FeFET\n\n" +
				"How it works:\n" +
				"Transistor acts as controlled\n" +
				"switch, isolating unselected cells.\n\n" +
				"Advantages:\n" +
				"* ~1000:1 sneak isolation\n" +
				"* Linear I-V characteristics\n" +
				"* Industry standard (SRAM-like)\n\n" +
				"Tradeoffs:\n" +
				"* 50% area overhead\n" +
				"* More complex fabrication\n\n" +
				"Best for: High-precision inference\n" +
				"(vision, language models)"
	}
	if arch == Architecture2T1R || arch == "2T1R" {
		return "2T1R Architecture",
			"2T1R = Two Transistors per FeFET\n\n" +
				"How it works:\n" +
				"Row transistor (WL) + Column\n" +
				"transistor (CSL) provide AND-gate\n" +
				"selection for individual cells.\n\n" +
				"Advantages:\n" +
				"* Individual cell addressing\n" +
				"* Zero sneak paths\n" +
				"* No write disturb\n" +
				"* Precise programming\n\n" +
				"Tradeoffs:\n" +
				"* 3x area vs passive\n" +
				"* Extra CSL routing\n" +
				"* More complex control\n\n" +
				"Best for: Precise weight\n" +
				"programming, multi-level storage"
	}
	return "0T1R Architecture",
		"0T1R = Passive Crossbar (no transistor)\n\n" +
			"How it works:\n" +
			"Direct connection between wires.\n" +
			"FeFET is the only device.\n\n" +
			"Advantages:\n" +
			"* Highest density (4F^2 per cell)\n" +
			"* Simpler fabrication\n" +
			"* Lower cost\n\n" +
			"Tradeoffs:\n" +
			"* Sneak paths (~1% coupling)\n" +
			"* Requires selector device OR\n" +
			"  self-rectifying FeFET\n\n" +
			"FeFET advantage: Natural\n" +
			"rectification in HfO2-ZrO2!"
}
