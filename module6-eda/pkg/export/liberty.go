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
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateLiberty generates a Liberty (.lib) timing file for the FeCIM bitcell
// This is required by synthesis and STA tools (OpenROAD, OpenLane)
// Format: Liberty Timing Format [Ref 1]
// Supports both passive and 1T1R architectures
//
// ⚠️ WARNING: Generated timing values are PLACEHOLDERS requiring characterization
func GenerateLiberty(cfg config.CellConfig) string {
	if cfg.CellType == "1t1r" {
		return Generate1T1RLiberty(cfg)
	}
	if cfg.CellType == "2t1r" {
		return Generate2T1RLiberty(cfg)
	}
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

// Generate1T1RLiberty generates Liberty file for 1T1R FeCIM bitcell with SL pin
// 1T1R has additional SL (Source Line) input for sneak path mitigation
func Generate1T1RLiberty(cfg config.CellConfig) string {
	cellName := cfg.Name
	if cellName == "fecim_bitcell" {
		cellName = "fecim_1t1r_bitcell"
	}
	// Use larger dimensions for 1T1R
	width := cfg.Width
	height := cfg.Height
	if width < 0.9 {
		width = 0.920
	}
	if height < 3.0 {
		height = 3.400
	}
	area := width * height

	return fmt.Sprintf(`library(fecim_1t1r_cells) {
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

    pin(SL) {
      direction : input ;
      capacitance : %.4f ;
    }

    pin(BL) {
      direction : output ;
      function : "(WL & SL)" ;

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
          values("0.060") ;
        }
        fall_transition(scalar) {
          values("0.060") ;
        }
      }

      timing() {
        related_pin : "SL" ;
        timing_sense : positive_unate ;

        cell_rise(scalar) {
          values("%.3f") ;
        }
        cell_fall(scalar) {
          values("%.3f") ;
        }
        rise_transition(scalar) {
          values("0.060") ;
        }
        fall_transition(scalar) {
          values("0.060") ;
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
`, cellName, area, cfg.LeakagePower, cfg.InputCap, cfg.InputCap, cfg.RiseTime, cfg.FallTime, cfg.RiseTime, cfg.FallTime)
}

// Generate2T1RLiberty generates Liberty file for 2T1R FeCIM bitcell with CSL pin
// 2T1R has WL (Word Line), CSL (Column Select Line), SL (Source Line), and BL (Bit Line)
// CSL enables individual cell addressing, differentiating it from 1T1R
func Generate2T1RLiberty(cfg config.CellConfig) string {
	cellName := cfg.Name
	if cellName == "fecim_bitcell" {
		cellName = "fecim_2t1r_bitcell"
	}
	// Use larger dimensions for 2T1R (more transistors than 1T1R)
	width := cfg.Width
	height := cfg.Height
	if width < 1.1 {
		width = 1.200
	}
	if height < 3.5 {
		height = 3.800
	}
	area := width * height

	return fmt.Sprintf(`library(fecim_2t1r_cells) {
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

    pin(CSL) {
      direction : input ;
      capacitance : %.4f ;
    }

    pin(SL) {
      direction : input ;
      capacitance : %.4f ;
    }

    pin(BL) {
      direction : output ;
      function : "(WL & CSL & SL)" ;

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
          values("0.070") ;
        }
        fall_transition(scalar) {
          values("0.070") ;
        }
      }

      timing() {
        related_pin : "CSL" ;
        timing_sense : positive_unate ;

        cell_rise(scalar) {
          values("%.3f") ;
        }
        cell_fall(scalar) {
          values("%.3f") ;
        }
        rise_transition(scalar) {
          values("0.070") ;
        }
        fall_transition(scalar) {
          values("0.070") ;
        }
      }

      timing() {
        related_pin : "SL" ;
        timing_sense : positive_unate ;

        cell_rise(scalar) {
          values("%.3f") ;
        }
        cell_fall(scalar) {
          values("%.3f") ;
        }
        rise_transition(scalar) {
          values("0.070") ;
        }
        fall_transition(scalar) {
          values("0.070") ;
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
`, cellName, area, cfg.LeakagePower, cfg.InputCap, cfg.InputCap, cfg.InputCap, cfg.RiseTime, cfg.FallTime, cfg.RiseTime, cfg.FallTime, cfg.RiseTime, cfg.FallTime)
}
