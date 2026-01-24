// pkg/export/liberty.go
// Liberty timing file generator for FeCIM bitcell
//
// References:
// [1] Liberty Timing Format - Synopsys standard for timing libraries
//     https://people.eecs.berkeley.edu/~alanmi/publications/other/liberty07_03.pdf
// [2] SkyWater SKY130 PDK: https://skywater-pdk.readthedocs.io/
//
// ⚠️ CRITICAL DISCLAIMER: ALL TIMING VALUES ARE PLACEHOLDERS
// Real Liberty files require:
// 1. SPICE characterization with FeFET compact model
// 2. Timing arc extraction (setup, hold, propagation delays)
// 3. Power analysis across process corners
// 4. Temperature/voltage derating tables
package export

import (
	"fmt"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// GenerateLiberty generates a Liberty (.lib) timing file for the FeCIM bitcell
// This is required by synthesis and STA tools (OpenROAD, OpenLane)
// Format: Liberty Timing Format [Ref 1]
//
// ⚠️ WARNING: Generated timing values are PLACEHOLDERS requiring characterization
func GenerateLiberty(cfg config.CellConfig) string {
	area := cfg.Width * cfg.Height
	
	return fmt.Sprintf(`library(fecim_cells) {
  technology (cmos) ;
  delay_model : table_lookup ;
  
  time_unit : "1ns" ;
  voltage_unit : "1V" ;
  current_unit : "1mA" ;
  capacitive_load_unit (1, pf) ;
  leakage_power_unit : "1nW" ;
  
  operating_conditions(typical) {
    process : 1.0 ;
    temperature : 25 ;
    voltage : 1.8 ;
  }
  default_operating_conditions : typical ;
  
  cell(%s) {
    area : %.4f ;
    cell_leakage_power : %.4f ;
    
    pin(WL) {
      direction : input ;
      capacitance : %.4f ;
    }
    
    pin(BL) {
      direction : output ;
      function : "WL" ;
      
      timing() {
        related_pin : "WL" ;
        timing_sense : positive_unate ;
        
        cell_rise(scalar) {
          values("%.3f") ;
        }
        cell_fall(scalar) {
          values("%.3f") ;
        }
        rise_transition(scalar) {
          values("0.050") ;
        }
        fall_transition(scalar) {
          values("0.050") ;
        }
      }
    }
    
    pin(VPWR) {
      direction : inout ;
      pg_type : primary_power ;
    }
    
    pin(VGND) {
      direction : inout ;
      pg_type : primary_ground ;
    }
  }
}
`, cfg.Name, area, cfg.LeakagePower, cfg.InputCap, cfg.RiseTime, cfg.FallTime)
}
