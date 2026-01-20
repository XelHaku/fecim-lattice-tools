package comparison

import (
	"fmt"
	"math"
	"strings"
)

// Renderer renders comparison visualizations.
type Renderer struct {
	UseColor bool
	Width    int
}

// NewRenderer creates a new renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		UseColor: true,
		Width:    70,
	}
}

// RenderArchitectureSpecs renders architecture specifications.
func (r *Renderer) RenderArchitectureSpecs(archs []*Architecture) string {
	var sb strings.Builder

	sb.WriteString("Architecture Specifications:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	// Header
	sb.WriteString(fmt.Sprintf("%-22s %10s %10s %10s %12s\n",
		"Architecture", "Node", "Area", "TDP", "Peak TOPS"))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	for _, arch := range archs {
		sb.WriteString(fmt.Sprintf("%-22s %8.0f nm %7.0f mm² %7.0f W %10.1f\n",
			arch.Name, arch.ProcessNode, arch.ChipArea, arch.TDP, arch.PeakTOPS))
	}

	sb.WriteString("\n")

	// Efficiency comparison
	sb.WriteString("Efficiency Metrics:\n")
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %15s %15s\n",
		"Architecture", "TOPS/W", "TOPS/mm²"))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	for _, arch := range archs {
		sb.WriteString(fmt.Sprintf("%-22s %15.3f %15.4f\n",
			arch.Name, arch.TOPSPerWatt, arch.TOPSPerMM2))
	}

	return sb.String()
}

// RenderInferenceComparison renders inference performance comparison.
func (r *Renderer) RenderInferenceComparison(results []InferenceResult, workload Workload) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Inference Comparison: %s\n", workload.Name))
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Workload: %s\n", workload.Description))
	sb.WriteString(fmt.Sprintf("Operations: %s MACs\n", formatNumber(float64(workload.TotalOps))))
	sb.WriteString(fmt.Sprintf("Parameters: %s\n\n", formatNumber(float64(workload.Parameters))))

	// Results table
	sb.WriteString(fmt.Sprintf("%-22s %12s %15s %12s\n",
		"Architecture", "Latency", "Throughput", "Energy"))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	for _, res := range results {
		latStr := formatLatency(res.Latency)
		tpStr := formatThroughput(res.Throughput)
		enStr := formatEnergy(res.Energy)

		sb.WriteString(fmt.Sprintf("%-22s %12s %15s %12s\n",
			res.Architecture, latStr, tpStr, enStr))
	}

	return sb.String()
}

// RenderBarChart renders a horizontal bar chart for a metric.
func (r *Renderer) RenderBarChart(title string, labels []string, values []float64, unit string, lowerIsBetter bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s:\n", title))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	// Find max for scaling
	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	barWidth := 35

	for i, label := range labels {
		val := values[i]
		barLen := int(val / maxVal * float64(barWidth))
		if barLen < 1 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)

		// Format value
		valStr := ""
		if val >= 1e9 {
			valStr = fmt.Sprintf("%.2fB", val/1e9)
		} else if val >= 1e6 {
			valStr = fmt.Sprintf("%.2fM", val/1e6)
		} else if val >= 1e3 {
			valStr = fmt.Sprintf("%.2fK", val/1e3)
		} else if val >= 1 {
			valStr = fmt.Sprintf("%.2f", val)
		} else if val >= 0.001 {
			valStr = fmt.Sprintf("%.3f", val)
		} else {
			valStr = fmt.Sprintf("%.2e", val)
		}

		indicator := ""
		if lowerIsBetter && i == indexOfMin(values) {
			indicator = " ← Best"
		} else if !lowerIsBetter && i == indexOfMax(values) {
			indicator = " ← Best"
		}

		sb.WriteString(fmt.Sprintf("%-18s [%s] %s %s%s\n",
			label, bar, valStr, unit, indicator))
	}

	return sb.String()
}

