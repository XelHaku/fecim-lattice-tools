// Package layers provides neural network layer implementations for CIM deployment.
// This file implements analog error correction and mixed-signal co-simulation
// for CIM accelerator validation and deployment.
// Based on research: Science (high-precision programming), DAC (CIM-ECC),
// Nature Communications (adaptive ADC), ACM (analog-mixed signal simulation)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// ANALOG ERROR CORRECTION FOR CIM
// ECC, redundancy, calibration, and write-verify programming
// =============================================================================

// ECCType defines types of error correcting codes for CIM.
type ECCType int

const (
	ECCSECDED    ECCType = iota // Single Error Correct, Double Error Detect
	ECCDEC                      // Double Error Correct
	ECCTEC                      // Triple Error Correct
	ECCLDPC                     // Low-Density Parity-Check
	ECCAnalog                   // Analog error correction
	ECCRedundancy               // Redundancy-based
)

// CIMECCConfig configures error correction for CIM.
type CIMECCConfig struct {
	Type            ECCType
	DataBits        int
	ParityBits      int

	// Correction capability
	CorrectableErrors int
	DetectableErrors  int

	// Analog-specific
	RedundancyFactor  float64 // >1.0 for redundant cells
	VotingThreshold   int     // For triple modular redundancy

	// Calibration
	EnableCalibration bool
	CalibrationPeriod int     // Operations between calibration
}

// DefaultCIMECCConfig returns default ECC config.
func DefaultCIMECCConfig() *CIMECCConfig {
	return &CIMECCConfig{
		Type:              ECCSECDED,
		DataBits:          64,
		ParityBits:        8,
		CorrectableErrors: 1,
		DetectableErrors:  2,
		RedundancyFactor:  1.0,
		VotingThreshold:   2,
		EnableCalibration: true,
		CalibrationPeriod: 10000,
	}
}

// CIMErrorCorrector implements error correction for CIM.
type CIMErrorCorrector struct {
	Config *CIMECCConfig

	// Parity check matrix (for LDPC/linear codes)
	ParityMatrix [][]int

	// Syndrome lookup table
	SyndromeLUT map[int]int // Syndrome -> error position

	// Statistics
	TotalOperations   int64
	ErrorsDetected    int64
	ErrorsCorrected   int64
	UncorrectableErrs int64

	// Calibration state
	LastCalibration   int64
}

// NewCIMErrorCorrector creates a CIM error corrector.
func NewCIMErrorCorrector(config *CIMECCConfig) *CIMErrorCorrector {
	if config == nil {
		config = DefaultCIMECCConfig()
	}

	ec := &CIMErrorCorrector{
		Config:      config,
		SyndromeLUT: make(map[int]int),
	}

	// Build parity check matrix based on ECC type
	switch config.Type {
	case ECCSECDED:
		ec.buildSECDEDMatrix()
	case ECCLDPC:
		ec.buildLDPCMatrix()
	default:
		ec.buildSECDEDMatrix()
	}

	return ec
}

// buildSECDEDMatrix builds SECDED parity check matrix.
func (ec *CIMErrorCorrector) buildSECDEDMatrix() {
	n := ec.Config.DataBits + ec.Config.ParityBits
	k := ec.Config.DataBits

	// Simple Hamming-like construction
	ec.ParityMatrix = make([][]int, ec.Config.ParityBits)
	for i := range ec.ParityMatrix {
		ec.ParityMatrix[i] = make([]int, n)
		for j := 0; j < n; j++ {
			// Set parity bits using Hamming pattern
			if (j+1)&(1<<i) != 0 {
				ec.ParityMatrix[i][j] = 1
			}
		}
	}

	// Build syndrome lookup
	for pos := 0; pos < n; pos++ {
		syndrome := 0
		for i := 0; i < ec.Config.ParityBits; i++ {
			if ec.ParityMatrix[i][pos] == 1 {
				syndrome |= 1 << i
			}
		}
		ec.SyndromeLUT[syndrome] = pos
	}

	_ = k // k used for data portion
}

