// Package physics provides shared physics utilities for FeCIM simulations.
// This file implements device-to-device variation modeling for Ec and Pr parameters.
//
// C11: Per Dr. Tour critique - implement Gaussian Ec/Pr distribution to model
// realistic device-to-device variation in manufactured FeFET arrays.
//
// Physics basis:
//   - Ec variation: σ_Ec/Ec ≈ 10-15% typical for HfO2-based ferroelectrics
//   - Pr variation: σ_Pr/Pr ≈ 15-20% typical for HfO2-based ferroelectrics
//   - Variations arise from: grain size distribution, dopant fluctuation,
//     thickness variation, interface quality, crystalline phase distribution
//
// References:
//   - IEEE IRPS 2022 (FeFET reliability and variability)
//   - Nature Commun. 2023 (FeFET array measurements)
//   - CrossSim device models (programming/read variability)
package physics

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"fecim-lattice-tools/shared/mathutil"
)

// DeviceVariationConfig configures device-to-device parameter variation.
type DeviceVariationConfig struct {
	Enable bool // Enable device variation modeling

	// Ec variation (coercive field)
	// Default: mean=1.0 MV/cm (material-dependent), sigma=0.15 (15% relative)
	EcSigmaRelative float64 // σ_Ec/Ec (relative standard deviation)

	// Pr variation (remanent polarization)
	// Default: sigma=0.20 (20% relative)
	PrSigmaRelative float64 // σ_Pr/Pr (relative standard deviation)

	// Correlation coefficient between Ec and Pr variations
	// Typically small positive correlation (0.1-0.3) due to shared grain structure
	EcPrCorrelation float64

	// Random seed (0 = use time-based seed)
	Seed int64
}

// DefaultDeviceVariationConfig returns default variation settings based on literature.
func DefaultDeviceVariationConfig() *DeviceVariationConfig {
	return &DeviceVariationConfig{
		Enable:          false, // Disabled by default for deterministic demos
		EcSigmaRelative: 0.15,  // 15% Ec variation (IEEE IRPS 2022)
		PrSigmaRelative: 0.20,  // 20% Pr variation (slightly higher)
		EcPrCorrelation: 0.2,   // Small positive correlation
		Seed:            0,     // Time-based seed
	}
}

// DeviceVariationEngine generates device-specific Ec/Pr values with Gaussian variation.
type DeviceVariationEngine struct {
	config *DeviceVariationConfig
	rng    *rand.Rand
	mu     sync.Mutex

	// Cache of generated variations per device (row, col)
	deviceCache map[deviceKey]*DeviceVariation
}

type deviceKey struct {
	row, col int
}

// DeviceVariation holds the variation factors for a single device.
type DeviceVariation struct {
	EcFactor float64 // Multiplicative factor for Ec (1.0 = nominal)
	PrFactor float64 // Multiplicative factor for Pr (1.0 = nominal)
}

// NewDeviceVariationEngine creates a new variation engine.
func NewDeviceVariationEngine(config *DeviceVariationConfig) *DeviceVariationEngine {
	if config == nil {
		config = DefaultDeviceVariationConfig()
	}

	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &DeviceVariationEngine{
		config:      config,
		rng:         rand.New(rand.NewSource(seed)),
		deviceCache: make(map[deviceKey]*DeviceVariation),
	}
}

