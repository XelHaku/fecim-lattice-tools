//go:build legacy_fyne

package gui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	mvCmToVm    = 1e8
	uCCm2ToCm2  = 1e-2
	cm2ToUCCm2  = 100.0
	defaultRows = 12
)

type DataUnits struct {
	EFieldUnit       string `json:"e_field_unit"`
	PolarizationUnit string `json:"polarization_unit"`
}

type LiteratureDataset struct {
	Source       string    `json:"source"`
	Material     string    `json:"material"`
	EField       []float64 `json:"e_field_v_m"`
	Polarization []float64 `json:"polarization_c_m2"`
	Units        DataUnits `json:"units"`
}

type FitMetrics struct {
	RMSE     float64
	MAE      float64
	MaxErr   float64
	NSamples int
}

type overlayJSON struct {
	Source   string    `json:"source"`
	Material string    `json:"material"`
	EVm      []float64 `json:"E_V_m"`
	PCm2     []float64 `json:"P_C_m2"`
}

func LoadLiteratureCSV(path string) (*LiteratureDataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open CSV: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	recs, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	if len(recs) < 2 {
		return nil, fmt.Errorf("malformed CSV: expected header + at least 1 data row")
	}
	h := recs[0]
	if len(h) != 2 {
		return nil, fmt.Errorf("malformed CSV header: expected 2 columns (E_MV_cm,P_uC_cm2), got %d", len(h))
	}
	if strings.TrimSpace(h[0]) != "E_MV_cm" || strings.TrimSpace(h[1]) != "P_uC_cm2" {
		return nil, fmt.Errorf("malformed CSV header: expected E_MV_cm,P_uC_cm2")
	}

	d := &LiteratureDataset{Units: DataUnits{EFieldUnit: "MV/cm", PolarizationUnit: "uC/cm2"}}
	for i := 1; i < len(recs); i++ {
		row := recs[i]
		if len(row) != 2 {
			return nil, fmt.Errorf("malformed CSV row %d: expected 2 columns, got %d", i+1, len(row))
		}
		eMV, err := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("malformed CSV row %d: invalid E_MV_cm value %q", i+1, row[0])
		}
		pUC, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("malformed CSV row %d: invalid P_uC_cm2 value %q", i+1, row[1])
		}
		d.EField = append(d.EField, eMV*mvCmToVm)
		d.Polarization = append(d.Polarization, pUC*uCCm2ToCm2)
	}
	if err := validateDataset(d); err != nil {
		return nil, err
	}
	return d, nil
}

func LoadLiteratureJSON(path string) (*LiteratureDataset, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JSON: %w", err)
	}
	var raw overlayJSON
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("malformed JSON: %w", err)
	}
	d := &LiteratureDataset{
		Source:       raw.Source,
		Material:     raw.Material,
		EField:       append([]float64(nil), raw.EVm...),
		Polarization: append([]float64(nil), raw.PCm2...),
		Units:        DataUnits{EFieldUnit: "V/m", PolarizationUnit: "C/m2"},
	}
	if err := validateDataset(d); err != nil {
		return nil, err
	}
	return d, nil
}

func validateDataset(d *LiteratureDataset) error {
	if d == nil {
		return fmt.Errorf("dataset is nil")
	}
	if len(d.EField) == 0 {
		return fmt.Errorf("dataset has no points")
	}
	if len(d.EField) != len(d.Polarization) {
		return fmt.Errorf("dataset length mismatch: E has %d points, P has %d", len(d.EField), len(d.Polarization))
	}
	if len(d.EField) < 2 {
		return fmt.Errorf("dataset must contain at least 2 points for interpolation")
	}
	for i := range d.EField {
		if math.IsNaN(d.EField[i]) || math.IsInf(d.EField[i], 0) {
			return fmt.Errorf("invalid E value at index %d", i)
		}
		if math.IsNaN(d.Polarization[i]) || math.IsInf(d.Polarization[i], 0) {
			return fmt.Errorf("invalid P value at index %d", i)
		}
	}
	if !isMonotonicStrict(d.EField) {
		return fmt.Errorf("E-field range must be strictly monotonic (increasing or decreasing)")
	}
	return nil
}

func isMonotonicStrict(v []float64) bool {
	if len(v) < 2 {
		return true
	}
	inc := v[1] > v[0]
	if v[1] == v[0] {
		return false
	}
	for i := 1; i < len(v); i++ {
		if inc && v[i] <= v[i-1] {
			return false
		}
		if !inc && v[i] >= v[i-1] {
			return false
		}
	}
	return true
}

func ComputeFitMetrics(sim, lit *LiteratureDataset) FitMetrics {
	if sim == nil || lit == nil || len(sim.EField) < 2 || len(lit.EField) == 0 || len(lit.EField) != len(lit.Polarization) {
		return FitMetrics{}
	}
	xs, ys := sortedPairs(sim.EField, sim.Polarization)
	if len(xs) < 2 {
		return FitMetrics{}
	}

	m := FitMetrics{}
	var sumSq, sumAbs float64
	for i := range lit.EField {
		psim := interpLinearClamp(xs, ys, lit.EField[i])
		err := math.Abs(psim - lit.Polarization[i])
		sumSq += err * err
		sumAbs += err
		if err > m.MaxErr {
			m.MaxErr = err
		}
		m.NSamples++
	}
	if m.NSamples > 0 {
		m.RMSE = math.Sqrt(sumSq / float64(m.NSamples))
		m.MAE = sumAbs / float64(m.NSamples)
	}
	return m
}

