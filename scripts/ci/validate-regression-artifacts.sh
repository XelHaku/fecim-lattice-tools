#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: $0 <module1|module4> <artifact-root> [--latest]" >&2
  exit 2
fi

MODULE="$1"
ROOT="$2"
shift 2

exec python3 scripts/ci/validate_regression_artifacts.py --module "$MODULE" --root "$ROOT" "$@"
