package literature

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

type peLoopDataset struct {
	Name       string
	DOI        string
	SourceCSV  string
	MaterialID string
	Material   *sharedphysics.HZOMaterial
	Engine     string // preisach
	FieldUnit  string // MV/cm
	PolarUnit  string // uC/cm2
	Conditions map[string]any
	Notes      string
}

type peLoopMetrics struct {
	MaterialID string `json:"material_id"`
	Material   string `json:"material"`
	Engine     string `json:"engine"`
	DOI        string `json:"doi"`
	Dataset    string `json:"dataset"`
	Generated  string `json:"generated_at"`

	PrData_uC_cm2 float64 `json:"pr_data_uC_cm2"`
	PrSim_uC_cm2  float64 `json:"pr_sim_uC_cm2"`
	PrErrPct      float64 `json:"pr_err_pct"`

	EcData_MV_cm float64 `json:"ec_data_MV_cm"`
	EcSim_MV_cm  float64 `json:"ec_sim_MV_cm"`
	EcErrPct     float64 `json:"ec_err_pct"`

	RMSE_uC_cm2    float64 `json:"rmse_uC_cm2"`
	RMSEOverPs     float64 `json:"rmse_over_ps"` // RMSE(P(E)) / Ps (dimensionless)
	Ps_uC_cm2      float64 `json:"ps_uC_cm2"`
	LoopAreaData   float64 `json:"loop_area_data_J_m3"`
	LoopAreaSim    float64 `json:"loop_area_sim_J_m3"`
	LoopAreaErrPct float64 `json:"loop_area_err_pct"`

	Pass bool `json:"pass"`
}

// REQUIRED THRESHOLDS (Juan request)
const (
	thPrPct   = 10.0
	thEcPct   = 10.0
	thRMSEps  = 0.05 // RMSE(P(E)) / Ps <= 5%
	thAreaPct = 25.0 // loop area is a derived metric; keep pragmatic bound
)