// buildLDPCMatrix builds sparse LDPC parity check matrix.
func (ec *CIMErrorCorrector) buildLDPCMatrix() {
	n := ec.Config.DataBits + ec.Config.ParityBits
	m := ec.Config.ParityBits

	// Regular LDPC with column weight 3
	colWeight := 3
	ec.ParityMatrix = make([][]int, m)
	for i := range ec.ParityMatrix {
		ec.ParityMatrix[i] = make([]int, n)
	}

	// Distribute 1s with column weight constraint
	for j := 0; j < n; j++ {
		positions := make([]int, 0, colWeight)
		for len(positions) < colWeight {
			pos := rand.Intn(m)
			found := false
			for _, p := range positions {
				if p == pos {
					found = true
					break
				}
			}
			if !found {
				positions = append(positions, pos)
			}
		}
		for _, pos := range positions {
			ec.ParityMatrix[pos][j] = 1
		}
	}
}

// Encode adds parity bits to data.
func (ec *CIMErrorCorrector) Encode(data []int) []int {
	n := ec.Config.DataBits + ec.Config.ParityBits
	codeword := make([]int, n)

	// Copy data bits
	copy(codeword, data)

	// Compute parity bits
	for i := 0; i < ec.Config.ParityBits; i++ {
		parity := 0
		for j := 0; j < ec.Config.DataBits; j++ {
			if ec.ParityMatrix[i][j] == 1 {
				parity ^= data[j]
			}
		}
		codeword[ec.Config.DataBits+i] = parity
	}

	return codeword
}

// Decode detects and corrects errors.
func (ec *CIMErrorCorrector) Decode(received []int) ([]int, *ECCResult) {
	ec.TotalOperations++

	// Compute syndrome
	syndrome := 0
	for i := 0; i < ec.Config.ParityBits; i++ {
		bit := 0
		for j := 0; j < len(received) && j < len(ec.ParityMatrix[i]); j++ {
			if ec.ParityMatrix[i][j] == 1 {
				bit ^= received[j]
			}
		}
		if bit == 1 {
			syndrome |= 1 << i
		}
	}

	result := &ECCResult{
		Syndrome:    syndrome,
		ErrorsFound: 0,
	}

	corrected := make([]int, len(received))
	copy(corrected, received)

	if syndrome != 0 {
		ec.ErrorsDetected++
		result.ErrorsFound = 1

		// Try to correct
		if errorPos, ok := ec.SyndromeLUT[syndrome]; ok && errorPos < len(corrected) {
			corrected[errorPos] ^= 1
			ec.ErrorsCorrected++
			result.Corrected = true
			result.CorrectedPositions = []int{errorPos}
		} else {
			// Uncorrectable error
			ec.UncorrectableErrs++
			result.Uncorrectable = true
		}
	}

	// Extract data bits
	data := corrected[:ec.Config.DataBits]

	return data, result
}

// ECCResult contains ECC decoding result.
type ECCResult struct {
	Syndrome           int
	ErrorsFound        int
	Corrected          bool
	CorrectedPositions []int
	Uncorrectable      bool
}

// GetBER returns bit error rate.
func (ec *CIMErrorCorrector) GetBER() float64 {
	if ec.TotalOperations == 0 {
		return 0
	}
	return float64(ec.ErrorsDetected) / float64(ec.TotalOperations*int64(ec.Config.DataBits))
}

// =============================================================================
// WRITE-VERIFY PROGRAMMING
// High-precision memristor conductance programming
// =============================================================================

// WriteVerifyConfig configures write-verify programming.
type WriteVerifyConfig struct {
	// Target precision
	TargetPrecisionBits int
	MaxIterations       int
	TolerancePercent    float64

	// Pulse parameters
	InitialPulseV       float64
	PulseStepV          float64
	PulseWidthNs        float64

	// Verification
	VerifyDelayNs       float64
	ReadVoltageV        float64

	// Adaptive programming
	UseAdaptive         bool
	LearningRate        float64
}

