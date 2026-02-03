// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

type termDetail struct {
	ID           string
	Title        string
	Equation     string
	Meaning      string
	Units        string
	DefaultValue string
	CodeRef      string
	References   []string
	Notes        string
}

type termDetailPanel struct {
	titleLabel   *widget.Label
	equation     *widget.Label
	meaning      *widget.Label
	units        *widget.Label
	defaultValue *widget.Label
	codeRef      *widget.Label
	references   *widget.Label
	notes        *widget.Label
}

func newTermDetailPanel() (*termDetailPanel, fyne.CanvasObject) {
	panel := &termDetailPanel{
		titleLabel:   widget.NewLabel("Tap a term to see its details."),
		equation:     widget.NewLabel(""),
		meaning:      widget.NewLabel(""),
		units:        widget.NewLabel(""),
		defaultValue: widget.NewLabel(""),
		codeRef:      widget.NewLabel(""),
		references:   widget.NewLabel(""),
		notes:        widget.NewLabel(""),
	}

	panel.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	panel.equation.TextStyle = fyne.TextStyle{Monospace: true}

	panel.equation.Wrapping = fyne.TextWrapWord
	panel.meaning.Wrapping = fyne.TextWrapWord
	panel.units.Wrapping = fyne.TextWrapWord
	panel.defaultValue.Wrapping = fyne.TextWrapWord
	panel.codeRef.Wrapping = fyne.TextWrapWord
	panel.references.Wrapping = fyne.TextWrapWord
	panel.notes.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		panel.titleLabel,
		fieldBlock("Equation", panel.equation),
		fieldBlock("Meaning", panel.meaning),
		fieldBlock("Units", panel.units),
		fieldBlock("Default / Typical", panel.defaultValue),
		fieldBlock("Code Mapping", panel.codeRef),
		fieldBlock("References", panel.references),
		fieldBlock("Notes", panel.notes),
	)

	card := widget.NewCard("Selected Term", "", content)
	return panel, card
}

func (p *termDetailPanel) SetDetail(termID, fallback string) {
	detail, ok := termDetails()[termID]
	if !ok {
		p.titleLabel.SetText("Selected Term")
		p.equation.SetText("")
		p.meaning.SetText(fallback)
		p.units.SetText("")
		p.defaultValue.SetText("")
		p.codeRef.SetText("")
		p.references.SetText("")
		p.notes.SetText("")
		return
	}

	p.titleLabel.SetText(detail.Title)
	p.equation.SetText(detail.Equation)
	p.meaning.SetText(detail.Meaning)
	p.units.SetText(detail.Units)
	p.defaultValue.SetText(detail.DefaultValue)
	p.codeRef.SetText(detail.CodeRef)
	p.references.SetText(bullets(detail.References))
	p.notes.SetText(detail.Notes)
}

func fieldBlock(title string, value *widget.Label) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(label, value)
}

