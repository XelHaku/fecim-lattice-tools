package circuits

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	sharedio "fecim-lattice-tools/shared/io"
	"fecim-lattice-tools/shared/mathutil"
	"fecim-lattice-tools/shared/peripherals"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/shared/viewmodel/design"
)

type Module struct{ state CircuitsState }

type OperationLogExport struct {
	Schema              string              `json:"schema"`
	Module              string              `json:"module"`
	OperationMode       string              `json:"operation_mode"`
	Architecture        string              `json:"architecture"`
	Rows                int                 `json:"rows"`
	Cols                int                 `json:"cols"`
	SelectedRow         int                 `json:"selected_row"`
	SelectedCol         int                 `json:"selected_col"`
	WriteTargetLevel    int                 `json:"write_target_level"`
	LastOperationStatus string              `json:"last_operation_status"`
	OperationLogTotal   int                 `json:"operation_log_total"`
	ExportedEntries     int                 `json:"exported_entries"`
	Entries             []OperationLogEntry `json:"entries"`
	ComputeRun          *ComputeRunLog      `json:"compute_run,omitempty"`
}

func New() *Module {
	m := &Module{state: CircuitsState{
		Rows: 8, Cols: 8,
		OperationMode:            OperationRead,
		Architecture:             ArchitecturePassive,
		WriteTargetLevel:         DefaultQuantLevels / 2,
		QuantLevels:              DefaultQuantLevels,
		CouplingTier:             CouplingTierA,
		ISPPEngine:               ISPPEngineLevel,
		LoggerVerbosity:          "off",
		TimingOperation:          "READ",
		TimingPlaybackIntervalMS: DefaultTimingPlaybackIntervalMS,
		ADCResolution:            5, DACResolution: 5, TIAGain: 1e4,
		ChargePumpStages: 4, SupplyVoltage: 1.8, ISPPEnabled: true,
	}}
	m.runISPPSimulation()
	m.computeHalfSelectStress()
	m.computeReferenceSpecs()
	m.computeReferenceTiming()
	return m
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleCircuits,
		Title:       "FeCIM Peripheral Circuits Visualizer",
		Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case ActionRunRead:
		m.state.OperationMode = OperationRead
		m.recordStatus("operation", "READ cell [%d,%d] through %s", m.state.SelectedRow, m.state.SelectedCol, m.state.Architecture)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
		return nil
	case ActionRunWrite:
		m.state.OperationMode = OperationWrite
		m.runISPPSimulation()
		m.recordStatus("operation", "WRITE level %d to cell [%d,%d] using %s", m.state.WriteTargetLevel, m.state.SelectedRow, m.state.SelectedCol, m.state.ISPPEngine)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
		return nil
	case ActionRunCompute:
		m.state.OperationMode = OperationCompute
		m.state.ComputeRunLog = m.buildComputeRunLog()
		m.recordStatus("operation", "COMPUTE on %dx%d %s array", m.state.Rows, m.state.Cols, m.state.Architecture)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
		return nil
	case ActionExportOperationLog:
		return m.exportOperationLog(action.Payload)
	case ActionExportReferenceSpecs:
		return m.exportReferenceSpecs(action.Payload)
	case ActionExportReferenceTiming:
		return m.exportReferenceTiming(action.Payload)
	case ActionExportReferenceTimingSVG:
		return m.exportReferenceTimingSVG(action.Payload)
	case ActionAnimateReferenceTiming:
		return m.animateReferenceTiming()
	case ActionPlayReferenceTiming:
		return m.playReferenceTiming(action.Payload)
	case ActionPauseReferenceTiming:
		return m.pauseReferenceTiming()
	case ActionStepReferenceTiming:
		return m.stepReferenceTiming()
	case ActionResetReferenceTiming:
		return m.resetReferenceTiming()
	case ActionToggleISPP:
		m.state.ISPPEnabled = !m.state.ISPPEnabled
		m.recordStatus("control", "ISPP enabled: %v", m.state.ISPPEnabled)
		return nil
	case ActionResizeArray:
		return m.resizeArray(action.Payload)
	case ActionSetOperationMode:
		return m.setOperationMode(action.Payload)
	case ActionSetArchitecture:
		return m.setArchitecture(action.Payload)
	case ActionSelectCell:
		return m.selectCell(action.Payload)
	case ActionSetWriteTarget:
		return m.setWriteTarget(action.Payload)
	case ActionSetDACBits:
		return m.setDACBits(action.Payload)
	case ActionSetADCBits:
		return m.setADCBits(action.Payload)
	case ActionSetTIAGain:
		return m.setTIAGain(action.Payload)
	case ActionSetCouplingTier:
		return m.setCouplingTier(action.Payload)
	case ActionSetISPPEngine:
		return m.setISPPEngine(action.Payload)
	case ActionSetTimingOperation:
		return m.setTimingOperation(action.Payload)
	case ActionSetLoggerVerbosity:
		return m.setLoggerVerbosity(action.Payload)
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

func (m *Module) resizeArray(payload map[string]string) error {
	rows, cols := m.state.Rows, m.state.Cols
	if sizeS, ok := payload["size"]; ok {
		size, err := parseInt(sizeS, "array size")
		if err != nil {
			return err
		}
		rows, cols = size, size
	}
	if rowsS, ok := payload["rows"]; ok {
		parsedRows, err := parseInt(rowsS, "rows")
		if err != nil {
			return err
		}
		rows = parsedRows
	}
	if colsS, ok := payload["cols"]; ok {
		parsedCols, err := parseInt(colsS, "cols")
		if err != nil {
			return err
		}
		cols = parsedCols
	}
	if !validArraySize(rows) || !validArraySize(cols) {
		return fmt.Errorf("circuits: invalid array size %dx%d", rows, cols)
	}
	m.state.Rows = rows
	m.state.Cols = cols
	m.clampSelectedCell()
	m.recordStatus("control", "Array resized to %dx%d", rows, cols)
	m.computeHalfSelectStress()
	m.computeReferenceSpecs()
	return nil
}

func (m *Module) setOperationMode(payload map[string]string) error {
	mode, err := viewmodel.PayloadString(payload, "mode")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	if !viewmodel.PayloadStringIn(mode, OperationRead, OperationWrite, OperationCompute) {
		return fmt.Errorf("circuits: invalid operation mode %q", mode)
	}
	m.state.OperationMode = mode
	m.recordStatus("control", "Operation mode set to %s", mode)
	m.computeHalfSelectStress()
	return nil
}

func (m *Module) setTimingOperation(payload map[string]string) error {
	operation, err := viewmodel.PayloadString(payload, "operation")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	rawOperation := operation
	operation = strings.ToUpper(strings.TrimSpace(operation))
	if !viewmodel.PayloadStringIn(operation, "READ", "WRITE", "COMPUTE") {
		return fmt.Errorf("circuits: invalid timing operation %q", rawOperation)
	}
	m.state.TimingOperation = operation
	m.computeReferenceTiming()
	m.recordStatus("control", "Timing operation set to %s", operation)
	return nil
}

func (m *Module) setArchitecture(payload map[string]string) error {
	architecture, err := viewmodel.PayloadString(payload, "architecture")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	if !viewmodel.PayloadStringIn(architecture, ArchitecturePassive, Architecture1T1R, Architecture2T1R) {
		return fmt.Errorf("circuits: invalid architecture %q", architecture)
	}
	m.state.Architecture = architecture
	m.recordStatus("control", "Architecture set to %s", architecture)
	m.computeHalfSelectStress()
	return nil
}

func (m *Module) selectCell(payload map[string]string) error {
	row, err := parsePayloadInt(payload, "row")
	if err != nil {
		return err
	}
	col, err := parsePayloadInt(payload, "col")
	if err != nil {
		return err
	}
	if row < 0 || row >= m.state.Rows || col < 0 || col >= m.state.Cols {
		return fmt.Errorf("circuits: selected cell [%d,%d] outside %dx%d array", row, col, m.state.Rows, m.state.Cols)
	}
	m.state.SelectedRow = row
	m.state.SelectedCol = col
	m.recordStatus("control", "Selected cell [%d,%d]", row, col)
	m.computeHalfSelectStress()
	return nil
}

func (m *Module) setWriteTarget(payload map[string]string) error {
	level, err := parsePayloadInt(payload, "level")
	if err != nil {
		return err
	}
	if m.state.QuantLevels <= 0 {
		m.state.QuantLevels = DefaultQuantLevels
	}
	if level < 0 || level >= m.state.QuantLevels {
		return fmt.Errorf("circuits: target level %d outside 0-%d", level, m.state.QuantLevels-1)
	}
	m.state.WriteTargetLevel = level
	m.recordStatus("control", "Write target set to level %d", level)
	m.computeHalfSelectStress()
	return nil
}

func (m *Module) setDACBits(payload map[string]string) error {
	bits, err := parsePayloadInt(payload, "bits")
	if err != nil {
		return err
	}
	if bits < 4 || bits > 8 {
		return fmt.Errorf("circuits: DAC bits must be 4-8, got %d", bits)
	}
	m.state.DACResolution = bits
	m.computeReferenceSpecs()
	m.recordStatus("control", "DAC resolution set to %d bits", bits)
	return nil
}

func (m *Module) setADCBits(payload map[string]string) error {
	bits, err := parsePayloadInt(payload, "bits")
	if err != nil {
		return err
	}
	if bits < 5 || bits > 8 {
		return fmt.Errorf("circuits: ADC bits must be 5-8, got %d", bits)
	}
	m.state.ADCResolution = bits
	m.computePVTCorners()
	m.computeReferenceSpecs()
	m.recordStatus("control", "ADC resolution set to %d bits", bits)
	return nil
}

func (m *Module) setTIAGain(payload map[string]string) error {
	gain, err := viewmodel.PayloadFloat(payload, "gain_ohm")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	if gain <= 0 {
		return fmt.Errorf("circuits: TIA gain must be positive, got %.3g", gain)
	}
	m.state.TIAGain = gain
	m.computePVTCorners()
	m.computeReferenceSpecs()
	m.recordStatus("control", "TIA gain set to %.0f ohm", gain)
	return nil
}

func (m *Module) setCouplingTier(payload map[string]string) error {
	tier, err := viewmodel.PayloadString(payload, "tier")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	if !viewmodel.PayloadStringIn(tier, CouplingIdeal, CouplingTierA, CouplingTierB) {
		return fmt.Errorf("circuits: invalid coupling tier %q", tier)
	}
	m.state.CouplingTier = tier
	m.recordStatus("control", "Coupling tier set to %s", tier)
	return nil
}

func (m *Module) setISPPEngine(payload map[string]string) error {
	engine, err := viewmodel.PayloadString(payload, "engine")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	if !viewmodel.PayloadStringIn(engine, ISPPEngineLevel, ISPPEngineLK) {
		return fmt.Errorf("circuits: invalid ISPP engine %q", engine)
	}
	m.state.ISPPEngine = engine
	m.recordStatus("control", "ISPP engine set to %s", engine)
	return nil
}

func (m *Module) setLoggerVerbosity(payload map[string]string) error {
	verbosity, err := viewmodel.PayloadString(payload, "verbosity")
	if err != nil {
		return fmt.Errorf("circuits: %w", err)
	}
	level, label, err := parseLoggerVerbosity(verbosity)
	if err != nil {
		return err
	}
	m.state.LoggerVerbosity = label
	m.state.LoggerVerbosityLevel = level
	m.recordStatus("control", "Logger verbosity set to %s", label)
	return nil
}

func (m *Module) recordStatus(kind, format string, args ...interface{}) {
	status := fmt.Sprintf(format, args...)
	m.state.LastOperationStatus = status
	m.state.OperationLogTotal++
	entry := OperationLogEntry{
		Sequence: m.state.OperationLogTotal,
		Kind:     kind,
		Message:  status,
	}
	m.state.OperationLog = append(m.state.OperationLog, entry)
	if len(m.state.OperationLog) > OperationLogLimit {
		m.state.OperationLog = m.state.OperationLog[len(m.state.OperationLog)-OperationLogLimit:]
	}
}

func (m *Module) exportOperationLog(payload map[string]string) error {
	path := payload["path"]

	export := m.operationLogExportPayload()
	artifact, err := sharedio.BufferOrWriteJSONArtifact(path, export)
	if err != nil {
		return fmt.Errorf("circuits: operation log export artifact: %w", err)
	}
	m.state.OperationLogExportStatus = fmt.Sprintf("%s %d entries", artifact.StatusVerb, export.ExportedEntries)
	m.state.OperationLogExportPath = artifact.Path
	m.state.OperationLogExportBytes = artifact.Bytes
	m.state.OperationLogExportJSON = artifact.Content
	return nil
}

func (m *Module) operationLogExportPayload() OperationLogExport {
	entries := make([]OperationLogEntry, len(m.state.OperationLog))
	copy(entries, m.state.OperationLog)
	export := OperationLogExport{
		Schema:              "fecim.circuits.operation_log.v1",
		Module:              string(viewmodel.ModuleCircuits),
		OperationMode:       m.state.OperationMode,
		Architecture:        m.state.Architecture,
		Rows:                m.state.Rows,
		Cols:                m.state.Cols,
		SelectedRow:         m.state.SelectedRow,
		SelectedCol:         m.state.SelectedCol,
		WriteTargetLevel:    m.state.WriteTargetLevel,
		LastOperationStatus: m.state.LastOperationStatus,
		OperationLogTotal:   m.state.OperationLogTotal,
		ExportedEntries:     len(entries),
		Entries:             entries,
	}
	if m.latestOperationLogIsCompute() && m.state.ComputeRunLog != nil {
		export.ComputeRun = cloneComputeRunLog(m.state.ComputeRunLog)
	}
	return export
}

func (m *Module) exportReferenceSpecs(payload map[string]string) error {
	path := payload["path"]

	export := m.referenceSpecExportPayload()
	artifact, err := sharedio.BufferOrWriteJSONArtifact(path, export)
	if err != nil {
		return fmt.Errorf("circuits: reference spec export artifact: %w", err)
	}
	m.state.ReferenceSpecExportStatus = fmt.Sprintf("%s %d cells", artifact.StatusVerb, export.Cells)
	m.state.ReferenceSpecExportPath = artifact.Path
	m.state.ReferenceSpecExportBytes = artifact.Bytes
	m.state.ReferenceSpecExportJSON = artifact.Content
	return nil
}

func (m *Module) referenceSpecExportPayload() ReferenceSpecExport {
	m.computeReferenceSpecs()
	quantLevels := m.state.QuantLevels
	if quantLevels <= 0 {
		quantLevels = DefaultQuantLevels
	}
	return ReferenceSpecExport{
		Schema:          "fecim.circuits.reference_specs.v1",
		Module:          string(viewmodel.ModuleCircuits),
		Rows:            m.state.Rows,
		Cols:            m.state.Cols,
		QuantLevels:     quantLevels,
		Cells:           m.state.SpecCells,
		BitsPerCell:     m.state.SpecBitsPerCell,
		DACCount:        m.state.SpecDACCount,
		TIACount:        m.state.SpecTIACount,
		ADCCount:        m.state.SpecADCCount,
		DACCodes:        m.state.SpecDACCodes,
		ADCCodes:        m.state.SpecADCCodes,
		TotalPowerMW:    m.state.SpecTotalPowerMW,
		LatencyNS:       m.state.SpecLatencyNS,
		ThroughputGOPS:  m.state.SpecThroughputGOPS,
		EfficiencyGOPSW: m.state.SpecEfficiencyGOPSW,
		Compliance:      m.state.SpecCompliance,
		BoundaryNotice:  "educational reference spec summary; power, latency, and throughput values are behavioral estimates, not calibrated silicon measurements.",
	}
}

func (m *Module) exportReferenceTiming(payload map[string]string) error {
	path := payload["path"]

	export := m.referenceTimingExportPayload()
	artifact, err := sharedio.BufferOrWriteJSONArtifact(path, export)
	if err != nil {
		return fmt.Errorf("circuits: reference timing export artifact: %w", err)
	}
	m.state.ReferenceTimingExportStatus = fmt.Sprintf("%s %d operations", artifact.StatusVerb, len(export.Operations))
	m.state.ReferenceTimingExportPath = artifact.Path
	m.state.ReferenceTimingExportBytes = artifact.Bytes
	m.state.ReferenceTimingExportJSON = artifact.Content
	return nil
}

func (m *Module) exportReferenceTimingSVG(payload map[string]string) error {
	path := payload["path"]

	m.computeReferenceTiming()
	waveform, ok := activeTimingWaveform(m.state)
	if !ok {
		return fmt.Errorf("circuits: no active timing waveform for %q", m.state.TimingActiveOp)
	}
	svg := buildReferenceTimingSVG(waveform)
	artifact, err := sharedio.BufferOrWriteTextArtifact(path, svg)
	if err != nil {
		return fmt.Errorf("circuits: write reference timing SVG export: %w", err)
	}
	m.state.ReferenceTimingSVGExportStatus = fmt.Sprintf("%s %s waveform", artifact.StatusVerb, waveform.Operation)
	m.state.ReferenceTimingSVGExportPath = artifact.Path
	m.state.ReferenceTimingSVGExportBytes = len(svg)
	m.state.ReferenceTimingSVGExport = svg
	return nil
}

func (m *Module) referenceTimingExportPayload() ReferenceTimingExport {
	m.computeReferenceTiming()
	return ReferenceTimingExport{
		Schema:          "fecim.circuits.reference_timing.v1",
		Module:          string(viewmodel.ModuleCircuits),
		OperationMode:   m.state.OperationMode,
		WriteTotalNS:    m.state.TimingWriteTotalNS,
		ReadTotalNS:     m.state.TimingReadTotalNS,
		ComputeTotalNS:  m.state.TimingComputeTotalNS,
		ActiveOperation: m.state.TimingActiveOp,
		ActiveTotalNS:   m.state.TimingActiveTotalNS,
		ActivePhases:    m.state.TimingActivePhases,
		Operations: []ReferenceTimingOperation{
			{
				Operation: "WRITE",
				TotalNS:   m.state.TimingWriteTotalNS,
				Phases: []ReferenceTimingPhase{
					{Name: "DAC", DurationNS: 10},
					{Name: "Pump", DurationNS: 88},
					{Name: "Pulse", DurationNS: 100},
					{Name: "Array", DurationNS: 5},
				},
			},
			{
				Operation: "READ",
				TotalNS:   m.state.TimingReadTotalNS,
				Phases: []ReferenceTimingPhase{
					{Name: "DAC", DurationNS: 10},
					{Name: "Array", DurationNS: 5},
					{Name: "TIA", DurationNS: 11},
					{Name: "ADC", DurationNS: 50},
				},
			},
			{
				Operation: "COMPUTE",
				TotalNS:   m.state.TimingComputeTotalNS,
				Phases: []ReferenceTimingPhase{
					{Name: "DAC", DurationNS: 10},
					{Name: "Array", DurationNS: 5},
					{Name: "TIA+ADC", DurationNS: 61},
				},
			},
		},
		BoundaryNotice: "educational reference timing summary; phase durations are behavioral estimates, not calibrated silicon measurements.",
	}
}

func buildReferenceTimingSVG(waveform ReferenceTimingWaveform) string {
	const (
		width     = 920
		left      = 150.0
		right     = 40.0
		top       = 82.0
		rowGap    = 46.0
		rowHeight = 22.0
	)
	signalRows := maxInt(1, len(waveform.Signals))
	height := 178 + signalRows*46
	plotWidth := float64(width) - left - right
	axisY := top + float64(signalRows)*rowGap + 6

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="Inter, Helvetica, Arial, sans-serif">
  <title>%s Timing Waveform</title>
  <defs>
    <style>
      .bg { fill: #ffffff; }
      .panel { fill: #f7faf8; stroke: #c8d7d0; stroke-width: 1; }
      .title { fill: #10251f; font-size: 22px; font-weight: 700; }
      .meta { fill: #52645e; font-size: 12px; }
      .label { fill: #233a33; font-size: 13px; font-weight: 600; }
      .base { stroke: #aebdb7; stroke-width: 1.2; }
      .high { fill: #2f7d68; opacity: 0.78; }
      .marker { stroke: #70847b; stroke-width: 1; stroke-dasharray: 4 4; }
      .phase { fill: #dfeae5; stroke: #aabbb3; stroke-width: 0.8; }
      .phase-label { fill: #40584f; font-size: 11px; }
      .notice { fill: #65756f; font-size: 11px; }
    </style>
  </defs>
  <rect class="bg" width="100%%" height="100%%"/>
  <rect class="panel" x="24" y="24" width="%d" height="%d" rx="8"/>
  <text class="title" x="44" y="58">%s Timing Waveform</text>
  <text class="meta" x="44" y="78">Total %d ns; generated from gogpu/ui-neutral Module 4 timing state.</text>
`, width, height, width, height, svgEscape(waveform.Operation), width-48, height-48, svgEscape(waveform.Operation), waveform.TotalNS))

	for _, marker := range waveform.TimeMarkers {
		x := left + float64(clampTimingPercent(marker.Percent, 0, 100))/100*plotWidth
		sb.WriteString(fmt.Sprintf("  <line class=\"marker\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n", x, top-12, x, axisY+8))
		sb.WriteString(fmt.Sprintf("  <text class=\"meta\" x=\"%.1f\" y=\"%.1f\" text-anchor=\"middle\">%s</text>\n", x, axisY+26, svgEscape(marker.Label)))
	}

	for i, signal := range waveform.Signals {
		y := top + float64(i)*rowGap
		sb.WriteString(fmt.Sprintf("  <text class=\"label\" x=\"44\" y=\"%.1f\">%s</text>\n", y+14, svgEscape(signal.Name)))
		sb.WriteString(fmt.Sprintf("  <line class=\"base\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n", left, y+rowHeight, left+plotWidth, y+rowHeight))
		for _, window := range signal.HighWindows {
			start := clampTimingPercent(window.StartPct, 0, 100)
			end := clampTimingPercent(window.EndPct, start, 100)
			if end <= start {
				continue
			}
			x := left + float64(start)/100*plotWidth
			w := float64(end-start) / 100 * plotWidth
			sb.WriteString(fmt.Sprintf("  <rect class=\"high\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" rx=\"2\"/>\n", x, y, w, rowHeight))
		}
	}

	phaseY := axisY + 44
	for _, phase := range waveform.PhaseMarkers {
		start := clampTimingPercent(phase.StartPct, 0, 100)
		end := clampTimingPercent(phase.EndPct, start, 100)
		if end <= start {
			continue
		}
		x := left + float64(start)/100*plotWidth
		w := float64(end-start) / 100 * plotWidth
		sb.WriteString(fmt.Sprintf("  <rect class=\"phase\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"24\" rx=\"3\"/>\n", x, phaseY, w))
		sb.WriteString(fmt.Sprintf("  <text class=\"phase-label\" x=\"%.1f\" y=\"%.1f\" text-anchor=\"middle\">%s %dns</text>\n", x+w/2, phaseY+16, svgEscape(phase.Label), phase.DurationNS))
	}

	sb.WriteString(fmt.Sprintf("  <text class=\"notice\" x=\"44\" y=\"%d\">educational reference timing waveform; behavioral estimates, not calibrated silicon timing.</text>\n", height-34))
	sb.WriteString("</svg>\n")
	return sb.String()
}

func svgEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return replacer.Replace(s)
}

func (m *Module) animateReferenceTiming() error {
	if err := m.ensureTimingPlaybackSteps(); err != nil {
		return err
	}
	operation := m.state.TimingAnimationOperation
	m.state.TimingAnimationStepIndex = 1
	m.state.TimingAnimationCurrentStep = m.state.TimingAnimationSteps[0]
	m.state.TimingAnimationStatus = fmt.Sprintf("%s timing animation step 1/%d: %s", operation, m.state.TimingAnimationStepTotal, m.state.TimingAnimationCurrentStep)
	return nil
}

func (m *Module) playReferenceTiming(payload map[string]string) error {
	if err := m.ensureTimingPlaybackSteps(); err != nil {
		return err
	}
	interval, err := timingPlaybackInterval(payload, m.state.TimingPlaybackIntervalMS)
	if err != nil {
		return err
	}
	m.state.TimingPlaybackIntervalMS = interval
	m.state.TimingPlaybackState = "playing"
	m.state.TimingPlaybackStatus = timingPlaybackStatus(m.state, "playing")
	return nil
}

func (m *Module) pauseReferenceTiming() error {
	if err := m.ensureTimingPlaybackSteps(); err != nil {
		return err
	}
	m.state.TimingPlaybackState = "paused"
	m.state.TimingPlaybackStatus = timingPlaybackStatus(m.state, "paused")
	return nil
}

func (m *Module) stepReferenceTiming() error {
	if err := m.ensureTimingPlaybackSteps(); err != nil {
		return err
	}
	if m.state.TimingAnimationStepIndex < m.state.TimingAnimationStepTotal {
		m.state.TimingAnimationStepIndex++
	}
	m.state.TimingAnimationCurrentStep = m.state.TimingAnimationSteps[m.state.TimingAnimationStepIndex-1]
	if m.state.TimingAnimationStepIndex >= m.state.TimingAnimationStepTotal {
		m.state.TimingPlaybackState = "completed"
		m.state.TimingPlaybackStatus = timingPlaybackStatus(m.state, "completed")
		return nil
	}
	if m.state.TimingPlaybackState == "" || m.state.TimingPlaybackState == "stopped" || m.state.TimingPlaybackState == "completed" {
		m.state.TimingPlaybackState = "paused"
	}
	m.state.TimingPlaybackStatus = timingPlaybackStatus(m.state, m.state.TimingPlaybackState)
	return nil
}

func (m *Module) resetReferenceTiming() error {
	if err := m.ensureTimingPlaybackSteps(); err != nil {
		return err
	}
	m.state.TimingAnimationStepIndex = 1
	m.state.TimingAnimationCurrentStep = m.state.TimingAnimationSteps[0]
	m.state.TimingPlaybackState = "stopped"
	m.state.TimingPlaybackStatus = timingPlaybackStatus(m.state, "reset")
	return nil
}

func (m *Module) ensureTimingPlaybackSteps() error {
	m.computeReferenceTiming()
	operation := m.state.TimingActiveOp
	if operation == "" {
		operation = strings.ToUpper(m.state.OperationMode)
	}
	steps := timingAnimationSteps(operation)
	if len(steps) == 0 {
		return fmt.Errorf("circuits: unsupported timing playback operation %q", operation)
	}
	if m.state.TimingAnimationOperation != operation || len(m.state.TimingAnimationSteps) == 0 {
		m.state.TimingAnimationOperation = operation
		m.state.TimingAnimationSteps = append([]string(nil), steps...)
		m.state.TimingAnimationStepIndex = 1
		m.state.TimingAnimationStepTotal = len(steps)
		m.state.TimingAnimationCurrentStep = steps[0]
		return nil
	}
	m.state.TimingAnimationSteps = append([]string(nil), steps...)
	m.state.TimingAnimationStepTotal = len(steps)
	if m.state.TimingAnimationStepIndex < 1 {
		m.state.TimingAnimationStepIndex = 1
	}
	if m.state.TimingAnimationStepIndex > len(steps) {
		m.state.TimingAnimationStepIndex = len(steps)
	}
	m.state.TimingAnimationCurrentStep = steps[m.state.TimingAnimationStepIndex-1]
	return nil
}

func timingPlaybackInterval(payload map[string]string, current int) (int, error) {
	interval := current
	if interval <= 0 {
		interval = DefaultTimingPlaybackIntervalMS
	}
	if value := strings.TrimSpace(payload["interval_ms"]); value != "" {
		parsed, err := parseInt(value, "timing playback interval")
		if err != nil {
			return 0, err
		}
		if parsed <= 0 {
			return 0, fmt.Errorf("circuits: invalid timing playback interval %d", parsed)
		}
		interval = parsed
	}
	return interval, nil
}

func timingPlaybackStatus(state CircuitsState, status string) string {
	operation := state.TimingAnimationOperation
	stepIndex := state.TimingAnimationStepIndex
	stepTotal := state.TimingAnimationStepTotal
	currentStep := state.TimingAnimationCurrentStep
	interval := state.TimingPlaybackIntervalMS
	if interval <= 0 {
		interval = DefaultTimingPlaybackIntervalMS
	}
	switch status {
	case "playing":
		return fmt.Sprintf("playing %s timing playback step %d/%d every %dms: %s", operation, stepIndex, stepTotal, interval, currentStep)
	case "paused":
		return fmt.Sprintf("paused %s timing playback at step %d/%d: %s", operation, stepIndex, stepTotal, currentStep)
	case "completed":
		return fmt.Sprintf("completed %s timing playback at step %d/%d: %s", operation, stepIndex, stepTotal, currentStep)
	case "reset":
		return fmt.Sprintf("reset %s timing playback to step %d/%d: %s", operation, stepIndex, stepTotal, currentStep)
	default:
		return "not playing"
	}
}

func timingAnimationSteps(operation string) []string {
	switch strings.ToUpper(operation) {
	case "WRITE":
		return []string{
			"Phase 1: DAC settle (0-10ns)...",
			"Phase 2: Charge pump rise (10-98ns)...",
			"Phase 3: V_PROG write pulse (98-198ns)...",
			"Phase 4: Array settle (198-203ns)...",
			"Phase 5: DONE asserted (203ns)...",
			"Write complete: Total 203ns",
		}
	case "COMPUTE":
		return []string{
			"Phase 1: INPUT_VALID asserted (0ns)...",
			"Phase 2: DAC_ALL converts inputs (0-10ns)...",
			"Phase 3: ARRAY_SETTLE (10-15ns)...",
			"Phase 4: TIA+ADC digitizes summed currents (15-76ns)...",
			"Phase 5: OUTPUT_VALID - MVM result ready (76ns)...",
			"Compute complete: Total 76ns for full MVM",
		}
	default:
		return []string{
			"Phase 1: DAC settle (0-10ns)...",
			"Phase 2: Array settle (10-15ns)...",
			"Phase 3: TIA settle (15-26ns)...",
			"Phase 4: ADC convert (26-76ns)...",
			"Phase 5: DATA_OUT valid (76ns)...",
			"Read complete: Total 76ns",
		}
	}
}

func (m *Module) latestOperationLogIsCompute() bool {
	if len(m.state.OperationLog) == 0 {
		return false
	}
	latest := m.state.OperationLog[len(m.state.OperationLog)-1]
	return latest.Kind == "operation" && strings.HasPrefix(latest.Message, "COMPUTE ")
}

func (m *Module) buildComputeRunLog() *ComputeRunLog {
	return newComputeRunWorkflow(m.state).buildLog()
}

func deterministicComputeInputVector(cols int) []float64 {
	input := make([]float64, cols)
	for c := 0; c < cols; c++ {
		input[c] = mathutil.LerpByIndex(c, cols, 0.2, 0.5)
	}
	return input
}

func conductanceForLevelUS(level, quantLevels int) float64 {
	if quantLevels <= 1 {
		quantLevels = DefaultQuantLevels
	}
	return mathutil.LerpByIndex(level, quantLevels, 1.0, 100.0)
}

func cloneComputeRunLog(src *ComputeRunLog) *ComputeRunLog {
	if src == nil {
		return nil
	}
	dst := *src
	dst.InputVector = append([]float64(nil), src.InputVector...)
	dst.Weights = cloneIntMatrix(src.Weights)
	dst.Conductances = cloneFloatMatrix(src.Conductances)
	dst.RowResults = make([]ComputeRowResult, len(src.RowResults))
	for i, row := range src.RowResults {
		dst.RowResults[i] = row
		dst.RowResults[i].CellDetail = append([]ComputeCellContribution(nil), row.CellDetail...)
	}
	return &dst
}

func cloneIntMatrix(src [][]int) [][]int {
	dst := make([][]int, len(src))
	for i := range src {
		dst[i] = append([]int(nil), src[i]...)
	}
	return dst
}

func cloneFloatMatrix(src [][]float64) [][]float64 {
	dst := make([][]float64, len(src))
	for i := range src {
		dst[i] = append([]float64(nil), src[i]...)
	}
	return dst
}

func (m *Module) clampSelectedCell() {
	if m.state.SelectedRow >= m.state.Rows {
		m.state.SelectedRow = m.state.Rows - 1
	}
	if m.state.SelectedCol >= m.state.Cols {
		m.state.SelectedCol = m.state.Cols - 1
	}
	if m.state.SelectedRow < 0 {
		m.state.SelectedRow = 0
	}
	if m.state.SelectedCol < 0 {
		m.state.SelectedCol = 0
	}
}

func (m *Module) computeHalfSelectStress() {
	m.state = newHalfSelectStressWorkflow(m.state).compute()
}

func parsePayloadInt(payload map[string]string, key string) (int, error) {
	n, err := viewmodel.PayloadInt(payload, key)
	if err != nil {
		return 0, fmt.Errorf("circuits: %w", err)
	}
	return n, nil
}

func parseInt(value, label string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("circuits: invalid %s %q: %w", label, value, err)
	}
	return n, nil
}

func validArraySize(size int) bool {
	for _, valid := range ValidArraySizes {
		if size == valid {
			return true
		}
	}
	return false
}

func parseLoggerVerbosity(value string) (int, string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "off", "none":
		return 0, "off", nil
	case "1", "info":
		return 1, "info", nil
	case "2", "debug":
		return 2, "debug", nil
	case "3", "trace", "all":
		return 3, "trace", nil
	default:
		return 0, "", fmt.Errorf("circuits: invalid logger verbosity %q", value)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) DesignState() design.ModuleDesignState {
	return design.ModuleDesignState{ADCResolution: m.state.ADCResolution, DACResolution: m.state.DACResolution}
}

func (m *Module) runISPPSimulation() {
	m.state = newISPPSimulationWorkflow(m.state).compute()
	m.computePVTCorners()
}

func (m *Module) computePVTCorners() {
	m.state = newPVTCornersWorkflow(m.state).compute()
}

func (m *Module) computeReferenceSpecs() {
	m.state = newReferenceSpecWorkflow(m.state).compute()
}

func referenceSpecCompliance(dacCodes, adcCodes, quantLevels int) string {
	if dacCodes < quantLevels {
		return fmt.Sprintf("CHECK: DAC %d codes < %d levels", dacCodes, quantLevels)
	}
	if adcCodes < quantLevels {
		return fmt.Sprintf("CHECK: ADC %d codes < %d levels", adcCodes, quantLevels)
	}
	return fmt.Sprintf("OK: DAC/ADC cover %d levels", quantLevels)
}

func (m *Module) computeReferenceTiming() {
	m.state = newReferenceTimingWorkflow(m.state).compute()
}

func referenceTimingWaveforms() []ReferenceTimingWaveform {
	return []ReferenceTimingWaveform{
		{
			Operation: "WRITE",
			TotalNS:   203,
			Signals: []ReferenceTimingSignal{
				{Name: "CLK", HighWindows: timingWindows(0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95)},
				{Name: "ROW_SEL", HighWindows: timingWindows(10, 80)},
				{Name: "COL_SEL", HighWindows: timingWindows(10, 80)},
				{Name: "DAC_EN", HighWindows: timingWindows(15, 75)},
				{Name: "V_PROG", HighWindows: timingWindows(20, 70)},
				{Name: "DONE", HighWindows: timingWindows(85, 95)},
			},
			TimeMarkers: []ReferenceTimingTimeMarker{
				{Percent: 0, Label: "0ns", TimeNS: 0},
				{Percent: 25, Label: "51ns", TimeNS: 51},
				{Percent: 50, Label: "102ns", TimeNS: 102},
				{Percent: 75, Label: "152ns", TimeNS: 152},
				{Percent: 100, Label: "203ns", TimeNS: 203},
			},
			PhaseMarkers: []ReferenceTimingPhaseMarker{
				{Label: "DAC", StartPct: 0, EndPct: 5, DurationNS: 10},
				{Label: "Pump", StartPct: 5, EndPct: 48, DurationNS: 88},
				{Label: "Pulse", StartPct: 48, EndPct: 98, DurationNS: 100},
				{Label: "Array", StartPct: 98, EndPct: 100, DurationNS: 5},
			},
		},
		{
			Operation: "READ",
			TotalNS:   76,
			Signals: []ReferenceTimingSignal{
				{Name: "CLK", HighWindows: timingWindows(0, 10, 20, 30, 40, 50, 60, 70, 80, 90)},
				{Name: "V_READ", HighWindows: timingWindows(10, 70)},
				{Name: "I_SENSE", HighWindows: timingWindows(15, 75)},
				{Name: "ADC_EN", HighWindows: timingWindows(40, 70)},
				{Name: "DATA_OUT", HighWindows: timingWindows(75, 100)},
			},
			TimeMarkers: []ReferenceTimingTimeMarker{
				{Percent: 0, Label: "0ns", TimeNS: 0},
				{Percent: 25, Label: "19ns", TimeNS: 19},
				{Percent: 50, Label: "38ns", TimeNS: 38},
				{Percent: 75, Label: "57ns", TimeNS: 57},
				{Percent: 100, Label: "76ns", TimeNS: 76},
			},
			PhaseMarkers: []ReferenceTimingPhaseMarker{
				{Label: "DAC", StartPct: 0, EndPct: 13, DurationNS: 10},
				{Label: "Array", StartPct: 13, EndPct: 20, DurationNS: 5},
				{Label: "TIA", StartPct: 20, EndPct: 34, DurationNS: 11},
				{Label: "ADC", StartPct: 34, EndPct: 100, DurationNS: 50},
			},
		},
		{
			Operation: "COMPUTE",
			TotalNS:   76,
			Signals: []ReferenceTimingSignal{
				{Name: "CLK", HighWindows: timingWindows(0, 8, 16, 24, 32, 40, 48, 56, 64, 72, 80, 88)},
				{Name: "INPUT_VALID", HighWindows: timingWindows(5, 85)},
				{Name: "DAC_ALL", HighWindows: timingWindows(10, 35)},
				{Name: "ARRAY_SETTLE", HighWindows: timingWindows(35, 60)},
				{Name: "ADC_ALL", HighWindows: timingWindows(55, 90)},
				{Name: "OUTPUT_VALID", HighWindows: timingWindows(90, 100)},
			},
			TimeMarkers: []ReferenceTimingTimeMarker{
				{Percent: 0, Label: "0ns", TimeNS: 0},
				{Percent: 25, Label: "19ns", TimeNS: 19},
				{Percent: 50, Label: "38ns", TimeNS: 38},
				{Percent: 75, Label: "57ns", TimeNS: 57},
				{Percent: 100, Label: "76ns", TimeNS: 76},
			},
			PhaseMarkers: []ReferenceTimingPhaseMarker{
				{Label: "DAC", StartPct: 10, EndPct: 35, DurationNS: 10},
				{Label: "Array", StartPct: 35, EndPct: 60, DurationNS: 5},
				{Label: "TIA+ADC", StartPct: 55, EndPct: 90, DurationNS: 61},
			},
		},
	}
}

func timingWindows(pcts ...int) []ReferenceTimingWindow {
	windows := make([]ReferenceTimingWindow, 0, len(pcts)/2)
	for i := 0; i+1 < len(pcts); i += 2 {
		windows = append(windows, ReferenceTimingWindow{StartPct: pcts[i], EndPct: pcts[i+1]})
	}
	return windows
}

func pvtTemperatureSweepStatus(mat *physics.HZOMaterial) string {
	tempsC := []float64{-40, 25, 85, 125}
	var prevEc, prevPr float64
	for i, tempC := range tempsC {
		tempK := tempC + 273.15
		ec := mat.CoerciveFieldAtTemp(tempK)
		pr := mat.PolarizationAtTemp(tempK)
		if ec <= 0 || pr <= 0 || math.IsNaN(ec) || math.IsNaN(pr) || math.IsInf(ec, 0) || math.IsInf(pr, 0) {
			return "check -40/25/85/125 C"
		}
		if i > 0 && (ec > prevEc || pr > prevPr) {
			return "check -40/25/85/125 C"
		}
		prevEc, prevPr = ec, pr
	}
	return "pass -40/25/85/125 C"
}

func pvtProcessYield(mat *physics.HZOMaterial) (float64, int, int) {
	const (
		samples      = 20
		sigmaEcFrac  = 0.03
		sigmaPrFrac  = 0.03
		accuracySpec = 0.90
	)
	rng := rand.New(rand.NewSource(42))
	pass := 0
	for i := 0; i < samples; i++ {
		ec := mat.Ec * (1 + sigmaEcFrac*rng.NormFloat64())
		pr := mat.Pr * (1 + sigmaPrFrac*rng.NormFloat64())
		if ec <= 0 || pr <= 0 {
			continue
		}
		normGain := (ec / mat.Ec) * (pr / mat.Pr)
		accuracy := 1 - math.Abs(normGain-1)
		if accuracy >= accuracySpec {
			pass++
		}
	}
	return float64(pass) / float64(samples), pass, samples
}

func pvtNoiseLimitedENOBCeiling(tiaGain float64) (float64, int) {
	const (
		vRange = 1.8
		tempK  = 300.0
		bwHz   = 10e6
	)
	if tiaGain <= 0 {
		tiaGain = 10e3
	}
	thermalVar := math.Pow(peripherals.ThermalNoiseRMS(tempK, tiaGain, bwHz), 2)
	signalRMS := vRange / (2 * math.Sqrt2)

	bestENOB := 0.0
	bestBits := 0
	for bits := 6; bits <= 16; bits++ {
		totalVar := thermalVar + peripherals.QuantizationNoiseVariance(vRange, bits)
		enob := (peripherals.SNRDB(signalRMS, math.Sqrt(totalVar)) - 1.76) / 6.02
		if enob > bestENOB {
			bestENOB = enob
			bestBits = bits
		}
	}
	return bestENOB, bestBits
}
