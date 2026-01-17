// Package layers provides neural network layer implementations for CIM simulation.
// ecc_power.go implements error correction codes (ECC) for CIM reliability and
// low-power inference optimization techniques.
//
// Research basis:
// - CIM ECC: 16,000× BER reduction, 29.1% area, 26.3% power overhead
// - Hard errors: stuck cells, short cells in crossbar arrays
// - Soft errors: noise-induced bit flips during analog computation
// - Multiple error correction for resistive crossbars
// - DVFS: 40-70% dynamic power reduction, 2-3× leakage improvement
// - Zero skipping: >50% energy reduction at 95% sparsity
// - Adaptive power gating: eliminates leakage in idle blocks
//
// Key metrics:
// - Raw BER in RRAM: ~10^-4 to 10^-6
// - Post-ECC BER: <10^-12 achievable
// - Sparsity-aware: 5 TOPS/W at 95% input sparsity
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// ERROR CORRECTION CODES FOR CIM
// =============================================================================

// ECCType represents different ECC schemes
type ECCType int

const (
	ECCNone           ECCType = iota // No error correction
	ECCHamming                       // Single error correction
	ECCSECDED                        // Single error correct, double detect
	ECCBCH                           // BCH code for multiple errors
	ECCReedSolomon                   // RS code for burst errors
	ECCAnalog                        // Analog domain error correction
	ECCSuccessive                    // Successive correction for CIM
)

// ErrorType represents different error sources in CIM
type ErrorType int

const (
	ErrorNone        ErrorType = iota
	ErrorStuckHigh                    // Cell stuck at high resistance
	ErrorStuckLow                     // Cell stuck at low resistance
	ErrorShortCircuit                 // Short between cells
	ErrorSoftNoise                    // Noise-induced error
	ErrorDrift                        // Resistance drift over time
	ErrorReadDisturb                  // Read operation disturbs state
)

// ECCConfig holds ECC configuration parameters
type ECCConfig struct {
	Type             ECCType
	DataBits         int     // Number of data bits
	ParityBits       int     // Number of parity/check bits
	CorrectionCapacity int   // Number of errors that can be corrected

	// Overhead metrics
	AreaOverheadPercent  float64 // Area overhead (29.1% typical)
	PowerOverheadPercent float64 // Power overhead (26.3% typical)
	LatencyOverheadNs    float64 // Latency overhead

	// BER targets
	RawBER       float64 // Raw bit error rate
	TargetBER    float64 // Target post-correction BER
}

// DefaultECCConfig returns typical CIM ECC configuration
func DefaultECCConfig() *ECCConfig {
	return &ECCConfig{
		Type:                ECCSuccessive,
		DataBits:            64,
		ParityBits:          8,
		CorrectionCapacity:  2, // Double error correction
		AreaOverheadPercent: 29.1,
		PowerOverheadPercent: 26.3,
		LatencyOverheadNs:   5,
		RawBER:              1e-5,
		TargetBER:           1e-12,
	}
}

// HammingECCConfig returns Hamming code configuration
func HammingECCConfig() *ECCConfig {
	return &ECCConfig{
		Type:                ECCHamming,
		DataBits:            64,
		ParityBits:          7, // (64,71) Hamming
		CorrectionCapacity:  1,
		AreaOverheadPercent: 15,
		PowerOverheadPercent: 12,
		LatencyOverheadNs:   2,
		RawBER:              1e-5,
		TargetBER:           1e-9,
	}
}

// BCHECCConfig returns BCH code configuration for multiple errors
func BCHECCConfig(t int) *ECCConfig {
	// BCH(n,k,t) where t is correction capacity
	// Parity bits ≈ m*t where m = ceil(log2(n+1))
	parityBits := 10 * t // Approximation for 64-bit data

	return &ECCConfig{
		Type:                ECCBCH,
		DataBits:            64,
		ParityBits:          parityBits,
		CorrectionCapacity:  t,
		AreaOverheadPercent: 20 + float64(t)*5,
		PowerOverheadPercent: 18 + float64(t)*4,
		LatencyOverheadNs:   float64(t) * 2,
		RawBER:              1e-5,
		TargetBER:           1e-15,
	}
}

// ECCEngine implements error correction for CIM
type ECCEngine struct {
	config *ECCConfig
	rng    *rand.Rand

	// Error injection map (for simulation)
	errorMap map[int]ErrorType

	// Statistics
	totalWords         int64
	uncorrectedErrors  int64
	correctedErrors    int64
	detectedUncorrectable int64
}

// NewECCEngine creates a new ECC engine
func NewECCEngine(config *ECCConfig) *ECCEngine {
	if config == nil {
		config = DefaultECCConfig()
	}

	return &ECCEngine{
		config:   config,
		rng:      rand.New(rand.NewSource(42)),
		errorMap: make(map[int]ErrorType),
	}
}

