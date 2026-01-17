// optical_inmem_training.go - Optical Interconnects and In-Memory Training for CIM
// Research iteration 118: Silicon photonics, MZI accelerators, and analog on-chip learning
//
// Key findings:
// - Silicon photonics: 1 Tbps WDM microring arrays, sub-10 fJ/bit
// - MZI coherent ONNs: 1.28 TOPS, complex-valued MVM
// - Taichi photonic chiplet: 160 TOPS/W (Science 2024)
// - Thin-film LiNbO3: 120 GOPS with in-situ training
// - CMO/HfOx ReRAM: All-in-one training + inference (IBM 2025)
// - Tiki-Taka algorithm: Handles device asymmetry via coupled matrices
// - Equilibrium propagation: Local learning without explicit gradients
// - Outer product updates: O(1) parallel weight updates in crossbar

package layers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"math/rand"
)

// ============================================================================
// SECTION 1: Silicon Photonics Fundamentals
// ============================================================================

// PhotonicConfig configures silicon photonic components
type PhotonicConfig struct {
	// Waveguide parameters
	WaveguideWidthNm    float64 // Typically 450-500 nm
	WaveguideLossdBcm   float64 // Propagation loss (0.5-2 dB/cm)

	// Operating wavelength
	CenterWavelengthNm  float64 // C-band: 1550 nm typical
	WavelengthChannels  int     // WDM channels
	ChannelSpacingGHz   float64 // Channel spacing (50-100 GHz)

	// Microring parameters
	MicroringRadiusUm   float64 // 5-20 µm typical
	MicroringQFactor    float64 // Quality factor (10k-100k)
	ThermalTuningRange  float64 // nm per mW

	// Modulator specs
	ModulationSpeedGbps float64 // Per-channel speed
	EnergyPerBitFJ      float64 // Energy efficiency
}

// DefaultPhotonicConfig returns typical silicon photonics configuration
func DefaultPhotonicConfig() *PhotonicConfig {
	return &PhotonicConfig{
		WaveguideWidthNm:    450,
		WaveguideLossdBcm:   1.0,
		CenterWavelengthNm:  1550,
		WavelengthChannels:  8,
		ChannelSpacingGHz:   100,
		MicroringRadiusUm:   10,
		MicroringQFactor:    50000,
		ThermalTuningRange:  0.1,
		ModulationSpeedGbps: 200,
		EnergyPerBitFJ:      10,
	}
}

// MicroringResonator models a silicon microring resonator
type MicroringResonator struct {
	Config            *PhotonicConfig
	ResonanceWavelength float64 // Current resonance (nm)
	CouplingRatio     float64   // Power coupling (0-1)
	ThermalShift      float64   // Thermal-induced shift (nm)

	// State
	TransmissionCoeff complex128 // Through port transmission
	DropCoeff         complex128 // Drop port transmission
}

// NewMicroringResonator creates a new microring resonator
func NewMicroringResonator(config *PhotonicConfig) *MicroringResonator {
	if config == nil {
		config = DefaultPhotonicConfig()
	}
	return &MicroringResonator{
		Config:              config,
		ResonanceWavelength: config.CenterWavelengthNm,
		CouplingRatio:       0.2,
	}
}

// ComputeTransfer calculates the transfer function at a given wavelength
func (mrr *MicroringResonator) ComputeTransfer(wavelengthNm float64) (through, drop complex128) {
	// Detuning from resonance
	deltaLambda := wavelengthNm - (mrr.ResonanceWavelength + mrr.ThermalShift)

	// Convert to frequency detuning
	fsr := mrr.Config.CenterWavelengthNm * mrr.Config.CenterWavelengthNm /
		(2 * math.Pi * mrr.Config.MicroringRadiusUm * 1000) // FSR in nm

	// Normalized detuning
	phi := 2 * math.Pi * deltaLambda / fsr

	// Coupling coefficients
	kappa := math.Sqrt(mrr.CouplingRatio)
	t := math.Sqrt(1 - mrr.CouplingRatio)

	// Round-trip loss
	alpha := math.Exp(-mrr.Config.WaveguideLossdBcm * 0.23 *
		2 * math.Pi * mrr.Config.MicroringRadiusUm / 10000)

	// Transfer matrix approach
	roundTrip := complex(alpha*math.Cos(phi), alpha*math.Sin(phi))

	// Through port
	numerator := complex(t, 0) - roundTrip
	denominator := complex(1, 0) - complex(t, 0)*roundTrip
	through = numerator / denominator

	// Drop port
	drop = complex(kappa*kappa*math.Sqrt(alpha), 0) * cmplx.Exp(complex(0, phi/2)) / denominator

	mrr.TransmissionCoeff = through
	mrr.DropCoeff = drop

	return through, drop
}

// SetWeight programs the microring as a programmable weight
func (mrr *MicroringResonator) SetWeight(weight float64) {
	// Map weight (-1 to 1) to thermal shift
	// Weight = 1 means on-resonance (full drop), weight = -1 means off-resonance
	maxShift := mrr.Config.ThermalTuningRange * 10 // 10 mW max power
	mrr.ThermalShift = (1 - weight) * maxShift / 2
}

// ============================================================================
// SECTION 2: Mach-Zehnder Interferometer (MZI) Networks
// ============================================================================

// MZIConfig configures a Mach-Zehnder interferometer
type MZIConfig struct {
	// Physical parameters
	ArmLengthMm       float64 // MZI arm length
	PhaseShifterType  string  // "thermal", "carrier", "pn"

	// Phase shifter specs
	VpiLengthMmV      float64 // Vπ·L product (mm·V)
	ThermalEfficiency float64 // rad/mW for thermal
	MaxPhaseShift     float64 // Maximum phase shift (rad)

	// Bandwidth and loss
	BandwidthGHz      float64
	InsertionLossdB   float64

	// Crosstalk
	ThermalCrosstalk  float64 // Fraction of heat to neighbor
}

// DefaultMZIConfig returns typical MZI configuration
func DefaultMZIConfig() *MZIConfig {
	return &MZIConfig{
		ArmLengthMm:       0.5,
		PhaseShifterType:  "thermal",
		VpiLengthMmV:      10.0,
		ThermalEfficiency: 0.1, // rad/mW
		MaxPhaseShift:     2 * math.Pi,
		BandwidthGHz:      40,
		InsertionLossdB:   0.5,
		ThermalCrosstalk:  0.05,
	}
}

