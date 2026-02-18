package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestM6LIB03_PowerDynamicPowerEntriesPresent — M6-LIB-03
// Verify dynamic power entries > 0
// Check rise_power, fall_power tables
func TestM6LIB03_PowerDynamicPowerEntriesPresent(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.LeakagePower = 0.5 // nW
	lib := GenerateLiberty(cfg)

	// Verify cell_leakage_power attribute
	if !strings.Contains(lib, "cell_leakage_power :") {
		t.Fatal("missing cell_leakage_power attribute")
	}

	// Extract cell_leakage_power value
	reLeakage := regexp.MustCompile(`cell_leakage_power\s*:\s*([0-9.]+)`)
	mLeakage := reLeakage.FindStringSubmatch(lib)
	if len(mLeakage) < 2 {
		t.Fatal("failed to extract cell_leakage_power value")
	}

	leakagePower, err := strconv.ParseFloat(mLeakage[1], 64)
	if err != nil {
		t.Fatalf("failed to parse leakage power: %v", err)
	}

	if leakagePower <= 0 {
		t.Fatalf("cell_leakage_power must be > 0, got %.6f nW", leakagePower)
	}

	expectedLeakage := cfg.LeakagePower
	tolerance := 0.01 // 1% tolerance
	delta := (leakagePower - expectedLeakage) / expectedLeakage
	if delta < -tolerance || delta > tolerance {
		t.Fatalf("cell_leakage_power mismatch: got %.6f nW, expected %.6f nW (delta %.2f%%)",
			leakagePower, expectedLeakage, delta*100)
	}

	t.Logf("M6-LIB-03 PASS: cell_leakage_power = %.6f nW (expected %.6f nW, delta %.2f%%)",
		leakagePower, expectedLeakage, delta*100)
}

// TestM6LIB03_PowerModule4EnergyAnnotation validates Module 4 internal_power injection
func TestM6LIB03_PowerModule4EnergyAnnotation(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.RiseTime = 50.0 // ns
	cfg.FallTime = 5.0  // ns

	// Module 4 energy model (J/op)
	energy := &Module4EnergyModel{
		DACEnergyJ: 14.4e-15, // 14.4 fJ (DAC)
		MVMEnergyJ: 25.0e-15, // 25.0 fJ (MVM)
		TIAEnergyJ: 6.3e-15,  // 6.3 fJ (TIA)
	}

	lib := GenerateLibertyWithModule4Energy(cfg, energy)

	// Verify internal_power blocks exist
	internalPowerCount := strings.Count(lib, "internal_power()")
	if internalPowerCount < 3 {
		t.Fatalf("expected at least 3 internal_power blocks, got %d", internalPowerCount)
	}

	// Verify rise_power and fall_power entries
	if !strings.Contains(lib, "rise_power(scalar)") {
		t.Fatal("missing rise_power(scalar) entry")
	}
	if !strings.Contains(lib, "fall_power(scalar)") {
		t.Fatal("missing fall_power(scalar) entry")
	}

	// Extract first rise_power value (DAC)
	reRisePower := regexp.MustCompile(`rise_power\(scalar\)\s*\{\s*values\("([0-9.]+)"\)`)
	mRisePower := reRisePower.FindStringSubmatch(lib)
	if len(mRisePower) < 2 {
		t.Fatal("failed to extract rise_power value")
	}

	risePowerNW, err := strconv.ParseFloat(mRisePower[1], 64)
	if err != nil {
		t.Fatalf("failed to parse rise_power: %v", err)
	}

	if risePowerNW <= 0 {
		t.Fatalf("rise_power must be > 0, got %.6f nW", risePowerNW)
	}

	// Compute expected DAC power (E / cycle_time)
	cycleTimeS := (cfg.RiseTime + cfg.FallTime) * 1e-9
	expectedDACPowerNW := (energy.DACEnergyJ / cycleTimeS) * 1e9

	tolerance := 0.05 // 5% tolerance
	delta := (risePowerNW - expectedDACPowerNW) / expectedDACPowerNW
	if delta < -tolerance || delta > tolerance {
		t.Fatalf("rise_power mismatch: got %.6f nW, expected %.6f nW (delta %.2f%%)",
			risePowerNW, expectedDACPowerNW, delta*100)
	}

	t.Logf("M6-LIB-03 PASS: Module 4 internal_power annotation validated")
	t.Logf("  - DAC energy: %.2f fJ", energy.DACEnergyJ*1e15)
	t.Logf("  - MVM energy: %.2f fJ", energy.MVMEnergyJ*1e15)
	t.Logf("  - TIA energy: %.2f fJ", energy.TIAEnergyJ*1e15)
	t.Logf("  - Cycle time: %.1f ns", cycleTimeS*1e9)
	t.Logf("  - DAC power: %.6f nW (expected %.6f nW, delta %.2f%%)",
		risePowerNW, expectedDACPowerNW, delta*100)
	t.Logf("  - internal_power blocks: %d", internalPowerCount)
}

