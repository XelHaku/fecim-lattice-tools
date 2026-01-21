// Package gui provides Fyne-based GUI for non-idealities visualization.
package gui

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/demo7-nonidealities/pkg/nonidealities"
)

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

var debug *log.Logger
var logFile *os.File

func init() {
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-nonidealities-demo07.log")

	var err error
	logFile, err = os.Create(logPath)
	if err != nil {
		debug = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
		debug.Printf("Failed to create log file: %v, using stdout", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	debug = log.New(multiWriter, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
	debug.Printf("Logging to: %s", logPath)
}

// NonIdealitiesApp is the main application for the non-idealities demo.
type NonIdealitiesApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Simulators
	irSim    *nonidealities.IRDropSimulator
	sneakSim *nonidealities.SneakPathAnalyzer
	driftSim *nonidealities.DriftSimulator

	// UI Components
	tabs         *container.AppTabs
	irHeatmap    *fyne.Container
	sneakHeatmap *fyne.Container
	driftChart   *fyne.Container
	statsLabel   *widget.Label
	statusLabel  *widget.Label

	// Settings
	arraySize int
}

// NewNonIdealitiesApp creates a new non-idealities visualization app.
func NewNonIdealitiesApp() *NonIdealitiesApp {
	return &NonIdealitiesApp{
		arraySize: 16,
	}
}

// Run starts the application.
func (na *NonIdealitiesApp) Run() {
	na.fyneApp = app.NewWithID("com.fecim.nonidealities-demo")
	na.fyneApp.Settings().SetTheme(&feCIMTheme{})

	na.window = na.fyneApp.NewWindow("FeCIM Demo 7: Non-Idealities Analysis")
	na.window.Resize(fyne.NewSize(1200, 800))

	// Initialize simulators
	na.initSimulators()

	content := na.createMainLayout()
	na.window.SetContent(content)

	// Initial update
	na.updateIRDrop()
	na.updateSneakPaths()
	na.updateDrift()

	debug.Printf("NonIdealitiesApp started")
	na.window.ShowAndRun()
}

func (na *NonIdealitiesApp) initSimulators() {
	// IR Drop simulator
	na.irSim = nonidealities.NewIRDropSimulator(na.arraySize, na.arraySize)
	for i := 0; i < na.arraySize; i++ {
		na.irSim.SetInputVoltage(i, 0.3+0.2*float64(i%5)/4.0)
	}
	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			distFromCenter := float64((i-na.arraySize/2)*(i-na.arraySize/2) + (j-na.arraySize/2)*(j-na.arraySize/2))
			g := 50e-6 + 30e-6*distFromCenter/float64(na.arraySize*na.arraySize/2)
			na.irSim.SetConductance(i, j, g)
		}
	}
	na.irSim.Simulate(100)

	// Sneak path analyzer
	na.sneakSim = nonidealities.NewSneakPathAnalyzer(na.arraySize, na.arraySize)
	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			g := (10 + float64((i*7+j*11)%80)) * 1e-6
			na.sneakSim.SetConductance(i, j, g)
		}
	}
	na.sneakSim.AnalyzeTarget(na.arraySize/2, na.arraySize/2, 0.5)

	// Drift simulator
	na.driftSim = nonidealities.NewDriftSimulator(na.arraySize, na.arraySize, 30)
	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			level := (i*3 + j*5) % 30
			na.driftSim.SetConductanceLevel(i, j, level)
		}
	}
	// Simulate some time
	for step := 0; step < 50; step++ {
		na.driftSim.SimulateTimeStep(200)
		na.driftSim.RecordSnapshot()
	}
}

func (na *NonIdealitiesApp) createMainLayout() fyne.CanvasObject {
	// Header
	header := na.createHeader()

	// Tabs for different analyses
	irTab := container.NewTabItem("IR Drop", na.createIRDropTab())
	sneakTab := container.NewTabItem("Sneak Paths", na.createSneakPathsTab())
	driftTab := container.NewTabItem("Drift", na.createDriftTab())
	compareTab := container.NewTabItem("Technology Comparison", na.createComparisonTab())

	na.tabs = container.NewAppTabs(irTab, sneakTab, driftTab, compareTab)
	na.tabs.SetTabLocation(container.TabLocationTop)

	// Status bar
	na.statusLabel = widget.NewLabel("Ready - Select a tab to view analysis")

	return container.NewBorder(
		header,
		na.statusLabel,
		nil,
		nil,
		na.tabs,
	)
}

