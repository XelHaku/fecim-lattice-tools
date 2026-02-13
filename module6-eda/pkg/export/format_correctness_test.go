package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

func makeSmallComputeDesign(t *testing.T) *compiler.ArrayDesign {
	t.Helper()
	weights := [][]float64{{0.2, -0.4, 0.8}, {-1.0, 0.0, 1.0}}
	cfg := compiler.NewComputeConfig(4, 4)
	cfg.WithWeights(weights)
	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}
	return design
}

func TestExportFormats_JSONCSVSPICEVerilogDEF(t *testing.T) {
	design := makeSmallComputeDesign(t)
	dir := t.TempDir()

	jsonPath := filepath.Join(dir, "a.json")
	csvPath := filepath.Join(dir, "a.csv")
	spPath := filepath.Join(dir, "a.sp")
	vPath := filepath.Join(dir, "a.v")
	defPath := filepath.Join(dir, "a.def")

	if err := ExportJSON(design, jsonPath); err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	if err := ExportCSV(design, csvPath); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	if err := ExportSPICE(design, spPath, 1.8); err != nil {
		t.Fatalf("ExportSPICE: %v", err)
	}
	if err := ExportVerilog(design, vPath); err != nil {
		t.Fatalf("ExportVerilog: %v", err)
	}
	if err := ExportDEF(design, defPath); err != nil {
		t.Fatalf("ExportDEF: %v", err)
	}

	// JSON format correctness
	jsonBytes, _ := os.ReadFile(jsonPath)
	var parsed compiler.ArrayDesign
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("invalid JSON export: %v", err)
	}
	if parsed.Stats.ActiveCells != 6 {
		t.Fatalf("json active_cells mismatch: got %d want 6", parsed.Stats.ActiveCells)
	}

	// CSV correctness + indexing
	csvFile, err := os.Open(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	defer csvFile.Close()
	records, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV export: %v", err)
	}
	if len(records) != 7 { // header + 6 mapped cells
		t.Fatalf("csv row count mismatch: got %d want 7", len(records))
	}
	if strings.Join(records[0], ",") != "row,col,weight,level,conductance_uS,resistance_ohm,program_V" {
		t.Fatalf("unexpected csv header: %v", records[0])
	}
	if records[1][0] != "0" || records[1][1] != "0" {
		t.Fatalf("csv first record index mismatch: %v", records[1])
	}
	if records[6][0] != "1" || records[6][1] != "2" {
		t.Fatalf("csv last record index mismatch: %v", records[6])
	}

	// SPICE correctness
	sp, _ := os.ReadFile(spPath)
	spice := string(sp)
	for _, token := range []string{".param VDD", "X_0_0", "X_1_2", ".end"} {
		if !strings.Contains(spice, token) {
			t.Fatalf("spice missing token: %s", token)
		}
	}

	// Verilog correctness
	vb, _ := os.ReadFile(vPath)
	verilog := string(vb)
	if !strings.Contains(verilog, "module fecim_crossbar") {
		t.Fatal("verilog missing module declaration")
	}
	if strings.Count(verilog, "R_") < 6 {
		t.Fatalf("verilog instance count too low: got %d", strings.Count(verilog, "R_"))
	}

	// DEF correctness
	db, _ := os.ReadFile(defPath)
	def := string(db)
	for _, token := range []string{"VERSION 5.8", "COMPONENTS 6", "- WL[0]", "- BL[2]", "END DESIGN"} {
		if !strings.Contains(def, token) {
			t.Fatalf("def missing token: %s", token)
		}
	}
}