// DefaultWriteVerifyConfig returns default write-verify config.
func DefaultWriteVerifyConfig() *WriteVerifyConfig {
	return &WriteVerifyConfig{
		TargetPrecisionBits: 6,
		MaxIterations:       30,
		TolerancePercent:    1.0,
		InitialPulseV:       1.0,
		PulseStepV:          0.1,
		PulseWidthNs:        100.0,
		VerifyDelayNs:       50.0,
		ReadVoltageV:        0.2,
		UseAdaptive:         true,
		LearningRate:        0.1,
	}
}

// WriteVerifyProgrammer implements write-verify programming.
type WriteVerifyProgrammer struct {
	Config *WriteVerifyConfig

	// State
	CurrentConductances [][]float64
	TargetConductances  [][]float64
	ProgrammingErrors   [][]float64

	// Statistics
	TotalProgramOps     int64
	TotalVerifyOps      int64
	AverageIterations   float64
	SuccessRate         float64

	// History for adaptive programming
	PulseHistory        []*ProgrammingPulse
}

// ProgrammingPulse records a programming pulse.
type ProgrammingPulse struct {
	Row, Col        int
	Voltage         float64
	Width           float64
	ConductanceBefore float64
	ConductanceAfter  float64
	DeltaConductance  float64
}

// NewWriteVerifyProgrammer creates a write-verify programmer.
func NewWriteVerifyProgrammer(rows, cols int, config *WriteVerifyConfig) *WriteVerifyProgrammer {
	if config == nil {
		config = DefaultWriteVerifyConfig()
	}

	p := &WriteVerifyProgrammer{
		Config:              config,
		CurrentConductances: make([][]float64, rows),
		TargetConductances:  make([][]float64, rows),
		ProgrammingErrors:   make([][]float64, rows),
		PulseHistory:        make([]*ProgrammingPulse, 0),
	}

	for i := 0; i < rows; i++ {
		p.CurrentConductances[i] = make([]float64, cols)
		p.TargetConductances[i] = make([]float64, cols)
		p.ProgrammingErrors[i] = make([]float64, cols)
	}

	return p
}

// ProgramCell programs a single cell to target conductance.
func (p *WriteVerifyProgrammer) ProgramCell(row, col int, target float64) *ProgramResult {
	p.TargetConductances[row][col] = target

	tolerance := target * p.Config.TolerancePercent / 100.0
	current := p.CurrentConductances[row][col]

	iterations := 0
	pulses := make([]*ProgrammingPulse, 0)

	for iterations < p.Config.MaxIterations {
		iterations++
		p.TotalVerifyOps++

		// Check if within tolerance
		error := target - current
		if math.Abs(error) <= tolerance {
			break
		}

		// Calculate programming pulse
		var voltage float64
		if p.Config.UseAdaptive {
			// Adaptive: predict voltage from error
			voltage = p.predictPulse(error, current, target)
		} else {
			// Fixed step
			if error > 0 {
				voltage = p.Config.InitialPulseV + float64(iterations)*p.Config.PulseStepV
			} else {
				voltage = -(p.Config.InitialPulseV + float64(iterations)*p.Config.PulseStepV)
			}
		}

		// Apply pulse (simulated)
		beforeG := current
		deltaG := p.simulatePulse(current, voltage)
		current += deltaG

		// Clamp to valid range
		if current < 0 {
			current = 0
		}
		if current > 1 {
			current = 1
		}

		pulse := &ProgrammingPulse{
			Row:               row,
			Col:               col,
			Voltage:           voltage,
			Width:             p.Config.PulseWidthNs,
			ConductanceBefore: beforeG,
			ConductanceAfter:  current,
			DeltaConductance:  deltaG,
		}
		pulses = append(pulses, pulse)
		p.PulseHistory = append(p.PulseHistory, pulse)
		p.TotalProgramOps++
	}

	// Update state
	p.CurrentConductances[row][col] = current
	p.ProgrammingErrors[row][col] = target - current

	// Update statistics
	successCount := 0
	for _, row := range p.ProgrammingErrors {
		for _, err := range row {
			if math.Abs(err) <= tolerance {
				successCount++
			}
		}
	}
	totalCells := len(p.CurrentConductances) * len(p.CurrentConductances[0])
	p.SuccessRate = float64(successCount) / float64(totalCells) * 100

	// Update average iterations
	if p.TotalProgramOps > 0 {
		p.AverageIterations = float64(len(p.PulseHistory)) / float64(p.TotalProgramOps) * float64(iterations)
	}

	return &ProgramResult{
		Row:           row,
		Col:           col,
		TargetG:       target,
		ActualG:       current,
		Error:         target - current,
		Iterations:    iterations,
		Success:       math.Abs(target-current) <= tolerance,
		Pulses:        pulses,
	}
}

