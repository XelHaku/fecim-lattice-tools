// Package widgets provides reusable UI components.
package widgets

import (
	"fmt"
	"math"
	"strings"

	"fecim-lattice-tools/config/physics"
)

// Property categories for organizing material data display.
const (
	CategoryPolarization = "Polarization"
	CategoryField        = "Field"
	CategoryDielectric   = "Dielectric"
	CategoryGeometry     = "Geometry"
	CategoryDynamics     = "Dynamics"
	CategoryTemperature  = "Temperature"
	CategoryReliability  = "Reliability"
	CategorySpecial      = "Special"
)

// ModelUsage indicates which physics models use a parameter.
type ModelUsage struct {
	Preisach bool // Used in Preisach hysteresis model
	// Landau   bool // Reserved for future Landau model implementation
}

// String returns the model indicator string (e.g., "[P]" or "").
func (m ModelUsage) String() string {
	if m.Preisach {
		return "[P]"
	}
	return ""
}

// FormattedProperty holds a material property with display formatting.
type FormattedProperty struct {
	Name        string     // Display name (e.g., "Remanent Polarization")
	Value       string     // Formatted value with units (e.g., "25 µC/cm²")
	RawValue    float64    // Raw value for sorting/comparison
	Category    string     // Physics category
	Description string     // Tooltip/help text
	Models      ModelUsage // Which physics models use this parameter
}

// FormatPolarization converts C/m² to µC/cm² display string.
func FormatPolarization(cM2 float64) string {
	// C/m² to µC/cm²: multiply by 100
	microCcm2 := cM2 * 100
	if microCcm2 >= 100 {
		return fmt.Sprintf("%.0f µC/cm²", microCcm2)
	}
	return fmt.Sprintf("%.1f µC/cm²", microCcm2)
}