// InjectError injects an error at specified position (for testing)
func (e *ECCEngine) InjectError(position int, errorType ErrorType) {
	e.errorMap[position] = errorType
}

// Encode adds ECC parity bits to data
func (e *ECCEngine) Encode(data []byte) []byte {
	switch e.config.Type {
	case ECCHamming:
		return e.encodeHamming(data)
	case ECCSECDED:
		return e.encodeSECDED(data)
	case ECCBCH:
		return e.encodeBCH(data)
	case ECCSuccessive:
		return e.encodeSuccessive(data)
	default:
		// No encoding, return copy
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}
}

// encodeHamming implements Hamming code encoding
func (e *ECCEngine) encodeHamming(data []byte) []byte {
	// Simplified: add parity bytes
	parityBytes := (e.config.ParityBits + 7) / 8
	encoded := make([]byte, len(data)+parityBytes)
	copy(encoded, data)

	// Calculate parity (XOR-based for simulation)
	parity := byte(0)
	for _, b := range data {
		parity ^= b
	}
	encoded[len(data)] = parity

	return encoded
}

// encodeSECDED implements SEC-DED encoding
func (e *ECCEngine) encodeSECDED(data []byte) []byte {
	// SEC-DED adds overall parity to Hamming
	hamming := e.encodeHamming(data)

	// Add overall parity byte
	overallParity := byte(0)
	for _, b := range hamming {
		overallParity ^= b
	}

	result := make([]byte, len(hamming)+1)
	copy(result, hamming)
	result[len(hamming)] = overallParity

	return result
}

// encodeBCH implements BCH code encoding (simplified)
func (e *ECCEngine) encodeBCH(data []byte) []byte {
	// Simplified BCH: polynomial-based syndrome
	parityBytes := (e.config.ParityBits + 7) / 8
	encoded := make([]byte, len(data)+parityBytes)
	copy(encoded, data)

	// Generate syndrome bytes
	for i := 0; i < parityBytes; i++ {
		syndrome := byte(0)
		for j, b := range data {
			syndrome ^= b * byte((j+1+i)%256)
		}
		encoded[len(data)+i] = syndrome
	}

	return encoded
}

// encodeSuccessive implements successive correction encoding
func (e *ECCEngine) encodeSuccessive(data []byte) []byte {
	// Multi-level encoding for CIM
	// Level 1: Hamming for soft errors
	level1 := e.encodeHamming(data)

	// Level 2: Additional redundancy for hard errors
	checksum := byte(0)
	for i, b := range data {
		checksum ^= byte((int(b) * (i + 1)) % 256)
	}

	result := make([]byte, len(level1)+1)
	copy(result, level1)
	result[len(level1)] = checksum

	return result
}

// Decode decodes and corrects errors in received data
func (e *ECCEngine) Decode(encoded []byte) ([]byte, int, error) {
	e.totalWords++

	// Simulate error injection based on BER
	corrupted := e.simulateErrors(encoded)

	switch e.config.Type {
	case ECCHamming:
		return e.decodeHamming(corrupted)
	case ECCSECDED:
		return e.decodeSECDED(corrupted)
	case ECCBCH:
		return e.decodeBCH(corrupted)
	case ECCSuccessive:
		return e.decodeSuccessive(corrupted)
	default:
		// No decoding
		return corrupted, 0, nil
	}
}

// simulateErrors injects random errors based on BER
func (e *ECCEngine) simulateErrors(data []byte) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	for i := range result {
		// Check for injected errors
		if errType, exists := e.errorMap[i]; exists {
			result[i] = e.applyError(result[i], errType)
		}

		// Random errors based on BER
		for bit := 0; bit < 8; bit++ {
			if e.rng.Float64() < e.config.RawBER {
				result[i] ^= (1 << bit) // Flip bit
			}
		}
	}

	return result
}

// applyError applies specific error type to byte
func (e *ECCEngine) applyError(b byte, errType ErrorType) byte {
	switch errType {
	case ErrorStuckHigh:
		return 0xFF
	case ErrorStuckLow:
		return 0x00
	case ErrorSoftNoise:
		return b ^ byte(e.rng.Intn(256))
	default:
		return b
	}
}