// predictPulse predicts optimal pulse voltage using history.
func (p *WriteVerifyProgrammer) predictPulse(error, current, target float64) float64 {
	// Simple linear prediction
	// In practice, this could use a neural network
	baseVoltage := p.Config.InitialPulseV

	// Scale voltage based on error magnitude
	errorMag := math.Abs(error)
	scaleFactor := math.Min(errorMag*10, 2.0)

	voltage := baseVoltage * scaleFactor
	if error < 0 {
		voltage = -voltage
	}

	return voltage
}

// simulatePulse simulates conductance change from pulse.
func (p *WriteVerifyProgrammer) simulatePulse(current, voltage float64) float64 {
	// Simplified memristor model
	// Delta G proportional to voltage with some nonlinearity
	sign := 1.0
	if voltage < 0 {
		sign = -1.0
	}

	absV := math.Abs(voltage)

	// Threshold behavior
	threshold := 0.5
	if absV < threshold {
		return 0
	}

	// Nonlinear switching
	deltaG := sign * 0.05 * math.Pow(absV-threshold, 1.5)

	// Add some randomness (device variation)
	deltaG *= (1 + 0.1*rand.NormFloat64())

	return deltaG
}

// ProgramArray programs entire array to targets.
func (p *WriteVerifyProgrammer) ProgramArray(targets [][]float64) *ArrayProgramResult {
	rows := len(targets)
	cols := len(targets[0])

	results := make([][]*ProgramResult, rows)
	totalIterations := 0
	successCount := 0

	for i := 0; i < rows; i++ {
		results[i] = make([]*ProgramResult, cols)
		for j := 0; j < cols; j++ {
			result := p.ProgramCell(i, j, targets[i][j])
			results[i][j] = result
			totalIterations += result.Iterations
			if result.Success {
				successCount++
			}
		}
	}

	// Compute overall error
	mse := 0.0
	maxError := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			err := math.Abs(p.ProgrammingErrors[i][j])
			mse += err * err
			if err > maxError {
				maxError = err
			}
		}
	}
	mse /= float64(rows * cols)

	return &ArrayProgramResult{
		Results:          results,
		TotalIterations:  totalIterations,
		AverageIterations: float64(totalIterations) / float64(rows*cols),
		SuccessRate:      float64(successCount) / float64(rows*cols) * 100,
		MSE:              mse,
		MaxError:         maxError,
	}
}

// ProgramResult contains cell programming result.
type ProgramResult struct {
	Row, Col    int
	TargetG     float64
	ActualG     float64
	Error       float64
	Iterations  int
	Success     bool
	Pulses      []*ProgrammingPulse
}

// ArrayProgramResult contains array programming result.
type ArrayProgramResult struct {
	Results           [][]*ProgramResult
	TotalIterations   int
	AverageIterations float64
	SuccessRate       float64
	MSE               float64
	MaxError          float64
}

// =============================================================================
// MIXED-SIGNAL CO-SIMULATION
// Behavioral models for analog-digital CIM simulation
// =============================================================================

// ComponentType defines mixed-signal component types.
type ComponentType int

