# Quarterly External-Validation Review Checklist

Run this review once per quarter to keep external validation trustworthy and reproducible.

## 1) Security (dependency risk)

- [ ] Review CVEs for pinned external tools and runtimes:
  - ngspice, iverilog, verilator, openroad/openlane, python stack, go toolchain
- [ ] Check package manager advisories (OS distro + language ecosystems)
- [ ] Record affected versions and exposure in this repo
- [ ] Open remediation issues for high/critical CVEs

## 2) Model quality (scientific validity)

- [ ] Re-run physics regression suite against current main
- [ ] Re-check key metrics against latest literature updates
- [ ] Confirm confidence-band outcomes (green/yellow/red) still hold for tracked baselines
- [ ] Document any drift in Pr, Ec, switching time, read margin, energy/write

## 3) Compatibility (toolchain reproducibility)

- [ ] Run `scripts/toolchain/check_tools.sh`
- [ ] Run `scripts/toolchain/drift_detector.sh`
- [ ] Confirm installed versions still match `tools/external/README.md` pins
- [ ] If versions changed, decide whether to:
  - [ ] keep old baseline and mark historical
  - [ ] refresh baseline for new version
  - [ ] support both versions temporarily

## 4) Baseline hygiene

- [ ] Ensure `validation/external/baselines/` has immutable historical artifacts
- [ ] Add new baseline folders for accepted version updates
- [ ] Update baseline README changelog and rationale

## 5) Action items template

Use this template in the quarterly review note:

- **Quarter:** YYYY-QN
- **Reviewer(s):**
- **Security findings:**
- **Compatibility findings:**
- **Model-quality findings:**
- **Baseline updates required:**
- **Decision summary:** keep pins / update pins / dual support
- **Action items:**
  1. owner — due date — task
  2. owner — due date — task