// decodeHamming implements Hamming code decoding
func (e *ECCEngine) decodeHamming(encoded []byte) ([]byte, int, error) {
	if len(encoded) < 2 {
		return nil, 0, fmt.Errorf("data too short for Hamming decode")
	}

	dataLen := len(encoded) - 1
	data := encoded[:dataLen]
	storedParity := encoded[dataLen]

	// Calculate parity
	calculatedParity := byte(0)
	for _, b := range data {
		calculatedParity ^= b
	}

	syndrome := storedParity ^ calculatedParity

	if syndrome == 0 {
		// No errors
		return data, 0, nil
	}

	// Single bit error - attempt correction
	e.correctedErrors++

	// Find error position (simplified)
	errorPos := int(syndrome) % len(data)
	corrected := make([]byte, len(data))
	copy(corrected, data)
	corrected[errorPos] ^= syndrome

	return corrected, 1, nil
}

// decodeSECDED implements SEC-DED decoding
func (e *ECCEngine) decodeSECDED(encoded []byte) ([]byte, int, error) {
	if len(encoded) < 3 {
		return nil, 0, fmt.Errorf("data too short for SEC-DED decode")
	}

	// Check overall parity
	overallParity := byte(0)
	for _, b := range encoded {
		overallParity ^= b
	}

	// Remove overall parity byte for Hamming decode
	hammingData := encoded[:len(encoded)-1]
	decoded, errors, err := e.decodeHamming(hammingData)

	if err != nil {
		return nil, 0, err
	}

	// If overall parity is wrong and Hamming detected error,
	// we have a correctable single error
	// If overall parity is correct and Hamming detected error,
	// we have an uncorrectable double error
	if overallParity != 0 && errors == 0 {
		// Double error detected but not correctable
		e.detectedUncorrectable++
		return decoded, 2, fmt.Errorf("double error detected")
	}

	return decoded, errors, nil
}

// decodeBCH implements BCH code decoding (simplified)
func (e *ECCEngine) decodeBCH(encoded []byte) ([]byte, int, error) {
	parityBytes := (e.config.ParityBits + 7) / 8
	if len(encoded) < parityBytes+1 {
		return nil, 0, fmt.Errorf("data too short for BCH decode")
	}

	dataLen := len(encoded) - parityBytes
	data := make([]byte, dataLen)
	copy(data, encoded[:dataLen])

	// Verify syndromes
	errorsFound := 0
	for i := 0; i < parityBytes; i++ {
		syndrome := byte(0)
		for j, b := range data {
			syndrome ^= b * byte((j+1+i)%256)
		}
		storedSyndrome := encoded[dataLen+i]

		if syndrome != storedSyndrome {
			errorsFound++
		}
	}

	if errorsFound > 0 {
		if errorsFound <= e.config.CorrectionCapacity {
			// Attempt correction (simplified)
			e.correctedErrors++
			// In reality, would use Berlekamp-Massey algorithm
		} else {
			e.uncorrectedErrors++
			return data, errorsFound, fmt.Errorf("too many errors to correct")
		}
	}

	return data, errorsFound, nil
}

// decodeSuccessive implements successive correction decoding
func (e *ECCEngine) decodeSuccessive(encoded []byte) ([]byte, int, error) {
	if len(encoded) < 3 {
		return nil, 0, fmt.Errorf("data too short for successive decode")
	}

	// Level 1: Hamming decode
	level1Data := encoded[:len(encoded)-1]
	decoded, errors1, _ := e.decodeHamming(level1Data)

	// Level 2: Checksum verification
	checksum := byte(0)
	for i, b := range decoded {
		checksum ^= byte((int(b) * (i + 1)) % 256)
	}

	storedChecksum := encoded[len(encoded)-1]
	if checksum != storedChecksum {
		// Additional error detected
		errors1++
		if errors1 > e.config.CorrectionCapacity {
			e.uncorrectedErrors++
			return decoded, errors1, fmt.Errorf("successive correction failed")
		}
		e.correctedErrors++
	}

	return decoded, errors1, nil
}

// GetBERReduction returns the BER reduction factor
func (e *ECCEngine) GetBERReduction() float64 {
	if e.totalWords == 0 {
		return 0
	}

	// Calculate effective BER
	effectiveBER := float64(e.uncorrectedErrors) / float64(e.totalWords*int64(e.config.DataBits))
	if effectiveBER == 0 {
		effectiveBER = e.config.TargetBER // Use target if no errors
	}

	return e.config.RawBER / effectiveBER
}

// GetStatistics returns ECC operation statistics
func (e *ECCEngine) GetStatistics() ECCStats {
	return ECCStats{
		TotalWords:            e.totalWords,
		CorrectedErrors:       e.correctedErrors,
		UncorrectedErrors:     e.uncorrectedErrors,
		DetectedUncorrectable: e.detectedUncorrectable,
		CorrectionRate:        float64(e.correctedErrors) / float64(e.totalWords+1),
		BERReduction:          e.GetBERReduction(),
	}
}