const (
	ComponentDAC      ComponentType = iota // Digital-to-Analog Converter
	ComponentADC                           // Analog-to-Digital Converter
	ComponentOpAmp                         // Operational Amplifier
	ComponentTIA                           // Transimpedance Amplifier
	ComponentCrossbar                      // Memristor Crossbar
	ComponentMux                           // Analog Multiplexer
)

// MixedSignalConfig configures mixed-signal simulation.
type MixedSignalConfig struct {
	// Simulation parameters
	TimeStepPs        float64 // Picoseconds
	MaxSimTimeNs      float64

	// ADC model
	ADCBits           int
	ADCSamplingRateHz float64
	ADCSNR            float64 // dB
	ADCLatencyCycles  int

	// DAC model
	DACBits           int
	DACSettlingTimeNs float64
	DACTHD            float64 // Total harmonic distortion

	// Crossbar model
	LineResistanceOhm float64
	ParasiticCapPF    float64

	// Noise
	ThermalNoiseK     float64 // Boltzmann constant factor
	FlickerNoiseAlpha float64
}

// DefaultMixedSignalConfig returns default mixed-signal config.
func DefaultMixedSignalConfig() *MixedSignalConfig {
	return &MixedSignalConfig{
		TimeStepPs:        100.0,
		MaxSimTimeNs:      1000.0,
		ADCBits:           6,
		ADCSamplingRateHz: 1e9,
		ADCSNR:            40.0,
		ADCLatencyCycles:  4,
		DACBits:           8,
		DACSettlingTimeNs: 10.0,
		DACTHD:            -60.0,
		LineResistanceOhm: 10.0,
		ParasiticCapPF:    1.0,
		ThermalNoiseK:     1.38e-23,
		FlickerNoiseAlpha: 1.0,
	}
}

// DACModel implements DAC behavioral model.
type DACModel struct {
	Config     *MixedSignalConfig
	NumBits    int
	VrefPos    float64
	VrefNeg    float64

	// Non-idealities
	DNL        []float64 // Differential nonlinearity
	INL        []float64 // Integral nonlinearity
	Offset     float64
	Gain       float64

	// State
	CurrentCode   int
	CurrentVoltage float64
	SettlingState float64
}

// NewDACModel creates a DAC model.
func NewDACModel(bits int, vref float64, config *MixedSignalConfig) *DACModel {
	if config == nil {
		config = DefaultMixedSignalConfig()
	}

	levels := 1 << bits

	dac := &DACModel{
		Config:  config,
		NumBits: bits,
		VrefPos: vref,
		VrefNeg: 0,
		DNL:     make([]float64, levels),
		INL:     make([]float64, levels),
		Offset:  rand.NormFloat64() * 0.001 * vref, // Small offset error
		Gain:    1.0 + rand.NormFloat64()*0.001,    // Small gain error
	}

	// Generate DNL/INL errors
	cumINL := 0.0
	for i := 0; i < levels; i++ {
		dac.DNL[i] = rand.NormFloat64() * 0.1 // 0.1 LSB typical
		cumINL += dac.DNL[i]
		dac.INL[i] = cumINL
	}

	return dac
}

// Convert converts digital code to analog voltage.
func (dac *DACModel) Convert(code int, timeNs float64) float64 {
	if code < 0 {
		code = 0
	}
	maxCode := (1 << dac.NumBits) - 1
	if code > maxCode {
		code = maxCode
	}

	// Ideal conversion
	lsb := (dac.VrefPos - dac.VrefNeg) / float64(1<<dac.NumBits)
	idealV := dac.VrefNeg + float64(code)*lsb

	// Apply non-idealities
	inlError := 0.0
	if code < len(dac.INL) {
		inlError = dac.INL[code] * lsb
	}

	// Apply gain and offset
	actualV := (idealV + inlError) * dac.Gain + dac.Offset

	// Settling behavior
	if dac.CurrentCode != code {
		dac.CurrentCode = code
		dac.SettlingState = dac.CurrentVoltage
	}

	// Exponential settling
	tau := dac.Config.DACSettlingTimeNs
	settled := actualV - (actualV-dac.SettlingState)*math.Exp(-timeNs/tau)

	dac.CurrentVoltage = settled
	return settled
}

