# G12 Evidence — GUI Parity Smoke Checklist

Date: 2026-02-11  
Repo: `fecim-lattice-tools`  
Module: `module1-hysteresis`

## Checklist

- [x] Capture GUI screenshot artifact (headless Xvfb)
- [x] Capture WRD/ISPP runtime log lines from same run
- [x] Verify target/phase widget wiring uses a single snapshot struct (`widgetSnapshot`)
- [x] Document expected WRD boundary log format and throttling behavior

## Commands run

```bash
xvfb-run -a ./fecim-screenshotter -only hysteresis \
  -out docs/development/evidence/g12-gui-parity-screenshots -w 1280 -h 800
```

## Screenshot artifacts

- `docs/development/evidence/g12-gui-parity-screenshots/hysteresis_initial.png`

## Log-line evidence (excerpt)

```text
[screenshotter] capturing hysteresis...
[screenshotter] hysteresis_initial: Run...
[screenshotter] onStarted hysteresis_initial
2026/02/11 19:08:58 ISPP START: target=21 fromSaturation=false Ec=8.5e+07 MaxField=2.12e+08 bounds=[0.000, 2.500]×Ec
2026/02/11 19:08:58 ISPP INIT: target=21, start=1.000×Ec (no calib), step=0.040×Ec
2026/02/11 19:08:58 ISPP APPLY: pulse=1 currentLevel=16 targetLevel=21 pulseDir=1 E=1.000×Ec bounds=[0.000, 2.500]×Ec
2026/02/11 19:08:58 ISPP WAIT: holding E=1.000×Ec for verify (pulse=1)
[screenshotter] capturing now hysteresis_initial
2026/02/11 19:08:58 ISPP READ: currentLevel=16 targetLevel=21 prevLevel=16 currentField=0.000×Ec
2026/02/11 19:08:58 ISPP VERIFY RESULT: error=-5 bounds=[1.000, 2.500]×Ec
```

## Notes

- This checklist focuses parity smoke validation (render + logs), not full functional correctness.
- WRD phase-boundary logging format/throttle contract is specified in:
  - `docs/development/GUI/WRD_PHASE_BOUNDARY_LOGGING_SPEC.md`