func TestModule1_PELoop_LiteratureBacked(t *testing.T) {
	datasets := []peLoopDataset{
		{
			Name:       "Park2015_Fig2a_HZO_10nm",
			DOI:        "10.1002/adma.201404531",
			SourceCSV:  filepath.Join("data", "park2015_fig2a_hzo_10nm.csv"),
			MaterialID: "park2015_hzo_10nm",
			Material:   sharedphysics.Park2015Fig2aHZO10nm(),
			Engine:     "preisach",
			FieldUnit:  "MV/cm",
			PolarUnit:  "uC/cm2",
			Notes:      "Digitized representative loop from Park et al. 2015 Fig 2a. Used as literature-backed loop-shape validation.",
		},
		{
			Name:       "Cheema2020_Fig2c_HZO_Superlattice_5nm",
			DOI:        "10.1038/s41586-020-2208-x",
			SourceCSV:  filepath.Join("data", "cheema2020_fig2c_hzo_superlattice_5nm.csv"),
			MaterialID: "cheema2020_superlattice_5nm",
			Material:   sharedphysics.Cheema2020Fig2cHZOSuperlattice5nm(),
			Engine:     "preisach",
			FieldUnit:  "MV/cm",
			PolarUnit:  "uC/cm2",
			Notes:      "Digitized representative loop from Cheema et al. 2020 Fig 2c. Used as literature-backed loop-shape validation.",
		},
		{
			Name:       "MDPI2020_Fig3a_HZO_10nm_Wakeup",
			DOI:        "10.3390/ma13132968",
			SourceCSV:  filepath.Join("data", "mdpi2020_ma13132968_fig3a_hzo_10nm_wakeup.csv"),
			MaterialID: "mdpi2020_hzo_10nm_wakeup",
			Material:   sharedphysics.MDPI2020Fig3aHZO10nmWakeup(),
			Engine:     "preisach",
			FieldUnit:  "MV/cm",
			PolarUnit:  "uC/cm2",
			Notes:      "Representative loop from Kim et al. Materials 2020 Fig 3a (10nm HZO after wake-up). Used as Tier-1 literature validation dataset.",
		},
		{
			Name:       "Micromachines2022_Fig6a_AlScN_Pt_200nm",
			DOI:        "10.3390/mi13101629",
			SourceCSV:  filepath.Join("data", "alscn2022_pmc9607415_fig6a_pt_200nm.csv"),
			MaterialID: "alscn2022_pmc9607415_fig6a_pt_200nm",
			Material:   sharedphysics.Micromachines2022Fig6aAlScNPt200nm(),
			Engine:     "preisach",
			FieldUnit:  "MV/cm",
			PolarUnit:  "uC/cm2",
			Notes:      "Al0.7Sc0.3N 200nm on Pt bottom electrode (PMC9607415 Fig 6a). Current CSV is provisional and should be replaced by direct digitization for final Tier-1 acceptance.",
		},
	}

	outDir := os.Getenv("FECIM_LITERATURE_JSON_DIR")
	if outDir == "" {
		repoRoot, err := findRepoRoot()
		if err != nil {
			t.Fatalf("resolve repo root: %v", err)
		}
		outDir = filepath.Join(repoRoot, "output", "validation", "literature")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	for _, ds := range datasets {
		ds := ds
		t.Run(ds.MaterialID, func(t *testing.T) {
			Edata, Pdata, err := loadCSVLoop(ds.SourceCSV)
			if err != nil {
				t.Fatalf("load %s: %v", ds.SourceCSV, err)
			}
			if len(Edata) < 10 {
				t.Fatalf("dataset too small: %d points", len(Edata))
			}

			Esim, Psim := simulatePreisachLoop(ds.Material, Edata)
			_ = Esim

			prData := estimatePrAtZeroField(Pdata, Edata)
			prSim := estimatePrAtZeroField(Psim, Edata)

			ecData := estimateEc(Pdata, Edata)
			ecSim := estimateEc(Psim, Edata)

			rmse := rmse(Psim, Pdata)
			ps := ds.Material.Ps * 1e2 // C/m^2 -> uC/cm^2
			rmsePs := 0.0
			if ps > 0 {
				rmsePs = rmse / ps
			}

			areaData := loopAreaEnergyDensity(Edata, Pdata)
			areaSim := loopAreaEnergyDensity(Edata, Psim)

			prErrPct := pctErr(prSim, prData)
			ecErrPct := pctErr(ecSim, ecData)
			areaErrPct := pctErr(areaSim, areaData)

			pass := true
			if prErrPct > thPrPct {
				pass = false
				t.Errorf("Pr error too large: pr_sim=%0.3f pr_data=%0.3f err=%0.2f%% (th=%0.1f%%)", prSim, prData, prErrPct, thPrPct)
			}
			if ecErrPct > thEcPct {
				pass = false
				t.Errorf("Ec error too large: ec_sim=%0.3f ec_data=%0.3f err=%0.2f%% (th=%0.1f%%)", ecSim, ecData, ecErrPct, thEcPct)
			}
			if rmsePs > thRMSEps {
				pass = false
				t.Errorf("RMSE too large: rmse=%0.3f uC/cm2 Ps=%0.3f uC/cm2 rmse/Ps=%0.4f (th=%0.4f)", rmse, ps, rmsePs, thRMSEps)
			}
			if areaErrPct > thAreaPct {
				pass = false
				t.Errorf("loop area error too large: area_sim=%0.3e area_data=%0.3e err=%0.2f%% (th=%0.1f%%)", areaSim, areaData, areaErrPct, thAreaPct)
			}

			m := peLoopMetrics{
				MaterialID:     ds.MaterialID,
				Material:       ds.Material.Name,
				Engine:         ds.Engine,
				DOI:            ds.DOI,
				Dataset:        ds.Name,
				Generated:      time.Now().UTC().Format(time.RFC3339),
				PrData_uC_cm2:  prData,
				PrSim_uC_cm2:   prSim,
				PrErrPct:       prErrPct,
				EcData_MV_cm:   ecData,
				EcSim_MV_cm:    ecSim,
				EcErrPct:       ecErrPct,
				RMSE_uC_cm2:    rmse,
				RMSEOverPs:     rmsePs,
				Ps_uC_cm2:      ps,
				LoopAreaData:   areaData,
				LoopAreaSim:    areaSim,
				LoopAreaErrPct: areaErrPct,
				Pass:           pass,
			}

			outPath := filepath.Join(outDir, fmt.Sprintf("module1_pe_loop_%s.json", ds.MaterialID))
			if err := writeJSON(outPath, &m); err != nil {
				t.Fatalf("write %s: %v", outPath, err)
			}

			t.Logf("LITERATURE_METRICS material=%s doi=%s pr_sim=%0.3f pr_data=%0.3f err=%0.2f%%(th=%0.1f%%) ec_sim=%0.3f ec_data=%0.3f err=%0.2f%%(th=%0.1f%%) rmse=%0.3f uC/cm2 Ps=%0.3f uC/cm2 rmse/Ps=%0.4f(th=%0.4f) areaErr=%0.2f%%(th=%0.1f%%) pass=%v artifact=%s",
				ds.MaterialID, ds.DOI, prSim, prData, prErrPct, thPrPct, ecSim, ecData, ecErrPct, thEcPct, rmse, ps, rmsePs, thRMSEps, areaErrPct, thAreaPct, pass, outPath)
		})
	}
}

func loadCSVLoop(path string) ([]float64, []float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	recs, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(recs) < 2 {
		return nil, nil, errors.New("empty csv")
	}

	e := make([]float64, 0, len(recs)-1)
	p := make([]float64, 0, len(recs)-1)
	for i := 1; i < len(recs); i++ {
		mvcm, err := strconv.ParseFloat(recs[i][0], 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d parse E: %w", i+1, err)
		}
		uc, err := strconv.ParseFloat(recs[i][1], 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d parse P: %w", i+1, err)
		}
		e = append(e, mvcm)
		p = append(p, uc)
	}
	return e, p, nil
}

func simulatePreisachLoop(mat *sharedphysics.HZOMaterial, E_MV_cm []float64) ([]float64, []float64) {
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	Eout := make([]float64, len(E_MV_cm))
	Pout := make([]float64, len(E_MV_cm))
	for i, e := range E_MV_cm {
		E_vpm := e * 1e8
		P_c_m2 := model.Update(E_vpm)
		Eout[i] = e
		Pout[i] = P_c_m2 * 1e2 // C/m2 -> uC/cm2
	}
	return Eout, Pout
}

func estimatePrAtZeroField(P, E []float64) float64 {
	// Prefer exact E=0 if present; otherwise use the closest |E|.
	best := math.Inf(1)
	pr := 0.0
	for i := range E {
		d := math.Abs(E[i])
		if d < best {
			best = d
			pr = math.Abs(P[i])
		}
	}
	return pr
}

func estimateEc(P, E []float64) float64 {
	// Find sign changes in P and linearly interpolate E at P=0.
	cross := make([]float64, 0, 4)
	for i := 1; i < len(P); i++ {
		if P[i-1] == 0 {
			cross = append(cross, math.Abs(E[i-1]))
			continue
		}
		if (P[i-1] > 0) != (P[i] > 0) {
			// interpolate between i-1 and i
			t := math.Abs(P[i-1]) / (math.Abs(P[i-1]) + math.Abs(P[i]))
			e0 := E[i-1] + t*(E[i]-E[i-1])
			cross = append(cross, math.Abs(e0))
		}
	}
	if len(cross) == 0 {
		return 0
	}
	// Use mean of crossings.
	sum := 0.0
	for _, c := range cross {
		sum += c
	}
	return sum / float64(len(cross))
}

func rmse(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return math.Inf(1)
	}
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s / float64(len(a)))
}