// FormatField converts V/m to MV/cm display string.
func FormatField(vM float64) string {
	// V/m to MV/cm: divide by 1e8
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
	// Very small areas (sub-nm²)
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
	if s < 31536000 { // < 1 year
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
		// Clean power of 10
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
	if celsius < 0 {
		return fmt.Sprintf("%.0f K (%.0f°C)", k, celsius)
	}
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

// preisachModel is a helper for marking Preisach model parameters.
var preisachModel = ModelUsage{Preisach: true}

// GetMaterialProperties extracts all properties from a Material into formatted display structs.
// Properties used in the Preisach hysteresis model are marked with [P].
func GetMaterialProperties(mat *physics.Material) []FormattedProperty {
	props := []FormattedProperty{}

	// Polarization properties
	// Pr is used in Preisach for initialization logging and normalization
	props = append(props, FormattedProperty{
		Name:        "Remanent Polarization (Pr)",
		Value:       FormatPolarization(mat.PrCM2),
		RawValue:    mat.PrCM2,
		Category:    CategoryPolarization,
		Description: "Polarization remaining after field removal. Used in Preisach model initialization.",
		Models:      preisachModel,
	})
	// Ps is a primary Preisach parameter for saturation and normalization
	props = append(props, FormattedProperty{
		Name:        "Saturation Polarization (Ps)",
		Value:       FormatPolarization(mat.PsCM2),
		RawValue:    mat.PsCM2,
		Category:    CategoryPolarization,
		Description: "Maximum achievable polarization. Primary parameter in Preisach model for hysteresis loop normalization.",
		Models:      preisachModel,
	})
	if mat.AnalogStates > 0 {
		props = append(props, FormattedProperty{
			Name:        "Analog States",
			Value:       fmt.Sprintf("%d (%.1f bits)", mat.AnalogStates, math.Log2(float64(mat.AnalogStates))),
			RawValue:    float64(mat.AnalogStates),
			Category:    CategoryPolarization,
			Description: "Number of discrete programmable states achievable through partial polarization switching.",
		})
	}

	// Field properties
	// Ec is the primary Preisach parameter for coercive field distribution
	props = append(props, FormattedProperty{
		Name:        "Coercive Field (Ec)",
		Value:       FormatField(mat.EcVM),
		RawValue:    mat.EcVM,
		Category:    CategoryField,
		Description: "Field required to switch polarization. Primary Preisach parameter - sets distribution mean and width (σ = 0.25·Ec).",
		Models:      preisachModel,
	})
	if mat.MemoryWindowV > 0 {
		props = append(props, FormattedProperty{
			Name:        "Memory Window",
			Value:       FormatVoltage(mat.MemoryWindowV),
			RawValue:    mat.MemoryWindowV,
			Category:    CategoryField,
			Description: "Voltage window between on/off states. Determines margin for reliable state detection.",
		})
	}

	// Dielectric properties
	props = append(props, FormattedProperty{
		Name:        "Permittivity (HF)",
		Value:       FormatDimensionless(mat.EpsilonHF),
		RawValue:    mat.EpsilonHF,
		Category:    CategoryDielectric,
		Description: "High-frequency relative permittivity (ε∞). Affects capacitance at MHz-GHz frequencies.",
	})
	props = append(props, FormattedProperty{
		Name:        "Permittivity (LF)",
		Value:       FormatDimensionless(mat.EpsilonLF),
		RawValue:    mat.EpsilonLF,
		Category:    CategoryDielectric,
		Description: "Low-frequency relative permittivity (εs). Affects DC capacitance and charge storage.",
	})
	props = append(props, FormattedProperty{
		Name:        "Loss Tangent (tan δ)",
		Value:       FormatPercent(mat.LossTangent),
		RawValue:    mat.LossTangent,
		Category:    CategoryDielectric,
		Description: "Dielectric loss factor. Ratio of energy dissipated to energy stored per cycle.",
	})

	// Geometry properties
	// Thickness is used in advanced Preisach for voltage calculations
	props = append(props, FormattedProperty{
		Name:        "Film Thickness",
		Value:       FormatThickness(mat.ThicknessM),
		RawValue:    mat.ThicknessM,
		Category:    CategoryGeometry,
		Description: "Ferroelectric layer thickness. Used in Preisach DiscreteStates() for field-to-voltage conversion.",
		Models:      preisachModel,
	})
	props = append(props, FormattedProperty{
		Name:        "Cell Area",
		Value:       FormatArea(mat.AreaM2),
		RawValue:    mat.AreaM2,
		Category:    CategoryGeometry,
		Description: "Active device area. Determines total charge capacity (Q = P × A).",
	})
	if mat.CellPitchNm > 0 {
		props = append(props, FormattedProperty{
			Name:        "Cell Pitch",
			Value:       fmt.Sprintf("%.0f nm", mat.CellPitchNm),
			RawValue:    mat.CellPitchNm,
			Category:    CategoryGeometry,
			Description: "Center-to-center cell spacing. Determines array density and potential crosstalk.",
		})
	}

	// Dynamics properties
	// Tau is used in Preisach KAI dynamics for domain switching simulation
	props = append(props, FormattedProperty{
		Name:        "Switching Time (τ)",
		Value:       FormatTime(mat.TauS),
		RawValue:    mat.TauS,
		Category:    CategoryDynamics,
		Description: "Characteristic switching time constant. Used in Preisach KAI dynamics: P(t) = Ps·(1 - exp(-(t/τ)^n)).",
		Models:      preisachModel,
	})
	// Tau0 is used in Preisach Merz law for field-dependent switching time
	props = append(props, FormattedProperty{
		Name:        "Attempt Time (τ₀)",
		Value:       FormatTime(mat.Tau0S),
		RawValue:    mat.Tau0S,
		Category:    CategoryDynamics,
		Description: "Thermal activation attempt time. Used in Preisach Merz law: τ(E) = τ₀·exp(Ea/|E|).",
		Models:      preisachModel,
	})
	// Activation energy is used in Preisach Merz law
	props = append(props, FormattedProperty{
		Name:        "Activation Energy",
		Value:       FormatEnergy(mat.ActivationEnergyEV),
		RawValue:    mat.ActivationEnergyEV,
		Category:    CategoryDynamics,
		Description: "Energy barrier for domain nucleation. Used in Preisach NLS dynamics for field-dependent switching.",
		Models:      preisachModel,
	})
	// KAI exponent is used in Preisach domain switching simulation
	props = append(props, FormattedProperty{
		Name:        "KAI Exponent",
		Value:       FormatDimensionless(mat.KAIExponent),
		RawValue:    mat.KAIExponent,
		Category:    CategoryDynamics,
		Description: "Kolmogorov-Avrami-Ishibashi exponent (n). n≈2 for 2D domain growth, n≈3 for 3D. Used in Preisach dynamics.",
		Models:      preisachModel,
	})

	// Temperature properties
	// Curie temperature is used in advanced Preisach for temperature-dependent Ec
	props = append(props, FormattedProperty{
		Name:        "Curie Temperature",
		Value:       FormatTemperature(mat.CurieTempK),
		RawValue:    mat.CurieTempK,
		Category:    CategoryTemperature,
		Description: "Ferroelectric transition temperature. Used in Preisach: Ec(T) = Ec·(1 - T/Tc)^0.5.",
		Models:      preisachModel,
	})
	props = append(props, FormattedProperty{
		Name:        "Temp. Coeff. Ec",
		Value:       fmt.Sprintf("%.1e V/m/K", mat.TempCoeffEc),
		RawValue:    mat.TempCoeffEc,
		Category:    CategoryTemperature,
		Description: "Temperature dependence of coercive field. Linear approximation: Ec(T) ≈ Ec + α·(T - T₀).",
	})
	props = append(props, FormattedProperty{
		Name:        "Temp. Coeff. Pr",
		Value:       fmt.Sprintf("%.1e C/m²/K", mat.TempCoeffPr),
		RawValue:    mat.TempCoeffPr,
		Category:    CategoryTemperature,
		Description: "Temperature dependence of remanent polarization. Pr decreases as T approaches Tc.",
	})
	if mat.OperatingTempK > 0 {
		props = append(props, FormattedProperty{
			Name:        "Operating Temperature",
			Value:       FormatTemperature(mat.OperatingTempK),
			RawValue:    mat.OperatingTempK,
			Category:    CategoryTemperature,
			Description: "Designed operating temperature. Properties are optimized for this temperature.",
		})
	}

	// Reliability properties
	props = append(props, FormattedProperty{
		Name:        "Endurance",
		Value:       FormatEndurance(mat.EnduranceCycles),
		RawValue:    mat.EnduranceCycles,
		Category:    CategoryReliability,
		Description: "Maximum write cycles before significant degradation. HZO typically achieves 10⁹-10¹² cycles.",
	})
	props = append(props, FormattedProperty{
		Name:        "Retention Time",
		Value:       FormatTime(mat.RetentionTimeS),
		RawValue:    mat.RetentionTimeS,
		Category:    CategoryReliability,
		Description: "Data retention time at specified temperature. Typically >10 years at 85°C for mature devices.",
	})
	if mat.ImprintFieldVM > 0 {
		props = append(props, FormattedProperty{
			Name:        "Imprint Field",
			Value:       FormatField(mat.ImprintFieldVM),
			RawValue:    mat.ImprintFieldVM,
			Category:    CategoryReliability,
			Description: "Voltage shift from polarization aging. Causes asymmetric hysteresis loop over time.",
		})
	}

	// Special properties (FTJ, AlScN, etc.)
	if mat.TERRatio > 0 {
		props = append(props, FormattedProperty{
			Name:        "TER Ratio",
			Value:       FormatConductanceRatio(mat.TERRatio),
			RawValue:    mat.TERRatio,
			Category:    CategorySpecial,
			Description: "Tunneling electroresistance ratio for FTJ devices. High/low resistance state ratio.",
		})
	}
	// Gmax/Gmin ratio is used in advanced Preisach for conductance mapping
	if mat.GmaxGminRatio > 0 {
		props = append(props, FormattedProperty{
			Name:        "Gmax/Gmin Ratio",
			Value:       FormatConductanceRatio(mat.GmaxGminRatio),
			RawValue:    mat.GmaxGminRatio,
			Category:    CategorySpecial,
			Description: "Conductance on/off ratio. Used in Preisach polarizationToConductance() for FeFET modeling.",
			Models:      preisachModel,
		})
	}
	if mat.ScFraction > 0 {
		props = append(props, FormattedProperty{
			Name:        "Sc Fraction",
			Value:       fmt.Sprintf("%.0f%% (Al%.2fSc%.2fN)", mat.ScFraction*100, 1-mat.ScFraction, mat.ScFraction),
			RawValue:    mat.ScFraction,
			Category:    CategorySpecial,
			Description: "Scandium fraction in AlScN alloy. Higher Sc increases piezoelectric response but reduces stability.",
		})
	}
	if mat.TRLLevel > 0 {
		props = append(props, FormattedProperty{
			Name:        "TRL Level",
			Value:       fmt.Sprintf("TRL %d", mat.TRLLevel),
			RawValue:    float64(mat.TRLLevel),
			Category:    CategorySpecial,
			Description: "Technology Readiness Level (1-9). 1=basic research, 9=flight proven/production.",
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

// CategoryOrder defines the display order for categories.
var CategoryOrder = []string{
	CategoryPolarization,
	CategoryField,
	CategoryDielectric,
	CategoryGeometry,
	CategoryDynamics,
	CategoryTemperature,
	CategoryReliability,
	CategorySpecial,
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
