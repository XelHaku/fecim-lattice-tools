// Package physics provides shared physics utilities for FeCIM simulations.
package physics

import (
	"math"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/shared/logging"
)

// Package-level logger for material calculations
var matLog *logging.Logger

func init() {
	matLog = logging.NewLogger("material")
}

// HZOMaterial contains material parameters for Hafnium-Zirconium Oxide.
//
// THREE-TIER MATERIAL SYSTEM:
//
//	DefaultHZO()           - Baseline: Standard Si-doped HZO from literature
//	FeCIMMaterial()        - FeCIM: Dr. Tour's DEMONSTRATED values (conservative)
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
	K_dep float64 // Depolarization coefficient (V*m/C) - creates "slant" for 30-level operation

	// Landau-Khalatnikov parameters (dynamic equation)
	BetaLandau   float64 // Landau β coefficient (J m^5 / C^4)
	GammaLandau  float64 // Landau γ coefficient (J m^9 / C^6)
	RhoViscosity float64 // Viscosity / damping (Ohm·m)

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
	// These are per-material since different ferroelectrics have different switching behavior
	Tau0NLS float64 // Attempt time for NLS (s), typically 1e-10 to 1e-12 for HfO2
	EaNLS   float64 // Activation field for NLS (V/m), typically 10-15 MV/cm for HfO2

	// FeFET conductance parameters (for CIM applications)
	// G = Gmin + (Gmax-Gmin) * (P/Ps + 1) / 2
	// Based on FeFET channel conductance modulation by ferroelectric polarization
	Gmin float64 // Minimum conductance (S) at P = -Ps (HRS state)
	Gmax float64 // Maximum conductance (S) at P = +Ps (LRS state)

	// Electrostrictive coefficients (m⁴/C²) for strain-polarization coupling
	// Used to calculate Ec shift under substrate strain:
	//   Ec_eff = Ec * (1 + strainFactor * strain)
	//   strainFactor ≈ 2 * Q11 * Ec / Ec = 2 * Q11 (simplified)
	// Ref: Park et al., J. Appl. Phys. 117, 074103 (2015)
	Q11 float64 // Longitudinal electrostrictive coefficient
	Q12 float64 // Transverse electrostrictive coefficient
}

// DefaultHZO returns material parameters for typical Si-doped HfO2 (Hf0.5Zr0.5O2).
// This is the BASELINE for comparison - standard HZO without superlattice enhancement.
// Use this to show how FeCIM's superlattice improves over conventional HZO.
//
// Based on literature values for 10nm ALD-deposited films.
// Ref: Park et al., Adv. Mater. 27, 1811 (2015)
func DefaultHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO (Si-doped)",
		Pr:              25e-2,   // 25 μC/cm² = 0.25 C/m²
		Ps:              30e-2,   // 30 μC/cm² = 0.30 C/m²
		Ec:              1.2e8,   // 1.2 MV/cm = 1.2e8 V/m
		Epsilon:         30,      // High frequency permittivity
		EpsilonLF:       38,      // Low frequency (includes domain contribution)
		LossAngle:       0.02,    // tan δ ≈ 2%
		Thickness:       10e-9,   // 10 nm
		Area:            100e-12, // 100 nm² typical device
		Tau:             1e-9,    // ~1 ns intrinsic switching
		Tau0:            1e-13,   // Attempt frequency ~10 THz
		Ea:              0.7,     // Activation energy ~0.7 eV
		Alpha:           2.0,     // KAI exponent (2D growth)
		CurieTemp:       723,     // ~450°C
		CurieConst:      1.5e5,   // Curie constant (K)
		TempCoeffEc:     -2e5,    // Ec decreases with T
		TempCoeffPr:     -5e-5,   // Pr decreases with T
		EnduranceCycles: 1e10,    // 10^10 cycles
		RetentionTime:   3.15e9,  // 100 years at 85°C
		ImrintField:     1e6,     // Small imprint
		NumLevels:       30,      // Demo baseline (conference claim; pending peer review)
		Tau0NLS:         1e-10,   // 100 ps attempt time (HfO2 typical)
		EaNLS:           12e8,    // 12 MV/cm activation field
		Gmin:            1e-6,    // 1 µS at HRS (P = -Ps)
		Gmax:            100e-6,  // 100 µS at LRS (P = +Ps), ~100x on/off ratio
		// Landau-Khalatnikov parameters (Golden Set I, 10nm HZO)
		BetaLandau:          -2.160e8,
		GammaLandau:         1.653e10,
		RhoViscosity:        0.05,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8, // Depolarization factor for analog slope
		StressGPa:           1.0,
		Q11:                 0.089,  // Longitudinal electrostriction (m⁴/C²)
		Q12:                 -0.026, // Transverse electrostriction (m⁴/C²)
	}
}