func termDetails() map[string]termDetail {
	mat := physics.FeCIMMaterial()
	return map[string]termDetail{
		"rho_eff_main": {
			ID:           "rho_eff_main",
			Title:        "rho_eff (effective viscosity)",
			Equation:     "rho_eff dP/dt",
			Meaning:      "Effective damping that folds intrinsic viscosity and RC delay into a single term.",
			Units:        "Ohm*m (effective viscosity term in the L-K equation).",
			DefaultValue: fmt.Sprintf("rho_eff = rho + (R_series*A/d) with rho=%.2f Ohm*m, R_series=%.0f Ohm, A=%.0f nm^2, d=%.0f nm.", mat.RhoViscosity, mat.SeriesResistanceOhm, mat.Area*1e18, mat.Thickness*1e9),
			CodeRef:      "shared/physics/landau.go (UseEffectiveViscosity, effectiveRho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
				"RC delay aggregation (effective viscosity approximation)",
			},
		},
		"e_applied": {
			ID:           "e_applied",
			Title:        "E_applied (external field)",
			Equation:     "E_applied = V/d",
			Meaning:      "External drive field applied across the ferroelectric film thickness.",
			Units:        "V/m",
			DefaultValue: fmt.Sprintf("Ec approx %.2f MV/cm for FeCIM HZO defaults.", mat.Ec/1e8),
			CodeRef:      "shared/physics/landau.go (Step, dPdT)",
			References: []string{
				"Landau-Khalatnikov dynamics (drive term)",
			},
		},
		"k_dep": {
			ID:           "k_dep",
			Title:        "k_dep (depolarization factor)",
			Equation:     "E_dep = k_dep * P",
			Meaning:      "Depolarization field from interfacial layers; adds slant for analog multi-level states.",
			Units:        "V*m/C",
			DefaultValue: fmt.Sprintf("k_dep = %.2e V*m/C (Golden Set I).", mat.K_dep),
			CodeRef:      "shared/physics/landau.go (dPdT, E_depolarization)",
			References: []string{
				"Thin-film depolarization field models (dead-layer)",
			},
		},
		"alpha": {
			ID:           "alpha",
			Title:        "alpha (dynamic stiffness)",
			Equation:     "alpha(T,sigma) = (T-T_C)/(2epsilon0C) - 2 Q12 sigma",
			Meaning:      "Temperature + stress dependent curvature of the Landau free energy wells.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: fmt.Sprintf("Derived from T_C=%.0f K, C=%.1e K, Q12=%.3f m^4/C^2, sigma=%.1f GPa.", mat.CurieTemp, mat.CurieConst, mat.Q12, mat.StressGPa),
			CodeRef:      "shared/physics/landau.go (UpdateParams)",
			References: []string{
				"Curie-Weiss law (alpha temperature dependence)",
				"Electrostriction coupling (Q12 stress term)",
			},
		},
		"alpha_def": {
			ID:           "alpha_def",
			Title:        "alpha(T,sigma) definition",
			Equation:     "alpha(T,sigma) = (T - T_C)/(2 epsilon0 C) - 2 Q12 sigma",
			Meaning:      "Explicit temperature + stress form of the Landau stiffness coefficient.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: fmt.Sprintf("Derived from T_C=%.0f K, C=%.1e K, Q12=%.3f m^4/C^2, sigma=%.1f GPa.", mat.CurieTemp, mat.CurieConst, mat.Q12, mat.StressGPa),
			CodeRef:      "shared/physics/landau.go (UpdateParams)",
			References: []string{
				"Curie-Weiss law (alpha temperature dependence)",
				"Electrostriction coupling (Q12 stress term)",
			},
		},
		"beta": {
			ID:           "beta",
			Title:        "beta (first-order nonlinearity)",
			Equation:     "4betaP^3",
			Meaning:      "Negative beta creates the first-order switching barrier in HZO.",
			Units:        "beta in J*m^5/C^4 (contributes an effective field term).",
			DefaultValue: fmt.Sprintf("beta = %.3e J*m^5/C^4 (Golden Set I).", mat.BetaLandau),
			CodeRef:      "shared/physics/landau.go (dPdT, Beta)",
			References: []string{
				"Landau-Devonshire free energy expansion",
			},
		},
		"gamma": {
			ID:           "gamma",
			Title:        "gamma (sixth-order stabilizer)",
			Equation:     "6gammaP^5",
			Meaning:      "Positive gamma keeps the energy bounded at high polarization.",
			Units:        "gamma in J*m^9/C^6 (contributes an effective field term).",
			DefaultValue: fmt.Sprintf("gamma = %.3e J*m^9/C^6 (Golden Set I).", mat.GammaLandau),
			CodeRef:      "shared/physics/landau.go (dPdT, Gamma)",
			References: []string{
				"Landau-Devonshire free energy expansion",
			},
		},
		"lk_terms": {
			ID:           "lk_terms",
			Title:        "Landau-Khalatnikov nonlinearity",
			Equation:     "2alphaP + 4betaP^3 + 6gammaP^5",
			Meaning:      "Nonlinear energy gradient that shapes the ferroelectric double-well potential.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: "Computed each step from alpha(T,sigma), beta, and gamma.",
			CodeRef:      "shared/physics/landau.go (dPdT, UpdateParams)",
			References: []string{
				"Landau-Devonshire free energy expansion",
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"noise": {
			ID:           "noise",
			Title:        "xi(t) (stochastic noise)",
			Equation:     "xi(t)",
			Meaning:      "Optional Langevin noise for thermal variability and cycle-to-cycle spread.",
			Units:        "Effective field contribution (V/m).",
			DefaultValue: "Disabled by default (EnableNoise=false).",
			CodeRef:      "shared/physics/landau.go (noiseTerm)",
			References: []string{
				"Langevin dynamics for ferroelectric variability",
			},
		},
		"rho_eff_def": {
			ID:           "rho_eff_def",
			Title:        "rho_eff definition",
			Equation:     "rho_eff = rho + (R_series*A/d)",
			Meaning:      "Aggregates series resistance delay into effective viscosity.",
			Units:        "Ohm*m",
			DefaultValue: fmt.Sprintf("rho=%.2f Ohm*m, R_series=%.0f Ohm, A=%.0f nm^2, d=%.0f nm.", mat.RhoViscosity, mat.SeriesResistanceOhm, mat.Area*1e18, mat.Thickness*1e9),
			CodeRef:      "shared/physics/landau.go (effectiveRho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"rho": {
			ID:           "rho",
			Title:        "rho (viscosity / damping)",
			Equation:     "rho",
			Meaning:      "Intrinsic damping in polarization dynamics.",
			Units:        "Ohm*m",
			DefaultValue: fmt.Sprintf("rho = %.2f Ohm*m (Golden Set I).", mat.RhoViscosity),
			CodeRef:      "shared/physics/landau.go (Rho)",
			References: []string{
				"Khalatnikov viscous dynamics (L-K form)",
			},
		},
		"r_series": {
			ID:           "r_series",
			Title:        "R_series (series resistance)",
			Equation:     "R_series*A/d",
			Meaning:      "Series resistance term folded into effective viscosity.",
			Units:        "Ohm",
			DefaultValue: fmt.Sprintf("R_series = %.0f Ohm.", mat.SeriesResistanceOhm),
			CodeRef:      "shared/physics/landau.go (SeriesResistance)",
			References: []string{
				"RC delay aggregation (effective viscosity approximation)",
			},
		},
		"preisach_mu": {
			ID:           "preisach_mu",
			Title:        "mu(alpha,beta) (hysteron density)",
			Equation:     "mu(alpha,beta)",
			Meaning:      "Weighting function that defines how many hysterons sit at each (alpha,beta) threshold pair.",
			Units:        "Model density (normalized weight).",
			DefaultValue: "Implemented via the Everett function (tanh-based distribution in the default model).",
			CodeRef:      "module1-hysteresis/pkg/ferroelectric/preisach.go (TanhEverett.Calculate)",
			References: []string{
				"Preisach hysteresis density function",
			},
		},
		"preisach_gamma": {
			ID:           "preisach_gamma",
			Title:        "gamma_{alpha,beta}(E) (hysteron state)",
			Equation:     "gamma_{alpha,beta}(E)",
			Meaning:      "Bistable relay output for a single hysteron (+1/-1) with memory between thresholds.",
			Units:        "Unitless (+1 / -1).",
			DefaultValue: "Switches at alpha or beta, otherwise holds last state.",
			CodeRef:      "shared/physics/preisach.go (Update, ComputePolarization)",
			References: []string{
				"Preisach relay operator (memory element)",
			},
		},
		"preisach_alpha": {
			ID:           "preisach_alpha",
			Title:        "alpha (upper switching threshold)",
			Equation:     "alpha",
			Meaning:      "Upper threshold where a hysteron switches to +1 on an increasing field.",
			Units:        "V/m",
			DefaultValue: "Distributed across the Preisach plane; not a single scalar.",
			CodeRef:      "shared/physics/preisach.go (TurningPoint, Update)",
			References: []string{
				"Preisach plane (alpha, beta thresholds)",
			},
		},
		"preisach_beta": {
			ID:           "preisach_beta",
			Title:        "beta (lower switching threshold)",
			Equation:     "beta",
			Meaning:      "Lower threshold where a hysteron switches to -1 on a decreasing field.",
			Units:        "V/m",
			DefaultValue: "Distributed across the Preisach plane; not a single scalar.",
			CodeRef:      "shared/physics/preisach.go (TurningPoint, Update)",
			References: []string{
				"Preisach plane (alpha, beta thresholds)",
			},
		},
		"preisach_history": {
			ID:           "preisach_history",
			Title:        "History / turning points",
			Equation:     "Turning points stack",
			Meaning:      "Compressed input history that determines which hysterons are currently switched.",
			Units:        "N/A (state memory).",
			DefaultValue: "Managed by the wipe-out stack (turning points).",
			CodeRef:      "shared/physics/preisach.go (PreisachStack, TurningPoint)",
			References: []string{
				"Wipe-out property and Preisach memory",
			},
		},
	}
}

func buildLkInfoTabs() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Overview", scrollSection(buildOverviewSection())),
		container.NewTabItem("Model Notes", scrollSection(buildLkNotesSection())),
		container.NewTabItem("alpha(T,sigma)", scrollSection(buildAlphaSection())),
		container.NewTabItem("Parameters", scrollSection(buildGoldenSetSection())),
		container.NewTabItem("Materials", scrollSection(buildMaterialDefaultsSection())),
		container.NewTabItem("Dynamics", scrollSection(buildDynamicsSection())),
		container.NewTabItem("Assumptions", scrollSection(buildAssumptionsSection())),
		container.NewTabItem("References", scrollSection(buildReferencesSection())),
	)
	return tabs
}