// ADCModel implements ADC behavioral model.
type ADCModel struct {
	Config     *MixedSignalConfig
	NumBits    int
	VrefPos    float64
	VrefNeg    float64

	// Non-idealities
	DNL        []float64
	INL        []float64
	Offset     float64
	Gain       float64
	NoiseRMS   float64

	// State
	SampleHold float64
	Pipeline   []int
}

// NewADCModel creates an ADC model.
func NewADCModel(bits int, vref float64, config *MixedSignalConfig) *ADCModel {
	if config == nil {
		config = DefaultMixedSignalConfig()
	}

	levels := 1 << bits

	// Calculate noise from SNR
	snrLinear := math.Pow(10, config.ADCSNR/20)
	noiseRMS := vref / snrLinear

	adc := &ADCModel{
		Config:   config,
		NumBits:  bits,
		VrefPos:  vref,
		VrefNeg:  0,
		DNL:      make([]float64, levels),
		INL:      make([]float64, levels),
		Offset:   rand.NormFloat64() * 0.002 * vref,
		Gain:     1.0 + rand.NormFloat64()*0.002,
		NoiseRMS: noiseRMS,
		Pipeline: make([]int, config.ADCLatencyCycles),
	}

	// Generate errors
	cumINL := 0.0
	for i := 0; i < levels; i++ {
		adc.DNL[i] = rand.NormFloat64() * 0.2
		cumINL += adc.DNL[i]
		adc.INL[i] = cumINL
	}

	return adc
}

// Convert converts analog voltage to digital code.
func (adc *ADCModel) Convert(voltage float64) int {
	// Apply input referred errors
	effectiveV := (voltage - adc.Offset) / adc.Gain

	// Add noise
	effectiveV += adc.NoiseRMS * rand.NormFloat64()

	// Quantize
	lsb := (adc.VrefPos - adc.VrefNeg) / float64(1<<adc.NumBits)
	code := int((effectiveV - adc.VrefNeg) / lsb)

	// Clamp
	if code < 0 {
		code = 0
	}
	maxCode := (1 << adc.NumBits) - 1
	if code > maxCode {
		code = maxCode
	}

	// Apply INL (code-dependent error)
	if code < len(adc.INL) && math.Abs(adc.INL[code]) > 0.5 {
		// Large INL can cause missing codes
		if adc.INL[code] > 0 {
			code--
		} else {
			code++
		}
		if code < 0 {
			code = 0
		}
		if code > maxCode {
			code = maxCode
		}
	}

	return code
}

// ConvertPipelined converts with pipeline latency.
func (adc *ADCModel) ConvertPipelined(voltage float64) int {
	// Shift pipeline
	output := adc.Pipeline[len(adc.Pipeline)-1]
	for i := len(adc.Pipeline) - 1; i > 0; i-- {
		adc.Pipeline[i] = adc.Pipeline[i-1]
	}
	adc.Pipeline[0] = adc.Convert(voltage)

	return output
}

// =============================================================================
// CROSSBAR BEHAVIORAL MODEL
// =============================================================================

// CrossbarBehavioralModel implements crossbar behavioral simulation.
type CrossbarBehavioralModel struct {
	Config *MixedSignalConfig
	Rows   int
	Cols   int

	// Conductance matrix
	Conductances [][]float64

	// Parasitics
	LineResistance [][]float64
	Capacitance    [][]float64

	// DAC/ADC interfaces
	RowDACs []*DACModel
	ColADCs []*ADCModel

	// Timing
	CurrentTimeNs float64
}