// LiteratureSuperlattice returns parameters from academic literature for
// HfO2/ZrO2 superlattices. These are THEORETICAL BEST values from published
// research, NOT IronLattice/FeCIM demonstrated performance.
//
// Use this mode to understand what superlattice engineering CAN achieve
// according to academic papers. For FeCIM-specific values, use FeCIMMaterial().
//
// Ref: Cheema et al., Nature 580, 478 (2020) - "Enhanced ferroelectricity
// in ultrathin films grown directly on silicon"
func LiteratureSuperlattice() *HZOMaterial {
	return &HZOMaterial{
		Name:            "Literature Superlattice (Cheema 2020)",
		Pr:              45e-2, // 45 μC/cm² (superlattice enhanced)
		Ps:              50e-2, // 50 μC/cm²
		Ec:              0.8e8, // 0.8 MV/cm (reduced by negative capacitance)
		Epsilon:         35,
		EpsilonLF:       50,
		LossAngle:       0.015,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             0.5e-9, // Faster switching
		Tau0:            1e-13,
		Ea:              0.5, // Lower activation energy
		Alpha:           2.5,
		CurieTemp:       773, // Higher Tc
		CurieConst:      1.5e5,
		TempCoeffEc:     -1.5e5,
		TempCoeffPr:     -3e-5,
		EnduranceCycles: 1e12, // Literature best-case
		RetentionTime:   1e10, // Literature best-case
		ImrintField:     0.5e6,
		NumLevels:       64,      // Enhanced superlattice can achieve more states
		Tau0NLS:         0.5e-10, // 50 ps (faster switching)
		EaNLS:           10e8,    // 10 MV/cm (lower barrier)
		Gmin:            0.5e-6,  // 0.5 µS at HRS (lower due to better control)
		Gmax:            150e-6,  // 150 µS at LRS (higher Pr → better modulation)
		// Landau-Khalatnikov parameters (superlattice-enhanced)
		BetaLandau:          -3.8e8,
		GammaLandau:         2.1e10,
		RhoViscosity:        0.02,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.5,
		// Electrostriction (assumed HZO-like; not explicitly reported in Cheema 2020).
		Q11: 0.089,
		Q12: -0.026,
	}
}

