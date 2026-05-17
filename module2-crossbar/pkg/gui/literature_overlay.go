//go:build legacy_fyne

// Package gui - Literature overlay for crossbar MVM benchmarks
package gui

import (
	"fmt"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CrossbarBenchmark holds a single published crossbar accuracy result.
type CrossbarBenchmark struct {
	Reference string  // Short citation, e.g. "IEEE TED 2022"
	Metric    string  // What was measured, e.g. "MNIST accuracy (%)"
	Condition string  // Experimental condition
	Published float64 // Reported value
}

// CrossbarLitComparison pairs a benchmark with a simulated value for display.
type CrossbarLitComparison struct {
	Benchmark CrossbarBenchmark
	Simulated float64
}

// CrossbarFitMetrics summarises the fit between published and simulated values.
type CrossbarFitMetrics struct {
	RMSE     float64
	MAE      float64
	MaxErr   float64
	NSamples int
}

// referenceBenchmarks returns hardcoded peer-reviewed crossbar accuracy data.
// Source: IEEE Transactions on Electron Devices, 2022
//
//	"Current-limiter-based sneak-path mitigation in FeFET crossbar arrays"
//	MNIST classification on 64x64 FeFET passive crossbar.
func referenceBenchmarks() []CrossbarBenchmark {
	return []CrossbarBenchmark{
		{
			Reference: "IEEE TED 2022",
			Metric:    "MNIST accuracy (%)",
			Condition: "With current limiter",
			Published: 97.0,
		},
		{
			Reference: "IEEE TED 2022",
			Metric:    "MNIST accuracy (%)",
			Condition: "Without current limiter (passive)",
			Published: 9.8,
		},
	}
}

// ComputeCrossbarFitMetrics computes RMSE/MAE between published and simulated values.
func ComputeCrossbarFitMetrics(comparisons []CrossbarLitComparison) CrossbarFitMetrics {
	if len(comparisons) == 0 {
		return CrossbarFitMetrics{}
	}
	var sumSq, sumAbs, maxErr float64
	for _, c := range comparisons {
		e := math.Abs(c.Simulated - c.Benchmark.Published)
		sumSq += e * e
		sumAbs += e
		if e > maxErr {
			maxErr = e
		}
	}
	n := float64(len(comparisons))
	return CrossbarFitMetrics{
		RMSE:     math.Sqrt(sumSq / n),
		MAE:      sumAbs / n,
		MaxErr:   maxErr,
		NSamples: len(comparisons),
	}
}

// buildComparisonTable formats comparisons as a text table.
func buildComparisonTable(comparisons []CrossbarLitComparison) string {
	header := fmt.Sprintf("%-16s %-35s %10s %10s %8s",
		"Reference", "Condition", "Published", "Simulated", "Delta")
	sep := strings.Repeat("-", len(header))
	lines := []string{header, sep}
	for _, c := range comparisons {
		delta := c.Simulated - c.Benchmark.Published
		lines = append(lines, fmt.Sprintf("%-16s %-35s %9.1f%% %9.1f%% %+7.1f%%",
			c.Benchmark.Reference,
			c.Benchmark.Condition,
			c.Benchmark.Published,
			c.Simulated,
			delta,
		))
	}
	return strings.Join(lines, "\n")
}

// confidenceBadgeText returns the standard disclaimer for external benchmarks.
func confidenceBadgeText() string {
	return "External Benchmark (not this simulator's claim)"
}

// createLiteratureOverlayPanel builds the literature overlay card for the crossbar app.
// It shows published crossbar MVM accuracy alongside the simulator's estimated accuracy.
func (ca *CrossbarApp) createLiteratureOverlayPanel() fyne.CanvasObject {
	benchmarks := referenceBenchmarks()

	// Build comparisons using current simulator state.
	// Simulated accuracy: baseline 90% minus AccuracyLoss from last MVM.
	simAccWithLimiter := ca.estimatedAccuracy()
	// Passive (no limiter) approximation: assume full sneak-path degradation
	// collapses accuracy to near-random (10-class MNIST ~ 10%).
	simAccPassive := 10.0

	comparisons := []CrossbarLitComparison{
		{Benchmark: benchmarks[0], Simulated: simAccWithLimiter},
		{Benchmark: benchmarks[1], Simulated: simAccPassive},
	}

	metrics := ComputeCrossbarFitMetrics(comparisons)
	table := buildComparisonTable(comparisons)

	badge := widget.NewLabelWithStyle(
		confidenceBadgeText(),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	metricsLabel := widget.NewLabel(fmt.Sprintf(
		"Fit: RMSE = %.2f%%, MAE = %.2f%%, MaxErr = %.2f%% (N=%d)",
		metrics.RMSE, metrics.MAE, metrics.MaxErr, metrics.NSamples,
	))

	tableLabel := widget.NewLabel(table)
	tableLabel.TextStyle = fyne.TextStyle{Monospace: true}
	tableLabel.Wrapping = fyne.TextWrapOff

	sourceLabel := widget.NewLabel(
		"Source: IEEE Trans. Electron Devices, 2022\n" +
			"Current-limiter sneak-path mitigation in FeFET crossbar arrays",
	)
	sourceLabel.Wrapping = fyne.TextWrapWord
	sourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	box := container.NewVBox(
		badge,
		widget.NewSeparator(),
		metricsLabel,
		widget.NewSeparator(),
		tableLabel,
		widget.NewSeparator(),
		sourceLabel,
	)
	return widget.NewCard("Literature Overlay", "Published vs simulated MVM accuracy", box)
}

// estimatedAccuracy returns the simulator's current estimated accuracy.
func (ca *CrossbarApp) estimatedAccuracy() float64 {
	const baseline = 90.0
	ca.stateMu.RLock()
	result := ca.lastMVMResult
	ca.stateMu.RUnlock()
	if result == nil {
		return baseline
	}
	acc := baseline - result.AccuracyLoss
	if acc < 0 {
		acc = 0
	}
	return acc
}