// MZI represents a single Mach-Zehnder interferometer
type MZI struct {
	Config      *MZIConfig
	Phase1      float64 // Internal phase (θ)
	Phase2      float64 // External phase (φ)

	// Computed transfer matrix
	TransferMatrix [2][2]complex128
}

// NewMZI creates a new MZI
func NewMZI(config *MZIConfig) *MZI {
	if config == nil {
		config = DefaultMZIConfig()
	}
	mzi := &MZI{Config: config}
	mzi.ComputeTransferMatrix()
	return mzi
}

// SetPhases sets the internal and external phase shifts
func (mzi *MZI) SetPhases(theta, phi float64) {
	mzi.Phase1 = theta
	mzi.Phase2 = phi
	mzi.ComputeTransferMatrix()
}

// ComputeTransferMatrix calculates the 2x2 unitary transfer matrix
func (mzi *MZI) ComputeTransferMatrix() {
	// MZI transfer matrix: T = R(φ) · BS · PS(θ) · BS
	// Where BS is 50:50 beam splitter, PS is phase shifter, R is rotation

	theta := mzi.Phase1
	phi := mzi.Phase2

	// Simplified Clements-style parameterization
	cosTheta := math.Cos(theta / 2)
	sinTheta := math.Sin(theta / 2)
	expPhi := cmplx.Exp(complex(0, phi))

	// Transfer matrix elements
	mzi.TransferMatrix[0][0] = complex(cosTheta, 0) * expPhi
	mzi.TransferMatrix[0][1] = complex(sinTheta, 0)
	mzi.TransferMatrix[1][0] = complex(-sinTheta, 0) * expPhi
	mzi.TransferMatrix[1][1] = complex(cosTheta, 0)

	// Apply insertion loss
	lossFactor := math.Pow(10, -mzi.Config.InsertionLossdB/20)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			mzi.TransferMatrix[i][j] *= complex(lossFactor, 0)
		}
	}
}

// Transform applies the MZI transformation to an input vector
func (mzi *MZI) Transform(input [2]complex128) [2]complex128 {
	var output [2]complex128
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			output[i] += mzi.TransferMatrix[i][j] * input[j]
		}
	}
	return output
}

// ============================================================================
// SECTION 3: Coherent Optical Neural Network
// ============================================================================

// CoherentONNConfig configures a coherent optical neural network
type CoherentONNConfig struct {
	// Network dimensions
	InputDim          int
	OutputDim         int
	HiddenDim         int

	// MZI mesh topology
	MeshType          string  // "rectangular", "triangular", "butterfly"
	NumMZILayers      int

	// Wavelength
	NumWavelengths    int
	WavelengthSpacing float64 // GHz

	// Performance targets
	TargetTOPS        float64 // Tera-ops per second
	TargetTOPSW       float64 // TOPS per watt
}

// DefaultCoherentONNConfig returns typical coherent ONN configuration
func DefaultCoherentONNConfig() *CoherentONNConfig {
	return &CoherentONNConfig{
		InputDim:          64,
		OutputDim:         64,
		HiddenDim:         64,
		MeshType:          "rectangular",
		NumMZILayers:      6,
		NumWavelengths:    8,
		WavelengthSpacing: 100,
		TargetTOPS:        1.28,
		TargetTOPSW:       160,
	}
}

// MZIMesh represents a mesh of MZIs for unitary transformation
type MZIMesh struct {
	Config      *CoherentONNConfig
	MZIs        [][]*MZI
	Dimension   int

	// Learned parameters
	ThetaAngles [][]float64
	PhiAngles   [][]float64
}

// NewMZIMesh creates a new MZI mesh
func NewMZIMesh(config *CoherentONNConfig) *MZIMesh {
	if config == nil {
		config = DefaultCoherentONNConfig()
	}

	dim := config.InputDim
	mesh := &MZIMesh{
		Config:    config,
		Dimension: dim,
	}

	// Create MZI layers based on mesh type
	switch config.MeshType {
	case "rectangular":
		// Clements architecture: dim layers, each with dim/2 MZIs
		mesh.MZIs = make([][]*MZI, dim)
		mesh.ThetaAngles = make([][]float64, dim)
		mesh.PhiAngles = make([][]float64, dim)

		for layer := 0; layer < dim; layer++ {
			numMZIs := dim / 2
			mesh.MZIs[layer] = make([]*MZI, numMZIs)
			mesh.ThetaAngles[layer] = make([]float64, numMZIs)
			mesh.PhiAngles[layer] = make([]float64, numMZIs)

			for m := 0; m < numMZIs; m++ {
				mesh.MZIs[layer][m] = NewMZI(DefaultMZIConfig())
				mesh.ThetaAngles[layer][m] = rand.Float64() * 2 * math.Pi
				mesh.PhiAngles[layer][m] = rand.Float64() * 2 * math.Pi
				mesh.MZIs[layer][m].SetPhases(mesh.ThetaAngles[layer][m], mesh.PhiAngles[layer][m])
			}
		}

	case "triangular":
		// Reck architecture
		numLayers := dim * (dim - 1) / 2
		mesh.MZIs = make([][]*MZI, numLayers)
		mesh.ThetaAngles = make([][]float64, numLayers)
		mesh.PhiAngles = make([][]float64, numLayers)

		for layer := 0; layer < numLayers; layer++ {
			mesh.MZIs[layer] = make([]*MZI, 1)
			mesh.ThetaAngles[layer] = make([]float64, 1)
			mesh.PhiAngles[layer] = make([]float64, 1)
			mesh.MZIs[layer][0] = NewMZI(DefaultMZIConfig())
		}
	}

	return mesh
}

// SetUnitaryMatrix programs the mesh to implement a target unitary matrix
func (mesh *MZIMesh) SetUnitaryMatrix(U [][]complex128) error {
	if len(U) != mesh.Dimension || len(U[0]) != mesh.Dimension {
		return fmt.Errorf("matrix dimension mismatch: expected %dx%d", mesh.Dimension, mesh.Dimension)
	}

	// Use Clements decomposition to find MZI parameters
	// This is a simplified placeholder - real implementation would use
	// the Clements or Reck decomposition algorithm
	for layer := range mesh.MZIs {
		for m := range mesh.MZIs[layer] {
			// Extract angles from matrix (simplified)
			i := layer
			j := m * 2
			if i < len(U) && j < len(U[0]) {
				magnitude := cmplx.Abs(U[i][j])
				phase := cmplx.Phase(U[i][j])

				theta := 2 * math.Acos(math.Min(1, math.Max(-1, magnitude)))
				phi := phase

				mesh.ThetaAngles[layer][m] = theta
				mesh.PhiAngles[layer][m] = phi
				mesh.MZIs[layer][m].SetPhases(theta, phi)
			}
		}
	}

	return nil
}