// ECCStats holds ECC operation statistics
type ECCStats struct {
	TotalWords            int64
	CorrectedErrors       int64
	UncorrectedErrors     int64
	DetectedUncorrectable int64
	CorrectionRate        float64
	BERReduction          float64
}

// =============================================================================
// ANALOG ERROR CORRECTION FOR CIM
// =============================================================================

// AnalogECCConfig configures analog domain error correction
type AnalogECCConfig struct {
	// Redundancy
	RedundantRows     int     // Extra rows for parity
	RedundantCols     int     // Extra columns for parity
	ChecksumWeights   bool    // Use checksum weights

	// Error detection thresholds
	OutlierThreshold  float64 // Threshold for outlier detection
	DriftThreshold    float64 // Threshold for drift detection

	// Correction parameters
	InterpolateErrors bool    // Use neighboring values for correction
	VotingScheme      bool    // Use majority voting
}

// DefaultAnalogECCConfig returns typical analog ECC config
func DefaultAnalogECCConfig() *AnalogECCConfig {
	return &AnalogECCConfig{
		RedundantRows:    2,
		RedundantCols:    2,
		ChecksumWeights:  true,
		OutlierThreshold: 3.0, // 3 sigma
		DriftThreshold:   0.1, // 10% drift
		InterpolateErrors: true,
		VotingScheme:      false,
	}
}

// AnalogECC implements analog domain error correction
type AnalogECC struct {
	config *AnalogECCConfig
	rng    *rand.Rand

	// Checksum weights (row and column sums)
	rowChecksums []float64
	colChecksums []float64

	// Statistics
	outliersCorrected int
	driftsCorrected   int
}

// NewAnalogECC creates a new analog ECC module
func NewAnalogECC(rows, cols int, config *AnalogECCConfig) *AnalogECC {
	if config == nil {
		config = DefaultAnalogECCConfig()
	}

	return &AnalogECC{
		config:       config,
		rng:          rand.New(rand.NewSource(42)),
		rowChecksums: make([]float64, rows),
		colChecksums: make([]float64, cols),
	}
}

// ComputeChecksums computes row and column checksums for weight matrix
func (a *AnalogECC) ComputeChecksums(weights [][]float64) {
	rows := len(weights)
	if rows == 0 {
		return
	}
	cols := len(weights[0])

	// Row checksums
	a.rowChecksums = make([]float64, rows)
	for i := 0; i < rows; i++ {
		sum := 0.0
		for j := 0; j < cols; j++ {
			sum += weights[i][j]
		}
		a.rowChecksums[i] = sum
	}

	// Column checksums
	a.colChecksums = make([]float64, cols)
	for j := 0; j < cols; j++ {
		sum := 0.0
		for i := 0; i < rows; i++ {
			sum += weights[i][j]
		}
		a.colChecksums[j] = sum
	}
}

// VerifyAndCorrect verifies matrix against checksums and corrects errors
func (a *AnalogECC) VerifyAndCorrect(weights [][]float64) ([][]float64, int) {
	rows := len(weights)
	if rows == 0 {
		return weights, 0
	}
	cols := len(weights[0])

	corrected := make([][]float64, rows)
	for i := range corrected {
		corrected[i] = make([]float64, cols)
		copy(corrected[i], weights[i])
	}

	errorsFound := 0

	// Check row checksums
	for i := 0; i < rows; i++ {
		sum := 0.0
		for j := 0; j < cols; j++ {
			sum += corrected[i][j]
		}

		diff := math.Abs(sum - a.rowChecksums[i])
		if diff > a.config.DriftThreshold*math.Abs(a.rowChecksums[i]+1e-10) {
			// Error in this row - find and correct outlier
			errorsFound++
			a.correctRowError(corrected, i)
		}
	}

	// Check column checksums
	for j := 0; j < cols; j++ {
		sum := 0.0
		for i := 0; i < rows; i++ {
			sum += corrected[i][j]
		}

		diff := math.Abs(sum - a.colChecksums[j])
		if diff > a.config.DriftThreshold*math.Abs(a.colChecksums[j]+1e-10) {
			errorsFound++
			a.correctColError(corrected, j)
		}
	}

	return corrected, errorsFound
}

// correctRowError finds and corrects error in a row
func (a *AnalogECC) correctRowError(weights [][]float64, row int) {
	cols := len(weights[row])

	// Find outlier using statistical method
	mean := 0.0
	for j := 0; j < cols; j++ {
		mean += weights[row][j]
	}
	mean /= float64(cols)

	variance := 0.0
	for j := 0; j < cols; j++ {
		diff := weights[row][j] - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(cols))

	// Find values beyond threshold
	for j := 0; j < cols; j++ {
		zScore := math.Abs(weights[row][j]-mean) / (stdDev + 1e-10)
		if zScore > a.config.OutlierThreshold {
			// Correct using interpolation
			if a.config.InterpolateErrors {
				weights[row][j] = a.interpolateValue(weights, row, j)
			} else {
				weights[row][j] = mean
			}
			a.outliersCorrected++
		}
	}
}