// GetDeviceVariation returns the Ec/Pr variation factors for a device at (row, col).
// Variations are cached per device for consistency across multiple accesses.
func (e *DeviceVariationEngine) GetDeviceVariation(row, col int) *DeviceVariation {
	if !e.config.Enable {
		return &DeviceVariation{EcFactor: 1.0, PrFactor: 1.0}
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	key := deviceKey{row, col}
	if cached, ok := e.deviceCache[key]; ok {
		return cached
	}

	// Generate correlated Gaussian samples for Ec and Pr
	// Using Cholesky decomposition for correlation
	z1 := e.rng.NormFloat64()
	z2 := e.rng.NormFloat64()

	// Correlation: X1 = Z1, X2 = ρ*Z1 + sqrt(1-ρ²)*Z2
	rho := e.config.EcPrCorrelation
	ecNoise := z1
	prNoise := rho*z1 + math.Sqrt(1-rho*rho)*z2

	// Apply relative sigma to get multiplicative factors
	// Factor = 1 + σ_rel * z (so mean=1, std=σ_rel)
	ecFactor := 1.0 + e.config.EcSigmaRelative*ecNoise
	prFactor := 1.0 + e.config.PrSigmaRelative*prNoise

	// Clamp to reasonable bounds (prevent negative or extreme values)
	ecFactor = mathutil.Clamp(ecFactor, 0.5, 1.5)
	prFactor = mathutil.Clamp(prFactor, 0.5, 1.5)

	variation := &DeviceVariation{
		EcFactor: ecFactor,
		PrFactor: prFactor,
	}

	e.deviceCache[key] = variation
	return variation
}

// ApplyToMaterial returns a material with device-specific Ec/Pr values.
func (e *DeviceVariationEngine) ApplyToMaterial(base *HZOMaterial, row, col int) *HZOMaterial {
	if !e.config.Enable {
		return base
	}

	variation := e.GetDeviceVariation(row, col)

	// Create a copy with varied parameters
	varied := *base
	varied.Ec = base.Ec * variation.EcFactor
	varied.Pr = base.Pr * variation.PrFactor
	// Ps scales with Pr (they're correlated)
	varied.Ps = base.Ps * variation.PrFactor

	return &varied
}

// GetArrayVariationStats returns statistics for an array of devices.
func (e *DeviceVariationEngine) GetArrayVariationStats(rows, cols int) *VariationStats {
	if !e.config.Enable {
		return &VariationStats{
			MeanEcFactor: 1.0, StdEcFactor: 0,
			MeanPrFactor: 1.0, StdPrFactor: 0,
		}
	}

	var sumEc, sumPr, sumEc2, sumPr2 float64
	n := float64(rows * cols)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v := e.GetDeviceVariation(r, c)
			sumEc += v.EcFactor
			sumPr += v.PrFactor
			sumEc2 += v.EcFactor * v.EcFactor
			sumPr2 += v.PrFactor * v.PrFactor
		}
	}

	meanEc := sumEc / n
	meanPr := sumPr / n
	varEc := sumEc2/n - meanEc*meanEc
	varPr := sumPr2/n - meanPr*meanPr

	return &VariationStats{
		MeanEcFactor: meanEc,
		StdEcFactor:  math.Sqrt(varEc),
		MeanPrFactor: meanPr,
		StdPrFactor:  math.Sqrt(varPr),
		NumDevices:   int(n),
	}
}

// VariationStats holds statistics about device variation across an array.
type VariationStats struct {
	MeanEcFactor float64
	StdEcFactor  float64
	MeanPrFactor float64
	StdPrFactor  float64
	NumDevices   int
}

// Reset clears the device cache, generating new variations on next access.
func (e *DeviceVariationEngine) Reset() {
	e.mu.Lock()
	e.deviceCache = make(map[deviceKey]*DeviceVariation)
	e.mu.Unlock()
}

// SetSeed sets a new random seed and clears the cache.
func (e *DeviceVariationEngine) SetSeed(seed int64) {
	e.mu.Lock()
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	e.rng = rand.New(rand.NewSource(seed))
	e.deviceCache = make(map[deviceKey]*DeviceVariation)
	e.mu.Unlock()
}

// EstimateYield estimates the fraction of devices within spec.
// A device is within spec if its Ec and Pr are within ±maxDeviation of nominal.
func (e *DeviceVariationEngine) EstimateYield(rows, cols int, maxDeviation float64) float64 {
	if !e.config.Enable {
		return 1.0 // Perfect yield without variation
	}

	inSpec := 0
	total := rows * cols

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			v := e.GetDeviceVariation(r, c)
			ecOK := math.Abs(v.EcFactor-1.0) <= maxDeviation
			prOK := math.Abs(v.PrFactor-1.0) <= maxDeviation
			if ecOK && prOK {
				inSpec++
			}
		}
	}

	return float64(inSpec) / float64(total)
}
