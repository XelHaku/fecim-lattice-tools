# Research Workflow Tutorial

Step-by-step guide for researchers using FeCIM Lattice Tools.

> **Prerequisite.** Build with `go build -o fecim-lattice-tools
> ./cmd/fecim-lattice-tools`. See [README.md](README.md) for installation.

---

## Step 1: Load Material

List available presets:

```bash
./fecim-lattice-tools --list-materials
```

| Preset | Pr (uC/cm^2) | Ec (MV/cm) | Levels | Source |
|--------|-------------|------------|--------|--------|
| HZO (Park 2015) | 24.5 | 1.2 | 30 | Park et al. Adv. Mater. 2015 |
| FeCIM HZO | 30 | 1.0 | 30 | Estimated (conference) |
| Literature Superlattice | 22 | 0.85 | 64 | RSC Nanoscale 2025 |
| HZO Standard 32 | 20 | 1.0 | 32 | Oh et al. IEEE EDL 2017 |
| HZO FTJ 140 | 25 | 1.2 | 140 | Song et al. Adv. Sci. 2024 |
| AlScN | 120 | 5.0 | 12 | Nature Commun. 2025 |

Programmatic loading (Go):

```go
import "fecim-lattice-tools/shared/physics"

mat := physics.FeCIMMaterial()          // Built-in preset
mat := &physics.HZOMaterial{            // Custom parameters
    Name: "Custom HZO", Pr: 25e-2, Ps: 30e-2, Ec: 1.0e8,
    Thickness: 10e-9, Area: 100e-12, NumLevels: 30,
    BetaLandau: -6.720e8, GammaLandau: 1.950e10,
    RhoViscosity: 0.05, K_dep: 2.5e8,
    Gmin: 1e-6, Gmax: 100e-6,
}
```

---

## Step 2: Run P-E Sweep

### CLI

```bash
./fecim-lattice-tools --mode hysteresis --engine preisach \
    --material fecim_hzo --json -o pe_loop.json
./fecim-lattice-tools --mode hysteresis --engine lk \
    --material fecim_hzo --json -o pe_loop_lk.json
```

### Go -- Preisach Engine

```go
mat := physics.FeCIMMaterial()
everett := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.5}
stack := physics.NewPreisachStack(mat.Ec*3, everett)

Emax := mat.Ec * 2.5
for i := 0; i <= 200; i++ {
    E := -Emax + 2*Emax*float64(i)/200
    P := stack.Update(E)
    fmt.Printf("%.6e, %.6e\n", E, P)   // V/m, C/m^2
}
```

### Go -- LK Engine

```go
solver := physics.NewLKSolver()
solver.ConfigureFromMaterial(mat)
dt, freq, Emax := 1e-9, 1e6, mat.Ec*2.5
for i := 0; i < int(1.0/(freq*dt)); i++ {
    t := float64(i) * dt
    E := Emax * math.Sin(2*math.Pi*freq*t)
    P := solver.Step(E, dt)
    fmt.Printf("%.6e, %.6e, %.6e\n", t, E, P)
}
```

---

## Step 3: Validate Against Literature

Extract Pr (P at E = 0 on descending branch) and Ec (E at P = 0 on ascending
branch), then compare to published values:

```go
PrSim := stack.ComputePolarization(0)
relError := math.Abs(PrSim-mat.Pr) / mat.Pr * 100
fmt.Printf("Pr error: %.1f%%\n", relError)
```

Chi-squared test against digitized experimental data:

```go
chiSq := 0.0
for i := range E_exp {
    Psim := stack.ComputePolarization(E_exp[i])
    chiSq += math.Pow((Psim-P_exp[i])/sigma_exp, 2)
}
fmt.Printf("Reduced chi^2 = %.3f\n", chiSq/float64(len(E_exp)-nParams))
```

Temperature dependence check:

```go
for _, T := range []float64{77, 200, 300, 400, 600} {
    fmt.Printf("T=%5.0f K: Pr=%.2f uC/cm^2, Ec=%.2f MV/cm\n",
        T, mat.PolarizationAtTemp(T)*1e2, mat.CoerciveFieldAtTemp(T)/1e8)
}
```

---

## Step 4: Run Crossbar Simulation