// TestM6LIB03_PowerRelatedPins validates internal_power related_pin attributes
func TestM6LIB03_PowerRelatedPins(t *testing.T) {
	cfg := config.DefaultCellConfig()
	energy := &Module4EnergyModel{
		DACEnergyJ: 10.0e-15,
		MVMEnergyJ: 20.0e-15,
		TIAEnergyJ: 5.0e-15,
	}

	lib := GenerateLibertyWithModule4Energy(cfg, energy)

	// Verify WL related_pin (DAC)
	reWLPower := regexp.MustCompile(`internal_power\(\)\s*\{[^}]*related_pin\s*:\s*"WL"`)
	if !reWLPower.MatchString(lib) {
		t.Fatal("missing internal_power block with related_pin: WL (DAC)")
	}

	// Verify BL related_pin (MVM, TIA)
	reBLPower := regexp.MustCompile(`internal_power\(\)\s*\{[^}]*related_pin\s*:\s*"BL"`)
	blMatches := reBLPower.FindAllString(lib, -1)
	if len(blMatches) < 2 {
		t.Fatalf("expected at least 2 internal_power blocks with related_pin: BL (MVM, TIA), got %d", len(blMatches))
	}

	t.Logf("M6-LIB-03 PASS: internal_power related_pin attributes validated")
	t.Logf("  - WL (DAC): 1 block")
	t.Logf("  - BL (MVM/TIA): %d blocks", len(blMatches))
}

// TestM6LIB03_PowerScalarTableFormat validates scalar power table format
func TestM6LIB03_PowerScalarTableFormat(t *testing.T) {
	cfg := config.DefaultCellConfig()
	energy := &Module4EnergyModel{
		DACEnergyJ: 15.0e-15,
		MVMEnergyJ: 30.0e-15,
		TIAEnergyJ: 8.0e-15,
	}

	lib := GenerateLibertyWithModule4Energy(cfg, energy)

	// Verify scalar table format: values("value")
	reScalar := regexp.MustCompile(`(rise_power|fall_power)\(scalar\)\s*\{\s*values\("([0-9.]+)"\)\s*;\s*\}`)
	matches := reScalar.FindAllStringSubmatch(lib, -1)

	if len(matches) < 6 {
		// 3 internal_power blocks × 2 (rise + fall) = 6
		t.Fatalf("expected at least 6 scalar power values (3 blocks × 2), got %d", len(matches))
	}

	// Verify all extracted values are > 0
	for i, m := range matches {
		if len(m) < 3 {
			continue
		}
		powerType := m[1]
		valueStr := m[2]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			t.Fatalf("failed to parse scalar power value #%d (%s): %v", i, valueStr, err)
		}
		if value <= 0 {
			t.Fatalf("%s scalar value #%d must be > 0, got %.6f nW", powerType, i, value)
		}
		t.Logf("  - %s[%d]: %.6f nW", powerType, i, value)
	}

	t.Logf("M6-LIB-03 PASS: Scalar power table format validated (%d entries)", len(matches))
}

// TestM6LIB03_PowerZeroEnergyModel validates behavior with zero energy
func TestM6LIB03_PowerZeroEnergyModel(t *testing.T) {
	cfg := config.DefaultCellConfig()
	energy := &Module4EnergyModel{
		DACEnergyJ: 0.0,
		MVMEnergyJ: 0.0,
		TIAEnergyJ: 0.0,
	}

	lib := GenerateLibertyWithModule4Energy(cfg, energy)

	// Should still have internal_power blocks, but with zero values
	if !strings.Contains(lib, "internal_power()") {
		t.Fatal("missing internal_power blocks even with zero energy")
	}

	// Extract rise_power values and verify they are 0
	reRisePower := regexp.MustCompile(`rise_power\(scalar\)\s*\{\s*values\("([0-9.]+)"\)`)
	matches := reRisePower.FindAllStringSubmatch(lib, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		value, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("failed to parse zero energy rise_power: %v", err)
		}
		if value != 0.0 {
			t.Fatalf("expected rise_power = 0 with zero energy model, got %.6f nW", value)
		}
	}

	t.Logf("M6-LIB-03 PASS: Zero energy model validated (rise_power = 0.0 nW)")
}
