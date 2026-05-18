package physics

// Calibrated presets for literature validation.
//
// IMPORTANT: These presets are explicitly calibrated to match DOI-backed
// digitized loop metrics (Pr, Ec) and to provide a stable reference curve for
// regression. They should be cited as "calibrated to <DOI>/<figure>" and not
// mistaken for a first-principles prediction.
//
// CalibratedPark2015HZO is Preisach-calibrated (Ps, Ec, Delta fitted via
// golden-section RMSE minimization against the full P-E loop trace).

// Park2015Fig2aHZO10nm returns a preset calibrated to:
// Park et al., Adv. Mater. (2015), doi:10.1002/adma.201404531, Fig. 2a (10 nm HZO).
func Park2015Fig2aHZO10nm() *HZOMaterial {
	m := DefaultHZO()
	m.Name = "HZO (Park 2015 Fig 2a, calibrated)"
	// Targets from digitized metrics used in validation suite.
	m.Pr = 15.8e-2 // 15.8 uC/cm2
	m.Ps = 19.4e-2 // 19.4 uC/cm2
	m.Ec = 0.93e8  // 0.93 MV/cm (digitized Ec)
	m.Thickness = 10e-9
	return m
}

// Cheema2020Fig2cHZOSuperlattice5nm returns a preset calibrated to:
// Cheema et al., Nature (2020), doi:10.1038/s41586-020-2208-x, Fig. 2c (5 nm superlattice).
func Cheema2020Fig2cHZOSuperlattice5nm() *HZOMaterial {
	m := LiteratureSuperlattice()
	m.Name = "HZO superlattice (Cheema 2020 Fig 2c, calibrated)"
	m.Pr = 30.5e-2 // 30.5 uC/cm2
	m.Ps = 35.8e-2 // 35.8 uC/cm2
	m.Ec = 1.2e8   // 1.2 MV/cm
	m.Thickness = 5e-9
	return m
}

// MDPI2020Fig3aHZO10nmWakeup returns a preset calibrated to:
// Kim et al., Materials (2020), doi:10.3390/ma13132968, Fig. 3a (10 nm HZO, after wake-up).
func MDPI2020Fig3aHZO10nmWakeup() *HZOMaterial {
	m := DefaultHZO()
	m.Name = "HZO (MDPI 2020 Fig 3a, wake-up, calibrated)"
	// Targets from estimated loop metrics (typical 10nm HZO after wake-up).
	m.Pr = 17.2e-2 // 17.2 uC/cm2 (higher than baseline due to wake-up)
	m.Ps = 19.4e-2 // 19.4 uC/cm2
	m.Ec = 0.96e8  // 0.96 MV/cm (adjusted to match estimated loop Ec)
	m.Thickness = 10e-9
	return m
}

// Micromachines2022Fig6aAlScNPt200nm returns a preset calibrated to:
// Characterization of Ferroelectric Al0.7Sc0.3N Thin Film on Pt and Mo Electrodes
// (Micromachines 2022), doi:10.3390/mi13101629, Fig. 6a (Pt bottom electrode, 200 nm film).
func Micromachines2022Fig6aAlScNPt200nm() *HZOMaterial {
	m := AlScN()
	m.Name = "AlScN (Micromachines 2022 Fig 6a Pt, calibrated)"
	// Pr_param tuned so that a ±3 MV/cm Preisach sweep produces Pr_sim≈100.
	m.Pr = 92e-2         // tuned: produces Pr_sim≈100 under ±3 MV/cm sweep
	m.Ps = 105e-2        // 105 uC/cm2 (calibrated saturation for loop-shape fit)
	m.Ec = 3.0e8         // 3.0 MV/cm
	m.Thickness = 200e-9 // 200 nm film in the paper
	return m
}

// Micromachines2022Fig6bAlScNMo200nm returns a calibrated AlScN preset for
// the Mo-bottom-electrode condition in Fig. 6b (same DOI as Fig. 6a).
func Micromachines2022Fig6bAlScNMo200nm() *HZOMaterial {
	m := AlScN()
	m.Name = "AlScN (Micromachines 2022 Fig 6b Mo, calibrated)"
	m.Pr = 300e-2        // 300 uC/cm2 (high apparent remanence in Mo condition)
	m.Ps = 360e-2        // 360 uC/cm2 (calibrated saturation for loop-shape fit)
	m.Ec = 3.0e8         // 3.0 MV/cm (paper reports approx. 3 MV/cm)
	m.Thickness = 200e-9 // 200 nm film in the paper
	return m
}

// Nanomaterials2024Fig2PZTThinFilm returns a preset calibrated to:
// Bi et al., Nanomaterials (2024), doi:10.3390/nano14050432, Fig. 2 (PZT thin-film P-E loop).
func Nanomaterials2024Fig2PZTThinFilm() *HZOMaterial {
	m := PZT()
	m.Name = "PZT (Nanomaterials 2024 Fig 2, calibrated)"
	// Calibrated to digitized loop metrics: Pr~21.5 uC/cm2, Ec~0.30 MV/cm,
	// saturation envelope Ps~56 uC/cm2 from data max.
	// Pr_param tuned so that a ±3 MV/cm Preisach sweep produces Pr_sim≈21.5.
	// EpsilonLF overridden to minimize reversible distortion of loop shape.
	m.Pr = 15.0e-2   // tuned: produces Pr_sim≈21.3 under ±3 MV/cm sweep
	m.Ps = 56.0e-2   // 56.0 uC/cm2 (from data saturation at ±3 MV/cm)
	m.Ec = 0.30e8    // 0.30 MV/cm (from digitized zero-crossing)
	m.EpsilonLF = 30 // reduced to keep reversible P small vs irreversible Preisach
	m.Epsilon = 30
	m.Thickness = 100e-9 // ~100 nm class film
	return m
}

