// cryo_reram_cim.go - Cryogenic CIM Operation and ReRAM Variability Simulation
// IronLattice Visualization Project - Iteration 122
//
// This module implements simulation models for:
// 1. Cryogenic RRAM/memristor operation (4K, 1.4K temperatures)
// 2. Superconducting memristors (CA-SQUID based)
// 3. Magnetic topological memristors for quantum-classical interfaces
// 4. ReRAM variability modeling (D2D, C2C variations)
// 5. Variation-aware training compensation (CTSF, VACTSF)
// 6. Bayesian neural networks for variation tolerance
//
// Research basis:
// - Nature Materials 2024: Magnetic topological memristors
// - C2RAM: Cryogenic Capacitorless RAM for quantum computing
// - HfO₂ RRAM at 1.4K with 4 resistance levels
// - Chemical Reviews 2025: RRAM applications and requirements
// - CTSF/VACTSF compensation training frameworks

package layers

import (
	"encoding/json"
	"math"
	"math/rand"
	"sync"
)

// ============================================================================
// CRYOGENIC RRAM OPERATION
// ============================================================================

// CryoRRAMConfig defines parameters for cryogenic RRAM operation
type CryoRRAMConfig struct {
	// Array dimensions
	Rows    int `json:"rows"`
	Cols    int `json:"cols"`

	// Temperature settings (Kelvin)
	Temperature     float64 `json:"temperature"`      // Operating temperature (4K, 1.4K typical)
	RoomTemperature float64 `json:"room_temperature"` // Reference: 300K

	// Device physics
	ActivationEnergy float64 `json:"activation_energy"` // Ea in eV
	AttemptFrequency float64 `json:"attempt_frequency"` // f0 in Hz
	FilamentRadius   float64 `json:"filament_radius"`   // Conductive filament radius (nm)

	// Resistance states
	HRS          float64 `json:"hrs"`            // High resistance state (Ohms)
	LRS          float64 `json:"lrs"`            // Low resistance state (Ohms)
	NumLevels    int     `json:"num_levels"`     // Multi-level states (4 typical at cryo)

	// Cryogenic-specific parameters
	TunnelBarrier     float64 `json:"tunnel_barrier"`      // Barrier height (eV)
	QuantumCorrection bool    `json:"quantum_correction"`  // Enable quantum effects
	PhononFreeze      float64 `json:"phonon_freeze_temp"`  // Temperature below which phonons freeze
}

// CryoRRAMCell represents a single cryogenic RRAM cell
type CryoRRAMCell struct {
	Resistance      float64 `json:"resistance"`
	State           int     `json:"state"`           // Discrete level (0 to NumLevels-1)
	FilamentGap     float64 `json:"filament_gap"`    // Gap in conductive filament (nm)
	OxygenVacancies float64 `json:"oxygen_vacancies"` // Vo concentration
	Temperature     float64 `json:"temperature"`
	CycleCount      int     `json:"cycle_count"`
}

// CryoRRAM implements cryogenic RRAM array simulation
type CryoRRAM struct {
	Config *CryoRRAMConfig
	Cells  [][]*CryoRRAMCell
	mu     sync.RWMutex
}

// NewCryoRRAM creates a new cryogenic RRAM array
func NewCryoRRAM(config *CryoRRAMConfig) *CryoRRAM {
	if config.RoomTemperature == 0 {
		config.RoomTemperature = 300.0
	}
	if config.ActivationEnergy == 0 {
		config.ActivationEnergy = 0.7 // Typical for HfO2
	}
	if config.AttemptFrequency == 0 {
		config.AttemptFrequency = 1e12 // THz range
	}
	if config.NumLevels == 0 {
		config.NumLevels = 4 // 4 levels demonstrated at 1.4K
	}
	if config.PhononFreeze == 0 {
		config.PhononFreeze = 50.0 // Phonons freeze below ~50K
	}

	cells := make([][]*CryoRRAMCell, config.Rows)
	for i := range cells {
		cells[i] = make([]*CryoRRAMCell, config.Cols)
		for j := range cells[i] {
			cells[i][j] = &CryoRRAMCell{
				Resistance:      config.HRS,
				State:           0,
				FilamentGap:     5.0, // Initial gap in nm
				OxygenVacancies: 0.1,
				Temperature:     config.Temperature,
				CycleCount:      0,
			}
		}
	}

	return &CryoRRAM{
		Config: config,
		Cells:  cells,
	}
}

// ArrheniusRate calculates temperature-dependent switching rate
func (c *CryoRRAM) ArrheniusRate(temperature float64) float64 {
	// Arrhenius equation: rate = f0 * exp(-Ea / kT)
	// kB in eV/K
	kB := 8.617e-5

	// At cryogenic temperatures, rate drops dramatically
	if temperature < 1.0 {
		temperature = 1.0 // Prevent division by zero
	}

	rate := c.Config.AttemptFrequency * math.Exp(-c.Config.ActivationEnergy/(kB*temperature))
	return rate
}

// QuantumTunnelingProbability calculates tunneling through filament gap
func (c *CryoRRAM) QuantumTunnelingProbability(cell *CryoRRAMCell) float64 {
	// WKB approximation for tunneling probability
	// T = exp(-2 * sqrt(2m*phi) * d / hbar)

	if !c.Config.QuantumCorrection {
		return 0.0
	}

	// Constants
	m_e := 9.109e-31    // Electron mass (kg)
	hbar := 1.055e-34   // Reduced Planck constant (J*s)
	eV := 1.602e-19     // eV to Joules
	nm := 1e-9          // nm to m

	phi := c.Config.TunnelBarrier * eV // Barrier height in J
	d := cell.FilamentGap * nm         // Gap in m

	kappa := math.Sqrt(2 * m_e * phi) / hbar
	probability := math.Exp(-2 * kappa * d)

	return probability
}