// FeCIMMaterial returns parameters matching FeCIM specifications.
// Based on Dr. Tour's Nov 2024 presentation data.
//
// ⚠️ HONESTY NOTE: Some parameters are TARGETS not DEMONSTRATED values.
// Dr. Tour's presentation (Nov 2024) disclosed:
//
// CONFERENCE CLAIM (PENDING PEER REVIEW):
//   - 30 discrete analog states (COSM 2025)
//   - Peer-reviewed MNIST: 96.6-98.24% (Nature Commun. 2023, ScienceDirect 2025)
//   - 10^9 endurance cycles DEMONSTRATED
//   - 10^7 seconds retention DEMONSTRATED (~116 days)
//
// TARGET (NOT YET ACHIEVED):
//   - 10^12 endurance cycles (stated as target, not demonstrated)
//   - 100 year retention (not demonstrated)
//
// ESTIMATED (NOT DISCLOSED):
//   - Pr, Ps, Ec (using standard HZO literature values)
//   - Tau (slides showed 10ns, using 10ns not 1ns)
//   - All other parameters are literature-based estimates
//
// See: opensource/papers/08_Documentation/HONESTY_AUDIT.md
func FeCIMMaterial() *HZOMaterial {
	return &HZOMaterial{
		Name:            "FeCIM HZO",
		Pr:              30e-2,         // 30 μC/cm² - ESTIMATED (not disclosed)
		Ps:              35e-2,         // 35 μC/cm² - ESTIMATED (not disclosed)
		Ec:              1.0e8,         // 1.0 MV/cm - ESTIMATED (not disclosed)
		Epsilon:         32,            // ESTIMATED
		EpsilonLF:       40,            // ESTIMATED
		LossAngle:       0.01,          // Low loss - ESTIMATED
		Thickness:       10e-9,         // ESTIMATED
		Area:            45e-9 * 45e-9, // 45nm pitch cell - ESTIMATED
		Tau:             10e-9,         // 10 ns (from Tour's slides, not 1ns)
		Tau0:            1e-13,
		Ea:              0.6,
		Alpha:           2.0,
		CurieTemp:       723,
		CurieConst:      1.5e5,
		TempCoeffEc:     -1.8e5,
		TempCoeffPr:     -4e-5,
		EnduranceCycles: 1e9, // 10^9 DEMONSTRATED (10^12 is TARGET)
		RetentionTime:   1e7, // 10^7 sec (~116 days) DEMONSTRATED
		ImrintField:     1e6,
		NumLevels:       30,     // 30-level baseline (conference claim; pending peer review)
		Tau0NLS:         1e-10,  // 100 ps attempt time
		EaNLS:           12e8,   // 12 MV/cm activation field
		Gmin:            1e-6,   // 1 µS at HRS
		Gmax:            100e-6, // 100 µS at LRS, on/off ~100x
		// Landau-Khalatnikov parameters (Golden Set I, 10nm HZO)
		BetaLandau:          -2.160e8,
		GammaLandau:         1.653e10,
		RhoViscosity:        0.05,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,  // Longitudinal electrostriction (m⁴/C²)
		Q12:                 -0.026, // Transverse electrostriction (m⁴/C²)
	}
}

// FeCIMMaterialTarget returns FeCIM TARGET specifications.
// ⚠️ WARNING: These are GOALS not demonstrated performance.
// Dr. Tour explicitly stated: "We still have to get this up to the
// required 10^12 cycles."
func FeCIMMaterialTarget() *HZOMaterial {
	mat := FeCIMMaterial()
	mat.Name = "FeCIM HZO (TARGET - NOT DEMONSTRATED)"
	mat.EnduranceCycles = 1e12 // TARGET - not yet achieved
	mat.RetentionTime = 3.15e9 // TARGET - 100 years (not demonstrated)
	mat.Tau = 1e-9             // TARGET - 1ns (demonstrated ~10ns)
	mat.Tau0NLS = 0.5e-10      // 50 ps target
	mat.EaNLS = 10e8           // 10 MV/cm target
	return mat
}

// CryogenicHZO returns HZO parameters at 4K for quantum computing integration.
// At cryogenic temperatures, HZO shows ENHANCED polarization (75 µC/cm² vs ~30 at RT).
//
// Ref: Adv. Electron. Mater. 2024, doi:10.1002/aelm.202300879
// Ref: Adv. Electron. Mater. 2025, doi:10.1002/aelm.202400840
func CryogenicHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "Cryogenic HZO (4K)",
		Pr:              75e-2, // 75 µC/cm² at 4K - RECORD (enhanced from RT!)
		Ps:              80e-2, // 80 µC/cm²
		Ec:              1.5e8, // 1.5 MV/cm (higher Ec at cryo)
		Epsilon:         28,    // Slightly reduced at cryo
		EpsilonLF:       35,
		LossAngle:       0.008, // Lower loss at cryo
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             1e-9, // ~1 ns maintained
		Tau0:            1e-13,
		Ea:              0.7,
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e9,     // 10^9 at ±5V verified at cryo
		RetentionTime:   3.15e10, // >10 years at cryo (improved)
		ImrintField:     0.3e6,   // Reduced imprint at cryo
		NumLevels:       48,      // Enhanced polarization allows more levels
		Tau0NLS:         2e-10,   // 200 ps (slower at cryo)
		EaNLS:           15e8,    // 15 MV/cm (higher barrier at cryo)
		Gmin:            0.5e-6,  // 0.5 µS at HRS (lower leakage at cryo)
		Gmax:            200e-6,  // 200 µS at LRS (higher Pr → better modulation)
	}
}

