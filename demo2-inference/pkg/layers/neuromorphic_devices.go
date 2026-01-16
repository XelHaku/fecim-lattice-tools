// neuromorphic_devices.go - Neuromorphic Processor and Emerging Device Models
//
// This module implements:
// - Neuromorphic processor architectures (Loihi, TrueNorth, BrainScaleS)
// - Spiking neural network primitives
// - Emerging ferroelectric materials (AlScN, HZO comparison)
// - Device-level simulation for next-gen FeFETs
// - Cross-platform neuromorphic benchmarking
//
// Based on research findings:
// - Intel Hala Point: 1.15B neurons, 20 petaops, 15 TOPS/W
// - IBM NorthPole: 22× faster than TrueNorth, 25× more efficient
// - AlScN: Pr 80-150 µC/cm², Ec 2000-5000 kV/cm, >3× HZO
// - BrainScaleS: 864× faster than biological neurons
//
// References:
// - Intel Loihi/Hala Point architecture
// - IBM TrueNorth/NorthPole
// - AlScN ferroelectrics (PMC 2024)

package layers

import (
	"math"
	"math/rand"
)

// ================== Neuromorphic Processor Architectures ==================

// NeuromorphicConfig configures a neuromorphic processor
type NeuromorphicConfig struct {
	// Architecture type
	Architecture     string // "loihi", "loihi2", "truenorth", "northpole", "brainscales"

	// Neuron configuration
	NumNeurons       int64
	NumSynapses      int64
	NumCores         int

	// Timing
	TimeStepUs       float64 // Time step in microseconds
	SpeedupFactor    float64 // vs biological (1.0 = real-time)

	// Power
	PowerWatts       float64
	EnergyPerSpike   float64 // pJ per spike

	// Precision
	WeightBits       int
	NeuronStateBits  int

	// Learning
	OnChipLearning   bool
	STDPSupport      bool
	LearningRules    []string
}

// Loihi2Config returns Intel Loihi 2 configuration
func Loihi2Config() *NeuromorphicConfig {
	return &NeuromorphicConfig{
		Architecture:    "loihi2",
		NumNeurons:      1000000,      // 1M neurons
		NumSynapses:     120000000,    // 120M synapses
		NumCores:        128,
		TimeStepUs:      1.0,          // 1 µs time step
		SpeedupFactor:   1000,         // 1000× real-time
		PowerWatts:      1.0,          // ~1W typical
		EnergyPerSpike:  23.6,         // ~24 pJ/spike
		WeightBits:      8,
		NeuronStateBits: 24,
		OnChipLearning:  true,
		STDPSupport:     true,
		LearningRules:   []string{"stdp", "r-stdp", "e-prop"},
	}
}

// HalaPointConfig returns Intel Hala Point configuration
func HalaPointConfig() *NeuromorphicConfig {
	return &NeuromorphicConfig{
		Architecture:    "hala_point",
		NumNeurons:      1150000000,   // 1.15B neurons
		NumSynapses:     128000000000, // 128B synapses
		NumCores:        1152 * 128,   // 1152 Loihi 2 chips
		TimeStepUs:      1.0,
		SpeedupFactor:   20000,        // 20× faster than human brain
		PowerWatts:      2600,         // ~2.6 kW
		EnergyPerSpike:  23.6,
		WeightBits:      8,
		NeuronStateBits: 24,
		OnChipLearning:  true,
		STDPSupport:     true,
		LearningRules:   []string{"stdp", "r-stdp", "e-prop"},
	}
}

// TrueNorthConfig returns IBM TrueNorth configuration
func TrueNorthConfig() *NeuromorphicConfig {
	return &NeuromorphicConfig{
		Architecture:    "truenorth",
		NumNeurons:      1000000,      // 1M neurons
		NumSynapses:     256000000,    // 256M synapses
		NumCores:        4096,
		TimeStepUs:      1000,         // 1 ms tick
		SpeedupFactor:   1.0,          // Real-time
		PowerWatts:      0.07,         // 70 mW
		EnergyPerSpike:  26.0,         // ~26 pJ/spike
		WeightBits:      1,            // Binary synapses
		NeuronStateBits: 32,
		OnChipLearning:  false,        // No on-chip learning
		STDPSupport:     false,
		LearningRules:   []string{},
	}
}