// CryogenicSwitching performs SET/RESET at cryogenic temperature
func (c *CryoRRAM) CryogenicSwitching(row, col int, targetState int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	cell := c.Cells[row][col]

	// Calculate thermal switching rate
	thermalRate := c.ArrheniusRate(cell.Temperature)

	// At cryogenic temps, thermal activation is suppressed
	// Switching relies more on electric field and quantum tunneling
	tunnelingProb := c.QuantumTunnelingProbability(cell)

	// Effective switching probability
	roomRate := c.ArrheniusRate(c.Config.RoomTemperature)
	thermalFactor := thermalRate / roomRate

	// Combined probability (field-assisted + quantum)
	effectiveProb := thermalFactor + tunnelingProb*(1-thermalFactor)

	// Stochastic switching
	if rand.Float64() < effectiveProb {
		// Calculate new resistance based on target state
		levelStep := (c.Config.HRS - c.Config.LRS) / float64(c.Config.NumLevels-1)
		cell.State = targetState
		cell.Resistance = c.Config.LRS + float64(c.Config.NumLevels-1-targetState)*levelStep

		// Update filament gap
		cell.FilamentGap = 1.0 + float64(c.Config.NumLevels-1-targetState)*1.0
		cell.CycleCount++
		return true
	}

	return false
}

// ReadCell reads resistance at cryogenic temperature
func (c *CryoRRAM) ReadCell(row, col int) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cell := c.Cells[row][col]

	// At cryogenic temperatures, resistance may change due to:
	// 1. Reduced thermal noise
	// 2. Carrier freeze-out
	// 3. Quantum effects

	baseResistance := cell.Resistance

	// Temperature coefficient (negative for semiconductors)
	tempRatio := cell.Temperature / c.Config.RoomTemperature

	// Carrier freeze-out effect (increases resistance at low T)
	if cell.Temperature < c.Config.PhononFreeze {
		freezeoutFactor := 1.0 + 0.1*(c.Config.PhononFreeze-cell.Temperature)/c.Config.PhononFreeze
		baseResistance *= freezeoutFactor
	}

	// Quantum correction for tunneling
	if c.Config.QuantumCorrection {
		tunnelingProb := c.QuantumTunnelingProbability(cell)
		// Tunneling reduces effective resistance
		baseResistance *= (1.0 - 0.5*tunnelingProb)
	}

	return baseResistance
}

// ============================================================================
// SUPERCONDUCTING MEMRISTOR (CA-SQUID)
// ============================================================================

// CASQUIDConfig defines parameters for superconducting memristor
type CASQUIDConfig struct {
	// Critical current
	Ic0           float64 `json:"ic0"`            // Base critical current (µA)
	IcModulation  float64 `json:"ic_modulation"`  // Flux modulation depth

	// SQUID parameters
	LoopInductance float64 `json:"loop_inductance"` // Loop inductance (pH)
	JJCapacitance  float64 `json:"jj_capacitance"`  // Junction capacitance (fF)

	// Flux states
	NumFluxStates int     `json:"num_flux_states"` // Number of programmable states
	FluxQuantum   float64 `json:"flux_quantum"`    // Φ0 = h/2e ≈ 2.07 mV·ps

	// Temperature
	CriticalTemp  float64 `json:"critical_temp"`   // Tc in Kelvin
	OperatingTemp float64 `json:"operating_temp"`  // T_op in Kelvin
}

// CASQUIDCell represents a single CA-SQUID memristor
type CASQUIDCell struct {
	FluxState      int     `json:"flux_state"`      // Trapped flux quanta
	CriticalCurrent float64 `json:"critical_current"` // Current Ic
	Phase          float64 `json:"phase"`           // Josephson phase
	IsSuperconducting bool  `json:"is_superconducting"`
}

// CASQUID implements Controllable-Asymmetric SQUID memristor array
type CASQUID struct {
	Config *CASQUIDConfig
	Cells  [][]*CASQUIDCell
	Rows   int
	Cols   int
	mu     sync.RWMutex
}

// NewCASQUID creates a new CA-SQUID array
func NewCASQUID(rows, cols int, config *CASQUIDConfig) *CASQUID {
	if config.FluxQuantum == 0 {
		config.FluxQuantum = 2.067e-15 // Wb (magnetic flux quantum)
	}
	if config.NumFluxStates == 0 {
		config.NumFluxStates = 16 // High precision states
	}
	if config.CriticalTemp == 0 {
		config.CriticalTemp = 9.2 // Niobium Tc
	}

	cells := make([][]*CASQUIDCell, rows)
	for i := range cells {
		cells[i] = make([]*CASQUIDCell, cols)
		for j := range cells[i] {
			cells[i][j] = &CASQUIDCell{
				FluxState:         0,
				CriticalCurrent:   config.Ic0,
				Phase:             0,
				IsSuperconducting: config.OperatingTemp < config.CriticalTemp,
			}
		}
	}

	return &CASQUID{
		Config: config,
		Cells:  cells,
		Rows:   rows,
		Cols:   cols,
	}
}

// ProgramFluxState sets the trapped flux quanta
func (s *CASQUID) ProgramFluxState(row, col int, state int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state < 0 || state >= s.Config.NumFluxStates {
		return
	}

	cell := s.Cells[row][col]
	cell.FluxState = state

	// Critical current modulation by flux
	// Ic(Φ) = Ic0 * |cos(π * Φ/Φ0)|
	fluxNormalized := float64(state) / float64(s.Config.NumFluxStates)
	cell.CriticalCurrent = s.Config.Ic0 * math.Abs(math.Cos(math.Pi*fluxNormalized*s.Config.IcModulation))
}

