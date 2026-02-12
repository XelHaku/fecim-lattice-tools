package physics

import (
	"fmt"
	"math"
	"testing"

	configphysics "fecim-lattice-tools/config/physics"
)

func TestMaterialPresets_AllNineProduceValidHysteresis(t *testing.T) {
	cfg, err := configphysics.Load()
	if err != nil {
		t.Fatalf("failed to load physics config: %v", err)
	}

	loadFromConfig := func(name string) (*HZOMaterial, error) {
		raw := cfg.GetMaterial(name)
		if raw == nil {
			return nil, fmt.Errorf("material %q not found in config", name)
		}
		return MaterialFromConfig(raw, cfg), nil
	}

	presetLoaders := map[string]func() (*HZOMaterial, error){
		"fecim_hzo": func() (*HZOMaterial, error) {
			return loadFromConfig("fecim_hzo")
		},
		"fecim_hzo_target": func() (*HZOMaterial, error) {
			return loadFromConfig("fecim_hzo_target")
		},
		"default_hzo": func() (*HZOMaterial, error) {
			return loadFromConfig("default_hzo")
		},
		"literature_superlattice": func() (*HZOMaterial, error) {
			return loadFromConfig("literature_superlattice")
		},
		"cryogenic_hzo": func() (*HZOMaterial, error) {
			return loadFromConfig("cryogenic_hzo")
		},
		"hzo_standard_32": func() (*HZOMaterial, error) {
			return loadFromConfig("hzo_standard_32")
		},
		"hzo_ftj_140": func() (*HZOMaterial, error) {
			return loadFromConfig("hzo_ftj_140")
		},
		"hzo_custom_14": func() (*HZOMaterial, error) {
			return HZOCustom14(), nil
		},
		"alscn": func() (*HZOMaterial, error) {
			return loadFromConfig("alscn")
		},
	}

	presetOrder := []string{
		"fecim_hzo",
		"fecim_hzo_target",
		"default_hzo",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"hzo_custom_14",
		"alscn",
	}

	for _, preset := range presetOrder {
		loader := presetLoaders[preset]
		if loader == nil {
			t.Fatalf("missing loader for preset %q", preset)
		}

		t.Run(preset, func(t *testing.T) {
			mat, err := loader()
			if err != nil {
				t.Fatalf("material failed to load: %v", err)
			}
			if mat == nil {
				t.Fatal("material loader returned nil")
			}

			if mat.Pr <= 0 {
				t.Fatalf("expected Pr > 0, got %.6e", mat.Pr)
			}
			if mat.Ec <= 0 {
				t.Fatalf("expected Ec > 0, got %.6e", mat.Ec)
			}
			if mat.Thickness <= 0 {
				t.Fatalf("expected thickness > 0, got %.6e", mat.Thickness)
			}

			s := NewLKSolver()
			s.ConfigureFromMaterial(mat)
			s.EnableNoise = false
			s.UseNLS = false
			s.SetState(-math.Abs(mat.Pr))

			Emax := 3.0 * mat.Ec
			const nSteps = 160
			const dt = 1e-11
			const settleSteps = 4

			fields := make([]float64, 0, 2*(nSteps+1))
			polarization := make([]float64, 0, 2*(nSteps+1))

			runPoint := func(E float64) {
				var P float64
				for k := 0; k < settleSteps; k++ {
					P = s.Step(E, dt)
				}
				if math.IsNaN(P) || math.IsInf(P, 0) {
					t.Fatalf("non-finite polarization for E=%.6e: P=%.6e", E, P)
				}
				fields = append(fields, E)
				polarization = append(polarization, P)
			}

			// Full loop: -Emax -> +Emax -> -Emax
			for i := 0; i <= nSteps; i++ {
				E := -Emax + 2.0*Emax*float64(i)/float64(nSteps)
				runPoint(E)
			}
			for i := 0; i <= nSteps; i++ {
				E := Emax - 2.0*Emax*float64(i)/float64(nSteps)
				runPoint(E)
			}

			if len(fields) != len(polarization) || len(fields) < 2 {
				t.Fatalf("invalid loop data sizes: len(E)=%d len(P)=%d", len(fields), len(polarization))
			}

			// Area in E-P plane (energy density per cycle): integral P dE
			area := 0.0
			for i := 0; i < len(fields)-1; i++ {
				dE := fields[i+1] - fields[i]
				area += 0.5 * (polarization[i] + polarization[i+1]) * dE
			}
			area = math.Abs(area)

			if math.IsNaN(area) || math.IsInf(area, 0) {
				t.Fatalf("computed non-finite loop area: %.6e", area)
			}
			if area <= 0 {
				t.Fatalf("expected loop area > 0, got %.6e", area)
			}
		})
	}
}