// Forward performs forward pass through the MZI mesh
func (mesh *MZIMesh) Forward(input []complex128) []complex128 {
	if len(input) != mesh.Dimension {
		// Pad or truncate
		padded := make([]complex128, mesh.Dimension)
		copy(padded, input)
		input = padded
	}

	current := make([]complex128, mesh.Dimension)
	copy(current, input)

	// Apply each MZI layer
	for layer := range mesh.MZIs {
		next := make([]complex128, mesh.Dimension)
		copy(next, current)

		offset := 0
		if layer%2 == 1 {
			offset = 1
		}

		for m, mzi := range mesh.MZIs[layer] {
			idx := m*2 + offset
			if idx+1 < mesh.Dimension {
				in := [2]complex128{current[idx], current[idx+1]}
				out := mzi.Transform(in)
				next[idx] = out[0]
				next[idx+1] = out[1]
			}
		}

		current = next
	}

	return current
}

// CoherentONN implements a coherent optical neural network
type CoherentONN struct {
	Config         *CoherentONNConfig

	// Optical components
	InputMesh      *MZIMesh    // Input unitary transformation
	DiagonalGains  []float64   // Programmable attenuators (Σ matrix)
	OutputMesh     *MZIMesh    // Output unitary transformation

	// WDM support
	WavelengthWeights [][]float64 // Per-wavelength weight banks

	// Performance tracking
	ComputeSpeedTOPS float64
	EnergyEffTOPSW   float64
}

// NewCoherentONN creates a new coherent optical neural network
func NewCoherentONN(config *CoherentONNConfig) *CoherentONN {
	if config == nil {
		config = DefaultCoherentONNConfig()
	}

	onn := &CoherentONN{
		Config:        config,
		InputMesh:     NewMZIMesh(config),
		OutputMesh:    NewMZIMesh(config),
		DiagonalGains: make([]float64, config.InputDim),
	}

	// Initialize diagonal gains (singular values)
	for i := range onn.DiagonalGains {
		onn.DiagonalGains[i] = rand.Float64()
	}

	// Initialize wavelength weights for WDM
	onn.WavelengthWeights = make([][]float64, config.NumWavelengths)
	for w := 0; w < config.NumWavelengths; w++ {
		onn.WavelengthWeights[w] = make([]float64, config.InputDim*config.OutputDim)
		for i := range onn.WavelengthWeights[w] {
			onn.WavelengthWeights[w][i] = rand.Float64()*2 - 1
		}
	}

	onn.ComputeSpeedTOPS = config.TargetTOPS
	onn.EnergyEffTOPSW = config.TargetTOPSW

	return onn
}

// MatrixVectorMultiply performs optical MVM
func (onn *CoherentONN) MatrixVectorMultiply(input []float64) []float64 {
	// Convert to complex
	complexInput := make([]complex128, len(input))
	for i, v := range input {
		complexInput[i] = complex(v, 0)
	}

	// U transformation
	afterU := onn.InputMesh.Forward(complexInput)

	// Diagonal (Σ) transformation
	afterSigma := make([]complex128, len(afterU))
	for i := range afterU {
		if i < len(onn.DiagonalGains) {
			afterSigma[i] = afterU[i] * complex(onn.DiagonalGains[i], 0)
		}
	}

	// V† transformation
	output := onn.OutputMesh.Forward(afterSigma)

	// Convert back to real (magnitude)
	result := make([]float64, len(output))
	for i, v := range output {
		result[i] = cmplx.Abs(v)
	}

	return result
}

// ============================================================================
// SECTION 4: In-Memory Training Fundamentals
// ============================================================================

// InMemoryTrainingConfig configures in-memory training
type InMemoryTrainingConfig struct {
	// Device parameters
	DeviceType        string  // "ReRAM", "PCM", "ECRAM", "FeFET"
	ConductanceLevels int     // Number of distinct states
	MinConductanceUS  float64 // Minimum conductance (µS)
	MaxConductanceUS  float64 // Maximum conductance (µS)

	// Update characteristics
	PulseAmplitudeV   float64 // Programming pulse amplitude
	PulseDurationNs   float64 // Programming pulse duration
	AsymmetryRatio    float64 // SET vs RESET asymmetry

	// Noise and variability
	CycleToVariation  float64 // Cycle-to-cycle variation
	DeviceToVariation float64 // Device-to-device variation
	ReadNoise         float64 // Read noise (fraction of signal)

	// Training algorithm
	Algorithm         string  // "SGD", "TikiTaka", "EP", "LocalLearning"
	LearningRate      float64
	BatchSize         int
}

// DefaultInMemoryTrainingConfig returns typical in-memory training configuration
func DefaultInMemoryTrainingConfig() *InMemoryTrainingConfig {
	return &InMemoryTrainingConfig{
		DeviceType:        "ReRAM",
		ConductanceLevels: 32,
		MinConductanceUS:  1.0,
		MaxConductanceUS:  10.0,
		PulseAmplitudeV:   2.0,
		PulseDurationNs:   100,
		AsymmetryRatio:    1.5,
		CycleToVariation:  0.05,
		DeviceToVariation: 0.1,
		ReadNoise:         0.02,
		Algorithm:         "TikiTaka",
		LearningRate:      0.01,
		BatchSize:         32,
	}
}

// MemristorDevice models a single memristive device for training
type MemristorDevice struct {
	Config           *InMemoryTrainingConfig
	Conductance      float64 // Current conductance (µS)
	SymmetryPoint    float64 // Conductance at symmetry point

	// History for analysis
	ConductanceHistory []float64
	UpdateCount        int
}

// NewMemristorDevice creates a new memristor device
func NewMemristorDevice(config *InMemoryTrainingConfig) *MemristorDevice {
	if config == nil {
		config = DefaultInMemoryTrainingConfig()
	}

	midConductance := (config.MinConductanceUS + config.MaxConductanceUS) / 2

	return &MemristorDevice{
		Config:             config,
		Conductance:        midConductance,
		SymmetryPoint:      midConductance,
		ConductanceHistory: make([]float64, 0),
	}
}

