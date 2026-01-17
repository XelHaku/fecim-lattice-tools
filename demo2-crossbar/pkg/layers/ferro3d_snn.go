// ferro3d_snn.go - 3D Ferroelectric visualization and Spiking Neural Network demo
// Implements FeNAND array simulation and all-ferroelectric SNN with STDP learning
// Based on research: SK hynix IEDM 2024 (4000x density), All-FeFET SNN (94.9% accuracy)

package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// 3D Ferroelectric NAND (FeNAND) Visualization
// ============================================================================

// FeNANDArchitecture defines the 3D FeNAND architecture type
type FeNANDArchitecture int

const (
	ArchVC_FeNAND FeNANDArchitecture = iota // Vertical Channel FeNAND
	ArchHC_FeNAND                           // Horizontal Channel FeNAND
	ArchFeAND                               // Ferroelectric AND array
	ArchHybridFeNAND                        // Hybrid charge-trap + ferroelectric
)

// FerroelectricPhase represents the crystal phase
type FerroelectricPhase int

const (
	PhaseOrthorhombic FerroelectricPhase = iota // Ferroelectric (desired)
	PhaseTetragonal                              // Antiferroelectric
	PhaseMonoclinic                              // Non-ferroelectric
	PhaseCubic                                   // Paraelectric
)

// DomainState represents the polarization state
type DomainState int

const (
	DomainUp   DomainState = 1  // P+
	DomainDown DomainState = -1 // P-
	DomainMixed DomainState = 0 // Mixed/transitioning
)

// FeNANDCell represents a single ferroelectric transistor cell
type FeNANDCell struct {
	Layer        int            // Vertical layer (wordline)
	String       int            // NAND string index
	BitLine      int            // Bit line index
	Phase        FerroelectricPhase
	Polarization float64        // -1 to +1 (normalized)
	Threshold    float64        // Threshold voltage (V)
	Conductance  float64        // Channel conductance
	DomainState  DomainState
	CycleCount   int64          // Write/erase cycles
	RetentionLoss float64       // Polarization loss over time
}

// FeNANDString represents a vertical NAND string
type FeNANDString struct {
	Index       int
	BitLine     int
	Cells       []*FeNANDCell
	SelectGate  float64 // SSL/GSL transistor
	ChannelType string  // "poly-Si", "IGZO", "InZnOx"
}

// FeNANDArray represents a 3D FeNAND array
type FeNANDArray struct {
	Architecture   FeNANDArchitecture
	NumLayers      int // Number of wordline layers
	NumStrings     int // Strings per block
	NumBitLines    int // Bit lines
	Strings        [][]*FeNANDString
	WordLines      []float64     // WL voltages
	BitLineVoltages []float64

	// Material parameters
	FerroMaterial  string  // "HZO", "La:HfO2", "Si:HfO2"
	FerroThickness float64 // nm
	GrainSize      float64 // nm
	CoerciveField  float64 // MV/cm
	Pr             float64 // Remanent polarization (µC/cm²)

	// Performance metrics
	MemoryWindow   float64 // V
	Endurance      float64 // Cycles
	Retention      float64 // Years
	BitsPerCell    int     // MLC levels
}

// NewFeNANDArray creates a new 3D FeNAND array
func NewFeNANDArray(layers, strings, bitlines int, arch FeNANDArchitecture) *FeNANDArray {
	array := &FeNANDArray{
		Architecture:   arch,
		NumLayers:      layers,
		NumStrings:     strings,
		NumBitLines:    bitlines,
		Strings:        make([][]*FeNANDString, bitlines),
		WordLines:      make([]float64, layers),
		BitLineVoltages: make([]float64, bitlines),
		FerroMaterial:  "HZO",
		FerroThickness: 10,    // nm
		GrainSize:      15,    // nm
		CoerciveField:  1.0,   // MV/cm
		Pr:             25,    // µC/cm²
		MemoryWindow:   2.0,   // V
		Endurance:      1e12,
		Retention:      10,    // years
		BitsPerCell:    3,     // 8 levels
	}

	// Initialize strings and cells
	for bl := 0; bl < bitlines; bl++ {
		array.Strings[bl] = make([]*FeNANDString, strings)
		for s := 0; s < strings; s++ {
			str := &FeNANDString{
				Index:       s,
				BitLine:     bl,
				Cells:       make([]*FeNANDCell, layers),
				SelectGate:  0,
				ChannelType: "poly-Si",
			}

			for l := 0; l < layers; l++ {
				str.Cells[l] = &FeNANDCell{
					Layer:        l,
					String:       s,
					BitLine:      bl,
					Phase:        PhaseOrthorhombic,
					Polarization: 0,
					Threshold:    0.5,
					Conductance:  1e-6,
					DomainState:  DomainMixed,
				}
			}

			array.Strings[bl][s] = str
		}
	}

	return array
}

