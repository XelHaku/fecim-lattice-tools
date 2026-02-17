package physics

// Calibrated presets for literature validation.
//
// IMPORTANT: These presets are explicitly calibrated to match DOI-backed
// digitized loop metrics (Pr, Ec) and to provide a stable reference curve for
// regression. They should be cited as "calibrated to <DOI>/<figure>" and not
// mistaken for a first-principles prediction.

// Park2015Fig2aHZO10nm returns a preset calibrated to:
// Park et al., Adv. Mater. (2015), doi:10.1002/adma.201404531, Fig. 2a (10 nm HZO).
func Park2015Fig2aHZO10nm() *HZOMaterial {
	m := DefaultHZO()
	m.Name = "HZO (Park 2015 Fig 2a, calibrated)"
	// Targets from digitized metrics used in validation suite.
	m.Pr = 15.8e-2 // 15.8 uC/cm2
	m.Ps = 19.4e-2 // 19.4 uC/cm2
	m.Ec = 1.0e8   // 1.0 MV/cm
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
	m.Pr = 100e-2        // 100 uC/cm2 (paper-reported remanent magnitude, Pt)
	m.Ps = 110e-2        // 110 uC/cm2 (calibrated saturation for loop-shape fit)
	m.Ec = 3.0e8         // 3.0 MV/cm
	m.Thickness = 200e-9 // 200 nm film in the paper
	return m
}

// Nanomaterials2024Fig2PZTThinFilm returns a preset calibrated to:
// Bi et al., Nanomaterials (2024), doi:10.3390/nano14050432, Fig. 2 (PZT thin-film P-E loop).
func Nanomaterials2024Fig2PZTThinFilm() *HZOMaterial {
	m := PZT()
	m.Name = "PZT (Nanomaterials 2024 Fig 2, calibrated)"
	m.Pr = 69.0e-2       // 69 uC/cm2 (paper-reported)
	m.Ps = 78.0e-2       // 78 uC/cm2 (calibrated saturation for loop fit)
	m.Ec = 1.148e8       // 1.148 MV/cm (11.48 kV/mm in paper)
	m.Thickness = 100e-9 // ~100 nm class film
	return m
}

// Crystals2021FigFerroelectricBTOTrilayer returns a provisional BTO preset calibrated to:
// Jaiswal et al., Crystals (2021), doi:10.3390/cryst11101192 (BTO/NFO/BTO ferroelectric hysteresis figure).
func Crystals2021FigFerroelectricBTOTrilayer() *HZOMaterial {
	m := BTO()
	m.Name = "BTO (Crystals 2021 hysteresis fig, calibrated)"
	m.Pr = 15.0e-2       // 15 uC/cm2 (provisional target)
	m.Ps = 20.0e-2       // 20 uC/cm2 (provisional saturation)
	m.Ec = 0.60e8        // 0.60 MV/cm (provisional target)
	m.Thickness = 100e-9 // nominal thin-film class for conversion consistency
	return m
}