func buildPreisachInfoTabs() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Overview", scrollSection(buildPreisachSection())),
		container.NewTabItem("Model Notes", scrollSection(buildPreisachNotesSection())),
		container.NewTabItem("alpha(T,sigma)", scrollSection(buildPreisachAlphaSection())),
		container.NewTabItem("Parameters", scrollSection(buildPreisachParametersSection())),
		container.NewTabItem("Materials", scrollSection(buildMaterialDefaultsSection())),
		container.NewTabItem("Dynamics", scrollSection(buildPreisachDynamicsSection())),
		container.NewTabItem("Assumptions", scrollSection(buildAssumptionsSection())),
		container.NewTabItem("References", scrollSection(buildReferencesSection())),
	)
	return tabs
}

func buildOverviewSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Master Equation (First-Order L-K / TDGL)"),
		bodyLabel("We solve polarization dynamics using the Landau-Khalatnikov form:"),
		equationBlock("rho dP/dt = -dG/dP"),
		bodyLabel("With effective field and Landau energy expansion:"),
		equationBlock("E_eff = E_applied - k_dep P"),
		equationBlock("dG/dP = 2alphaP + 4betaP^3 + 6gammaP^5"),
		bodyLabel("This widget adds depolarization and series-resistance aggregation used in the headless L-K path."),
	)
}

func buildLkNotesSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Model Notes"),
		bodyLabel("Landau-Khalatnikov captures dynamic switching with an explicit dP/dt term."),
		bodyLabel(bullets([]string{
			"Tap a coefficient or the LK nonlinearity row to see its meaning and code mapping.",
			"Includes depolarization and effective viscosity (series resistance aggregation).",
			"Use L-K for rate-dependent behavior; use Preisach for static loop shape.",
		})),
	)
}

func buildPreisachSection() fyne.CanvasObject {
	section := []fyne.CanvasObject{
		sectionTitle("Preisach Model (Quasi-Static)"),
	}

	if img := loadPreisachEquationSVG(); img != nil {
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(900, 140))
		section = append(section, img)
	}

	section = append(section,
		bodyLabel("The quasi-static Preisach model represents polarization as a weighted sum of bistable hysterons:"),
		equationBlock("P(E) = double_integral mu(alpha,beta) * gamma_{alpha,beta}(E) d alpha d beta"),
		bodyLabel("Each hysteron switches based on its thresholds and retains memory between them:"),
		equationBlock("gamma_{alpha,beta}(E) = +1 if E >= alpha; -1 if E <= beta; hold if beta < E < alpha"),
		bodyLabel("Quasi-static means rate-independent: there is no explicit dP/dt term and switching depends on input history."),
		bodyLabel("If you sweep faster or slower but preserve the same ordering of field values, the predicted loop is unchanged."),
		bodyLabel("Dynamics like viscosity, switching delay, and RC effects are intentionally omitted in this model."),
	)

	return container.NewVBox(section...)
}

func buildPreisachAlphaSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("alpha(T,sigma) in Preisach"),
		bodyLabel("The Preisach model here does not use the L-K alpha(T,sigma) stiffness term."),
		bodyLabel("Temperature and stress are handled by scaling the effective coercive field Ec and saturation polarization Ps."),
		bodyLabel("See the Parameters tab for how Ec and Ps are updated in the Everett distribution."),
	)
}

func buildPreisachParametersSection() fyne.CanvasObject {
	rows := [][]string{
		{"Parameter", "Meaning", "Notes"},
		{"Ec", "Effective coercive field", "Scaled with temperature and stress."},
		{"Ps", "Saturation polarization", "Scaled with temperature; used as Everett amplitude."},
		{"Delta", "Distribution width", "Set to 0.25 * Ec in TanhEverett."},
		{"E_sat", "Saturation field", "Set to 5 * Ec for the Preisach stack."},
	}
	return container.NewVBox(
		sectionTitle("Preisach Parameter Mapping"),
		bodyLabel("Parameters are mapped into the Everett function and Preisach stack:"),
		tableFromRows(rows, []float32{120, 240, 260}),
	)
}

func buildPreisachDynamicsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Quasi-Static Dynamics"),
		bodyLabel("Preisach uses a turning-point stack to encode history (wipe-out property)."),
		bodyLabel("There is no explicit time integration or dP/dt term, so results are rate-independent."),
		bodyLabel("To model switching delay or inertia, use the L-K dynamics tab."),
	)
}

func loadPreisachEquationSVG() *canvas.Image {
	const svgPath = "shared/assets/equations/preisach.svg"
	if _, err := os.Stat(svgPath); err != nil {
		return nil
	}
	return loadFrankesteinEquationSVG(svgPath)
}

func buildAlphaSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Unified alpha(T,sigma) Coefficient"),
		equationBlock("alpha(T,sigma) = (T - T_C) / (2 epsilon0 C) - 2 Q12 sigma"),
		bodyLabel("As temperature approaches T_C, alpha -> 0 and the wells flatten (more volatile)."),
		bodyLabel("Stress shifts alpha via electrostriction; tensile vs compressive sign depends on Q12."),
	)
}

func buildGoldenSetSection() fyne.CanvasObject {
	rows := [][]string{
		{"Parameter", "Value", "Units", "Role"},
		{"beta", "-2.160e8", "J*m^5/C^4", "First-order barrier (negative)"},
		{"gamma", "1.653e10", "J*m^9/C^6", "Stability (positive)"},
		{"rho", "0.05", "Ohm*m", "Viscosity / damping"},
		{"Q12", "-0.026", "m^4/C^2", "Electrostriction"},
		{"T_C", "723", "K", "Curie temperature"},
		{"k_dep", "2.5e8", "V*m/C", "Depolarization"},
	}
	return container.NewVBox(
		sectionTitle("Golden Parameter Set (10 nm HZO, Set I)"),
		tableFromRows(rows, []float32{120, 120, 140, 280}),
	)
}

