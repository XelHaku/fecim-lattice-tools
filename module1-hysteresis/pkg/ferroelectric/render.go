// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"fmt"
	"math"
	"strings"
)

// PERenderer renders P-E hysteresis loops and related visualizations.
type PERenderer struct {
	Width  int  // Character width for plots
	Height int  // Character height for plots
	Color  bool // Use ANSI colors
}

// NewPERenderer creates a new renderer with default settings.
func NewPERenderer() *PERenderer {
	return &PERenderer{
		Width:  60,
		Height: 25,
		Color:  true,
	}
}

// RenderPELoop renders a P-E hysteresis loop as ASCII art.
func (r *PERenderer) RenderPELoop(E, P []float64, material *HZOMaterial) string {
	var sb strings.Builder

	// Find bounds
	Emax := 0.0
	Pmax := 0.0
	for i := range E {
		if math.Abs(E[i]) > Emax {
			Emax = math.Abs(E[i])
		}
		if math.Abs(P[i]) > Pmax {
			Pmax = math.Abs(P[i])
		}
	}

	// Create plot grid
	grid := make([][]rune, r.Height)
	for i := range grid {
		grid[i] = make([]rune, r.Width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Draw axes
	midY := r.Height / 2
	midX := r.Width / 2

	for x := 0; x < r.Width; x++ {
		grid[midY][x] = '─'
	}
	for y := 0; y < r.Height; y++ {
		grid[y][midX] = '│'
	}
	grid[midY][midX] = '┼'

	// Plot the hysteresis loop
	for i := range E {
		// Map to grid coordinates
		x := int((E[i]/Emax + 1) / 2 * float64(r.Width-1))
		y := int((1 - (P[i]/Pmax+1)/2) * float64(r.Height-1))

		if x >= 0 && x < r.Width && y >= 0 && y < r.Height {
			grid[y][x] = '●'
		}
	}

	// Mark special points
	// Coercive field points
	EcX := int((material.Ec/Emax + 1) / 2 * float64(r.Width-1))
	if EcX >= 0 && EcX < r.Width {
		grid[midY][EcX] = '◆'
	}
	EcX = int((-material.Ec/Emax + 1) / 2 * float64(r.Width-1))
	if EcX >= 0 && EcX < r.Width {
		grid[midY][EcX] = '◆'
	}

	// Remanent polarization points
	PrY := int((1 - (material.Pr/Pmax+1)/2) * float64(r.Height-1))
	if PrY >= 0 && PrY < r.Height {
		grid[PrY][midX] = '◇'
	}
	PrY = int((1 - (-material.Pr/Pmax+1)/2) * float64(r.Height-1))
	if PrY >= 0 && PrY < r.Height {
		grid[PrY][midX] = '◇'
	}

	// Header
	sb.WriteString("P-E Hysteresis Loop:\n")
	sb.WriteString(strings.Repeat("═", r.Width+4) + "\n")

	// Top axis label
	sb.WriteString(fmt.Sprintf("  P (C/m²)  [Ps = %.2f C/m²]\n", material.Ps))
	sb.WriteString(fmt.Sprintf("     ↑ +%.2f\n", Pmax))

	// Render grid
	for _, row := range grid {
		sb.WriteString("  ")
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}

	// Bottom labels
	sb.WriteString(fmt.Sprintf("     ↓ -%.2f\n", Pmax))
	sb.WriteString(fmt.Sprintf("  -%.1e ← E (V/m) → +%.1e\n", Emax, Emax))
	sb.WriteString(fmt.Sprintf("  [Ec = %.1e V/m]\n", material.Ec))

	// Legend
	sb.WriteString("\n  Legend: ● Loop  ◆ Ec (coercive)  ◇ Pr (remanent)\n")

	return sb.String()
}

// RenderDomainStates renders the Preisach plane domain states.
func (r *PERenderer) RenderDomainStates(alphas, betas []float64, states []int) string {
	var sb strings.Builder

	// Find bounds
	maxAlpha := 0.0
	minBeta := 0.0
	for i := range alphas {
		if alphas[i] > maxAlpha {
			maxAlpha = alphas[i]
		}
		if betas[i] < minBeta {
			minBeta = betas[i]
		}
	}

	size := 30 // Grid size

	// Create grid
	grid := make([][]rune, size)
	for i := range grid {
		grid[i] = make([]rune, size)
		for j := range grid[i] {
			grid[i][j] = '·'
		}
	}

	// Draw the valid region boundary (α = β line)
	for i := 0; i < size; i++ {
		grid[size-1-i][i] = '╲'
	}

	// Plot hysteron states
	upCount := 0
	downCount := 0
	for i := range alphas {
		// Map to grid (β on x-axis, α on y-axis)
		x := int((betas[i] - minBeta) / (maxAlpha - minBeta) * float64(size-1))
		y := size - 1 - int((alphas[i]-minBeta)/(maxAlpha-minBeta)*float64(size-1))

		if x >= 0 && x < size && y >= 0 && y < size {
			if states[i] == +1 {
				grid[y][x] = '█'
				upCount++
			} else {
				grid[y][x] = '░'
				downCount++
			}
		}
	}

	sb.WriteString("Preisach Plane (Domain States):\n")
	sb.WriteString(strings.Repeat("═", size+4) + "\n")
	sb.WriteString("  α (up-switch field)\n")
	sb.WriteString("  ↑\n")

	for _, row := range grid {
		sb.WriteString("  ")
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}

	sb.WriteString("  └" + strings.Repeat("─", size-1) + "→ β (down-switch field)\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  █ Up (+1): %d    ░ Down (-1): %d\n", upCount, downCount))
	sb.WriteString(fmt.Sprintf("  Switched fraction: %.1f%%\n", float64(upCount)/float64(upCount+downCount)*100))

	return sb.String()
}

// RenderDiscreteStates renders the 30 programmable states.
func (r *PERenderer) RenderDiscreteStates(states []DiscreteState) string {
	var sb strings.Builder

	sb.WriteString("30 Discrete Analog States (FeCIM):\n")
	sb.WriteString(strings.Repeat("═", 65) + "\n\n")

	// Header
	sb.WriteString(fmt.Sprintf("%-6s %-12s %-10s %-12s %-12s\n",
		"Level", "P (C/m²)", "P/Ps", "Voltage", "Conductance"))
	sb.WriteString(strings.Repeat("─", 65) + "\n")

	// Show every 5th state for brevity
	for i := 0; i < len(states); i += 3 {
		s := states[i]
		bar := r.makeBar(s.NormalizedP, 20)
		sb.WriteString(fmt.Sprintf("%5d  %+.4f    %+.2f [%s]  %+.3f V   %.1f µS\n",
			s.Level, s.Polarization, s.NormalizedP, bar, s.Voltage*1e3, s.Conductance*1e6))
	}

	sb.WriteString(strings.Repeat("─", 65) + "\n")
	sb.WriteString(fmt.Sprintf("Total states: %d  (%.1f bits/cell)\n",
		len(states), math.Log2(float64(len(states)))))

	return sb.String()
}

// RenderSwitchingDynamics renders domain switching over time.
func (r *PERenderer) RenderSwitchingDynamics(times, pols []float64, switched []int, material *HZOMaterial) string {
	var sb strings.Builder

	sb.WriteString("Domain Switching Dynamics (KAI Model):\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	// Find max values
	maxT := times[len(times)-1]
	maxP := material.Ps
	maxSwitch := switched[len(switched)-1]

	height := 15
	width := 50

	// Create plot grid
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot polarization
	for i := range times {
		x := int(times[i] / maxT * float64(width-1))
		y := height - 1 - int(pols[i]/maxP*float64(height-1))
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = '●'
		}
	}

	// Draw axes
	for x := 0; x < width; x++ {
		grid[height-1][x] = '─'
	}
	for y := 0; y < height; y++ {
		grid[y][0] = '│'
	}
	grid[height-1][0] = '└'

	sb.WriteString(fmt.Sprintf("  P/Ps\n  ↑ %.2f\n", maxP))
	for _, row := range grid {
		sb.WriteString("  ")
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("  └%s→ t\n", strings.Repeat("─", width-1)))
	sb.WriteString(fmt.Sprintf("  0                              %.1f ns\n", maxT*1e9))

	// Statistics
	sb.WriteString(fmt.Sprintf("\n  Switching time (τ): %.2f ns\n", material.Tau*1e9))
	sb.WriteString(fmt.Sprintf("  Final polarization: %.3f C/m²\n", pols[len(pols)-1]))
	sb.WriteString(fmt.Sprintf("  Domains switched: %d\n", maxSwitch))

	return sb.String()
}

// RenderTemperatureDependence shows how properties vary with temperature.
func (r *PERenderer) RenderTemperatureDependence(material *HZOMaterial) string {
	var sb strings.Builder

	sb.WriteString("Temperature Dependence:\n")
	sb.WriteString(strings.Repeat("═", 60) + "\n\n")

	temps := []float64{200, 250, 300, 350, 400, 450, 500, 550, 600}

	sb.WriteString(fmt.Sprintf("%-8s %-15s %-15s %-15s\n",
		"T (K)", "Ec (MV/cm)", "Pr (µC/cm²)", "τ (ns)"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for _, T := range temps {
		Ec := material.CoerciveFieldAtTemp(T)
		Pr := material.PolarizationAtTemp(T)
		tau := material.SwitchingTime(T)

		EcMV := Ec / 1e8 // Convert to MV/cm (1 MV/cm = 1e8 V/m)
		PruC := Pr * 100 // Convert to µC/cm² (1 C/m² = 100 µC/cm²)
		tauNs := tau * 1e9

		bar := r.makeTempBar(T, material.CurieTemp)
		sb.WriteString(fmt.Sprintf("%-8.0f %-15.3f %-15.2f %-15.2f %s\n",
			T, EcMV, PruC, tauNs, bar))
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Curie Temperature: %.0f K (%.0f°C)\n",
		material.CurieTemp, material.CurieTemp-273))

	return sb.String()
}

// RenderMaterialComparison compares different HZO variants.
func (r *PERenderer) RenderMaterialComparison() string {
	var sb strings.Builder

	materials := AllMaterials()

	sb.WriteString("HZO Material Comparison:\n")
	sb.WriteString(strings.Repeat("═", 70) + "\n\n")

	sb.WriteString(fmt.Sprintf("%-20s %-12s %-12s %-12s %-10s\n",
		"Material", "Pr (µC/cm²)", "Ec (MV/cm)", "τ (ns)", "Endurance"))
	sb.WriteString(strings.Repeat("─", 70) + "\n")

	for _, m := range materials {
		PruC := m.Pr * 100 // C/m² to µC/cm²
		EcMV := m.Ec / 1e8 // V/m to MV/cm
		tauNs := m.Tau * 1e9
		endurance := fmt.Sprintf("%.0e", m.EnduranceCycles)

		sb.WriteString(fmt.Sprintf("%-20s %-12.1f %-12.2f %-12.2f %-10s\n",
			m.Name, PruC, EcMV, tauNs, endurance))
	}

	return sb.String()
}

// Helper functions

func (r *PERenderer) makeBar(value float64, width int) string {
	// Map -1 to +1 to bar
	pos := int((value + 1) / 2 * float64(width))
	if pos < 0 {
		pos = 0
	}
	if pos >= width {
		pos = width - 1
	}

	bar := make([]rune, width)
	for i := range bar {
		if i < pos {
			bar[i] = '█'
		} else {
			bar[i] = '░'
		}
	}
	return string(bar)
}

func (r *PERenderer) makeTempBar(T, Tc float64) string {
	width := 15
	ratio := T / Tc
	pos := int(ratio * float64(width))
	if pos >= width {
		pos = width - 1
	}

	bar := make([]rune, width)
	for i := range bar {
		if i <= pos {
			bar[i] = '█'
		} else {
			bar[i] = '░'
		}
	}
	return "[" + string(bar) + "]"
}
