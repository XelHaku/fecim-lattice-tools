package presets

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/shared/peripherals"
	"fecim-lattice-tools/shared/physics"
)

func TestComprehensive_MaterialPresetsLoadableAndPhysicsValid(t *testing.T) {
	builtins := getAllBuiltInPresets()
	if len(builtins) == 0 {
		t.Fatal("no built-in presets found")
	}

	materialsByName := make(map[string]*physics.HZOMaterial)
	for _, m := range physics.AllMaterials() {
		materialsByName[normalizePresetToken(m.Name)] = m
	}

	// Common UI aliases used by presets.
	aliasToCanonical := map[string]string{
		normalizePresetToken("HZO (optimized)"):         normalizePresetToken("FeCIM HZO"),
		normalizePresetToken("HZO (Si-doped)"):          normalizePresetToken("HZO (Si-doped)"),
		normalizePresetToken("Literature Superlattice"): normalizePresetToken("Literature Superlattice"),
		normalizePresetToken("Cryogenic HZO"):           normalizePresetToken("Cryogenic HZO"),
		normalizePresetToken("AlScN"):                   normalizePresetToken("AlScN (8-16 states)"),
	}

	materialPresetCount := 0
	for _, p := range builtins {
		raw, ok := p.Config["material"]
		if !ok {
			continue
		}
		materialPresetCount++

		name, ok := raw.(string)
		if !ok || strings.TrimSpace(name) == "" {
			t.Fatalf("preset %q has invalid material field: %T (%v)", p.Metadata.Name, raw, raw)
		}

		norm := normalizePresetToken(name)
		if mapped, exists := aliasToCanonical[norm]; exists {
			norm = mapped
		}

		mat := materialsByName[norm]
		if mat == nil {
			t.Fatalf("preset %q references unknown material %q", p.Metadata.Name, name)
		}

		assertValidPhysicsMaterial(t, p.Metadata.Name, mat)
	}

	if materialPresetCount == 0 {
		t.Fatal("expected at least one preset to include a material selection")
	}
}

func TestComprehensive_MeasurementPresetsProduceValidPeripheralConfigs(t *testing.T) {
	builtins := getAllBuiltInPresets()

	measurementPresetCount := 0
	for _, p := range builtins {
		if !isMeasurementPreset(p) {
			continue
		}
		measurementPresetCount++

		dac := peripherals.DefaultDAC()
		adc := peripherals.DefaultADC()
		tia := peripherals.DefaultTIA()

		if bits, ok := getConfigInt(p.Config, "adc_bits"); ok {
			adc.Bits = bits
		}

		if bits, ok := getConfigInt(p.Config, "dac_bits"); ok {
			dac.Bits = bits
		}

		if gainKOhm, ok := getConfigFloat(p.Config, "tia_gain_kohm"); ok {
			tia.Gain = gainKOhm * 1e3
		}

		if err := validatePeripheralChain(dac, adc, tia); err != nil {
			t.Fatalf("measurement preset %q produced invalid peripheral config: %v", p.Metadata.Name, err)
		}
	}

	if measurementPresetCount == 0 {
		t.Fatal("expected at least one measurement-related preset")
	}
}

func TestComprehensive_PresetNamesAreUnique(t *testing.T) {
	builtins := getAllBuiltInPresets()
	seen := make(map[string]string)

	for _, p := range builtins {
		key := strings.ToLower(strings.TrimSpace(p.Metadata.Name))
		if prevID, exists := seen[key]; exists {
			t.Fatalf("duplicate preset name %q found (ids: %s and %s)", p.Metadata.Name, prevID, p.Metadata.ID)
		}
		seen[key] = p.Metadata.ID
	}
}