// ProgramCell programs a cell to a target polarization
func (arr *FeNANDArray) ProgramCell(layer, str, bl int, targetPol float64) error {
	if bl >= arr.NumBitLines || str >= arr.NumStrings || layer >= arr.NumLayers {
		return fmt.Errorf("cell index out of range")
	}

	cell := arr.Strings[bl][str].Cells[layer]

	// Apply programming pulse (simplified model)
	// Actual polarization depends on pulse amplitude and duration
	oldPol := cell.Polarization

	// Ferroelectric switching dynamics (tanh approximation)
	switchingRate := 0.8 // Depends on field strength
	cell.Polarization = oldPol + switchingRate*(targetPol-oldPol)

	// Update threshold voltage based on polarization
	cell.Threshold = 0.5 - cell.Polarization*arr.MemoryWindow/2

	// Update domain state
	if cell.Polarization > 0.5 {
		cell.DomainState = DomainUp
	} else if cell.Polarization < -0.5 {
		cell.DomainState = DomainDown
	} else {
		cell.DomainState = DomainMixed
	}

	cell.CycleCount++
	return nil
}

// ReadCell reads a cell's conductance state
func (arr *FeNANDArray) ReadCell(layer, str, bl int) (float64, error) {
	if bl >= arr.NumBitLines || str >= arr.NumStrings || layer >= arr.NumLayers {
		return 0, fmt.Errorf("cell index out of range")
	}

	cell := arr.Strings[bl][str].Cells[layer]

	// Conductance depends on threshold and pass voltage
	passVoltage := arr.WordLines[layer]
	vt := cell.Threshold

	// Simple conductance model
	if passVoltage > vt {
		cell.Conductance = 1e-5 * (passVoltage - vt)
	} else {
		cell.Conductance = 1e-9 // Leakage
	}

	return cell.Conductance, nil
}

// GetAnalogWeight returns multi-level weight from cell
func (arr *FeNANDArray) GetAnalogWeight(layer, str, bl int) float64 {
	cell := arr.Strings[bl][str].Cells[layer]
	// Map polarization to weight levels
	levels := float64(1 << arr.BitsPerCell)
	quantized := math.Round((cell.Polarization + 1) / 2 * (levels - 1))
	return quantized / (levels - 1) * 2 - 1
}

// ============================================================================
// Domain Visualization (PFM-style)
// ============================================================================

// DomainPixel represents a pixel in domain visualization
type DomainPixel struct {
	X, Y         int
	Polarization float64
	Phase        float64 // PFM phase angle
	Amplitude    float64 // PFM amplitude
	DomainWall   bool    // Is domain wall
}

// DomainMap represents a 2D domain structure visualization
type DomainMap struct {
	Width        int
	Height       int
	Pixels       [][]*DomainPixel
	GrainBoundaries [][]bool
	TimeStep     float64
}

// NewDomainMap creates a domain visualization map
func NewDomainMap(width, height int) *DomainMap {
	dm := &DomainMap{
		Width:           width,
		Height:          height,
		Pixels:          make([][]*DomainPixel, height),
		GrainBoundaries: make([][]bool, height),
	}

	for y := 0; y < height; y++ {
		dm.Pixels[y] = make([]*DomainPixel, width)
		dm.GrainBoundaries[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			dm.Pixels[y][x] = &DomainPixel{
				X:            x,
				Y:            y,
				Polarization: 0,
				Phase:        0,
				Amplitude:    1,
			}
		}
	}

	return dm
}

// InitializeRandomDomains initializes random domain structure
func (dm *DomainMap) InitializeRandomDomains(grainSize int) {
	// Create grain structure
	numGrainsX := dm.Width / grainSize
	numGrainsY := dm.Height / grainSize

	grainPolarizations := make([][]float64, numGrainsY)
	for gy := 0; gy < numGrainsY; gy++ {
		grainPolarizations[gy] = make([]float64, numGrainsX)
		for gx := 0; gx < numGrainsX; gx++ {
			// Random initial polarization per grain
			if rand.Float64() > 0.5 {
				grainPolarizations[gy][gx] = 1
			} else {
				grainPolarizations[gy][gx] = -1
			}
		}
	}

	// Assign polarization to pixels
	for y := 0; y < dm.Height; y++ {
		for x := 0; x < dm.Width; x++ {
			gx := x / grainSize
			gy := y / grainSize
			if gx >= numGrainsX {
				gx = numGrainsX - 1
			}
			if gy >= numGrainsY {
				gy = numGrainsY - 1
			}

			dm.Pixels[y][x].Polarization = grainPolarizations[gy][gx]

			// Mark grain boundaries
			if x%grainSize == 0 || y%grainSize == 0 {
				dm.GrainBoundaries[y][x] = true
			}
		}
	}

	dm.identifyDomainWalls()
}

