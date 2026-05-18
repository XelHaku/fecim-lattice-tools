// Package physics provides shared physics utilities for FeCIM simulations.
package physics

import (
	"math"
	"sync"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/shared/logging"
)

// Package-level logger for material calculations
var matLog *logging.Logger

var (
	allMaterialsOnce   sync.Once
	allMaterialsCached []*HZOMaterial
)

func init() {
	matLog = logging.NewLogger("material")
}

// HZOMaterial contains material parameters for Hafnium-Zirconium Oxide.
//
// THREE-TIER MATERIAL SYSTEM:
//
//	DefaultHZO()           - Baseline: Standard Si-doped HZO from literature
//	FeCIMMaterial()        - FeCIM: conference-baseline demonstrated values (conservative)
//	FeCIMMaterialTarget()  - FeCIM: conference-claim targets (aspirational)
//	LiteratureSuperlattice() - Academic: Best-case from Cheema et al. 2020
//
// For honest educational visualization, prefer FeCIMMaterial() which uses
// only demonstrated values. Use LiteratureSuperlattice() to show what
// superlattice engineering CAN theoretically achieve.
//
// References:
// - Park et al., Adv. Mater. 2015 (HZO ferroelectricity)
// - Müller et al., Nano Lett. 2012 (HfO2 ferroelectric properties)
// - Pesic et al., Adv. Funct. Mater. 2016 (wake-up, fatigue)
// - Cheema et al., Nature 2020 (superlattice enhancement)
type HZOMaterial struct {
	Name string // Material name/identifier

	// Polarization parameters (C/m²)
	Pr float64 // Remanent polarization
	Ps float64 // Saturation polarization

	// Field parameters (V/m)
	Ec float64 // Coercive field

	// Dielectric properties
	Epsilon   float64 // Relative permittivity (high frequency)
	EpsilonLF float64 // Low frequency permittivity
	LossAngle float64 // Dielectric loss tangent (tan δ)

	// Film properties
	Thickness float64 // Film thickness (m)
	Area      float64 // Active area (m²)

	// Dynamics
	Tau   float64 // Characteristic switching time (s)
	Tau0  float64 // Activation attempt frequency inverse (s)
	Ea    float64 // Activation energy for switching (eV)
	Alpha float64 // Switching exponent (KAI model)

	// Depolarization (Polycrystalline Analog Behavior)
	// Physics rule: E_dep = -k_dep·P. For stack-derived models, k_dep can be estimated from
	// k_dep ≈ (ε_FE/ε_dead)·(d_dead/d_FE) with dead-layer assumptions.
	K_dep float64 // Depolarization coefficient (V*m/C); educational assumed value commonly set in HZO compact models (~1-5×10^8 V·m/C depending on dead-layer stack), e.g., Hoffmann et al., Adv. Funct. Mater. 2016

	// Landau-Khalatnikov parameters (dynamic equation)
	// Materlik et al., J. Appl. Phys. 117, 134109 (2015) report LGD coefficients
	// for orthorhombic ferroelectric HfO2 (Pca21), commonly used as a baseline:
	// β ≈ -6.72×10^8 J·m^5/C^4, γ ≈ +1.95×10^10 J·m^9/C^6,
	// with Curie-Weiss α(T) parameterized by Tc≈598 K and C0≈5.3×10^5 K.
	BetaLandau   float64 // Landau β coefficient (J m^5 / C^4)
	GammaLandau  float64 // Landau γ coefficient (J m^9 / C^6)
	RhoViscosity float64 // Viscosity / damping (Ohm·m), calibrated dynamic parameter

	// Circuit parasitics (for series resistance coupling)
	SeriesResistanceOhm float64 // Series resistance (Ohms)

	// Temperature properties
	CurieTemp   float64 // Curie temperature (K)
	CurieConst  float64 // Curie constant (K)
	TempCoeffEc float64 // Temperature coefficient of Ec (V/m/K)
	TempCoeffPr float64 // Temperature coefficient of Pr (C/m²/K)

	// Mechanical stress (for electrostriction coupling)
	StressGPa float64 // In-plane tensile stress (GPa)

	// Reliability parameters
	EnduranceCycles float64 // Endurance limit (cycles)
	RetentionTime   float64 // Retention time at 85°C (s)
	ImrintField     float64 // Imprint field shift (V/m)

	// Analog state capability
	NumLevels       int     // Number of discrete analog states (0 = use default 30)
	TargetRangeFrac float64 // Fraction of Ps used for outer level targets (0..1)

	// NLS (Nucleation-Limited Switching) parameters for Merz law dynamics
	// tau(E) = Tau0NLS * exp(EaNLS / |E|)
	// These are per-material since different ferroelectrics have different switching behavior.
	// Literature anchors for NLS-style switching-time parameters in HfO2/HZO ferroelectrics:
	//   - Müller et al., IEEE Trans. Electron Devices, 60(12), 2013 (switching-time statistics/kinetics)
	//   - Trentzsch et al., IEDM 2016 (HfO2-based FeFET switching dynamics and compact-model fitting)
	Tau0NLS  float64 // Attempt time for NLS (s), typically 1e-10 to 1e-12 for HfO2
	EaNLS    float64 // Activation field for NLS (V/m), typically 10-15 MV/cm for HfO2
	NLSSigma float64 // Log-normal switching-time sigma (dimensionless), default 1.5 for HfO2 per Guo et al., APL 112, 262903 (2018)

	// FeFET conductance parameters (for CIM applications)
	// G = Gmin + (Gmax-Gmin) * (P/Ps + 1) / 2
	// Default 10:1 to 100:1 windows in this repo are simulation-facing presets.
	// Educational placeholder note: this repo often uses 10:1-100:1 windows for demos;
	// production FeFET ON/OFF spans ~10^2 to 10^5 depending on stack/cycling.
	// Literature examples: Jerry et al., IEDM 2017; Ni et al., IEEE Electron Device Lett. 2018.
	Gmin float64 // Minimum conductance (S) at P = -Ps (HRS state)
	Gmax float64 // Maximum conductance (S) at P = +Ps (LRS state)
	// Conductance transfer model: linear (default), subthreshold, saturation.
	ConductanceModel string
	KvT              float64 // ΔVt scale (V): ΔVt = KvT*(P/Ps)
	VGSReadV         float64 // Read-bias gate voltage (V), used by saturation model
	VT0V             float64 // Zero-polarization threshold voltage (V), used by saturation model

	// State-dependent cycle-to-cycle (C2C) variation (LIT-P1-06/07)
	// sigma(G) = C2CSigmaBase * (G / Gmin)^C2CExponent
	// Inspired by 28nm HKMG FeFET characterization showing state-dependent C2C:
	// - C2CSigmaBase ≈ 0.03 (3% base relative variation at Gmin state)
	// - C2CExponent ≈ 0.5 (sqrt dependence, weaker variation at high G)
	// - At HRS (Gmin): sigma = C2CSigmaBase * Gmin = 3% of Gmin
	// - At LRS (Gmax): sigma = C2CSigmaBase * sqrt(Gmax/Gmin) * Gmin
	// Refs:
	//   Soliman et al., Nature Commun. 14, 6348 (2023), doi:10.1038/s41467-023-42110-y
	//     — 28nm HKMG FeFET crossbar, multi-level cell characterization
	//   Reis et al., arXiv:2312.15444 (2023)
	//     — state-dependent C2C/D2D variation model from 28nm FeFET measurements
	// NOTE: sigma_base=0.03 and exponent=0.5 are representative educational defaults
	// derived from the above papers' qualitative findings, not a direct numerical
	// extraction. Treat as [CALIBRATION NEEDED] for quantitative use.
	C2CSigmaBase float64 // Base relative C2C variation (dimensionless, typ. 0.02-0.05)
	C2CExponent  float64 // State-dependence exponent (typ. 0.5 for sqrt, 1.0 for linear)

	// Electrostrictive coefficients (m⁴/C²) for strain-polarization coupling
	// Used to calculate Ec shift under substrate strain:
	//   Ec_eff = Ec * (1 + strainFactor * strain)
	//   strainFactor ≈ 2 * Q11 * Ec / Ec = 2 * Q11 (simplified)
	// Ref: Park et al., J. Appl. Phys. 117, 074103 (2015)
	Q11 float64 // Longitudinal electrostrictive coefficient
	// Q12: transverse electrostrictive coefficient.
	// Default -0.026 m⁴/C² is an empirical calibration value for the Ec(stress) scaling
	// in updateEffectiveParameters(). DFT calculations for orthorhombic HfO2 report
	// much larger magnitudes (|Q12| ~ 0.5 m⁴/C²; see Materlik et al. 2015 and
	// Gong et al., arXiv:1811.09787). The smaller value here is retained because
	// changing it would alter the stress-sensitivity of Ec by ~20x and requires
	// re-validation of all temperature/stress sweep tests.
	Q12 float64 // Transverse electrostrictive coefficient [CALIBRATION NEEDED vs HfO2 DFT]
}