// GetConductance returns the effective conductance for CIM operation
func (s *CASQUID) GetConductance(row, col int) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cell := s.Cells[row][col]

	if !cell.IsSuperconducting {
		return 0 // Normal state - high resistance
	}

	// Conductance proportional to critical current
	// Normalized to [0, 1] range
	return cell.CriticalCurrent / s.Config.Ic0
}

// ============================================================================
// MAGNETIC TOPOLOGICAL MEMRISTOR
// ============================================================================

// MagneticTopoConfig defines parameters for magnetic topological memristors
type MagneticTopoConfig struct {
	// Material: typically Cr2Te3 or similar
	CurieTemp        float64 `json:"curie_temp"`        // Ferromagnetic transition (K)
	ExchangeEnergy   float64 `json:"exchange_energy"`   // J in meV
	AnisotropyEnergy float64 `json:"anisotropy_energy"` // K in meV

	// Domain wall properties
	DomainWallWidth  float64 `json:"domain_wall_width"` // nm
	DomainWallEnergy float64 `json:"domain_wall_energy"` // mJ/m²

	// Topological properties
	SkyrmionRadius   float64 `json:"skyrmion_radius"`   // nm
	DMIStrength      float64 `json:"dmi_strength"`      // Dzyaloshinskii-Moriya interaction (mJ/m²)

	// Operating conditions
	OperatingTemp    float64 `json:"operating_temp"`    // K
	MagneticField    float64 `json:"magnetic_field"`    // Applied field (T)

	// States
	NumStates        int     `json:"num_states"`
}

// MagneticTopoCell represents a magnetic topological memristor cell
type MagneticTopoCell struct {
	Magnetization    [3]float64 `json:"magnetization"`     // mx, my, mz components
	SkyrmionNumber   int        `json:"skyrmion_number"`   // Topological charge
	DomainWallPos    float64    `json:"domain_wall_pos"`   // Position along nanowire
	Resistance       float64    `json:"resistance"`
	State            int        `json:"state"`
}

// MagneticTopoMemristor implements magnetic topological memristor array
type MagneticTopoMemristor struct {
	Config *MagneticTopoConfig
	Cells  [][]*MagneticTopoCell
	Rows   int
	Cols   int
	mu     sync.RWMutex
}

// NewMagneticTopoMemristor creates a new magnetic topological memristor array
func NewMagneticTopoMemristor(rows, cols int, config *MagneticTopoConfig) *MagneticTopoMemristor {
	if config.CurieTemp == 0 {
		config.CurieTemp = 180.0 // Cr2Te3 typical
	}
	if config.NumStates == 0 {
		config.NumStates = 8
	}

	cells := make([][]*MagneticTopoCell, rows)
	for i := range cells {
		cells[i] = make([]*MagneticTopoCell, cols)
		for j := range cells[i] {
			cells[i][j] = &MagneticTopoCell{
				Magnetization:  [3]float64{0, 0, 1}, // Initially saturated +z
				SkyrmionNumber: 0,
				DomainWallPos:  0,
				Resistance:     1000.0, // Base resistance
				State:          0,
			}
		}
	}

	return &MagneticTopoMemristor{
		Config: config,
		Cells:  cells,
		Rows:   rows,
		Cols:   cols,
	}
}

// CreateSkyrmion nucleates a skyrmion in the cell
func (m *MagneticTopoMemristor) CreateSkyrmion(row, col int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cell := m.Cells[row][col]

	// Skyrmion creation requires sufficient DMI and applied field
	// Temperature must be below Curie temperature
	if m.Config.OperatingTemp >= m.Config.CurieTemp {
		return // Paramagnetic state
	}

	cell.SkyrmionNumber++
	cell.State = cell.SkyrmionNumber % m.Config.NumStates

	// Topological Hall effect changes resistance
	// Each skyrmion adds a resistance contribution
	cell.Resistance = 1000.0 * (1.0 + 0.1*float64(cell.SkyrmionNumber))
}

// MoveDomainWall shifts domain wall position using spin-orbit torque
func (m *MagneticTopoMemristor) MoveDomainWall(row, col int, currentDensity float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cell := m.Cells[row][col]

	// Domain wall velocity proportional to current density
	// v = (γ * hbar / 2e * Ms) * J * η
	// Simplified: v ∝ J

	velocity := currentDensity * 100.0 // nm/ns per A/m² (simplified)
	cell.DomainWallPos += velocity

	// Clamp to track length (assume 100 nm track)
	if cell.DomainWallPos < 0 {
		cell.DomainWallPos = 0
	} else if cell.DomainWallPos > 100 {
		cell.DomainWallPos = 100
	}

	// Update state based on position
	cell.State = int(cell.DomainWallPos / 100.0 * float64(m.Config.NumStates-1))

	// Resistance depends on domain wall position (AMR/GMR effect)
	cell.Resistance = 1000.0 * (1.0 + 0.2*cell.DomainWallPos/100.0)
}

// GetConductance returns normalized conductance for CIM
func (m *MagneticTopoMemristor) GetConductance(row, col int) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cell := m.Cells[row][col]

	// Conductance inversely proportional to resistance
	// Normalized to [0, 1]
	maxR := 1200.0 // Maximum expected resistance
	minR := 1000.0 // Minimum resistance

	conductance := (maxR - cell.Resistance) / (maxR - minR)
	return math.Max(0, math.Min(1, conductance))
}

// ============================================================================
// RERAM VARIABILITY MODELING
// ============================================================================

