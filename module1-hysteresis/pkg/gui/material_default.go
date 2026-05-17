//go:build legacy_fyne

package gui

import "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"

func defaultMaterialSelection(materials []*ferroelectric.HZOMaterial) (*ferroelectric.HZOMaterial, int) {
	preferredNames := []string{
		"Literature Superlattice",
		"Literature Superlattice (Cheema 2020)",
	}

	for _, name := range preferredNames {
		for i, m := range materials {
			if m != nil && m.Name == name {
				return m, i
			}
		}
	}

	if len(materials) > 0 && materials[0] != nil {
		return materials[0], 0
	}

	return ferroelectric.DefaultHZO(), 0
}
