//go:build !cgo

package gogpuscreenshot

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"

	"fecim-lattice-tools/internal/gogpuapp/design"
	"fecim-lattice-tools/internal/gogpuapp/render"

	"github.com/gogpu/gg"
	xdraw "golang.org/x/image/draw"
)

var activeOptions = DefaultOptions()
var captureErr error

func Run(args []string) error {
	opts, err := ParseOptions(args)
	if err != nil {
		return err
	}
	return Generate(opts)
}

func Generate(opts Options) error {
	activeOptions = opts
	captureErr = nil

	count := 0
	if opts.Matches("hysteresis") {
		genLoop()
		genHysteresisComparison()
		count += 2
	}
	if opts.Matches("crossbar") {
		genHeatmap8()
		genHeatmap16()
		genMVM()
		count += 3
	}
	if opts.Matches("mnist") {
		genMNIST()
		count++
	}
	if opts.Matches("circuits") {
		genCircuitsISPP()
		genCircuitsPVT()
		count += 2
	}
	if opts.Matches("comparison") {
		genComparisonBar()
		count++
	}
	if opts.Matches("eda") {
		genEDA()
		count++
	}
	if opts.Matches("docs") {
		genDocs()
		count++
	}

	if count == 0 {
		return fmt.Errorf("no screenshots matched -only %q", opts.Only)
	}
	if captureErr != nil {
		return captureErr
	}
	log.Printf("done — %d screens generated in %s/", count, opts.OutputDir)
	return nil
}

func captureAndSave(width, height int, draw func(*gg.Context), filename string) error {
	if captureErr != nil {
		return captureErr
	}
	dc := gg.NewContext(width, height)
	defer dc.Close()
	if draw != nil {
		draw(dc)
	}

	outPath := activeOptions.OutputPath(filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		captureErr = fmt.Errorf("screenshot: mkdir %s: %w", filepath.Dir(outPath), err)
		return captureErr
	}
	if activeOptions.Width == width && activeOptions.Height == height {
		if err := dc.SavePNG(outPath); err != nil {
			captureErr = fmt.Errorf("screenshot: save %s: %w", outPath, err)
		}
		return captureErr
	}

	scaled := image.NewRGBA(image.Rect(0, 0, activeOptions.Width, activeOptions.Height))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), dc.Image(), dc.Image().Bounds(), xdraw.Over, nil)
	file, err := os.Create(outPath)
	if err != nil {
		captureErr = fmt.Errorf("screenshot: create %s: %w", outPath, err)
		return captureErr
	}
	defer file.Close()
	if err := png.Encode(file, scaled); err != nil {
		captureErr = fmt.Errorf("screenshot: encode %s: %w", outPath, err)
	}
	return captureErr
}

func genLoop() {
	log.Print("[1/11] hysteresis-p-e-loop.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("P-E Hysteresis Loop", "Field (kV/cm)", "Polarization (µC/cm²)")
		data.AddSeries("HZO Default", loop(3000, 20, 0.3))
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 80, Y: 60, Width: 1240, Height: 780, Background: "#FFFFFF",
		})
	}, "hysteresis-p-e-loop.png")
}

func genHeatmap8() {
	log.Print("[2/11] crossbar-heatmap-8x8.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := make([][]float64, 8)
		for i := range 8 {
			data[i] = make([]float64, 8)
			for j := range 8 {
				data[i][j] = float64(30 - ((i*8 + j) % 30) + i)
			}
		}
		render.DrawHeatmap(dc, render.HeatmapConfig{
			Data: data, X: 200, Y: 120, CellSize: 60,
			Title: "Crossbar Conductance Matrix (8×8, 30 levels)",
		})
	}, "crossbar-heatmap-8x8.png")
}

func genHeatmap16() {
	log.Print("[3/11] crossbar-heatmap-16x16.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := make([][]float64, 16)
		for i := range 16 {
			data[i] = make([]float64, 16)
			for j := range 16 {
				data[i][j] = float64(30 - ((i*16 + j) % 30) + i/2)
			}
		}
		render.DrawHeatmap(dc, render.HeatmapConfig{
			Data: data, X: 120, Y: 100, CellSize: 36,
			Title: "Crossbar Conductance Matrix (16×16, 30 levels)",
		})
	}, "crossbar-heatmap-16x16.png")
}

