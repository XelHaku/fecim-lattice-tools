#!/usr/bin/env bash
set -euo pipefail

# Required gate: CI test-all must run fully headless.
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

# CI-friendly defaults (override via env vars)
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-10m}"

# -count=1 avoids cached test results hiding flakes.
# -shuffle=off keeps execution order stable by default.
exec go test \
  -tags=ci \
  -count=1 \
  -shuffle=off \
  -trimpath \
  -timeout="${GO_TEST_TIMEOUT}" \
  ./...
