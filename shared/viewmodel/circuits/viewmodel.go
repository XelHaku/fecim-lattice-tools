package circuits

import (
	"fmt"
	"math"
	"strconv"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state CircuitsState }

func New() *Module {
	m := &Module{state: CircuitsState{
		Rows: 8, Cols: 8,
		OperationMode:    OperationRead,
		Architecture:     ArchitecturePassive,
		WriteTargetLevel: DefaultQuantLevels / 2,
		QuantLevels:      DefaultQuantLevels,
		CouplingTier:     CouplingTierA,
		ISPPEngine:       ISPPEngineLevel,
		ADCResolution:    5, DACResolution: 5, TIAGain: 1e4,
		ChargePumpStages: 4, SupplyVoltage: 1.8, ISPPEnabled: true,
	}}
	m.runISPPSimulation()
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
		m.state.LastOperationStatus = fmt.Sprintf("READ cell [%d,%d] through %s", m.state.SelectedRow, m.state.SelectedCol, m.state.Architecture)
		return nil
	case ActionRunWrite:
		m.state.OperationMode = OperationWrite
		m.runISPPSimulation()
		m.state.LastOperationStatus = fmt.Sprintf("WRITE level %d to cell [%d,%d] using %s", m.state.WriteTargetLevel, m.state.SelectedRow, m.state.SelectedCol, m.state.ISPPEngine)
		return nil
	case ActionRunCompute:
		m.state.OperationMode = OperationCompute
		m.state.LastOperationStatus = fmt.Sprintf("COMPUTE on %dx%d %s array", m.state.Rows, m.state.Cols, m.state.Architecture)
		return nil
	case ActionToggleISPP:
		m.state.ISPPEnabled = !m.state.ISPPEnabled
		m.state.LastOperationStatus = fmt.Sprintf("ISPP enabled: %v", m.state.ISPPEnabled)
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
	m.state.LastOperationStatus = fmt.Sprintf("Array resized to %dx%d", rows, cols)
	return nil
}

func (m *Module) setOperationMode(payload map[string]string) error {
	mode, ok := payload["mode"]
	if !ok {
		return fmt.Errorf("circuits: missing operation mode")
	}
	if !validString(mode, OperationRead, OperationWrite, OperationCompute) {
		return fmt.Errorf("circuits: invalid operation mode %q", mode)
	}
	m.state.OperationMode = mode
	m.state.LastOperationStatus = fmt.Sprintf("Operation mode set to %s", mode)
	return nil
}

func (m *Module) setArchitecture(payload map[string]string) error {
	architecture, ok := payload["architecture"]
	if !ok {
		return fmt.Errorf("circuits: missing architecture")
	}
	if !validString(architecture, ArchitecturePassive, Architecture1T1R, Architecture2T1R) {
		return fmt.Errorf("circuits: invalid architecture %q", architecture)
	}
	m.state.Architecture = architecture
	m.state.LastOperationStatus = fmt.Sprintf("Architecture set to %s", architecture)
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
	m.state.LastOperationStatus = fmt.Sprintf("Selected cell [%d,%d]", row, col)
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
	m.state.LastOperationStatus = fmt.Sprintf("Write target set to level %d", level)
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
	m.state.LastOperationStatus = fmt.Sprintf("DAC resolution set to %d bits", bits)
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
	m.state.LastOperationStatus = fmt.Sprintf("ADC resolution set to %d bits", bits)
	return nil
}

func (m *Module) setTIAGain(payload map[string]string) error {
	gainS, ok := payload["gain_ohm"]
	if !ok {
		return fmt.Errorf("circuits: missing TIA gain")
	}
	gain, err := strconv.ParseFloat(gainS, 64)
	if err != nil {
		return fmt.Errorf("circuits: invalid TIA gain %q: %w", gainS, err)
	}
	if gain <= 0 {
		return fmt.Errorf("circuits: TIA gain must be positive, got %.3g", gain)
	}
	m.state.TIAGain = gain
	m.state.LastOperationStatus = fmt.Sprintf("TIA gain set to %.0f ohm", gain)
	return nil
}

func (m *Module) setCouplingTier(payload map[string]string) error {
	tier, ok := payload["tier"]
	if !ok {
		return fmt.Errorf("circuits: missing coupling tier")
	}
	if !validString(tier, CouplingIdeal, CouplingTierA, CouplingTierB) {
		return fmt.Errorf("circuits: invalid coupling tier %q", tier)
	}
	m.state.CouplingTier = tier
	m.state.LastOperationStatus = fmt.Sprintf("Coupling tier set to %s", tier)
	return nil
}

func (m *Module) setISPPEngine(payload map[string]string) error {
	engine, ok := payload["engine"]
	if !ok {
		return fmt.Errorf("circuits: missing ISPP engine")
	}
	if !validString(engine, ISPPEngineLevel, ISPPEngineLK) {
		return fmt.Errorf("circuits: invalid ISPP engine %q", engine)
	}
	m.state.ISPPEngine = engine
	m.state.LastOperationStatus = fmt.Sprintf("ISPP engine set to %s", engine)
	return nil
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

func parsePayloadInt(payload map[string]string, key string) (int, error) {
	value, ok := payload[key]
	if !ok {
		return 0, fmt.Errorf("circuits: missing %s", key)
	}
	return parseInt(value, key)
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

func validString(value string, validValues ...string) bool {
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) runISPPSimulation() {
	mat := physics.DefaultHZO()
	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ctrl := physics.NewWriteController(solver, mat)

	numLevels := m.state.QuantLevels
	if numLevels <= 0 {
		numLevels = DefaultQuantLevels
		m.state.QuantLevels = numLevels
	}
	m.state.ISPPAttempts = make([]int, numLevels)
	m.state.ISPPConverged = make([]bool, numLevels)

	successCount := 0
	totalAttempts := 0
	for level := 0; level < numLevels; level++ {
		targetG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(level)/float64(numLevels-1)
		attempts, success, _ := ctrl.WriteTarget(targetG)
		m.state.ISPPAttempts[level] = attempts
		m.state.ISPPConverged[level] = success
		if success {
			successCount++
		}
		totalAttempts += attempts
	}

	m.state.ISPPTotalAttempts = totalAttempts
	m.state.ISPPConvergedCount = successCount
	if totalAttempts > 0 {
		m.state.ISPPAvgAttempts = float64(totalAttempts) / float64(numLevels)
	}
	m.state.ISPPExecuted = true

	m.computePVTCorners()
}

func (m *Module) computePVTCorners() {
	vref := m.state.SupplyVoltage
	bits := m.state.ADCResolution
	lsb := vref / float64(int(1)<<bits)

	enobForINL := func(inlLSB float64) float64 {
		return math.Max(float64(bits)-math.Log2(inlLSB+1.0), 1.0)
	}
	m.state.ENOBtt = enobForINL(0.5)
	m.state.ENOBff = enobForINL(0.5 * 0.80)
	m.state.ENOBss = enobForINL(0.5 * 1.25)
	m.state.ADCNoiseLSB = math.Sqrt(lsb * lsb / 12.0)
	m.state.SNRdB = 6.02*float64(bits) + 1.76

	_ = lsb
	_ = vref
}