// identifyDomainWalls marks pixels at domain walls
func (dm *DomainMap) identifyDomainWalls() {
	for y := 0; y < dm.Height; y++ {
		for x := 0; x < dm.Width; x++ {
			p := dm.Pixels[y][x].Polarization
			isWall := false

			// Check neighbors
			neighbors := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
			for _, n := range neighbors {
				nx, ny := x+n[0], y+n[1]
				if nx >= 0 && nx < dm.Width && ny >= 0 && ny < dm.Height {
					if dm.Pixels[ny][nx].Polarization*p < 0 {
						isWall = true
						break
					}
				}
			}

			dm.Pixels[y][x].DomainWall = isWall
		}
	}
}

// ApplyElectricField simulates domain switching under field
func (dm *DomainMap) ApplyElectricField(field float64, coerciveField float64, dt float64) {
	dm.TimeStep += dt

	// Nucleation-limited switching model
	nucleationRate := math.Exp(-coerciveField / math.Abs(field))

	for y := 0; y < dm.Height; y++ {
		for x := 0; x < dm.Width; x++ {
			pixel := dm.Pixels[y][x]

			// Check if field exceeds coercive field
			if math.Abs(field) > coerciveField {
				targetPol := float64(1)
				if field < 0 {
					targetPol = -1
				}

				// Switching probability
				if pixel.Polarization != targetPol {
					// Domain wall velocity
					if pixel.DomainWall {
						// Faster switching at domain walls
						switchProb := nucleationRate * 10 * dt
						if rand.Float64() < switchProb {
							pixel.Polarization = targetPol
						}
					} else {
						// Nucleation in domain interior
						switchProb := nucleationRate * dt
						if rand.Float64() < switchProb {
							pixel.Polarization = targetPol
						}
					}
				}
			}

			// Update PFM-style visualization
			pixel.Phase = math.Atan2(pixel.Polarization, 0.1) * 180 / math.Pi
			pixel.Amplitude = math.Abs(pixel.Polarization)
		}
	}

	dm.identifyDomainWalls()
}

// GetPolarizationFraction returns fraction of up/down domains
func (dm *DomainMap) GetPolarizationFraction() (upFrac, downFrac float64) {
	upCount, downCount := 0, 0
	total := dm.Width * dm.Height

	for y := 0; y < dm.Height; y++ {
		for x := 0; x < dm.Width; x++ {
			if dm.Pixels[y][x].Polarization > 0 {
				upCount++
			} else {
				downCount++
			}
		}
	}

	return float64(upCount) / float64(total), float64(downCount) / float64(total)
}

// ============================================================================
// 3D Visualization for FeNAND
// ============================================================================

// Voxel3D represents a 3D voxel for visualization
type Voxel3D struct {
	X, Y, Z      int
	Value        float64
	Color        [3]float64 // RGB
	Visible      bool
	Type         string // "cell", "wordline", "channel", "oxide"
}

// FeNAND3DView represents a 3D visualization of FeNAND
type FeNAND3DView struct {
	Array        *FeNANDArray
	Voxels       [][][]*Voxel3D
	ResX, ResY, ResZ int
	ViewAngle    [3]float64 // Rotation angles
	Scale        float64
}

// NewFeNAND3DView creates a 3D view of FeNAND array
func NewFeNAND3DView(array *FeNANDArray, resolution int) *FeNAND3DView {
	view := &FeNAND3DView{
		Array: array,
		ResX:  array.NumBitLines * resolution,
		ResY:  array.NumStrings * resolution,
		ResZ:  array.NumLayers * resolution,
		Scale: 1.0,
	}

	// Initialize voxel grid
	view.Voxels = make([][][]*Voxel3D, view.ResZ)
	for z := 0; z < view.ResZ; z++ {
		view.Voxels[z] = make([][]*Voxel3D, view.ResY)
		for y := 0; y < view.ResY; y++ {
			view.Voxels[z][y] = make([]*Voxel3D, view.ResX)
			for x := 0; x < view.ResX; x++ {
				view.Voxels[z][y][x] = &Voxel3D{
					X: x, Y: y, Z: z,
					Visible: false,
					Type:    "oxide",
				}
			}
		}
	}

	view.updateVoxels()
	return view
}