// VariabilityConfig defines ReRAM variation parameters
type VariabilityConfig struct {
	// Device-to-device (D2D) variations - fixed per device
	D2D_HRS_Sigma    float64 `json:"d2d_hrs_sigma"`    // σ of HRS distribution
	D2D_LRS_Sigma    float64 `json:"d2d_lrs_sigma"`    // σ of LRS distribution
	D2D_Vset_Sigma   float64 `json:"d2d_vset_sigma"`   // σ of SET voltage
	D2D_Vreset_Sigma float64 `json:"d2d_vreset_sigma"` // σ of RESET voltage

	// Cycle-to-cycle (C2C) variations - changes each cycle
	C2C_HRS_Sigma    float64 `json:"c2c_hrs_sigma"`
	C2C_LRS_Sigma    float64 `json:"c2c_lrs_sigma"`

	// Temporal variations
	RTN_Amplitude    float64 `json:"rtn_amplitude"`    // Random telegraph noise
	RTN_Frequency    float64 `json:"rtn_frequency"`    // RTN switching rate

	// Read disturb
	ReadDisturbProb  float64 `json:"read_disturb_prob"` // Probability per read

	// Endurance degradation
	EnduranceCycles  int     `json:"endurance_cycles"`  // Cycles before degradation
	DegradationRate  float64 `json:"degradation_rate"`  // Resistance drift per cycle
}

// VariableRRAMCell represents an RRAM cell with variability
type VariableRRAMCell struct {
	// Nominal values
	NominalHRS float64 `json:"nominal_hrs"`
	NominalLRS float64 `json:"nominal_lrs"`

	// D2D offsets (fixed at fabrication)
	D2D_HRS_Offset float64 `json:"d2d_hrs_offset"`
	D2D_LRS_Offset float64 `json:"d2d_lrs_offset"`
	D2D_Vset_Offset float64 `json:"d2d_vset_offset"`
	D2D_Vreset_Offset float64 `json:"d2d_vreset_offset"`

	// Current state
	CurrentResistance float64 `json:"current_resistance"`
	IsHRS            bool    `json:"is_hrs"`
	CycleCount       int     `json:"cycle_count"`

	// RTN state
	RTN_State        bool    `json:"rtn_state"`
}

// VariableRRAM implements RRAM array with comprehensive variability model
type VariableRRAM struct {
	Config  *VariabilityConfig
	Cells   [][]*VariableRRAMCell
	Rows    int
	Cols    int
	mu      sync.RWMutex
}

// NewVariableRRAM creates RRAM array with variability
func NewVariableRRAM(rows, cols int, nominalHRS, nominalLRS float64, config *VariabilityConfig) *VariableRRAM {
	cells := make([][]*VariableRRAMCell, rows)

	for i := range cells {
		cells[i] = make([]*VariableRRAMCell, cols)
		for j := range cells[i] {
			// Sample D2D variations (fixed per device)
			d2d_hrs := rand.NormFloat64() * config.D2D_HRS_Sigma
			d2d_lrs := rand.NormFloat64() * config.D2D_LRS_Sigma
			d2d_vset := rand.NormFloat64() * config.D2D_Vset_Sigma
			d2d_vreset := rand.NormFloat64() * config.D2D_Vreset_Sigma

			cells[i][j] = &VariableRRAMCell{
				NominalHRS:        nominalHRS,
				NominalLRS:        nominalLRS,
				D2D_HRS_Offset:    d2d_hrs,
				D2D_LRS_Offset:    d2d_lrs,
				D2D_Vset_Offset:   d2d_vset,
				D2D_Vreset_Offset: d2d_vreset,
				CurrentResistance: nominalHRS * (1 + d2d_hrs),
				IsHRS:             true,
				CycleCount:        0,
				RTN_State:         false,
			}
		}
	}

	return &VariableRRAM{
		Config: config,
		Cells:  cells,
		Rows:   rows,
		Cols:   cols,
	}
}

// SET performs SET operation with C2C variation
func (v *VariableRRAM) SET(row, col int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	cell := v.Cells[row][col]

	// Sample C2C variation
	c2c := rand.NormFloat64() * v.Config.C2C_LRS_Sigma

	// Calculate effective LRS with D2D and C2C
	effectiveLRS := cell.NominalLRS * (1 + cell.D2D_LRS_Offset + c2c)

	// Endurance degradation
	if cell.CycleCount > v.Config.EnduranceCycles {
		degradation := float64(cell.CycleCount-v.Config.EnduranceCycles) * v.Config.DegradationRate
		effectiveLRS *= (1 + degradation)
	}

	cell.CurrentResistance = effectiveLRS
	cell.IsHRS = false
	cell.CycleCount++
}

// RESET performs RESET operation with C2C variation
func (v *VariableRRAM) RESET(row, col int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	cell := v.Cells[row][col]

	// Sample C2C variation
	c2c := rand.NormFloat64() * v.Config.C2C_HRS_Sigma

	// Calculate effective HRS with D2D and C2C
	effectiveHRS := cell.NominalHRS * (1 + cell.D2D_HRS_Offset + c2c)

	// Endurance degradation
	if cell.CycleCount > v.Config.EnduranceCycles {
		degradation := float64(cell.CycleCount-v.Config.EnduranceCycles) * v.Config.DegradationRate
		effectiveHRS *= (1 - degradation*0.5) // HRS tends to decrease
	}

	cell.CurrentResistance = effectiveHRS
	cell.IsHRS = true
	cell.CycleCount++
}

// Read returns resistance with RTN and read disturb effects
func (v *VariableRRAM) Read(row, col int) float64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	cell := v.Cells[row][col]

	resistance := cell.CurrentResistance

	// Random telegraph noise
	if rand.Float64() < v.Config.RTN_Frequency {
		cell.RTN_State = !cell.RTN_State
	}
	if cell.RTN_State {
		resistance *= (1 + v.Config.RTN_Amplitude)
	}

	// Read disturb
	if rand.Float64() < v.Config.ReadDisturbProb {
		// Small random walk in resistance
		resistance *= (1 + rand.NormFloat64()*0.01)
		cell.CurrentResistance = resistance
	}

	return resistance
}

