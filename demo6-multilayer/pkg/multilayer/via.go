package multilayer

import (
	"math"
)

// Via represents a vertical interconnect between layers.
type Via struct {
	X           int     // X position in the layer
	FromLayer   int     // Source layer index
	ToLayer     int     // Destination layer index
	Resistance  float64 // Via resistance (Ohms)
	Capacitance float64 // Via capacitance (fF)
	Current     float64 // Current flowing through via (µA)
}

// ViaArray represents all vias between two adjacent layers.
type ViaArray struct {
	Vias        []*Via  // Individual vias
	FromLayer   int     // Source layer index
	ToLayer     int     // Destination layer index
	Count       int     // Number of vias
	Pitch       float64 // Via pitch in nm
	Diameter    float64 // Via diameter in nm
	AspectRatio float64 // Height/diameter ratio
}

// ViaNetwork manages all via connections in a stack.
type ViaNetwork struct {
	Arrays      []*ViaArray // Via arrays between each layer pair
	TotalVias   int         // Total number of vias
	TotalLength float64     // Total via length in µm
}

// NewViaArray creates a via array between two layers.
func NewViaArray(fromLayer, toLayer, count int) *ViaArray {
	vias := make([]*Via, count)
	for i := range vias {
		vias[i] = &Via{
			X:           i,
			FromLayer:   fromLayer,
			ToLayer:     toLayer,
			Resistance:  10.0,  // 10 Ohms typical
			Capacitance: 0.1,   // 0.1 fF typical
			Current:     0.0,
		}
	}

	return &ViaArray{
		Vias:        vias,
		FromLayer:   fromLayer,
		ToLayer:     toLayer,
		Count:       count,
		Pitch:       90.0, // 90nm pitch (2x cell pitch)
		Diameter:    40.0, // 40nm diameter
		AspectRatio: 1.25, // 50nm height / 40nm diameter
	}
}

// NewViaNetwork creates a via network for a stack.
func NewViaNetwork(stack *Stack) *ViaNetwork {
	if len(stack.Layers) < 2 {
		return &ViaNetwork{
			Arrays:    make([]*ViaArray, 0),
			TotalVias: 0,
		}
	}

	arrays := make([]*ViaArray, len(stack.Layers)-1)
	totalVias := 0
	totalLength := 0.0

	for i := 0; i < len(stack.Layers)-1; i++ {
		// Vias connect layer i outputs to layer i+1 inputs
		viaCount := stack.Layers[i].Cols
		arrays[i] = NewViaArray(i, i+1, viaCount)
		totalVias += viaCount
		// Via length = layer height
		totalLength += float64(viaCount) * stack.LayerHeight / 1000.0 // µm
	}

	return &ViaNetwork{
		Arrays:      arrays,
		TotalVias:   totalVias,
		TotalLength: totalLength,
	}
}

// TotalResistance returns total via resistance in the network.
func (v *ViaNetwork) TotalResistance() float64 {
	total := 0.0
	for _, array := range v.Arrays {
		for _, via := range array.Vias {
			total += via.Resistance
		}
	}
	return total
}

// TotalCapacitance returns total via capacitance in fF.
func (v *ViaNetwork) TotalCapacitance() float64 {
	total := 0.0
	for _, array := range v.Arrays {
		for _, via := range array.Vias {
			total += via.Capacitance
		}
	}
	return total
}

// PropagationDelay estimates signal propagation delay through vias.
func (v *ViaNetwork) PropagationDelay() float64 {
	// RC delay: τ = R * C
	totalDelay := 0.0
	for _, array := range v.Arrays {
		for _, via := range array.Vias {
			// Convert fF to F for calculation
			delay := via.Resistance * via.Capacitance * 1e-15
			totalDelay += delay
		}
	}
	// Return in ps
	return totalDelay * 1e12
}

// SetViaCurrents sets current flow based on layer activations.
func (v *ViaNetwork) SetViaCurrents(activations [][]float64) {
	for i, array := range v.Arrays {
		if i < len(activations) {
			for j, via := range array.Vias {
				if j < len(activations[i]) {
					// Current proportional to activation (simplified)
					via.Current = activations[i][j] * 10.0 // Scale to µA
				}
			}
		}
	}
}

// MaxCurrent returns the maximum current through any via.
func (v *ViaNetwork) MaxCurrent() float64 {
	maxI := 0.0
	for _, array := range v.Arrays {
		for _, via := range array.Vias {
			if math.Abs(via.Current) > maxI {
				maxI = math.Abs(via.Current)
			}
		}
	}
	return maxI
}

// ViaStats contains statistics about the via network.
type ViaStats struct {
	TotalVias        int
	ViaArrays        int
	TotalLength      float64 // µm
	AvgResistance    float64 // Ohms
	TotalCapacitance float64 // fF
	PropagationDelay float64 // ps
	ViaDensity       float64 // Vias per µm²
}

// GetStats returns via network statistics.
func (v *ViaNetwork) GetStats(footprintArea float64) ViaStats {
	avgR := 0.0
	if v.TotalVias > 0 {
		avgR = v.TotalResistance() / float64(v.TotalVias)
	}

	density := 0.0
	if footprintArea > 0 {
		density = float64(v.TotalVias) / footprintArea
	}

	return ViaStats{
		TotalVias:        v.TotalVias,
		ViaArrays:        len(v.Arrays),
		TotalLength:      v.TotalLength,
		AvgResistance:    avgR,
		TotalCapacitance: v.TotalCapacitance(),
		PropagationDelay: v.PropagationDelay(),
		ViaDensity:       density,
	}
}

// ViaCurrentDistribution returns histogram of via currents.
func (v *ViaNetwork) ViaCurrentDistribution(bins int) []int {
	if v.TotalVias == 0 {
		return make([]int, bins)
	}

	// Find max current
	maxI := v.MaxCurrent()
	if maxI == 0 {
		maxI = 1.0
	}

	histogram := make([]int, bins)
	binWidth := maxI / float64(bins)

	for _, array := range v.Arrays {
		for _, via := range array.Vias {
			bin := int(math.Abs(via.Current) / binWidth)
			if bin >= bins {
				bin = bins - 1
			}
			histogram[bin]++
		}
	}

	return histogram
}

// EstimateViaYield estimates manufacturing yield based on via count.
func (v *ViaNetwork) EstimateViaYield(defectDensity float64) float64 {
	// Poisson yield model: Y = exp(-D * A)
	// Where D = defect density, A = total via area
	if v.TotalVias == 0 {
		return 1.0
	}

	// Assume average via diameter of 40nm
	viaDiameter := 40e-9 // meters
	viaArea := math.Pi * math.Pow(viaDiameter/2, 2) * float64(v.TotalVias)

	// Convert defect density from defects/cm² to defects/m²
	defectDensityM2 := defectDensity * 1e4

	return math.Exp(-defectDensityM2 * viaArea)
}