```go
import "fecim-lattice-tools/shared/crossbar"

cfg := &crossbar.Config{Rows: 64, Cols: 64, NoiseLevel: 0.02, ADCBits: 4, DACBits: 4}
arr, _ := crossbar.NewArray(cfg)

// Program and execute MVM
weights := make([][]float64, 64)
for i := range weights { weights[i] = make([]float64, 64); /* fill */ }
arr.ProgramWeightMatrix(weights)

input := make([]float64, 64) // normalized [0,1]
output, _ := arr.MVM(input)
```

With non-idealities (IR drop, variation, drift):

```go
opts := crossbar.DefaultMVMOptions()
opts.IRDropEnabled = true
opts.VariationEnabled = true
result, _ := arr.MVMWithNonIdealities(input, opts)
```

---

## Step 5: Quantify Uncertainty

Monte Carlo for process variation confidence intervals:

```go
N := 1000
errors := make([]float64, N)
for trial := 0; trial < N; trial++ {
    cfg.NoiseLevel = 0.02 + rand.NormFloat64()*0.005
    arr, _ := crossbar.NewArray(cfg)
    arr.ProgramWeightMatrix(weights)
    output, _ := arr.MVM(input)
    // compute RMSE vs ideal
    errors[trial] = rmse(output, idealOutput)
}
sort.Float64s(errors)
fmt.Printf("RMSE 90%% CI: [%.4f, %.4f]\n",
    errors[int(0.05*float64(N))], errors[int(0.95*float64(N))])
```

Drift sensitivity:

```go
drift := crossbar.NewDriftSimulator(32, 32, 30)
for _, t := range []float64{1, 3600, 86400, 3.156e7} {
    drift.SimulateTimeStep(t - drift.Time)
    stats := drift.GetStats()
    fmt.Printf("t=%.0fs: level_errors=%d (%.2f%%)\n",
        t, stats.NumLevelErrors, stats.LevelErrorRate)
}
```

---

## Step 6: Export for Publication

CSV:

```go
f, _ := os.Create("pe_loop.csv")
fmt.Fprintln(f, "E_Vm,P_Cm2")
for i := range fieldData { fmt.Fprintf(f, "%.6e,%.6e\n", fieldData[i], polData[i]) }
f.Close()
```

JSON reproducibility pack:

```go
import "fecim-lattice-tools/shared/io"
io.WriteJSON("repro_pack.json", map[string]interface{}{
    "material": mat, "engine": "preisach",
    "timestamp": time.Now().UTC().Format(time.RFC3339),
    "data": map[string]interface{}{"E_Vm": fieldData, "P_Cm2": polData},
})
```

SPICE netlist export via EDA module:

```bash
./fecim-lattice-tools --module eda   # Use Export tab for .spice/.v/.def
```

---

## Step 7: Cross-Validate

### badcrossbar (Python, ground-truth nodal analysis)

```python
import badcrossbar, numpy as np, json
with open("crossbar_config.json") as f: data = json.load(f)
R = 1.0 / np.array(data["conductance_matrix"])
solution = badcrossbar.compute(np.array(data["input_voltages"]), R, r_i=2.5)
rel_error = np.abs(data["output_currents"] - solution.currents.output)
print(f"Max error: {np.max(rel_error):.2f}%")
```

### ngspice

```bash
ngspice -b crossbar_8x8.spice -o output.raw
```

### Preisach vs LK cross-check

```bash
./fecim-lattice-tools --mode hysteresis --engine preisach --material fecim_hzo --json -o p.json
./fecim-lattice-tools --mode hysteresis --engine lk --material fecim_hzo --json -o l.json
```

Both engines should agree on Pr (within 5%) and Ec (within 10%). Systematic
differences indicate Delta or K_dep need re-calibration.

---

## Publication Checklist

- [ ] Material parameters traced to DOI-backed source
- [ ] Approximations from [physics-models.md](../4-research/physics-models.md) Section 7 acknowledged
- [ ] Pr/Ec validated against at least one literature data point
- [ ] MVM accuracy verified against badcrossbar or ngspice
- [ ] Monte Carlo confidence intervals reported (>= 100 trials)
- [ ] Reproducibility pack (JSON) committed with manuscript
- [ ] All parameters in SI units
- [ ] Calibration-needed parameters flagged as educational defaults