// HZOStandard32 returns parameters for standard HZO demonstrating 32 analog states.
// This is the peer-reviewed benchmark from Oh et al. 2017.
//
// Ref: Oh et al., IEEE EDL 38(6), 732 (2017)
func HZOStandard32() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO Standard (32 states)",
		Pr:              20e-2, // 20 µC/cm²
		Ps:              25e-2, // 25 µC/cm²
		Ec:              1.0e8, // 1.0 MV/cm
		Epsilon:         28,
		EpsilonLF:       35,
		LossAngle:       0.02,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             10e-9, // 10 ns
		Tau0:            1e-13,
		Ea:              0.7,
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e8,    // 10^8 cycles (2017 era)
		RetentionTime:   3.15e8, // 10 years projected
		ImrintField:     1e6,
		NumLevels:       32,    // 32 states - VERIFIED (Oh et al. 2017)
		Tau0NLS:         1e-10, // 100 ps attempt time
		EaNLS:           12e8,  // 12 MV/cm activation field
		Gmin:            2e-6,  // 2 µS at HRS (2017 era device)
		Gmax:            80e-6, // 80 µS at LRS, ~40x on/off ratio
	}
}

// HZOFJT140 returns parameters for HZO Ferroelectric Tunnel Junction with 140 states.
// FTJs achieve highest state count through ultrathin tunneling barriers.
//
// Ref: Song et al., Adv. Science 2024, doi:10.1002/advs.202308588
// Ref: ACS Appl. Mater. Interfaces 2022, doi:10.1021/acsami.2c04441
func HZOFJT140() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO FTJ (140 states)",
		Pr:              25e-2, // 25 µC/cm²
		Ps:              30e-2, // 30 µC/cm²
		Ec:              1.2e8, // 1.2 MV/cm
		Epsilon:         30,
		EpsilonLF:       40,
		LossAngle:       0.015,
		Thickness:       4.5e-9, // 4.5 nm - ULTRATHIN for tunneling
		Area:            100e-12,
		Tau:             20e-9, // 20 ns switching
		Tau0:            1e-13,
		Ea:              0.5, // Lower barrier for tunneling
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e7,    // 10^7 cycles demonstrated
		RetentionTime:   3.15e8, // 10 years extrapolated
		ImrintField:     1e6,
		NumLevels:       140,    // 140 states - VERIFIED (Song et al. 2024)
		Tau0NLS:         2e-10,  // 200 ps (FTJ tunneling)
		EaNLS:           10e8,   // 10 MV/cm
		Gmin:            0.1e-6, // 0.1 µS at HRS (FTJ: tunneling → lower current)
		Gmax:            10e-6,  // 10 µS at LRS (FTJ: ~100x on/off but lower absolute)
	}
}