func (na *NonIdealitiesApp) createHeader() fyne.CanvasObject {
	title := canvas.NewText("FeCIM Demo 7: Non-Idealities Analysis", color.White)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("IR Drop - Sneak Paths - Conductance Drift")
	subtitle.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
	)
}

func (na *NonIdealitiesApp) createIRDropTab() fyne.CanvasObject {
	// Heatmap
	na.irHeatmap = container.NewWithoutLayout()
	na.irHeatmap.Resize(fyne.NewSize(400, 400))

	heatmapScroll := container.NewScroll(na.irHeatmap)
	heatmapScroll.SetMinSize(fyne.NewSize(350, 350))

	// Stats panel
	irStats := na.irSim.GetStats()
	statsText := fmt.Sprintf(`IR Drop Analysis
================

Max IR Drop: %.3f mV
Avg IR Drop: %.3f mV
Max Output Error: %.2f%%
Avg Output Error: %.2f%%
Worst Cell: (%d, %d)

Severity: %s`,
		irStats.MaxIRDrop*1000,
		irStats.AvgIRDrop*1000,
		irStats.MaxOutputError,
		irStats.AvgOutputError,
		irStats.WorstCellRow, irStats.WorstCellCol,
		na.getIRSeverity(irStats.MaxOutputError))

	statsLabel := widget.NewLabel(statsText)
	statsLabel.Wrapping = fyne.TextWrapWord

	// Mitigation button
	mitigateBtn := widget.NewButton("Apply 2x Wider Lines", func() {
		mitigation := nonidealities.IRDropMitigation{
			UseWidenedLines:   true,
			LineWidthIncrease: 2.0,
		}
		na.irSim.ApplyMitigation(mitigation)
		na.updateIRDrop()
		na.statusLabel.SetText("Mitigation applied: 2x wider metal lines")
	})

	resetBtn := widget.NewButton("Reset", func() {
		na.irSim = nonidealities.NewIRDropSimulator(na.arraySize, na.arraySize)
		for i := 0; i < na.arraySize; i++ {
			na.irSim.SetInputVoltage(i, 0.3+0.2*float64(i%5)/4.0)
		}
		na.irSim.Simulate(100)
		na.updateIRDrop()
		na.statusLabel.SetText("IR drop simulation reset")
	})

	rightPanel := container.NewVBox(
		statsLabel,
		widget.NewSeparator(),
		mitigateBtn,
		resetBtn,
	)

	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("IR Drop Heatmap (darker = higher drop)"),
			nil, nil, nil,
			heatmapScroll,
		),
		rightPanel,
	)
}

func (na *NonIdealitiesApp) createSneakPathsTab() fyne.CanvasObject {
	// Heatmap
	na.sneakHeatmap = container.NewWithoutLayout()
	na.sneakHeatmap.Resize(fyne.NewSize(400, 400))

	heatmapScroll := container.NewScroll(na.sneakHeatmap)
	heatmapScroll.SetMinSize(fyne.NewSize(350, 350))

	// Stats panel
	sneakStats := na.sneakSim.GetStats(0.5)
	statsText := fmt.Sprintf(`Sneak Path Analysis
===================

Target Cell: (%d, %d)
Target Current: %.3f uA
Total Sneak Current: %.3f uA
Sneak Ratio: %.2f%%
Number of Paths: %d
Signal-to-Noise: %.1f dB

Severity: %s`,
		na.sneakSim.TargetRow, na.sneakSim.TargetCol,
		sneakStats.TargetCurrent*1e6,
		sneakStats.TotalSneakCurrent*1e6,
		sneakStats.SneakRatio*100,
		sneakStats.NumSneakPaths,
		sneakStats.SignalToNoiseRatio,
		na.getSneakSeverity(sneakStats.SignalToNoiseRatio))

	statsLabel := widget.NewLabel(statsText)
	statsLabel.Wrapping = fyne.TextWrapWord

	// Mitigation button
	mitigateBtn := widget.NewButton("Apply Selector (1000:1)", func() {
		mitigation := nonidealities.SneakMitigation{
			UseSelector:   true,
			SelectorOnOff: 1000,
		}
		na.sneakSim.AnalyzeWithMitigation(na.arraySize/2, na.arraySize/2, 0.5, mitigation)
		na.updateSneakPaths()
		na.statusLabel.SetText("Mitigation applied: selector device (1000:1 on/off)")
	})

	rightPanel := container.NewVBox(
		statsLabel,
		widget.NewSeparator(),
		mitigateBtn,
	)

	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("Sneak Current Map (X = target cell)"),
			nil, nil, nil,
			heatmapScroll,
		),
		rightPanel,
	)
}

