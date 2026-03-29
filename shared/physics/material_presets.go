// material_presets.go provides hardcoded material preset constructors for the
// HZOMaterial type.  Each function returns a self-contained *HZOMaterial with
// literature-anchored (or explicitly flagged) simulation defaults.
//
// Extracted from material.go to keep individual files under 500 lines.
package physics

// DefaultHZO returns baseline parameters for Si-doped HfO2 (Hf0.5Zr0.5O2).
// Based on 10nm ALD films. Ref: Park et al., Adv. Mater. 27, 1811 (2015)
func DefaultHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO (Si-doped, Park 2015 midpoint)",
		Pr:        24.5e-2, // 24.5 μC/cm² (Park 2015: 15-34 range)
		Ps:        30e-2,   // 30 μC/cm²
		Ec:        1.2e8,   // 1.2 MV/cm
		Epsilon:   30,
		EpsilonLF: 38,
		LossAngle: 0.02,    // tan δ ≈ 2%
		Thickness: 10e-9,   // 10 nm
		Area:      100e-12, // 100 nm²
		Tau:       1e-9,    // ~1 ns
		Tau0:      1e-13,
		Ea:        0.7,
		Alpha:     2.0,
		// NOTE: CurieTemp and CurieConst here are empirical fits for the DefaultHZO
		// Preisach temperature-scaling path and differ from the Materlik 2015 LGD
		// values used in MaterlikHfO2() (Tc=598 K, C0=5.3e5 K). The Materlik values
		// are for the orthorhombic Pca21 phase free-energy surface; these defaults
		// are tuned for the phenomenological Curie-Weiss Ec(T)/Pr(T) scaling in
		// updateEffectiveParameters() and should not be substituted without
		// re-validating the temperature sweep behavior.
		CurieTemp:  723,   // Empirical — differs from Materlik 2015 (Tc=598 K for o-HfO2)
		CurieConst: 1.5e5, // Empirical — differs from Materlik 2015 (C0=5.3e5 K for o-HfO2)
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e10,   // 10^10 cycles
		RetentionTime:   3.15e9, // 100 years at 85°C
		ImrintField:     1e6,
		NumLevels:       30,     // Conference claim; pending peer review
		TargetRangeFrac: 0.90,
		Tau0NLS:         1e-10,  // 100 ps (HfO2 typical)
		EaNLS:           12e8,   // 12 MV/cm
		NLSSigma:        1.5,    // Guo et al., APL 2018
		Gmin:            1e-6,   // 1 µS HRS
		Gmax:            100e-6, // 100 µS LRS, ~100x on/off
		C2CSigmaBase:        0.03,    // 3% base C2C (Soliman 2023 / Reis 2023)
		C2CExponent:         0.5,     // sqrt(G/Gmin) scaling
		BetaLandau:          -6.720e8,  // Materlik 2015 LGD baseline
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,  // m⁴/C²
		Q12:                 -0.026, // see struct Q12 comment re: DFT discrepancy
	}
}

// MaterlikHfO2 returns an LK-focused HfO2 preset anchored to:
//
//	T. Materlik et al., J. Appl. Phys. 117, 134109 (2015), doi:10.1063/1.4916229
//
// using β=-6.72e8 J·m^5/C^4, γ=1.95e10 J·m^9/C^6, Tc≈598 K, C0≈5.3e5 K.
//
// Operating target used in this repository for validation:
// Pr≈20 µC/cm², Ec≈1.0 MV/cm at 300 K and 10 nm thickness.
func MaterlikHfO2() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HfO2 (Materlik 2015)",
		Pr:              20e-2,
		Ps:              25e-2,
		Ec:              1.5e8,
		Epsilon:         30,
		EpsilonLF:       38,
		LossAngle:       0.02,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             1e-9,
		Tau0:            1e-13,
		Ea:              0.7,
		Alpha:           2.0,
		CurieTemp:       598,
		CurieConst:      5.3e5,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e10,
		RetentionTime:   3.15e9,
		ImrintField:     1e6,
		NumLevels:       30,
		TargetRangeFrac: 0.90,
		Tau0NLS:         1e-10,
		EaNLS:           12e8,
		NLSSigma:        1.5,
		Gmin:            1e-6,
		Gmax:            100e-6,
		BetaLandau:      -6.720e8,
		GammaLandau:     1.950e10,
		RhoViscosity:    0.05,
		K_dep:           2.5e8,
		Q11:             0.089,
		Q12:             -0.026,
	}
}

