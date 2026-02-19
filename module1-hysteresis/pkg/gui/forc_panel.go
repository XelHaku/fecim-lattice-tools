package gui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

// FORCDensityMatrix stores FORC density on an Ea/Eb grid.
type FORCDensityMatrix struct {
	EaAxis  []float64   `json:"ea_axis_vm"`
	EbAxis  []float64   `json:"eb_axis_vm"`
	Density [][]float64 `json:"density"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
}

type FORCWorkflowResult struct {
	Sweep       sharedphysics.FORCResult `json:"sweep"`
	Matrix      FORCDensityMatrix        `json:"matrix"`
	Stats       string                   `json:"stats"`
	HeatmapText string                   `json:"heatmap_text"`
}

type guiFORCEverett struct {
	satE float64
	ps   float64
}

func (e guiFORCEverett) Calculate(alpha, beta float64) float64 {
	if alpha <= beta || e.satE <= 0 {
		return 0
	}
	x := (alpha - beta) / (2 * e.satE)
	if x < 0 {
		x = 0
	}
	if x > 1 {
		x = 1
	}
	center := (alpha + beta) / (2 * e.satE)
	weight := math.Exp(-(center * center) / (2 * 0.2 * 0.2))
	return e.ps * x * weight
}

// runFORCWorkflow executes FORC sweep + Preisach density extraction.
// Density is computed in shared/physics as rho = -0.5*∂²M/(∂Ha∂Hb)
// using central finite differences on the reversal-curve grid.
func runFORCWorkflow(materialPs, emax float64, numReversals, resolution int) (FORCWorkflowResult, error) {
	if emax <= 0 {
		return FORCWorkflowResult{}, fmt.Errorf("Emax must be > 0")
	}
	if numReversals < 3 {
		return FORCWorkflowResult{}, fmt.Errorf("numReversals must be >= 3")
	}
	if resolution < 3 {
		return FORCWorkflowResult{}, fmt.Errorf("resolution must be >= 3")
	}
	if materialPs <= 0 {
		materialPs = 0.25
	}

	stack := sharedphysics.NewPreisachStack(emax, guiFORCEverett{satE: emax, ps: materialPs})
	result, err := sharedphysics.RunFORCSweep(stack, emax, numReversals)
	if err != nil {
		return FORCWorkflowResult{}, err
	}

	matrix := buildFORCMatrix(result.ReversalFields_Vm, result.PreisachDensity, resolution)
	stats := summarizeFORCDensity(matrix)
	heatmap := renderHeatmapText(matrix)

	return FORCWorkflowResult{
		Sweep:       result,
		Matrix:      matrix,
		Stats:       stats,
		HeatmapText: heatmap,
	}, nil
}

func buildFORCMatrix(reversal []float64, density [][]float64, resolution int) FORCDensityMatrix {
	rows := len(density)
	if rows == 0 {
		return FORCDensityMatrix{Rows: 0, Cols: 0}
	}
	if resolution < 1 {
		resolution = rows
	}
	cols := resolution
	out := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			src := int(math.Round(float64(j) * float64(len(density[i])-1) / math.Max(float64(cols-1), 1)))
			if src < 0 {
				src = 0
			}
			if src >= len(density[i]) {
				src = len(density[i]) - 1
			}
			out[i][j] = density[i][src]
		}
	}
	eaAxis := make([]float64, cols)
	if len(reversal) == 0 {
		reversal = []float64{0, 1}
	}
	for j := 0; j < cols; j++ {
		src := int(math.Round(float64(j) * float64(len(reversal)-1) / math.Max(float64(cols-1), 1)))
		if src < 0 {
			src = 0
		}
		if src >= len(reversal) {
			src = len(reversal) - 1
		}
		eaAxis[j] = reversal[src]
	}
	ebAxis := append([]float64(nil), reversal...)
	return FORCDensityMatrix{EaAxis: eaAxis, EbAxis: ebAxis, Density: out, Rows: rows, Cols: cols}
}

func summarizeFORCDensity(matrix FORCDensityMatrix) string {
	if matrix.Rows == 0 || matrix.Cols == 0 {
		return "FORC density unavailable"
	}
	minD := math.Inf(1)
	maxD := math.Inf(-1)
	peak := 0.0
	peakI, peakJ := 0, 0
	for i := 0; i < matrix.Rows; i++ {
		for j := 0; j < matrix.Cols; j++ {
			v := matrix.Density[i][j]
			if v < minD {
				minD = v
			}
			if v > maxD {
				maxD = v
			}
			if math.Abs(v) > math.Abs(peak) {
				peak = v
				peakI, peakJ = i, j
			}
		}
	}
	peakEa := matrix.EaAxis[peakJ]
	peakEb := matrix.EbAxis[peakI]
	return fmt.Sprintf("num_curves=%d, peak_density=%.6e at (Ea=%.6e, Eb=%.6e), density_range=[%.6e, %.6e]", matrix.Rows, peak, peakEa, peakEb, minD, maxD)
}

func renderHeatmapText(matrix FORCDensityMatrix) string {
	if matrix.Rows == 0 || matrix.Cols == 0 {
		return "(empty heatmap)"
	}
	var b strings.Builder
	b.WriteString("FORC density grid (rows=Eb, cols=Ea)\n")
	b.WriteString("Eb\\Ea")
	for j := 0; j < matrix.Cols; j++ {
		b.WriteString(fmt.Sprintf(",%.3e", matrix.EaAxis[j]))
	}
	b.WriteString("\n")
	for i := 0; i < matrix.Rows; i++ {
		b.WriteString(fmt.Sprintf("%.3e", matrix.EbAxis[i]))
		for j := 0; j < matrix.Cols; j++ {
			b.WriteString(fmt.Sprintf(",%.3e", matrix.Density[i][j]))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// NewFORCDensityRaster creates a Fyne canvas.Raster rendering the FORC density matrix.
// Color map: 0=dark blue → mid=green → max=red (scientific hot colormap).
func NewFORCDensityRaster(matrix FORCDensityMatrix) *canvas.Raster {
	if matrix.Rows == 0 || matrix.Cols == 0 {
		return canvas.NewRaster(func(w, h int) image.Image {
			return image.NewRGBA(image.Rect(0, 0, w, h))
		})
	}

	// Find max density for normalization
	maxDensity := 0.0
	for i := 0; i < matrix.Rows; i++ {
		for j := 0; j < matrix.Cols; j++ {
			if matrix.Density[i][j] > maxDensity {
				maxDensity = matrix.Density[i][j]
			}
		}
	}
	if maxDensity == 0 {
		maxDensity = 1
	}

	return canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for py := 0; py < h; py++ {
			// Map pixel y to matrix row (flip y: row 0 = bottom)
			row := matrix.Rows - 1 - int(float64(py)*float64(matrix.Rows)/float64(h))
			if row < 0 {
				row = 0
			}
			if row >= matrix.Rows {
				row = matrix.Rows - 1
			}
			for px := 0; px < w; px++ {
				col := int(float64(px) * float64(matrix.Cols) / float64(w))
				if col >= matrix.Cols {
					col = matrix.Cols - 1
				}
				v := matrix.Density[row][col] / maxDensity
				img.Set(px, py, heatColor(v))
			}
		}
		return img
	})
}

// heatColor maps a value in [0,1] to a hot colormap color.
func heatColor(v float64) color.RGBA {
	if v <= 0 {
		return color.RGBA{0, 0, 80, 255} // dark blue
	}
	if v >= 1 {
		return color.RGBA{255, 255, 0, 255} // yellow
	}
	// 4-segment linear hot colormap: blue→cyan→green→yellow→red
	switch {
	case v < 0.25:
		t := v / 0.25
		return color.RGBA{0, uint8(t * 255), 255, 255}
	case v < 0.5:
		t := (v - 0.25) / 0.25
		return color.RGBA{0, 255, uint8((1 - t) * 255), 255}
	case v < 0.75:
		t := (v - 0.5) / 0.25
		return color.RGBA{uint8(t * 255), 255, 0, 255}
	default:
		t := (v - 0.75) / 0.25
		return color.RGBA{255, uint8((1 - t) * 255), 0, 255}
	}
}

func ExportFORCSweepCSV(result sharedphysics.FORCResult, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"reversal_field", "E", "P"}); err != nil {
		return err
	}
	for _, curve := range result.Curves {
		for i := range curve.AppliedField_Vm {
			if i >= len(curve.Polarization_Cm2) {
				continue
			}
			row := []string{
				fmt.Sprintf("%.9e", curve.ReversalField_Vm),
				fmt.Sprintf("%.9e", curve.AppliedField_Vm[i]),
				fmt.Sprintf("%.9e", curve.Polarization_Cm2[i]),
			}
			if err := w.Write(row); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func ExportFORCMatrixCSV(density FORCDensityMatrix, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"Ea", "Eb", "density"}); err != nil {
		return err
	}
	for i := 0; i < density.Rows; i++ {
		for j := 0; j < density.Cols; j++ {
			if err := w.Write([]string{
				fmt.Sprintf("%.9e", density.EaAxis[j]),
				fmt.Sprintf("%.9e", density.EbAxis[i]),
				fmt.Sprintf("%.9e", density.Density[i][j]),
			}); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func gitCommitShort() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func ExportFORCMetadata(material, engine string, params map[string]any, path string) error {
	payload := map[string]any{
		"timestamp":  time.Now().Format(time.RFC3339),
		"git_commit": gitCommitShort(),
		"material":   material,
		"engine":     engine,
		"params":     params,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func (a *App) createFORCPanel() fyne.CanvasObject {
	emaxEntry := widget.NewEntry()
	emaxEntry.SetText("1.0")
	reversalEntry := widget.NewEntry()
	reversalEntry.SetText("21")
	resolutionEntry := widget.NewEntry()
	resolutionEntry.SetText("21")

	resultLabel := widget.NewLabel("Ready")
	resultLabel.Wrapping = fyne.TextWrapOff
	resultScroll := container.NewHScroll(resultLabel)
	resultScroll.SetMinSize(fyne.NewSize(0, 200))

	// Density raster display (updated after each FORC run)
	rasterPlaceholder := NewFORCDensityRaster(FORCDensityMatrix{})
	rasterPlaceholder.SetMinSize(fyne.NewSize(300, 300))
	rasterContainer := container.NewStack(rasterPlaceholder)

	var last FORCWorkflowResult
	runFORC := func() {
		a.mu.RLock()
		ps := 0.25
		if a.material != nil {
			ps = a.material.Ps
		}
		a.mu.RUnlock()

		emax, eErr := parseFloatOrDefault(emaxEntry.Text, 1.0)
		nrev, nErr := parseIntOrDefault(reversalEntry.Text, 21)
		res, rErr := parseIntOrDefault(resolutionEntry.Text, 21)
		if eErr != nil || nErr != nil || rErr != nil {
			fyne.Do(func() { resultLabel.SetText("invalid parameters") })
			return
		}
		out, err := runFORCWorkflow(ps, emax, nrev, res)
		if err != nil {
			fyne.Do(func() { resultLabel.SetText(fmt.Sprintf("FORC error: %v", err)) })
			return
		}
		last = out
		fyne.Do(func() {
			resultLabel.SetText(out.Stats + "\n\n" + out.HeatmapText)
			newRaster := NewFORCDensityRaster(out.Matrix)
			newRaster.SetMinSize(fyne.NewSize(300, 300))
			rasterContainer.Objects = []fyne.CanvasObject{newRaster}
			rasterContainer.Refresh()
		})
	}

	exportSweepBtn := widget.NewButton("Sweep CSV", func() {
		if len(last.Sweep.Curves) == 0 {
			runFORC()
		}
		dir := "exports"
		_ = os.MkdirAll(dir, 0755)
		name := filepath.Join(dir, fmt.Sprintf("forc-sweep-%s.csv", time.Now().Format("2006-01-02T15-04-05")))
		if err := ExportFORCSweepCSV(last.Sweep, name); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}
		a.setStatus("Exported " + name)
	})

	exportMatrixBtn := widget.NewButton("Matrix CSV", func() {
		if len(last.Sweep.Curves) == 0 {
			runFORC()
		}
		dir := "exports"
		_ = os.MkdirAll(dir, 0755)
		ts := time.Now().Format("2006-01-02T15-04-05")
		csvPath := filepath.Join(dir, fmt.Sprintf("forc-matrix-%s.csv", ts))
		jsonPath := filepath.Join(dir, fmt.Sprintf("forc-matrix-%s.json", ts))
		if err := ExportFORCMatrixCSV(last.Matrix, csvPath); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}
		b, _ := json.MarshalIndent(last.Matrix, "", "  ")
		if err := os.WriteFile(jsonPath, b, 0644); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}
		a.setStatus("Exported " + csvPath + " and JSON")
	})

	exportMetaBtn := widget.NewButton("Meta JSON", func() {
		a.mu.RLock()
		mat := "unknown"
		engine := a.physicsEngine.String()
		if a.material != nil {
			mat = a.material.Name
		}
		a.mu.RUnlock()
		params := map[string]any{
			"Emax":         emaxEntry.Text,
			"numReversals": reversalEntry.Text,
			"resolution":   resolutionEntry.Text,
		}
		dir := "exports"
		_ = os.MkdirAll(dir, 0755)
		name := filepath.Join(dir, fmt.Sprintf("forc-metadata-%s.json", time.Now().Format("2006-01-02T15-04-05")))
		if err := ExportFORCMetadata(mat, engine, params, name); err != nil {
			dialog.ShowError(err, a.mainWindow)
			return
		}
		a.setStatus("Exported " + name)
	})

	runBtn := widget.NewButton("Run FORC", runFORC)

	controls := container.NewGridWithColumns(2,
		widget.NewLabel("Emax (V/m):"), emaxEntry,
		widget.NewLabel("Reversals:"), reversalEntry,
		widget.NewLabel("Resolution:"), resolutionEntry,
	)

	return widget.NewCard("FORC Workflow", "Mixed derivative ∂²M/∂Ha∂Hb via central finite differences", container.NewVBox(
		controls,
		runBtn,
		container.NewGridWithColumns(3, exportSweepBtn, exportMatrixBtn, exportMetaBtn),
		widget.NewLabel("Density map (run FORC to populate):"),
		rasterContainer,
		resultScroll,
	))
}

func parseFloatOrDefault(s string, d float64) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return d, err
	}
	return v, nil
}

func parseIntOrDefault(s string, d int) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return d, err
	}
	return v, nil
}
