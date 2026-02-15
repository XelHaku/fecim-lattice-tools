#!/usr/bin/env bash
set -euo pipefail

# Literature-backed validation runner (Module 1 P-E loops)
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

unset DISPLAY
unset WAYLAND_DISPLAY

OUT_DIR="${FECIM_LITERATURE_JSON_DIR:-$ROOT/output/validation/literature}"
mkdir -p "$OUT_DIR"

echo "[lit] out_dir=$OUT_DIR"

# Run suite
FECIM_LITERATURE_JSON_DIR="$OUT_DIR" go test -v ./validation/literature/... | tee /tmp/literature_validation.out

# Summary table from JSON artifacts
python3 - <<'PY'
import glob, json, os
out = os.environ.get('FECIM_LITERATURE_JSON_DIR', 'output/validation/literature')
files = sorted(glob.glob(os.path.join(out, 'module1_pe_loop_*.json')))
print('\n[lit] Summary (module1 P-E loop literature validation)')
print('material_id\tPr_err_%\tEc_err_%\tRMSE_FS\tArea_err_%\tPASS')
for f in files:
    d=json.load(open(f))
    print(f"{d['material_id']}\t{d['pr_err_pct']:.2f}\t{d['ec_err_pct']:.2f}\t{d['rmse_full_scale']:.4f}\t{d['loop_area_err_pct']:.2f}\t{d['pass']}")
print(f"\n[lit] artifacts={len(files)} dir={out}")
PY