// NorthPoleConfig returns IBM NorthPole configuration
func NorthPoleConfig() *NeuromorphicConfig {
	return &NeuromorphicConfig{
		Architecture:    "northpole",
		NumNeurons:      22000000,     // 22M "neurons" (activation units)
		NumSynapses:     22000000000,  // 22B weights
		NumCores:        256,
		TimeStepUs:      10,           // ~100 MHz operation
		SpeedupFactor:   4000,         // 4000× faster than TrueNorth
		PowerWatts:      12,           // ~12W
		EnergyPerSpike:  0.5,          // ~0.5 pJ/op (highly efficient)
		WeightBits:      8,
		NeuronStateBits: 8,
		OnChipLearning:  false,
		STDPSupport:     false,
		LearningRules:   []string{},
	}
}

// BrainScaleSConfig returns BrainScaleS-2 configuration
func BrainScaleSConfig() *NeuromorphicConfig {
	return &NeuromorphicConfig{
		Architecture:    "brainscales2",
		NumNeurons:      512,          // Per chip
		NumSynapses:     130000,       // Per chip
		NumCores:        1,
		TimeStepUs:      0.001,        // Sub-µs (analog)
		SpeedupFactor:   864,          // 864× biological speed
		PowerWatts:      0.1,          // ~100 mW per chip
		EnergyPerSpike:  1.0,          // ~1 pJ/spike (analog)
		WeightBits:      6,
		NeuronStateBits: 10,           // Analog precision
		OnChipLearning:  true,
		STDPSupport:     true,
		LearningRules:   []string{"stdp", "reward_modulated"},
	}
}

// NeuromorphicProcessor simulates a neuromorphic processor
type NeuromorphicProcessor struct {
	Config         *NeuromorphicConfig
	Neurons        []*SpikingNeuron
	Synapses       [][]*Synapse
	SpikeQueue     []SpikeEvent
	CurrentTime    float64 // µs
	TotalSpikes    int64
	TotalEnergy    float64 // pJ

	// Performance metrics
	TOPS           float64
	EffectiveTOPS  float64
	Utilization    float64
}

// SpikingNeuron represents a spiking neuron
type SpikingNeuron struct {
	ID             int
	Potential      float64 // Membrane potential
	Threshold      float64 // Spike threshold
	ResetPotential float64 // Reset value after spike
	LeakRate       float64 // Leak time constant
	RefractoryTime float64 // Refractory period (µs)
	LastSpikeTime  float64
	SpikeCount     int64
	InputCurrent   float64
}

// Synapse represents a synaptic connection
type Synapse struct {
	PreNeuron      int
	PostNeuron     int
	Weight         float64
	Delay          float64 // Synaptic delay (µs)
	Plastic        bool    // Learning enabled
	EligibilityTrace float64
}

// SpikeEvent represents a spike in the queue
type SpikeEvent struct {
	NeuronID  int
	Time      float64
	Processed bool
}

// NewNeuromorphicProcessor creates a neuromorphic processor
func NewNeuromorphicProcessor(config *NeuromorphicConfig, numNeurons, numSynapsesPerNeuron int) *NeuromorphicProcessor {
	neurons := make([]*SpikingNeuron, numNeurons)
	for i := range neurons {
		neurons[i] = &SpikingNeuron{
			ID:             i,
			Potential:      0,
			Threshold:      1.0,
			ResetPotential: 0,
			LeakRate:       0.1, // 10% leak per time step
			RefractoryTime: 2.0, // 2 µs refractory
			LastSpikeTime:  -100,
		}
	}

	// Create random synaptic connections
	synapses := make([][]*Synapse, numNeurons)
	for i := range synapses {
		synapses[i] = make([]*Synapse, 0, numSynapsesPerNeuron)
		for j := 0; j < numSynapsesPerNeuron; j++ {
			post := rand.Intn(numNeurons)
			if post != i { // No self-connections
				syn := &Synapse{
					PreNeuron:  i,
					PostNeuron: post,
					Weight:     rand.Float64()*2 - 1, // [-1, 1]
					Delay:      rand.Float64() * 5,   // 0-5 µs delay
					Plastic:    config.OnChipLearning,
				}
				synapses[i] = append(synapses[i], syn)
			}
		}
	}

	return &NeuromorphicProcessor{
		Config:      config,
		Neurons:    neurons,
		Synapses:   synapses,
		SpikeQueue: make([]SpikeEvent, 0),
		CurrentTime: 0,
	}
}

