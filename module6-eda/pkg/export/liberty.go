package export

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/shared/logging"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

var logLiberty = logging.NewLogger("eda-export-liberty")

const (
	publishedWriteRiseNS      = 50.0
	publishedReadFallNS       = 5.0
	publishedInputCapPF       = 0.015
	publishedLeakagePowerNW   = 0.0003
	publishedTransitionFactor = 0.30
)

type CharacterizationResult = sharedphysics.CharacterizationResult

// Module4EnergyModel carries per-operation energy from Module 4 (J/op).
// These terms are back-annotated into Liberty internal_power groups.
type Module4EnergyModel struct {
	DACEnergyJ float64
	MVMEnergyJ float64
	TIAEnergyJ float64
}

type libertyCorner struct {
	Name         string
	Process      float64
	TempC        float64
	TimingScale  float64
	VoltageScale float64
}

var defaultCorners = []libertyCorner{
	{Name: "ff_n40c", Process: 0.8, TempC: -40.0, TimingScale: 0.76, VoltageScale: 1.05},
	{Name: "ff_25c", Process: 0.8, TempC: 25.0, TimingScale: 0.82, VoltageScale: 1.02},
	{Name: "ff_125c", Process: 0.8, TempC: 125.0, TimingScale: 0.93, VoltageScale: 0.98},
	{Name: "tt_n40c", Process: 1.0, TempC: -40.0, TimingScale: 0.90, VoltageScale: 1.02},
	{Name: "tt_25c", Process: 1.0, TempC: 25.0, TimingScale: 1.00, VoltageScale: 1.00},
	{Name: "tt_125c", Process: 1.0, TempC: 125.0, TimingScale: 1.12, VoltageScale: 0.97},
	{Name: "ss_n40c", Process: 1.2, TempC: -40.0, TimingScale: 1.08, VoltageScale: 0.98},
	{Name: "ss_25c", Process: 1.2, TempC: 25.0, TimingScale: 1.20, VoltageScale: 0.95},
	{Name: "ss_125c", Process: 1.2, TempC: 125.0, TimingScale: 1.35, VoltageScale: 0.92},
}

func GenerateLibertyFromCharacterization(cfg config.CellConfig, char *CharacterizationResult) string {
	if char == nil {
		return GenerateLiberty(cfg)
	}
	cfgWithChar := cfg
	if char.WriteTimeNs > 0 {
		cfgWithChar.RiseTime = char.WriteTimeNs
	}
	if char.ReadTimeNs > 0 {
		cfgWithChar.FallTime = char.ReadTimeNs
	}
	if char.ReadEnergy_fJ > 0 {
		cfgWithChar.LeakagePower = char.ReadEnergy_fJ * 1e-6
	}
	return GenerateLiberty(cfgWithChar)
}

// GenerateLibertyWithModule4Energy back-annotates Module 4 energies into internal_power groups.
func GenerateLibertyWithModule4Energy(cfg config.CellConfig, energy *Module4EnergyModel) string {
	return injectModule4InternalPower(GenerateLiberty(cfg), cfg, energy)
}

// GenerateLibertyFromCharacterizationWithEnergy applies timing characterization + M4 power annotation.
func GenerateLibertyFromCharacterizationWithEnergy(cfg config.CellConfig, char *CharacterizationResult, energy *Module4EnergyModel) string {
	if char == nil {
		return GenerateLibertyWithModule4Energy(cfg, energy)
	}
	cfgWithChar := cfg
	if char.WriteTimeNs > 0 {
		cfgWithChar.RiseTime = char.WriteTimeNs
	}
	if char.ReadTimeNs > 0 {
		cfgWithChar.FallTime = char.ReadTimeNs
	}
	return injectModule4InternalPower(GenerateLiberty(cfgWithChar), cfgWithChar, energy)
}

