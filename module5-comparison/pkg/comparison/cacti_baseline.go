// cacti_baseline.go provides CACTI-based SRAM/DRAM energy baselines for
// normalizing FeCIM comparison metrics.
//
// CACTI references:
//   - HP CACTI 7.0 (Balasubramonian et al., 2017)
//   - CACTI-P: Architecture-Level Modeling for SRAM-based Structures
//   - "CACTI 6.0: A Tool to Model Large Caches" (HP Labs, 2009)
//
// All energy values are from published CACTI simulation outputs or derived
// from peer-reviewed literature. They are NOT validated device measurements.
package comparison

import (
	"fmt"
	"strings"
)

// CACTIBaseline stores energy/area metrics from CACTI for comparison.
type CACTIBaseline struct {
	Technology    string  // e.g., "45nm", "28nm", "7nm"
	MemoryType    string  // "SRAM", "DRAM", "eDRAM"
	Capacity      string  // e.g., "256KB", "1MB"
	ReadEnergyPJ  float64 // Energy per read access (pJ)
	WriteEnergyPJ float64 // Energy per write access (pJ)
	AreaMm2       float64 // Total area (mm²)
	AccessTimeNs  float64 // Access latency (ns)
	LeakageMW     float64 // Leakage power (mW)
	Source        string  // "CACTI 7.0" or specific paper citation
}

// CACTINormalizedComparison normalizes FeCIM metrics against CACTI baseline.
type CACTINormalizedComparison struct {
	Baseline         CACTIBaseline
	FeCIM_pJperMAC   float64 // FeCIM energy per MAC operation (pJ)
	SRAM_pJperMAC    float64 // SRAM energy per equivalent MAC (pJ)
	DRAM_pJperMAC    float64 // DRAM energy per equivalent MAC (pJ)
	EnergyRatio_SRAM float64 // FeCIM / SRAM (< 1 means FeCIM better)
	EnergyRatio_DRAM float64 // FeCIM / DRAM (< 1 means FeCIM better)
	AreaRatio_SRAM   float64 // FeCIM area / SRAM area (< 1 means FeCIM smaller)
	Source           string  // Data source attribution
}

// DefaultCACTIBaselines returns published CACTI numbers for 28nm and 45nm
// SRAM and DRAM configurations.
//
// Values sourced from:
//   - HP CACTI 7.0 simulation outputs (Balasubramonian et al., 2017)
//   - Muralimanohar, Balasubramonian & Jouppi, "CACTI 6.0" (HP Labs, 2009)
//
// These are simulation-based numbers, not silicon measurements.
func DefaultCACTIBaselines() []CACTIBaseline {
	return []CACTIBaseline{
		// 45nm SRAM — CACTI 7.0 typical configuration
		{
			Technology:    "45nm",
			MemoryType:    "SRAM",
			Capacity:      "256KB",
			ReadEnergyPJ:  120.0, // ~120 pJ per read (CACTI 7.0, 256KB, 45nm)
			WriteEnergyPJ: 130.0, // ~130 pJ per write
			AreaMm2:       0.48,  // ~0.48 mm²
			AccessTimeNs:  1.8,   // ~1.8 ns access time
			LeakageMW:     45.0,  // ~45 mW leakage
			Source:        "CACTI 7.0, 45nm, 256KB SRAM",
		},
		// 45nm DRAM — CACTI 7.0 typical configuration
		{
			Technology:    "45nm",
			MemoryType:    "DRAM",
			Capacity:      "256KB",
			ReadEnergyPJ:  250.0, // ~250 pJ per read (includes row buffer)
			WriteEnergyPJ: 280.0, // ~280 pJ per write
			AreaMm2:       0.12,  // ~0.12 mm² (denser than SRAM)
			AccessTimeNs:  8.0,   // ~8 ns access time (slower)
			LeakageMW:     15.0,  // ~15 mW leakage (needs refresh)
			Source:        "CACTI 7.0, 45nm, 256KB DRAM",
		},
		// 28nm SRAM — CACTI 7.0 scaled configuration
		{
			Technology:    "28nm",
			MemoryType:    "SRAM",
			Capacity:      "256KB",
			ReadEnergyPJ:  65.0, // ~65 pJ per read (28nm scaling)
			WriteEnergyPJ: 72.0, // ~72 pJ per write
			AreaMm2:       0.22, // ~0.22 mm²
			AccessTimeNs:  1.2,  // ~1.2 ns access time
			LeakageMW:     60.0, // ~60 mW leakage (increases with scaling)
			Source:        "CACTI 7.0, 28nm, 256KB SRAM",
		},
		// 28nm DRAM — CACTI 7.0 scaled configuration
		{
			Technology:    "28nm",
			MemoryType:    "DRAM",
			Capacity:      "256KB",
			ReadEnergyPJ:  140.0, // ~140 pJ per read
			WriteEnergyPJ: 160.0, // ~160 pJ per write
			AreaMm2:       0.06,  // ~0.06 mm²
			AccessTimeNs:  6.5,   // ~6.5 ns access time
			LeakageMW:     20.0,  // ~20 mW leakage
			Source:        "CACTI 7.0, 28nm, 256KB DRAM",
		},
		// 45nm SRAM — 1MB configuration for larger arrays
		{
			Technology:    "45nm",
			MemoryType:    "SRAM",
			Capacity:      "1MB",
			ReadEnergyPJ:  180.0, // ~180 pJ per read
			WriteEnergyPJ: 200.0, // ~200 pJ per write
			AreaMm2:       1.85,  // ~1.85 mm²
			AccessTimeNs:  2.5,   // ~2.5 ns access time
			LeakageMW:     170.0, // ~170 mW leakage
			Source:        "CACTI 7.0, 45nm, 1MB SRAM",
		},
	}
}

