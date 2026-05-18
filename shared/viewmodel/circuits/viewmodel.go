package circuits

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"fecim-lattice-tools/shared/peripherals"
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
		m.state.LastOperationStatus = fmt.Sprintf("READ cell [%d,%d] through %s", m.state.SelectedRow, m.state.SelectedCol, m.state.Architecture)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
		return nil
	case ActionRunWrite:
		m.state.OperationMode = OperationWrite
		m.runISPPSimulation()
		m.state.LastOperationStatus = fmt.Sprintf("WRITE level %d to cell [%d,%d] using %s", m.state.WriteTargetLevel, m.state.SelectedRow, m.state.SelectedCol, m.state.ISPPEngine)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
		return nil
	case ActionRunCompute:
		m.state.OperationMode = OperationCompute
		m.state.LastOperationStatus = fmt.Sprintf("COMPUTE on %dx%d %s array", m.state.Rows, m.state.Cols, m.state.Architecture)
		m.computeHalfSelectStress()
		m.computeReferenceTiming()
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
	m.computeHalfSelectStress()
	m.computeReferenceSpecs()
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
	m.computeHalfSelectStress()
	m.computeReferenceTiming()
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
	m.state.LastOperationStatus = fmt.Sprintf("Selected cell [%d,%d]", row, col)
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
	m.state.LastOperationStatus = fmt.Sprintf("Write target set to level %d", level)
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
	m.computeReferenceSpecs()
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
	m.computePVTCorners()
	m.computeReferenceSpecs()
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

func (m *Module) computeHalfSelectStress() {
	m.state.HalfSelectState = HalfSelectStateInactive
	m.state.HalfSelectCells = 0
	m.state.DisturbVoltage = 0
	m.state.StressPerPulse = 0
	m.state.StressCyclesToLevel = 0

	if m.state.OperationMode != OperationWrite {
		return
	}

	switch m.state.Architecture {
	case ArchitecturePassive:
		m.state.HalfSelectState = HalfSelectStateColumnWriteActive
		m.state.HalfSelectCells = maxInt(0, m.state.Rows-1)
		m.state.DisturbVoltage = DefaultDisturbVoltage
		m.state.StressPerPulse = PassiveStressPerPulse
	case Architecture1T1R:
		m.state.HalfSelectState = HalfSelectStateAttenuated
		m.state.HalfSelectCells = maxInt(0, m.state.Rows-1)
		m.state.DisturbVoltage = DefaultDisturbVoltage / OneTOneRStressAttenuation
		m.state.StressPerPulse = PassiveStressPerPulse / OneTOneRStressAttenuation
	case Architecture2T1R:
		m.state.HalfSelectState = HalfSelectStateIsolated
		return
	default:
		return
	}
	if m.state.StressPerPulse > 0 {
		m.state.StressCyclesToLevel = int(math.Ceil(1.0 / m.state.StressPerPulse))
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
	mat := physics.DefaultHZO()
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
	m.state.PVTTemperatureSweep = pvtTemperatureSweepStatus(mat)
	m.state.PVTProcessYield, m.state.PVTPassSamples, m.state.PVTSamples = pvtProcessYield(mat)
	m.state.PVTENOBNoiseCeiling, m.state.PVTENOBCeilingBits = pvtNoiseLimitedENOBCeiling(m.state.TIAGain)

	_ = lsb
	_ = vref
}

func (m *Module) computeReferenceSpecs() {
	quantLevels := m.state.QuantLevels
	if quantLevels <= 0 {
		quantLevels = DefaultQuantLevels
	}
	rows := maxInt(1, m.state.Rows)
	cols := maxInt(1, m.state.Cols)
	cells := rows * cols
	dacCodes := 1 << m.state.DACResolution
	adcCodes := 1 << m.state.ADCResolution

	const (
		arrayPowerMW       = 0.1
		controlPowerMW     = 0.5
		dacPowerPerColMW   = 0.1
		tiaPowerPerRowMW   = 0.05
		adcPowerPerRowMW   = 0.5
		referenceLatencyNS = 76.0
	)
	totalPowerMW := arrayPowerMW + controlPowerMW +
		dacPowerPerColMW*float64(cols) +
		tiaPowerPerRowMW*float64(rows) +
		adcPowerPerRowMW*float64(rows)
	throughputGOPS := float64(cells) / referenceLatencyNS
	efficiencyGOPSW := 0.0
	if totalPowerMW > 0 {
		efficiencyGOPSW = throughputGOPS * 1000 / totalPowerMW
	}

	m.state.SpecCells = cells
	m.state.SpecBitsPerCell = math.Log2(float64(quantLevels))
	m.state.SpecDACCount = cols
	m.state.SpecTIACount = rows
	m.state.SpecADCCount = rows
	m.state.SpecDACCodes = dacCodes
	m.state.SpecADCCodes = adcCodes
	m.state.SpecTotalPowerMW = totalPowerMW
	m.state.SpecLatencyNS = referenceLatencyNS
	m.state.SpecThroughputGOPS = throughputGOPS
	m.state.SpecEfficiencyGOPSW = efficiencyGOPSW
	m.state.SpecCompliance = referenceSpecCompliance(dacCodes, adcCodes, quantLevels)
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
	const (
		writeTotalNS   = 203
		readTotalNS    = 76
		computeTotalNS = 76
	)
	m.state.TimingWriteTotalNS = writeTotalNS
	m.state.TimingReadTotalNS = readTotalNS
	m.state.TimingComputeTotalNS = computeTotalNS

	switch m.state.OperationMode {
	case OperationWrite:
		m.state.TimingActiveOp = "WRITE"
		m.state.TimingActiveTotalNS = writeTotalNS
		m.state.TimingActivePhases = "DAC 10 / Pump 88 / Pulse 100 / Array 5 ns"
	case OperationCompute:
		m.state.TimingActiveOp = "COMPUTE"
		m.state.TimingActiveTotalNS = computeTotalNS
		m.state.TimingActivePhases = "DAC 10 / Array 5 / TIA+ADC 61 ns"
	default:
		m.state.TimingActiveOp = "READ"
		m.state.TimingActiveTotalNS = readTotalNS
		m.state.TimingActivePhases = "DAC 10 / Array 5 / TIA 11 / ADC 50 ns"
	}
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
