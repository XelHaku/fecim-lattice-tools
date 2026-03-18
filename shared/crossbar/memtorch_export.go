// memtorch_export.go provides MemTorch-compatible device parameter export
// and import for FeCIM crossbar arrays.
//
// MemTorch (https://github.com/coreylammie/MemTorch) uses resistance-domain
// parameters (RON, ROFF) whereas FeCIM uses conductance-domain (GMin, GMax).
// This file bridges the two representations via R = 1/G.
//
// Variation model:
//   - MemTorch models device-to-device (D2D) variation as Gaussian sigma on
//     RON and ROFF independently.
//   - FeCIM models variation as a single normalized sigma on conductance.
//   - This export maps FeCIM's sigma to both RON_sigma and ROFF_sigma using
//     error propagation: sigma_R = sigma_G / G^2 (from R = 1/G).
package crossbar

import (
	"encoding/json"
	"fmt"
	"math"

	fecimexport "fecim-lattice-tools/shared/export"
	"fecim-lattice-tools/shared/physics"
)

// MemTorchDeviceParams represents MemTorch-format device parameters.
type MemTorchDeviceParams struct {
	ROn       float64 `json:"r_on"`        // Low resistance state (Ohms)
	ROff      float64 `json:"r_off"`       // High resistance state (Ohms)
	ROnSigma  float64 `json:"r_on_sigma"`  // D2D variation at LRS
	ROffSigma float64 `json:"r_off_sigma"` // D2D variation at HRS
	PNoiseStd float64 `json:"p_noise_std"` // Programming noise std
}

// MemTorchWeightExport is the JSON structure for MemTorch weight matrix export.
type MemTorchWeightExport struct {
	Metadata     fecimexport.SimulationMetadata `json:"metadata"`
	DeviceParams MemTorchDeviceParams           `json:"device_params"`
	Rows         int                            `json:"rows"`
	Cols         int                            `json:"cols"`
	Weights      [][]float64                    `json:"weights"`
}

// ExportToMemTorch converts FeCIM crossbar Config conductance parameters
// to MemTorch resistance-domain parameters.
//
// Conversion:
//
//	R_ON  = 1 / G_MAX  (high conductance = low resistance)
//	R_OFF = 1 / G_MIN  (low conductance = high resistance)
//
// Variation mapping uses error propagation from G to R domain:
//
//	sigma_R = sigma_G * R^2  (from delta_R = delta_G / G^2)
//
// For the normalized noise level in Config, we interpret it as a fractional
// sigma relative to the conductance range.
func ExportToMemTorch(cfg *Config) MemTorchDeviceParams {
	gMin := physics.GMin // 10 µS
	gMax := physics.GMax // 100 µS

	rOn := 1.0 / gMax  // High-G state = low-R state
	rOff := 1.0 / gMin // Low-G state = high-R state

	// Map conductance-domain noise to resistance-domain sigma
	// NoiseLevel is normalized sigma (fraction of conductance range)
	gRange := gMax - gMin
	sigmaG := cfg.NoiseLevel * gRange

	// Error propagation: sigma_R ≈ sigma_G / G^2
	rOnSigma := 0.0
	rOffSigma := 0.0
	if gMax > 0 {
		rOnSigma = sigmaG / (gMax * gMax)
	}
	if gMin > 0 {
		rOffSigma = sigmaG / (gMin * gMin)
	}

	// Programming noise: same fractional noise as device variation
	pNoiseStd := cfg.NoiseLevel

	return MemTorchDeviceParams{
		ROn:       rOn,
		ROff:      rOff,
		ROnSigma:  rOnSigma,
		ROffSigma: rOffSigma,
		PNoiseStd: pNoiseStd,
	}
}

// ImportFromMemTorch converts MemTorch resistance-domain parameters back
// to a FeCIM crossbar Config.
//
// Conversion:
//
//	G_MAX = 1 / R_ON   (low resistance = high conductance)
//	G_MIN = 1 / R_OFF  (high resistance = low conductance)
//
// The noise level is recovered from the R_ON sigma using inverse error
// propagation.
func ImportFromMemTorch(params MemTorchDeviceParams) *Config {
	if params.ROn <= 0 || params.ROff <= 0 {
		return &Config{
			Rows:       0,
			Cols:       0,
			NoiseLevel: 0,
		}
	}

	gMax := 1.0 / params.ROn
	gMin := 1.0 / params.ROff

	// Recover noise level from R_ON sigma
	// sigma_R = sigma_G / G^2, so sigma_G = sigma_R * G^2
	// NoiseLevel = sigma_G / (GMax - GMin)
	noiseLevel := 0.0
	gRange := gMax - gMin
	if gRange > 0 && params.ROnSigma > 0 {
		sigmaG := params.ROnSigma * (gMax * gMax)
		noiseLevel = sigmaG / gRange
	}

	return &Config{
		Rows:       0, // Caller must set array dimensions
		Cols:       0,
		NoiseLevel: noiseLevel,
	}
}

// ExportWeightsAsMemTorchJSON exports a weight matrix with device parameters
// in MemTorch-compatible JSON format.
//
// Weights are expected as normalized values in [0, 1] where 0 maps to the
// minimum conductance state and 1 maps to the maximum conductance state.
func ExportWeightsAsMemTorchJSON(weights [][]float64, params MemTorchDeviceParams) ([]byte, error) {
	if len(weights) == 0 {
		return nil, fmt.Errorf("weight matrix is empty")
	}

	rows := len(weights)
	cols := len(weights[0])

	// Validate all rows have the same number of columns
	for i, row := range weights {
		if len(row) != cols {
			return nil, fmt.Errorf("row %d has %d columns, expected %d (jagged matrix)", i, len(row), cols)
		}
	}

	// Validate weight values are in [0, 1]
	for i, row := range weights {
		for j, w := range row {
			if math.IsNaN(w) || math.IsInf(w, 0) {
				return nil, fmt.Errorf("weight[%d][%d] is NaN or Inf", i, j)
			}
			if w < 0 || w > 1 {
				return nil, fmt.Errorf("weight[%d][%d] = %f is outside [0, 1] range", i, j, w)
			}
		}
	}

	result := MemTorchWeightExport{
		Metadata:     fecimexport.DefaultSimulationMetadata(),
		DeviceParams: params,
		Rows:         rows,
		Cols:         cols,
		Weights:      weights,
	}

	return json.MarshalIndent(result, "", "  ")
}
