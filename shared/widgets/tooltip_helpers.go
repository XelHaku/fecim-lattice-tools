// Package widgets provides reusable UI components.
// tooltip_helpers.go provides helper functions for adding tooltips to existing widgets.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ════════════════════════════════════════════════════════════════════════════════
// TOOLTIP HELPER FUNCTIONS
// Easy-to-use functions for adding educational tooltips to GUI elements
// ════════════════════════════════════════════════════════════════════════════════

// WithInfoButton wraps content with an info button that shows tooltip on click.
// This is the recommended way to add tooltips for important parameters.
func WithInfoButton(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance
	return container.NewBorder(nil, nil, nil, infoBtn, content)
}

// WithInfoButtonLeft places the info button on the left side.
func WithInfoButtonLeft(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance
	return container.NewBorder(nil, nil, infoBtn, nil, content)
}

// LabelWithTooltip creates a label with an adjacent info button.
func LabelWithTooltip(text string, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	label := widget.NewLabel(text)
	return WithInfoButtonLeft(label, tc, window)
}

// SliderWithTooltip creates a complete slider row with label and info button.
func SliderWithTooltip(labelText string, slider *widget.Slider, valueLabel *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance

	label := widget.NewLabel(labelText)
	header := container.NewHBox(label, infoBtn)

	if valueLabel != nil {
		return container.NewVBox(
			container.NewBorder(nil, nil, header, valueLabel, nil),
			slider,
		)
	}
	return container.NewVBox(header, slider)
}

// SelectWithTooltip creates a select widget with info button.
func SelectWithTooltip(label string, options []string, onChanged func(string), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	sel := widget.NewSelect(options, onChanged)
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance

	labelWidget := widget.NewLabel(label)
	return container.NewBorder(nil, nil, container.NewHBox(labelWidget, infoBtn), nil, sel)
}

// ButtonWithTooltip creates a button with an info button next to it.
func ButtonWithTooltip(label string, onTapped func(), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	btn := widget.NewButton(label, onTapped)
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance
	return container.NewHBox(btn, infoBtn)
}

// EntryWithTooltip creates an entry widget with info button.
func EntryWithTooltip(placeholder string, tc TooltipContent, window fyne.Window) (*widget.Entry, fyne.CanvasObject) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance
	return entry, container.NewBorder(nil, nil, nil, infoBtn, entry)
}

// CheckWithTooltip creates a checkbox with info button.
func CheckWithTooltip(label string, onChanged func(bool), tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	check := widget.NewCheck(label, onChanged)
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance
	return container.NewHBox(check, infoBtn)
}

// ShowQuickTooltip shows a brief tooltip near the cursor (for hover effects).
// This creates a temporary popup that auto-hides.
func ShowQuickTooltip(text string, pos fyne.Position, canvas fyne.Canvas) *widget.PopUp {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	popup := widget.NewPopUp(container.NewPadded(label), canvas)
	popup.ShowAtPosition(pos)
	return popup
}

// CreateTooltipCard creates a card-style container for a section with an overall tooltip.
func CreateTooltipCard(title string, tc TooltipContent, window fyne.Window, content ...fyne.CanvasObject) *widget.Card {
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance

	// Add info button to the right of the title using subtitle hack
	card := widget.NewCard(title, "", container.NewVBox(content...))
	return card
}

