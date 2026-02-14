#!/usr/bin/env bash
set -euo pipefail

# Required gate: race lane must run fully headless.
# Set FECIM_ALLOW_DISPLAY_IN_CI=1 only for explicit local debugging.
if [[ "${FECIM_ALLOW_DISPLAY_IN_CI:-0}" != "1" ]]; then
  if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
    echo "[headless-gate] DISPLAY/WAYLAND_DISPLAY must be unset for required CI lanes" >&2
    echo "[headless-gate] unset DISPLAY WAYLAND_DISPLAY (or set FECIM_ALLOW_DISPLAY_IN_CI=1 for local debug only)" >&2
    exit 1
  fi
fi

# Keep tests display-agnostic even when launched from desktop shells.
unset DISPLAY
unset WAYLAND_DISPLAY

GO_TEST_RACE_TIMEOUT="${GO_TEST_RACE_TIMEOUT:-20m}"

# Keep the race set small/targeted to control runtime.
# Avoid GUI-heavy packages to keep CI headless and deterministic.
RACE_PKGS=(
  "./shared/compute"
  "./shared/gpu"
  "./shared/io"
  "./shared/logging"
  "./shared/peripherals"
  "./shared/physics"
  "./shared/recording"
  "./shared/theme"
  "./shared/utils"
  "./module1-hysteresis/pkg/algo"
  "./module1-hysteresis/pkg/ferroelectric"
  "./module1-hysteresis/pkg/simulation"
  "./module2-crossbar/pkg/crossbar"
  "./module2-crossbar/pkg/network"
  "./module2-crossbar/pkg/training"
  "./module2-crossbar/pkg/weights"
  "./module3-mnist/pkg/core"
  "./module3-mnist/pkg/mnist"
  "./module3-mnist/pkg/training"
  "./module4-circuits/pkg/arraysim"
  "./module5-comparison/pkg/comparison"
  "./module6-eda/pkg/compiler"
  "./module6-eda/pkg/config"
)

exec go test \
  -tags=ci \
  -race \
  -count=1 \
  -shuffle=off \
  -trimpath \
  -timeout="${GO_TEST_RACE_TIMEOUT}" \
  "${RACE_PKGS[@]}"