// GetVariationStats returns statistical analysis of array variations
func (v *VariableRRAM) GetVariationStats() map[string]float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var hrs_values, lrs_values []float64

	for i := range v.Cells {
		for _, cell := range v.Cells[i] {
			if cell.IsHRS {
				hrs_values = append(hrs_values, cell.CurrentResistance)
			} else {
				lrs_values = append(lrs_values, cell.CurrentResistance)
			}
		}
	}

	stats := make(map[string]float64)

	if len(hrs_values) > 0 {
		stats["hrs_mean"] = mean(hrs_values)
		stats["hrs_std"] = stddev(hrs_values)
	}
	if len(lrs_values) > 0 {
		stats["lrs_mean"] = mean(lrs_values)
		stats["lrs_std"] = stddev(lrs_values)
	}

	// On/Off ratio statistics
	if len(hrs_values) > 0 && len(lrs_values) > 0 {
		stats["on_off_ratio"] = stats["hrs_mean"] / stats["lrs_mean"]
	}

	return stats
}

// ============================================================================
// VARIATION-AWARE TRAINING (CTSF/VACTSF)
// ============================================================================

// CTSFConfig defines Cycle-Time Scaling Factor training parameters
type CTSFConfig struct {
	// Scaling factors
	Alpha float64 `json:"alpha"` // Weight scaling factor
	Beta  float64 `json:"beta"`  // Gradient scaling factor

	// Noise injection during training
	NoiseScale float64 `json:"noise_scale"` // σ of injected noise

	// Quantization-aware
	NumBits    int     `json:"num_bits"`     // Weight precision
	ClipValue  float64 `json:"clip_value"`   // Weight clipping threshold

	// Variation modeling
	D2D_Model  bool    `json:"d2d_model"`    // Include D2D in training
	C2C_Model  bool    `json:"c2c_model"`    // Include C2C in training
}

// VACTSFConfig extends CTSF with variation-aware features
type VACTSFConfig struct {
	CTSFConfig

	// Bayesian components
	UseBayesian       bool    `json:"use_bayesian"`
	PriorSigma        float64 `json:"prior_sigma"`
	PosteriorSamples  int     `json:"posterior_samples"`

	// Robustness targets
	TargetD2D_Sigma   float64 `json:"target_d2d_sigma"`
	TargetC2C_Sigma   float64 `json:"target_c2c_sigma"`
}

// VariationAwareTrainer implements CTSF and VACTSF training
type VariationAwareTrainer struct {
	Config    *VACTSFConfig
	Weights   [][]float64
	Gradients [][]float64

	// Bayesian posterior
	WeightMean     [][]float64
	WeightVariance [][]float64

	// Training state
	Iteration int
	mu        sync.RWMutex
}

// NewVariationAwareTrainer creates a new variation-aware trainer
func NewVariationAwareTrainer(rows, cols int, config *VACTSFConfig) *VariationAwareTrainer {
	weights := make([][]float64, rows)
	gradients := make([][]float64, rows)
	weightMean := make([][]float64, rows)
	weightVar := make([][]float64, rows)

	for i := range weights {
		weights[i] = make([]float64, cols)
		gradients[i] = make([]float64, cols)
		weightMean[i] = make([]float64, cols)
		weightVar[i] = make([]float64, cols)

		for j := range weights[i] {
			weights[i][j] = rand.NormFloat64() * 0.1
			weightMean[i][j] = weights[i][j]
			weightVar[i][j] = config.PriorSigma * config.PriorSigma
		}
	}

	return &VariationAwareTrainer{
		Config:         config,
		Weights:        weights,
		Gradients:      gradients,
		WeightMean:     weightMean,
		WeightVariance: weightVar,
		Iteration:      0,
	}
}

// InjectVariation adds D2D and C2C noise to simulate hardware
func (t *VariationAwareTrainer) InjectVariation(weights [][]float64) [][]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	noisy := make([][]float64, len(weights))
	for i := range weights {
		noisy[i] = make([]float64, len(weights[i]))
		for j := range weights[i] {
			noise := 0.0

			if t.Config.D2D_Model {
				noise += rand.NormFloat64() * t.Config.TargetD2D_Sigma
			}
			if t.Config.C2C_Model {
				noise += rand.NormFloat64() * t.Config.TargetC2C_Sigma
			}

			noisy[i][j] = weights[i][j] * (1 + noise)
		}
	}

	return noisy
}

// CTSFForward performs forward pass with CTSF scaling
func (t *VariationAwareTrainer) CTSFForward(input []float64) []float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Apply weight scaling
	scaledWeights := make([][]float64, len(t.Weights))
	for i := range t.Weights {
		scaledWeights[i] = make([]float64, len(t.Weights[i]))
		for j := range t.Weights[i] {
			scaledWeights[i][j] = t.Weights[i][j] * t.Config.Alpha
		}
	}

	// Inject variation for hardware-aware training
	if t.Config.D2D_Model || t.Config.C2C_Model {
		scaledWeights = t.InjectVariation(scaledWeights)
	}

	// Matrix-vector multiplication
	output := make([]float64, len(t.Weights))
	for i := range t.Weights {
		sum := 0.0
		for j := range input {
			if j < len(t.Weights[i]) {
				sum += scaledWeights[i][j] * input[j]
			}
		}
		output[i] = sum
	}

	return output
}