// LiteratureSuperlattice returns THEORETICAL BEST values from published HfO2/ZrO2
// superlattice research (NOT FeCIM demonstrated performance).
// Best-in-class HZO/ZrO₂ nanolaminate (1nm layers); PMC 12254504; RSC Nanoscale 2025.
func LiteratureSuperlattice() *HZOMaterial {
	return &HZOMaterial{
		Name:            "Literature Superlattice (HZO nanolaminate 2025)",
		Pr:              22e-2,  // 22 µC/cm² (HZO/ZrO₂ 1nm nanolaminate, 2025); revised from Cheema 2020 (50 µC/cm²)
		Ps:              27e-2,  // 27 µC/cm² (≈1.2×Pr)
		Ec:              0.85e8, // 0.85 MV/cm (Nature Commun. 2025) [3]
		Epsilon:         35,
		EpsilonLF:       50,
		LossAngle:       0.015,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             0.36e-9, // 360 ps RECORD (Purdue sub-µm arrays) [8]
		Tau0:            1e-13,
		Ea:              0.5, // Lower activation energy
		Alpha:           2.5,
		CurieTemp:       773, // Higher Tc
		CurieConst:      1.5e5,
		TempCoeffEc:     -1.5e5,
		TempCoeffPr:     -3e-5,
		EnduranceCycles: 1e10,   // 10^10 verified; 10^12 (La-doped, ACS AMI 2024)
		RetentionTime:   3.15e8, // 10y at 85°C verified
		ImrintField:     0.5e6,
		NumLevels:       64,      // Enhanced superlattice
		TargetRangeFrac: 0.90,
		Tau0NLS:         0.5e-10, // 50 ps
		EaNLS:           10e8,    // 10 MV/cm
		NLSSigma:        1.5,
		Gmin:                0.5e-6, // 0.5 µS HRS
		Gmax:                150e-6, // 150 µS LRS
		BetaLandau:          -3.8e8, // Superlattice-enhanced LGD
		GammaLandau:         2.1e10,
		RhoViscosity:        0.02,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.5,
		Q11:                 0.089,
		Q12:                 -0.035, // Stronger coupling in superlattice
	}
}

// FeCIMMaterial returns parameters matching FeCIM specifications (Tour, Nov 2024).
//
// DEMONSTRATED: 30 states (COSM 2025, pending peer review), 10^9 endurance, 10^7s retention.
// TARGET (not achieved): 10^12 endurance, 100y retention.
// ESTIMATED (not disclosed): Pr, Ps, Ec, most other params use HZO literature values.
// See: opensource/papers/08_Documentation/HONESTY_AUDIT.md
func FeCIMMaterial() *HZOMaterial {
	return &HZOMaterial{
		Name:            "FeCIM HZO",
		Pr:        30e-2,         // ESTIMATED
		Ps:        35e-2,         // ESTIMATED
		Ec:        1.0e8,         // ESTIMATED
		Epsilon:   32,            // ESTIMATED
		EpsilonLF: 40,            // ESTIMATED
		LossAngle: 0.01,          // ESTIMATED
		Thickness: 10e-9,         // ESTIMATED
		Area:      45e-9 * 45e-9, // 45nm pitch - ESTIMATED
		Tau:       10e-9,         // 10 ns (Tour's slides, not 1ns)
		Tau0:            1e-13,
		Ea:              0.6,
		Alpha:           2.0,
		CurieTemp:       723,
		CurieConst:      1.5e5,
		TempCoeffEc:     -1.8e5,
		TempCoeffPr:     -4e-5,
		EnduranceCycles: 1e9, // DEMONSTRATED (10^12 is TARGET)
		RetentionTime:   1e7, // DEMONSTRATED (~116 days)
		ImrintField:     1e6,
		NumLevels:       30,    // Conference claim; pending peer review
		TargetRangeFrac: 0.90,
		Tau0NLS:         1e-10,
		EaNLS:           12e8,
		NLSSigma:        1.5,
		Gmin:                1e-6,   // 1 µS HRS
		Gmax:                100e-6, // 100 µS LRS, ~100x on/off
		C2CSigmaBase:        0.03,   // 3% base C2C (Soliman 2023)
		C2CExponent:         0.5,
		BetaLandau:          -6.720e8, // Materlik 2015 LGD baseline
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05, // Model default for HZO; see Alessandri et al., IEEE EDL 2018 for range 0.005-0.05
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,  // m⁴/C²
		Q12:                 -0.026, // see struct Q12 comment re: DFT discrepancy
	}
}