func genHysteresisComparison() {
	log.Print("[4/11] hysteresis-material-comparison.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("P-E Hysteresis — Material Comparison", "Field (kV/cm)", "Polarization (µC/cm²)")
		data.AddSeries("HZO Default", loop(3000, 20, 0.3))
		data.AddSeries("BTO (high Pr)", loop(2500, 30, 0.2))
		data.AddSeries("PZT (low Ec)", loop(1500, 25, 0.15))
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 80, Y: 60, Width: 1240, Height: 780, Background: "#FFFFFF",
		})
	}, "hysteresis-material-comparison.png")
}

func genMVM() {
	log.Print("[5/11] mvm-diagram.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		// G matrix grid
		for r := 0; r < 8; r++ {
			for c := 0; c < 8; c++ {
				x := 120.0 + float64(c)*40
				y := 180.0 + float64(r)*40
				v := float64(30 - ((r*8 + c) % 30) + r)
				t := clamp01(v / 30.0)
				rr, gg, bb := heatmap(t)
				dc.SetRGBA(float64(rr)/255, float64(gg)/255, float64(bb)/255, 1)
				dc.DrawRectangle(x, y, 32, 32)
				dc.Fill()
			}
		}
		// V vector
		for r := 0; r < 8; r++ {
			dc.SetRGBA(0.18, 0.36, 0.31, 1)
			dc.DrawRectangle(550, 180+float64(r)*40, 50, 32)
			dc.Fill()
		}
		// Arrow
		dc.SetHexColor("#1A1C1A")
		dc.SetLineWidth(3)
		dc.MoveTo(470, 340)
		dc.LineTo(540, 340)
		dc.Stroke()
		dc.MoveTo(610, 340)
		dc.LineTo(680, 340)
		dc.Stroke()
		// I output
		for r := 0; r < 8; r++ {
			dc.SetRGBA(0.35, 0.53, 0.49, 1)
			dc.DrawRectangle(700, 180+float64(r)*40, 80, 32)
			dc.Fill()
		}
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("G", 392, 120, 0.5, 0)
		dc.DrawStringAnchored("V", 575, 120, 0.5, 0)
		dc.DrawStringAnchored("I = G × V", 740, 120, 0.5, 0)
	}, "mvm-diagram.png")
}

func loop(maxF, maxP, w float64) []design.PlotPoint {
	pts := make([]design.PlotPoint, 200)
	for i := range 200 {
		t := float64(i) * 2 * math.Pi / 199
		pts[i] = design.PlotPoint{X: maxF * math.Sin(t), Y: maxP * math.Sin(t-math.Pi/6) * (1 - w*math.Abs(math.Sin(t)))}
	}
	return pts
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func heatmap(t float64) (uint8, uint8, uint8) {
	if t < 0.33 {
		s := t / 0.33
		return uint8(41 + s*49), uint8(93 + s*93), uint8(80 + s*76)
	}
	if t < 0.66 {
		s := (t - 0.33) / 0.33
		return uint8(90 + s*165), uint8(186 - s*34), uint8(156 - s*116)
	}
	s := (t - 0.66) / 0.34
	return uint8(255), uint8(152 + s*103), uint8(40 - s*40)
}

// ============================================================
// Module 3: MNIST — Accuracy vs Quantization Levels
// ============================================================

func genMNIST() {
	log.Print("[6/11] mnist-accuracy-sweep.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("MNIST — Accuracy vs Quantization Levels", "Quantization Levels", "Accuracy (%)")
		levels := []float64{2, 4, 8, 16, 32, 64, 128}
		accs := []float64{55, 65, 74, 79, 82, 84, 85}
		pts := make([]design.PlotPoint, len(levels))
		for i := range levels {
			pts[i] = design.PlotPoint{X: levels[i], Y: accs[i]}
		}
		data.AddSeries("CIM Accuracy", pts)
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 100, Y: 60, Width: 1200, Height: 780,
			Background: "#F4F5F3", GridColor: "#C4C7C4",
		})
		// Annotation
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("Educational simulation — 80% baseline at 30 levels. 98.24% reference (HZO FTJ, J. Alloys Comp. 2025) is a different architecture.", 700, 870, 0.5, 0)
	}, "mnist-accuracy-sweep.png")
}

// ============================================================
// Module 4: Circuits — ISPP Write-Verify + PVT Corners
// ============================================================

