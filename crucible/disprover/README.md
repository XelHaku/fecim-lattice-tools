# Disprover

The Disprover searches for incorrect physics, numerical instability, misleading docs, and user-facing overclaims.

## Stance

How is this wrong?

## Inputs

- module documentation
- validation reports
- UI claims and screenshots
- paper draft
- public README claims

## Outputs

Reports should be saved under `crucible/disprover/reports/`.

Expected JSON shape:

```json
{
  "module": "module_name",
  "issues_found": [],
  "untested_claims": [],
  "documentation_issues": [],
  "overall_trust_level": "MEDIUM"
}
```

## Minimum Standard

Every finding must include severity, reproduction steps, expected behavior, actual behavior, impact, and suggested fix.