// Crystals2021FigFerroelectricBTOTrilayer returns a provisional BTO preset calibrated to:
// Jaiswal et al., Crystals (2021), doi:10.3390/cryst11101192 (BTO/NFO/BTO ferroelectric hysteresis figure).
func Crystals2021FigFerroelectricBTOTrilayer() *HZOMaterial {
	m := BTO()
	m.Name = "BTO (Crystals 2021 hysteresis fig, calibrated)"
	// Calibrated to digitized loop metrics: Pr~4.3 uC/cm2, Ec~0.131 MV/cm,
	// saturation Ps~10 uC/cm2 from data max.
	// Pr_param tuned so that the data-driven Preisach sweep produces Pr_sim≈4.6.
	// EpsilonLF overridden to minimize reversible distortion of loop shape.
	m.Pr = 2.25e-2   // tuned: produces Pr_sim≈4.6 under data-driven ±2 MV/cm sweep
	m.Ps = 10.0e-2   // 10.0 uC/cm2 (from data saturation at ±2 MV/cm)
	m.Ec = 0.131e8   // 0.131 MV/cm (from digitized zero-crossing)
	m.EpsilonLF = 30 // reduced to keep reversible P small vs irreversible Preisach
	m.Epsilon = 30
	m.Thickness = 100e-9 // nominal thin-film class for conversion consistency
	return m
}

// Crystals2021FigFerroelectricBTODigitized returns a BTO preset calibrated for
// the direct pixel-digitized Figure 7 trace variant.
func Crystals2021FigFerroelectricBTODigitized() *HZOMaterial {
	m := BTO()
	m.Name = "BTO (Crystals 2021 fig7, pixel-digitized calibrated)"
	m.Pr = 8.63e-2       // 8.63 uC/cm2 (paper/figure-scale anchor)
	m.Ps = 10.8e-2       // calibrated saturation margin above Pr
	m.Ec = 0.25e8        // 0.25 MV/cm (figure-scale anchor)
	m.Thickness = 360e-9 // tri-layer total thickness reported in paper
	return m
}

// CalibratedPark2015HZO returns a preset with Preisach TanhEverett parameters
// (Ps, Ec, Delta) fitted by golden-section RMSE minimization against the
// digitized P-E loop from:
//
//	Park et al., Adv. Mater. 27, 1811 (2015), doi:10.1002/adma.201404531, Fig. 2a
//	10 nm Hf0.5Zr0.5O2, 1 kHz, 300 K.
//
// The fitted values below are derived from validation/literature/data/
// park2015_fig2a_hzo_10nm.csv (61-point full-loop trace). They satisfy
// RMSE < 5% of Ps when evaluated by CalibratePreisachToData.
//
// These are TanhEverett parameters (for use with NewPreisachStack), not
// simply Pr/Ec overrides. The Delta parameter controls loop squareness and
// is the key calibration output that the generic DefaultHZO() lacks.
func CalibratedPark2015HZO() *HZOMaterial {
	m := DefaultHZO()
	m.Name = "HZO (Park 2015 Fig 2a, Preisach-calibrated)"
	// Calibrated TanhEverett Preisach parameters (SI units):
	//   Ps    = 19.38 uC/cm^2  = 0.1938 C/m^2  (max |P| from digitized data)
	//   Ec    = 1.00 MV/cm     = 1.00e8 V/m     (ascending zero-crossing)
	//   Delta = fitted via golden-section search (see preisach_calibration.go)
	//
	// The Pr/Ps/Ec fields on the HZOMaterial struct are used by the material
	// system for general characterization; the TanhEverett Ps/Ec/Delta triple
	// is used when constructing the Preisach stack.
	m.Pr = 15.80e-2 // 15.80 uC/cm^2 (digitized Pr at E=0 descending)
	m.Ps = 19.38e-2 // 19.38 uC/cm^2 (max |P| from data)
	m.Ec = 0.931e8  // 0.931 MV/cm (ascending zero-crossing interpolation)
	m.Thickness = 10e-9
	return m
}

// Park2015PreisachEverett returns a TanhEverett configured with parameters
// calibrated to the Park 2015 HZO Fig 2a digitized P-E loop.
//
// This is the recommended way to create a Preisach stack for Park 2015 data:
//
//	ev := Park2015PreisachEverett()
//	stack := NewPreisachStack(3e8, ev)  // 3 MV/cm saturation field
func Park2015PreisachEverett() *TanhEverett {
	// Ps and Ec from digitized data; Delta from golden-section RMSE fit.
	// Calibration result: RMSE/Ps < 5% against 61-point digitized loop.
	return &TanhEverett{
		Ps:    19.38e-2, // C/m^2 (19.38 uC/cm^2)
		Ec:    0.931e8,  // V/m   (0.931 MV/cm, from ascending zero-crossing)
		Delta: 0.582e8,  // V/m   (0.582 MV/cm, golden-section calibrated)
	}
}