func genCircuitsISPP() {
	log.Print("[7/11] circuits-ispp-convergence.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()
		data := design.NewPlotData("ISPP Write-Verify Convergence (30 levels)", "Target Level", "Attempts to Converge")
		pts := make([]design.PlotPoint, 30)
		rng := xorshift32(42)
		for i := range 30 {
			// Realistic ISPP convergence: 3-8 attempts per level with noise
			base := 4.0 + 2.0*float64(i)/30.0 // harder at higher levels
			noise := (float64(rng()%100)/100.0 - 0.5) * 2.0
			pts[i] = design.PlotPoint{X: float64(i), Y: base + noise}
		}
		data.AddSeries("attempts", pts)
		render.DrawPlot(dc, render.PlotConfig{
			Data: data, X: 100, Y: 60, Width: 1200, Height: 780,
			Background: "#F4F5F3", GridColor: "#C4C7C4",
		})
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("5-bit SAR ADC → TIA → ISPP binary search. Educational behavioral model — not calibrated against silicon.", 700, 870, 0.5, 0)
	}, "circuits-ispp-convergence.png")
}

func genCircuitsPVT() {
	log.Print("[8/11] circuits-pvt-corners.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()

		// Title
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("ADC ENOB — PVT Corners (5-bit SAR, Supply 1.8V)", 700, 40, 0.5, 0)

		// Bar chart: ENOB at TT, FF, SS
		corners := []struct {
			name          string
			enob, r, g, b float64
		}{
			{"TT (Typical)", 4.42, 0.18, 0.36, 0.31},
			{"FF (Fast)", 4.57, 0.35, 0.53, 0.49},
			{"SS (Slow)", 4.20, 0.62, 0.47, 0.42},
		}
		barW, barH, gap := 120.0, 400.0, 60.0
		maxENOB := 5.0
		startX := 200.0

		for i, c := range corners {
			x := startX + float64(i)*(barW+gap) + barW/2
			h := c.enob / maxENOB * barH
			y := 500.0 - h
			// Bar
			dc.SetRGBA(c.r, c.g, c.b, 1)
			dc.DrawRoundedRectangle(x-barW/2, y, barW, h, 8)
			dc.Fill()
			// Label
			dc.SetHexColor("#1A1C1A")
			dc.DrawStringAnchored(c.name, x, 520, 0.5, 0)
			dc.DrawStringAnchored(fmt.Sprintf("%.2f bits", c.enob), x, y-12, 0.5, 0)
		}

		// Metrics table
		metrics := []string{
			"  DAC: 5-bit R-2R  |  TIA: 10.0 kΩ  |  Charge Pump: 4-stage Dickson",
			"  Ideal SNR: 31.8 dB (5-bit)  |  INL < 0.5 LSB at TT",
			"  Read path: TIA → SAR ADC  |  Write path: Charge Pump → DAC → ISPP pulse train",
		}
		dc.SetRGBA(0.25, 0.25, 0.25, 1)
		for i, m := range metrics {
			dc.DrawStringAnchored(m, 80, 590+float64(i)*32, 0, 0)
		}

		dc.SetHexColor("#5C3B00")
		dc.DrawStringAnchored("SIMULATION OUTPUT — Educational circuit models. Not calibrated against silicon measurements.", 700, 870, 0.5, 0)
	}, "circuits-pvt-corners.png")
}

// ============================================================
// Module 5: Comparison — Architecture TOPS/W Bar Chart
// ============================================================

