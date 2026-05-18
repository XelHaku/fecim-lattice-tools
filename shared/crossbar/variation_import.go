package crossbar

import (
	"math"

	"fecim-lattice-tools/shared/export"
)

// CrossSimVariationConfig represents CrossSim-format device variation parameters.
//
// CrossSim (Sandia National Labs) uses a set of noise/variation parameters
// that decompose device non-idealities into programming noise, read noise,
// drift, device-to-device (D2D), and cycle-to-cycle (C2C) components.
//
// Reference: Agarwal et al., "Achieving Ideal Accuracies in Analog Neuromorphic
// Computing Using Periodic Carry", Proc. CSNDSP 2016.
// https://github.com/sandialabs/cross-sim
type CrossSimVariationConfig struct {
	// Programming noise
	ProgramNoiseSigma float64 `json:"sigma_prog"` // Programming noise std dev (relative)

	// Read noise
	ReadNoiseSigma float64 `json:"sigma_read"` // Read noise std dev (relative)

	// Drift model (power-law: G(t) = G0 * (t/t0)^(-v))
	DriftCoeff float64 `json:"drift_coeff"` // Power-law drift coefficient (v)
	DriftT0    float64 `json:"drift_t0"`    // Reference time for drift (s)

	// Device-to-device variation
	D2DSigmaHRS float64 `json:"sigma_d2d_hrs"` // D2D variation at high resistance state (relative)
	D2DSigmaLRS float64 `json:"sigma_d2d_lrs"` // D2D variation at low resistance state (relative)

	// Cycle-to-cycle variation
	C2CSigma float64 `json:"sigma_c2c"` // Cycle-to-cycle variation (relative)

	// Simulation disclaimer (auto-injected on export)
	Disclaimer string `json:"disclaimer,omitempty"`
}

// DefaultCrossSimHZO returns typical HZO FeFET variation parameters from
// published CrossSim studies and FeFET characterization literature.
//
// References:
//   - Soliman et al., Nature Commun. 14, 6348 (2023) — 28nm HKMG FeFET crossbar
//   - Reis et al., arXiv:2312.15444 (2023) — state-dependent C2C/D2D model
//   - CrossSim documentation (Sandia) — generic device error model parameters
//
// NOTE: These are representative educational defaults. Treat as [CALIBRATION NEEDED]
// for quantitative hardware claims.
func DefaultCrossSimHZO() CrossSimVariationConfig {
	return CrossSimVariationConfig{
		ProgramNoiseSigma: 0.03, // 3% programming noise
		ReadNoiseSigma:    0.01, // 1% read noise
		DriftCoeff:        0.05, // Mild power-law drift (v=0.05)
		DriftT0:           1.0,  // 1 second reference time
		D2DSigmaHRS:       0.05, // 5% D2D at HRS
		D2DSigmaLRS:       0.03, // 3% D2D at LRS
		C2CSigma:          0.03, // 3% cycle-to-cycle variation
	}
}

// ImportCrossSimVariation converts CrossSim-format variation parameters to our
// ProcessVariationConfig format for use with crossbar arrays.
//
// The mapping combines CrossSim's D2D sigma components into a single DeviceSigma
// using a geometric mean, and maps the programming noise to gradient/edge
// placeholders. The exact mapping is approximate since the CrossSim model
// has a richer decomposition than our compact ProcessVariationConfig.
func ImportCrossSimVariation(config CrossSimVariationConfig) *ProcessVariationConfig {
	// Combine D2D HRS and LRS sigmas into a single representative device sigma.
	// Geometric mean gives a balanced estimate across the conductance range.
	deviceSigma := config.D2DSigmaHRS
	if config.D2DSigmaLRS > 0 && config.D2DSigmaHRS > 0 {
		deviceSigma = math.Sqrt(config.D2DSigmaHRS * config.D2DSigmaLRS)
	} else if config.D2DSigmaLRS > 0 {
		deviceSigma = config.D2DSigmaLRS
	}

	// Add programming noise contribution in quadrature (independent error sources).
	if config.ProgramNoiseSigma > 0 {
		deviceSigma = math.Sqrt(deviceSigma*deviceSigma + config.ProgramNoiseSigma*config.ProgramNoiseSigma)
	}

	// Gradient defaults: CrossSim does not model spatial gradients, so we use
	// a small fraction of the programming noise as a heuristic placeholder.
	gradientX := config.ProgramNoiseSigma * 0.1
	gradientY := config.ProgramNoiseSigma * 0.1

	// Edge effect: estimated as 2x the read noise sigma (edge cells tend to
	// have higher variability due to lithographic proximity effects).
	edgeEffect := config.ReadNoiseSigma * 2.0
	if edgeEffect > 0.20 {
		edgeEffect = 0.20 // Cap at 20% to avoid unrealistic degradation
	}

	return &ProcessVariationConfig{
		DeviceSigma: deviceSigma,
		GradientX:   gradientX,
		GradientY:   gradientY,
		EdgeEffect:  edgeEffect,
	}
}

// ExportToCrossSimFormat converts our ProcessVariationConfig back to CrossSim
// format for interoperability with Sandia's CrossSim simulator.
//
// Since our model is less granular than CrossSim's, the reverse mapping uses
// reasonable assumptions to decompose our DeviceSigma into the CrossSim fields.
func ExportToCrossSimFormat(config *ProcessVariationConfig) CrossSimVariationConfig {
	if config == nil {
		return CrossSimVariationConfig{}
	}

	// Split DeviceSigma into D2D and programming noise using a 70/30 heuristic:
	// most of the sigma is D2D, with a smaller programming noise component.
	totalSigma := config.DeviceSigma
	d2dContrib := totalSigma * 0.70
	progContrib := totalSigma * 0.30

	// Asymmetric D2D split: HRS typically has ~1.5x the variation of LRS
	// due to the wider distribution of defect configurations in the reset state.
	d2dHRS := d2dContrib * 1.2
	d2dLRS := d2dContrib * 0.8

	// Read noise from edge effect (reversed from import mapping).
	readNoise := config.EdgeEffect / 2.0

	return CrossSimVariationConfig{
		ProgramNoiseSigma: progContrib,
		ReadNoiseSigma:    readNoise,
		DriftCoeff:        0.05, // Default drift (not represented in ProcessVariationConfig)
		DriftT0:           1.0,  // Default reference time
		D2DSigmaHRS:       d2dHRS,
		D2DSigmaLRS:       d2dLRS,
		C2CSigma:          progContrib, // C2C ≈ programming noise for compact mapping
		Disclaimer:        export.SimulationDisclaimer(),
	}
}