// correctColError finds and corrects error in a column
func (a *AnalogECC) correctColError(weights [][]float64, col int) {
	rows := len(weights)

	mean := 0.0
	for i := 0; i < rows; i++ {
		mean += weights[i][col]
	}
	mean /= float64(rows)

	variance := 0.0
	for i := 0; i < rows; i++ {
		diff := weights[i][col] - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(rows))

	for i := 0; i < rows; i++ {
		zScore := math.Abs(weights[i][col]-mean) / (stdDev + 1e-10)
		if zScore > a.config.OutlierThreshold {
			if a.config.InterpolateErrors {
				weights[i][col] = a.interpolateValue(weights, i, col)
			} else {
				weights[i][col] = mean
			}
			a.outliersCorrected++
		}
	}
}

// interpolateValue estimates value from neighbors
func (a *AnalogECC) interpolateValue(weights [][]float64, row, col int) float64 {
	rows := len(weights)
	cols := len(weights[0])

	sum := 0.0
	count := 0

	// Average of 4 neighbors
	neighbors := []struct{ r, c int }{
		{row - 1, col}, {row + 1, col}, {row, col - 1}, {row, col + 1},
	}

	for _, n := range neighbors {
		if n.r >= 0 && n.r < rows && n.c >= 0 && n.c < cols {
			sum += weights[n.r][n.c]
			count++
		}
	}

	if count > 0 {
		return sum / float64(count)
	}
	return 0
}

// =============================================================================
// LOW-POWER INFERENCE OPTIMIZATION
// =============================================================================

// PowerState represents power management state
type PowerState int

const (
	PowerActive     PowerState = iota // Full power
	PowerLowVoltage                   // Reduced voltage (DVS)
	PowerClockGated                   // Clock gated
	PowerGated                        // Power gated (off)
)

// PowerConfig configures power management
type PowerConfig struct {
	// Voltage scaling
	NominalVoltageV    float64 // Normal operating voltage
	MinVoltageV        float64 // Minimum safe voltage
	VoltageStepV       float64 // Voltage step size

	// Power gating
	EnablePowerGating  bool
	WakeupLatencyNs    float64 // Time to wake from power gated
	PowerGatingThresholdNs float64 // Min idle time for power gating

	// Clock gating
	EnableClockGating  bool

	// Sparsity optimization
	EnableZeroSkipping bool
	SparsityThreshold  float64 // Min sparsity for optimization

	// Activity monitoring
	ActivityWindowSize int
}

// DefaultPowerConfig returns typical power configuration
func DefaultPowerConfig() *PowerConfig {
	return &PowerConfig{
		NominalVoltageV:        1.0,
		MinVoltageV:            0.6,
		VoltageStepV:           0.05,
		EnablePowerGating:      true,
		WakeupLatencyNs:        100,
		PowerGatingThresholdNs: 1000,
		EnableClockGating:      true,
		EnableZeroSkipping:     true,
		SparsityThreshold:      0.5, // 50% zeros
		ActivityWindowSize:     100,
	}
}

// PowerManager implements power optimization for CIM
type PowerManager struct {
	config *PowerConfig

	// Current state
	currentState    PowerState
	currentVoltage  float64
	idleTimeNs      float64

	// Activity tracking
	activityHistory []float64
	activityIndex   int

	// Energy tracking
	totalEnergyPJ   float64
	activeEnergyPJ  float64
	leakageEnergyPJ float64

	// Statistics
	powerGateEvents int
	clockGateEvents int
	dvfsEvents      int
}

// NewPowerManager creates a new power manager
func NewPowerManager(config *PowerConfig) *PowerManager {
	if config == nil {
		config = DefaultPowerConfig()
	}

	return &PowerManager{
		config:          config,
		currentState:    PowerActive,
		currentVoltage:  config.NominalVoltageV,
		activityHistory: make([]float64, config.ActivityWindowSize),
	}
}

// UpdateActivity updates activity tracking
func (pm *PowerManager) UpdateActivity(activity float64) {
	pm.activityHistory[pm.activityIndex] = activity
	pm.activityIndex = (pm.activityIndex + 1) % pm.config.ActivityWindowSize
}

// GetAverageActivity returns average activity over window
func (pm *PowerManager) GetAverageActivity() float64 {
	sum := 0.0
	for _, a := range pm.activityHistory {
		sum += a
	}
	return sum / float64(len(pm.activityHistory))
}

