package hysteresis

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sharedio "fecim-lattice-tools/shared/io"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state HysteresisState
}

func New() *Module {
	materials := physics.AllMaterials()
	defaultMat := "HZO (Si-doped, Park 2015 midpoint)"
	if len(materials) > 0 && materials[0] != nil {
		defaultMat = materials[0].Name
	}
	m := &Module{
		state: HysteresisState{
			SelectedMaterial: defaultMat,
			Materials:        materials,
			FieldRange:       FieldRange{MinField: -3000, MaxField: 3000},
			Waveform:         WaveformSine,
		},
	}
	m.computeLoopForCurrentMaterial()
	return m
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleHysteresis,
		Title:       "FeCIM Hysteresis Simulation",
		Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case EventSelectMaterial:
		if name, ok := action.Payload["material"]; ok {
			for _, mat := range m.state.Materials {
				if mat != nil && mat.Name == name {
					m.state.SelectedMaterial = name
					m.computeLoopForCurrentMaterial()
					return nil
				}
			}
		}
		return fmt.Errorf("hysteresis: material %q not found", action.Payload["material"])
	case EventToggleSimulation:
		m.state.IsRunning = !m.state.IsRunning
		return nil
	case EventSetFieldRange:
		if minS, ok := action.Payload["min"]; ok {
			fmt.Sscanf(minS, "%f", &m.state.FieldRange.MinField)
		}
		if maxS, ok := action.Payload["max"]; ok {
			fmt.Sscanf(maxS, "%f", &m.state.FieldRange.MaxField)
		}
		m.computeLoopForCurrentMaterial()
		return nil
	case EventSetWaveform:
		return m.setWaveform(action.Payload["waveform"])
	case EventExportCSV:
		return m.exportCSV(action.Payload["path"])
	case EventRunPUND:
		return m.runPUND()
	case EventRunFORC:
		return m.runFORC(action.Payload)
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) setWaveform(waveform string) error {
	if !isValidWaveform(waveform) {
		return fmt.Errorf("hysteresis: unsupported waveform %q", waveform)
	}
	m.state.Waveform = waveform
	m.computeLoopForCurrentMaterial()
	return nil
}

func isValidWaveform(waveform string) bool {
	switch waveform {
	case WaveformSine, WaveformTriangle, WaveformSquare, WaveformManual:
		return true
	default:
		return false
	}
}

func (m *Module) exportCSV(path string) error {
	content, err := buildLoopCSV(m.state.LoopPoints)
	if err != nil {
		return err
	}

	exportPath := "artifact buffer"
	statusVerb := "buffered"
	if trimmed := strings.TrimSpace(path); trimmed != "" {
		cleanPath, err := sharedio.ValidatePath(trimmed)
		if err != nil {
			return fmt.Errorf("hysteresis: invalid CSV export path: %w", err)
		}
		if err := writeTextArtifact(cleanPath, content); err != nil {
			return fmt.Errorf("hysteresis: write CSV export: %w", err)
		}
		exportPath = cleanPath
		statusVerb = "wrote"
	}

	m.state.CSVExportStatus = fmt.Sprintf("%s %d points", statusVerb, len(m.state.LoopPoints))
	m.state.CSVExportPath = exportPath
	m.state.CSVExportBytes = len(content)
	m.state.CSVExportContent = content
	return nil
}