// updateVoxels updates voxel values from array state
func (v *FeNAND3DView) updateVoxels() {
	resolution := v.ResX / v.Array.NumBitLines

	for bl := 0; bl < v.Array.NumBitLines; bl++ {
		for s := 0; s < v.Array.NumStrings; s++ {
			for l := 0; l < v.Array.NumLayers; l++ {
				cell := v.Array.Strings[bl][s].Cells[l]

				// Map cell to voxels
				baseX := bl * resolution
				baseY := s * resolution
				baseZ := l * resolution

				for dz := 0; dz < resolution; dz++ {
					for dy := 0; dy < resolution; dy++ {
						for dx := 0; dx < resolution; dx++ {
							x := baseX + dx
							y := baseY + dy
							z := baseZ + dz

							if x < v.ResX && y < v.ResY && z < v.ResZ {
								voxel := v.Voxels[z][y][x]
								voxel.Value = cell.Polarization
								voxel.Visible = true
								voxel.Type = "cell"

								// Color based on polarization
								if cell.Polarization > 0 {
									// Blue for P+
									voxel.Color = [3]float64{0.2, 0.2, 0.8 + 0.2*cell.Polarization}
								} else {
									// Red for P-
									voxel.Color = [3]float64{0.8 - 0.2*cell.Polarization, 0.2, 0.2}
								}
							}
						}
					}
				}
			}
		}
	}
}

// RenderSlice renders a 2D slice of the 3D view
func (v *FeNAND3DView) RenderSlice(axis string, position int) [][]float64 {
	var slice [][]float64

	switch axis {
	case "xy": // Top view
		if position >= v.ResZ {
			position = v.ResZ - 1
		}
		slice = make([][]float64, v.ResY)
		for y := 0; y < v.ResY; y++ {
			slice[y] = make([]float64, v.ResX)
			for x := 0; x < v.ResX; x++ {
				slice[y][x] = v.Voxels[position][y][x].Value
			}
		}

	case "xz": // Front view
		if position >= v.ResY {
			position = v.ResY - 1
		}
		slice = make([][]float64, v.ResZ)
		for z := 0; z < v.ResZ; z++ {
			slice[z] = make([]float64, v.ResX)
			for x := 0; x < v.ResX; x++ {
				slice[z][x] = v.Voxels[z][position][x].Value
			}
		}

	case "yz": // Side view
		if position >= v.ResX {
			position = v.ResX - 1
		}
		slice = make([][]float64, v.ResZ)
		for z := 0; z < v.ResZ; z++ {
			slice[z] = make([]float64, v.ResY)
			for y := 0; y < v.ResY; y++ {
				slice[z][y] = v.Voxels[z][y][position].Value
			}
		}
	}

	return slice
}

// ============================================================================
// Spiking Neural Network with Ferroelectric Synapses
// ============================================================================

// Spike represents a neural spike event
type Spike struct {
	NeuronID  int
	Timestamp float64 // ms
	Layer     int
}

// LIFNeuronConfig configures a LIF neuron
type LIFNeuronConfig struct {
	Threshold       float64 // Spike threshold (mV)
	RestPotential   float64 // Resting potential (mV)
	ResetPotential  float64 // Reset after spike (mV)
	LeakConstant    float64 // Membrane time constant (ms)
	RefractoryTime  float64 // Refractory period (ms)
	Capacitance     float64 // Membrane capacitance (pF)
}

// DefaultLIFConfig returns default LIF neuron parameters
func DefaultLIFConfig() LIFNeuronConfig {
	return LIFNeuronConfig{
		Threshold:      -55,  // mV
		RestPotential:  -70,  // mV
		ResetPotential: -75,  // mV
		LeakConstant:   20,   // ms
		RefractoryTime: 2,    // ms
		Capacitance:    100,  // pF
	}
}

// LIFNeuron implements a Leaky Integrate-and-Fire neuron
type LIFNeuron struct {
	ID              int
	Config          LIFNeuronConfig
	MembranePotential float64
	LastSpikeTime   float64
	InRefractory    bool
	InputCurrent    float64
	SpikeHistory    []float64
	mu              sync.Mutex
}

// NewLIFNeuron creates a new LIF neuron
func NewLIFNeuron(id int, config LIFNeuronConfig) *LIFNeuron {
	return &LIFNeuron{
		ID:                id,
		Config:            config,
		MembranePotential: config.RestPotential,
		LastSpikeTime:     -1000,
		SpikeHistory:      make([]float64, 0),
	}
}