// HZOCustom14 returns a custom HZO configuration with 14 discrete levels.
// This provides a balance between precision and programming complexity.
// 14 levels = 3.81 bits/cell
func HZOCustom14() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO Custom (14 states)",
		Pr:              25e-2, // 25 µC/cm²
		Ps:              30e-2, // 30 µC/cm²
		Ec:              1.1e8, // 1.1 MV/cm
		Epsilon:         30,
		EpsilonLF:       38,
		LossAngle:       0.02,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             5e-9, // 5 ns
		Tau0:            1e-13,
		Ea:              0.65,
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e10,
		RetentionTime:   3.15e9,
		ImrintField:     1e6,
		NumLevels:       14, // 14 discrete levels (3.81 bits/cell)
		Tau0NLS:         1e-10,
		EaNLS:           12e8,
		Gmin:            1e-6,
		Gmax:            100e-6,
	}
}

// AlScN returns parameters for Aluminum Scandium Nitride.
// AlScN has VERY HIGH Pr (120 µC/cm²) but also very high Ec (5 MV/cm),
// which limits practical state granularity to 8-16 levels.
//
// Ref: Nature Commun. 2025, doi:10.1038/s41467-025-62904-6
// Ref: APL Materials 2023, doi:10.1063/5.0148068
func AlScN() *HZOMaterial {
	return &HZOMaterial{
		Name:            "AlScN (8-16 states)",
		Pr:              120e-2, // 120 µC/cm² - VERY HIGH (range: 100-172)
		Ps:              150e-2, // 150 µC/cm²
		Ec:              5.0e8,  // 5.0 MV/cm - VERY HIGH (limits granularity)
		Epsilon:         10,     // Lower permittivity than HZO
		EpsilonLF:       12,
		LossAngle:       0.01,
		Thickness:       20e-9, // 20 nm typical
		Area:            100e-12,
		Tau:             10e-9, // ~10 ns
		Tau0:            1e-13,
		Ea:              1.0, // Higher barrier
		Alpha:           2.0,
		CurieTemp:       1273, // >1000°C - excellent thermal stability
		TempCoeffEc:     -1e5,
		TempCoeffPr:     -2e-5,
		EnduranceCycles: 1e9, // 10^9 cycles
		RetentionTime:   3.15e8,
		ImrintField:     1e6,
		NumLevels:       12,     // 8-16 states (high Ec limits granularity)
		Tau0NLS:         1e-11,  // 10 ps (faster attempt time)
		EaNLS:           22e8,   // 22 MV/cm (higher Ec material)
		Gmin:            0.2e-6, // 0.2 µS at HRS (high Pr → excellent off state)
		Gmax:            300e-6, // 300 µS at LRS (very high Pr → strong modulation)
	}
}

// AllMaterials returns a slice of all available CMOS-compatible materials.
// Use this to populate material selection dropdowns in the GUI.
// Prefers loading from config/physics.yaml, falls back to hardcoded values.
func AllMaterials() []*HZOMaterial {
	// Try to load from YAML config first
	cfg, err := physics.Load()
	if err == nil && cfg != nil && len(cfg.Materials) > 0 {
		return AllMaterialsFromConfig(cfg)
	}

	// Fallback to hardcoded values
	return []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		FeCIMMaterialTarget(),
		LiteratureSuperlattice(),
		CryogenicHZO(),
		HZOStandard32(),
		HZOCustom14(),
		HZOFJT140(),
		AlScN(),
	}
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
		Tau0NLS:             m.NLS.TauInfS,
		EaNLS:               m.NLS.ActivationFieldVM,
		Gmin:                m.Conductance.GminS,
		Gmax:                m.Conductance.GmaxS,
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
	matLog.Input("DiscreteLevel", map[string]interface{}{
		"level":       level,
		"totalLevels": totalLevels,
	})

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

	// FeFET conductance model: G = Gmin + (Gmax-Gmin) * (normalizedP + 1) / 2
	// At P = -Ps (normalizedP = -1): G = Gmin
	// At P = +Ps (normalizedP = +1): G = Gmax
	conductance := Gmin + (Gmax-Gmin)*(normalizedP+1)/2

	matLog.Calculation("DiscreteLevel", map[string]interface{}{
		"level":       level,
		"normalizedP": normalizedP,
		"Gmin":        Gmin,
		"Gmax":        Gmax,
	}, conductance)

	return conductance
}