// AllMaterials returns a slice of all available CMOS-compatible materials.
// Use this to populate material selection dropdowns in the GUI.
// Prefers loading from config/physics.yaml, falls back to hardcoded values.
//
// Material construction is cached after the first load to avoid hot-path
// allocations in callers such as benchmarks and UI refresh loops.
func AllMaterials() []*HZOMaterial {
	allMaterialsOnce.Do(func() {
		// Try to load from YAML config first
		cfg, err := physics.Load()
		if err == nil && cfg != nil && len(cfg.Materials) > 0 {
			allMaterialsCached = AllMaterialsFromConfig(cfg)
			return
		}

		// Fallback to hardcoded values
		allMaterialsCached = []*HZOMaterial{
			DefaultHZO(),
			FeCIMMaterial(),
			FeCIMMaterialTarget(),
			LiteratureSuperlattice(),
			CryogenicHZO(),
			HZOStandard32(),
			HZOFJT140(),
			PZT(),
			BTO(),
			AlScN(),
		}
	})

	// Return a shallow copy to avoid accidental slice-header mutation by callers.
	materials := make([]*HZOMaterial, len(allMaterialsCached))
	copy(materials, allMaterialsCached)
	return materials
}

// AllMaterialsFromConfig loads all materials from the physics config.
// Returns materials in a consistent order for GUI display.
func AllMaterialsFromConfig(cfg *physics.Config) []*HZOMaterial {
	// Define order for consistent GUI display
	order := []string{
		"default_hzo",
		"fecim_hzo",
		"fecim_hzo_target",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"pzt",
		"bto",
		"alscn",
		"in2se3",
	}

	var materials []*HZOMaterial
	for _, name := range order {
		if mat := cfg.GetMaterial(name); mat != nil {
			materials = append(materials, MaterialFromConfig(mat, cfg))
		}
	}

	// Add any materials not in the predefined order
	for name, mat := range cfg.Materials {
		found := false
		for _, orderedName := range order {
			if name == orderedName {
				found = true
				break
			}
		}
		if !found {
			materials = append(materials, MaterialFromConfig(mat, cfg))
		}
	}

	return materials
}