// OptimizePowerState determines optimal power state
func (pm *PowerManager) OptimizePowerState(workloadSize int, sparsity float64) PowerState {
	avgActivity := pm.GetAverageActivity()

	// Check if should power gate
	if pm.config.EnablePowerGating && avgActivity < 0.01 {
		if pm.idleTimeNs > pm.config.PowerGatingThresholdNs {
			pm.currentState = PowerGated
			pm.powerGateEvents++
			return PowerGated
		}
	}

	// Check if should clock gate
	if pm.config.EnableClockGating && avgActivity < 0.1 {
		pm.currentState = PowerClockGated
		pm.clockGateEvents++
		return PowerClockGated
	}

	// DVFS based on workload
	if workloadSize < 1000 && avgActivity < 0.5 {
		// Reduce voltage for light workloads
		newVoltage := pm.config.NominalVoltageV * 0.8
		if newVoltage >= pm.config.MinVoltageV {
			pm.currentVoltage = newVoltage
			pm.currentState = PowerLowVoltage
			pm.dvfsEvents++
			return PowerLowVoltage
		}
	}

	pm.currentState = PowerActive
	pm.currentVoltage = pm.config.NominalVoltageV
	return PowerActive
}

// CalculatePowerSavings calculates power savings from current optimizations
func (pm *PowerManager) CalculatePowerSavings() PowerSavings {
	// Dynamic power scales as V²
	voltageRatio := pm.currentVoltage / pm.config.NominalVoltageV
	dynamicReduction := 1 - voltageRatio*voltageRatio

	// Leakage reduction from power gating
	leakageReduction := 0.0
	if pm.currentState == PowerGated {
		leakageReduction = 0.99 // 99% leakage reduction when gated
	} else if pm.currentState == PowerClockGated {
		leakageReduction = 0.0 // Clock gating doesn't reduce leakage
	}

	return PowerSavings{
		DynamicPowerReduction: dynamicReduction * 100,
		LeakageReduction:      leakageReduction * 100,
		CurrentVoltage:        pm.currentVoltage,
		CurrentState:          pm.currentState,
	}
}

// PowerSavings holds power saving metrics
type PowerSavings struct {
	DynamicPowerReduction float64 // Percentage
	LeakageReduction      float64 // Percentage
	CurrentVoltage        float64
	CurrentState          PowerState
}

// =============================================================================
// SPARSITY-AWARE ACCELERATION
// =============================================================================

// SparsityConfig configures sparsity optimization
type SparsityConfig struct {
	EnableInputSparsity  bool
	EnableWeightSparsity bool
	EnableOutputSparsity bool

	// Zero detection threshold
	ZeroThreshold float64 // Values below this are treated as zero

	// Structured sparsity
	BlockSize       int  // For block sparsity (0 = unstructured)
	ChannelPruning  bool // Enable channel-wise pruning
}

// DefaultSparsityConfig returns typical sparsity configuration
func DefaultSparsityConfig() *SparsityConfig {
	return &SparsityConfig{
		EnableInputSparsity:  true,
		EnableWeightSparsity: true,
		EnableOutputSparsity: false,
		ZeroThreshold:        1e-6,
		BlockSize:            0, // Unstructured
		ChannelPruning:       false,
	}
}

// SparsityAccelerator implements sparsity-aware computation
type SparsityAccelerator struct {
	config *SparsityConfig

	// Statistics
	totalOps     int64
	skippedOps   int64
	effectiveOps int64
}

// NewSparsityAccelerator creates a new sparsity accelerator
func NewSparsityAccelerator(config *SparsityConfig) *SparsityAccelerator {
	if config == nil {
		config = DefaultSparsityConfig()
	}

	return &SparsityAccelerator{
		config: config,
	}
}

// ComputeSparsity calculates sparsity of a vector
func (sa *SparsityAccelerator) ComputeSparsity(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	zeros := 0
	for _, v := range data {
		if math.Abs(v) < sa.config.ZeroThreshold {
			zeros++
		}
	}

	return float64(zeros) / float64(len(data))
}