// FeCIMMaterialTarget returns FeCIM TARGET specifications.
// WARNING: These are GOALS not demonstrated performance.
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

// CryogenicHZO returns HZO parameters at 4K (enhanced Pr: 75 vs ~30 µC/cm² at RT).
// Refs: Adv. Electron. Mater. 2024 (doi:10.1002/aelm.202300879), 2025 (202400840).
func CryogenicHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "Cryogenic HZO (4K)",
		Pr:        75e-2, // 75 µC/cm² at 4K - RECORD
		Ps:        80e-2,
		Ec:        1.5e8, // Higher Ec at cryo
		Epsilon:   28,
		EpsilonLF:       35,
		LossAngle: 0.008, // Lower loss at cryo
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:       1e-9, // ~1 ns maintained
		Tau0:            1e-13,
		Ea:              0.7,
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -2e5,
		TempCoeffPr:     -5e-5,
		EnduranceCycles: 1e9,     // Verified at cryo
		RetentionTime:   3.15e10, // >10y at cryo
		ImrintField:     0.3e6,
		NumLevels:       48,      // Enhanced Pr allows more levels
		TargetRangeFrac: 0.90,
		Tau0NLS:         2e-10,   // 200 ps (slower at cryo)
		EaNLS:           15e8,    // Higher barrier at cryo
		NLSSigma:        1.5,
		Gmin:                0.5e-6, // 0.5 µS HRS (lower leakage at cryo)
		Gmax:                200e-6, // 200 µS LRS
		BetaLandau:          -6.720e8, // Materlik 2015 LGD baseline; cryo-specific not widely published
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05,
		CurieConst:          1.5e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,
		Q12:                 -0.026,
	}
}

// HZOStandard32 returns parameters for standard HZO with 32 analog states.
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
		NumLevels:       32,   // VERIFIED (Oh et al. 2017)
		TargetRangeFrac: 0.90,
		Tau0NLS:         1e-10,
		EaNLS:           12e8,
		NLSSigma:        1.5,
		Gmin:                2e-6,     // 2 µS HRS (2017 era device)
		Gmax:                80e-6,   // 80 µS LRS, ~40x on/off
		BetaLandau:          -6.720e8, // Materlik 2015 LGD baseline
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05,
		CurieConst:          1.5e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,
		Q12:                 -0.026,
	}
}

// HZOFJT140 returns parameters for HZO FTJ with 140 states (ultrathin tunneling).
// Refs: Song et al., Adv. Science 2024; ACS AMI 2022.
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
		NumLevels:       140,  // VERIFIED (Song et al. 2024)
		TargetRangeFrac: 0.90,
		Tau0NLS:         2e-10,
		EaNLS:           10e8,
		NLSSigma:        1.5,
		Gmin:                0.1e-6,   // 0.1 µS HRS (FTJ tunneling)
		Gmax:                10e-6,    // 10 µS LRS (~100x on/off)
		BetaLandau:          -6.720e8, // Materlik 2015 LGD baseline
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05,
		CurieConst:          1.5e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,
		Q12:                 -0.026,
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
		TargetRangeFrac: 0.90,
		Tau0NLS:         1e-10,
		EaNLS:           12e8,
		NLSSigma:        1.5,
		Gmin:                1e-6,
		Gmax:                100e-6,
		BetaLandau:          -6.720e8, // Materlik 2015 LGD baseline
		GammaLandau:         1.950e10,
		RhoViscosity:        0.05,
		CurieConst:          1.5e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.5e8,
		StressGPa:           1.0,
		Q11:                 0.089,
		Q12:                 -0.026,
	}
}