func injectModule4InternalPower(lib string, cfg config.CellConfig, energy *Module4EnergyModel) string {
	if energy == nil {
		return lib
	}
	cfg = normalizeCellConfig(cfg)
	cycleTimeS := (cfg.RiseTime + cfg.FallTime) * 1e-9
	if cycleTimeS <= 0 {
		cycleTimeS = 1e-9
	}
	toNW := func(eJ float64) float64 {
		if eJ <= 0 {
			return 0
		}
		return (eJ / cycleTimeS) * 1e9
	}
	dacNW := toNW(energy.DACEnergyJ)
	mvmNW := toNW(energy.MVMEnergyJ)
	tiaNW := toNW(energy.TIAEnergyJ)

	var sb strings.Builder
	sb.WriteString("\n      internal_power() {\n")
	sb.WriteString("        related_pin : \"WL\" ;\n")
	sb.WriteString(fmt.Sprintf("        rise_power(scalar) { values(\"%.6f\") ; }\n", dacNW))
	sb.WriteString(fmt.Sprintf("        fall_power(scalar) { values(\"%.6f\") ; }\n", dacNW))
	sb.WriteString("      }\n")
	sb.WriteString("\n      internal_power() {\n")
	sb.WriteString("        related_pin : \"BL\" ;\n")
	sb.WriteString(fmt.Sprintf("        rise_power(scalar) { values(\"%.6f\") ; }\n", mvmNW))
	sb.WriteString(fmt.Sprintf("        fall_power(scalar) { values(\"%.6f\") ; }\n", mvmNW))
	sb.WriteString("      }\n")
	sb.WriteString("\n      internal_power() {\n")
	sb.WriteString("        related_pin : \"BL\" ;\n")
	sb.WriteString(fmt.Sprintf("        rise_power(scalar) { values(\"%.6f\") ; }\n", tiaNW))
	sb.WriteString(fmt.Sprintf("        fall_power(scalar) { values(\"%.6f\") ; }\n", tiaNW))
	sb.WriteString("      }\n")

	return strings.Replace(lib, "    pin(VPWR)", sb.String()+"\n    pin(VPWR)", 1)
}

// GenerateMultiCornerLiberty emits 9 corner libraries: FF/TT/SS at -40/25/125C.
func GenerateMultiCornerLiberty(cfg config.CellConfig) string {
	parts := make([]string, 0, len(defaultCorners))
	for _, c := range defaultCorners {
		ccfg := cfg
		ccfg.Process = c.Process
		ccfg.Temperature = c.TempC
		if ccfg.Voltage <= 0 {
			ccfg.Voltage = defaultVoltageForTechnology(cfg.Technology)
		}
		ccfg.Voltage *= c.VoltageScale
		parts = append(parts, generateCornerLiberty(ccfg, c))
	}
	return strings.Join(parts, "\n\n")
}

func GenerateLiberty(cfg config.CellConfig) string {
	logLiberty.Input("GenerateLiberty", map[string]interface{}{"cellName": cfg.Name, "cellType": cfg.CellType})
	return generateCornerLiberty(normalizeCellConfig(cfg), libertyCorner{Name: "typical", Process: cfg.Process, TempC: cfg.Temperature, TimingScale: 1.0, VoltageScale: 1.0})
}

func generateCornerLiberty(cfg config.CellConfig, corner libertyCorner) string {
	cfg = normalizeCellConfig(cfg)
	riseTime := cfg.RiseTime * corner.TimingScale
	fallTime := cfg.FallTime * corner.TimingScale
	riseTransition := riseTime * publishedTransitionFactor
	fallTransition := fallTime * publishedTransitionFactor

	libraryName := libraryNameForCellType(cfg.CellType, corner.Name)
	cellName, area, blFunction, inputPins := cellMetadata(cfg)

	return fmt.Sprintf(characterizationProvenanceBlockC+`library(%s) {
  technology (cmos) ;
  delay_model : table_lookup ;

  time_unit : "1ns" ;
  voltage_unit : "1V" ;
  current_unit : "1mA" ;
  capacitive_load_unit (1, pf) ;
  leakage_power_unit : "1nW" ;

  lu_table_template(fecim_nldm_7x7) {
    variable_1 : input_net_transition ;
    variable_2 : total_output_net_capacitance ;
    index_1("0.005, 0.010, 0.020, 0.040, 0.080, 0.120, 0.180") ;
    index_2("0.005, 0.010, 0.020, 0.040, 0.080, 0.120, 0.180") ;
  }

  operating_conditions(%s) {
    process : %.1f ;
    temperature : %.1f ;
    voltage : %.3f ;
  }
  default_operating_conditions : %s ;

  cell(%s) {
    area : %.4f ;
    cell_leakage_power : %.6f ;
%s
    pin(BL) {
      direction : output ;
      function : "%s" ;
%s
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
`, libraryName, corner.Name, corner.Process, corner.TempC, cfg.Voltage, corner.Name, cellName, area, cfg.LeakagePower, inputPins, blFunction,
		buildTimingBlocks(cfg.CellType, riseTime, fallTime, riseTransition, fallTransition))
}