// Step advances simulation by one time step
func (np *NeuromorphicProcessor) Step() int {
	cfg := np.Config
	dt := cfg.TimeStepUs
	spikesThisStep := 0

	// Process pending spikes
	for i := range np.SpikeQueue {
		if !np.SpikeQueue[i].Processed && np.SpikeQueue[i].Time <= np.CurrentTime {
			np.SpikeQueue[i].Processed = true
			// Deliver spike to post-synaptic neurons
			preID := np.SpikeQueue[i].NeuronID
			for _, syn := range np.Synapses[preID] {
				np.Neurons[syn.PostNeuron].InputCurrent += syn.Weight
			}
		}
	}

	// Update all neurons
	for _, neuron := range np.Neurons {
		// Check refractory period
		if np.CurrentTime-neuron.LastSpikeTime < neuron.RefractoryTime {
			continue
		}

		// Leak
		neuron.Potential *= (1 - neuron.LeakRate)

		// Integrate input
		neuron.Potential += neuron.InputCurrent
		neuron.InputCurrent = 0 // Reset input

		// Check threshold
		if neuron.Potential >= neuron.Threshold {
			// Spike!
			neuron.Potential = neuron.ResetPotential
			neuron.LastSpikeTime = np.CurrentTime
			neuron.SpikeCount++
			np.TotalSpikes++
			spikesThisStep++

			// Add spike to queue
			np.SpikeQueue = append(np.SpikeQueue, SpikeEvent{
				NeuronID: neuron.ID,
				Time:     np.CurrentTime,
			})

			// Energy cost
			np.TotalEnergy += cfg.EnergyPerSpike
		}
	}

	// Advance time
	np.CurrentTime += dt

	return spikesThisStep
}

// InjectSpikes injects external spikes
func (np *NeuromorphicProcessor) InjectSpikes(neuronIDs []int) {
	for _, id := range neuronIDs {
		if id >= 0 && id < len(np.Neurons) {
			np.Neurons[id].Potential = np.Neurons[id].Threshold + 0.1 // Ensure spike
		}
	}
}

// GetMetrics returns performance metrics
func (np *NeuromorphicProcessor) GetMetrics() map[string]float64 {
	cfg := np.Config
	timeSeconds := np.CurrentTime / 1e6 // Convert µs to s

	// Calculate effective TOPS
	totalOps := np.TotalSpikes * int64(len(np.Synapses[0])) // Spikes × synapses
	if timeSeconds > 0 {
		np.TOPS = float64(totalOps) / timeSeconds / 1e12
	}

	// Utilization
	maxSpikes := float64(len(np.Neurons)) * np.CurrentTime / cfg.TimeStepUs
	if maxSpikes > 0 {
		np.Utilization = float64(np.TotalSpikes) / maxSpikes
	}

	return map[string]float64{
		"TotalSpikes":    float64(np.TotalSpikes),
		"TotalEnergyPJ":  np.TotalEnergy,
		"SimTimeUs":      np.CurrentTime,
		"TOPS":           np.TOPS,
		"Utilization":    np.Utilization,
		"AvgFiringRate":  float64(np.TotalSpikes) / float64(len(np.Neurons)) / (np.CurrentTime / 1e6),
		"EnergyPerSpike": cfg.EnergyPerSpike,
	}
}

// ================== Emerging Ferroelectric Materials ==================

// FerroelectricMaterialType represents material types
type FerroelectricMaterialType int

const (
	MaterialHZO FerroelectricMaterialType = iota
	MaterialAlScN
	MaterialAlBN
	MaterialPZT
	MaterialHZOSuperlattice
)

// FerroelectricMaterial defines material properties
type FerroelectricMaterial struct {
	Type             FerroelectricMaterialType
	Name             string

	// Polarization (µC/cm²)
	RemanentPolarization float64 // Pr
	SaturationPolarization float64 // Ps

	// Coercive field (kV/cm)
	CoerciveField    float64 // Ec

	// Memory window (V)
	MemoryWindow     float64

	// Curie temperature (°C)
	CurieTemp        float64

	// Endurance (cycles)
	EnduranceCycles  float64

	// Retention (years at room temp)
	RetentionYears   float64

	// Operating voltage (V)
	ProgramVoltage   float64
	EraseVoltage     float64

	// CMOS compatibility
	CMOSCompatible   bool
	MaxProcessTemp   float64 // °C

	// Scaling
	MinThickness     float64 // nm
	ScalingLimit     float64 // nm (minimum feature)
}

