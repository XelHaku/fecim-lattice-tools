// Package widgets provides reusable UI components.
package widgets

import (
	"fmt"
	"math"
	"strings"

	"fecim-lattice-tools/config/physics"
)

// Property categories organized by role in the Frankenstein L-K equation:
//   ρ_eff·dP/dt = E_applied - k_dep·P - (2αP + 4βP³ + 6γP⁵) + ξ(t)
//   ρ_eff = ρ + (R_series·A)/d
//   α(T,σ) = (T-Tc)/(2ε₀C) - 2Q₁₂σ
const (
	CategoryCore         = "Core (Pr, Ps, Ec)"      // Reference polarization/field
	CategoryGeometry     = "Geometry (A, d)"        // For ρ_eff calculation
	CategoryLandau       = "Landau (β, γ, ρ)"       // L-K coefficients
	CategoryAlpha        = "Alpha (Tc, C, Q₁₂, σ)"  // Dynamic stiffness α(T,σ)
	CategoryDepol        = "Depolarization (k_dep)" // Slanted loop
	CategoryCircuit      = "Circuit (R_series)"     // Series resistance
	CategoryNLS          = "NLS (τ∞, Ea)"           // Merz law dynamics
	CategoryConductance  = "Conductance (G)"        // P→G mapping
)

// ModelUsage indicates which physics models use a parameter.
type ModelUsage struct {
	Preisach bool // Used in Preisach hysteresis model
	LandauKh bool // Used in Landau-Khalatnikov solver
}

// String returns the model indicator string.
func (m ModelUsage) String() string {
	if m.LandauKh && m.Preisach {
		return "[L+P]"
	}
	if m.LandauKh {
		return "[L-K]"
	}
	if m.Preisach {
		return "[P]"
	}
	return ""
}

// FormattedProperty holds a material property with display formatting.
type FormattedProperty struct {
	Name        string     // Display name
	Value       string     // Formatted value with units
	RawValue    float64    // Raw value for sorting/comparison
	Category    string     // Physics category
	Description string     // Tooltip/help text
	Models      ModelUsage // Which physics models use this parameter
}

// Model usage markers
var (
	lkModel      = ModelUsage{LandauKh: true}
	preisachModel = ModelUsage{Preisach: true}
	bothModels   = ModelUsage{LandauKh: true, Preisach: true}
)

// FormatPolarization converts C/m² to µC/cm² display string.
func FormatPolarization(cM2 float64) string {
	microCcm2 := cM2 * 100
	if microCcm2 >= 100 {
		return fmt.Sprintf("%.0f µC/cm²", microCcm2)
	}
	return fmt.Sprintf("%.1f µC/cm²", microCcm2)
}

// FormatField converts V/m to MV/cm display string.
func FormatField(vM float64) string {
	mvCm := vM / 1e8
	if mvCm >= 1 {
		return fmt.Sprintf("%.1f MV/cm", mvCm)
	}
	return fmt.Sprintf("%.2f MV/cm", mvCm)
}

// FormatThickness converts m to nm display string.
func FormatThickness(m float64) string {
	nm := m * 1e9
	if nm >= 10 {
		return fmt.Sprintf("%.0f nm", nm)
	}
	return fmt.Sprintf("%.1f nm", nm)
}

// FormatArea converts m² to nm² display string.
func FormatArea(m2 float64) string {
	nm2 := m2 * 1e18
	if nm2 >= 1000 {
		return fmt.Sprintf("%.0f nm²", nm2)
	} else if nm2 >= 1 {
		return fmt.Sprintf("%.1f nm²", nm2)
	}
	return fmt.Sprintf("%.3f nm²", nm2)
}

// FormatTime converts seconds to appropriate time unit.
func FormatTime(s float64) string {
	if s <= 0 {
		return "N/A"
	}
	if s < 1e-12 {
		return fmt.Sprintf("%.1f fs", s*1e15)
	}
	if s < 1e-9 {
		return fmt.Sprintf("%.0f ps", s*1e12)
	}
	if s < 1e-6 {
		return fmt.Sprintf("%.1f ns", s*1e9)
	}
	if s < 1e-3 {
		return fmt.Sprintf("%.1f µs", s*1e6)
	}
	if s < 1 {
		return fmt.Sprintf("%.1f ms", s*1e3)
	}
	if s < 60 {
		return fmt.Sprintf("%.1f s", s)
	}
	if s < 3600 {
		return fmt.Sprintf("%.1f min", s/60)
	}
	if s < 86400 {
		return fmt.Sprintf("%.1f hr", s/3600)
	}
	if s < 31536000 {
		return fmt.Sprintf("%.0f days", s/86400)
	}
	years := s / 31536000
	if years >= 99.5 {
		return fmt.Sprintf("%.0f years", math.Round(years))
	}
	return fmt.Sprintf("%.1f years", years)
}