// ApplyPulse applies a programming pulse and updates conductance
func (m *MemristorDevice) ApplyPulse(positive bool, numPulses int) float64 {
	// Calculate conductance change per pulse
	conductanceRange := m.Config.MaxConductanceUS - m.Config.MinConductanceUS
	deltaG := conductanceRange / float64(m.Config.ConductanceLevels)

	// Apply asymmetry
	if positive {
		deltaG *= 1.0 // SET direction
	} else {
		deltaG *= -m.Config.AsymmetryRatio // RESET direction (often slower)
	}

	// Add variation
	variation := 1 + rand.NormFloat64()*m.Config.CycleToVariation
	deltaG *= variation

	// Update conductance
	for i := 0; i < numPulses; i++ {
		m.Conductance += deltaG
	}

	// Clamp to valid range
	if m.Conductance < m.Config.MinConductanceUS {
		m.Conductance = m.Config.MinConductanceUS
	}
	if m.Conductance > m.Config.MaxConductanceUS {
		m.Conductance = m.Config.MaxConductanceUS
	}

	m.ConductanceHistory = append(m.ConductanceHistory, m.Conductance)
	m.UpdateCount++

	return m.Conductance
}

// ReadConductance returns conductance with read noise
func (m *MemristorDevice) ReadConductance() float64 {
	noise := rand.NormFloat64() * m.Config.ReadNoise * m.Conductance
	return m.Conductance + noise
}

// GetWeight converts conductance to normalized weight
func (m *MemristorDevice) GetWeight() float64 {
	normalized := (m.Conductance - m.Config.MinConductanceUS) /
		(m.Config.MaxConductanceUS - m.Config.MinConductanceUS)
	return normalized*2 - 1 // Map to [-1, 1]
}

// ============================================================================
// SECTION 5: Tiki-Taka Training Algorithm
// ============================================================================

// TikiTakaConfig configures the Tiki-Taka algorithm
type TikiTakaConfig struct {
	// Matrix coupling
	GammaFactor       float64 // Mixing factor for A and C matrices

	// Update parameters
	UpdateRate        float64 // Base update rate
	MomentumFactor    float64 // Momentum for C matrix updates

	// Convergence
	MaxIterations     int
	ConvergenceThresh float64
}

// DefaultTikiTakaConfig returns default Tiki-Taka configuration
func DefaultTikiTakaConfig() *TikiTakaConfig {
	return &TikiTakaConfig{
		GammaFactor:       0.5,
		UpdateRate:        0.01,
		MomentumFactor:    0.9,
		MaxIterations:     1000,
		ConvergenceThresh: 1e-6,
	}
}

// TikiTakaTrainer implements the Tiki-Taka training algorithm
type TikiTakaTrainer struct {
	Config            *TikiTakaConfig
	DeviceConfig      *InMemoryTrainingConfig

	// Dual matrix representation: W = A + γC
	MatrixA           [][]*MemristorDevice // Primary matrix
	MatrixC           [][]*MemristorDevice // Auxiliary matrix

	// Dimensions
	Rows              int
	Cols              int

	// Momentum storage
	CMomentum         [][]float64

	// Training state
	Iteration         int
	Loss              float64
	Converged         bool
}

// NewTikiTakaTrainer creates a new Tiki-Taka trainer
func NewTikiTakaTrainer(rows, cols int, config *TikiTakaConfig, deviceConfig *InMemoryTrainingConfig) *TikiTakaTrainer {
	if config == nil {
		config = DefaultTikiTakaConfig()
	}
	if deviceConfig == nil {
		deviceConfig = DefaultInMemoryTrainingConfig()
	}

	trainer := &TikiTakaTrainer{
		Config:       config,
		DeviceConfig: deviceConfig,
		Rows:         rows,
		Cols:         cols,
		MatrixA:      make([][]*MemristorDevice, rows),
		MatrixC:      make([][]*MemristorDevice, rows),
		CMomentum:    make([][]float64, rows),
	}

	// Initialize device arrays
	for i := 0; i < rows; i++ {
		trainer.MatrixA[i] = make([]*MemristorDevice, cols)
		trainer.MatrixC[i] = make([]*MemristorDevice, cols)
		trainer.CMomentum[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			trainer.MatrixA[i][j] = NewMemristorDevice(deviceConfig)
			trainer.MatrixC[i][j] = NewMemristorDevice(deviceConfig)
		}
	}

	return trainer
}

// GetEffectiveWeight returns W = A + γC
func (t *TikiTakaTrainer) GetEffectiveWeight(i, j int) float64 {
	weightA := t.MatrixA[i][j].GetWeight()
	weightC := t.MatrixC[i][j].GetWeight()
	return weightA + t.Config.GammaFactor*weightC
}

// GetWeightMatrix returns the full effective weight matrix
func (t *TikiTakaTrainer) GetWeightMatrix() [][]float64 {
	weights := make([][]float64, t.Rows)
	for i := 0; i < t.Rows; i++ {
		weights[i] = make([]float64, t.Cols)
		for j := 0; j < t.Cols; j++ {
			weights[i][j] = t.GetEffectiveWeight(i, j)
		}
	}
	return weights
}

// UpdateStep performs one Tiki-Taka update step
func (t *TikiTakaTrainer) UpdateStep(gradients [][]float64) {
	for i := 0; i < t.Rows; i++ {
		for j := 0; j < t.Cols; j++ {
			grad := gradients[i][j]

			// Update C matrix with momentum
			t.CMomentum[i][j] = t.Config.MomentumFactor*t.CMomentum[i][j] -
				t.Config.UpdateRate*grad
			deltaC := t.CMomentum[i][j]

			// Apply pulses to C matrix
			numPulses := int(math.Abs(deltaC) * 10)
			if numPulses > 0 {
				t.MatrixC[i][j].ApplyPulse(deltaC > 0, numPulses)
			}

			// Update A matrix: A_new = A + γ(C_new - C_old)
			currentC := t.MatrixC[i][j].GetWeight()
			prevC := currentC - deltaC
			deltaA := t.Config.GammaFactor * (currentC - prevC)

			numPulsesA := int(math.Abs(deltaA) * 10)
			if numPulsesA > 0 {
				t.MatrixA[i][j].ApplyPulse(deltaA > 0, numPulsesA)
			}
		}
	}

	t.Iteration++
}

// ComputeForward performs forward pass using effective weights
func (t *TikiTakaTrainer) ComputeForward(input []float64) []float64 {
	output := make([]float64, t.Rows)

	for i := 0; i < t.Rows; i++ {
		sum := 0.0
		for j := 0; j < t.Cols && j < len(input); j++ {
			weight := t.GetEffectiveWeight(i, j)
			sum += weight * input[j]
		}
		output[i] = sum
	}

	return output
}

// ============================================================================
// SECTION 6: Equilibrium Propagation
// ============================================================================