// RenderDataCenterComparison renders data center scale comparison.
func (r *Renderer) RenderDataCenterComparison(metrics []DataCenterMetrics, targetThroughput float64) string {
	var sb strings.Builder

	sb.WriteString("Data Center Comparison:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	sb.WriteString(fmt.Sprintf("Target Throughput: %s inferences/sec\n\n", formatNumber(targetThroughput)))

	// Chips and power
	sb.WriteString(fmt.Sprintf("%-22s %10s %12s %10s\n",
		"Architecture", "Chips", "Power (kW)", "Rack Units"))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	for _, m := range metrics {
		sb.WriteString(fmt.Sprintf("%-22s %10d %12.1f %10d\n",
			m.Architecture, m.ChipsRequired, m.TotalPower, m.RackSpace))
	}

	sb.WriteString("\n")

	// Cost and emissions
	sb.WriteString(fmt.Sprintf("%-22s %15s %12s %12s\n",
		"Architecture", "Cost/Inference", "TCO/Year", "CO2/Day"))
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	for _, m := range metrics {
		sb.WriteString(fmt.Sprintf("%-22s %14.6f$ %11.0f$ %10.1f kg\n",
			m.Architecture, m.CostPerInference, m.TCO, m.CO2Emissions))
	}

	return sb.String()
}

// RenderAdvantages renders FeCIM advantages.
func (r *Renderer) RenderAdvantages(adv FeCIMAdvantage) string {
	var sb strings.Builder

	sb.WriteString("FeCIM Advantages:\n")
	sb.WriteString(strings.Repeat("═", r.Width) + "\n\n")

	// vs CPU
	sb.WriteString("vs Traditional CPU+DRAM:\n")
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	cpuAdvantages := []struct {
		name  string
		value float64
	}{
		{"Energy Reduction", adv.VsCPU.EnergyReduction},
		{"Latency Reduction", adv.VsCPU.LatencyReduction},
		{"Throughput Increase", adv.VsCPU.ThroughputIncrease},
		{"Power Reduction", adv.VsCPU.PowerReduction},
		{"TCO Reduction", adv.VsCPU.CostReduction},
	}

	for _, a := range cpuAdvantages {
		bar := r.makeAdvantageBar(a.value)
		sb.WriteString(fmt.Sprintf("  %-20s [%s] %.0fx\n", a.name, bar, a.value))
	}

	sb.WriteString("\n")

	// vs GPU
	sb.WriteString("vs GPU Accelerator:\n")
	sb.WriteString(strings.Repeat("─", r.Width) + "\n")

	gpuAdvantages := []struct {
		name  string
		value float64
	}{
		{"Energy Reduction", adv.VsGPU.EnergyReduction},
		{"Area Reduction", adv.VsGPU.AreaReduction},
		{"Power Reduction", adv.VsGPU.PowerReduction},
		{"TCO Reduction", adv.VsGPU.CostReduction},
	}

	for _, a := range gpuAdvantages {
		bar := r.makeAdvantageBar(a.value)
		sb.WriteString(fmt.Sprintf("  %-20s [%s] %.0fx\n", a.name, bar, a.value))
	}

	return sb.String()
}

// RenderSummary renders comparison summary.
func (r *Renderer) RenderSummary(comparison ComparisonResult, adv FeCIMAdvantage) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    COMPARISON SUMMARY                            ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	// Find FeCIM results
	var fecimResult InferenceResult
	var fecimDC DataCenterMetrics
	for i, arch := range comparison.Architectures {
		if arch.Name == "FeCIM CIM" {
			fecimResult = comparison.Results[i]
			fecimDC = comparison.DataCenter[i]
		}
	}

	sb.WriteString(fmt.Sprintf("Workload: %s (%s MACs)\n\n", comparison.Workload.Name,
		formatNumber(float64(comparison.Workload.TotalOps))))

	sb.WriteString("FeCIM Key Metrics:\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("  Inference Latency:    %s\n", formatLatency(fecimResult.Latency)))
	sb.WriteString(fmt.Sprintf("  Energy per Inference: %s\n", formatEnergy(fecimResult.Energy)))
	sb.WriteString(fmt.Sprintf("  Throughput:           %s\n", formatThroughput(fecimResult.Throughput)))
	sb.WriteString(fmt.Sprintf("  Data Center Power:    %.1f kW\n", fecimDC.TotalPower))
	sb.WriteString(fmt.Sprintf("  Annual TCO:           $%.0f\n", fecimDC.TCO))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("vs CPU:  %.0fx energy, %.0fx throughput, %.0fx lower TCO\n",
		adv.VsCPU.EnergyReduction, adv.VsCPU.ThroughputIncrease, adv.VsCPU.CostReduction))
	sb.WriteString(fmt.Sprintf("vs GPU:  %.0fx energy, %.0fx smaller, %.0fx lower power\n",
		adv.VsGPU.EnergyReduction, adv.VsGPU.AreaReduction, adv.VsGPU.PowerReduction))
	sb.WriteString("\n")

	// Dr. Tour quote
	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString("  \"This could lower the requirements in a data center\n")
	sb.WriteString("   by 80 to 90%.\" - Dr. external research group\n")
	sb.WriteString("─────────────────────────────────────────────────────────────\n")

	return sb.String()
}