// EducationalDialog shows a detailed explanation dialog with formatted content.
func EducationalDialog(title string, tc TooltipContent, window fyne.Window) {
	titleLabel := widget.NewLabelWithStyle(tc.Title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Description section
	descLabel := widget.NewLabel(tc.Description)
	descLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		descLabel,
	)

	// Range section
	if tc.Range != "" {
		rangeTitle := widget.NewLabelWithStyle("📊 Valid Range", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		rangeLabel := widget.NewLabel(tc.Range)
		rangeLabel.Wrapping = fyne.TextWrapWord
		content.Add(widget.NewSeparator())
		content.Add(rangeTitle)
		content.Add(rangeLabel)
	}

	// Physics section
	if tc.Physics != "" {
		physicsTitle := widget.NewLabelWithStyle("🔬 Physical Meaning", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		physicsLabel := widget.NewLabel(tc.Physics)
		physicsLabel.Wrapping = fyne.TextWrapWord
		content.Add(widget.NewSeparator())
		content.Add(physicsTitle)
		content.Add(physicsLabel)
	}

	// Tips section
	if len(tc.Tips) > 0 {
		tipsTitle := widget.NewLabelWithStyle("💡 Tips for New Users", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		content.Add(widget.NewSeparator())
		content.Add(tipsTitle)
		for _, tip := range tc.Tips {
			tipLabel := widget.NewLabel("• " + tip)
			tipLabel.Wrapping = fyne.TextWrapWord
			content.Add(tipLabel)
		}
	}

	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 350))

	dialog.ShowCustom(title, "Got it!", container.NewPadded(scroll), window)
}

// ════════════════════════════════════════════════════════════════════════════════
// SECTION BUILDERS
// Create labeled sections with tooltips for control panels
// ════════════════════════════════════════════════════════════════════════════════

// SectionWithTooltip creates a collapsible section with a tooltip for the header.
type SectionWithTooltip struct {
	widget.BaseWidget
	title    string
	tc       TooltipContent
	window   fyne.Window
	content  fyne.CanvasObject
	expanded bool
}

// NewSectionWithTooltip creates a new section with tooltip support.
func NewSectionWithTooltip(title string, tc TooltipContent, window fyne.Window, content fyne.CanvasObject) *SectionWithTooltip {
	s := &SectionWithTooltip{
		title:    title,
		tc:       tc,
		window:   window,
		content:  content,
		expanded: true,
	}
	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer implements fyne.Widget.
func (s *SectionWithTooltip) CreateRenderer() fyne.WidgetRenderer {
	infoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowTooltipDialog(s.tc, s.window)
	})
	infoBtn.Importance = widget.LowImportance

	titleLabel := widget.NewLabelWithStyle(s.title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewBorder(nil, nil, titleLabel, infoBtn, nil)

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		s.content,
	)

	return widget.NewSimpleRenderer(container.NewPadded(content))
}

// ════════════════════════════════════════════════════════════════════════════════
// COMPOSITE TOOLTIPS
// Tooltips that combine multiple parameters for overview explanations
// ════════════════════════════════════════════════════════════════════════════════