// EquilibriumPropConfig configures equilibrium propagation
type EquilibriumPropConfig struct {
	// Network architecture
	LayerSizes        []int

	// Energy function parameters
	Beta              float64 // Clamping strength
	Epsilon           float64 // Learning rate

	// Dynamics
	NumFreePhaseSteps int     // Steps to reach free equilibrium
	NumClampedSteps   int     // Steps in clamped phase
	DynamicsRate      float64 // Rate of state updates

	// Activation
	ActivationType    string  // "sigmoid", "tanh", "softplus"
}

// DefaultEquilibriumPropConfig returns default EP configuration
func DefaultEquilibriumPropConfig() *EquilibriumPropConfig {
	return &EquilibriumPropConfig{
		LayerSizes:        []int{784, 500, 10},
		Beta:              0.5,
		Epsilon:           0.1,
		NumFreePhaseSteps: 20,
		NumClampedSteps:   4,
		DynamicsRate:      0.5,
		ActivationType:    "sigmoid",
	}
}

// EPLayer represents a layer in the EP network
type EPLayer struct {
	NumNeurons      int
	States          []float64 // Neuron states
	FreeStates      []float64 // States at free equilibrium
	ClampedStates   []float64 // States at clamped equilibrium
}

// EquilibriumPropNetwork implements equilibrium propagation
type EquilibriumPropNetwork struct {
	Config          *EquilibriumPropConfig
	Layers          []*EPLayer
	Weights         [][][]*MemristorDevice // Weights between layers
	Biases          [][]*MemristorDevice   // Bias terms

	// Training state
	LossHistory     []float64
}

// NewEquilibriumPropNetwork creates a new EP network
func NewEquilibriumPropNetwork(config *EquilibriumPropConfig, deviceConfig *InMemoryTrainingConfig) *EquilibriumPropNetwork {
	if config == nil {
		config = DefaultEquilibriumPropConfig()
	}
	if deviceConfig == nil {
		deviceConfig = DefaultInMemoryTrainingConfig()
	}

	numLayers := len(config.LayerSizes)

	net := &EquilibriumPropNetwork{
		Config:      config,
		Layers:      make([]*EPLayer, numLayers),
		Weights:     make([][][]*MemristorDevice, numLayers-1),
		Biases:      make([][]*MemristorDevice, numLayers),
		LossHistory: make([]float64, 0),
	}

	// Initialize layers
	for l := 0; l < numLayers; l++ {
		size := config.LayerSizes[l]
		net.Layers[l] = &EPLayer{
			NumNeurons:    size,
			States:        make([]float64, size),
			FreeStates:    make([]float64, size),
			ClampedStates: make([]float64, size),
		}

		// Initialize biases
		net.Biases[l] = make([]*MemristorDevice, size)
		for i := 0; i < size; i++ {
			net.Biases[l][i] = NewMemristorDevice(deviceConfig)
		}
	}

	// Initialize weights
	for l := 0; l < numLayers-1; l++ {
		rows := config.LayerSizes[l]
		cols := config.LayerSizes[l+1]

		net.Weights[l] = make([][]*MemristorDevice, rows)
		for i := 0; i < rows; i++ {
			net.Weights[l][i] = make([]*MemristorDevice, cols)
			for j := 0; j < cols; j++ {
				net.Weights[l][i][j] = NewMemristorDevice(deviceConfig)
			}
		}
	}

	return net
}

// activation applies the configured activation function
func (net *EquilibriumPropNetwork) activation(x float64) float64 {
	switch net.Config.ActivationType {
	case "sigmoid":
		return 1.0 / (1.0 + math.Exp(-x))
	case "tanh":
		return math.Tanh(x)
	case "softplus":
		return math.Log(1 + math.Exp(x))
	default:
		return 1.0 / (1.0 + math.Exp(-x))
	}
}

// activationDerivative computes derivative of activation
func (net *EquilibriumPropNetwork) activationDerivative(x float64) float64 {
	switch net.Config.ActivationType {
	case "sigmoid":
		s := net.activation(x)
		return s * (1 - s)
	case "tanh":
		t := math.Tanh(x)
		return 1 - t*t
	case "softplus":
		return 1.0 / (1.0 + math.Exp(-x))
	default:
		s := net.activation(x)
		return s * (1 - s)
	}
}

// ComputeEnergy computes the energy function
func (net *EquilibriumPropNetwork) ComputeEnergy() float64 {
	energy := 0.0

	// Sum over all neurons
	for l := 0; l < len(net.Layers); l++ {
		for i := 0; i < net.Layers[l].NumNeurons; i++ {
			s := net.Layers[l].States[i]
			// Primitive function of activation
			energy += s*s/2 - net.Biases[l][i].GetWeight()*s
		}
	}

	// Subtract interaction terms
	for l := 0; l < len(net.Weights); l++ {
		for i := 0; i < len(net.Weights[l]); i++ {
			for j := 0; j < len(net.Weights[l][i]); j++ {
				s_i := net.Layers[l].States[i]
				s_j := net.Layers[l+1].States[j]
				w := net.Weights[l][i][j].GetWeight()
				energy -= w * s_i * s_j
			}
		}
	}

	return energy
}

// RelaxToEquilibrium runs network dynamics to equilibrium
func (net *EquilibriumPropNetwork) RelaxToEquilibrium(numSteps int, clampInput, clampOutput bool, input, target []float64) {
	for step := 0; step < numSteps; step++ {
		// Update each layer (except clamped ones)
		for l := 0; l < len(net.Layers); l++ {
			// Skip clamped layers
			if l == 0 && clampInput {
				continue
			}
			if l == len(net.Layers)-1 && clampOutput {
				continue
			}

			for i := 0; i < net.Layers[l].NumNeurons; i++ {
				// Compute input from neighboring layers
				totalInput := net.Biases[l][i].GetWeight()

				// Input from previous layer
				if l > 0 {
					for j := 0; j < net.Layers[l-1].NumNeurons; j++ {
						w := net.Weights[l-1][j][i].GetWeight()
						totalInput += w * net.Layers[l-1].States[j]
					}
				}

				// Input from next layer
				if l < len(net.Layers)-1 {
					for j := 0; j < net.Layers[l+1].NumNeurons; j++ {
						w := net.Weights[l][i][j].GetWeight()
						totalInput += w * net.Layers[l+1].States[j]
					}
				}

				// Update state
				newState := net.activation(totalInput)
				net.Layers[l].States[i] = (1-net.Config.DynamicsRate)*net.Layers[l].States[i] +
					net.Config.DynamicsRate*newState
			}
		}

		// Apply clamping
		if clampInput && input != nil {
			copy(net.Layers[0].States, input)
		}
		if clampOutput && target != nil {
			for i := 0; i < len(target) && i < len(net.Layers[len(net.Layers)-1].States); i++ {
				outputLayer := net.Layers[len(net.Layers)-1]
				outputLayer.States[i] = (1-net.Config.Beta)*outputLayer.States[i] +
					net.Config.Beta*target[i]
			}
		}
	}
}

