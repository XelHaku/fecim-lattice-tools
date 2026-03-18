# Reproducibility Guide

How to create, share, and verify reproducibility packs for FeCIM simulations.

## What's in a Pack

| File | Purpose |
|------|---------|
| `config.yaml` | Full simulation configuration snapshot |
| `random_seeds.json` | All RNG seeds (including `lk_solver` key) |
| `git_commit.txt` | Source code version hash |
| `go_version.txt` | Go toolchain version (e.g., `go1.24.12`) |
| `config_version.txt` | Config schema version (e.g., `1.0.0`) |
| `DISCLAIMER.txt` | Simulation-only disclaimer (auto-generated) |
| `test_results.txt` | Test suite status at export time |
| `parameter_provenance.json` | Confidence ledger (included when attached) |
| `artifacts/` | Generated data files (CSV, JSON, plots) |

## Creating a Pack

### From Go Code

```go
import "fecim-lattice-tools/shared/export"

pack := export.ReproducibilityPackInput{
    ConfigYAML:    configBytes,           // raw YAML snapshot
    RandomSeeds:   map[string]int64{      // additional RNG seeds
        "global": 12345,
    },
    SolverSeed:    42,                    // written as "lk_solver" in random_seeds.json
    GitCommitHash: "a1b2c3d",
    GoVersion:     "",                    // auto-detected via runtime.Version() when empty
    ConfigVersion: "1.0.0",
    TestResults:   "ok  ./... 3.2s",
    GeneratedAssets: []string{            // files copied into artifacts/
        "output/hysteresis.csv",
        "output/pe_curve.png",
    },
    ParameterProvenance: ledger.ExportForReproPack(), // nil to omit
}

path, err := export.CreateReproducibilityPack("output/repro-pack", pack)
// Also supports zip: CreateReproducibilityPack("output/repro-pack.zip", pack)
```

### From CLI

```bash
fecim-lattice-tools --seed 42 --mode hysteresis
# Pack is auto-created in output/repro-pack/
```

## Reproducing Results

### Step 1: Verify Environment

Install the exact Go version from `go_version.txt` and clone the repo at the
commit recorded in `git_commit.txt`:

```bash
cat repro-pack/go_version.txt   # e.g., go1.24.12
cat repro-pack/git_commit.txt   # e.g., a1b2c3d
git checkout $(cat repro-pack/git_commit.txt)
```

### Step 2: Restore Configuration

Copy the config snapshot into the project and verify the schema version:

```bash
cp repro-pack/config.yaml config/constants.yaml
cat repro-pack/config_version.txt   # must match your config schema
```

### Step 3: Set Seed

Extract the solver seed and pass it to the simulation:

```bash
# Read the lk_solver seed from the pack
cat repro-pack/random_seeds.json
# {"lk_solver": 42, ...}

fecim-lattice-tools --seed 42 --mode hysteresis
```

### Step 4: Run and Compare

Run the simulation and compare output artifacts with those in the pack's
`artifacts/` directory. Bitwise-identical results are expected for deterministic
runs using the same Go version, commit, config, and seed.

## Validating a Pack

Use `ValidateReproducibilityPack` to check that all required files are present:

```go
err := export.ValidateReproducibilityPack("path/to/repro-pack")
if err != nil {
    log.Fatalf("invalid pack: %v", err)
}
```

Required files: `DISCLAIMER.txt`, `config.yaml`, `random_seeds.json`,
`git_commit.txt`, `go_version.txt`, `test_results.txt`, `artifacts/`.

Optional files: `config_version.txt`, `parameter_provenance.json`.

## Provenance Tiers

The confidence ledger (`parameter_provenance.json`) tags every physics parameter
with a provenance tier and confidence score:

| Tier | Confidence | Meaning |
|------|-----------|---------|
| Measured | 0.90--0.98 | Directly measured in published experiments |
| Calibrated | 0.75--0.86 | Fitted to experimental data via Landau models |
| Estimated | 0.60--0.72 | Derived from related measurements or theory |
| Placeholder | 0.20--0.40 | Educational default, not validated |

Example entries from the default ledger:

| Parameter | Tier | Confidence |
|-----------|------|-----------|
| `Pr` (remanent polarization) | Measured | 0.95 |
| `beta_landau` | Calibrated | 0.86 |
| `rho_viscosity` | Estimated | 0.72 |
| `imprint_field` | Placeholder | 0.20 |

Unregistered parameters receive a fallback confidence of 0.10 with
`placeholder` provenance.

## For Publication

Include your reproducibility pack as supplementary material. Reviewers can:

1. **Verify parameter provenance** via `parameter_provenance.json` -- every
   parameter's confidence tier is transparent.
2. **Check simulation-only claims** against `DISCLAIMER.txt` -- all outputs
   carry the notice: *"SIMULATION ONLY -- not validated against silicon."*
3. **Reproduce exact results** using the seed, config, and Go version recorded
   in the pack.

See also: `docs/4-research/honesty-audit.md` for the full accuracy policy.