// ModuleOverviewTooltip provides a high-level overview for each module.
var ModuleOverviewTooltips = struct {
	Hysteresis TooltipContent
	Crossbar   TooltipContent
	MNIST      TooltipContent
	Circuits   TooltipContent
	Comparison TooltipContent
	EDA        TooltipContent
}{
	Hysteresis: TooltipContent{
		Title:       "Hysteresis Module Overview",
		Description: "Explore ferroelectric P-E curves and multi-level polarization switching. This module demonstrates the fundamental physics that enables FeCIM: the ability to store and read multiple analog states in a single ferroelectric cell.",
		Physics:     "Ferroelectric materials like HfO₂ (HZO) have two stable polarization states. By carefully controlling the applied electric field, we can access many intermediate states, each corresponding to a distinct conductance level for analog computing.",
		Tips: []string{
			"Model limitation: Preisach/LK options are reduced-order models (not full domain-field micromagnetics).",
			"Start with 'ISPP (Write/Read)' mode to see multi-level programming",
			"Use 'Manual' mode to explore the P-E curve interactively",
			"Try different materials to see how they affect the loop shape",
			"Watch the level indicator to see discrete quantization",
		},
	},
	Crossbar: TooltipContent{
		Title:       "Crossbar Module Overview",
		Description: "Visualize how FeCIM performs matrix-vector multiplication in memory. The crossbar architecture enables massively parallel analog computation using Ohm's law: current = conductance × voltage.",
		Physics:     "Each cell stores a weight as conductance G. Input voltages V applied to rows produce currents I = G×V. Columns sum these currents (Kirchhoff's law), computing the dot product in O(1) time regardless of matrix size.",
		Tips: []string{
			"Model limitation: IR-drop/sneak analyses use compact circuit abstractions, not full post-layout parasitic extraction.",
			"'Run MVM' to see matrix multiplication in action",
			"'Analyze IR Drop' shows voltage loss along wires",
			"'Analyze Sneak Paths' reveals parasitic current problems",
			"Try different architectures (0T1R, 1T1R) to see sneak path mitigation",
		},
	},
	MNIST: TooltipContent{
		Title:       "MNIST Module Overview",
		Description: "Test a real neural network running on simulated FeCIM hardware. Draw digits and watch the network recognize them, comparing ideal floating-point to quantized analog computation.",
		Physics:     "The network weights are stored as discrete conductance levels. Input pixels become voltages, and the crossbar computes weighted sums. ReLU activation and multiple layers classify the digit.",
		Tips: []string{
			"Model limitation: MNIST CIM path is inference-only and omits full training-time hardware adaptation.",
			"Draw thick, centered digits for best recognition",
			"Use 'Hardware' preset to see realistic accuracy",
			"'Noisy' preset shows degradation under stress",
			"Compare 'FP32 vs CIM' to see quantization effects",
		},
	},
	Circuits: TooltipContent{
		Title:       "Circuits Module Overview",
		Description: "Learn about the peripheral electronics that interface with the FeCIM array: DACs that generate input voltages, TIAs that sense output currents, and ADCs that digitize results.",
		Physics:     "Real CIM systems need precision analog circuits. DACs convert digital inputs to voltages. Crossbar computes analog sums. TIAs convert currents to voltages. ADCs digitize for further processing.",
		Tips: []string{
			"Model limitation: peripheral blocks are first-order behavioral models (not full transistor-level SPICE).",
			"Explore DAC operation: digital codes → analog voltages",
			"See how TIA gain affects current sensing",
			"Understand ADC resolution vs. noise trade-offs",
			"View timing diagrams for read/write sequences",
		},
	},
	Comparison: TooltipContent{
		Title:       "Comparison Module Overview",
		Description: "Compare FeCIM to conventional computing: CPU, GPU, DRAM, and NAND. Understand the energy efficiency advantage and see projected data center impact.",
		Physics:     "The 'memory wall' limits conventional computers: moving data costs 100-1000× more energy than computing. CIM eliminates data movement by computing where data lives, potentially saving orders of magnitude in energy.",
		Tips: []string{
			"Model limitation: comparison numbers are scenario projections, not measured product benchmarks.",
			"All FeCIM numbers are PROJECTIONS (TRL 4)",
			"Energy per MAC is the key efficiency metric",
			"Data center savings assume full technology maturation",
			"Compare different workloads to see scaling effects",
		},
	},
	EDA: TooltipContent{
		Title:       "EDA Module Overview",
		Description: "Introduction to electronic design automation for chip layout. This educational module shows how FeCIM arrays would be implemented in silicon using standard EDA tools.",
		Physics:     "Chip design uses hierarchical abstraction: transistors → cells → blocks → chip. Layout must satisfy manufacturing rules (DRC) and match intended connectivity (LVS).",
		Tips: []string{
			"Model limitation: EDA flow is educational and omits foundry sign-off corners and full DFM checks.",
			"This is educational, not tapeout-ready",
			"View standard cell layouts and I/O pad arrangements",
			"Learn industry formats: GDS, LEF, DEF",
			"See how FeCIM arrays fit into the chip floorplan",
		},
	},
}

// ShowModuleOverview displays the overview tooltip for a module.
func ShowModuleOverview(module string, window fyne.Window) {
	var tc TooltipContent
	switch module {
	case "hysteresis":
		tc = ModuleOverviewTooltips.Hysteresis
	case "crossbar":
		tc = ModuleOverviewTooltips.Crossbar
	case "mnist":
		tc = ModuleOverviewTooltips.MNIST
	case "circuits":
		tc = ModuleOverviewTooltips.Circuits
	case "comparison":
		tc = ModuleOverviewTooltips.Comparison
	case "eda":
		tc = ModuleOverviewTooltips.EDA
	default:
		return
	}
	EducationalDialog("Module Overview", tc, window)
}

// HelpButton creates a help button that shows the module overview.
func HelpButton(module string, window fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("Help", theme.HelpIcon(), func() {
		ShowModuleOverview(module, window)
	})
	btn.Importance = widget.LowImportance
	return btn
}

// QuickReferenceCard creates a compact card with key information for a parameter.
func QuickReferenceCard(tc TooltipContent) *widget.Card {
	summary := tc.Description
	if tc.Range != "" {
		summary += "\n\n📊 " + tc.Range
	}
	if len(tc.Tips) > 0 {
		summary += "\n\n💡 " + tc.Tips[0]
	}

	label := widget.NewLabel(summary)
	label.Wrapping = fyne.TextWrapWord

	return widget.NewCard(tc.Title, "", label)
}
