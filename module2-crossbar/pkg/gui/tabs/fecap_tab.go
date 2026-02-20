package tabs

// LIT-P2-05: FeCAP-specific GUI visualizations.
//
// Displays charge integration (Q = C × V) and displacement current
// (I_disp = ΔQ/Δt) for a ferroelectric capacitor crossbar array.
//
// Educational contrast with FeFET mode:
//
//	FeFET: I = G × V  (continuous DC, sneak paths, IR drop)
//	FeCAP: Q = C × V  (transient charge, DC-blocked, no sneak paths)

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/crossbar"
)

// FeCAPTab provides the FeCAP charge-domain MVM visualization.
type FeCAPTab struct {
	array *crossbar.Array
	rows  int
	cols  int

	chargeLabel *widget.Label // Per-column Q[col] bar chart
	dispLabel   *widget.Label // Displacement current I_disp
	statusLabel *widget.Label

	lastInput  []float64 // Last input voltages (V)
	lastCharge []float64 // Last computed charge vector (C)
}

// NewFeCAPTab creates a FeCAP visualization tab with an internal FeCAP array.
func NewFeCAPTab(arraySize int) *FeCAPTab {
	cfg := crossbar.DefaultFeCAPConfig(arraySize, arraySize)
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		arr, _ = crossbar.NewArray(crossbar.DefaultFeCAPConfig(4, 4))
	}

	tab := &FeCAPTab{
		array:       arr,
		rows:        arr.Rows(),
		cols:        arr.Cols(),
		chargeLabel: widget.NewLabel("Program capacitances and run Charge MVM to see Q[col]."),
		dispLabel:   widget.NewLabel(""),
		statusLabel: widget.NewLabel("Ready — FeCAP mode (charge-domain)"),
	}
	tab.chargeLabel.TextStyle = fyne.TextStyle{Monospace: true}
	tab.chargeLabel.Wrapping = fyne.TextWrapWord
	tab.dispLabel.TextStyle = fyne.TextStyle{Monospace: true}
	tab.dispLabel.Wrapping = fyne.TextWrapWord
	return tab
}