func (na *NonIdealitiesApp) createDriftTab() fyne.CanvasObject {
	// Drift chart
	na.driftChart = container.NewWithoutLayout()
	na.driftChart.Resize(fyne.NewSize(500, 300))

	chartScroll := container.NewScroll(na.driftChart)
	chartScroll.SetMinSize(fyne.NewSize(450, 250))

	// Stats
	driftStats := na.driftSim.GetStats()
	statsText := fmt.Sprintf(`Conductance Drift Analysis
==========================

Elapsed Time: %.1f seconds
Average Drift: %.4f%%
Maximum Drift: %.4f%%
Level Errors: %d (%.4f%%)
10-Year Retention: %.2f%%

FeCIM Drift Coefficient: %.4f
(50x better than RRAM!)`,
		driftStats.ElapsedTime,
		driftStats.AvgDriftPercent,
		driftStats.MaxDriftPercent,
		driftStats.NumLevelErrors,
		driftStats.LevelErrorRate,
		driftStats.RetentionPrediction,
		driftStats.TechnologyComparison.FeFETDrift)

	statsLabel := widget.NewLabel(statsText)
	statsLabel.Wrapping = fyne.TextWrapWord

	// Simulate more button
	simBtn := widget.NewButton("Simulate +1 hour", func() {
		for step := 0; step < 18; step++ { // 18 * 200s = 1 hour
			na.driftSim.SimulateTimeStep(200)
			na.driftSim.RecordSnapshot()
		}
		na.updateDrift()
		na.statusLabel.SetText("Simulated additional 1 hour of drift")
	})

	resetBtn := widget.NewButton("Reset", func() {
		na.driftSim.Reset()
		for step := 0; step < 50; step++ {
			na.driftSim.SimulateTimeStep(200)
			na.driftSim.RecordSnapshot()
		}
		na.updateDrift()
		na.statusLabel.SetText("Drift simulation reset")
	})

	rightPanel := container.NewVBox(
		statsLabel,
		widget.NewSeparator(),
		simBtn,
		resetBtn,
	)

	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("Drift over Time"),
			nil, nil, nil,
			chartScroll,
		),
		rightPanel,
	)
}

func (na *NonIdealitiesApp) createComparisonTab() fyne.CanvasObject {
	// Run comparison
	results := nonidealities.CompareTechnologies(na.arraySize, na.arraySize, 86400) // 1 day

	// Create comparison bars
	bars := container.NewVBox()

	// Find max for scaling
	maxDrift := 0.0
	for _, stats := range results {
		if stats.MaxDriftPercent > maxDrift {
			maxDrift = stats.MaxDriftPercent
		}
	}

	order := []string{"FeCIM (FeFET)", "Flash", "RRAM", "PCM"}
	barColors := []color.Color{
		color.RGBA{0, 220, 150, 255}, // Green for FeCIM
		color.RGBA{255, 200, 0, 255}, // Yellow for Flash
		color.RGBA{255, 150, 0, 255}, // Orange for RRAM
		color.RGBA{255, 80, 80, 255}, // Red for PCM
	}

	for i, name := range order {
		stats, ok := results[name]
		if !ok {
			continue
		}

		// Label
		label := widget.NewLabel(fmt.Sprintf("%-15s Drift: %.3f%%  Retention: %.2f%%",
			name, stats.MaxDriftPercent, stats.RetentionPrediction))

		// Progress bar (scaled)
		bar := widget.NewProgressBar()
		bar.SetValue(stats.MaxDriftPercent / maxDrift)

		// Color indicator
		colorRect := canvas.NewRectangle(barColors[i])
		colorRect.SetMinSize(fyne.NewSize(20, 20))

		row := container.NewHBox(colorRect, label)
		bars.Add(row)
		bars.Add(bar)
		bars.Add(layout.NewSpacer())
	}

	// Summary
	summaryText := `Technology Comparison Summary
=============================

FeCIM (FeFET) achieves:
- 50x lower drift than RRAM
- 100x lower drift than PCM
- 20x lower drift than Flash

This is due to ferroelectric polarization
providing stable, non-volatile states with
minimal drift over time.

FeCIM 10-year retention: >99.9%`

	summaryLabel := widget.NewLabel(summaryText)
	summaryLabel.Wrapping = fyne.TextWrapWord

	return container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("Drift Comparison (after 24 hours)"),
			nil, nil, nil,
			container.NewScroll(bars),
		),
		summaryLabel,
	)
}

