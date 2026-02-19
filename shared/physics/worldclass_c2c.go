package physics

import (
	"fmt"
	"math/rand"
)

// StateDepC2CConfig configures state-dependent cycle-to-cycle conductance variation.
// Physics: lower conductance states have proportionally higher relative noise.
// Model: σ_G = k_abs * G_ref / G  (relative noise scales inversely with G)
// where G_ref is a reference conductance (typically G_max of the device).
type StateDepC2CConfig struct {
	// AbsoluteNoiseSigma is the base noise sigma at G = G_ref (conductance units).
	// Typical value: 0.05 * G_ref (5% at max conductance).
	AbsoluteNoiseSigma float64

	// G_ref is the reference conductance where AbsoluteNoiseSigma applies.
	// Set to the maximum device conductance (e.g., G_HRS or G_max).
	G_ref float64
}

// DefaultStateDepC2CConfig returns parameters calibrated for HZO FeFET arrays.
// Based on: 5% C2C at G_max, scaling as 1/G (literature: IEEE EDL 2023).
func DefaultStateDepC2CConfig(G_max float64) StateDepC2CConfig {
	return StateDepC2CConfig{
		AbsoluteNoiseSigma: 0.05 * G_max,
		G_ref:              G_max,
	}
}

// C2CSigma returns the standard deviation of conductance noise for a device at conductance G.
// σ(G) = AbsoluteNoiseSigma * G_ref / G
// Clamped: sigma never exceeds 0.5*G to prevent unphysical sign flips.
func (c StateDepC2CConfig) C2CSigma(G float64) (float64, error) {
	if G <= 0 {
		return 0, fmt.Errorf("conductance G must be > 0, got %g", G)
	}
	if c.G_ref <= 0 {
		return 0, fmt.Errorf("G_ref must be > 0, got %g", c.G_ref)
	}
	sigma := c.AbsoluteNoiseSigma * c.G_ref / G
	maxSigma := 0.5 * G
	if sigma > maxSigma {
		sigma = maxSigma
	}
	return sigma, nil
}

// ApplyStateDepC2C draws a noisy conductance sample given nominal G and config.
// Returns the noisy G value, clamped to [0, 2*G] to prevent negative conductance.
func ApplyStateDepC2C(G float64, cfg StateDepC2CConfig, rng *rand.Rand) (float64, error) {
	sigma, err := cfg.C2CSigma(G)
	if err != nil {
		return G, err
	}
	noise := rng.NormFloat64() * sigma
	noisy := G + noise
	if noisy < 0 {
		noisy = 0
	}
	if noisy > 2*G {
		noisy = 2 * G
	}
	return noisy, nil
}

// C2CRelativeNoise returns σ_G/G for conductance G (for plotting/reporting).
func C2CRelativeNoise(G float64, cfg StateDepC2CConfig) (float64, error) {
	sigma, err := cfg.C2CSigma(G)
	if err != nil {
		return 0, err
	}
	return sigma / G, nil
}

// SweepC2CRelativeNoise returns relative noise σ/G at each conductance value in G_values.
// Useful for plotting σ/G vs G to verify the 1/G relationship.
func SweepC2CRelativeNoise(G_values []float64, cfg StateDepC2CConfig) ([]float64, error) {
	if len(G_values) == 0 {
		return nil, fmt.Errorf("G_values must not be empty")
	}
	out := make([]float64, len(G_values))
	for i, G := range G_values {
		rel, err := C2CRelativeNoise(G, cfg)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		out[i] = rel
	}
	return out, nil
}