func (r *Renderer) makeAdvantageBar(advantage float64) string {
	maxBar := 30
	// Log scale for better visualization
	logAdv := math.Log10(advantage)
	if logAdv < 0 {
		logAdv = 0
	}
	barLen := int(logAdv / 3.0 * float64(maxBar)) // 1000x = full bar
	if barLen > maxBar {
		barLen = maxBar
	}
	if barLen < 1 {
		barLen = 1
	}
	return strings.Repeat("█", barLen) + strings.Repeat("░", maxBar-barLen)
}

func formatNumber(n float64) string {
	if n >= 1e12 {
		return fmt.Sprintf("%.1fT", n/1e12)
	} else if n >= 1e9 {
		return fmt.Sprintf("%.1fB", n/1e9)
	} else if n >= 1e6 {
		return fmt.Sprintf("%.1fM", n/1e6)
	} else if n >= 1e3 {
		return fmt.Sprintf("%.1fK", n/1e3)
	}
	return fmt.Sprintf("%.0f", n)
}

func formatLatency(ms float64) string {
	if ms >= 1000 {
		return fmt.Sprintf("%.2f s", ms/1000)
	} else if ms >= 1 {
		return fmt.Sprintf("%.2f ms", ms)
	} else if ms >= 0.001 {
		return fmt.Sprintf("%.2f µs", ms*1000)
	}
	return fmt.Sprintf("%.2f ns", ms*1e6)
}

func formatThroughput(ips float64) string {
	if ips >= 1e6 {
		return fmt.Sprintf("%.2fM inf/s", ips/1e6)
	} else if ips >= 1e3 {
		return fmt.Sprintf("%.2fK inf/s", ips/1e3)
	}
	return fmt.Sprintf("%.2f inf/s", ips)
}

func formatEnergy(mj float64) string {
	if mj >= 1000 {
		return fmt.Sprintf("%.2f J", mj/1000)
	} else if mj >= 1 {
		return fmt.Sprintf("%.2f mJ", mj)
	} else if mj >= 0.001 {
		return fmt.Sprintf("%.2f µJ", mj*1000)
	}
	return fmt.Sprintf("%.2f nJ", mj*1e6)
}

func indexOfMin(values []float64) int {
	minIdx := 0
	minVal := values[0]
	for i, v := range values {
		if v < minVal {
			minVal = v
			minIdx = i
		}
	}
	return minIdx
}

func indexOfMax(values []float64) int {
	maxIdx := 0
	maxVal := values[0]
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}