// SparseMVM performs sparsity-aware matrix-vector multiply
func (sa *SparsityAccelerator) SparseMVM(
	inputs []float64,
	weights [][]float64,
) ([]float64, SparseMVMStats) {
	if len(weights) == 0 {
		return nil, SparseMVMStats{}
	}

	rows := len(weights)
	cols := len(weights[0])
	outputs := make([]float64, rows)

	totalOps := int64(0)
	skippedOps := int64(0)

	// Compute input sparsity mask
	inputMask := make([]bool, len(inputs))
	if sa.config.EnableInputSparsity {
		for i, v := range inputs {
			inputMask[i] = math.Abs(v) >= sa.config.ZeroThreshold
		}
	} else {
		for i := range inputMask {
			inputMask[i] = true
		}
	}

	// Sparse computation
	for i := 0; i < rows; i++ {
		sum := 0.0
		for j := 0; j < cols; j++ {
			totalOps++

			// Skip if input is zero
			if sa.config.EnableInputSparsity && !inputMask[j] {
				skippedOps++
				continue
			}

			// Skip if weight is zero
			if sa.config.EnableWeightSparsity &&
				math.Abs(weights[i][j]) < sa.config.ZeroThreshold {
				skippedOps++
				continue
			}

			// Actual computation
			sum += inputs[j] * weights[i][j]
		}
		outputs[i] = sum
	}

	sa.totalOps += totalOps
	sa.skippedOps += skippedOps
	sa.effectiveOps += totalOps - skippedOps

	inputSparsity := sa.ComputeSparsity(inputs)
	weightSparsity := 0.0
	totalWeights := 0
	zeroWeights := 0
	for i := range weights {
		for j := range weights[i] {
			totalWeights++
			if math.Abs(weights[i][j]) < sa.config.ZeroThreshold {
				zeroWeights++
			}
		}
	}
	if totalWeights > 0 {
		weightSparsity = float64(zeroWeights) / float64(totalWeights)
	}

	return outputs, SparseMVMStats{
		TotalOps:        totalOps,
		SkippedOps:      skippedOps,
		EffectiveOps:    totalOps - skippedOps,
		SkipRatio:       float64(skippedOps) / float64(totalOps+1),
		InputSparsity:   inputSparsity,
		WeightSparsity:  weightSparsity,
		EnergySavings:   float64(skippedOps) / float64(totalOps+1) * 100,
	}
}

// SparseMVMStats holds sparse MVM statistics
type SparseMVMStats struct {
	TotalOps       int64
	SkippedOps     int64
	EffectiveOps   int64
	SkipRatio      float64
	InputSparsity  float64
	WeightSparsity float64
	EnergySavings  float64 // Percentage
}

// GetOverallStats returns overall sparsity acceleration statistics
func (sa *SparsityAccelerator) GetOverallStats() SparsityStats {
	return SparsityStats{
		TotalOps:     sa.totalOps,
		SkippedOps:   sa.skippedOps,
		EffectiveOps: sa.effectiveOps,
		OverallSkipRatio: float64(sa.skippedOps) / float64(sa.totalOps+1),
		EffectiveTOPSMultiplier: float64(sa.totalOps) / float64(sa.effectiveOps+1),
	}
}

// SparsityStats holds overall sparsity statistics
type SparsityStats struct {
	TotalOps               int64
	SkippedOps             int64
	EffectiveOps           int64
	OverallSkipRatio       float64
	EffectiveTOPSMultiplier float64
}

// =============================================================================
// INTEGRATED POWER-EFFICIENT CIM SYSTEM
// =============================================================================

// EfficientCIMConfig configures power-efficient CIM
type EfficientCIMConfig struct {
	ECCConfig      *ECCConfig
	PowerConfig    *PowerConfig
	SparsityConfig *SparsityConfig
	AnalogECCConfig *AnalogECCConfig
}

// DefaultEfficientCIMConfig returns typical efficient CIM config
func DefaultEfficientCIMConfig() *EfficientCIMConfig {
	return &EfficientCIMConfig{
		ECCConfig:       DefaultECCConfig(),
		PowerConfig:     DefaultPowerConfig(),
		SparsityConfig:  DefaultSparsityConfig(),
		AnalogECCConfig: DefaultAnalogECCConfig(),
	}
}

// EfficientCIMSystem combines ECC and power optimization
type EfficientCIMSystem struct {
	config      *EfficientCIMConfig
	eccEngine   *ECCEngine
	analogECC   *AnalogECC
	powerMgr    *PowerManager
	sparsityAcc *SparsityAccelerator

	// Array dimensions
	rows int
	cols int
}

// NewEfficientCIMSystem creates an integrated efficient CIM system
func NewEfficientCIMSystem(rows, cols int, config *EfficientCIMConfig) *EfficientCIMSystem {
	if config == nil {
		config = DefaultEfficientCIMConfig()
	}

	return &EfficientCIMSystem{
		config:      config,
		eccEngine:   NewECCEngine(config.ECCConfig),
		analogECC:   NewAnalogECC(rows, cols, config.AnalogECCConfig),
		powerMgr:    NewPowerManager(config.PowerConfig),
		sparsityAcc: NewSparsityAccelerator(config.SparsityConfig),
		rows:        rows,
		cols:        cols,
	}
}

