// Package physics provides shared physics utilities for FeCIM simulations.
//
// This file implements Rule 6 of the FeCIM Shared Architecture Constitution:
// domain-specific constants must be loadable from YAML/JSON config files,
// not hardcoded as Go constants.
//
// Usage:
//
//	cfg, err := physics.LoadMaterialConfig("configs/materials/hfo2.yaml")
//	if err != nil { log.Fatal(err) }
//	mat := cfg.ToHZOMaterial()
package physics

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MaterialConfig is a YAML-serialisable representation of ferroelectric material
// parameters. It covers the most frequently varied parameters in research.
//
// Inspired by ferro_scripts' bto_params.yaml / crca_params.yaml pattern:
// adding a new material = creating a new YAML file, not editing Go source.
//
// Example file: configs/materials/hfo2.yaml
type MaterialConfig struct {
	// Identity
	Name        string `yaml:"name"`        // e.g. "HfO2", "BaTiO3"
	Description string `yaml:"description"` // free-text summary
	Source      string `yaml:"source"`      // literature citation

	// Polarization (µC/cm²) — converted to C/m² internally (* 1e-2)
	RemnantPolarization    float64 `yaml:"remnant_polarization_uc_cm2"`    // Pr
	SaturationPolarization float64 `yaml:"saturation_polarization_uc_cm2"` // Ps

	// Coercive field (kV/cm) — converted to V/m internally (* 1e5)
	CoerciveField float64 `yaml:"coercive_field_kv_cm"` // Ec

	// Dielectric
	DielectricConstant   float64 `yaml:"dielectric_constant"`    // ε_r (HF)
	DielectricConstantLF float64 `yaml:"dielectric_constant_lf"` // ε_r (LF), 0 = same as HF

	// Film geometry
	FilmThicknessNM float64 `yaml:"film_thickness_nm"` // in nm, converted to m (* 1e-9)

	// Landau-Khalatnikov coefficients
	Landau struct {
		Beta  float64 `yaml:"beta"`  // J·m⁵/C⁴
		Gamma float64 `yaml:"gamma"` // J·m⁹/C⁶ (0 = ignore cubic term)
		Rho   float64 `yaml:"rho"`   // viscosity/damping Ω·m
	} `yaml:"landau"`

	// Depolarization
	DepolarizationCoeff float64 `yaml:"depolarization_coeff"` // k_dep (V·m/C)

	// Dynamics (Merz / KAI model)
	Tau0NLS  float64 `yaml:"tau0_nls_s"`     // attempt time (s)
	EaNLS    float64 `yaml:"ea_nls_v_per_m"` // activation field (V/m)
	NLSSigma float64 `yaml:"nls_sigma"`      // log-normal sigma

	// Temperature
	CurieTempK float64 `yaml:"curie_temp_k"`  // Tc (K)
	CurieConst float64 `yaml:"curie_const_k"` // C0 (K)

	// Conductance window (for FeFET/RRAM models, Siemens)
	Gmin float64 `yaml:"g_min_s"` // off-state conductance
	Gmax float64 `yaml:"g_max_s"` // on-state conductance

	// Analog state count
	NumLevels int `yaml:"num_levels"` // discrete analog states (0 → use package default 30)

	// Symmetrize loop
	Symmetrize bool `yaml:"symmetrize"` // if true, force symmetric P-E loop
}

// LoadMaterialConfig reads a YAML file and returns a parsed MaterialConfig.
// Returns an error if the file cannot be read or the YAML is invalid.
func LoadMaterialConfig(path string) (*MaterialConfig, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("material config: read %q: %w", path, err)
	}
	var cfg MaterialConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("material config: parse %q: %w", path, err)
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("material config: %q: 'name' field is required", path)
	}
	return &cfg, nil
}

// ToHZOMaterial converts the YAML-loaded config into the canonical HZOMaterial
// struct used throughout the simulator. Unit conversions are applied here so
// YAML files use human-friendly units (kV/cm, µC/cm², nm).
func (c *MaterialConfig) ToHZOMaterial() *HZOMaterial {
	// Unit conversions
	Pr := c.RemnantPolarization * 1e-2    // µC/cm² → C/m²
	Ps := c.SaturationPolarization * 1e-2 // µC/cm² → C/m²
	Ec := c.CoerciveField * 1e5           // kV/cm  → V/m
	t := c.FilmThicknessNM * 1e-9         // nm     → m

	epsLF := c.DielectricConstantLF
	if epsLF == 0 {
		epsLF = c.DielectricConstant
	}

	gmin := c.Gmin
	gmax := c.Gmax
	if gmin == 0 && gmax == 0 {
		// Sensible defaults: 100:1 window around 1 µS
		gmin = 1e-8
		gmax = 1e-6
	}

	numLevels := c.NumLevels
	if numLevels == 0 {
		numLevels = DefaultLevels
	}

	return &HZOMaterial{
		Name:      c.Name,
		Pr:        Pr,
		Ps:        Ps,
		Ec:        Ec,
		Epsilon:   c.DielectricConstant,
		EpsilonLF: epsLF,
		Thickness: t,

		BetaLandau:   c.Landau.Beta,
		GammaLandau:  c.Landau.Gamma,
		RhoViscosity: c.Landau.Rho,

		K_dep: c.DepolarizationCoeff,

		Tau0NLS:  c.Tau0NLS,
		EaNLS:    c.EaNLS,
		NLSSigma: c.NLSSigma,

		CurieTemp:  c.CurieTempK,
		CurieConst: c.CurieConst,

		Gmin:      gmin,
		Gmax:      gmax,
		NumLevels: numLevels,
	}
}