// NewHZOMaterial creates HfO2-ZrO2 material
func NewHZOMaterial() *FerroelectricMaterial {
	return &FerroelectricMaterial{
		Type:                  MaterialHZO,
		Name:                  "Hf0.5Zr0.5O2 (HZO)",
		RemanentPolarization:  25,    // 15-40 µC/cm²
		SaturationPolarization: 40,
		CoerciveField:         1500,  // 1000-2000 kV/cm
		MemoryWindow:          1.5,   // ~1.5 V
		CurieTemp:             450,   // ~450 °C
		EnduranceCycles:       1e10,  // 10^10 (superlattice)
		RetentionYears:        10,
		ProgramVoltage:        3.0,
		EraseVoltage:          -3.0,
		CMOSCompatible:        true,
		MaxProcessTemp:        400,
		MinThickness:          5,     // 5 nm
		ScalingLimit:          5,
	}
}

// NewAlScNMaterial creates AlScN wurtzite material
func NewAlScNMaterial() *FerroelectricMaterial {
	return &FerroelectricMaterial{
		Type:                  MaterialAlScN,
		Name:                  "Al0.64Sc0.36N (AlScN)",
		RemanentPolarization:  110,   // 80-150 µC/cm²
		SaturationPolarization: 150,
		CoerciveField:         3500,  // 2000-5000 kV/cm
		MemoryWindow:          3.5,   // >3V
		CurieTemp:             1100,  // >1100 °C
		EnduranceCycles:       1e5,   // 10^5 (current limit)
		RetentionYears:        10,
		ProgramVoltage:        10.0,  // Higher due to Ec
		EraseVoltage:          -10.0,
		CMOSCompatible:        true,  // BEOL compatible
		MaxProcessTemp:        400,
		MinThickness:          10,    // 10 nm
		ScalingLimit:          40,    // 40 nm demonstrated
	}
}

// NewAlBNMaterial creates AlBN wurtzite material
func NewAlBNMaterial() *FerroelectricMaterial {
	return &FerroelectricMaterial{
		Type:                  MaterialAlBN,
		Name:                  "AlBN",
		RemanentPolarization:  130,   // 100-150 µC/cm²
		SaturationPolarization: 160,
		CoerciveField:         4000,  // 3000-5000 kV/cm
		MemoryWindow:          4.0,
		CurieTemp:             1200,
		EnduranceCycles:       1e4,   // Limited data
		RetentionYears:        10,
		ProgramVoltage:        12.0,
		EraseVoltage:          -12.0,
		CMOSCompatible:        true,
		MaxProcessTemp:        400,
		MinThickness:          15,
		ScalingLimit:          50,
	}
}

// NewHZOSuperlattice creates HZO superlattice (IronLattice)
func NewHZOSuperlattice() *FerroelectricMaterial {
	return &FerroelectricMaterial{
		Type:                  MaterialHZOSuperlattice,
		Name:                  "HfO2/ZrO2 Superlattice (IronLattice)",
		RemanentPolarization:  45,    // Enhanced vs standard HZO
		SaturationPolarization: 60,
		CoerciveField:         850,   // Reduced Ec
		MemoryWindow:          3.5,   // ~3.5 V
		CurieTemp:             500,
		EnduranceCycles:       1e10,  // 10^10 cycles
		RetentionYears:        10,
		ProgramVoltage:        3.0,   // Lower voltage
		EraseVoltage:          -3.0,
		CMOSCompatible:        true,
		MaxProcessTemp:        400,
		MinThickness:          4,     // 4 nm (4×[HfO2/ZrO2])
		ScalingLimit:          4,
	}
}

// CompareMaterials compares two ferroelectric materials
func CompareMaterials(m1, m2 *FerroelectricMaterial) *MaterialComparison {
	return &MaterialComparison{
		Material1:         m1.Name,
		Material2:         m2.Name,
		PrRatio:           m1.RemanentPolarization / m2.RemanentPolarization,
		EcRatio:           m1.CoerciveField / m2.CoerciveField,
		MemoryWindowRatio: m1.MemoryWindow / m2.MemoryWindow,
		EnduranceRatio:    m1.EnduranceCycles / m2.EnduranceCycles,
		VoltageRatio:      m1.ProgramVoltage / m2.ProgramVoltage,
		ScalingRatio:      m2.ScalingLimit / m1.ScalingLimit, // Lower is better
	}
}