func TestComprehensive_PresetSerializationRoundTrip(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-comprehensive-"+time.Now().Format("20060102150405.000000000"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	for idx, builtIn := range getAllBuiltInPresets() {
		copyPreset := builtIn.Clone()
		copyPreset.Metadata.BuiltIn = false
		copyPreset.Metadata.ID = fmt.Sprintf("roundtrip-%03d-%s", idx, sanitizeName(copyPreset.Metadata.Name))

		if err := manager.Save(copyPreset); err != nil {
			t.Fatalf("failed saving preset %q: %v", copyPreset.Metadata.Name, err)
		}

		loaded, err := manager.Load(copyPreset.Metadata.ID)
		if err != nil {
			t.Fatalf("failed loading preset %q: %v", copyPreset.Metadata.Name, err)
		}

		if !reflect.DeepEqual(copyPreset.Metadata, loaded.Metadata) {
			t.Fatalf("metadata mismatch for %q after round-trip", copyPreset.Metadata.Name)
		}

		expectedCfg := normalizeJSONMap(t, copyPreset.Config)
		actualCfg := normalizeJSONMap(t, loaded.Config)
		if !reflect.DeepEqual(expectedCfg, actualCfg) {
			t.Fatalf("config mismatch for %q after round-trip", copyPreset.Metadata.Name)
		}
	}
}

func normalizePresetToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func assertValidPhysicsMaterial(t *testing.T, presetName string, m *physics.HZOMaterial) {
	t.Helper()
	if m == nil {
		t.Fatalf("preset %q: material is nil", presetName)
	}
	if !isFinitePositive(m.Ps) {
		t.Fatalf("preset %q: invalid Ps=%g", presetName, m.Ps)
	}
	if !isFinitePositive(m.Pr) {
		t.Fatalf("preset %q: invalid Pr=%g", presetName, m.Pr)
	}
	if m.Ps < m.Pr {
		t.Fatalf("preset %q: Ps (%g) must be >= Pr (%g)", presetName, m.Ps, m.Pr)
	}
	if !isFinitePositive(m.Ec) {
		t.Fatalf("preset %q: invalid Ec=%g", presetName, m.Ec)
	}
	if !isFinitePositive(m.Thickness) {
		t.Fatalf("preset %q: invalid Thickness=%g", presetName, m.Thickness)
	}
	if !isFinitePositive(m.Area) {
		t.Fatalf("preset %q: invalid Area=%g", presetName, m.Area)
	}
	if !isFinitePositive(m.Epsilon) {
		t.Fatalf("preset %q: invalid Epsilon=%g", presetName, m.Epsilon)
	}
	if m.GetNumLevels() < 2 {
		t.Fatalf("preset %q: invalid level count=%d", presetName, m.GetNumLevels())
	}
}

func isMeasurementPreset(p *Preset) bool {
	if p.Metadata.Module == ModuleCircuits || p.Metadata.Module == ModuleCrossbar {
		return true
	}

	for key := range p.Config {
		switch key {
		case "adc_bits", "dac_bits", "tia_gain_kohm", "operation", "show_timing", "show_currents", "show_voltages":
			return true
		}
	}
	return false
}

func validatePeripheralChain(dac *peripherals.DAC, adc *peripherals.ADC, tia *peripherals.TIA) error {
	if dac == nil || adc == nil || tia == nil {
		return fmt.Errorf("nil peripheral in chain")
	}
	if dac.Bits <= 0 || dac.Bits > 16 {
		return fmt.Errorf("invalid dac bits: %d", dac.Bits)
	}
	if adc.Bits <= 0 || adc.Bits > 16 {
		return fmt.Errorf("invalid adc bits: %d", adc.Bits)
	}
	if !isFinitePositive(tia.Gain) {
		return fmt.Errorf("invalid TIA gain: %g", tia.Gain)
	}
	if !isFinitePositive(tia.Bandwidth) {
		return fmt.Errorf("invalid TIA bandwidth: %g", tia.Bandwidth)
	}
	if !isFinitePositive(tia.MaxOutputVoltage) {
		return fmt.Errorf("invalid TIA max output voltage: %g", tia.MaxOutputVoltage)
	}
	if !(dac.VrefHigh > dac.VrefLow) {
		return fmt.Errorf("invalid DAC reference window: [%.4g, %.4g]", dac.VrefLow, dac.VrefHigh)
	}
	if !(adc.VrefHigh > adc.VrefLow) {
		return fmt.Errorf("invalid ADC reference window: [%.4g, %.4g]", adc.VrefLow, adc.VrefHigh)
	}
	return nil
}

func isFinitePositive(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0
}

func getConfigInt(cfg map[string]interface{}, key string) (int, bool) {
	v, ok := cfg[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}

func getConfigFloat(cfg map[string]interface{}, key string) (float64, bool) {
	v, ok := cfg[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	default:
		return 0, false
	}
}

func normalizeJSONMap(t *testing.T, in map[string]interface{}) map[string]interface{} {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return out
}