// NewCrossbarBehavioralModel creates a crossbar model.
func NewCrossbarBehavioralModel(rows, cols int, config *MixedSignalConfig) *CrossbarBehavioralModel {
	if config == nil {
		config = DefaultMixedSignalConfig()
	}

	model := &CrossbarBehavioralModel{
		Config:         config,
		Rows:           rows,
		Cols:           cols,
		Conductances:   make([][]float64, rows),
		LineResistance: make([][]float64, rows),
		Capacitance:    make([][]float64, rows),
		RowDACs:        make([]*DACModel, rows),
		ColADCs:        make([]*ADCModel, cols),
	}

	// Initialize arrays
	for i := 0; i < rows; i++ {
		model.Conductances[i] = make([]float64, cols)
		model.LineResistance[i] = make([]float64, cols)
		model.Capacitance[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			model.LineResistance[i][j] = config.LineResistanceOhm * float64(i+j+1) / float64(rows+cols)
			model.Capacitance[i][j] = config.ParasiticCapPF
		}

		model.RowDACs[i] = NewDACModel(config.DACBits, 1.0, config)
	}

	for j := 0; j < cols; j++ {
		model.ColADCs[j] = NewADCModel(config.ADCBits, 1.0, config)
	}

	return model
}

// SetConductances sets the conductance matrix.
func (cb *CrossbarBehavioralModel) SetConductances(g [][]float64) {
	for i := 0; i < cb.Rows && i < len(g); i++ {
		for j := 0; j < cb.Cols && j < len(g[i]); j++ {
			cb.Conductances[i][j] = g[i][j]
		}
	}
}

// SimulateMVM simulates matrix-vector multiplication.
func (cb *CrossbarBehavioralModel) SimulateMVM(input []int) *MixedSignalMVMResult {
	cb.CurrentTimeNs += cb.Config.DACSettlingTimeNs

	// Convert inputs through DACs
	rowVoltages := make([]float64, cb.Rows)
	for i := 0; i < cb.Rows && i < len(input); i++ {
		rowVoltages[i] = cb.RowDACs[i].Convert(input[i], cb.CurrentTimeNs)
	}

	// Compute column currents with IR drop
	colCurrents := make([]float64, cb.Cols)
	idealCurrents := make([]float64, cb.Cols)

	for j := 0; j < cb.Cols; j++ {
		for i := 0; i < cb.Rows; i++ {
			// Ideal current
			ideal := rowVoltages[i] * cb.Conductances[i][j]
			idealCurrents[j] += ideal

			// With IR drop
			effectiveV := rowVoltages[i] - colCurrents[j]*cb.LineResistance[i][j]
			colCurrents[j] += effectiveV * cb.Conductances[i][j]
		}

		// Add thermal noise
		noiseK := cb.Config.ThermalNoiseK
		temp := 300.0 // Room temperature
		bandwidth := cb.Config.ADCSamplingRateHz
		noisePower := 4 * noiseK * temp * bandwidth
		colCurrents[j] += math.Sqrt(noisePower) * rand.NormFloat64() * 1e-9
	}

	// Convert through ADCs
	outputCodes := make([]int, cb.Cols)
	for j := 0; j < cb.Cols; j++ {
		// Scale current to voltage for ADC
		scaledV := colCurrents[j] * 1000 // TIA gain simulation
		if scaledV > 1 {
			scaledV = 1
		}
		if scaledV < 0 {
			scaledV = 0
		}
		outputCodes[j] = cb.ColADCs[j].ConvertPipelined(scaledV)
	}

	// Compute errors
	mse := 0.0
	for j := 0; j < cb.Cols; j++ {
		idealCode := int(idealCurrents[j] * 1000 * float64((1<<cb.Config.ADCBits)-1))
		if idealCode > (1<<cb.Config.ADCBits)-1 {
			idealCode = (1 << cb.Config.ADCBits) - 1
		}
		err := float64(outputCodes[j]-idealCode) / float64(1<<cb.Config.ADCBits)
		mse += err * err
	}
	mse /= float64(cb.Cols)

	return &MixedSignalMVMResult{
		InputCodes:    input,
		RowVoltages:   rowVoltages,
		ColCurrents:   colCurrents,
		IdealCurrents: idealCurrents,
		OutputCodes:   outputCodes,
		MSE:           mse,
		LatencyNs:     cb.Config.DACSettlingTimeNs + float64(cb.Config.ADCLatencyCycles)/cb.Config.ADCSamplingRateHz*1e9,
	}
}