// Update updates neuron state for one time step
func (n *LIFNeuron) Update(dt float64, currentTime float64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Check refractory period
	if currentTime-n.LastSpikeTime < n.Config.RefractoryTime {
		n.InRefractory = true
		n.MembranePotential = n.Config.ResetPotential
		return false
	}
	n.InRefractory = false

	// Leaky integration
	// dV/dt = -(V - V_rest)/tau + I/C
	leak := -(n.MembranePotential - n.Config.RestPotential) / n.Config.LeakConstant
	input := n.InputCurrent / n.Config.Capacitance * 1000 // Convert to mV/ms

	n.MembranePotential += (leak + input) * dt

	// Reset input current
	n.InputCurrent = 0

	// Check for spike
	if n.MembranePotential >= n.Config.Threshold {
		n.MembranePotential = n.Config.ResetPotential
		n.LastSpikeTime = currentTime
		n.SpikeHistory = append(n.SpikeHistory, currentTime)
		return true
	}

	return false
}

// InjectCurrent injects current into the neuron
func (n *LIFNeuron) InjectCurrent(current float64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.InputCurrent += current
}

// FerroelectricSynapse implements a FeFET-based synapse with STDP
type FerroelectricSynapse struct {
	PreNeuronID     int
	PostNeuronID    int
	Weight          float64 // -1 to +1 (polarization)
	Conductance     float64 // nS

	// FeFET parameters
	Pr              float64 // Remanent polarization
	CoerciveVoltage float64 // V
	SwitchingTime   float64 // ns

	// STDP parameters
	TauPlus         float64 // Potentiation time constant (ms)
	TauMinus        float64 // Depression time constant (ms)
	APlus           float64 // Max potentiation
	AMinus          float64 // Max depression

	LastPreSpike    float64
	LastPostSpike   float64
}

// NewFerroelectricSynapse creates a new FeFET synapse
func NewFerroelectricSynapse(preID, postID int) *FerroelectricSynapse {
	return &FerroelectricSynapse{
		PreNeuronID:     preID,
		PostNeuronID:    postID,
		Weight:          0,
		Conductance:     100, // nS
		Pr:              25,  // µC/cm²
		CoerciveVoltage: 1.5, // V
		SwitchingTime:   10,  // ns
		TauPlus:         20,  // ms
		TauMinus:        20,  // ms
		APlus:           0.1,
		AMinus:          0.12,
		LastPreSpike:    -1000,
		LastPostSpike:   -1000,
	}
}

// ApplySTDP applies spike-timing-dependent plasticity
func (s *FerroelectricSynapse) ApplySTDP(preTime, postTime float64) {
	dt := postTime - preTime

	if dt > 0 {
		// Pre before post: potentiation (LTP)
		deltaW := s.APlus * math.Exp(-dt/s.TauPlus)
		s.Weight += deltaW
	} else if dt < 0 {
		// Post before pre: depression (LTD)
		deltaW := -s.AMinus * math.Exp(dt/s.TauMinus)
		s.Weight += deltaW
	}

	// Clamp weight to [-1, 1]
	if s.Weight > 1 {
		s.Weight = 1
	}
	if s.Weight < -1 {
		s.Weight = -1
	}

	// Update conductance based on weight
	s.Conductance = 100 * (s.Weight + 1) / 2 // 0 to 100 nS
}

// GetCurrent returns synaptic current for a given voltage
func (s *FerroelectricSynapse) GetCurrent(voltage float64) float64 {
	// I = g * V
	return s.Conductance * voltage / 1000 // pA
}

// SpikingNeuralNetwork implements an all-ferroelectric SNN
type SpikingNeuralNetwork struct {
	Neurons         [][]*LIFNeuron           // Layers of neurons
	Synapses        [][]*FerroelectricSynapse // Layer-to-layer synapses
	InputSize       int
	HiddenSizes     []int
	OutputSize      int
	TimeStep        float64 // ms
	CurrentTime     float64
	SpikeHistory    []Spike
	LearningEnabled bool
	mu              sync.Mutex
}

