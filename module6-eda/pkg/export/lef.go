// pkg/export/lef.go
// LEF (Library Exchange Format) generator for FeCIM bitcell
//
// References:
// [1] LEF/DEF 5.8 Specification - Si2/OpenAccess Coalition
// [2] SkyWater SKY130 PDK: https://skywater-pdk.readthedocs.io/
//
// ⚠️ DISCLAIMER: This generates an ABSTRACT VIEW only.
// Actual Magic layout (.mag file) required for real GDS generation and DRC/LVS.
package export

import (
	"fmt"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateLEF generates a LEF (Library Exchange Format) file for the FeCIM bitcell
// LEF provides the abstract/physical view needed by place-and-route tools (OpenLane/OpenROAD)
// Format: LEF 5.8 [Ref 1]
// Supports passive, 1T1R, and 2T1R architectures
// Includes minimal layer and site definitions for standalone validation
func GenerateLEF(cfg config.CellConfig) string {
	if cfg.CellType == "1t1r" {
		return Generate1T1RLEF(cfg)
	}
	if cfg.CellType == "2t1r" {
		return Generate2T1RLEF(cfg)
	}
	return fmt.Sprintf(`VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

# Minimal layer definition for standalone validation
LAYER met1
  TYPE ROUTING ;
  DIRECTION HORIZONTAL ;
  PITCH 0.46 ;
  WIDTH 0.14 ;
END met1

# Site definition for placement
SITE fecim_site
  CLASS CORE ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
END fecim_site

MACRO %s
  CLASS CORE ;
  ORIGIN 0 0 ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
  SITE fecim_site ;

  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.000 1.200 0.100 1.400 ;
    END
  END WL

  PIN BL
    DIRECTION OUTPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.360 1.200 0.460 1.400 ;
    END
  END BL

  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
      RECT 0.000 2.620 0.460 2.720 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ;
    USE GROUND ;
    PORT
      LAYER met1 ;
      RECT 0.000 0.000 0.460 0.100 ;
    END
  END VGND

  OBS
    LAYER met1 ;
    RECT 0.100 0.100 0.360 2.620 ;
  END

END %s

END LIBRARY
`, cfg.Width, cfg.Height, cfg.Name, cfg.Width, cfg.Height, cfg.Name)
}

// Generate1T1RLEF generates LEF for 1T1R FeCIM bitcell with SL (Source Line) pin
// 1T1R cells are larger (0.92µm pitch) due to select transistor overhead
// Includes minimal layer and site definitions for standalone validation
func Generate1T1RLEF(cfg config.CellConfig) string {
	cellName := cfg.Name
	if cellName == "fecim_bitcell" {
		cellName = "fecim_1t1r_bitcell"
	}
	// Use larger dimensions for 1T1R (transistor overhead)
	width := cfg.Width
	height := cfg.Height
	if width < 0.9 {
		width = 0.920 // 1T1R minimum pitch
	}
	if height < 3.0 {
		height = 3.400 // Taller for transistor + FeFET stack
	}

	return fmt.Sprintf(`VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

# Minimal layer definition for standalone validation
LAYER met1
  TYPE ROUTING ;
  DIRECTION HORIZONTAL ;
  PITCH 0.46 ;
  WIDTH 0.14 ;
END met1

# Site definition for 1T1R placement
SITE fecim_1t1r_site
  CLASS CORE ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
END fecim_1t1r_site

MACRO %s
  CLASS CORE ;
  ORIGIN 0 0 ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
  SITE fecim_1t1r_site ;

  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.000 1.600 0.100 1.800 ;
    END
  END WL

  PIN BL
    DIRECTION OUTPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.820 1.600 0.920 1.800 ;
    END
  END BL

  PIN SL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.410 0.100 0.510 0.300 ;
    END
  END SL

  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
      RECT 0.000 3.300 0.920 3.400 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ;
    USE GROUND ;
    PORT
      LAYER met1 ;
      RECT 0.000 0.000 0.920 0.100 ;
    END
  END VGND

  OBS
    LAYER met1 ;
    RECT 0.100 0.300 0.820 3.300 ;
  END

END %s

END LIBRARY
`, width, height, cellName, width, height, cellName)
}

// Generate2T1RLEF generates LEF for 2T1R FeCIM bitcell with CSL (Column Select Line) pin
// 2T1R cells are larger (1.38µm pitch) due to dual select transistor overhead
// Includes minimal layer and site definitions for standalone validation
func Generate2T1RLEF(cfg config.CellConfig) string {
	cellName := cfg.Name
	if cellName == "fecim_bitcell" {
		cellName = "fecim_2t1r_bitcell"
	}
	// Use larger dimensions for 2T1R (dual transistor overhead)
	width := cfg.Width
	height := cfg.Height
	if width < 1.3 {
		width = 1.380 // 2T1R minimum pitch (~3x passive)
	}
	if height < 3.0 {
		height = 3.400 // Taller for dual transistor + FeFET stack
	}

	return fmt.Sprintf(`VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

# Minimal layer definition for standalone validation
LAYER met1
  TYPE ROUTING ;
  DIRECTION HORIZONTAL ;
  PITCH 0.46 ;
  WIDTH 0.14 ;
END met1

# Site definition for 2T1R placement
SITE fecim_2t1r_site
  CLASS CORE ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
END fecim_2t1r_site

MACRO %s
  CLASS CORE ;
  ORIGIN 0 0 ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
  SITE fecim_2t1r_site ;

  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.000 1.600 0.100 1.800 ;
    END
  END WL

  PIN CSL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 1.280 1.600 1.380 1.800 ;
    END
  END CSL

  PIN BL
    DIRECTION OUTPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.640 3.300 0.740 3.400 ;
    END
  END BL

  PIN SL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
      RECT 0.640 0.000 0.740 0.100 ;
    END
  END SL

  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
      RECT 0.000 3.300 1.380 3.400 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ;
    USE GROUND ;
    PORT
      LAYER met1 ;
      RECT 0.000 0.000 1.380 0.100 ;
    END
  END VGND

  OBS
    LAYER met1 ;
    RECT 0.100 0.100 1.280 3.300 ;
  END

END %s

END LIBRARY
`, width, height, cellName, width, height, cellName)
}