// macEnergyFromRead estimates the energy per MAC operation from a memory
// read energy. A MAC in compute-in-memory requires reading a weight (one
// memory access) plus analog multiply-accumulate overhead.
//
// For conventional architectures, a MAC requires at minimum:
//   - 1 weight read + 1 activation read + 1 multiply + 1 accumulate
//
// We approximate MAC energy as 2x read energy (conservative lower bound).
func macEnergyFromRead(readEnergyPJ float64) float64 {
	return 2.0 * readEnergyPJ
}

// CompareFeCIMvsCACTI computes normalized energy ratios between FeCIM and
// CACTI baselines for each technology/memory-type combination.
func CompareFeCIMvsCACTI(fecimEnergyPJperMAC float64, baselines []CACTIBaseline) []CACTINormalizedComparison {
	// FeCIM area estimate: 50 mm² chip with embedded CIM
	// This is a model input, not a validated measurement.
	const fecimAreaMm2 = 50.0

	var results []CACTINormalizedComparison

	// Group baselines by technology to create paired SRAM/DRAM comparisons
	type techGroup struct {
		sram *CACTIBaseline
		dram *CACTIBaseline
	}
	groups := make(map[string]*techGroup)

	for i := range baselines {
		b := &baselines[i]
		key := b.Technology + "_" + b.Capacity
		if groups[key] == nil {
			groups[key] = &techGroup{}
		}
		switch b.MemoryType {
		case "SRAM":
			groups[key].sram = b
		case "DRAM":
			groups[key].dram = b
		}
	}

	for _, g := range groups {
		if g.sram == nil {
			continue
		}

		sramMACEnergy := macEnergyFromRead(g.sram.ReadEnergyPJ)
		dramMACEnergy := 0.0
		energyRatioDRAM := 0.0
		if g.dram != nil {
			dramMACEnergy = macEnergyFromRead(g.dram.ReadEnergyPJ)
			if dramMACEnergy > 0 {
				energyRatioDRAM = fecimEnergyPJperMAC / dramMACEnergy
			}
		}

		energyRatioSRAM := 0.0
		if sramMACEnergy > 0 {
			energyRatioSRAM = fecimEnergyPJperMAC / sramMACEnergy
		}

		areaRatioSRAM := 0.0
		if g.sram.AreaMm2 > 0 {
			areaRatioSRAM = fecimAreaMm2 / g.sram.AreaMm2
		}

		comp := CACTINormalizedComparison{
			Baseline:         *g.sram,
			FeCIM_pJperMAC:   fecimEnergyPJperMAC,
			SRAM_pJperMAC:    sramMACEnergy,
			DRAM_pJperMAC:    dramMACEnergy,
			EnergyRatio_SRAM: energyRatioSRAM,
			EnergyRatio_DRAM: energyRatioDRAM,
			AreaRatio_SRAM:   areaRatioSRAM,
			Source:           g.sram.Source,
		}
		results = append(results, comp)
	}

	return results
}

// FormatComparisonTable renders a text table of CACTINormalizedComparison results.
func FormatComparisonTable(comparisons []CACTINormalizedComparison) string {
	var sb strings.Builder

	sb.WriteString("FeCIM vs CACTI Baseline Comparison\n")
	sb.WriteString(strings.Repeat("=", 90) + "\n\n")

	// Header
	sb.WriteString(fmt.Sprintf("%-10s %-8s %-12s %-12s %-12s %-12s %-12s\n",
		"Tech", "Cap", "FeCIM pJ/MAC", "SRAM pJ/MAC", "DRAM pJ/MAC", "Ratio SRAM", "Ratio DRAM"))
	sb.WriteString(strings.Repeat("-", 90) + "\n")

	for _, c := range comparisons {
		dramStr := "N/A"
		ratioStr := "N/A"
		if c.DRAM_pJperMAC > 0 {
			dramStr = fmt.Sprintf("%.1f", c.DRAM_pJperMAC)
			ratioStr = fmt.Sprintf("%.4f", c.EnergyRatio_DRAM)
		}

		sb.WriteString(fmt.Sprintf("%-10s %-8s %-12.1f %-12.1f %-12s %-12.4f %-12s\n",
			c.Baseline.Technology,
			c.Baseline.Capacity,
			c.FeCIM_pJperMAC,
			c.SRAM_pJperMAC,
			dramStr,
			c.EnergyRatio_SRAM,
			ratioStr))
	}

	sb.WriteString("\n")
	sb.WriteString("Ratio < 1.0 means FeCIM is more energy-efficient.\n")
	sb.WriteString("Note: FeCIM values are model inputs (TRL 4), not validated measurements.\n")

	return sb.String()
}