// Train performs one EP training step
func (net *EquilibriumPropNetwork) Train(input, target []float64) {
	// Set input
	copy(net.Layers[0].States, input)

	// Free phase: relax to equilibrium without output clamping
	net.RelaxToEquilibrium(net.Config.NumFreePhaseSteps, true, false, input, nil)

	// Store free equilibrium states
	for l := range net.Layers {
		copy(net.Layers[l].FreeStates, net.Layers[l].States)
	}

	// Weakly clamped phase: relax with output nudged toward target
	net.RelaxToEquilibrium(net.Config.NumClampedSteps, true, true, input, target)

	// Store clamped equilibrium states
	for l := range net.Layers {
		copy(net.Layers[l].ClampedStates, net.Layers[l].States)
	}

	// Update weights using contrastive Hebbian rule
	for l := 0; l < len(net.Weights); l++ {
		for i := 0; i < len(net.Weights[l]); i++ {
			for j := 0; j < len(net.Weights[l][i]); j++ {
				// ΔW ∝ (s_i^+ * s_j^+ - s_i^- * s_j^-) / β
				s_i_free := net.Layers[l].FreeStates[i]
				s_j_free := net.Layers[l+1].FreeStates[j]
				s_i_clamp := net.Layers[l].ClampedStates[i]
				s_j_clamp := net.Layers[l+1].ClampedStates[j]

				deltaW := (s_i_clamp*s_j_clamp - s_i_free*s_j_free) / net.Config.Beta
				deltaW *= net.Config.Epsilon

				// Apply to memristor
				numPulses := int(math.Abs(deltaW) * 100)
				if numPulses > 0 {
					net.Weights[l][i][j].ApplyPulse(deltaW > 0, numPulses)
				}
			}
		}
	}
}

// ============================================================================
// SECTION 7: Outer Product Weight Update
// ============================================================================

// OuterProductUpdateConfig configures parallel outer product updates
type OuterProductUpdateConfig struct {
	// Pulse parameters
	RowPulseAmplitude float64
	ColPulseAmplitude float64
	PulseDurationNs   float64

	// Timing
	UpdateParallelism int     // Number of simultaneous updates
	SettlingTimeNs    float64 // Time to settle after update
}

// DefaultOuterProductUpdateConfig returns default configuration
func DefaultOuterProductUpdateConfig() *OuterProductUpdateConfig {
	return &OuterProductUpdateConfig{
		RowPulseAmplitude: 0.5,
		ColPulseAmplitude: 0.5,
		PulseDurationNs:   50,
		UpdateParallelism: 256,
		SettlingTimeNs:    100,
	}
}

// CrossbarArray represents a memristor crossbar for in-memory training
type CrossbarArray struct {
	Config            *OuterProductUpdateConfig
	DeviceConfig      *InMemoryTrainingConfig
	Devices           [][]*MemristorDevice
	Rows              int
	Cols              int

	// Peripheral circuits
	RowDrivers        []float64 // Row voltage drivers
	ColDrivers        []float64 // Column voltage drivers
	SenseAmplifiers   []float64 // Column sense amplifiers

	// Statistics
	TotalUpdates      int64
	EnergyConsumedPJ  float64
}

// NewCrossbarArray creates a new crossbar array for training
func NewCrossbarArray(rows, cols int, config *OuterProductUpdateConfig, deviceConfig *InMemoryTrainingConfig) *CrossbarArray {
	if config == nil {
		config = DefaultOuterProductUpdateConfig()
	}
	if deviceConfig == nil {
		deviceConfig = DefaultInMemoryTrainingConfig()
	}

	array := &CrossbarArray{
		Config:          config,
		DeviceConfig:    deviceConfig,
		Rows:            rows,
		Cols:            cols,
		Devices:         make([][]*MemristorDevice, rows),
		RowDrivers:      make([]float64, rows),
		ColDrivers:      make([]float64, cols),
		SenseAmplifiers: make([]float64, cols),
	}

	for i := 0; i < rows; i++ {
		array.Devices[i] = make([]*MemristorDevice, cols)
		for j := 0; j < cols; j++ {
			array.Devices[i][j] = NewMemristorDevice(deviceConfig)
		}
	}

	return array
}

// ForwardPass performs matrix-vector multiplication
func (ca *CrossbarArray) ForwardPass(input []float64) []float64 {
	output := make([]float64, ca.Cols)

	// Apply input voltages to rows
	for i := 0; i < ca.Rows && i < len(input); i++ {
		ca.RowDrivers[i] = input[i]
	}

	// Compute output currents (Ohm's law + Kirchhoff's current law)
	for j := 0; j < ca.Cols; j++ {
		current := 0.0
		for i := 0; i < ca.Rows; i++ {
			conductance := ca.Devices[i][j].ReadConductance()
			current += ca.RowDrivers[i] * conductance
		}
		ca.SenseAmplifiers[j] = current
		output[j] = current
	}

	return output
}

// OuterProductUpdate performs parallel weight update via outer product
func (ca *CrossbarArray) OuterProductUpdate(rowActivations, colErrors []float64, learningRate float64) {
	// In hardware: ΔW[i][j] = η * x[i] * δ[j]
	// Implemented via simultaneous row and column pulses

	for i := 0; i < ca.Rows && i < len(rowActivations); i++ {
		for j := 0; j < ca.Cols && j < len(colErrors); j++ {
			// Compute desired weight change
			deltaW := learningRate * rowActivations[i] * colErrors[j]

			// Convert to pulse count
			numPulses := int(math.Abs(deltaW) * 100)

			if numPulses > 0 {
				ca.Devices[i][j].ApplyPulse(deltaW > 0, numPulses)
				ca.TotalUpdates++

				// Energy: V^2 * G * t
				voltage := ca.Config.RowPulseAmplitude + ca.Config.ColPulseAmplitude
				energy := voltage * voltage * ca.Devices[i][j].Conductance *
					ca.Config.PulseDurationNs * 1e-9 * float64(numPulses)
				ca.EnergyConsumedPJ += energy * 1e12
			}
		}
	}
}