// NewSpikingNeuralNetwork creates a new SNN
func NewSpikingNeuralNetwork(inputSize int, hiddenSizes []int, outputSize int) *SpikingNeuralNetwork {
	snn := &SpikingNeuralNetwork{
		InputSize:       inputSize,
		HiddenSizes:     hiddenSizes,
		OutputSize:      outputSize,
		TimeStep:        0.1, // ms
		LearningEnabled: true,
		SpikeHistory:    make([]Spike, 0),
	}

	// Create neuron layers
	layerSizes := append([]int{inputSize}, hiddenSizes...)
	layerSizes = append(layerSizes, outputSize)

	snn.Neurons = make([][]*LIFNeuron, len(layerSizes))
	config := DefaultLIFConfig()

	neuronID := 0
	for l, size := range layerSizes {
		snn.Neurons[l] = make([]*LIFNeuron, size)
		for i := 0; i < size; i++ {
			snn.Neurons[l][i] = NewLIFNeuron(neuronID, config)
			neuronID++
		}
	}

	// Create synapses between layers
	snn.Synapses = make([][]*FerroelectricSynapse, len(layerSizes)-1)
	for l := 0; l < len(layerSizes)-1; l++ {
		preSize := layerSizes[l]
		postSize := layerSizes[l+1]
		snn.Synapses[l] = make([]*FerroelectricSynapse, preSize*postSize)

		for pre := 0; pre < preSize; pre++ {
			for post := 0; post < postSize; post++ {
				idx := pre*postSize + post
				preID := snn.Neurons[l][pre].ID
				postID := snn.Neurons[l+1][post].ID
				syn := NewFerroelectricSynapse(preID, postID)
				// Initialize with random weights
				syn.Weight = rand.Float64()*2 - 1
				snn.Synapses[l][idx] = syn
			}
		}
	}

	return snn
}

// Step advances the network by one time step
func (snn *SpikingNeuralNetwork) Step() {
	snn.mu.Lock()
	defer snn.mu.Unlock()

	snn.CurrentTime += snn.TimeStep

	// Process each layer
	for l := range snn.Neurons {
		// Collect spikes from previous layer
		if l > 0 {
			prevLayer := l - 1
			synLayer := snn.Synapses[prevLayer]

			for post := range snn.Neurons[l] {
				postNeuron := snn.Neurons[l][post]
				totalCurrent := float64(0)

				for pre := range snn.Neurons[prevLayer] {
					preNeuron := snn.Neurons[prevLayer][pre]
					synIdx := pre*len(snn.Neurons[l]) + post
					syn := synLayer[synIdx]

					// Check if pre-neuron spiked recently
					if len(preNeuron.SpikeHistory) > 0 {
						lastSpike := preNeuron.SpikeHistory[len(preNeuron.SpikeHistory)-1]
						if snn.CurrentTime-lastSpike < snn.TimeStep*2 {
							// Add synaptic current
							current := syn.GetCurrent(20) // 20mV driving force
							totalCurrent += current
						}
					}
				}

				postNeuron.InjectCurrent(totalCurrent)
			}
		}

		// Update all neurons in layer
		for _, neuron := range snn.Neurons[l] {
			spiked := neuron.Update(snn.TimeStep, snn.CurrentTime)
			if spiked {
				snn.SpikeHistory = append(snn.SpikeHistory, Spike{
					NeuronID:  neuron.ID,
					Timestamp: snn.CurrentTime,
					Layer:     l,
				})

				// Apply STDP if learning enabled
				if snn.LearningEnabled && l > 0 {
					snn.applySTDPForNeuron(neuron, l)
				}
			}
		}
	}
}

// applySTDPForNeuron applies STDP for a post-synaptic spike
func (snn *SpikingNeuralNetwork) applySTDPForNeuron(postNeuron *LIFNeuron, layer int) {
	prevLayer := layer - 1
	synLayer := snn.Synapses[prevLayer]
	postIdx := -1

	// Find post neuron index in its layer
	for i, n := range snn.Neurons[layer] {
		if n.ID == postNeuron.ID {
			postIdx = i
			break
		}
	}

	if postIdx < 0 {
		return
	}

	// Update synapses from all pre-neurons
	for pre, preNeuron := range snn.Neurons[prevLayer] {
		synIdx := pre*len(snn.Neurons[layer]) + postIdx
		syn := synLayer[synIdx]

		// Get last spike times
		preTime := float64(-1000)
		if len(preNeuron.SpikeHistory) > 0 {
			preTime = preNeuron.SpikeHistory[len(preNeuron.SpikeHistory)-1]
		}
		postTime := snn.CurrentTime

		syn.ApplySTDP(preTime, postTime)
	}
}

// Present presents input spikes to the network
func (snn *SpikingNeuralNetwork) Present(inputRates []float64, duration float64) {
	snn.mu.Lock()
	defer snn.mu.Unlock()

	// Generate Poisson spike trains based on rates
	steps := int(duration / snn.TimeStep)

	for step := 0; step < steps; step++ {
		// Inject input spikes
		for i, rate := range inputRates {
			if i < len(snn.Neurons[0]) {
				// Poisson probability
				prob := rate * snn.TimeStep / 1000 // rate in Hz
				if rand.Float64() < prob {
					// Inject large current to cause spike
					snn.Neurons[0][i].InjectCurrent(500) // pA
				}
			}
		}

		snn.mu.Unlock()
		snn.Step()
		snn.mu.Lock()
	}
}