// MaterialComparison contains comparison results
type MaterialComparison struct {
	Material1         string
	Material2         string
	PrRatio           float64 // M1/M2
	EcRatio           float64
	MemoryWindowRatio float64
	EnduranceRatio    float64
	VoltageRatio      float64
	ScalingRatio      float64
}

// ================== FeFET Device Model ==================

// FeFETConfig configures FeFET device
type FeFETConfig struct {
	Material         *FerroelectricMaterial
	FeThickness      float64 // Ferroelectric thickness (nm)
	ChannelLength    float64 // Channel length (nm)
	ChannelWidth     float64 // Channel width (nm)
	InterfaceLayer   float64 // Interface layer thickness (nm)
	ChannelMaterial  string  // "Si", "IGZO", "MoS2", "WSe2"
	OperatingTemp    float64 // °C
}

// DefaultFeFETConfig returns typical FeFET parameters
func DefaultFeFETConfig(material *FerroelectricMaterial) *FeFETConfig {
	return &FeFETConfig{
		Material:        material,
		FeThickness:    10,    // 10 nm
		ChannelLength:  28,    // 28 nm node
		ChannelWidth:   100,   // 100 nm
		InterfaceLayer: 1,     // 1 nm IL
		ChannelMaterial: "Si",
		OperatingTemp:  25,    // Room temp
	}
}

// FeFETDevice models a ferroelectric FET
type FeFETDevice struct {
	Config            *FeFETConfig
	Polarization      float64 // Current polarization state
	ThresholdVoltage  float64 // Current Vth
	DrainCurrent      float64 // Current Id
	State             int     // Multi-level state
	CycleCount        int64   // Write cycles
	RetentionLoss     float64 // Retention degradation

	// Device characteristics
	OnCurrent         float64 // Ion (µA/µm)
	OffCurrent        float64 // Ioff (pA/µm)
	SubthresholdSwing float64 // SS (mV/dec)
}

// NewFeFETDevice creates a FeFET device model
func NewFeFETDevice(config *FeFETConfig) *FeFETDevice {
	mat := config.Material

	// Calculate threshold voltage window
	// Vth depends on polarization and interface
	vthWindow := mat.MemoryWindow * (mat.RemanentPolarization / 100)

	return &FeFETDevice{
		Config:            config,
		Polarization:      0,
		ThresholdVoltage:  0.5, // Default Vth
		DrainCurrent:      0,
		State:             0,
		CycleCount:        0,
		RetentionLoss:     0,
		OnCurrent:         500,   // 500 µA/µm typical
		OffCurrent:        1,     // 1 pA/µm
		SubthresholdSwing: 70,    // 70 mV/dec (near ideal)
	}
}

// Program programs the FeFET to a state
func (dev *FeFETDevice) Program(state int, voltage float64) error {
	mat := dev.Config.Material

	// Check if voltage exceeds coercive field
	effectiveField := math.Abs(voltage) * 1e7 / dev.Config.FeThickness // kV/cm
	if effectiveField < mat.CoerciveField*0.5 {
		// Insufficient field for switching
		return nil
	}

	// Calculate polarization based on voltage
	// Simplified Preisach model
	normalizedV := voltage / mat.ProgramVoltage
	targetP := math.Tanh(normalizedV * 3) * mat.RemanentPolarization

	// Partial switching
	switchFraction := math.Min(1.0, effectiveField/mat.CoerciveField)
	dev.Polarization = dev.Polarization*(1-switchFraction) + targetP*switchFraction

	// Update threshold voltage
	dev.ThresholdVoltage = 0.5 + dev.Polarization/mat.RemanentPolarization*mat.MemoryWindow/2

	// Update state
	dev.State = state
	dev.CycleCount++

	// Check endurance
	if float64(dev.CycleCount) > mat.EnduranceCycles {
		dev.RetentionLoss += 0.01 // 1% degradation per excess cycle
	}

	return nil
}

