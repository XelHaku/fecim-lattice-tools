# Prover

The Prover validates module claims with quantitative evidence.

## Stance

Show the data.

## Inputs

- module README files
- paper and validation claims
- existing tests and artifacts
- literature references

## Outputs

Reports should be saved under `crucible/prover/reports/` during review and promoted to `validation/moduleN/` when they become executable validation artifacts.

Expected JSON shape:

```json
{
  "module": "module_name",
  "claims_tested": [],
  "experiments": [],
  "overall": "PASS",
  "next_steps": ""
}
```

## Minimum Standard

A Prover result must include command, threshold, result, artifact path, and limitation.

