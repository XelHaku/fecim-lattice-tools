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
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// GenerateLEF generates a LEF (Library Exchange Format) file for the FeCIM bitcell
// LEF provides the abstract/physical view needed by place-and-route tools (OpenLane/OpenROAD)
// Format: LEF 5.8 [Ref 1]
func GenerateLEF(cfg config.CellConfig) string {
	return fmt.Sprintf(`VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

MACRO %s
  CLASS CORE ;
  ORIGIN 0 0 ;
  SIZE %.3f BY %.3f ;
  SYMMETRY X Y ;
  SITE unithd ;
  
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
`, cfg.Name, cfg.Width, cfg.Height, cfg.Name)
}