func buildMaterialDefaultsSection() fyne.CanvasObject {
	mat := physics.FeCIMMaterial()
	coerciveV := mat.CoerciveVoltage()
	capacitance := mat.Capacitance()
	switchEnergy := mat.SwitchingEnergy()
	rows := [][]string{
		{"Parameter", "Value"},
		{"Pr (remanent polarization)", fmt.Sprintf("%.2f C/m^2 (%.0f uC/cm^2)", mat.Pr, mat.Pr*100)},
		{"Ps (saturation polarization)", fmt.Sprintf("%.2f C/m^2 (%.0f uC/cm^2)", mat.Ps, mat.Ps*100)},
		{"Ec (coercive field)", fmt.Sprintf("%.2f MV/cm", mat.Ec/1e8)},
		{"Vc (coercive voltage)", fmt.Sprintf("%.2f V (Ec * t)", coerciveV)},
		{"Thickness (t)", fmt.Sprintf("%.0f nm", mat.Thickness*1e9)},
		{"Area (A)", fmt.Sprintf("%.0f nm^2", mat.Area*1e18)},
		{"Capacitance (C)", fmt.Sprintf("%.2f fF", capacitance*1e15)},
		{"Tau (pulse width)", fmt.Sprintf("%.0f ns", mat.Tau*1e9)},
		{"Tau0 (attempt time)", fmt.Sprintf("%.0f ps", mat.Tau0*1e12)},
		{"Ea (activation energy)", fmt.Sprintf("%.2f eV", mat.Ea)},
		{"Alpha (KAI exponent)", fmt.Sprintf("%.2f", mat.Alpha)},
		{"Tau0NLS (Merz attempt)", fmt.Sprintf("%.0f ps", mat.Tau0NLS*1e12)},
		{"EaNLS (Merz field)", fmt.Sprintf("%.1f MV/cm", mat.EaNLS/1e8)},
		{"SwitchingEnergy", fmt.Sprintf("%.2f fJ", switchEnergy*1e15)},
		{"Epsilon (relative)", fmt.Sprintf("%.0f", mat.Epsilon)},
		{"EpsilonLF (low freq)", fmt.Sprintf("%.0f", mat.EpsilonLF)},
		{"LossAngle (tan delta)", fmt.Sprintf("%.3f", mat.LossAngle)},
		{"BetaLandau", fmt.Sprintf("%.3e J*m^5/C^4", mat.BetaLandau)},
		{"GammaLandau", fmt.Sprintf("%.3e J*m^9/C^6", mat.GammaLandau)},
		{"RhoViscosity", fmt.Sprintf("%.2f Ohm*m", mat.RhoViscosity)},
		{"SeriesResistance", fmt.Sprintf("%.0f Ohm", mat.SeriesResistanceOhm)},
		{"K_dep", fmt.Sprintf("%.2e V*m/C", mat.K_dep)},
		{"Q11", fmt.Sprintf("%.3f m^4/C^2", mat.Q11)},
		{"Q12", fmt.Sprintf("%.3f m^4/C^2", mat.Q12)},
		{"Stress", fmt.Sprintf("%.1f GPa", mat.StressGPa)},
		{"CurieTemp", fmt.Sprintf("%.0f K", mat.CurieTemp)},
		{"CurieConst", fmt.Sprintf("%.1e K", mat.CurieConst)},
		{"TempCoeffEc", fmt.Sprintf("%.2e V/m/K", mat.TempCoeffEc)},
		{"TempCoeffPr", fmt.Sprintf("%.2e C/m^2/K", mat.TempCoeffPr)},
		{"EnduranceCycles", fmt.Sprintf("%.1e cycles", mat.EnduranceCycles)},
		{"RetentionTime", fmt.Sprintf("%.1e s", mat.RetentionTime)},
		{"ImprintField", fmt.Sprintf("%.2e V/m", mat.ImrintField)},
		{"NumLevels", fmt.Sprintf("%d", mat.GetNumLevels())},
		{"Gmin/Gmax", fmt.Sprintf("%.0f / %.0f uS", mat.Gmin*1e6, mat.Gmax*1e6)},
	}

	return container.NewVBox(
		sectionTitle("FeCIM HZO Material Defaults"),
		bodyLabel("Defaults are pulled from shared/physics/material.go (FeCIMMaterial)."),
		tableFromRows(rows, []float32{180, 360}),
	)
}

func buildDynamicsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Numerical Solver & Write Loop"),
		bodyLabel("RK4 integration is used to stabilize the stiff L-K dynamics at sub-ns steps."),
		bodyLabel("Effective viscosity aggregates series resistance: rho_eff = rho + (R_series*A/d)."),
		bodyLabel("Headless and GUI use the same real incremental ISPP with overshoot handling."),
		bodyLabel("Optional NLS (Merz's law) modulates switching time at low fields."),
	)
}

func buildAssumptionsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Model Assumptions"),
		bodyLabel(bullets([]string{
			"Single-domain effective medium (no explicit multi-domain walls).",
			"Depolarization modeled by k_dep term (interfacial layer approximation).",
			"Series resistance folded into rho_eff for RC delay.",
			"Noise term optional; default disabled for deterministic checks.",
			"GUI can run either L-K dynamics or Preisach (toggle in controls); headless uses L-K.",
		})),
	)
}

func buildReferencesSection() fyne.CanvasObject {
	refs := []string{
		"Landau & Devonshire - phenomenological free-energy expansion for ferroelectrics (classic theory).",
		"Khalatnikov - viscous polarization dynamics (Landau-Khalatnikov equation, classic theory).",
		"Curie-Weiss law - temperature dependence of dielectric stiffness alpha(T).",
		"Electrostriction coupling - Q12 stress term in alpha(T,sigma).",
		"Merz's law - nucleation-limited switching kinetics (field-dependent tau).",
		"Park et al., Advanced Materials (2015) - HZO ferroelectricity in thin films.",
		"Muller et al., Nano Letters (2012) - ferroelectric HfO2 discovery and properties.",
		"Pesic et al., Advanced Functional Materials (2016) - wake-up and fatigue in HZO.",
		"Cheema et al., Nature 580 (2020) 478 - superlattice-enhanced ferroelectricity on silicon.",
	}
	return container.NewVBox(
		sectionTitle("Literature & Canonical References"),
		bodyLabel(bullets(refs)),
	)
}

func sectionTitle(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord
	return label
}

func bodyLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func equationBlock(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Monospace: true}
	label.Wrapping = fyne.TextWrapWord
	return label
}

func tableFromRows(rows [][]string, colWidths []float32) fyne.CanvasObject {
	rowCount := len(rows)
	colCount := 0
	if rowCount > 0 {
		colCount = len(rows[0])
	}
	table := widget.NewTable(
		func() (int, int) {
			return rowCount, colCount
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row >= rowCount || id.Col >= colCount {
				label.SetText("")
				return
			}
			label.SetText(rows[id.Row][id.Col])
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
		},
	)
	for col, width := range colWidths {
		if col < colCount {
			table.SetColumnWidth(col, width)
		}
	}
	return table
}

func scrollSection(content fyne.CanvasObject) fyne.CanvasObject {
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(560, 240))
	return scroll
}

func bullets(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- ")
		b.WriteString(item)
	}
	return b.String()
}