func (na *NonIdealitiesApp) updateIRDrop() {
	na.irHeatmap.Objects = nil

	cellSize := float32(20)
	padding := float32(2)

	maxDrop := na.irSim.GetMaxIRDrop()
	if maxDrop == 0 {
		maxDrop = 1
	}

	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			drop := na.irSim.IRDropMap[i][j]
			normalized := drop / maxDrop

			// Color from green (low) to red (high)
			r := uint8(normalized * 255)
			g := uint8((1 - normalized) * 200)
			b := uint8(50)

			rect := canvas.NewRectangle(color.RGBA{r, g, b, 255})
			rect.Resize(fyne.NewSize(cellSize, cellSize))
			rect.Move(fyne.NewPos(float32(j)*(cellSize+padding)+30, float32(i)*(cellSize+padding)+30))
			na.irHeatmap.Add(rect)
		}
	}

	// Add legend
	legend := canvas.NewText("Low IR Drop", color.RGBA{0, 200, 50, 255})
	legend.TextSize = 10
	legend.Move(fyne.NewPos(30, 5))
	na.irHeatmap.Add(legend)

	legend2 := canvas.NewText("High IR Drop", color.RGBA{255, 50, 50, 255})
	legend2.TextSize = 10
	legend2.Move(fyne.NewPos(150, 5))
	na.irHeatmap.Add(legend2)

	na.irHeatmap.Refresh()
}

func (na *NonIdealitiesApp) updateSneakPaths() {
	na.sneakHeatmap.Objects = nil

	cellSize := float32(20)
	padding := float32(2)

	// Find max sneak current
	maxSneak := 0.0
	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			if na.sneakSim.SneakCurrents[i][j] > maxSneak {
				maxSneak = na.sneakSim.SneakCurrents[i][j]
			}
		}
	}
	if maxSneak == 0 {
		maxSneak = 1
	}

	for i := 0; i < na.arraySize; i++ {
		for j := 0; j < na.arraySize; j++ {
			var rectColor color.Color

			if i == na.sneakSim.TargetRow && j == na.sneakSim.TargetCol {
				// Target cell - cyan
				rectColor = colorPrimary
			} else {
				sneak := na.sneakSim.SneakCurrents[i][j]
				normalized := sneak / maxSneak

				// Color from blue (low) to yellow (high)
				r := uint8(normalized * 255)
				g := uint8(normalized * 200)
				b := uint8((1 - normalized) * 150)
				rectColor = color.RGBA{r, g, b, 255}
			}

			rect := canvas.NewRectangle(rectColor)
			rect.Resize(fyne.NewSize(cellSize, cellSize))
			rect.Move(fyne.NewPos(float32(j)*(cellSize+padding)+30, float32(i)*(cellSize+padding)+30))
			na.sneakHeatmap.Add(rect)
		}
	}

	// Mark target
	targetX := float32(na.sneakSim.TargetCol)*(cellSize+padding) + 30 + cellSize/2 - 5
	targetY := float32(na.sneakSim.TargetRow)*(cellSize+padding) + 30 + cellSize/2 - 6
	targetMark := canvas.NewText("X", color.White)
	targetMark.TextSize = 14
	targetMark.TextStyle = fyne.TextStyle{Bold: true}
	targetMark.Move(fyne.NewPos(targetX, targetY))
	na.sneakHeatmap.Add(targetMark)

	na.sneakHeatmap.Refresh()
}