// Content returns the tab's Fyne UI content.
func (t *FeCAPTab) Content() fyne.CanvasObject {
	// --- Left panel: array info and educational text ---
	physicsText := fmt.Sprintf(
		"Array: %d×%d FeCAP cells\n"+
			"C_min = 0.5 fF  C_max = 2.0 fF\n\n"+
			"MVM physics:\n"+
			"  Q[j] = Σᵢ C[i,j] × V_DAC[i]\n"+
			"  (charge-domain, not current)\n\n"+
			"Sensing:\n"+
			"  V_out = Q / C_fb  (charge amp)\n"+
			"  C_fb = 128 fF (default)\n\n"+
			"Pulse duration: %.0f ns\n\n"+
			"Key advantage over FeFET:\n"+
			"  • No DC path → no sneak paths\n"+
			"  • No IR drop (capacitor blocks DC)\n"+
			"  • Energy: E = ½CV² per cell\n"+
			"  • 14-57× lower energy per MVM\n"+
			"    (Adv. Intell. Syst. 2022)",
		t.rows, t.cols,
		crossbar.DefaultFeCAPPulseDuration*1e9,
	)
	physicsLabel := widget.NewLabel(physicsText)
	physicsLabel.TextStyle = fyne.TextStyle{Monospace: true}
	physicsLabel.Wrapping = fyne.TextWrapWord

	leftContent := container.NewBorder(
		widget.NewLabelWithStyle("FeCAP Array (Charge Domain)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		container.NewVScroll(physicsLabel),
	)

	// --- Right panel: controls + results ---
	programRandomBtn := widget.NewButton("Program Random Capacitances", func() {
		for r := 0; r < t.rows; r++ {
			for c := 0; c < t.cols; c++ {
				_ = t.array.ProgramCapacitance(r, c, rand.Float64())
			}
		}
		t.statusLabel.SetText(fmt.Sprintf("Programmed %d×%d random capacitances (w ∈ [0,1])", t.rows, t.cols))
	})

	programSatBtn := widget.NewButton("Program Full Saturation (+Pr)", func() {
		for r := 0; r < t.rows; r++ {
			for c := 0; c < t.cols; c++ {
				_ = t.array.ProgramCapacitance(r, c, 1.0)
			}
		}
		t.statusLabel.SetText("All cells at C_max (w=1.0, +Pr state)")
	})

	clearBtn := widget.NewButton("Clear Array (−Pr, C_min)", func() {
		for r := 0; r < t.rows; r++ {
			for c := 0; c < t.cols; c++ {
				_ = t.array.ProgramCapacitance(r, c, 0.0)
			}
		}
		t.chargeLabel.SetText("Array cleared — all cells at C_min (w=0)")
		t.dispLabel.SetText("")
		t.statusLabel.SetText("Cleared — cells at C_min = 0.5 fF")
	})

	runMVMBtn := widget.NewButton("Run Charge MVM", func() {
		t.runChargeMVM()
	})
	runMVMBtn.Importance = widget.HighImportance

	chargeScroll := container.NewVScroll(t.chargeLabel)
	chargeScroll.SetMinSize(fyne.NewSize(300, 130))

	dispScroll := container.NewVScroll(t.dispLabel)
	dispScroll.SetMinSize(fyne.NewSize(300, 110))

	controls := container.NewVBox(
		widget.NewLabel("Array Programming:"),
		programRandomBtn,
		programSatBtn,
		clearBtn,
		widget.NewSeparator(),
		runMVMBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Charge Accumulation  Q[col]", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		chargeScroll,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Displacement Current  I_disp = ΔQ/Δt", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dispScroll,
	)
	controlsScroll := container.NewVScroll(controls)

	content := container.NewHSplit(
		container.NewBorder(
			nil,
			t.statusLabel,
			nil, nil,
			leftContent,
		),
		controlsScroll,
	)
	content.SetOffset(0.40) // 40% charge display, 60% controls

	return content
}

// runChargeMVM runs a charge-domain MVM with random inputs and updates the display.
func (t *FeCAPTab) runChargeMVM() {
	input := make([]float64, t.rows)
	for i := range input {
		input[i] = rand.Float64()
	}
	t.lastInput = input

	charge, err := t.array.MVMCharge(input)
	if err != nil {
		t.statusLabel.SetText("Error: " + err.Error())
		return
	}
	t.lastCharge = charge

	energy := t.array.MVMChargeEnergy(input)
	t.chargeLabel.SetText(t.formatChargeAccumulation(charge))
	t.dispLabel.SetText(t.formatDisplacementCurrent(charge))
	t.statusLabel.SetText(fmt.Sprintf(
		"Charge MVM done — Total E = %.3f fJ  (no sneak-path or IR-drop losses)",
		energy*1e15,
	))
}

// formatChargeAccumulation renders per-column charge Q[col] as a bar chart.
func (t *FeCAPTab) formatChargeAccumulation(charge []float64) string {
	if len(charge) == 0 {
		return "(no data)"
	}
	maxQ := 0.0
	for _, q := range charge {
		if math.Abs(q) > maxQ {
			maxQ = math.Abs(q)
		}
	}
	if maxQ == 0 {
		return "All cells at zero capacitance — Q[col] = 0"
	}

	const barWidth = 14
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-4s  %-10s  %s\n", "Col", "Q [fC]", "Charge Bar"))

	limit := len(charge)
	truncated := false
	if limit > 8 {
		limit = 8
		truncated = true
	}
	for c := 0; c < limit; c++ {
		q := charge[c]
		filled := int(math.Round(math.Abs(q) / maxQ * barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		sb.WriteString(fmt.Sprintf("%-4d  %8.3f fC  %s\n", c, q*1e15, bar))
	}
	if truncated {
		sb.WriteString(fmt.Sprintf("  … %d more cols (total %d)\n", len(charge)-limit, len(charge)))
	}
	sb.WriteString(fmt.Sprintf("\nMax Q = %.3f fC  (C_max×V_max×%d rows)",
		maxQ*1e15, t.rows))
	return sb.String()
}

// formatDisplacementCurrent shows I_disp = ΔQ/Δt for the configured pulse duration.
func (t *FeCAPTab) formatDisplacementCurrent(charge []float64) string {
	const pulseDur = crossbar.DefaultFeCAPPulseDuration // 10 ns

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("t_pulse = %.0f ns  →  I_disp = Q / t_pulse\n\n", pulseDur*1e9))
	sb.WriteString(fmt.Sprintf("%-4s  %-12s  %s\n", "Col", "I_disp [µA]", "Current Bar"))

	maxI := 0.0
	currents := make([]float64, len(charge))
	for c, q := range charge {
		i := math.Abs(q) / pulseDur
		currents[c] = i
		if i > maxI {
			maxI = i
		}
	}

	const barW = 10
	limit := len(currents)
	truncated := false
	if limit > 8 {
		limit = 8
		truncated = true
	}
	for c := 0; c < limit; c++ {
		i := currents[c]
		filled := 0
		if maxI > 0 {
			filled = int(math.Round(i / maxI * barW))
		}
		bar := strings.Repeat("▪", filled) + strings.Repeat("·", barW-filled)
		sb.WriteString(fmt.Sprintf("%-4d  %9.2f µA  %s\n", c, i*1e6, bar))
	}
	if truncated {
		sb.WriteString(fmt.Sprintf("  … %d more cols\n", len(currents)-limit))
	}
	sb.WriteString("\nAfter pulse: I_disp → 0  (capacitor blocks DC)\n")
	sb.WriteString("FeFET contrast: I = G×V persists → sneak paths")
	return sb.String()
}
