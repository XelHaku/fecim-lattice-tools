package literature

// FORCReferenceData contains expected FORC characteristics from published
// literature for validation of the Preisach FORC implementation.
//
// The FORC (First-Order Reversal Curve) density rho(Hc, Hu) is the mixed second
// derivative of polarization with respect to reversal and measurement fields.
// In the coercivity-interaction coordinate system:
//
//	Hc = (Hb - Ha) / 2   (coercivity axis, Hc >= 0)
//	Hu = (Ha + Hb) / 2   (interaction / bias axis)
//
// A well-behaved ferroelectric FORC density should:
//   - Peak near (Hc = Ec, Hu = 0) for an unbiased material
//   - Be concentrated along the Ec axis (not spread uniformly)
//   - Show approximate symmetry about Hu = 0
type FORCReferenceData struct {
	// MaterialName is a human-readable identifier.
	MaterialName string

	// Source is the literature citation.
	Source string

	// DOI is the digital object identifier for the reference paper.
	DOI string

	// PeakHc_MVcm is the expected location of peak FORC density on the
	// coercivity axis (MV/cm).
	PeakHc_MVcm float64

	// PeakHcTolerance_pct is the acceptable relative error in peak Hc
	// location compared to the material Ec.
	PeakHcTolerance_pct float64

	// DensityFWHM_Hc_MVcm is the expected full-width-at-half-maximum of the
	// FORC density distribution along the Hc axis (MV/cm).
	DensityFWHM_Hc_MVcm float64

	// SymmetryRatio is the expected ratio of integrated density for Hu > 0
	// vs Hu < 0. A value of 1.0 means perfectly symmetric.
	SymmetryRatio float64

	// SymmetryTolerance is the acceptable deviation from perfect symmetry.
	// |ratio - 1.0| should be less than this value.
	SymmetryTolerance float64

	// ExpectedPr_uCcm2 is the expected remanent polarization (uC/cm2)
	// extractable from the FORC data (for cross-validation with P-E loops).
	ExpectedPr_uCcm2 float64

	// ExpectedEc_MVcm is the expected coercive field (MV/cm).
	ExpectedEc_MVcm float64

	// Notes contains any additional context about the reference data.
	Notes string
}

// HZO10nm_FORCReference returns expected FORC metrics for 10nm HZO thin films.
//
// Reference: Park et al., Adv. Mater. 27, 1811 (2015)
// "Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films"
//
// The FORC density for HZO is expected to show:
//   - Peak density concentrated near the coercive field (~0.93 MV/cm for 10nm films)
//   - Narrow distribution along Hc (FWHM ~ 0.3-0.5 MV/cm), reflecting the
//     relatively narrow switching-field distribution of polycrystalline HZO
//   - Good symmetry about Hu = 0 (no significant built-in bias)
//   - Pr ~ 15-16 uC/cm2 extractable from FORC integration
func HZO10nm_FORCReference() FORCReferenceData {
	return FORCReferenceData{
		MaterialName:        "HZO 10nm (Park 2015)",
		Source:              "Park et al., Adv. Mater. 27, 1811 (2015)",
		DOI:                 "10.1002/adma.201404531",
		PeakHc_MVcm:         0.93,
		PeakHcTolerance_pct: 20.0,
		DensityFWHM_Hc_MVcm: 0.4,
		SymmetryRatio:       1.0,
		SymmetryTolerance:   0.5,
		ExpectedPr_uCcm2:    15.8,
		ExpectedEc_MVcm:     0.93,
		Notes: "Calibrated to Park 2015 Fig 2a digitized loop. " +
			"FWHM and symmetry are qualitative expectations from " +
			"Preisach density theory for polycrystalline HZO. " +
			"Symmetry tolerance is widened to 0.5 because the FORC " +
			"density on a finite grid has inherent asymmetry from " +
			"the triangular Preisach domain boundary.",
	}
}

// BTO_FORCReference returns expected FORC metrics for BaTiO3 thin films.
//
// BTO is a classic ferroelectric with lower Ec (~30 kV/cm = 0.03 MV/cm) and
// higher Pr (~20 uC/cm2) compared to HZO. Its FORC density should show a
// very different distribution shape than HZO.
//
// Reference: Educational placeholder based on standard BTO thin-film parameters.
// See Li et al., J. Appl. Phys. 98, 064101 (2005) for LGD calibration data.
func BTO_FORCReference() FORCReferenceData {
	return FORCReferenceData{
		MaterialName:        "BTO (educational)",
		Source:              "Li et al., J. Appl. Phys. 98, 064101 (2005)",
		DOI:                 "10.1063/1.2042528",
		PeakHc_MVcm:         0.03,
		PeakHcTolerance_pct: 25.0,
		DensityFWHM_Hc_MVcm: 0.02,
		SymmetryRatio:       1.0,
		SymmetryTolerance:   0.3,
		ExpectedPr_uCcm2:    20.0,
		ExpectedEc_MVcm:     0.03,
		Notes: "BTO has much lower Ec than HZO, so peak Hc should be " +
			"correspondingly lower. Educational placeholder values.",
	}
}

// FORCDensityPeakLocation finds the peak of the Preisach density distribution
// in FORC result data. It returns the (Ea, Eb) coordinates of the maximum
// density value, or (0, 0) if the density grid is empty.
//
// To convert to coercivity-interaction coordinates:
//
//	Hc = (Eb - Ea) / 2
//	Hu = (Ea + Eb) / 2
func FORCDensityPeakLocation(density [][]float64, reversalFields []float64) (peakEa, peakEb float64) {
	if len(density) == 0 || len(reversalFields) == 0 {
		return 0, 0
	}

	maxRho := -1e300
	peakI, peakJ := 0, 0

	for i, row := range density {
		for j, rho := range row {
			if rho > maxRho {
				maxRho = rho
				peakI = i
				peakJ = j
			}
		}
	}

	n := len(reversalFields)
	if peakI < n && peakJ < n {
		peakEa = reversalFields[peakJ]
		peakEb = reversalFields[peakI]
	}
	return peakEa, peakEb
}
