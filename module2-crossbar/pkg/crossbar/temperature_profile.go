package crossbar

// TemperatureProfile controls which temperature-dependent effects are applied.
//
// Units:
//   - AmbientK: Kelvin
//
// Notes:
//   - Wire resistance temperature scaling is handled by IR-drop code paths.
//   - Additional effects (conductance window, noise, drift) are gated behind this
//     profile so existing simulations remain unchanged unless explicitly enabled.
//   - When nil, callers get legacy behavior (wire resistance only).
//
// TODO M2-P2: This struct enables temperature scalings beyond wire resistance.
//
// Rationale: a profile object keeps MVMOptions reasonably small while allowing
// coherent defaults and future extensions (e.g., per-line thermal gradients).

type TemperatureProfile struct {
	// Enable turns on additional temperature-dependent effects beyond wire resistance.
	Enable bool

	// ApplyConductanceWindow scales the effective conductance window (Gmin/Gmax)
	// as a function of temperature.
	ApplyConductanceWindow bool

	// ApplyVariationNoise scales device/process variation magnitude with
	// temperature (thermal noise ~ sqrt(T)).
	ApplyVariationNoise bool

	// ApplyDrift applies a simplified drift-with-temperature model during MVM.
	// Drift is usually simulated separately over time; this is an optional fast
	// approximation for read-time drift impact.
	ApplyDrift bool
}

// DefaultTemperatureProfile returns a conservative profile that enables all
// additional temperature-dependent effects.
func DefaultTemperatureProfile() *TemperatureProfile {
	return &TemperatureProfile{
		Enable:                 true,
		ApplyConductanceWindow: true,
		ApplyVariationNoise:    true,
		ApplyDrift:             true,
	}
}