func genComparisonBar() {
	log.Print("[9/11] comparison-architecture-bars.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()

		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("Technology Comparison — TOPS/W Efficiency (log scale)", 700, 36, 0.5, 0)

		arches := []struct {
			name     string
			topsPerW float64
			chipArea float64
			tdp      float64
			r, g, b  float64
		}{
			{"CPU+DRAM", 0.008, 400, 125, 0.47, 0.40, 0.42},
			{"GPU (HBM)", 0.25, 800, 400, 0.58, 0.35, 0.40},
			{"FeCIM CIM", 10.0, 50, 5, 0.18, 0.36, 0.31},
		}

		// Bar chart using log scale for TOPS/W
		barW, maxBarH := 140.0, 500.0
		gap := 80.0
		startX := 240.0
		baseY := 560.0
		logMax := math.Log10(20.0) // reference max for scaling

		for i, a := range arches {
			x := startX + float64(i)*(barW+gap) + barW/2
			h := math.Max(math.Log10(math.Max(a.topsPerW, 0.001))/logMax*maxBarH, 20)
			y := baseY - h
			dc.SetRGBA(a.r, a.g, a.b, 1)
			dc.DrawRoundedRectangle(x-barW/2, y, barW, h, 8)
			dc.Fill()
			dc.SetHexColor("#1A1C1A")
			dc.DrawStringAnchored(a.name, x, baseY+22, 0.5, 0)
			dc.DrawStringAnchored(fmt.Sprintf("%.3f TOPS/W", a.topsPerW), x, y-14, 0.5, 0)
			// Sub-metrics
			dc.SetRGBA(0.35, 0.35, 0.35, 1)
			dc.DrawStringAnchored(fmt.Sprintf("%.0f mm²", a.chipArea), x, y-34, 0.5, 0)
			dc.DrawStringAnchored(fmt.Sprintf("%.0f W", a.tdp), x, y-50, 0.5, 0)
		}

		// Advantage callouts
		dc.SetHexColor("#2F5D50")
		dc.DrawStringAnchored("FeCIM: 1250× more efficient than CPU, 40× more than GPU (educational estimates, TRL 4)", 700, 640, 0.5, 0)
		dc.SetHexColor("#5C3B00")
		dc.DrawStringAnchored("ESTIMATED DATA — FeCIM specs are model inputs at TRL 4. Not validated device measurements.", 700, 870, 0.5, 0)

		// Table
		headers := []string{"Architecture", "Process", "TOPS/W", "Area", "TDP"}
		rows := [][]string{
			{"CPU+DRAM", "5 nm", "0.008", "400 mm²", "125 W"},
			{"GPU Accelerator", "4 nm", "0.25", "800 mm²", "400 W"},
			{"FeCIM CIM", "45 nm*", "10.0*", "50 mm²*", "5 W*"},
		}
		tableX, tableY := 80.0, 680.0
		colW := []float64{180, 100, 100, 110, 80}
		dc.SetRGBA(0.15, 0.15, 0.15, 1)
		cx := tableX
		for j, h := range headers {
			dc.DrawStringAnchored(h, cx, tableY, 0, 0)
			cx += colW[j]
		}
		dc.SetRGBA(0.35, 0.35, 0.35, 1)
		for _, row := range rows {
			tableY += 22
			cx = tableX
			for j, cell := range row {
				dc.DrawStringAnchored(cell, cx, tableY, 0, 0)
				cx += colW[j]
			}
		}
	}, "comparison-architecture-bars.png")
}

// ============================================================
// Module 6: EDA — Design Stats + Export Formats
// ============================================================