func maxAbs(x []float64) float64 {
	m := 0.0
	for _, v := range x {
		if a := math.Abs(v); a > m {
			m = a
		}
	}
	return m
}

func loopAreaEnergyDensity(E_MV_cm, P_uC_cm2 []float64) float64 {
	// Approximate W = ∮ E dP. Units:
	// E: MV/cm -> V/m (×1e8)
	// P: uC/cm2 -> C/m2 (×1e-2)
	// So E*dP -> (V/m)*(C/m2)=J/m3.
	if len(E_MV_cm) < 2 {
		return 0
	}
	w := 0.0
	for i := 1; i < len(E_MV_cm); i++ {
		e1 := E_MV_cm[i-1] * 1e8
		e2 := E_MV_cm[i] * 1e8
		p1 := P_uC_cm2[i-1] * 1e-2
		p2 := P_uC_cm2[i] * 1e-2
		dp := p2 - p1
		eAvg := 0.5 * (e1 + e2)
		w += eAvg * dp
	}
	return math.Abs(w)
}

func pctErr(sim, ref float64) float64 {
	den := math.Abs(ref)
	if den == 0 {
		return math.Inf(1)
	}
	return math.Abs(sim-ref) / den * 100.0
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("go.mod not found from %s upward", wd)
		}
		cur = parent
	}
}

// park2015HZO10nmPreset removed: use sharedphysics.Park2015Fig2aHZO10nm()