// MaterialFromConfig converts a physics.Material to an HZOMaterial.
func MaterialFromConfig(m *physics.Material, cfg *physics.Config) *HZOMaterial {
	// NLS hydration fallback: some config profiles omit the `nls` block.
	// Keep material kinetics usable in Arrhenius/Merz validation by deriving
	// conservative defaults from core switching fields instead of zeroing out.
	tau0NLS := m.NLS.TauInfS
	eaNLS := m.NLS.ActivationFieldVM
	nlsSigma := m.NLS.Sigma
	if tau0NLS <= 0 {
		tau0NLS = m.Tau0S
	}
	if eaNLS <= 0 {
		// Match legacy AlScN/FE fallback behavior (~4.4*Ec for activation field)
		// while remaining generic for any config-driven material lacking NLS.
		eaNLS = 4.4 * m.EcVM
	}
	if nlsSigma <= 0 {
		nlsSigma = 1.5
	}

	return &HZOMaterial{
		Name:                m.Name,
		Pr:                  m.PrCM2,
		Ps:                  m.PsCM2,
		Ec:                  m.EcVM,
		Epsilon:             m.EpsilonHF,
		EpsilonLF:           m.EpsilonLF,
		LossAngle:           m.LossTangent,
		Thickness:           m.ThicknessM,
		Area:                m.AreaM2,
		Tau:                 m.TauS,
		Tau0:                m.Tau0S,
		Ea:                  m.ActivationEnergyEV,
		Alpha:               m.KAIExponent,
		CurieTemp:           m.CurieTempK,
		TempCoeffEc:         m.TempCoeffEc,
		TempCoeffPr:         m.TempCoeffPr,
		EnduranceCycles:     m.EnduranceCycles,
		RetentionTime:       m.RetentionTimeS,
		ImrintField:         m.ImprintFieldVM,
		NumLevels:           m.GetNumLevels(cfg),
		TargetRangeFrac:     m.TargetRangeFrac,
		K_dep:               m.DepolarizationFactorVMC,
		BetaLandau:          m.Thermodynamics.BetaLandau,
		GammaLandau:         m.Thermodynamics.GammaLandau,
		RhoViscosity:        m.Thermodynamics.RhoViscosity,
		CurieConst:          m.Thermodynamics.CurieConstK,
		Q11:                 m.Coupling.Q11Electrostriction,
		Q12:                 m.Coupling.Q12Electrostriction,
		StressGPa:           m.Coupling.StressGPa,
		SeriesResistanceOhm: m.Circuit.SeriesResistanceOhm,
		Tau0NLS:             tau0NLS,
		EaNLS:               eaNLS,
		NLSSigma:            nlsSigma,
		Gmin:                m.Conductance.GminS,
		Gmax:                m.Conductance.GmaxS,
		ConductanceModel:    m.Conductance.Model,
		KvT:                 m.Conductance.KvT,
		VGSReadV:            m.Conductance.VGSReadV,
		VT0V:                m.Conductance.VT0V,
	}
}