// GetWeightMatrix extracts the weight matrix
func (ca *CrossbarArray) GetWeightMatrix() [][]float64 {
	weights := make([][]float64, ca.Rows)
	for i := 0; i < ca.Rows; i++ {
		weights[i] = make([]float64, ca.Cols)
		for j := 0; j < ca.Cols; j++ {
			weights[i][j] = ca.Devices[i][j].GetWeight()
		}
	}
	return weights
}

// ============================================================================
// SECTION 8: Integrated Optical-Electronic Training System
// ============================================================================

// HybridTrainingSystem combines optical inference with in-memory training
type HybridTrainingSystem struct {
	// Optical inference path
	OpticalProcessor  *CoherentONN

	// Electronic training path
	TrainingArray     *CrossbarArray
	TikiTakaTrainer   *TikiTakaTrainer

	// Synchronization
	WeightSyncInterval int
	LastSyncIteration  int

	// Performance metrics
	InferenceTOPS     float64
	TrainingTOPS      float64
	TotalPowerW       float64
}

// NewHybridTrainingSystem creates a hybrid optical-electronic system
func NewHybridTrainingSystem(inputDim, outputDim int) *HybridTrainingSystem {
	onnConfig := DefaultCoherentONNConfig()
	onnConfig.InputDim = inputDim
	onnConfig.OutputDim = outputDim

	system := &HybridTrainingSystem{
		OpticalProcessor:   NewCoherentONN(onnConfig),
		TrainingArray:      NewCrossbarArray(outputDim, inputDim, nil, nil),
		TikiTakaTrainer:    NewTikiTakaTrainer(outputDim, inputDim, nil, nil),
		WeightSyncInterval: 100,
		InferenceTOPS:      1.28,
		TrainingTOPS:       0.1,
		TotalPowerW:        10.0,
	}

	return system
}

// TrainStep performs one training step
func (h *HybridTrainingSystem) TrainStep(input, target []float64) []float64 {
	// Forward pass through optical processor
	output := h.OpticalProcessor.MatrixVectorMultiply(input)

	// Compute error
	errors := make([]float64, len(output))
	for i := range output {
		if i < len(target) {
			errors[i] = target[i] - output[i]
		}
	}

	// Backward pass and weight update in electronic crossbar
	h.TrainingArray.OuterProductUpdate(input, errors, 0.01)

	// Periodically sync weights to optical processor
	if h.TikiTakaTrainer.Iteration-h.LastSyncIteration >= h.WeightSyncInterval {
		h.SyncWeights()
		h.LastSyncIteration = h.TikiTakaTrainer.Iteration
	}

	h.TikiTakaTrainer.Iteration++

	return output
}

// SyncWeights transfers learned weights to optical processor
func (h *HybridTrainingSystem) SyncWeights() {
	weights := h.TrainingArray.GetWeightMatrix()

	// Update optical processor diagonal gains (simplified SVD approach)
	for i := 0; i < len(h.OpticalProcessor.DiagonalGains); i++ {
		sum := 0.0
		for j := 0; j < len(weights[0]) && i < len(weights); j++ {
			sum += weights[i][j] * weights[i][j]
		}
		h.OpticalProcessor.DiagonalGains[i] = math.Sqrt(sum)
	}
}

// ============================================================================
// SECTION 9: Serialization and Export
// ============================================================================

// TrainingSnapshot captures training state
type TrainingSnapshot struct {
	Iteration          int                    `json:"iteration"`
	Algorithm          string                 `json:"algorithm"`
	Loss               float64                `json:"loss"`
	WeightStats        map[string]float64     `json:"weight_stats"`
	DeviceStats        map[string]float64     `json:"device_stats"`
	PerformanceMetrics map[string]float64     `json:"performance_metrics"`
}

// ExportSnapshot exports training state to JSON
func ExportTrainingSnapshot(trainer *TikiTakaTrainer, writer io.Writer) error {
	snapshot := TrainingSnapshot{
		Iteration:          trainer.Iteration,
		Algorithm:          "TikiTaka",
		Loss:               trainer.Loss,
		WeightStats:        make(map[string]float64),
		DeviceStats:        make(map[string]float64),
		PerformanceMetrics: make(map[string]float64),
	}

	// Compute weight statistics
	weights := trainer.GetWeightMatrix()
	sum, sumSq, minW, maxW := 0.0, 0.0, weights[0][0], weights[0][0]
	count := 0

	for i := range weights {
		for j := range weights[i] {
			w := weights[i][j]
			sum += w
			sumSq += w * w
			if w < minW {
				minW = w
			}
			if w > maxW {
				maxW = w
			}
			count++
		}
	}

	mean := sum / float64(count)
	variance := sumSq/float64(count) - mean*mean

	snapshot.WeightStats["mean"] = mean
	snapshot.WeightStats["std"] = math.Sqrt(variance)
	snapshot.WeightStats["min"] = minW
	snapshot.WeightStats["max"] = maxW

	// Device statistics
	totalUpdates := 0
	for i := range trainer.MatrixA {
		for j := range trainer.MatrixA[i] {
			totalUpdates += trainer.MatrixA[i][j].UpdateCount
			totalUpdates += trainer.MatrixC[i][j].UpdateCount
		}
	}
	snapshot.DeviceStats["total_updates"] = float64(totalUpdates)
	snapshot.DeviceStats["avg_updates_per_device"] = float64(totalUpdates) / float64(2*trainer.Rows*trainer.Cols)

	return json.NewEncoder(writer).Encode(snapshot)
}

// ============================================================================
// SECTION 10: ASCII Diagrams
// ============================================================================