// BayesianUpdate updates weight posterior using variational inference
func (t *VariationAwareTrainer) BayesianUpdate(gradient [][]float64, learningRate float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.Config.UseBayesian {
		// Standard gradient descent
		for i := range t.Weights {
			for j := range t.Weights[i] {
				t.Weights[i][j] -= learningRate * gradient[i][j] * t.Config.Beta

				// Clip weights
				if t.Weights[i][j] > t.Config.ClipValue {
					t.Weights[i][j] = t.Config.ClipValue
				} else if t.Weights[i][j] < -t.Config.ClipValue {
					t.Weights[i][j] = -t.Config.ClipValue
				}
			}
		}
		return
	}

	// Bayesian update with natural gradient approximation
	for i := range t.Weights {
		for j := range t.Weights[i] {
			// Update mean
			precisionOld := 1.0 / t.WeightVariance[i][j]
			precisionGrad := 1.0 / (t.Config.NoiseScale * t.Config.NoiseScale)

			newPrecision := precisionOld + precisionGrad
			newMean := (precisionOld*t.WeightMean[i][j] - learningRate*gradient[i][j]) / newPrecision

			t.WeightMean[i][j] = newMean
			t.WeightVariance[i][j] = 1.0 / newPrecision

			// Sample weight from posterior
			t.Weights[i][j] = t.WeightMean[i][j] + rand.NormFloat64()*math.Sqrt(t.WeightVariance[i][j])
		}
	}

	t.Iteration++
}

// Quantize applies quantization-aware training
func (t *VariationAwareTrainer) Quantize() [][]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	quantized := make([][]float64, len(t.Weights))
	levels := math.Pow(2, float64(t.Config.NumBits))

	for i := range t.Weights {
		quantized[i] = make([]float64, len(t.Weights[i]))
		for j := range t.Weights[i] {
			// Clip to range
			w := t.Weights[i][j]
			if w > t.Config.ClipValue {
				w = t.Config.ClipValue
			} else if w < -t.Config.ClipValue {
				w = -t.Config.ClipValue
			}

			// Quantize
			normalized := (w + t.Config.ClipValue) / (2 * t.Config.ClipValue)
			quantLevel := math.Round(normalized * (levels - 1))
			quantized[i][j] = (quantLevel/(levels-1))*2*t.Config.ClipValue - t.Config.ClipValue
		}
	}

	return quantized
}

// ============================================================================
// BAYESIAN NEURAL NETWORK FOR VARIATION TOLERANCE
// ============================================================================

// BayesianLayerConfig defines a Bayesian layer configuration
type BayesianLayerConfig struct {
	InputSize   int     `json:"input_size"`
	OutputSize  int     `json:"output_size"`
	PriorMu     float64 `json:"prior_mu"`
	PriorSigma  float64 `json:"prior_sigma"`
	PriorMixture float64 `json:"prior_mixture"` // For scale mixture prior
}

// BayesianLayer implements a single Bayesian neural network layer
type BayesianLayer struct {
	Config *BayesianLayerConfig

	// Weight distribution parameters
	WeightMu    [][]float64 // Mean
	WeightRho   [][]float64 // log(σ) parameterization

	// Bias distribution parameters
	BiasMu      []float64
	BiasRho     []float64

	// Sampled weights (for forward pass)
	WeightSample [][]float64
	BiasSample   []float64

	mu sync.RWMutex
}

// NewBayesianLayer creates a new Bayesian layer
func NewBayesianLayer(config *BayesianLayerConfig) *BayesianLayer {
	weightMu := make([][]float64, config.OutputSize)
	weightRho := make([][]float64, config.OutputSize)
	biasMu := make([]float64, config.OutputSize)
	biasRho := make([]float64, config.OutputSize)
	weightSample := make([][]float64, config.OutputSize)
	biasSample := make([]float64, config.OutputSize)

	// Initialize with prior
	initScale := math.Sqrt(2.0 / float64(config.InputSize))

	for i := 0; i < config.OutputSize; i++ {
		weightMu[i] = make([]float64, config.InputSize)
		weightRho[i] = make([]float64, config.InputSize)
		weightSample[i] = make([]float64, config.InputSize)

		for j := 0; j < config.InputSize; j++ {
			weightMu[i][j] = rand.NormFloat64() * initScale
			weightRho[i][j] = math.Log(math.Exp(config.PriorSigma) - 1) // Softplus inverse
		}

		biasMu[i] = 0.0
		biasRho[i] = math.Log(math.Exp(config.PriorSigma) - 1)
	}

	return &BayesianLayer{
		Config:       config,
		WeightMu:     weightMu,
		WeightRho:    weightRho,
		BiasMu:       biasMu,
		BiasRho:      biasRho,
		WeightSample: weightSample,
		BiasSample:   biasSample,
	}
}

// Softplus computes log(1 + exp(x)) for rho → sigma conversion
func softplus(x float64) float64 {
	if x > 20 {
		return x // Avoid overflow
	}
	return math.Log(1 + math.Exp(x))
}

// SampleWeights draws weights from posterior distribution
func (l *BayesianLayer) SampleWeights() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.WeightMu {
		for j := range l.WeightMu[i] {
			sigma := softplus(l.WeightRho[i][j])
			epsilon := rand.NormFloat64()
			l.WeightSample[i][j] = l.WeightMu[i][j] + sigma*epsilon
		}

		sigmaBias := softplus(l.BiasRho[i])
		epsilonBias := rand.NormFloat64()
		l.BiasSample[i] = l.BiasMu[i] + sigmaBias*epsilonBias
	}
}

// Forward performs forward pass with sampled weights
func (l *BayesianLayer) Forward(input []float64) []float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	output := make([]float64, l.Config.OutputSize)

	for i := 0; i < l.Config.OutputSize; i++ {
		sum := l.BiasSample[i]
		for j := 0; j < len(input) && j < l.Config.InputSize; j++ {
			sum += l.WeightSample[i][j] * input[j]
		}
		output[i] = sum
	}

	return output
}