// GetOutputSpikeCounts returns spike counts for output neurons
func (snn *SpikingNeuralNetwork) GetOutputSpikeCounts(window float64) []int {
	snn.mu.Lock()
	defer snn.mu.Unlock()

	counts := make([]int, snn.OutputSize)
	outputLayer := len(snn.Neurons) - 1
	startTime := snn.CurrentTime - window

	for _, spike := range snn.SpikeHistory {
		if spike.Layer == outputLayer && spike.Timestamp >= startTime {
			// Find neuron index in output layer
			for i, n := range snn.Neurons[outputLayer] {
				if n.ID == spike.NeuronID {
					counts[i]++
					break
				}
			}
		}
	}

	return counts
}

// Classify returns the predicted class based on spike counts
func (snn *SpikingNeuralNetwork) Classify() int {
	counts := snn.GetOutputSpikeCounts(100) // 100ms window

	maxCount := 0
	maxIdx := 0
	for i, c := range counts {
		if c > maxCount {
			maxCount = c
			maxIdx = i
		}
	}

	return maxIdx
}

// ============================================================================
// SNN Training and Evaluation
// ============================================================================

// SNNTrainer trains an SNN on a dataset
type SNNTrainer struct {
	Network         *SpikingNeuralNetwork
	LearningRate    float64
	NumEpochs       int
	PresentationTime float64 // ms per sample
}

// NewSNNTrainer creates a new SNN trainer
func NewSNNTrainer(network *SpikingNeuralNetwork) *SNNTrainer {
	return &SNNTrainer{
		Network:          network,
		LearningRate:     0.01,
		NumEpochs:        10,
		PresentationTime: 100, // ms
	}
}

// ConvertToSpikeRates converts image data to spike rates
func (t *SNNTrainer) ConvertToSpikeRates(data []float64, maxRate float64) []float64 {
	rates := make([]float64, len(data))
	for i, v := range data {
		// Map pixel value to spike rate
		rates[i] = v * maxRate // 0-1 -> 0-maxRate Hz
	}
	return rates
}

// TrainSample trains on a single sample with target label
func (t *SNNTrainer) TrainSample(data []float64, label int) {
	// Convert to spike rates (max 100 Hz)
	rates := t.ConvertToSpikeRates(data, 100)

	// Present to network
	t.Network.Present(rates, t.PresentationTime)

	// Target neuron should fire more
	// This is simplified - real implementation would use R-STDP or reward modulation
}

// Evaluate evaluates accuracy on test data
func (t *SNNTrainer) Evaluate(testData [][]float64, testLabels []int) float64 {
	correct := 0
	total := len(testLabels)

	for i, data := range testData {
		// Reset network state
		t.Network.CurrentTime = 0
		t.Network.SpikeHistory = make([]Spike, 0)
		for _, layer := range t.Network.Neurons {
			for _, n := range layer {
				n.MembranePotential = n.Config.RestPotential
				n.SpikeHistory = make([]float64, 0)
			}
		}

		// Present sample
		rates := t.ConvertToSpikeRates(data, 100)
		t.Network.LearningEnabled = false
		t.Network.Present(rates, t.PresentationTime)

		// Get prediction
		pred := t.Network.Classify()
		if pred == testLabels[i] {
			correct++
		}
	}

	return float64(correct) / float64(total)
}

// ============================================================================
// SNN Benchmark and Demo
// ============================================================================

// SNNBenchmark benchmarks SNN performance
type SNNBenchmark struct {
	Network         *SpikingNeuralNetwork
	TotalSpikes     int
	TotalTime       float64 // ms
	EnergyPerSpike  float64 // fJ
	Results         []SNNBenchmarkResult
}

// SNNBenchmarkResult holds benchmark results
type SNNBenchmarkResult struct {
	Dataset       string
	Accuracy      float64
	SpikesPerInf  int
	InferenceTime float64 // ms
	EnergyMJ      float64
	SynapticOps   int64
}

// NewSNNBenchmark creates a new SNN benchmark
func NewSNNBenchmark(network *SpikingNeuralNetwork) *SNNBenchmark {
	return &SNNBenchmark{
		Network:        network,
		EnergyPerSpike: 2, // fJ per spike (from literature)
		Results:        make([]SNNBenchmarkResult, 0),
	}
}