// GetNumLevels returns the number of discrete analog states for this material.
// Returns 30 as default if not specified.
func (m *HZOMaterial) GetNumLevels() int {
	if m.NumLevels <= 0 {
		return DefaultLevels // Use shared constant
	}
	return m.NumLevels
}

// CoerciveVoltage returns the coercive voltage for a given thickness.
func (m *HZOMaterial) CoerciveVoltage() float64 {
	matLog.Input("CoerciveVoltage", map[string]interface{}{
		"Ec":        m.Ec,
		"thickness": m.Thickness,
	})

	voltage := m.Ec * m.Thickness

	matLog.Calculation("CoerciveVoltage", map[string]interface{}{
		"Ec":        m.Ec,
		"thickness": m.Thickness,
	}, voltage)

	return voltage
}

// RemanentCharge returns the remanent charge per unit area.
func (m *HZOMaterial) RemanentCharge() float64 {
	return m.Pr
}

// Capacitance returns the capacitance of the ferroelectric layer.
func (m *HZOMaterial) Capacitance() float64 {
	epsilon0 := 8.854e-12 // F/m
	return epsilon0 * m.Epsilon * m.Area / m.Thickness
}

// SwitchingEnergy returns the energy required for complete switching.
// E = 2 * Pr * Ec * Volume
func (m *HZOMaterial) SwitchingEnergy() float64 {
	matLog.Input("SwitchingEnergy", map[string]interface{}{
		"Pr":        m.Pr,
		"Ec":        m.Ec,
		"area":      m.Area,
		"thickness": m.Thickness,
	})

	volume := m.Area * m.Thickness
	energy := 2 * m.Pr * m.Ec * volume

	matLog.Calculation("SwitchingEnergy", map[string]interface{}{
		"volume": volume,
		"Pr":     m.Pr,
		"Ec":     m.Ec,
	}, energy)

	return energy
}

// SwitchingTime returns temperature-dependent switching time.
// τ(T) = τ0 * exp(Ea / kT)
func (m *HZOMaterial) SwitchingTime(T float64) float64 {
	kB := 8.617e-5 // eV/K
	return m.Tau0 * math.Exp(m.Ea/(kB*T))
}

// CoerciveFieldAtTemp returns temperature-dependent coercive field.
// Ec(T) = Ec0 * (1 - T/Tc)^0.5
func (m *HZOMaterial) CoerciveFieldAtTemp(T float64) float64 {
	if T >= m.CurieTemp {
		return 0
	}
	return m.Ec * math.Pow(1-T/m.CurieTemp, 0.5)
}

// PolarizationAtTemp returns temperature-dependent remanent polarization.
// Pr(T) = Pr0 * (1 - T/Tc)^0.5
func (m *HZOMaterial) PolarizationAtTemp(T float64) float64 {
	if T >= m.CurieTemp {
		return 0
	}
	return m.Pr * math.Pow(1-T/m.CurieTemp, 0.5)
}

// EnduranceAtCycles returns estimated Pr degradation after N cycles.
// Using stretched exponential model: Pr(N) = Pr0 * exp(-(N/N0)^β)
func (m *HZOMaterial) EnduranceAtCycles(N float64) float64 {
	beta := 0.3 // Stretched exponential factor
	return m.Pr * math.Exp(-math.Pow(N/m.EnduranceCycles, beta))
}