// ProcessLayer processes a layer with all optimizations
func (sys *EfficientCIMSystem) ProcessLayer(
	inputs []float64,
	weights [][]float64,
) ([]float64, LayerStats) {
	// Compute sparsity
	inputSparsity := sys.sparsityAcc.ComputeSparsity(inputs)

	// Optimize power state
	sys.powerMgr.UpdateActivity(1 - inputSparsity)
	powerState := sys.powerMgr.OptimizePowerState(len(inputs)*len(weights), inputSparsity)

	// Verify and correct weights (analog ECC)
	correctedWeights, weightErrors := sys.analogECC.VerifyAndCorrect(weights)

	// Sparse computation
	outputs, sparseStats := sys.sparsityAcc.SparseMVM(inputs, correctedWeights)

	// Power savings
	powerSavings := sys.powerMgr.CalculatePowerSavings()

	return outputs, LayerStats{
		InputSparsity:   inputSparsity,
		WeightErrors:    weightErrors,
		SkippedOps:      sparseStats.SkippedOps,
		EnergySavings:   sparseStats.EnergySavings + powerSavings.DynamicPowerReduction,
		PowerState:      powerState,
	}
}

// LayerStats holds layer processing statistics
type LayerStats struct {
	InputSparsity float64
	WeightErrors  int
	SkippedOps    int64
	EnergySavings float64
	PowerState    PowerState
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// FormatECCReport generates ECC analysis report
func FormatECCReport(engine *ECCEngine) string {
	stats := engine.GetStatistics()

	report := "=== CIM ECC Analysis Report ===\n\n"
	report += fmt.Sprintf("ECC Type: %s\n", eccTypeName(engine.config.Type))
	report += fmt.Sprintf("Data Bits: %d, Parity Bits: %d\n",
		engine.config.DataBits, engine.config.ParityBits)
	report += fmt.Sprintf("Correction Capacity: %d errors\n", engine.config.CorrectionCapacity)
	report += fmt.Sprintf("\nStatistics:\n")
	report += fmt.Sprintf("  Total Words: %d\n", stats.TotalWords)
	report += fmt.Sprintf("  Corrected Errors: %d\n", stats.CorrectedErrors)
	report += fmt.Sprintf("  Uncorrected Errors: %d\n", stats.UncorrectedErrors)
	report += fmt.Sprintf("  BER Reduction: %.2e×\n", stats.BERReduction)
	report += fmt.Sprintf("\nOverhead:\n")
	report += fmt.Sprintf("  Area: %.1f%%\n", engine.config.AreaOverheadPercent)
	report += fmt.Sprintf("  Power: %.1f%%\n", engine.config.PowerOverheadPercent)

	return report
}

func eccTypeName(t ECCType) string {
	names := map[ECCType]string{
		ECCNone:       "None",
		ECCHamming:    "Hamming",
		ECCSECDED:     "SEC-DED",
		ECCBCH:        "BCH",
		ECCReedSolomon: "Reed-Solomon",
		ECCAnalog:     "Analog",
		ECCSuccessive: "Successive Correction",
	}
	return names[t]
}

// FormatPowerReport generates power optimization report
func FormatPowerReport(pm *PowerManager) string {
	savings := pm.CalculatePowerSavings()

	stateNames := map[PowerState]string{
		PowerActive:     "Active",
		PowerLowVoltage: "Low Voltage (DVS)",
		PowerClockGated: "Clock Gated",
		PowerGated:      "Power Gated",
	}

	report := "=== Power Optimization Report ===\n\n"
	report += fmt.Sprintf("Current State: %s\n", stateNames[savings.CurrentState])
	report += fmt.Sprintf("Current Voltage: %.2f V\n", savings.CurrentVoltage)
	report += fmt.Sprintf("Dynamic Power Reduction: %.1f%%\n", savings.DynamicPowerReduction)
	report += fmt.Sprintf("Leakage Reduction: %.1f%%\n", savings.LeakageReduction)
	report += fmt.Sprintf("\nEvents:\n")
	report += fmt.Sprintf("  Power Gate Events: %d\n", pm.powerGateEvents)
	report += fmt.Sprintf("  Clock Gate Events: %d\n", pm.clockGateEvents)
	report += fmt.Sprintf("  DVFS Events: %d\n", pm.dvfsEvents)

	return report
}

// EstimateEfficiency estimates overall system efficiency
func EstimateEfficiency(
	baseEfficiencyTOPSW float64,
	sparsity float64,
	eccOverhead float64,
	dvfsReduction float64,
) float64 {
	// Sparsity improves effective throughput
	sparsityMultiplier := 1 / (1 - sparsity + 0.01)

	// ECC adds overhead
	eccFactor := 1 - eccOverhead/100

	// DVS reduces power
	dvsFactor := 1 + dvfsReduction/100

	return baseEfficiencyTOPSW * sparsityMultiplier * eccFactor * dvsFactor
}
