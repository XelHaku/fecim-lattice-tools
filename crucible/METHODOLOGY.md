# The Crucible Methodology

The Crucible is a three-agent validation protocol for FeCIM Lattice Tools. It exists to separate real validation from validation theater.

## Definition of Success

### Tier 1: Educational Success

A student can use this tool and emerge understanding FeCIM concepts they could not grasp from papers alone.

Requirements:

- visual learning
- interactive exploration
- progressive complexity
- honest limitation labels
- self-contained setup
- glossary and references
- visible failure modes such as sneak paths and IR drop
- comparison views for passive and transistor-isolated arrays when supported

Validation criteria:

- a student can complete a tutorial in about 30 minutes
- a student can explain why conductance levels matter
- a student can identify sneak paths visually
- a student can predict IR drop direction
- a student can distinguish educational simulation from production validation

### Tier 2: Research Success

A researcher can trust scoped outputs enough to inform experimental design or a hypothesis.

Requirements:

- mathematically correct physics within documented model scope
- conservation-law checks
- comparison to published data
- cross-tool validation when external tools are available
- documented assumptions
- reproducibility scripts
- configurable parameters
- exportable CSV and JSON results

Validation criteria:

- KCL residual below `1e-9 A`
- P-E loop comparison to Park 2015 with a documented RMSE threshold
- SPICE comparison with a documented current-deviation threshold
- deterministic seeded runs
- cited physics parameters
- reproducible inference metrics only when artifacts exist

### Tier 3: Scientific Success

The tool's claims are reproducible, validated, versioned, and able to withstand peer scrutiny.

Requirements:

- arXiv-style methods paper
- published validation methodology
- citation metadata and source records
- external users
- peer feedback incorporated
- versioned releases
- MIT license clarity
- public issue tracker activity

Long-term indicators:

- external users
- citations
- course adoption
- reproducible reports
- no unretracted major overclaims

## Citation Integration

The Crucible treats citations as part of validation, not as formatting.

- Prover must fail or mark partial any external claim that lacks a citation record or validation artifact.
- Disprover must check whether cited records actually support the claim being made.
- Builder must add source comments, citation records, or wording changes when a claim is missing support.
- Weak evidence such as preprints, conference material, or project assumptions must stay labeled as weak evidence.
- Contested claims belong in `citations/disputed.md` until resolved by stronger sources or project validation.

## Agent 1: Prover

Mission: verify scientific correctness through mathematical checks, empirical comparisons, internal consistency tests, and cross-tool validation.

Prompt template:

```markdown
You are a Scientific Validator for FeCIM Lattice Tools.

Your job: Prove module [X] works correctly.

Methodology:
1. Read the module's claims in README and docs.
2. Identify testable predictions.
3. Design experiments to verify each.
4. Run experiments programmatically.
5. Compare against published references when available.
6. Report PASS/FAIL/PARTIAL with quantitative evidence.

Rules:
- Every claim must be tested or marked untested.
- Every test must produce a number.
- Every number must have a tolerance.
- Every comparison needs a citation or artifact.
- "It compiles" is not validation.
- "It runs without error" is not validation.
- "Tests pass" is not enough without saying what they test.

Output JSON:
{
  "module": "module_name",
  "claims_tested": [],
  "experiments": [
    {
      "name": "",
      "method": "",
      "expected": "",
      "actual": "",
      "tolerance": "",
      "verdict": "PASS|FAIL|PARTIAL",
      "evidence_file": ""
    }
  ],
  "overall": "PASS|FAIL|PARTIAL",
  "next_steps": ""
}
```

### Prover Tasks by Module

Module 1, hysteresis:

- generate P-E loop with HZO parameters
- compare to Park 2015 digitized data
- compute RMSE
- verify symmetry of major loop where model assumptions require it
- verify minor loops stay inside the major loop
- verify saturation behavior at high field
- document parameter sources
- output `validation/module1/prover_report.json`

Module 2, crossbar:

- verify KCL for deterministic random arrays
- verify single-cell Ohm's law
- verify linear superposition for supported ideal paths
- verify IR drop direction along wires
- compare with ngspice when installed
- output `validation/module2/prover_report.json`

Module 3, MNIST:

- report FP32 baseline accuracy only from a reproducible run
- report quantized accuracy only from a reproducible run
- check seed determinism
- generate confusion matrix
- compare to published state-count results only when citation and setup match
- output `validation/module3/prover_report.json`

Module 4, peripherals:

- DAC INL/DNL sweep
- ADC quantization and SNR check
- TIA transfer sweep
- combined chain roundtrip
- output `validation/module4/prover_report.json`

Module 6, EDA:

- Yosys synthesis or syntax check
- DEF/LEF/Liberty parsing checks
- optional OpenLane smoke test
- generate-parse-regenerate roundtrip where supported
- output `validation/module6/prover_report.json`

## Agent 2: Disprover

Mission: find ways the tool is wrong, misleading, unstable, or overclaimed.

Prompt template:

```markdown
You are an Adversarial Auditor for FeCIM Lattice Tools.

Your job: Find ways this tool is wrong, misleading, or breaks.

Methodology:
1. Read the module documentation skeptically.
2. Identify implicit assumptions.
3. Test boundary conditions.
4. Test invalid inputs.
5. Test physical extremes.
6. Check for overstated claims.
7. Verify visualizations do not deceive.

Rules:
- Assume bugs exist.
- Test impossible inputs.
- Try to make the module fail.
- Check edge cases of physics and units.
- Verify error bars on claims.
- Identify untested assumptions.

Output JSON:
{
  "module": "module_name",
  "issues_found": [
    {
      "severity": "CRITICAL|HIGH|MEDIUM|LOW",
      "type": "physics|numerical|UI|documentation",
      "description": "",
      "reproduction": "",
      "expected": "",
      "actual": "",
      "impact": "",
      "suggested_fix": ""
    }
  ],
  "untested_claims": [],
  "documentation_issues": [],
  "overall_trust_level": "HIGH|MEDIUM|LOW"
}
```

### Disprover Attack Vectors

Module 1:

- fields beyond supported model range
- temperature extremes
- negative or zero material parameters
- degenerate Preisach grids
- minor-loop escape
- unit mismatches
- overclaims about conductance states, wake-up, or retention

Module 2:

- `1x1` arrays
- large arrays such as `64x64`, `128x128`, and stress sizes
- all-zero and all-max conductances
- negative voltages
- zero or extreme wire resistance
- passive array sneak-path behavior
- color-scale or visualization ambiguity

Module 3:

- zero epochs
- one-level quantization
- high-level quantization near FP32 behavior
- all-zero weights
- corrupted inputs
- seed sensitivity
- inconsistent accuracy claims

Module 4:

- DAC code overflow
- `V_ref = 0`
- ADC input saturation
- TIA compliance limits
- bandwidth assumptions
- noise determinism

Module 6:

- degenerate arrays
- large generated arrays
- overlapping DEF placements
- placeholder Liberty timing
- unsupported tape-out implications

## Agent 3: Builder

Mission: resolve Prover failures and Disprover findings in priority order.

Priority order:

1. Critical physics errors.
2. Critical documentation overclaims.
3. High-severity bugs.
4. Failed Prover gates.
5. Medium and low Disprover findings.
6. New features.

Prompt template:

```markdown
You are a Senior Engineer working on FeCIM Lattice Tools.

Input:
- Prover report
- Disprover report
- Current codebase

Your job: Resolve issues in priority order.

Methodology:
1. Pick the highest-priority issue.
2. Reproduce the failure.
3. Identify root cause.
4. Propose the fix.
5. Implement the fix.
6. Add regression test.
7. Verify Prover passes.
8. Verify Disprover finding is resolved.
9. Update documentation if behavior or trust boundary changed.
10. Commit with a descriptive message.

Output JSON:
{
  "issue_addressed": "",
  "root_cause": "",
  "fix_implemented": "",
  "tests_added": [],
  "files_modified": [],
  "regression_check": "PASS|FAIL",
  "documentation_updated": true,
  "commit_message": ""
}
```

## Continuous Validation Loop

Daily or per-feature:

1. Prover runs validation and emits reports.
2. Disprover attacks assumptions and emits findings.
3. Builder fixes the highest-priority item.
4. Prover reruns the relevant gates.
5. Findings are documented in changelog and trust docs.

## Weekly Metrics

Scientific rigor:

- percent of claims with citations
- percent of parameters with sources
- number of validation experiments
- RMSE values against published data
- KCL residual magnitude
- passing unit and validation tests

Coverage:

- percent of modules with Prover reports
- percent of Disprover findings resolved
- number of edge cases tested
- number of documentation fixes

External validation:

- stars, forks, issues, and contributors
- citations and external references
- course or tutorial usage

Progress:

- commits shipped
- bugs fixed
- features added
- papers cited

## Truth Filter

Red flags:

- "All tests pass" without saying what they test.
- "Validated" without artifact or reference.
- "Industry-standard" without source.
- No comparison to references.
- No error bars.
- No reproducibility script.
- No failure modes.

Green flags:

- RMSE vs a named paper and figure.
- KCL residual with threshold and sample count.
- Accuracy with seed, run count, and confidence interval.
- SPICE deviation with tool version.
- One-command reproducibility.
- Limitations visible next to claims.
- Citations and artifact paths.
- Failure modes documented.

## 90-Day Plan

Days 1-7:

- send first outreach email
- keep validation and paper scaffolds public
- define module claims explicitly

Days 8-21:

- finish Module 2 Prover report
- prepare Park 2015 comparison artifact
- generate SPICE comparison report if ngspice is available
- document each result

Days 22-35:

- run Disprover passes on all modules
- document edge cases
- document overclaims
- prioritize findings

Days 36-60:

- fix critical issues
- add regression tests
- update documentation
- rerun Prover gates

Days 61-90:

- write paper with validation data
- submit preprint only when claims are reproducible
- publish release notes and citation metadata
- engage external users through issues and tutorials

## Done Criteria

A module is done for a release when:

- Prover passes the module's declared gates.
- It has at least one reference comparison or a documented reason why that is not applicable.
- Disprover findings above the release threshold are resolved or explicitly accepted.
- Documentation is complete enough for a new user.
- Trust status is recorded in `docs/TRUST.md`.
