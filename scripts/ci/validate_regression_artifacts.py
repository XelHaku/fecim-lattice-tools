#!/usr/bin/env python3
"""Validate required regression artifacts for Module 1 / Module 4 plans.

Usage examples:
  scripts/ci/validate_regression_artifacts.py --module module4 \
    --root output/regression/module4 --latest

  scripts/ci/validate_regression_artifacts.py --module module1 \
    --root output/validation/module1
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Iterable


MODULE4_REQUIRED_FILES = [
    "summary.json",
    "matrix.json",
    "thresholds.json",
    "results.json",
    "material_snapshot.json",
    "signal_chain_trace.json",
    "solver_diagnostics.json",
    "thermodynamics.json",
    "confidence_ledger.json",
    "uncertainty.json",
]

MODULE4_REQUIRED_KEYS = {
    "summary.json": ["git_hash"],
    "results.json": ["verdict"],
    "material_snapshot.json": ["material"],
    "solver_diagnostics.json": ["kcl_residual_max"],
    "thermodynamics.json": ["power"],
    "confidence_ledger.json": ["measured", "estimated", "placeholder"],
    "uncertainty.json": ["method", "confidence", "sample_size"],
}

MODULE1_REQUIRED_KEYS = [
    "schema_version",
    "timestamp_utc",
    "commit",
    "gate",
    "test_id",
    "material",
    "dataset",
    "metrics",
    "uncertainty",
    "thresholds",
    "verdict",
]


class ValidationError(Exception):
    pass


def load_json(path: Path):
    try:
        with path.open("r", encoding="utf-8") as f:
            return json.load(f)
    except FileNotFoundError as e:
        raise ValidationError(f"missing file: {path}") from e
    except json.JSONDecodeError as e:
        raise ValidationError(f"invalid json: {path} ({e})") from e


def resolve_run_root(root: Path, latest: bool) -> Path:
    if not latest:
        return root
    if not root.exists():
        raise ValidationError(f"root not found: {root}")
    dirs = [p for p in root.iterdir() if p.is_dir()]
    if not dirs:
        return root
    return sorted(dirs, key=lambda p: p.stat().st_mtime, reverse=True)[0]


def require_keys(obj: dict, keys: Iterable[str], where: str):
    missing = [k for k in keys if k not in obj]
    if missing:
        raise ValidationError(f"missing keys in {where}: {', '.join(missing)}")


def validate_module4(run_root: Path):
    for name in MODULE4_REQUIRED_FILES:
        p = run_root / name
        data = load_json(p)
        keys = MODULE4_REQUIRED_KEYS.get(name)
        if keys:
            require_keys(data, keys, name)

    # Minimal semantic checks
    summary = load_json(run_root / "summary.json")
    if not str(summary.get("git_hash", "")).strip():
        raise ValidationError("summary.json: git_hash is empty")

    uncertainty = load_json(run_root / "uncertainty.json")
    if int(uncertainty.get("sample_size", 0)) <= 0:
        raise ValidationError("uncertainty.json: sample_size must be > 0")



def validate_module1(root: Path):
    # Only validate PE-loop artifacts; other artifact types (FORC, arrhenius,
    # kinetics, chi-squared, etc.) have distinct schemas and are not required to
    # carry the full envelope.
    files = sorted(root.rglob("module1_pe_loop_*.json"))
    if not files:
        raise ValidationError(f"no module1 PE-loop artifacts found under {root}")

    checked = 0
    for p in files:
        data = load_json(p)
        if not isinstance(data, dict):
            raise ValidationError(f"artifact not object: {p}")
        require_keys(data, MODULE1_REQUIRED_KEYS, str(p))

        uncertainty = data.get("uncertainty", {})
        if not isinstance(uncertainty, dict):
            raise ValidationError(f"uncertainty not object: {p}")
        require_keys(uncertainty, ["method", "confidence", "sample_size"], f"{p}:uncertainty")
        if int(uncertainty.get("sample_size", 0)) <= 0:
            raise ValidationError(f"sample_size must be > 0: {p}")

        checked += 1

    print(f"[artifact-validate] module1 checked {checked} artifact files")



def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--module", choices=["module1", "module4"], required=True)
    ap.add_argument("--root", required=True, help="artifact root directory")
    ap.add_argument("--latest", action="store_true", help="use latest timestamped run subdir")
    args = ap.parse_args()

    root = Path(args.root)
    run_root = resolve_run_root(root, args.latest)

    try:
        if args.module == "module4":
            validate_module4(run_root)
        else:
            validate_module1(run_root)
    except ValidationError as e:
        print(f"[artifact-validate] FAIL: {e}", file=sys.stderr)
        return 1

    print(f"[artifact-validate] PASS: {args.module} artifacts valid at {run_root}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
