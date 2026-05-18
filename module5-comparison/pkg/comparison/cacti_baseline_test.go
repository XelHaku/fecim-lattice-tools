package comparison

import (
	"strings"
	"testing"
)

// TestDefaultCACTIBaselines_DataIntegrity validates that all CACTI baseline
// entries have physically reasonable values.
func TestDefaultCACTIBaselines_DataIntegrity(t *testing.T) {
	baselines := DefaultCACTIBaselines()

	if len(baselines) == 0 {
		t.Fatal("DefaultCACTIBaselines returned empty slice")
	}

	validTechs := map[string]bool{"45nm": true, "28nm": true, "7nm": true, "14nm": true, "22nm": true}
	validTypes := map[string]bool{"SRAM": true, "DRAM": true, "eDRAM": true}

	for i, b := range baselines {
		// Technology node must be recognized
		if !validTechs[b.Technology] {
			t.Errorf("Baseline[%d]: unrecognized technology %q", i, b.Technology)
		}

		// Memory type must be valid
		if !validTypes[b.MemoryType] {
			t.Errorf("Baseline[%d]: unrecognized memory type %q", i, b.MemoryType)
		}

		// Capacity must not be empty
		if b.Capacity == "" {
			t.Errorf("Baseline[%d]: capacity is empty", i)
		}

		// All energy values must be positive
		if b.ReadEnergyPJ <= 0 {
			t.Errorf("Baseline[%d] (%s %s): ReadEnergyPJ must be positive, got %.2f",
				i, b.Technology, b.MemoryType, b.ReadEnergyPJ)
		}
		if b.WriteEnergyPJ <= 0 {
			t.Errorf("Baseline[%d] (%s %s): WriteEnergyPJ must be positive, got %.2f",
				i, b.Technology, b.MemoryType, b.WriteEnergyPJ)
		}

		// Area must be positive
		if b.AreaMm2 <= 0 {
			t.Errorf("Baseline[%d] (%s %s): AreaMm2 must be positive, got %.4f",
				i, b.Technology, b.MemoryType, b.AreaMm2)
		}

		// Access time must be positive
		if b.AccessTimeNs <= 0 {
			t.Errorf("Baseline[%d] (%s %s): AccessTimeNs must be positive, got %.2f",
				i, b.Technology, b.MemoryType, b.AccessTimeNs)
		}

		// Leakage must be non-negative
		if b.LeakageMW < 0 {
			t.Errorf("Baseline[%d] (%s %s): LeakageMW must be non-negative, got %.2f",
				i, b.Technology, b.MemoryType, b.LeakageMW)
		}

		// Source must not be empty
		if b.Source == "" {
			t.Errorf("Baseline[%d] (%s %s): Source attribution is empty",
				i, b.Technology, b.MemoryType)
		}

		// Write energy should be >= read energy (physical expectation)
		if b.WriteEnergyPJ < b.ReadEnergyPJ*0.5 {
			t.Errorf("Baseline[%d] (%s %s): WriteEnergyPJ (%.1f) unexpectedly much less than ReadEnergyPJ (%.1f)",
				i, b.Technology, b.MemoryType, b.WriteEnergyPJ, b.ReadEnergyPJ)
		}

		// Sanity: energy should be in reasonable range (0.1 pJ to 10000 pJ)
		if b.ReadEnergyPJ < 0.1 || b.ReadEnergyPJ > 10000 {
			t.Errorf("Baseline[%d] (%s %s): ReadEnergyPJ %.2f outside reasonable range [0.1, 10000]",
				i, b.Technology, b.MemoryType, b.ReadEnergyPJ)
		}
	}
}

// TestDefaultCACTIBaselines_SRAMDRAMRelationship validates physical
// relationships between SRAM and DRAM within the same technology node.
func TestDefaultCACTIBaselines_SRAMDRAMRelationship(t *testing.T) {
	baselines := DefaultCACTIBaselines()

	// Group by technology+capacity
	type pair struct{ sram, dram *CACTIBaseline }
	pairs := make(map[string]*pair)

	for i := range baselines {
		b := &baselines[i]
		key := b.Technology + "_" + b.Capacity
		if pairs[key] == nil {
			pairs[key] = &pair{}
		}
		switch b.MemoryType {
		case "SRAM":
			pairs[key].sram = b
		case "DRAM":
			pairs[key].dram = b
		}
	}

	for key, p := range pairs {
		if p.sram == nil || p.dram == nil {
			continue // Skip unpaired entries
		}

		// DRAM should be denser (smaller area) than SRAM
		if p.dram.AreaMm2 >= p.sram.AreaMm2 {
			t.Errorf("%s: DRAM area (%.4f) should be less than SRAM area (%.4f)",
				key, p.dram.AreaMm2, p.sram.AreaMm2)
		}

		// DRAM should be slower than SRAM
		if p.dram.AccessTimeNs <= p.sram.AccessTimeNs {
			t.Errorf("%s: DRAM access time (%.1f ns) should be greater than SRAM (%.1f ns)",
				key, p.dram.AccessTimeNs, p.sram.AccessTimeNs)
		}
	}
}

