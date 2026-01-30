package physics

import (
	"math"
	"testing"
)

func TestDeviceVariationEngine_Disabled(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = false

	engine := NewDeviceVariationEngine(config)
	v := engine.GetDeviceVariation(0, 0)

	if v.EcFactor != 1.0 {
		t.Errorf("Expected EcFactor=1.0 when disabled, got %v", v.EcFactor)
	}
	if v.PrFactor != 1.0 {
		t.Errorf("Expected PrFactor=1.0 when disabled, got %v", v.PrFactor)
	}
}

func TestDeviceVariationEngine_Enabled(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 42 // Deterministic

	engine := NewDeviceVariationEngine(config)

	// Get variation for device (0,0)
	v := engine.GetDeviceVariation(0, 0)

	// Should be near 1.0 but with some variation
	if v.EcFactor < 0.5 || v.EcFactor > 1.5 {
		t.Errorf("EcFactor out of bounds: %v", v.EcFactor)
	}
	if v.PrFactor < 0.5 || v.PrFactor > 1.5 {
		t.Errorf("PrFactor out of bounds: %v", v.PrFactor)
	}

	// Verify caching - same device should return same values
	v2 := engine.GetDeviceVariation(0, 0)
	if v.EcFactor != v2.EcFactor || v.PrFactor != v2.PrFactor {
		t.Error("Device variation not cached correctly")
	}

	// Different device should have different values (very likely with random)
	v3 := engine.GetDeviceVariation(1, 1)
	if v.EcFactor == v3.EcFactor && v.PrFactor == v3.PrFactor {
		t.Log("Warning: Different devices have same variation (possible but unlikely)")
	}
}

func TestDeviceVariationEngine_Statistics(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 123
	config.EcSigmaRelative = 0.15
	config.PrSigmaRelative = 0.20

	engine := NewDeviceVariationEngine(config)

	// Generate 100x100 array statistics
	stats := engine.GetArrayVariationStats(100, 100)

	// Mean should be close to 1.0
	if math.Abs(stats.MeanEcFactor-1.0) > 0.05 {
		t.Errorf("Mean EcFactor should be ~1.0, got %v", stats.MeanEcFactor)
	}
	if math.Abs(stats.MeanPrFactor-1.0) > 0.05 {
		t.Errorf("Mean PrFactor should be ~1.0, got %v", stats.MeanPrFactor)
	}

	// Std should be close to configured sigma
	if math.Abs(stats.StdEcFactor-config.EcSigmaRelative) > 0.03 {
		t.Errorf("Std EcFactor should be ~%v, got %v", config.EcSigmaRelative, stats.StdEcFactor)
	}
	if math.Abs(stats.StdPrFactor-config.PrSigmaRelative) > 0.05 {
		t.Errorf("Std PrFactor should be ~%v, got %v", config.PrSigmaRelative, stats.StdPrFactor)
	}
}

func TestDeviceVariationEngine_ApplyToMaterial(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 42

	engine := NewDeviceVariationEngine(config)
	base := DefaultHZO()

	varied := engine.ApplyToMaterial(base, 5, 5)

	// Varied material should have different Ec/Pr
	v := engine.GetDeviceVariation(5, 5)
	expectedEc := base.Ec * v.EcFactor
	expectedPr := base.Pr * v.PrFactor

	if varied.Ec != expectedEc {
		t.Errorf("Expected Ec=%v, got %v", expectedEc, varied.Ec)
	}
	if varied.Pr != expectedPr {
		t.Errorf("Expected Pr=%v, got %v", expectedPr, varied.Pr)
	}

	// Original should be unchanged
	if base.Ec == varied.Ec && v.EcFactor != 1.0 {
		t.Error("Original material was modified")
	}
}

func TestDeviceVariationEngine_Yield(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 42
	config.EcSigmaRelative = 0.10
	config.PrSigmaRelative = 0.10

	engine := NewDeviceVariationEngine(config)

	// With 10% sigma, most devices should be within 30% (3-sigma)
	yield := engine.EstimateYield(50, 50, 0.30)
	if yield < 0.95 {
		t.Errorf("Expected >95%% yield at 3-sigma, got %.1f%%", yield*100)
	}

	// Fewer should be within 10% (1-sigma)
	// With two correlated variables, combined yield ≈ 0.68^2 to 0.68 depending on correlation
	yield1sigma := engine.EstimateYield(50, 50, 0.10)
	if yield1sigma > 0.70 || yield1sigma < 0.35 {
		t.Errorf("Expected ~46-68%% yield at 1-sigma (combined), got %.1f%%", yield1sigma*100)
	}
}

func TestDeviceVariationEngine_Reset(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 42

	engine := NewDeviceVariationEngine(config)

	v1 := engine.GetDeviceVariation(0, 0)

	// Reset and regenerate
	engine.SetSeed(99)
	v2 := engine.GetDeviceVariation(0, 0)

	// Should be different after reset with new seed
	if v1.EcFactor == v2.EcFactor && v1.PrFactor == v2.PrFactor {
		t.Error("Reset with new seed should produce different variations")
	}
}

func TestDeviceVariationEngine_Correlation(t *testing.T) {
	config := DefaultDeviceVariationConfig()
	config.Enable = true
	config.Seed = 42
	config.EcPrCorrelation = 0.8 // High correlation

	engine := NewDeviceVariationEngine(config)

	// With high correlation, Ec and Pr variations should track together
	var sumProduct, sumEc, sumPr, sumEc2, sumPr2 float64
	n := 1000

	for i := 0; i < n; i++ {
		v := engine.GetDeviceVariation(i, 0)
		ecDev := v.EcFactor - 1.0
		prDev := v.PrFactor - 1.0
		sumProduct += ecDev * prDev
		sumEc += ecDev
		sumPr += prDev
		sumEc2 += ecDev * ecDev
		sumPr2 += prDev * prDev
	}

	// Calculate sample correlation coefficient
	nf := float64(n)
	cov := sumProduct/nf - (sumEc/nf)*(sumPr/nf)
	stdEc := math.Sqrt(sumEc2/nf - (sumEc/nf)*(sumEc/nf))
	stdPr := math.Sqrt(sumPr2/nf - (sumPr/nf)*(sumPr/nf))
	correlation := cov / (stdEc * stdPr)

	// Should be close to configured correlation
	if math.Abs(correlation-config.EcPrCorrelation) > 0.15 {
		t.Errorf("Expected correlation ~%v, got %v", config.EcPrCorrelation, correlation)
	}
}