func buildTimingBlocks(cellType string, riseTime, fallTime, riseTransition, fallTransition float64) string {
	relatedPins := []string{"WL"}
	if cellType == "1t1r" {
		relatedPins = []string{"WL", "SL"}
	} else if cellType == "2t1r" {
		relatedPins = []string{"WL", "CSL", "SL"}
	}
	blocks := make([]string, 0, len(relatedPins))
	for _, pin := range relatedPins {
		blocks = append(blocks, fmt.Sprintf(`
      timing() {
        related_pin : "%s" ;
        timing_sense : positive_unate ;

        cell_rise(fecim_nldm_7x7) {
%s
        }
        cell_fall(fecim_nldm_7x7) {
%s
        }
        rise_transition(fecim_nldm_7x7) {
%s
        }
        fall_transition(fecim_nldm_7x7) {
%s
        }
      }`, pin,
			formatNLDMValues(riseTime), formatNLDMValues(fallTime), formatNLDMValues(riseTransition), formatNLDMValues(fallTransition)))
	}
	return strings.Join(blocks, "\n")
}

func formatNLDMValues(base float64) string {
	slewScale := []float64{0.85, 0.92, 1.00, 1.10, 1.22, 1.35, 1.50}
	loadScale := []float64{0.82, 0.90, 1.00, 1.12, 1.25, 1.40, 1.56}
	rows := make([]string, 0, 7)
	for _, ss := range slewScale {
		vals := make([]string, 0, 7)
		for _, ls := range loadScale {
			vals = append(vals, fmt.Sprintf("%.3f", base*ss*ls))
		}
		rows = append(rows, fmt.Sprintf("          \"%s\"", strings.Join(vals, ", ")))
	}
	return fmt.Sprintf("          values(\\\n%s\\\n          ) ;", strings.Join(rows, ", \\\n"))
}

func normalizeCellConfig(cfg config.CellConfig) config.CellConfig {
	if cfg.Name == "" {
		cfg.Name = "fecim_bitcell"
	}
	if cfg.CellType == "" {
		cfg.CellType = "passive"
	}
	tech := sharedphysics.TechnologyNodeFromName(cfg.Technology)
	if cfg.Width <= 0 {
		cfg.Width = tech.CellPitchX * 1e6
	}
	if cfg.Height <= 0 {
		cfg.Height = tech.CellRowHeight * 1e6
	}
	if cfg.Voltage <= 0 {
		cfg.Voltage = tech.VDD
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 25.0
	}
	if cfg.Process <= 0 {
		cfg.Process = 1.0
	}
	if cfg.RiseTime <= 0 {
		cfg.RiseTime = publishedWriteRiseNS
	}
	if cfg.FallTime <= 0 {
		cfg.FallTime = publishedReadFallNS
	}
	if cfg.InputCap <= 0 {
		cfg.InputCap = publishedInputCapPF
	}
	if cfg.LeakagePower <= 0 {
		cfg.LeakagePower = publishedLeakagePowerNW
	}
	return cfg
}

func defaultVoltageForTechnology(tech string) float64 {
	return sharedphysics.TechnologyNodeFromName(tech).VDD
}

func libraryNameForCellType(cellType, corner string) string {
	prefix := "fecim_cells"
	switch cellType {
	case "1t1r":
		prefix = "fecim_1t1r_cells"
	case "2t1r":
		prefix = "fecim_2t1r_cells"
	}
	if corner == "typical" {
		return prefix
	}
	return fmt.Sprintf("%s_%s", prefix, corner)
}

func cellMetadata(cfg config.CellConfig) (cellName string, area float64, blFunction string, inputPins string) {
	width, height := cfg.Width, cfg.Height
	cellName = cfg.Name
	switch cfg.CellType {
	case "1t1r":
		if cellName == "fecim_bitcell" {
			cellName = "fecim_1t1r_bitcell"
		}
		if width < 0.9 {
			width = 0.920
		}
		if height < 3.0 {
			height = 3.400
		}
		blFunction = "(WL & SL)"
		inputPins = fmt.Sprintf(`    pin(WL) {
      direction : input ;
      capacitance : %.4f ;
    }

    pin(SL) {
      direction : input ;
      capacitance : %.4f ;
    }
`, cfg.InputCap, cfg.InputCap)
	case "2t1r":
		if cellName == "fecim_bitcell" {
			cellName = "fecim_2t1r_bitcell"
		}
		if width < 1.1 {
			width = 1.200
		}
		if height < 3.5 {
			height = 3.800
		}
		blFunction = "(WL & CSL & SL)"
		inputPins = fmt.Sprintf(`    pin(WL) {
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
`, cfg.InputCap, cfg.InputCap, cfg.InputCap)
	default:
		blFunction = "WL"
		inputPins = fmt.Sprintf(`    pin(WL) {
      direction : input ;
      capacitance : %.4f ;
    }
`, cfg.InputCap)
	}
	return cellName, width * height, blFunction, inputPins
}