// TestCompareFeCIMvsCACTI_BetterThanSRAM validates that at typical FeCIM energy
// (~0.1 pJ/MAC), the energy ratio is less than 1 versus SRAM baselines.
func TestCompareFeCIMvsCACTI_BetterThanSRAM(t *testing.T) {
	// Typical FeCIM CIM energy: ~0.1 pJ/MAC (model input, not validated)
	fecimEnergy := 0.1 // pJ per MAC
	baselines := DefaultCACTIBaselines()

	comparisons := CompareFeCIMvsCACTI(fecimEnergy, baselines)

	if len(comparisons) == 0 {
		t.Fatal("CompareFeCIMvsCACTI returned no comparisons")
	}

	for _, c := range comparisons {
		// At 0.1 pJ/MAC, FeCIM should be significantly better than SRAM
		// (SRAM MAC energy is typically 100+ pJ)
		if c.EnergyRatio_SRAM >= 1.0 {
			t.Errorf("FeCIM (%.1f pJ/MAC) vs %s SRAM (%.1f pJ/MAC): "+
				"energy ratio %.4f should be < 1.0",
				c.FeCIM_pJperMAC, c.Baseline.Technology,
				c.SRAM_pJperMAC, c.EnergyRatio_SRAM)
		}

		// FeCIM should also be better than DRAM
		if c.DRAM_pJperMAC > 0 && c.EnergyRatio_DRAM >= 1.0 {
			t.Errorf("FeCIM (%.1f pJ/MAC) vs %s DRAM (%.1f pJ/MAC): "+
				"energy ratio %.4f should be < 1.0",
				c.FeCIM_pJperMAC, c.Baseline.Technology,
				c.DRAM_pJperMAC, c.EnergyRatio_DRAM)
		}
	}
}

// TestCompareFeCIMvsCACTI_ReasonableRatios validates that energy ratios are
// within a physically plausible range (0.001x to 10x).
func TestCompareFeCIMvsCACTI_ReasonableRatios(t *testing.T) {
	testCases := []struct {
		name     string
		fecimPJ  float64
		minRatio float64
		maxRatio float64
	}{
		{"typical CIM (0.1 pJ)", 0.1, 0.0001, 1.0},
		{"pessimistic CIM (10 pJ)", 10.0, 0.01, 10.0},
		{"optimistic CIM (0.01 pJ)", 0.01, 0.00001, 0.5},
	}

	baselines := DefaultCACTIBaselines()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comparisons := CompareFeCIMvsCACTI(tc.fecimPJ, baselines)

			for _, c := range comparisons {
				if c.EnergyRatio_SRAM < tc.minRatio || c.EnergyRatio_SRAM > tc.maxRatio {
					t.Errorf("SRAM energy ratio %.6f outside expected range [%.6f, %.1f] for %s",
						c.EnergyRatio_SRAM, tc.minRatio, tc.maxRatio, c.Baseline.Source)
				}
				if c.DRAM_pJperMAC > 0 {
					if c.EnergyRatio_DRAM < tc.minRatio || c.EnergyRatio_DRAM > tc.maxRatio {
						t.Errorf("DRAM energy ratio %.6f outside expected range [%.6f, %.1f] for %s",
							c.EnergyRatio_DRAM, tc.minRatio, tc.maxRatio, c.Baseline.Source)
					}
				}
			}
		})
	}
}

// TestCompareFeCIMvsCACTI_ZeroEnergy validates handling of zero FeCIM energy.
func TestCompareFeCIMvsCACTI_ZeroEnergy(t *testing.T) {
	baselines := DefaultCACTIBaselines()
	comparisons := CompareFeCIMvsCACTI(0.0, baselines)

	for _, c := range comparisons {
		// Zero FeCIM energy should give ratio of 0 (not NaN or Inf)
		if c.EnergyRatio_SRAM != 0 {
			t.Errorf("Zero FeCIM energy should give SRAM ratio of 0, got: %f", c.EnergyRatio_SRAM)
		}
	}
}

// TestFormatComparisonTable validates the table formatter produces readable output.
func TestFormatComparisonTable(t *testing.T) {
	baselines := DefaultCACTIBaselines()
	comparisons := CompareFeCIMvsCACTI(0.1, baselines)

	table := FormatComparisonTable(comparisons)

	if len(table) == 0 {
		t.Fatal("FormatComparisonTable returned empty string")
	}

	// Should contain header
	if !strings.Contains(table, "FeCIM vs CACTI") {
		t.Error("Table should contain title")
	}

	// Should contain column headers
	expectedHeaders := []string{"Tech", "FeCIM pJ/MAC", "SRAM pJ/MAC", "Ratio SRAM"}
	for _, h := range expectedHeaders {
		if !strings.Contains(table, h) {
			t.Errorf("Table should contain header %q", h)
		}
	}

	// Should contain the disclaimer about model inputs
	if !strings.Contains(table, "model input") {
		t.Error("Table should contain disclaimer about model inputs (TRL 4)")
	}
}

// TestMACEnergyFromRead validates the MAC energy estimation function.
func TestMACEnergyFromRead(t *testing.T) {
	// MAC energy should be > read energy (MAC = read + compute overhead)
	readEnergy := 100.0
	macEnergy := macEnergyFromRead(readEnergy)

	if macEnergy <= readEnergy {
		t.Errorf("MAC energy (%.1f) should be greater than read energy (%.1f)",
			macEnergy, readEnergy)
	}

	// Zero read energy should give zero MAC energy
	if macEnergyFromRead(0) != 0 {
		t.Error("Zero read energy should give zero MAC energy")
	}
}