// FormatEndurance formats cycle count with superscript notation.
func FormatEndurance(cycles float64) string {
	if cycles <= 0 {
		return "N/A"
	}
	exp := math.Log10(cycles)
	if exp >= 1 && math.Abs(exp-math.Round(exp)) < 0.01 {
		return fmt.Sprintf("10^%.0f cycles", exp)
	}
	if cycles >= 1e12 {
		return fmt.Sprintf("%.1f×10^12 cycles", cycles/1e12)
	}
	if cycles >= 1e9 {
		return fmt.Sprintf("%.1f×10^9 cycles", cycles/1e9)
	}
	if cycles >= 1e6 {
		return fmt.Sprintf("%.1f×10^6 cycles", cycles/1e6)
	}
	return fmt.Sprintf("%.0f cycles", cycles)
}

// FormatTemperature converts K to display string with Celsius.
func FormatTemperature(k float64) string {
	celsius := k - 273.15
	return fmt.Sprintf("%.0f K (%.0f°C)", k, celsius)
}

// FormatEnergy formats energy in eV.
func FormatEnergy(ev float64) string {
	return fmt.Sprintf("%.2f eV", ev)
}

// FormatConductanceRatio formats a ratio for display.
func FormatConductanceRatio(ratio float64) string {
	if ratio <= 0 {
		return "N/A"
	}
	if ratio >= 1e5 {
		return fmt.Sprintf(">10^5:1")
	}
	if ratio >= 1e4 {
		return fmt.Sprintf("%.0f×10^4:1", ratio/1e4)
	}
	if ratio >= 1000 {
		return fmt.Sprintf("%.0fk:1", ratio/1000)
	}
	return fmt.Sprintf("%.0f:1", ratio)
}

// FormatVoltage formats voltage in V.
func FormatVoltage(v float64) string {
	if v >= 1 {
		return fmt.Sprintf("%.1f V", v)
	}
	return fmt.Sprintf("%.0f mV", v*1000)
}

