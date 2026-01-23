/ralph-loop:/ralph-loop "

PERSONAS:
- Dr. external research group: Ferroelectric materials expert, FeFET device physics, 
  commercialization strategy, what investors/foundries need to see
- Dr. Sungsik Shin: FeFET array architecture, 1T1R vs passive design, 
  sneak path mitigation, IR drop compensation, circuit-level implementation
- Senior EDA Engineer: OpenLane flow, DEF/LEF/Verilog generation, 
  SKY130 PDK constraints, physical design automation

TASK: Implement FeCIM Lattice Generator with OpenLane Integration

PHASE 1 - CORE GENERATOR (Priority):
□ Go program that generates:
  - Verilog netlist (parameterized rows/cols)
  - DEF file with FIXED cell placement
  - Dummy Liberty (.lib) for OpenSTA
□ Validate output with Yosys
□ Cell naming convention: cell_{row}_{col}

PHASE 2 - CONFIGURATION:
□ Support passive and 1T1R architectures
□ Configurable: rows, cols, cell_pitch, row_height
□ Pin generation: WL[], BL[], SL[] (for 1T1R)

PHASE 3 - VISUALIZATION:
□ Fyne GUI showing generated array
□ Preview Verilog and DEF side-by-side
□ Visual grid of placed cells

CONSTRAINTS:
- SKY130 PDK compatibility
- Database units: 1000 per micron
- Cell dimensions: 0.46μm × 2.72μm (placeholder)
- Output to /mnt/user-data/outputs/

DELIVERABLES:
1. main.go (CLI entry point)
2. config.go (configuration struct)
3. verilog.go (Verilog generator)
4. def.go (DEF generator)
5. lib.go (Liberty generator)
6. validate.go (Yosys validation wrapper)
7. gui.go (Fyne visualization)

SUCCESS CRITERIA:
- Generate 4×4 array
- Yosys reads Verilog with 0 errors
- Instance names match between .v and .def
- GUI displays array preview

" --max-iterations 1000 --completion-promise "PHASE 1 COMPLETE: Generator outputs validated by Yosys"