func sortedPairs(x, y []float64) ([]float64, []float64) {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return x[idx[i]] < x[idx[j]] })
	xs := make([]float64, 0, n)
	ys := make([]float64, 0, n)
	for _, i := range idx {
		if len(xs) > 0 && x[i] == xs[len(xs)-1] {
			continue
		}
		xs = append(xs, x[i])
		ys = append(ys, y[i])
	}
	return xs, ys
}

func interpLinearClamp(x, y []float64, q float64) float64 {
	if q <= x[0] {
		return y[0]
	}
	last := len(x) - 1
	if q >= x[last] {
		return y[last]
	}
	i := sort.Search(len(x), func(i int) bool { return x[i] >= q })
	if i == 0 {
		return y[0]
	}
	x0, x1 := x[i-1], x[i]
	y0, y1 := y[i-1], y[i]
	t := (q - x0) / (x1 - x0)
	return y0 + t*(y1-y0)
}

func SaveOverlayProfile(path string, d *LiteratureDataset) error {
	if err := validateDataset(d); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}
	return nil
}

func LoadOverlayProfile(path string) (*LiteratureDataset, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}
	var d LiteratureDataset
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("malformed profile JSON: %w", err)
	}
	if err := validateDataset(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (a *App) createLiteratureOverlayPanel() fyne.CanvasObject {
	a.literatureSummaryLabel = widget.NewLabel("Literature: none loaded")
	a.literatureMetricsLabel = widget.NewLabel("Fit: RMSE = -, MAE = -")
	a.literatureTableLabel = widget.NewLabel("E(V/m)\tP_sim(µC/cm²)\tP_lit(µC/cm²)\terror(µC/cm²)")
	a.literatureTableLabel.Wrapping = fyne.TextWrapWord

	box := container.NewVBox(
		a.literatureSummaryLabel,
		a.literatureMetricsLabel,
		widget.NewSeparator(),
		a.literatureTableLabel,
	)
	return widget.NewCard("Literature Overlay", "Simulated vs measured", box)
}

func (a *App) SetLiteratureDataset(ds *LiteratureDataset) {
	a.mu.Lock()
	a.literatureDataset = ds
	a.updateLiteratureOverlayLocked()
	a.mu.Unlock()
}

func (a *App) updateLiteratureOverlayLocked() {
	a.updateLiteratureOverlayFromData(a.eHistory, a.pHistory)
}

func (a *App) updateLiteratureOverlayFromData(eHist, pHist []float64) {
	if a.literatureSummaryLabel == nil || a.literatureMetricsLabel == nil || a.literatureTableLabel == nil {
		return
	}
	if a.literatureDataset == nil {
		a.literatureSummaryLabel.SetText("Literature: none loaded")
		a.literatureMetricsLabel.SetText("Fit: RMSE = -, MAE = -")
		a.literatureTableLabel.SetText("E(V/m)\tP_sim(µC/cm²)\tP_lit(µC/cm²)\terror(µC/cm²)")
		return
	}
	lit := a.literatureDataset
	sim := &LiteratureDataset{EField: append([]float64(nil), eHist...), Polarization: append([]float64(nil), pHist...)}
	m := ComputeFitMetrics(sim, lit)
	a.literatureSummaryLabel.SetText(fmt.Sprintf("Literature: %s, N=%d points", safeText(lit.Source, "(unspecified source)"), len(lit.EField)))
	a.literatureMetricsLabel.SetText(fmt.Sprintf("Fit: RMSE = %.3f µC/cm², MAE = %.3f µC/cm², MaxErr = %.3f µC/cm²", m.RMSE*cm2ToUCCm2, m.MAE*cm2ToUCCm2, m.MaxErr*cm2ToUCCm2))

	xs, ys := sortedPairs(sim.EField, sim.Polarization)
	lines := []string{"E(V/m)\tP_sim(µC/cm²)\tP_lit(µC/cm²)\terror(µC/cm²)"}
	rows := len(lit.EField)
	if rows > defaultRows {
		rows = defaultRows
	}
	for i := 0; i < rows && len(xs) >= 2; i++ {
		psim := interpLinearClamp(xs, ys, lit.EField[i])
		err := (psim - lit.Polarization[i]) * cm2ToUCCm2
		lines = append(lines, fmt.Sprintf("%.3e\t%.3f\t%.3f\t%.3f", lit.EField[i], psim*cm2ToUCCm2, lit.Polarization[i]*cm2ToUCCm2, err))
	}
	if len(lit.EField) > rows {
		lines = append(lines, fmt.Sprintf("... (%d more rows)", len(lit.EField)-rows))
	}
	a.literatureTableLabel.SetText(strings.Join(lines, "\n"))
}

func safeText(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