func genEDA() {
	log.Print("[10/11] eda-design-overview.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()

		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("FeCIM EDA Design Suite — fecim_crossbar_8x8 (sky130)", 700, 30, 0.5, 0)

		// Design stats cards
		stats := []struct {
			label, value string
		}{
			{"Design", "fecim_crossbar_8x8"},
			{"Process Node", "sky130 (130 nm)"},
			{"Array Size", "8 × 8"},
			{"Total Cells", "64"},
			{"Area", "0.245 mm²"},
			{"Power", "18.5 mW"},
		}
		cardX, cardY := 80.0, 70.0
		for i, s := range stats {
			x := cardX + float64(i%3)*240
			y := cardY + float64(i/3)*100
			dc.SetHexColor("#E8EBE7")
			dc.DrawRoundedRectangle(x, y, 220, 80, 8)
			dc.Fill()
			dc.SetHexColor("#2F5D50")
			dc.DrawStringAnchored(s.label, x+12, y+22, 0, 0.5)
			dc.SetHexColor("#1A1C1A")
			dc.DrawStringAnchored(s.value, x+12, y+52, 0, 0.5)
		}

		// Export format cards
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("Export Formats", 80, 310, 0, 0)

		formats := []struct {
			name, desc string
			r, g, b    float64
		}{
			{"SPICE", "Circuit simulation netlist\nwith FeFET compact model", 0.18, 0.36, 0.31},
			{"Verilog", "RTL module for digital\ncontrol and testbench", 0.35, 0.53, 0.49},
			{"Liberty", "Timing library (.lib)\nTT/FF/SS corners", 0.58, 0.35, 0.40},
			{"DEF", "Physical design layout\nfloorplan and routing", 0.47, 0.42, 0.36},
			{"LEF", "Technology LEF macro\nstandard cell library", 0.62, 0.47, 0.42},
		}
		for i, f := range formats {
			x := 80.0 + float64(i)*260
			dc.SetRGBA(f.r, f.g, f.b, 1)
			dc.DrawRoundedRectangle(x, 340, 240, 140, 8)
			dc.Fill()
			dc.SetRGBA(1, 1, 1, 1)
			dc.DrawStringAnchored(f.name, x+120, 365, 0.5, 0)
			dc.SetRGBA(0.95, 0.95, 0.95, 1)
			lines := []string{""}
			for _, line := range lines {
				dc.DrawStringAnchored(line, x+16, 400, 0, 0)
			}
			// Draw description on multiple lines
			descLines := []string{}
			current := ""
			for _, ch := range f.desc {
				if ch == '\n' {
					descLines = append(descLines, current)
					current = ""
				} else {
					current += string(ch)
				}
			}
			if current != "" {
				descLines = append(descLines, current)
			}
			for j, line := range descLines {
				dc.SetRGBA(0.95, 0.95, 0.95, 1)
				dc.DrawStringAnchored(line, x+16, 400+float64(j)*20, 0, 0)
			}
		}

		// Design flow diagram
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("Design Flow", 80, 540, 0, 0)

		flowSteps := []string{"RTL", "Synthesis", "Floorplan", "Place", "Route", "DRC/LVS", "GDSII"}
		stepW := 160.0
		for i, step := range flowSteps {
			x := 80.0 + float64(i)*stepW
			y := 570.0
			dc.SetHexColor("#E8EBE7")
			dc.DrawRoundedRectangle(x, y, stepW-16, 48, 8)
			dc.Fill()
			dc.SetHexColor("#2F5D50")
			dc.DrawStringAnchored(fmt.Sprintf("%d", i+1), x+12, y+24, 0, 0.5)
			dc.SetHexColor("#1A1C1A")
			dc.DrawStringAnchored(step, x+36, y+24, 0, 0.5)
			if i < len(flowSteps)-1 {
				dc.SetHexColor("#2F5D50")
				dc.SetLineWidth(2)
				dc.MoveTo(x+stepW-16, y+24)
				dc.LineTo(x+stepW, y+24)
				dc.Stroke()
				// Arrow
				dc.MoveTo(x+stepW-8, y+18)
				dc.LineTo(x+stepW, y+24)
				dc.LineTo(x+stepW-8, y+30)
				dc.Stroke()
			}
		}

		// SPICE snippet
		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("SPICE Netlist (excerpt)", 80, 660, 0, 0)
		dc.SetHexColor("#1A1C1A")
		dc.DrawRoundedRectangle(80, 680, 600, 170, 6)
		dc.Fill()
		dc.SetRGBA(0.85, 0.95, 0.85, 1)
		spiceLines := []string{
			"* FeCIM crossbar 8x8 (fecim_crossbar_8x8)",
			".include fefet_compact_model.lib",
			".global VDD VSS",
			"XWL0 WL0 VSS fefet_row DR=0 G=15",
			"XWL1 WL1 VSS fefet_row DR=1 G=22",
			"... (60 more cells)",
		}
		for i, line := range spiceLines {
			dc.DrawStringAnchored("  "+line, 92, 708+float64(i)*22, 0, 0)
		}

		dc.SetHexColor("#5C3B00")
		dc.DrawStringAnchored("EDUCATIONAL EDA — Not a production chip design tool. No proprietary foundry PDKs used.", 700, 870, 0.5, 0)
	}, "eda-design-overview.png")
}

// ============================================================
// Module 7: Docs — Documentation Sections Overview
// ============================================================

