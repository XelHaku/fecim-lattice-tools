// Package gui provides overlay rendering for Lab Bench status display.
package gui

import (
	"fmt"
	"strings"
)

// Overlay renders Lab Bench status information as text overlays.
// This is designed to be drawn on top of the main visualization.
type Overlay struct {
	labBench *LabBench

	// Display options
	ShowHelp       bool
	ShowParameters bool
	ShowPhysics    bool

	// Physics state (updated from simulation)
	Polarization   float64
	Energy         float64
	CoerciveField  float64
	SimulationTime float64
}

// NewOverlay creates an overlay for the given Lab Bench.
func NewOverlay(lb *LabBench) *Overlay {
	return &Overlay{
		labBench:       lb,
		ShowParameters: true,
		ShowPhysics:    true,
		ShowHelp:       false,
	}
}

// ToggleHelp toggles the help display.
func (o *Overlay) ToggleHelp() {
	o.ShowHelp = !o.ShowHelp
}

// UpdatePhysics updates the physics state for display.
func (o *Overlay) UpdatePhysics(polarization, energy, ec, time float64) {
	o.Polarization = polarization
	o.Energy = energy
	o.CoerciveField = ec
	o.SimulationTime = time
}

// RenderText returns the overlay as formatted text lines.
// Each line is a string that can be rendered by the display system.
func (o *Overlay) RenderText() []string {
	var lines []string

	// Status bar at top
	lines = append(lines, o.renderStatusBar())
	lines = append(lines, "")

	// Parameters panel (left side)
	if o.ShowParameters {
		lines = append(lines, o.renderParametersPanel()...)
		lines = append(lines, "")
	}

	// Physics readout (right side, but appended here)
	if o.ShowPhysics {
		lines = append(lines, o.renderPhysicsPanel()...)
		lines = append(lines, "")
	}

	// Help at bottom (if enabled)
	if o.ShowHelp {
		lines = append(lines, o.labBench.ControlsHelp())
	} else {
		lines = append(lines, "Press H for help")
	}

	return lines
}

// renderStatusBar creates the top status line.
func (o *Overlay) renderStatusBar() string {
	status := "▶ RUNNING"
	if o.labBench.Paused {
		status = "⏸ PAUSED"
	}

	return fmt.Sprintf("IronLattice Lab Bench │ %s │ t = %.3f µs",
		status,
		o.SimulationTime*1e6,
	)
}

// renderParametersPanel creates the parameter display.
func (o *Overlay) renderParametersPanel() []string {
	lb := o.labBench

	// ASCII slider bar
	eFieldBar := renderSliderBar(lb.ElectricFieldAmplitude, lb.EFieldMin, lb.EFieldMax, 20)
	tempBar := renderSliderBar(lb.Temperature, lb.TempMin, lb.TempMax, 20)

	return []string{
		"┌─ CONTROL PARAMETERS ─────────────────┐",
		fmt.Sprintf("│ Electric Field: %s %.2f MV/cm │", eFieldBar, lb.ElectricFieldAmplitude/1e8),
		fmt.Sprintf("│ Temperature:    %s %.0f K      │", tempBar, lb.Temperature),
		fmt.Sprintf("│ Frequency:      %s                  │", formatFrequency(lb.Frequency)),
		fmt.Sprintf("│ Waveform:       %s                        │", lb.WaveformNames[lb.WaveformIndex]),
		"└──────────────────────────────────────┘",
	}
}

// renderPhysicsPanel creates the physics readout display.
func (o *Overlay) renderPhysicsPanel() []string {
	// Temperature effects
	effects := ComputeTemperatureEffects(o.labBench.Temperature)

	phaseStr := "Ferroelectric"
	if !effects.IsFerro {
		phaseStr = "Paraelectric"
	}

	// Polarization as bar
	polBar := renderPolarizationBar(o.Polarization, 20)

	return []string{
		"┌─ PHYSICS STATE ──────────────────────┐",
		fmt.Sprintf("│ Polarization: %s %+.3f    │", polBar, o.Polarization),
		fmt.Sprintf("│ Phase:        %s              │", phaseStr),
		fmt.Sprintf("│ α(T):         %+.2e C⁻²m²J      │", effects.AlphaT),
		fmt.Sprintf("│ T/Tc:         %.2f                    │", o.labBench.Temperature/effects.Tc),
		"└──────────────────────────────────────┘",
	}
}

// renderSliderBar creates an ASCII slider bar visualization.
func renderSliderBar(value, min, max float64, width int) string {
	if max <= min {
		return strings.Repeat("─", width)
	}

	normalized := (value - min) / (max - min)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	pos := int(normalized * float64(width-1))

	bar := make([]rune, width)
	for i := range bar {
		if i < pos {
			bar[i] = '█'
		} else if i == pos {
			bar[i] = '▓'
		} else {
			bar[i] = '░'
		}
	}

	return string(bar)
}

// renderPolarizationBar creates a centered polarization bar.
// Negative polarization extends left from center, positive extends right.
func renderPolarizationBar(normP float64, width int) string {
	// Clamp to [-1, 1]
	if normP < -1 {
		normP = -1
	}
	if normP > 1 {
		normP = 1
	}

	center := width / 2
	bar := make([]rune, width)

	// Fill with empty
	for i := range bar {
		bar[i] = '·'
	}

	// Mark center
	bar[center] = '│'

	// Fill based on polarization
	if normP < 0 {
		// Negative: fill left of center
		fillEnd := center
		fillStart := center + int(normP*float64(center))
		for i := fillStart; i < fillEnd; i++ {
			if i >= 0 && i < width {
				bar[i] = '▓'
			}
		}
	} else {
		// Positive: fill right of center
		fillStart := center + 1
		fillEnd := center + 1 + int(normP*float64(center))
		for i := fillStart; i < fillEnd; i++ {
			if i >= 0 && i < width {
				bar[i] = '▓'
			}
		}
	}

	return string(bar)
}

// ConsoleOverlay provides a simple console-based rendering for debugging.
type ConsoleOverlay struct {
	overlay *Overlay
}

// NewConsoleOverlay creates a console-based overlay renderer.
func NewConsoleOverlay(lb *LabBench) *ConsoleOverlay {
	return &ConsoleOverlay{
		overlay: NewOverlay(lb),
	}
}

// Render outputs the overlay to stdout.
func (co *ConsoleOverlay) Render() {
	// Clear screen (ANSI escape)
	fmt.Print("\033[H\033[2J")

	lines := co.overlay.RenderText()
	for _, line := range lines {
		fmt.Println(line)
	}
}

// UpdatePhysics passes through to the overlay.
func (co *ConsoleOverlay) UpdatePhysics(polarization, energy, ec, time float64) {
	co.overlay.UpdatePhysics(polarization, energy, ec, time)
}