// RetentionAtTime returns estimated Pr retention after time t at temperature T.
// Using Arrhenius model for retention loss.
func (m *HZOMaterial) RetentionAtTime(t, T float64) float64 {
	// Acceleration factor for temperature
	Tref := 358.0 // 85°C reference
	kB := 8.617e-5
	Ea := 1.0 // Retention activation energy (eV)

	accel := math.Exp(Ea / kB * (1/Tref - 1/T))
	effectiveTime := t * accel

	if effectiveTime >= m.RetentionTime {
		return 0.9 * m.Pr // 10% loss at end of life
	}
	return m.Pr * (1 - 0.1*effectiveTime/m.RetentionTime)
}

// DiscreteLevel returns the conductance for a given state level (0 to totalLevels-1).
// Uses FeFET physics model: conductance is linearly modulated by polarization.
//
// Physics basis:
//   - Polarization P varies from -Ps to +Ps across levels
//   - FeFET threshold voltage: ΔVth ∝ P (polarization shifts channel threshold)
//   - Channel conductance: G ∝ (Vgs - Vth) in linear region
//   - Net effect: G = Gmin + (Gmax-Gmin) * (P/Ps + 1) / 2
//
// This matches the Preisach model's polarizationToConductance() for consistency.
func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64 {
	if totalLevels <= 1 {
		return (m.Gmin + m.Gmax) / 2 // Return midpoint for degenerate case
	}

	// Map level to normalized polarization: level 0 → P/Ps = -1, level N-1 → P/Ps = +1
	normalizedP := -1.0 + 2.0*float64(level)/float64(totalLevels-1)

	// Get conductance bounds from material (use defaults if not set)
	Gmin := m.Gmin
	Gmax := m.Gmax
	if Gmin == 0 && Gmax == 0 {
		// Fallback to typical FeFET values if material doesn't specify
		Gmin = 1e-6   // 1 µS (high resistance state, P = -Ps)
		Gmax = 100e-6 // 100 µS (low resistance state, P = +Ps)
	}

	// FeFET conductance model with selectable transfer behavior.
	model := ParseConductanceModel(m.ConductanceModel)
	// Use normalized polarization directly so DiscreteLevel remains well-defined
	// even when Ps is unset/zero in lightweight test fixtures.
	return PolarizationToConductanceWithParams(
		normalizedP,
		1.0,
		Gmin,
		Gmax,
		model,
		m.KvT,
		m.VGSReadV,
		m.VT0V,
	)
}

// C2CVariation returns the cycle-to-cycle conductance standard deviation (S) at state G.
//
// State-dependent C2C variation model (LIT-P1-06/07):
//
//	sigma(G) = C2CSigmaBase * Gmin * (G/Gmin)^C2CExponent
//
// Inspired by 28nm HKMG FeFET data (Soliman et al., Nature Commun. 14, 6348, 2023;
// Reis et al., arXiv:2312.15444, 2023):
//   - At HRS (G=Gmin): sigma = C2CSigmaBase * Gmin  (base variation)
//   - At LRS (G=Gmax): sigma = C2CSigmaBase * Gmin * (Gmax/Gmin)^0.5  (sqrt scaling)
//   - Exponent 0.5 means weaker relative variation at high conductance states
//
// NOTE: numerical defaults are educational; treat as [CALIBRATION NEEDED].
//
// If C2CSigmaBase is zero, returns 0 (no C2C variation modeled).
func (m *HZOMaterial) C2CVariation(G float64) float64 {
	if m.C2CSigmaBase == 0 || m.Gmin == 0 {
		return 0
	}
	exp := m.C2CExponent
	if exp == 0 {
		exp = 0.5 // Default: sqrt dependence (Soliman 2023 / Reis 2023)
	}
	return m.C2CSigmaBase * m.Gmin * math.Pow(G/m.Gmin, exp)
}

// C2CVariationRelative returns the relative C2C variation (sigma/G) at state G.
// A value of 0.05 means 5% standard deviation relative to the conductance level.
func (m *HZOMaterial) C2CVariationRelative(G float64) float64 {
	if G == 0 {
		return 0
	}
	return m.C2CVariation(G) / G
}