// RunMNISTBenchmark runs MNIST classification benchmark
func (sb *SNNBenchmark) RunMNISTBenchmark(numSamples int) SNNBenchmarkResult {
	// Generate synthetic MNIST-like data
	testData := make([][]float64, numSamples)
	testLabels := make([]int, numSamples)

	for i := 0; i < numSamples; i++ {
		testData[i] = make([]float64, sb.Network.InputSize)
		for j := range testData[i] {
			testData[i][j] = rand.Float64()
		}
		testLabels[i] = rand.Intn(sb.Network.OutputSize)
	}

	// Run inference
	startTime := time.Now()
	trainer := NewSNNTrainer(sb.Network)
	accuracy := trainer.Evaluate(testData, testLabels)
	elapsed := time.Since(startTime).Seconds() * 1000 // ms

	// Count total spikes
	totalSpikes := len(sb.Network.SpikeHistory)
	spikesPerInf := totalSpikes / numSamples

	// Calculate energy
	energyMJ := float64(totalSpikes) * sb.EnergyPerSpike / 1e9

	// Synaptic operations
	var synOps int64
	for _, layer := range sb.Network.Synapses {
		synOps += int64(len(layer)) * int64(spikesPerInf)
	}

	result := SNNBenchmarkResult{
		Dataset:       "MNIST-synthetic",
		Accuracy:      accuracy,
		SpikesPerInf:  spikesPerInf,
		InferenceTime: elapsed / float64(numSamples),
		EnergyMJ:      energyMJ / float64(numSamples),
		SynapticOps:   synOps,
	}

	sb.Results = append(sb.Results, result)
	return result
}

// CompareToDNN compares SNN to equivalent DNN
func (sb *SNNBenchmark) CompareToDNN() map[string]float64 {
	// Estimate equivalent DNN energy
	dnnMACs := int64(0)
	for l := 0; l < len(sb.Network.Neurons)-1; l++ {
		preSize := len(sb.Network.Neurons[l])
		postSize := len(sb.Network.Neurons[l+1])
		dnnMACs += int64(preSize * postSize)
	}

	dnnEnergyPerMAC := 0.5e-12 // 0.5 pJ/MAC for digital
	snnEnergyPerSpike := 2e-15 // 2 fJ/spike for SNN

	avgSpikesPerInf := 0
	if len(sb.Results) > 0 {
		avgSpikesPerInf = sb.Results[len(sb.Results)-1].SpikesPerInf
	}

	dnnEnergy := float64(dnnMACs) * dnnEnergyPerMAC
	snnEnergy := float64(avgSpikesPerInf) * snnEnergyPerSpike * float64(len(sb.Network.Synapses))

	return map[string]float64{
		"dnn_energy_j":        dnnEnergy,
		"snn_energy_j":        snnEnergy,
		"energy_reduction":    (dnnEnergy - snnEnergy) / dnnEnergy * 100,
		"dnn_ops":             float64(dnnMACs),
		"snn_ops":             float64(avgSpikesPerInf * len(sb.Network.Synapses)),
		"sparsity_advantage":  float64(dnnMACs) / float64(avgSpikesPerInf*len(sb.Network.Synapses)),
	}
}

// PrintSummary prints benchmark summary
func (sb *SNNBenchmark) PrintSummary() string {
	output := "All-Ferroelectric SNN Benchmark Summary\n"
	output += "========================================\n\n"

	for _, r := range sb.Results {
		output += fmt.Sprintf("Dataset: %s\n", r.Dataset)
		output += fmt.Sprintf("  Accuracy: %.2f%%\n", r.Accuracy*100)
		output += fmt.Sprintf("  Spikes/inference: %d\n", r.SpikesPerInf)
		output += fmt.Sprintf("  Inference time: %.3f ms\n", r.InferenceTime)
		output += fmt.Sprintf("  Energy: %.6f mJ\n", r.EnergyMJ)
		output += "\n"
	}

	comparison := sb.CompareToDNN()
	output += "DNN Comparison:\n"
	output += fmt.Sprintf("  Energy reduction: %.1f%%\n", comparison["energy_reduction"])
	output += fmt.Sprintf("  Sparsity advantage: %.1fx\n", comparison["sparsity_advantage"])

	return output
}

// ============================================================================
// Utility Functions
// ============================================================================

// SortSpikesByTime sorts spikes chronologically
func SortSpikesByTime(spikes []Spike) []Spike {
	sorted := make([]Spike, len(spikes))
	copy(sorted, spikes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp < sorted[j].Timestamp
	})
	return sorted
}

// ComputeSpikeRate computes average spike rate
func ComputeSpikeRate(spikes []Spike, duration float64) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(len(spikes)) / duration * 1000 // Hz
}