// FormatDimensionless formats a dimensionless value.
func FormatDimensionless(v float64) string {
	if v == math.Floor(v) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

// FormatPercent formats a fraction as percentage.
func FormatPercent(v float64) string {
	return fmt.Sprintf("%.1f%%", v*100)
}

// GetMaterialProperties extracts properties relevant to the Frankenstein L-K equation.
// Parameters are organized by their role in:
//   ρ_eff·dP/dt = E_applied - k_dep·P - (2αP + 4βP³ + 6γP⁵) + ξ(t)
func GetMaterialProperties(mat *physics.Material) []FormattedProperty {
	props := []FormattedProperty{}

	// ═══════════════════════════════════════════════════════════════════════════
	// CORE: Reference polarization and field (for normalization and comparison)
	// ═══════════════════════════════════════════════════════════════════════════
	props = append(props, FormattedProperty{
		Name:        "Pr (Remanent)",
		Value:       FormatPolarization(mat.PrCM2),
		RawValue:    mat.PrCM2,
		Category:    CategoryCore,
		Description: "Remanent polarization. Reference for P→G mapping normalization.",
		Models:      bothModels,
	})
	props = append(props, FormattedProperty{
		Name:        "Ps (Saturation)",
		Value:       FormatPolarization(mat.PsCM2),
		RawValue:    mat.PsCM2,
		Category:    CategoryCore,
		Description: "Saturation polarization. Sets P_max for conductance mapping: G = f(P/Ps).",
		Models:      bothModels,
	})
	props = append(props, FormattedProperty{
		Name:        "Ec (Coercive)",
		Value:       FormatField(mat.EcVM),
		RawValue:    mat.EcVM,
		Category:    CategoryCore,
		Description: "Coercive field. Reference for E_applied normalization. Vc = Ec × d.",
		Models:      bothModels,
	})
	if mat.AnalogStates > 0 {
		props = append(props, FormattedProperty{
			Name:        "Analog States",
			Value:       fmt.Sprintf("%d (%.1f bits)", mat.AnalogStates, math.Log2(float64(mat.AnalogStates))),
			RawValue:    float64(mat.AnalogStates),
			Category:    CategoryCore,
			Description: "Number of discrete programmable states via partial polarization switching.",
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// GEOMETRY: Film dimensions for ρ_eff = ρ + (R_series·A)/d
	// ═══════════════════════════════════════════════════════════════════════════
	props = append(props, FormattedProperty{
		Name:        "d (Thickness)",
		Value:       FormatThickness(mat.ThicknessM),
		RawValue:    mat.ThicknessM,
		Category:    CategoryGeometry,
		Description: "Film thickness. Used in ρ_eff = ρ + (R·A)/d and E = V/d conversion.",
		Models:      lkModel,
	})
	props = append(props, FormattedProperty{
		Name:        "A (Area)",
		Value:       FormatArea(mat.AreaM2),
		RawValue:    mat.AreaM2,
		Category:    CategoryGeometry,
		Description: "Active cell area. Used in ρ_eff = ρ + (R·A)/d calculation.",
		Models:      lkModel,
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// LANDAU: Core L-K coefficients (β, γ, ρ) in dP/dt equation
	// ═══════════════════════════════════════════════════════════════════════════
	if mat.Thermodynamics.BetaLandau != 0 {
		props = append(props, FormattedProperty{
			Name:        "β (Landau)",
			Value:       fmt.Sprintf("%.3e J·m⁵/C⁴", mat.Thermodynamics.BetaLandau),
			RawValue:    mat.Thermodynamics.BetaLandau,
			Category:    CategoryLandau,
			Description: "First-order barrier. NEGATIVE for HZO → creates 4βP³ switching barrier.",
			Models:      lkModel,
		})
	}
	if mat.Thermodynamics.GammaLandau != 0 {
		props = append(props, FormattedProperty{
			Name:        "γ (Landau)",
			Value:       fmt.Sprintf("%.3e J·m⁹/C⁶", mat.Thermodynamics.GammaLandau),
			RawValue:    mat.Thermodynamics.GammaLandau,
			Category:    CategoryLandau,
			Description: "Stability term. POSITIVE → prevents runaway at high P via 6γP⁵.",
			Models:      lkModel,
		})
	}
	if mat.Thermodynamics.RhoViscosity > 0 {
		props = append(props, FormattedProperty{
			Name:        "ρ (Viscosity)",
			Value:       fmt.Sprintf("%.3f Ω·m", mat.Thermodynamics.RhoViscosity),
			RawValue:    mat.Thermodynamics.RhoViscosity,
			Category:    CategoryLandau,
			Description: "Khalatnikov damping. ρ<0.1 for GHz operation. Enters ρ_eff = ρ + R·A/d.",
			Models:      lkModel,
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// ALPHA: Dynamic stiffness α(T,σ) = (T-Tc)/(2ε₀C) - 2Q₁₂σ
	// ═══════════════════════════════════════════════════════════════════════════
	props = append(props, FormattedProperty{
		Name:        "Tc (Curie)",
		Value:       FormatTemperature(mat.CurieTempK),
		RawValue:    mat.CurieTempK,
		Category:    CategoryAlpha,
		Description: "Curie temperature. α → 0 as T → Tc (wells become shallow → data loss).",
		Models:      lkModel,
	})
	if mat.Thermodynamics.CurieConstK > 0 {
		props = append(props, FormattedProperty{
			Name:        "C (Curie Const)",
			Value:       fmt.Sprintf("%.2e K", mat.Thermodynamics.CurieConstK),
			RawValue:    mat.Thermodynamics.CurieConstK,
			Category:    CategoryAlpha,
			Description: "Curie-Weiss constant. α = (T-Tc)/(2ε₀C) - 2Q₁₂σ.",
			Models:      lkModel,
		})
	}
	if mat.Coupling.Q12Electrostriction != 0 {
		props = append(props, FormattedProperty{
			Name:        "Q₁₂ (Electrostric.)",
			Value:       fmt.Sprintf("%.3f m⁴/C²", mat.Coupling.Q12Electrostriction),
			RawValue:    mat.Coupling.Q12Electrostriction,
			Category:    CategoryAlpha,
			Description: "Transverse electrostriction. α includes -2Q₁₂σ term. Q₁₂<0 typical.",
			Models:      lkModel,
		})
	}
	if mat.Coupling.StressGPa > 0 {
		props = append(props, FormattedProperty{
			Name:        "σ (Stress)",
			Value:       fmt.Sprintf("%.1f GPa", mat.Coupling.StressGPa),
			RawValue:    mat.Coupling.StressGPa,
			Category:    CategoryAlpha,
			Description: "In-plane stress (TiN capping). Tensile σ>0 with Q₁₂<0 raises α.",
			Models:      lkModel,
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// DEPOLARIZATION: k_dep term for slanted hysteresis (analog states)
	// ═══════════════════════════════════════════════════════════════════════════
	if mat.DepolarizationFactorVMC > 0 {
		props = append(props, FormattedProperty{
			Name:        "k_dep",
			Value:       fmt.Sprintf("%.2e V·m/C", mat.DepolarizationFactorVMC),
			RawValue:    mat.DepolarizationFactorVMC,
			Category:    CategoryDepol,
			Description: "Depolarization factor. E_eff = E_applied - k_dep·P creates slanted loop.",
			Models:      lkModel,
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// CIRCUIT: Series resistance for ρ_eff = ρ + (R_series·A)/d
	// ═══════════════════════════════════════════════════════════════════════════
	if mat.Circuit.SeriesResistanceOhm > 0 {
		props = append(props, FormattedProperty{
			Name:        "R_series",
			Value:       fmt.Sprintf("%.0f Ω", mat.Circuit.SeriesResistanceOhm),
			RawValue:    mat.Circuit.SeriesResistanceOhm,
			Category:    CategoryCircuit,
			Description: "Series resistance (contact + wire). ρ_eff = ρ + (R·A)/d adds RC delay.",
			Models:      lkModel,
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// NLS: Nucleation-Limited Switching (Merz law) τ = τ∞·exp(Ea/|E|)
	// ═══════════════════════════════════════════════════════════════════════════
	if mat.NLS.TauInfS > 0 {
		props = append(props, FormattedProperty{
			Name:        "τ∞ (NLS)",
			Value:       FormatTime(mat.NLS.TauInfS),
			RawValue:    mat.NLS.TauInfS,
			Category:    CategoryNLS,
			Description: "NLS attempt time. τ(E) = τ∞·exp(Ea/|E|). ~100ps for HZO.",
			Models:      lkModel,
		})
	}
	if mat.NLS.ActivationFieldVM > 0 {
		props = append(props, FormattedProperty{
			Name:        "Ea (NLS)",
			Value:       FormatField(mat.NLS.ActivationFieldVM),
			RawValue:    mat.NLS.ActivationFieldVM,
			Category:    CategoryNLS,
			Description: "NLS activation field. τ(E) = τ∞·exp(Ea/|E|). 10-20 MV/cm for HZO.",
			Models:      lkModel,
		})
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// CONDUCTANCE: P→G mapping for analog level readout
	// G = Gmin + (Gmax-Gmin)·(P/Ps + 1)/2
	// ═══════════════════════════════════════════════════════════════════════════
	if mat.Conductance.GminS > 0 {
		props = append(props, FormattedProperty{
			Name:        "Gmin (HRS)",
			Value:       fmt.Sprintf("%.1f µS", mat.Conductance.GminS*1e6),
			RawValue:    mat.Conductance.GminS,
			Category:    CategoryConductance,
			Description: "High resistance state at P=-Ps. G = Gmin + (Gmax-Gmin)·(P/Ps+1)/2.",
			Models:      bothModels,
		})
	}
	if mat.Conductance.GmaxS > 0 {
		props = append(props, FormattedProperty{
			Name:        "Gmax (LRS)",
			Value:       fmt.Sprintf("%.1f µS", mat.Conductance.GmaxS*1e6),
			RawValue:    mat.Conductance.GmaxS,
			Category:    CategoryConductance,
			Description: "Low resistance state at P=+Ps. G = Gmin + (Gmax-Gmin)·(P/Ps+1)/2.",
			Models:      bothModels,
		})
	}
	// Show on/off ratio if both Gmin and Gmax are set
	if mat.Conductance.GminS > 0 && mat.Conductance.GmaxS > 0 {
		ratio := mat.Conductance.GmaxS / mat.Conductance.GminS
		props = append(props, FormattedProperty{
			Name:        "Gmax/Gmin",
			Value:       FormatConductanceRatio(ratio),
			RawValue:    ratio,
			Category:    CategoryConductance,
			Description: "On/off ratio. Higher → more distinguishable analog levels.",
			Models:      bothModels,
		})
	}

	return props
}

// GetPropertiesByCategory filters properties by category.
func GetPropertiesByCategory(props []FormattedProperty, category string) []FormattedProperty {
	filtered := []FormattedProperty{}
	for _, p := range props {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// CategoryOrder defines the display order matching the L-K equation structure.
var CategoryOrder = []string{
	CategoryCore,
	CategoryGeometry,
	CategoryLandau,
	CategoryAlpha,
	CategoryDepol,
	CategoryCircuit,
	CategoryNLS,
	CategoryConductance,
}

// HasCategory checks if any properties exist in the given category.
func HasCategory(props []FormattedProperty, category string) bool {
	for _, p := range props {
		if p.Category == category {
			return true
		}
	}
	return false
}

// TruncateString truncates a string to maxLen with ellipsis.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// WrapText wraps text to a maximum width (in characters).
func WrapText(s string, maxWidth int) string {
	if maxWidth <= 0 || len(s) <= maxWidth {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for i, word := range words {
		if i > 0 {
			if lineLen+1+len(word) > maxWidth {
				result.WriteString("\n")
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}
