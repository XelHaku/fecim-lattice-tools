package nonidealities

import (
	"fmt"
	"math"
	"strings"
)

// Renderer renders non-ideality visualizations as ASCII art.
type Renderer struct {
	UseColor bool
	Width    int
}

// NewRenderer creates a new renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		UseColor: true,
		Width:    60,
	}
}

// RenderIRDropMap renders the IR drop map.
func (r *Renderer) RenderIRDropMap(ir *IRDropSimulator) string {
	var sb strings.Builder

	sb.WriteString("IR Drop Map:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	if len(ir.IRDropMap) == 0 || len(ir.IRDropMap[0]) == 0 {
		return sb.String() + "No data available\n"
	}

	// Find max drop for scaling
	maxDrop := ir.GetMaxIRDrop()
	if maxDrop == 0 {
		maxDrop = 1
	}

	// Render header
	sb.WriteString("    ")
	for j := 0; j < ir.Cols && j < 32; j++ {
		sb.WriteString(fmt.Sprintf("%2d", j%10))
	}
	sb.WriteString(" Col\n")
	sb.WriteString("    " + strings.Repeat("──", min(ir.Cols, 32)) + "\n")

	chars := []rune{'·', '░', '▒', '▓', '█'}

	for i := 0; i < ir.Rows && i < 16; i++ {
		sb.WriteString(fmt.Sprintf("%2d │ ", i))
		for j := 0; j < ir.Cols && j < 32; j++ {
			drop := ir.IRDropMap[i][j]
			normalized := drop / maxDrop
			charIdx := int(normalized * float64(len(chars)-1))
			if charIdx >= len(chars) {
				charIdx = len(chars) - 1
			}
			sb.WriteRune(chars[charIdx])
			sb.WriteRune(' ')
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Row\n")

	// Legend
	sb.WriteString("\nLegend: · = minimal, ░ = low, ▒ = medium, ▓ = high, █ = maximum\n")
	sb.WriteString(fmt.Sprintf("Max IR Drop: %.3f mV\n", maxDrop*1000))

	return sb.String()
}

// RenderSneakPathMap renders sneak current distribution.
func (r *Renderer) RenderSneakPathMap(sp *SneakPathAnalyzer) string {
	var sb strings.Builder

	sb.WriteString("Sneak Current Map:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	if len(sp.SneakCurrents) == 0 {
		return sb.String() + "No data available\n"
	}

	// Find max sneak current for scaling
	maxSneak := 0.0
	for i := 0; i < sp.Rows; i++ {
		for j := 0; j < sp.Cols; j++ {
			if sp.SneakCurrents[i][j] > maxSneak {
				maxSneak = sp.SneakCurrents[i][j]
			}
		}
	}
	if maxSneak == 0 {
		maxSneak = 1
	}

	// Mark target cell
	sb.WriteString(fmt.Sprintf("Target Cell: (%d, %d)\n\n", sp.TargetRow, sp.TargetCol))

	sb.WriteString("    ")
	for j := 0; j < sp.Cols && j < 32; j++ {
		sb.WriteString(fmt.Sprintf("%2d", j%10))
	}
	sb.WriteString(" Col\n")
	sb.WriteString("    " + strings.Repeat("──", min(sp.Cols, 32)) + "\n")

	for i := 0; i < sp.Rows && i < 16; i++ {
		sb.WriteString(fmt.Sprintf("%2d │ ", i))
		for j := 0; j < sp.Cols && j < 32; j++ {
			if i == sp.TargetRow && j == sp.TargetCol {
				sb.WriteString("╳ ")
			} else {
				sneak := sp.SneakCurrents[i][j]
				normalized := sneak / maxSneak
				if normalized < 0.1 {
					sb.WriteString("· ")
				} else if normalized < 0.3 {
					sb.WriteString("░ ")
				} else if normalized < 0.6 {
					sb.WriteString("▒ ")
				} else if normalized < 0.9 {
					sb.WriteString("▓ ")
				} else {
					sb.WriteString("█ ")
				}
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Row\n")

	sb.WriteString("\n╳ = Target cell, intensity shows sneak current magnitude\n")

	return sb.String()
}

// RenderDriftHistory renders drift over time.
func (r *Renderer) RenderDriftHistory(d *DriftSimulator) string {
	var sb strings.Builder

	sb.WriteString("Conductance Drift History:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	if len(d.DriftHistory) == 0 {
		return sb.String() + "No history available\n"
	}

	// Find max drift for scaling
	maxDrift := 0.0
	for _, snap := range d.DriftHistory {
		if snap.MaxDrift > maxDrift {
			maxDrift = snap.MaxDrift
		}
	}
	if maxDrift == 0 {
		maxDrift = 1e-6
	}

	// Time axis
	maxTime := d.DriftHistory[len(d.DriftHistory)-1].Time
	sb.WriteString("Drift over Time:\n")
	sb.WriteString("│\n")

	// Graph height
	height := 10
	for h := height; h >= 0; h-- {
		threshold := float64(h) / float64(height) * maxDrift

		if h == height {
			sb.WriteString(fmt.Sprintf("%.1e │", maxDrift))
		} else if h == 0 {
			sb.WriteString(fmt.Sprintf("%.1e │", 0.0))
		} else {
			sb.WriteString("        │")
		}

		for _, snap := range d.DriftHistory {
			if snap.MaxDrift >= threshold {
				sb.WriteString("█")
			} else if snap.AvgDrift >= threshold {
				sb.WriteString("▒")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("        └" + strings.Repeat("─", len(d.DriftHistory)) + "\n")
	sb.WriteString(fmt.Sprintf("        0%s%.0fs\n",
		strings.Repeat(" ", len(d.DriftHistory)-5), maxTime))
	sb.WriteString("\n█ = Max drift, ▒ = Avg drift\n")

	return sb.String()
}

// RenderTechComparison renders technology comparison.
func (r *Renderer) RenderTechComparison(comparisons map[string]DriftStats) string {
	var sb strings.Builder

	sb.WriteString("Technology Drift Comparison:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	// Find max for scaling
	maxDrift := 0.0
	for _, stats := range comparisons {
		if stats.MaxDriftPercent > maxDrift {
			maxDrift = stats.MaxDriftPercent
		}
	}
	if maxDrift == 0 {
		maxDrift = 1
	}

	// Header
	sb.WriteString(fmt.Sprintf("%-20s %10s %10s %15s\n",
		"Technology", "Avg Drift", "Max Drift", "10yr Retention"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// Order matters for display
	order := []string{"FeCIM (FeFET)", "Flash", "RRAM", "PCM"}

	for _, name := range order {
		stats, ok := comparisons[name]
		if !ok {
			continue
		}

		barWidth := int(stats.MaxDriftPercent / maxDrift * 20)
		if barWidth < 1 {
			barWidth = 1
		}
		bar := strings.Repeat("█", barWidth)

		sb.WriteString(fmt.Sprintf("%-20s %9.3f%% %9.3f%% %14.2f%%\n",
			name, stats.AvgDriftPercent, stats.MaxDriftPercent, stats.RetentionPrediction))
		sb.WriteString(fmt.Sprintf("                     [%s%s]\n",
			bar, strings.Repeat(" ", 20-barWidth)))
	}

	// Highlight FeCIM advantage
	if fecimStats, ok := comparisons["FeCIM (FeFET)"]; ok {
		if rramStats, ok := comparisons["RRAM"]; ok {
			advantage := rramStats.MaxDriftPercent / fecimStats.MaxDriftPercent
			sb.WriteString(fmt.Sprintf("\nFeCIM advantage: %.0fx lower drift than RRAM!\n", advantage))
		}
	}

	return sb.String()
}

// RenderIRDropStats renders IR drop statistics.
func (r *Renderer) RenderIRDropStats(stats IRDropStats) string {
	var sb strings.Builder

	sb.WriteString("IR Drop Analysis:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Maximum IR Drop:     %.3f mV\n", stats.MaxIRDrop*1000))
	sb.WriteString(fmt.Sprintf("Average IR Drop:     %.3f mV\n", stats.AvgIRDrop*1000))
	sb.WriteString(fmt.Sprintf("Maximum Output Error: %.2f%%\n", stats.MaxOutputError))
	sb.WriteString(fmt.Sprintf("Average Output Error: %.2f%%\n", stats.AvgOutputError))
	sb.WriteString(fmt.Sprintf("Worst Cell:          (%d, %d)\n", stats.WorstCellRow, stats.WorstCellCol))

	// Visual severity indicator
	sb.WriteString("\nSeverity: ")
	if stats.MaxOutputError < 1 {
		sb.WriteString("[██████████] Excellent (<1% error)")
	} else if stats.MaxOutputError < 5 {
		sb.WriteString("[████████░░] Good (<5% error)")
	} else if stats.MaxOutputError < 10 {
		sb.WriteString("[██████░░░░] Acceptable (<10% error)")
	} else {
		sb.WriteString("[████░░░░░░] Needs Mitigation (>10% error)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// RenderSneakStats renders sneak path statistics.
func (r *Renderer) RenderSneakStats(stats SneakPathStats) string {
	var sb strings.Builder

	sb.WriteString("Sneak Path Analysis:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Target Current:      %.3f µA\n", stats.TargetCurrent*1e6))
	sb.WriteString(fmt.Sprintf("Total Sneak Current: %.3f µA\n", stats.TotalSneakCurrent*1e6))
	sb.WriteString(fmt.Sprintf("Sneak Ratio:         %.3f (%.1f%%)\n",
		stats.SneakRatio, stats.SneakRatio*100))
	sb.WriteString(fmt.Sprintf("Number of Paths:     %d\n", stats.NumSneakPaths))
	sb.WriteString(fmt.Sprintf("Worst Single Path:   %.3f µA\n", stats.WorstSneakPath*1e6))
	sb.WriteString(fmt.Sprintf("Signal-to-Noise:     %.1f dB\n", stats.SignalToNoiseRatio))

	// Visual severity
	sb.WriteString("\nSneak Current Level: ")
	if stats.SignalToNoiseRatio > 30 {
		sb.WriteString("[██████████] Excellent (SNR > 30dB)")
	} else if stats.SignalToNoiseRatio > 20 {
		sb.WriteString("[████████░░] Good (SNR > 20dB)")
	} else if stats.SignalToNoiseRatio > 10 {
		sb.WriteString("[██████░░░░] Acceptable (SNR > 10dB)")
	} else {
		sb.WriteString("[████░░░░░░] Needs Selector (SNR < 10dB)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// RenderDriftStats renders drift statistics.
func (r *Renderer) RenderDriftStats(stats DriftStats) string {
	var sb strings.Builder

	sb.WriteString("Conductance Drift Analysis:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Elapsed Time:        %.1f seconds\n", stats.ElapsedTime))
	sb.WriteString(fmt.Sprintf("Average Drift:       %.4f%% of Gmax\n", stats.AvgDriftPercent))
	sb.WriteString(fmt.Sprintf("Maximum Drift:       %.4f%% of Gmax\n", stats.MaxDriftPercent))
	sb.WriteString(fmt.Sprintf("Level Errors:        %d (%.4f%%)\n",
		stats.NumLevelErrors, stats.LevelErrorRate))
	sb.WriteString(fmt.Sprintf("10-Year Retention:   %.2f%%\n", stats.RetentionPrediction))

	// Technology comparison
	sb.WriteString("\nDrift Coefficient Comparison:\n")
	techs := []struct {
		name  string
		coeff float64
	}{
		{"FeCIM (FeFET)", stats.TechnologyComparison.FeFETDrift},
		{"Flash", stats.TechnologyComparison.FlashDrift},
		{"RRAM", stats.TechnologyComparison.RRAMDrift},
		{"PCM", stats.TechnologyComparison.PCMDrift},
	}

	maxCoeff := stats.TechnologyComparison.PCMDrift
	for _, tech := range techs {
		barWidth := int(tech.coeff / maxCoeff * 30)
		if barWidth < 1 {
			barWidth = 1
		}
		bar := strings.Repeat("█", barWidth) + strings.Repeat("░", 30-barWidth)
		sb.WriteString(fmt.Sprintf("  %-20s [%s] %.3f\n", tech.name, bar, tech.coeff))
	}

	sb.WriteString(fmt.Sprintf("\nFeCIM: %.0fx better retention than RRAM!\n",
		stats.TechnologyComparison.FeFETAdvantage))

	return sb.String()
}

// RenderMitigationComparison shows before/after mitigation.
func (r *Renderer) RenderMitigationComparison(before, after IRDropStats, strategy string) string {
	var sb strings.Builder

	sb.WriteString("Mitigation Comparison:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Strategy: %s\n\n", strategy))

	sb.WriteString(fmt.Sprintf("%-25s %12s %12s %12s\n",
		"Metric", "Before", "After", "Improvement"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// Max IR Drop
	improvement := (before.MaxIRDrop - after.MaxIRDrop) / before.MaxIRDrop * 100
	sb.WriteString(fmt.Sprintf("%-25s %10.3f mV %10.3f mV %10.1f%%\n",
		"Max IR Drop", before.MaxIRDrop*1000, after.MaxIRDrop*1000, improvement))

	// Avg IR Drop
	improvement = (before.AvgIRDrop - after.AvgIRDrop) / before.AvgIRDrop * 100
	sb.WriteString(fmt.Sprintf("%-25s %10.3f mV %10.3f mV %10.1f%%\n",
		"Avg IR Drop", before.AvgIRDrop*1000, after.AvgIRDrop*1000, improvement))

	// Max Error
	improvement = (before.MaxOutputError - after.MaxOutputError) / before.MaxOutputError * 100
	sb.WriteString(fmt.Sprintf("%-25s %10.2f %% %10.2f %% %10.1f%%\n",
		"Max Output Error", before.MaxOutputError, after.MaxOutputError, improvement))

	return sb.String()
}

// RenderSummary renders overall non-idealities summary.
func (r *Renderer) RenderSummary(irStats IRDropStats, sneakStats SneakPathStats, driftStats DriftStats) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║            NON-IDEALITIES ANALYSIS SUMMARY                   ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	// Overall assessment
	sb.WriteString("Overall Assessment:\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// IR Drop
	irSeverity := "PASS"
	if irStats.MaxOutputError > 10 {
		irSeverity = "NEEDS ATTENTION"
	}
	sb.WriteString(fmt.Sprintf("IR Drop:           Max %.3f mV, Error %.1f%% [%s]\n",
		irStats.MaxIRDrop*1000, irStats.MaxOutputError, irSeverity))

	// Sneak Paths
	sneakSeverity := "PASS"
	if sneakStats.SignalToNoiseRatio < 15 {
		sneakSeverity = "NEEDS ATTENTION"
	}
	sb.WriteString(fmt.Sprintf("Sneak Paths:       SNR %.1f dB, Ratio %.2f%% [%s]\n",
		sneakStats.SignalToNoiseRatio, sneakStats.SneakRatio*100, sneakSeverity))

	// Drift
	driftSeverity := "EXCELLENT"
	if driftStats.RetentionPrediction < 99 {
		driftSeverity = "GOOD"
	}
	sb.WriteString(fmt.Sprintf("Drift/Retention:   10yr retention %.2f%% [%s]\n",
		driftStats.RetentionPrediction, driftSeverity))

	sb.WriteString("\n")
	sb.WriteString("FeCIM Advantages:\n")
	sb.WriteString("  • Ferroelectric polarization enables excellent retention\n")
	sb.WriteString("  • 30 discrete levels provide noise margin\n")
	sb.WriteString("  • Low operating voltages reduce IR drop effects\n")
	sb.WriteString("  • Selector-free operation possible with proper biasing\n")

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper for absolute value
func abs(x float64) float64 {
	return math.Abs(x)
}
