# Known Limitations

Honest documentation of current simulator limitations. See also: `docs/4-research/honesty-audit.md`.

## Physics Model Limitations

### Single-Domain LK Solver
- The default LK solver is single-domain; polydomain behavior requires ensemble mode
- Ensemble converges at N=100 domains (delta < 2% Pr) but adds computational cost
- No grain boundary effects or polycrystalline texture modeling

### Preisach Model
- Uses TanhEverett analytical density (no measured Preisach distribution input)
- Chi-squared fit is PASS (<5) for HZO, but EXPECTED_FAIL for PZT/BTO (uncalibrated)
- No temperature dependence in Preisach (use LK for thermal studies)

### Cross-Solver Agreement
- LK and Preisach agree on Pr within 7% but Ec differs by ~43%
- This is physically expected: LK has continuous switching, Preisach has sharp thresholds
- Use Preisach for fast parameter sweeps, LK for physically accurate dynamics

## Crossbar Array Limitations

### MVM Uncertainty Not Propagated
- `MVM()` applies device variation via conductance noise but returns only `[]float64`
- No uncertainty bounds on output values
- **Workaround**: Use Monte Carlo tests or `BuildResearchTrace()` for uncertainty analysis

### Thermal Model
- Independent RC cells with no lateral heat coupling
- Valid for pitches > 100 nm; for sub-100nm, use FEM (Elmer, 3D-ICE)
- Forward Euler integration (1st order); adequate when dt << tau

### IR Drop
- Linear resistive network model
- No frequency-dependent effects or capacitive coupling between lines

## Export Limitations

### No HDF5 Export
- All exports are JSON/CSV/SPICE text format
- HDF5 would enable larger dataset export and compatibility with Python/MATLAB workflows

### No Dublin Core Metadata
- Exports use custom `SimulationMetadata` struct, not standardized Dublin Core or DataCite

## Data Limitations

### Educational Defaults Only
- All parameters are simulation defaults unless explicitly cited with DOI
- No silicon measurement data included
- See confidence ledger for per-parameter provenance tiers
