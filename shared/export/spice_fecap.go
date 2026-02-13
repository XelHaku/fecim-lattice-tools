package export

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

const vacuumPermittivity = 8.8541878128e-12 // F/m

// FECapParams contains parameters for Landau-Khalatnikov ferroelectric capacitor SPICE export.
type FECapParams struct {
	Name    string  // e.g. "FECAP_HZO"
	Alpha   float64 // Landau alpha coefficient (SI)
	Beta    float64 // Landau beta coefficient (SI)
	Gamma   float64 // Landau gamma coefficient (SI)
	Rho     float64 // damping/viscosity (Ohm*m)
	EpsR    float64 // relative permittivity for linear dielectric branch
	Area_m2 float64 // device area
	Thick_m float64 // film thickness
}

// DefaultMaterlikFECapParams returns LK defaults anchored to Materlik 2015
// coefficients used in shared/physics/material.go.
func DefaultMaterlikFECapParams() FECapParams {
	m := physics.MaterlikHfO2()
	return FECapParams{
		Name:    "FECAP_HZO",
		Alpha:   -1.0e8, // practical ngspice default; full alpha(T) calibration needs full LK flow
		Beta:    m.BetaLandau,
		Gamma:   m.GammaLandau,
		Rho:     m.RhoViscosity,
		EpsR:    25.0,
		Area_m2: 2.025e-15,
		Thick_m: 10e-9,
	}
}

// GenerateFECapSubcircuit returns an ngspice-compatible Landau-Khalatnikov FeCap subcircuit.
func GenerateFECapSubcircuit(params FECapParams) string {
	if strings.TrimSpace(params.Name) == "" {
		params.Name = "FECAP_HZO"
	}
	if params.Alpha == 0 {
		params.Alpha = -1.0e8
	}
	if params.Beta == 0 || params.Gamma == 0 || params.Rho <= 0 || params.EpsR <= 0 || params.Area_m2 <= 0 || params.Thick_m <= 0 {
		d := DefaultMaterlikFECapParams()
		if params.Beta == 0 {
			params.Beta = d.Beta
		}
		if params.Gamma == 0 {
			params.Gamma = d.Gamma
		}
		if params.Rho <= 0 {
			params.Rho = d.Rho
		}
		if params.EpsR <= 0 {
			params.EpsR = d.EpsR
		}
		if params.Area_m2 <= 0 {
			params.Area_m2 = d.Area_m2
		}
		if params.Thick_m <= 0 {
			params.Thick_m = d.Thick_m
		}
	}

	return fmt.Sprintf(`.subckt %s pos neg PARAMS: alpha=%.6e beta=%.6e gamma=%.6e rho=%.6e thick=%.6e area=%.6e eps0=%.12e
* Landau-Khalatnikov ferroelectric capacitor
* Reference: Sivasubramanian & Widom, IEEE (2003)
* Materlik defaults used for beta/gamma (J. Appl. Phys. 117, 134109, 2015).
* Limitation: full nonlinear LK production models generally require Verilog-A.
* This ngspice B-source form is suitable for small-signal and transient studies.
Cfe pos n1 {eps0*%.6g*area/thick}
Rvisc n1 neg {rho*thick/area}
* Nonlinear Landau voltage via ngspice B-source
B_landau n1 neg V = -(2*alpha*v(n1)*thick + 4*beta*v(n1)**3*thick + 6*gamma*v(n1)**5*thick)
.ends %s
`, params.Name, params.Alpha, params.Beta, params.Gamma, params.Rho, params.Thick_m, params.Area_m2, vacuumPermittivity, params.EpsR, params.Name)
}

// Generate1T1RSubcircuit returns an ngspice 1T1R wrapper with one NMOS selector and one FeCap.
func Generate1T1RSubcircuit(fecap FECapParams, mosfetModel string) string {
	if strings.TrimSpace(fecap.Name) == "" {
		fecap.Name = "FECAP"
	}

	sel := physics.SKY130NMOS()
	modelName := strings.TrimSpace(mosfetModel)
	modelCard := ""
	if modelName == "" {
		modelName = "SKY130NMOS"
		modelCard = fmt.Sprintf(".model %s NMOS (LEVEL=1 VTO=%.6g KP=120e-6 LAMBDA=0.03)\n", modelName, sel.Vth)
	}

	return fmt.Sprintf(`%s
.subckt %s_1T1R bl wl sl
* 1T1R wrapper: NMOS selector in series with Landau-Khalatnikov FeCap
M_sel nmid wl sl sl %s W=%.12e L=%.12e
X_fecap bl nmid %s
.ends %s_1T1R
`, modelCard, fecap.Name, modelName, sel.W, sel.L, fecap.Name, fecap.Name)
}