// Read reads the FeFET state
func (dev *FeFETDevice) Read(gateVoltage, drainVoltage float64) float64 {
	// Calculate drain current using simplified model
	// Id = µn × Cox × W/L × ((Vgs - Vth) × Vds - Vds²/2)

	vgs := gateVoltage
	vth := dev.ThresholdVoltage * (1 - dev.RetentionLoss)
	vds := drainVoltage

	if vgs < vth {
		// Subthreshold
		dev.DrainCurrent = dev.OffCurrent * 1e-6 * math.Exp((vgs-vth)/(dev.SubthresholdSwing/1000/2.3))
	} else {
		// Linear/saturation
		vov := vgs - vth // Overdrive voltage
		if vds < vov {
			// Linear region
			dev.DrainCurrent = dev.OnCurrent * 1e-6 * (vov*vds - vds*vds/2) / (vov * vov / 2)
		} else {
			// Saturation
			dev.DrainCurrent = dev.OnCurrent * 1e-6 * vov * vov / (2 * vov)
		}
	}

	return dev.DrainCurrent
}

// GetOnOffRatio returns the on/off current ratio
func (dev *FeFETDevice) GetOnOffRatio() float64 {
	return dev.OnCurrent * 1e6 / dev.OffCurrent // µA/µm to pA/µm
}

// ================== Device Comparison Framework ==================

// DeviceBenchmark benchmarks different device types
type DeviceBenchmark struct {
	Devices       map[string]*FeFETDevice
	Measurements  map[string]*DeviceMeasurement
}

// DeviceMeasurement contains measurement results
type DeviceMeasurement struct {
	DeviceName      string
	MemoryWindow    float64 // V
	OnOffRatio      float64
	WriteLatency    float64 // ns
	ReadLatency     float64 // ns
	WriteEnergy     float64 // fJ
	ReadEnergy      float64 // fJ
	Endurance       float64 // cycles
	Retention       float64 // years
	CellArea        float64 // F²
}

// NewDeviceBenchmark creates a benchmark framework
func NewDeviceBenchmark() *DeviceBenchmark {
	return &DeviceBenchmark{
		Devices:      make(map[string]*FeFETDevice),
		Measurements: make(map[string]*DeviceMeasurement),
	}
}

// AddDevice adds a device for benchmarking
func (db *DeviceBenchmark) AddDevice(name string, device *FeFETDevice) {
	db.Devices[name] = device
}

// RunBenchmark runs benchmarks on all devices
func (db *DeviceBenchmark) RunBenchmark() {
	for name, device := range db.Devices {
		mat := device.Config.Material
		cfg := device.Config

		// Write latency (based on RC and switching time)
		// Simplified: ~100 ns typical
		writeLatency := 100.0 * (mat.CoerciveField / 1500) // Scale with Ec

		// Read latency
		readLatency := 10.0 // ~10 ns typical

		// Write energy (CV²)
		capacitance := mat.RemanentPolarization * 1e-6 / mat.ProgramVoltage * cfg.ChannelWidth * cfg.ChannelLength * 1e-14
		writeEnergy := 0.5 * capacitance * mat.ProgramVoltage * mat.ProgramVoltage * 1e15 // fJ

		// Read energy (much lower)
		readEnergy := writeEnergy * 0.01 // ~1% of write

		// Cell area (in F²)
		cellArea := 4.0 * (cfg.ChannelLength / mat.ScalingLimit) * (cfg.ChannelWidth / mat.ScalingLimit)

		db.Measurements[name] = &DeviceMeasurement{
			DeviceName:   name,
			MemoryWindow: mat.MemoryWindow,
			OnOffRatio:   device.GetOnOffRatio(),
			WriteLatency: writeLatency,
			ReadLatency:  readLatency,
			WriteEnergy:  writeEnergy,
			ReadEnergy:   readEnergy,
			Endurance:    mat.EnduranceCycles,
			Retention:    mat.RetentionYears,
			CellArea:     cellArea,
		}
	}
}

// GetRankings returns devices ranked by a metric
func (db *DeviceBenchmark) GetRankings(metric string) []string {
	type deviceScore struct {
		name  string
		score float64
	}

	scores := make([]deviceScore, 0, len(db.Measurements))
	for name, m := range db.Measurements {
		var score float64
		switch metric {
		case "memory_window":
			score = m.MemoryWindow
		case "on_off_ratio":
			score = m.OnOffRatio
		case "write_speed":
			score = 1.0 / m.WriteLatency
		case "write_energy":
			score = 1.0 / m.WriteEnergy
		case "endurance":
			score = m.Endurance
		case "density":
			score = 1.0 / m.CellArea
		default:
			score = 0
		}
		scores = append(scores, deviceScore{name, score})
	}

	// Sort by score (descending)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	rankings := make([]string, len(scores))
	for i, s := range scores {
		rankings[i] = s.name
	}
	return rankings
}
