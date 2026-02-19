package system

// DSEConfig parameterises a design-space exploration sweep over crossbar array
// configurations. Fields use yaml tags so configs can be loaded from YAML files.
type DSEConfig struct {
	// ArraySizes lists square array edge lengths to sweep (e.g. [32, 64, 128]).
	ArraySizes []int `yaml:"array_sizes"`

	// ADCBits lists ADC resolutions to sweep (e.g. [4, 6, 8]).
	ADCBits []int `yaml:"adc_bits"`

	// CellBits lists per-cell bit depths to sweep (e.g. [1, 2, 4]).
	// Used to derive the number of conductance levels (2^CellBits).
	CellBits []int `yaml:"cell_bits"`

	// TechNode selects the fabrication node for area and latency estimates.
	TechNode TechnologyNode `yaml:"tech_node"`

	// DeviceType selects the cell type for area estimates.
	DeviceType CellType `yaml:"device_type"`

	// Metrics lists which output metrics to populate (currently informational;
	// all metrics are always computed). Example: ["area", "energy", "latency"].
	Metrics []string `yaml:"metrics"`
}

// DSEResult holds the estimated system metrics for one point in the design space.
type DSEResult struct {
	ArraySize int // edge length of the square array (cells)
	ADCBits   int // ADC resolution (bits)
	CellBits  int // per-cell storage depth (bits)

	AreaUM2   float64 // total chip area estimate (µm²)
	EnergyPJ  float64 // MVM cycle energy estimate (pJ)
	LatencyNS float64 // MVM pipeline latency estimate (ns)
	PowerUW   float64 // total power at 100 MHz (µW)
}

// RunDSE sweeps all combinations of cfg.ArraySizes × cfg.ADCBits × cfg.CellBits
// and returns one DSEResult per combination.
//
// Default values are applied for empty slices:
//   - ArraySizes → [32, 64, 128]
//   - ADCBits    → [4, 6, 8]
//   - CellBits   → [1, 2, 4]
//   - TechNode   → Node65nm
//   - DeviceType → CellFeFET
func RunDSE(cfg DSEConfig) []DSEResult {
	if len(cfg.ArraySizes) == 0 {
		cfg.ArraySizes = []int{32, 64, 128}
	}
	if len(cfg.ADCBits) == 0 {
		cfg.ADCBits = []int{4, 6, 8}
	}
	if len(cfg.CellBits) == 0 {
		cfg.CellBits = []int{1, 2, 4}
	}
	if cfg.TechNode == "" {
		cfg.TechNode = Node65nm
	}
	if cfg.DeviceType == "" {
		cfg.DeviceType = CellFeFET
	}

	const freqMHz = 100.0 // evaluation frequency for power estimate

	results := make([]DSEResult, 0, len(cfg.ArraySizes)*len(cfg.ADCBits)*len(cfg.CellBits))

	for _, sz := range cfg.ArraySizes {
		for _, adcBits := range cfg.ADCBits {
			for _, cellBits := range cfg.CellBits {
				levels := 1 << cellBits // 2^cellBits conductance levels

				area := NewCrossbarAreaModel(sz, sz, cfg.TechNode, cfg.DeviceType)
				lat := NewLatencyModel(sz, sz, cfg.TechNode)

				// Energy: one MVM = sz DAC conversions + sz² MACs + sz ADC conversions.
				macEnergy := float64(sz*sz) * MACEnergyJ(levels)
				dacEnergy := float64(sz) * DefaultEnergyPerDACJ
				adcEnergy := float64(sz) * DefaultEnergyPerADCJ
				totalEnergyPJ := (macEnergy + dacEnergy + adcEnergy) * 1e12

				// Power: leakage ~ 1 pA/cell typical for FeFET at 1 V supply.
				pwr := NewPowerModel(sz, sz, 1.0, 1e-12)

				results = append(results, DSEResult{
					ArraySize: sz,
					ADCBits:   adcBits,
					CellBits:  cellBits,
					AreaUM2:   area.TotalAreaUM2(adcBits),
					EnergyPJ:  totalEnergyPJ,
					LatencyNS: lat.TotalPipelineNS(adcBits),
					PowerUW:   pwr.TotalPowerUW(freqMHz),
				})
			}
		}
	}

	return results
}