// KLDivergence computes KL(q||p) for the layer
func (l *BayesianLayer) KLDivergence() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	kl := 0.0

	// KL for weights: KL(N(mu_q, sigma_q) || N(mu_p, sigma_p))
	priorVar := l.Config.PriorSigma * l.Config.PriorSigma

	for i := range l.WeightMu {
		for j := range l.WeightMu[i] {
			sigma := softplus(l.WeightRho[i][j])
			variance := sigma * sigma

			// KL = log(sigma_p/sigma_q) + (sigma_q^2 + (mu_q - mu_p)^2) / (2*sigma_p^2) - 0.5
			muDiff := l.WeightMu[i][j] - l.Config.PriorMu
			kl += math.Log(l.Config.PriorSigma/sigma) +
				  (variance + muDiff*muDiff)/(2*priorVar) - 0.5
		}
	}

	// KL for biases
	for i := range l.BiasMu {
		sigma := softplus(l.BiasRho[i])
		variance := sigma * sigma
		muDiff := l.BiasMu[i] - l.Config.PriorMu
		kl += math.Log(l.Config.PriorSigma/sigma) +
			  (variance + muDiff*muDiff)/(2*priorVar) - 0.5
	}

	return kl
}

// ============================================================================
// CRYOGENIC CIM ACCELERATOR
// ============================================================================

// CryoCIMAccelerator combines cryogenic memory with variation-aware inference
type CryoCIMAccelerator struct {
	// Memory technologies
	CryoRRAM          *CryoRRAM
	CASQUID           *CASQUID
	MagneticTopo      *MagneticTopoMemristor
	VariableRRAM      *VariableRRAM

	// Training/inference
	VariationTrainer  *VariationAwareTrainer
	BayesianLayers    []*BayesianLayer

	// Configuration
	OperatingTemp     float64
	UseCryogenic      bool
	UseSuperconducting bool

	mu sync.RWMutex
}

// CryoCIMConfig defines accelerator configuration
type CryoCIMConfig struct {
	Rows              int     `json:"rows"`
	Cols              int     `json:"cols"`
	Temperature       float64 `json:"temperature"`
	UseCryoRRAM       bool    `json:"use_cryo_rram"`
	UseCASQUID        bool    `json:"use_ca_squid"`
	UseMagneticTopo   bool    `json:"use_magnetic_topo"`
	UseVariableRRAM   bool    `json:"use_variable_rram"`
}

// NewCryoCIMAccelerator creates a comprehensive cryogenic CIM accelerator
func NewCryoCIMAccelerator(config *CryoCIMConfig) *CryoCIMAccelerator {
	acc := &CryoCIMAccelerator{
		OperatingTemp:      config.Temperature,
		UseCryogenic:       config.Temperature < 77.0, // Below liquid nitrogen
		UseSuperconducting: config.Temperature < 10.0, // Below typical Tc
	}

	if config.UseCryoRRAM {
		acc.CryoRRAM = NewCryoRRAM(&CryoRRAMConfig{
			Rows:        config.Rows,
			Cols:        config.Cols,
			Temperature: config.Temperature,
			HRS:         1e6,
			LRS:         1e4,
			NumLevels:   4,
			TunnelBarrier: 1.5,
			QuantumCorrection: true,
		})
	}

	if config.UseCASQUID && config.Temperature < 10.0 {
		acc.CASQUID = NewCASQUID(config.Rows, config.Cols, &CASQUIDConfig{
			Ic0:           100.0, // 100 µA
			IcModulation:  0.9,
			NumFluxStates: 16,
			CriticalTemp:  9.2,
			OperatingTemp: config.Temperature,
		})
	}

	if config.UseMagneticTopo {
		acc.MagneticTopo = NewMagneticTopoMemristor(config.Rows, config.Cols, &MagneticTopoConfig{
			CurieTemp:     180.0,
			OperatingTemp: config.Temperature,
			NumStates:     8,
			DMIStrength:   2.0,
		})
	}

	if config.UseVariableRRAM {
		acc.VariableRRAM = NewVariableRRAM(config.Rows, config.Cols, 1e6, 1e4, &VariabilityConfig{
			D2D_HRS_Sigma:    0.15, // 15% D2D variation
			D2D_LRS_Sigma:    0.10,
			C2C_HRS_Sigma:    0.05, // 5% C2C variation
			C2C_LRS_Sigma:    0.03,
			RTN_Amplitude:    0.02,
			RTN_Frequency:    0.01,
			ReadDisturbProb:  1e-6,
			EnduranceCycles:  1e6,
			DegradationRate:  1e-8,
		})
	}

	return acc
}