func buildLoopCSV(points []LoopPoint) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"Index", "E_field_kV_cm", "Polarization_uC_cm2"}); err != nil {
		return "", err
	}
	for i, point := range points {
		if err := writer.Write([]string{
			strconv.Itoa(i),
			strconv.FormatFloat(point.Field, 'f', 6, 64),
			strconv.FormatFloat(point.Polarization, 'f', 6, 64),
		}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeTextArtifact(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *Module) runPUND() error {
	mat := m.currentMaterial()
	ec, ps, area := materialPUNDFORCParameters(mat)
	stack := physics.NewPreisachStack(ec, &physics.TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})
	if stack == nil {
		return fmt.Errorf("hysteresis: could not initialize PUND Preisach stack")
	}
	result, traces, err := physics.RunPUNDSimulation(stack, 5*ec, 100e-9, 5e-9, area)
	if err != nil {
		return fmt.Errorf("hysteresis: run PUND: %w", err)
	}
	ratio := 0.0
	if result.SwitchingNegative_C != 0 {
		ratio = math.Abs(result.SwitchingPositive_C / result.SwitchingNegative_C)
	}
	samplesPerPulse := 0
	if len(traces) > 0 {
		samplesPerPulse = len(traces[0])
	}
	m.state.PUND = PUNDSummary{
		Available:         true,
		QP_C:              result.QP_C,
		QU_C:              result.QU_C,
		QN_C:              result.QN_C,
		QD_C:              result.QD_C,
		SwitchingPositive: result.SwitchingPositive_C,
		SwitchingNegative: result.SwitchingNegative_C,
		SwitchingRatio:    ratio,
		SamplesPerPulse:   samplesPerPulse,
		Summary: fmt.Sprintf(
			"QP=%.3e C, QU=%.3e C, QN=%.3e C, QD=%.3e C; Qsw+=%.3e C, Qsw-=%.3e C; Switching ratio |Qsw+/Qsw-|=%.3f; samples_per_pulse=%d",
			result.QP_C, result.QU_C, result.QN_C, result.QD_C,
			result.SwitchingPositive_C, result.SwitchingNegative_C, ratio, samplesPerPulse,
		),
	}
	return nil
}

func (m *Module) runFORC(payload map[string]string) error {
	mat := m.currentMaterial()
	ec, ps, _ := materialPUNDFORCParameters(mat)
	emax := math.Max(math.Abs(m.state.FieldRange.MinField), math.Abs(m.state.FieldRange.MaxField)) * 1e5
	if emax < 2*ec {
		emax = 2 * ec
	}
	reversals := 21
	if raw := strings.TrimSpace(payload["reversals"]); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("hysteresis: invalid FORC reversals %q", raw)
		}
		reversals = parsed
	}
	if reversals < 3 {
		return fmt.Errorf("hysteresis: FORC reversals must be >= 3")
	}

	stack := physics.NewPreisachStack(emax, &physics.TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})
	if stack == nil {
		return fmt.Errorf("hysteresis: could not initialize FORC Preisach stack")
	}
	result, err := physics.RunFORCSweep(stack, emax, reversals)
	if err != nil {
		return fmt.Errorf("hysteresis: run FORC: %w", err)
	}
	m.state.FORC = summarizeFORCResult(result)
	return nil
}

func materialPUNDFORCParameters(mat *physics.HZOMaterial) (ec, ps, area float64) {
	ec = 3e7
	ps = 0.25
	area = 1e-12
	if mat == nil {
		return ec, ps, area
	}
	if mat.Ec > 0 {
		ec = mat.Ec
	}
	if mat.Ps > 0 {
		ps = mat.Ps
	}
	if mat.Area > 0 {
		area = mat.Area
	}
	return ec, ps, area
}