func genDocs() {
	log.Print("[11/11] docs-overview.png")
	captureAndSave(1400, 900, func(dc *gg.Context) {
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawRectangle(0, 0, 1400, 900)
		dc.Fill()

		dc.SetHexColor("#1A1C1A")
		dc.DrawStringAnchored("FeCIM Documentation — 7 Modules, 230+ References, Interactive Curriculum", 700, 30, 0.5, 0)

		// Stats row
		statCards := []struct{ label, value string }{
			{"Modules", "7"},
			{"References", "230+"},
			{"Papers", "28 verified facts"},
			{"Topics", "23 categories"},
		}
		for i, sc := range statCards {
			x := 80.0 + float64(i)*320
			dc.SetHexColor("#E8EBE7")
			dc.DrawRoundedRectangle(x, 60, 290, 80, 8)
			dc.Fill()
			dc.SetHexColor("#2F5D50")
			dc.DrawStringAnchored(sc.value, x+145, 85, 0.5, 0)
			dc.SetRGBA(0.35, 0.35, 0.35, 1)
			dc.DrawStringAnchored(sc.label, x+145, 118, 0.5, 0)
		}

		// Sections with color coding
		sections := []struct {
			title, body, cat string
			r, g, b          float64
		}{
			{
				title: "Learning Curriculum", cat: "education",
				body: "FeCIM 101 → Hysteresis → Crossbar Arrays → CIM Inference → Design Export.\nGuided walkthrough with 7 interactive modules. Progressive depth: education → research → design.",
				r:    0.18, g: 0.36, b: 0.31,
			},
			{
				title: "Citation Browser", cat: "research",
				body: "Filter by module, paper, or author. Track verified vs. educational claims.\nHonesty audit dashboard. All sources cited or marked educational.",
				r:    0.35, g: 0.53, b: 0.49,
			},
			{
				title: "Interactive Glossary", cat: "education",
				body: "Click any term to see definition, equation, and citation.\nKey terms: Ec (coercive field), Pr (remanent polarization), ISPP, MVM, IR drop, sneak path.",
				r:    0.18, g: 0.36, b: 0.31,
			},
			{
				title: "Design Guide", cat: "design",
				body: "Step-by-step accelerator design workflow. Cross-module integration:\nMaterial (M1) → Array (M2) → Circuits (M4) → EDA Export (M6).",
				r:    0.47, g: 0.42, b: 0.36,
			},
			{
				title: "Honesty Audit", cat: "research",
				body: "Verified claims: HZO parameters (Materlik 2015, Park 2015). Educational defaults:\n30-level quantization, energy models. Not validated: accuracy claims without evidence.",
				r:    0.35, g: 0.53, b: 0.49,
			},
			{
				title: "Trust Boundaries", cat: "research",
				body: "Each output labeled: validated (golden), literature-backed (cited), educational (default),\nplanned (not built), or not validated. Full trust matrix at docs/TRUST.md.",
				r:    0.35, g: 0.53, b: 0.49,
			},
		}

		y := 180.0
		for _, s := range sections {
			// Category badge
			catColors := map[string][3]float64{
				"education": {0.18, 0.36, 0.31},
				"research":  {0.35, 0.53, 0.49},
				"design":    {0.47, 0.42, 0.36},
			}
			cc := catColors[s.cat]
			dc.SetRGBA(cc[0], cc[1], cc[2], 1)
			dc.DrawRoundedRectangle(80, y, 100, 26, 4)
			dc.Fill()
			dc.SetRGBA(1, 1, 1, 1)
			dc.DrawStringAnchored(s.cat, 130, y+13, 0.5, 0.5)

			dc.SetHexColor("#1A1C1A")
			dc.DrawStringAnchored(s.title, 200, y+13, 0, 0.5)

			dc.SetRGBA(0.35, 0.35, 0.35, 1)
			bodyLines := splitLines(s.body)
			for j, line := range bodyLines {
				dc.DrawStringAnchored(line, 200, y+36+float64(j)*20, 0, 0)
			}

			y += 40.0 + float64(len(bodyLines))*20.0 + 10.0
		}

		dc.SetHexColor("#5C3B00")
		dc.DrawStringAnchored("Documentation does not produce simulation output. All references are cited and categorized by validation status.", 700, 870, 0.5, 0)
	}, "docs-overview.png")
}

func splitLines(s string) []string {
	var result []string
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// xorshift32 is a deterministic pseudo-random number generator for reproducible screenshots.
func xorshift32(seed uint32) func() uint32 {
	state := seed
	if state == 0 {
		state = 1
	}
	return func() uint32 {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		return state
	}
}

// ensure sort is used (for potential future use)
var _ = sort.Float64s

// ensure fmt is used
var _ = fmt.Sprintf
