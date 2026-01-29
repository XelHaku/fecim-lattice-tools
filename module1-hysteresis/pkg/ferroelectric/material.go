// Package ferroelectric provides physics models for ferroelectric materials.
//
// This file re-exports the HZOMaterial type and factory functions from the shared
// physics package for backward compatibility. New code should import from
// "fecim-lattice-tools/shared/physics" directly.
package ferroelectric

import (
	"fecim-lattice-tools/config/physics"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// HZOMaterial is re-exported from shared/physics for backward compatibility.
// New code should import from "fecim-lattice-tools/shared/physics" directly.
type HZOMaterial = sharedphysics.HZOMaterial

// DefaultHZO returns material parameters for typical Si-doped HfO2 (Hf0.5Zr0.5O2).
// Re-exported from shared/physics for backward compatibility.
func DefaultHZO() *HZOMaterial {
	return sharedphysics.DefaultHZO()
}

// LiteratureSuperlattice returns parameters from academic literature for
// HfO2/ZrO2 superlattices.
// Re-exported from shared/physics for backward compatibility.
func LiteratureSuperlattice() *HZOMaterial {
	return sharedphysics.LiteratureSuperlattice()
}

// OptimizedHZO is deprecated. Use LiteratureSuperlattice() instead.
// Re-exported from shared/physics for backward compatibility.
func OptimizedHZO() *HZOMaterial {
	return sharedphysics.OptimizedHZO()
}

// FeCIMMaterial returns parameters matching FeCIM specifications.
// Re-exported from shared/physics for backward compatibility.
func FeCIMMaterial() *HZOMaterial {
	return sharedphysics.FeCIMMaterial()
}

// FeCIMMaterialTarget returns FeCIM TARGET specifications.
// Re-exported from shared/physics for backward compatibility.
func FeCIMMaterialTarget() *HZOMaterial {
	return sharedphysics.FeCIMMaterialTarget()
}

// CryogenicHZO returns HZO parameters at 4K for quantum computing integration.
// Re-exported from shared/physics for backward compatibility.
func CryogenicHZO() *HZOMaterial {
	return sharedphysics.CryogenicHZO()
}

// HZOStandard32 returns parameters for standard HZO demonstrating 32 analog states.
// Re-exported from shared/physics for backward compatibility.
func HZOStandard32() *HZOMaterial {
	return sharedphysics.HZOStandard32()
}

// HZOFJT140 returns parameters for HZO Ferroelectric Tunnel Junction with 140 states.
// Re-exported from shared/physics for backward compatibility.
func HZOFJT140() *HZOMaterial {
	return sharedphysics.HZOFJT140()
}

// HZOCustom14 returns a custom HZO configuration with 14 discrete levels.
// Re-exported from shared/physics for backward compatibility.
func HZOCustom14() *HZOMaterial {
	return sharedphysics.HZOCustom14()
}

// AlScN returns parameters for Aluminum Scandium Nitride.
// Re-exported from shared/physics for backward compatibility.
func AlScN() *HZOMaterial {
	return sharedphysics.AlScN()
}

// AllMaterials returns a slice of all available CMOS-compatible materials.
// Re-exported from shared/physics for backward compatibility.
func AllMaterials() []*HZOMaterial {
	return sharedphysics.AllMaterials()
}

// AllMaterialsFromConfig loads all materials from the physics config.
// Re-exported from shared/physics for backward compatibility.
func AllMaterialsFromConfig(cfg *physics.Config) []*HZOMaterial {
	return sharedphysics.AllMaterialsFromConfig(cfg)
}

// MaterialFromConfig converts a physics.Material to an HZOMaterial.
// Re-exported from shared/physics for backward compatibility.
func MaterialFromConfig(m *physics.Material, cfg *physics.Config) *HZOMaterial {
	return sharedphysics.MaterialFromConfig(m, cfg)
}