/*
Silicon Photonics MZI Mesh for Neural Networks:
===============================================

┌─────────────────────────────────────────────────────────────────────────────┐
│                    MZI-Based Optical Neural Network                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Input      ┌─────┐     ┌─────┐     ┌─────┐     ┌─────┐                    │
│  x₁  ──────▶│ MZI │─────│ MZI │─────│ MZI │─────│ MZI │──────▶ y₁         │
│             │ θ₁₁ │     │ θ₁₂ │     │ θ₁₃ │     │ θ₁₄ │                    │
│  x₂  ──────▶└──┬──┘     └──┬──┘     └──┬──┘     └──┬──┘──────▶ y₂         │
│                │  ╲    ╱   │  ╲    ╱   │  ╲    ╱   │                        │
│  x₃  ──────▶┌──┴──┐ ╲╱  ┌──┴──┐ ╲╱  ┌──┴──┐ ╲╱  ┌──┴──┐──────▶ y₃         │
│             │ MZI │ ╱╲  │ MZI │ ╱╲  │ MZI │ ╱╲  │ MZI │                    │
│  x₄  ──────▶│ θ₂₁ │╱  ╲ │ θ₂₂ │╱  ╲ │ θ₂₃ │╱  ╲ │ θ₂₄ │──────▶ y₄         │
│             └─────┘     └─────┘     └─────┘     └─────┘                    │
│                                                                             │
│  MZI = Mach-Zehnder Interferometer (programmable 2×2 unitary)              │
│  θ = Phase shift angle (thermal or electro-optic)                          │
│                                                                             │
│  Performance: 1.28 TOPS at 160 TOPS/W (Taichi, Science 2024)               │
│  Energy: ~0.66 photons per multiplication (2.5×10⁻¹⁹ J optical)            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


WDM Microring Weight Bank:
==========================

┌────────────────────────────────────────────────────────────────────────┐
│                                                                        │
│     λ₁ λ₂ λ₃ λ₄ λ₅ λ₆ λ₇ λ₈    (8 wavelength channels)               │
│      │  │  │  │  │  │  │  │                                           │
│      ▼  ▼  ▼  ▼  ▼  ▼  ▼  ▼                                           │
│    ┌──────────────────────────┐                                        │
│    │   ○   ○   ○   ○   ○   ○  │ ◀── Microring array                   │
│    │   MRR MRR MRR MRR MRR MRR│     (each tuned to different λ)       │
│    └──────────────────────────┘                                        │
│                 │                                                      │
│                 ▼                                                      │
│         Weighted output                                                │
│                                                                        │
│  Each MRR acts as wavelength-selective weight:                         │
│  • On-resonance: high transmission (weight ≈ 1)                        │
│  • Off-resonance: low transmission (weight ≈ 0)                        │
│  • Thermal tuning: continuous weight control                           │
│                                                                        │
│  Bandwidth: 1 Tbps (5×200 Gbps), Energy: <10 fJ/bit                   │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘


In-Memory Training with Outer Product:
======================================

┌─────────────────────────────────────────────────────────────────────────────┐
│                    Crossbar Outer Product Update                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│        Activation vector x          Error vector δ                          │
│        ┌─┬─┬─┬─┐                   ┌─┬─┬─┬─┐                               │
│        │x│x│x│x│                   │δ│δ│δ│δ│                               │
│        │₁│₂│₃│₄│                   │₁│₂│₃│₄│                               │
│        └┬┴┬┴┬┴┬┘                   └┬┴┬┴┬┴┬┘                               │
│         │ │ │ │                     │ │ │ │                                 │
│    ─────┼─┼─┼─┼─────           ────┼─┼─┼─┼────                             │
│    │    ▼ ▼ ▼ ▼    │           │   ▼ ▼ ▼ ▼   │                             │
│    │  ┌─┬─┬─┬─┐    │           │ ┌─┬─┬─┬─┐   │                             │
│    │  │●│●│●│●│◀───┤ Row       │ │ │ │ │ │   │                             │
│    │  ├─┼─┼─┼─┤    │ Pulse     │ ├─┼─┼─┼─┤   │                             │
│    │  │●│●│●│●│◀───┤           │ │ │ │ │ │   │                             │
│    │  ├─┼─┼─┼─┤    │           │ ├─┼─┼─┼─┤   │                             │
│    │  │●│●│●│●│◀───┤           │ │ │ │ │ │   │                             │
│    │  └─┴─┴─┴─┘    │           │ └─┴─┴─┴─┘   │                             │
│    │       ▲       │           │      ▲      │                             │
│    │       │       │           │      │      │                             │
│    │   Col Pulse   │           │  ΔW = η·x·δᵀ │                            │
│    └───────────────┘           └─────────────┘                             │
│                                                                             │
│  Parallel update: O(1) time complexity for entire matrix!                  │
│  ΔW[i][j] = learning_rate × x[i] × δ[j]                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


Tiki-Taka Dual Matrix Algorithm:
================================

┌────────────────────────────────────────────────────────────────────────┐
│                    Tiki-Taka Weight Representation                      │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│     Effective Weight: W = A + γC                                       │
│                                                                        │
│     ┌─────────┐         ┌─────────┐                                   │
│     │ Matrix A│    +γ×  │ Matrix C│   =   Effective W                 │
│     │ (slow)  │         │ (fast)  │                                   │
│     └─────────┘         └─────────┘                                   │
│          │                   │                                         │
│          │                   │                                         │
│          ▼                   ▼                                         │
│     Handles              Captures                                      │
│     asymmetry            gradients                                     │
│                                                                        │
│  Algorithm:                                                            │
│  1. C ← C - η·∇L        (gradient update to C)                        │
│  2. A ← A + γ·ΔC        (transfer change to A)                        │
│                                                                        │
│  Benefit: Works with asymmetric devices (SET ≠ RESET)                 │
│  Demonstrated: 30+ states, 5% skew from center                         │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘


Performance Summary:
====================

┌────────────────────────────────────────────────────────────────────────┐
│                    Optical vs Electronic Comparison                     │
├────────────────────┬───────────────┬───────────────┬──────────────────┤
│ Metric             │ Electronic    │ Optical       │ Hybrid           │
│                    │ (GPU)         │ (Photonic)    │ (Opt+Mem)        │
├────────────────────┼───────────────┼───────────────┼──────────────────┤
│ Speed              │ ~100 TOPS     │ 1.28 TOPS     │ ~10 TOPS         │
│ Energy Eff.        │ ~1 TOPS/W     │ 160 TOPS/W    │ ~50 TOPS/W       │
│ Training Support   │ Native        │ Limited       │ In-memory        │
│ Precision          │ FP32/16/8     │ Analog        │ Analog + Digital │
│ Scalability        │ Limited by    │ WDM parallel  │ Best of both     │
│                    │ memory BW     │ (100+ λ)      │                  │
└────────────────────┴───────────────┴───────────────┴──────────────────┘

Key Numbers:
• Taichi photonic chiplet: 160 TOPS/W (Science 2024)
• LiNbO₃ tensor core: 120 GOPS with in-situ training
• WDM microring: 1 Tbps (5×200 Gbps), <10 fJ/bit
• Coherent processor: 1.28 TOPS complex MVM
• CMO/HfOx ReRAM: All-in-one training + inference (IBM 2025)
• Optical efficiency: 0.66 photons/multiplication (99% accuracy)

*/
