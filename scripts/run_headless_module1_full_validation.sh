#!/usr/bin/env bash
set -euo pipefail

# Module 1 full headless validation lane (research-grade)
# Runs the key Tier-1 gates for Module 1:
# - Build
# - Unit tests
# - ISPP regression artifacts (Preisach + LK)
# - GUI/headless parity (WRD/ISPP)
# - Timestep convergence (RK4)

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

unset DISPLAY
unset WAYLAND_DISPLAY

OUT_DIR="${FECIM_REGRESSION_JSON_DIR:-$ROOT/output/regression/module1}"
mkdir -p "$OUT_DIR"

printf "[m1-full] repo=%s\n" "$ROOT"
printf "[m1-full] out_dir=%s\n" "$OUT_DIR"

printf "[m1-full] go build ./...\n"
go build ./...

printf "[m1-full] go test ./module1-hysteresis/...\n"
go test ./module1-hysteresis/...

printf "[m1-full] ISPP regressions (Preisach) -> JSON\n"
FECIM_REGRESSION_JSON_DIR="$OUT_DIR" go test ./module1-hysteresis/pkg/controller -run TestHeadlessRegression_WRD_ISPP_Preisach -count=1

printf "[m1-full] ISPP regressions (LK) -> JSON\n"
FECIM_REGRESSION_JSON_DIR="$OUT_DIR" go test ./module1-hysteresis/pkg/controller -run TestHeadlessRegression_WRD_ISPP_LK -count=1

printf "[m1-full] GUI/headless parity\n"
go test ./module1-hysteresis/tests -run TestM1_GUIHeadlessParity -count=1

printf "[m1-full] timestep convergence\n"
go test ./shared/physics -run TimestepConvergence -count=1

printf "[m1-full] OK\n"