func summarizeFORCResult(result physics.FORCResult) FORCSummary {
	rows := len(result.PreisachDensity)
	cols := 0
	if rows > 0 {
		cols = len(result.PreisachDensity[0])
	}
	minD, maxD := math.Inf(1), math.Inf(-1)
	peak := 0.0
	peakI, peakJ := 0, 0
	visited := false
	for i := 0; i < rows; i++ {
		for j := 0; j < len(result.PreisachDensity[i]); j++ {
			v := result.PreisachDensity[i][j]
			visited = true
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
	if !visited {
		minD, maxD = 0, 0
	}
	peakEa, peakEb := 0.0, 0.0
	if peakJ >= 0 && peakJ < len(result.ReversalFields_Vm) {
		peakEa = result.ReversalFields_Vm[peakJ]
	}
	if peakI >= 0 && peakI < len(result.ReversalFields_Vm) {
		peakEb = result.ReversalFields_Vm[peakI]
	}
	summary := fmt.Sprintf(
		"curves=%d, density_grid=%dx%d, peak_density=%.3e at (Ea=%.3e V/m, Eb=%.3e V/m), density_range=[%.3e, %.3e]",
		len(result.Curves), rows, cols, peak, peakEa, peakEb, minD, maxD,
	)
	return FORCSummary{
		Available:   true,
		Curves:      len(result.Curves),
		DensityRows: rows,
		DensityCols: cols,
		PeakDensity: peak,
		PeakEa_Vm:   peakEa,
		PeakEb_Vm:   peakEb,
		MinDensity:  minD,
		MaxDensity:  maxD,
		Summary:     summary,
	}
}

func (m *Module) currentMaterial() *physics.HZOMaterial {
	for _, candidate := range m.state.Materials {
		if candidate != nil && candidate.Name == m.state.SelectedMaterial {
			return candidate
		}
	}
	return nil
}

func (m *Module) computeLoopForCurrentMaterial() {
	mat := m.currentMaterial()
	if mat == nil {
		return
	}

	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	maxFieldKVcm := math.Max(math.Abs(m.state.FieldRange.MinField), math.Abs(m.state.FieldRange.MaxField))
	ecKVcm := mat.Ec * 1e-5
	if maxFieldKVcm < ecKVcm*2 {
		maxFieldKVcm = ecKVcm * 2
	}
	maxFieldSI := maxFieldKVcm * 1e5

	const numPoints = 200
	dt := 5e-5 // ~100Hz equivalent quasi-static timestep

	for cycle := 0; cycle < 2; cycle++ {
		for i := 0; i < numPoints; i++ {
			fieldSI := waveformField(i, numPoints, maxFieldSI, m.state.Waveform)
			solver.Step(fieldSI, dt)
		}
	}

	pts := make([]LoopPoint, numPoints)
	for i := 0; i < numPoints; i++ {
		fieldSI := waveformField(i, numPoints, maxFieldSI, m.state.Waveform)
		polSI := solver.Step(fieldSI, dt)
		pts[i] = LoopPoint{
			Field:        fieldSI * 1e-5, // V/m → kV/cm
			Polarization: polSI * 100,    // C/m² → µC/cm²
		}
	}
	m.state.LoopPoints = pts
	m.computeLoopMetrics()
}

func waveformField(index, numPoints int, maxFieldSI float64, waveform string) float64 {
	if numPoints <= 1 {
		return 0
	}
	phase := float64(index) / float64(numPoints-1)
	switch waveform {
	case WaveformTriangle:
		if phase <= 0.5 {
			return maxFieldSI * (-1 + 4*phase)
		}
		return maxFieldSI * (3 - 4*phase)
	case WaveformSquare:
		if phase < 0.5 {
			return maxFieldSI
		}
		return -maxFieldSI
	case WaveformManual:
		return 0
	default:
		return maxFieldSI * math.Sin(2*math.Pi*phase)
	}
}

func (m *Module) computeLoopMetrics() {
	pts := m.state.LoopPoints
	if len(pts) < 4 {
		return
	}

	minP, maxP := pts[0].Polarization, pts[0].Polarization
	for _, p := range pts {
		if p.Polarization < minP {
			minP = p.Polarization
		}
		if p.Polarization > maxP {
			maxP = p.Polarization
		}
	}
	m.state.Psat = maxP
	m.state.PsatNeg = minP

	prPos, prNeg := 0.0, 0.0
	ecPos, ecNeg := 0.0, 0.0
	for i := 1; i < len(pts); i++ {
		if (pts[i-1].Field < 0 && pts[i].Field >= 0) || (pts[i-1].Field > 0 && pts[i].Field <= 0) {
			interp := pts[i-1].Polarization + (0.0-pts[i-1].Field)*(pts[i].Polarization-pts[i-1].Polarization)/(pts[i].Field-pts[i-1].Field+1e-12)
			if interp > 0 {
				prPos = interp
			} else {
				prNeg = interp
			}
		}
		if (pts[i-1].Polarization < 0 && pts[i].Polarization >= 0) || (pts[i-1].Polarization > 0 && pts[i].Polarization <= 0) {
			interp := pts[i-1].Field + (0.0-pts[i-1].Polarization)*(pts[i].Field-pts[i-1].Field)/(pts[i].Polarization-pts[i-1].Polarization+1e-12)
			if pts[i-1].Polarization < 0 || (pts[i-1].Polarization == 0 && pts[i].Polarization > 0 && interp > 0) {
				ecPos = math.Abs(interp)
			} else {
				ecNeg = math.Abs(interp)
			}
		}
	}
	if math.Abs(prNeg) > math.Abs(prPos) {
		m.state.Pr = math.Abs(prNeg)
	} else {
		m.state.Pr = prPos
	}
	m.state.EcPlus = ecPos
	m.state.EcMinus = ecNeg

	area := 0.0
	for i := 0; i < len(pts); i++ {
		j := (i + 1) % len(pts)
		area += pts[i].Field * pts[j].Polarization
		area -= pts[j].Field * pts[i].Polarization
	}
	m.state.LoopArea = math.Abs(area) * 0.5

	times := []float64{1, 10, 100, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8}
	m.state.RetentionTimes = times
	m.state.RetentionPr = make([]float64, len(times))
	if m.state.Pr > 0 {
		prSI := m.state.Pr * 0.01
		points, err := physics.SimulateRetentionPowerLaw(prSI, 1.0, 0.02, times)
		if err == nil {
			for i, p := range points {
				m.state.RetentionPr[i] = p.Polarization_Cm * 100
			}
		}
	}
}