// PZT returns parameters for Lead Zirconate Titanate (Pb[Zr_xTi_{1-x}]O3).
// LGD coefficients are educational placeholders; for calibrated PZT LGD see
// Haun et al., J. Appl. Phys. 62, 3331 (1987).
func PZT() *HZOMaterial {
	return &HZOMaterial{
		Name:                "PZT",
		Pr:                  30e-2, // 30 µC/cm²
		Ps:                  40e-2, // 40 µC/cm²
		Ec:                  6.0e6, // 60 kV/cm
		Epsilon:             450,
		EpsilonLF:           650,
		LossAngle:           0.03,
		Thickness:           100e-9,
		Area:                100e-12,
		Tau:                 20e-9,
		Tau0:                1e-12,
		Ea:                  0.4,
		Alpha:               2.0,
		CurieTemp:           673,
		TempCoeffEc:         -5e4,
		TempCoeffPr:         -2e-5,
		EnduranceCycles:     1e9,
		RetentionTime:       3.15e8,
		ImrintField:         1e6,
		NumLevels:           30,
		TargetRangeFrac:     0.90,
		Tau0NLS:             1e-10,
		EaNLS:               6e7,
		NLSSigma:            1.5,
		Gmin:                1e-6,
		Gmax:                120e-6,
		BetaLandau:          -2.0e8,  // Educational placeholder — not PZT-calibrated (see func comment)
		GammaLandau:         1.0e10,  // Educational placeholder — not PZT-calibrated (see func comment)
		RhoViscosity:        0.05,
		CurieConst:          1.0e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.0e8,
		StressGPa:           0.5,
		Q11:                 0.08,
		Q12:                 -0.03,
	}
}

// BTO returns parameters for Barium Titanate (BaTiO3). Educational preset;
// LGD coefficients are placeholders. For calibrated BTO LGD see
// Li et al., J. Appl. Phys. 98, 064101 (2005).
func BTO() *HZOMaterial {
	return &HZOMaterial{
		Name:                "BTO",
		Pr:                  20e-2, // 20 µC/cm²
		Ps:                  26e-2, // 26 µC/cm²
		Ec:                  3.0e6, // 30 kV/cm
		Epsilon:             800,
		EpsilonLF:           1200,
		LossAngle:           0.02,
		Thickness:           100e-9,
		Area:                100e-12,
		Tau:                 50e-9,
		Tau0:                1e-12,
		Ea:                  0.4,
		Alpha:               2.0,
		CurieTemp:           393, // ~120°C
		TempCoeffEc:         -2e4,
		TempCoeffPr:         -2e-5,
		EnduranceCycles:     1e8,
		RetentionTime:       3.15e8,
		ImrintField:         5e5,
		NumLevels:           30,
		TargetRangeFrac:     0.90,
		Tau0NLS:             1e-10,
		EaNLS:               3e7,
		NLSSigma:            1.5,
		Gmin:                1e-6,
		Gmax:                100e-6,
		BetaLandau:          -2.0e8,  // Educational placeholder — not BTO-calibrated (see func comment)
		GammaLandau:         1.0e10,  // Educational placeholder — not BTO-calibrated (see func comment)
		RhoViscosity:        0.05,
		CurieConst:          1.0e5,
		SeriesResistanceOhm: 50.0,
		K_dep:               2.0e8,
		StressGPa:           0.5,
		Q11:                 0.09,
		Q12:                 -0.03,
	}
}

// AlScN returns parameters for Aluminum Scandium Nitride (very high Pr/Ec,
// limits granularity to 8-16 levels).
// Refs: Nature Commun. 2025 (doi:10.1038/s41467-025-62904-6); APL Mater. 2023.
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
		NumLevels:       12, // 8-16 states (high Ec limits granularity)
		// AlScN switching defaults are placeholders pending calibration
		// against APL Materials (2023) and Nature Commun. (2025).
		Tau0NLS: 1e-11,  // 10 ps placeholder
		EaNLS:   22e8,   // 22 MV/cm placeholder
		Gmin:    0.2e-6, // 0.2 µS HRS
		Gmax:    300e-6, // 300 µS LRS
	}
}