// MatrixVectorMultiply performs MVM using the configured memory technology
func (a *CryoCIMAccelerator) MatrixVectorMultiply(input []float64) []float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var weights [][]float64

	// Select memory technology based on temperature
	if a.UseSuperconducting && a.CASQUID != nil {
		// Use CA-SQUID at very low temperatures
		weights = make([][]float64, a.CASQUID.Rows)
		for i := 0; i < a.CASQUID.Rows; i++ {
			weights[i] = make([]float64, a.CASQUID.Cols)
			for j := 0; j < a.CASQUID.Cols; j++ {
				weights[i][j] = a.CASQUID.GetConductance(i, j)
			}
		}
	} else if a.UseCryogenic && a.CryoRRAM != nil {
		// Use cryogenic RRAM
		weights = make([][]float64, a.CryoRRAM.Config.Rows)
		for i := 0; i < a.CryoRRAM.Config.Rows; i++ {
			weights[i] = make([]float64, a.CryoRRAM.Config.Cols)
			for j := 0; j < a.CryoRRAM.Config.Cols; j++ {
				r := a.CryoRRAM.ReadCell(i, j)
				// Convert resistance to conductance (normalized)
				weights[i][j] = 1.0 / r * 1e4 // Normalize by LRS
			}
		}
	} else if a.VariableRRAM != nil {
		// Use room-temperature RRAM with variations
		weights = make([][]float64, a.VariableRRAM.Rows)
		for i := 0; i < a.VariableRRAM.Rows; i++ {
			weights[i] = make([]float64, a.VariableRRAM.Cols)
			for j := 0; j < a.VariableRRAM.Cols; j++ {
				r := a.VariableRRAM.Read(i, j)
				weights[i][j] = 1.0 / r * 1e4
			}
		}
	}

	if weights == nil {
		return nil
	}

	// MVM: output = W * input
	output := make([]float64, len(weights))
	for i := range weights {
		sum := 0.0
		for j := range input {
			if j < len(weights[i]) {
				sum += weights[i][j] * input[j]
			}
		}
		output[i] = sum
	}

	return output
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mean calculates the arithmetic mean
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stddev calculates standard deviation
func stddev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	sumSq := 0.0
	for _, v := range values {
		diff := v - m
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)-1))
}

// ============================================================================
// SERIALIZATION
// ============================================================================

// CryoCIMState holds serializable state
type CryoCIMState struct {
	OperatingTemp float64               `json:"operating_temp"`
	CryoRRAMState *CryoRRAMStateData    `json:"cryo_rram_state,omitempty"`
	VariabilityStats map[string]float64 `json:"variability_stats,omitempty"`
}

// CryoRRAMStateData holds RRAM array state
type CryoRRAMStateData struct {
	Rows        int         `json:"rows"`
	Cols        int         `json:"cols"`
	Temperature float64     `json:"temperature"`
	Resistances [][]float64 `json:"resistances"`
	States      [][]int     `json:"states"`
}

// ExportState exports accelerator state for analysis
func (a *CryoCIMAccelerator) ExportState() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	state := &CryoCIMState{
		OperatingTemp: a.OperatingTemp,
	}

	if a.CryoRRAM != nil {
		resistances := make([][]float64, a.CryoRRAM.Config.Rows)
		states := make([][]int, a.CryoRRAM.Config.Rows)

		for i := 0; i < a.CryoRRAM.Config.Rows; i++ {
			resistances[i] = make([]float64, a.CryoRRAM.Config.Cols)
			states[i] = make([]int, a.CryoRRAM.Config.Cols)
			for j := 0; j < a.CryoRRAM.Config.Cols; j++ {
				resistances[i][j] = a.CryoRRAM.Cells[i][j].Resistance
				states[i][j] = a.CryoRRAM.Cells[i][j].State
			}
		}

		state.CryoRRAMState = &CryoRRAMStateData{
			Rows:        a.CryoRRAM.Config.Rows,
			Cols:        a.CryoRRAM.Config.Cols,
			Temperature: a.CryoRRAM.Config.Temperature,
			Resistances: resistances,
			States:      states,
		}
	}

	if a.VariableRRAM != nil {
		state.VariabilityStats = a.VariableRRAM.GetVariationStats()
	}

	return json.MarshalIndent(state, "", "  ")
}

// ============================================================================
// BENCHMARK UTILITIES
// ============================================================================

// BenchmarkConfig defines benchmark parameters
type BenchmarkConfig struct {
	Iterations      int       `json:"iterations"`
	Temperatures    []float64 `json:"temperatures"` // Test at multiple temps
	ArraySizes      []int     `json:"array_sizes"`
	MeasureAccuracy bool      `json:"measure_accuracy"`
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	Temperature      float64 `json:"temperature"`
	ArraySize        int     `json:"array_size"`
	SwitchingSuccess float64 `json:"switching_success"` // Success rate
	ReadVariation    float64 `json:"read_variation"`    // Coefficient of variation
	EnergyEstimate   float64 `json:"energy_estimate"`   // Relative energy
}

// RunCryogenicBenchmark tests cryogenic RRAM at various temperatures
func RunCryogenicBenchmark(config *BenchmarkConfig) []BenchmarkResult {
	results := []BenchmarkResult{}

	for _, temp := range config.Temperatures {
		for _, size := range config.ArraySizes {
			rram := NewCryoRRAM(&CryoRRAMConfig{
				Rows:              size,
				Cols:              size,
				Temperature:       temp,
				HRS:               1e6,
				LRS:               1e4,
				NumLevels:         4,
				QuantumCorrection: true,
				TunnelBarrier:     1.5,
			})

			// Test switching
			successCount := 0
			for iter := 0; iter < config.Iterations; iter++ {
				row := rand.Intn(size)
				col := rand.Intn(size)
				targetState := rand.Intn(4)

				if rram.CryogenicSwitching(row, col, targetState) {
					successCount++
				}
			}

			// Measure read variation
			var readings []float64
			for i := 0; i < size; i++ {
				for j := 0; j < size; j++ {
					readings = append(readings, rram.ReadCell(i, j))
				}
			}

			m := mean(readings)
			s := stddev(readings)
			cv := 0.0
			if m > 0 {
				cv = s / m // Coefficient of variation
			}

			// Energy estimate (relative, based on switching rate)
			thermalRate := rram.ArrheniusRate(temp)
			roomRate := rram.ArrheniusRate(300.0)
			energyRatio := thermalRate / roomRate

			results = append(results, BenchmarkResult{
				Temperature:      temp,
				ArraySize:        size,
				SwitchingSuccess: float64(successCount) / float64(config.Iterations),
				ReadVariation:    cv,
				EnergyEstimate:   energyRatio,
			})
		}
	}

	return results
}
