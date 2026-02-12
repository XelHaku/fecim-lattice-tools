# G08 Evidence - Literature Superlattice MID Stability

Date: 2026-02-11 18:49 CST

Command:

```bash
./scripts/run_headless_ispp_regressions.sh
```

Artifacts produced and committed:

- `validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json`
- `validation/testdata/ispp_regression/lk_wrd_ispp_regression.json`

MID case values:

- Preisach MID (`target_level=15`):
  - `final_level=15`
  - `level_error=0`
  - `pulses=8`
  - `overshoots=6`
  - `retries=0`
  - `converged=true`
  - `reached_done=true`

- LK MID (`target_level=15`):
  - `final_level=24`
  - `level_error=9`
  - `pulses=30`
  - `overshoots=0`
  - `retries=1`
  - `converged=false`
  - `reached_done=true`

Interpretation used for G08:
- Preisach is suitable for strict target-lock MID acceptance.
- LK (single-domain baseline) is currently suitable for bounded-completion acceptance until LK stabilization/polydomain work is complete.