func (na *NonIdealitiesApp) updateDrift() {
	na.driftChart.Objects = nil

	if len(na.driftSim.DriftHistory) == 0 {
		return
	}

	// Find max drift for scaling
	maxDrift := 0.0
	for _, snap := range na.driftSim.DriftHistory {
		if snap.MaxDrift > maxDrift {
			maxDrift = snap.MaxDrift
		}
	}
	if maxDrift == 0 {
		maxDrift = 1e-9
	}

	chartWidth := float32(450)
	chartHeight := float32(200)
	marginLeft := float32(50)

	// Draw axes
	yAxis := canvas.NewLine(color.White)
	yAxis.Position1 = fyne.NewPos(marginLeft, 10)
	yAxis.Position2 = fyne.NewPos(marginLeft, chartHeight)
	na.driftChart.Add(yAxis)

	xAxis := canvas.NewLine(color.White)
	xAxis.Position1 = fyne.NewPos(marginLeft, chartHeight)
	xAxis.Position2 = fyne.NewPos(chartWidth, chartHeight)
	na.driftChart.Add(xAxis)

	// Draw drift line
	barWidth := (chartWidth - marginLeft) / float32(len(na.driftSim.DriftHistory))
	for i, snap := range na.driftSim.DriftHistory {
		// Max drift bar
		height := float32(snap.MaxDrift/maxDrift) * (chartHeight - 20)
		x := marginLeft + float32(i)*barWidth

		maxBar := canvas.NewRectangle(color.RGBA{255, 100, 100, 200})
		maxBar.Resize(fyne.NewSize(barWidth-1, height))
		maxBar.Move(fyne.NewPos(x, chartHeight-height))
		na.driftChart.Add(maxBar)

		// Avg drift bar
		avgHeight := float32(snap.AvgDrift/maxDrift) * (chartHeight - 20)
		avgBar := canvas.NewRectangle(color.RGBA{100, 200, 100, 200})
		avgBar.Resize(fyne.NewSize(barWidth-1, avgHeight))
		avgBar.Move(fyne.NewPos(x, chartHeight-avgHeight))
		na.driftChart.Add(avgBar)
	}

	// Labels
	yLabel := canvas.NewText(fmt.Sprintf("%.1e", maxDrift), color.White)
	yLabel.TextSize = 9
	yLabel.Move(fyne.NewPos(5, 10))
	na.driftChart.Add(yLabel)

	xLabel := canvas.NewText("Time", color.White)
	xLabel.TextSize = 10
	xLabel.Move(fyne.NewPos(chartWidth/2, chartHeight+10))
	na.driftChart.Add(xLabel)

	// Legend
	maxLegend := canvas.NewText("Max Drift", color.RGBA{255, 100, 100, 255})
	maxLegend.TextSize = 10
	maxLegend.Move(fyne.NewPos(chartWidth-100, 10))
	na.driftChart.Add(maxLegend)

	avgLegend := canvas.NewText("Avg Drift", color.RGBA{100, 200, 100, 255})
	avgLegend.TextSize = 10
	avgLegend.Move(fyne.NewPos(chartWidth-100, 25))
	na.driftChart.Add(avgLegend)

	na.driftChart.Refresh()
}

func (na *NonIdealitiesApp) getIRSeverity(maxError float64) string {
	if maxError < 1 {
		return "Excellent (<1% error)"
	} else if maxError < 5 {
		return "Good (<5% error)"
	} else if maxError < 10 {
		return "Acceptable (<10% error)"
	}
	return "Needs Mitigation (>10% error)"
}

func (na *NonIdealitiesApp) getSneakSeverity(snr float64) string {
	if snr > 30 {
		return "Excellent (SNR > 30dB)"
	} else if snr > 20 {
		return "Good (SNR > 20dB)"
	} else if snr > 10 {
		return "Acceptable (SNR > 10dB)"
	}
	return "Needs Selector (SNR < 10dB)"
}