// MixedSignalMVMResult contains mixed-signal MVM result.
type MixedSignalMVMResult struct {
	InputCodes    []int
	RowVoltages   []float64
	ColCurrents   []float64
	IdealCurrents []float64
	OutputCodes   []int
	MSE           float64
	LatencyNs     float64
}

// =============================================================================
// CO-SIMULATION FRAMEWORK
// =============================================================================

// CoSimulator manages mixed-signal co-simulation.
type CoSimulator struct {
	Config    *MixedSignalConfig
	Crossbar  *CrossbarBehavioralModel

	// Simulation state
	CurrentTimeNs float64
	EventQueue    []*SimEvent

	// Results
	WaveformHistory []*WaveformSample
	MVMResults      []*MixedSignalMVMResult
}

// SimEvent represents a simulation event.
type SimEvent struct {
	TimeNs    float64
	Type      string
	Component string
	Data      interface{}
}

// WaveformSample records a waveform sample.
type WaveformSample struct {
	TimeNs    float64
	Node      string
	Voltage   float64
	Current   float64
}

// NewCoSimulator creates a co-simulator.
func NewCoSimulator(rows, cols int, config *MixedSignalConfig) *CoSimulator {
	if config == nil {
		config = DefaultMixedSignalConfig()
	}

	return &CoSimulator{
		Config:          config,
		Crossbar:        NewCrossbarBehavioralModel(rows, cols, config),
		EventQueue:      make([]*SimEvent, 0),
		WaveformHistory: make([]*WaveformSample, 0),
		MVMResults:      make([]*MixedSignalMVMResult, 0),
	}
}

// LoadWeights loads weights to crossbar.
func (cs *CoSimulator) LoadWeights(weights [][]float64) {
	cs.Crossbar.SetConductances(weights)
}

// RunMVM runs MVM simulation.
func (cs *CoSimulator) RunMVM(input []int) *MixedSignalMVMResult {
	result := cs.Crossbar.SimulateMVM(input)
	cs.MVMResults = append(cs.MVMResults, result)
	cs.CurrentTimeNs = cs.Crossbar.CurrentTimeNs

	// Record waveforms
	for i, v := range result.RowVoltages {
		cs.WaveformHistory = append(cs.WaveformHistory, &WaveformSample{
			TimeNs:  cs.CurrentTimeNs,
			Node:    "row_" + string(rune('0'+i)),
			Voltage: v,
		})
	}

	return result
}

// GetSimulationSummary returns simulation summary.
func (cs *CoSimulator) GetSimulationSummary() *CoSimSummary {
	totalMSE := 0.0
	maxMSE := 0.0
	totalLatency := 0.0

	for _, r := range cs.MVMResults {
		totalMSE += r.MSE
		if r.MSE > maxMSE {
			maxMSE = r.MSE
		}
		totalLatency += r.LatencyNs
	}

	numOps := len(cs.MVMResults)
	if numOps == 0 {
		numOps = 1
	}

	return &CoSimSummary{
		TotalOperations:   len(cs.MVMResults),
		TotalSimTimeNs:    cs.CurrentTimeNs,
		AverageMSE:        totalMSE / float64(numOps),
		MaxMSE:            maxMSE,
		TotalLatencyNs:    totalLatency,
		WaveformSamples:   len(cs.WaveformHistory),
	}
}

// CoSimSummary contains co-simulation summary.
type CoSimSummary struct {
	TotalOperations int
	TotalSimTimeNs  float64
	AverageMSE      float64
	MaxMSE          float64
	TotalLatencyNs  float64
	WaveformSamples int
}
