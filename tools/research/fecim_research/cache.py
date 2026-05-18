import hashlib
import json
import shutil
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class CacheSpec:
    name: str
    kind: str
    path: str
    manifest: str
    rebuild_command: str
    ignore_rule: str
    required: bool


CACHE_SPECS = [
    CacheSpec(
        name="pyserini",
        kind="bm25",
        path="research/index/pyserini",
        manifest="research/manifests/index-latest.json",
        rebuild_command="fecim research index",
        ignore_rule="/index/pyserini/",
        required=True,
    ),
    CacheSpec(
        name="lancedb",
        kind="semantic-vector",
        path="research/index/lancedb",
        manifest="research/manifests/index-latest.json",
        rebuild_command="fecim research index --semantic",
        ignore_rule="/index/lancedb/",
        required=False,
    ),
    CacheSpec(
        name="models",
        kind="embedding-model-cache",
        path="research/index/models",
        manifest="",
        rebuild_command="fecim research index --semantic",
        ignore_rule="/index/models/",
        required=False,
    ),
    CacheSpec(
        name="scratch",
        kind="tool-scratch",
        path="research/.cache",
        manifest="",
        rebuild_command="fecim research rebuild",
        ignore_rule="/.cache/",
        required=False,
    ),
]


def build_cache_report(root: Path) -> dict[str, Any]:
    caches = [_build_cache_entry(root, spec) for spec in CACHE_SPECS]
    required_caches = [cache for cache in caches if cache["required"]]
    stale = [cache for cache in required_caches if cache.get("stale")]
    missing = [cache for cache in required_caches if cache["status"] != "ready"]
    return {
        "ok": not stale and not missing,
        "caches": caches,
        "summary": {
            "total": len(caches),
            "required": len(required_caches),
            "ready": sum(1 for cache in caches if cache["status"] == "ready"),
            "missing": sum(1 for cache in required_caches if cache["status"] == "missing"),
            "stale": len(stale),
            "optional": sum(1 for cache in caches if not cache["required"]),
        },
    }


def run_cache(root: Path, clean: bool = False) -> int:
    cleanup = _clean_cache_dirs(root) if clean else None
    report = build_cache_report(root)
    if cleanup is not None:
        report["cleanup"] = cleanup
    report_path = root / "research" / "reports" / "cache-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    summary = report["summary"]
    print(
        "research cache complete: "
        f"required={summary['required']} "
        f"ready={summary['ready']} "
        f"missing={summary['missing']} "
        f"stale={summary['stale']} "
        f"report={_repo_relative(root, report_path)}"
    )
    if cleanup is not None:
        print(
            "research cache clean: "
            f"status={cleanup['status']} "
            f"removed={len(cleanup['removed'])} "
            f"refused={len(cleanup['refused'])}"
        )
        return 0 if cleanup["status"] == "ok" else 2
    return 0 if report["ok"] else 1


def _clean_cache_dirs(root: Path) -> dict[str, Any]:
    entries = [_build_cache_entry(root, spec) for spec in CACHE_SPECS]
    existing = [entry for entry in entries if entry["exists"]]
    refused = sorted(entry["path"] for entry in existing if not entry["ignored_by_policy"])
    if refused:
        return {
            "status": "refused",
            "removed": [],
            "refused": refused,
        }

    removed: list[str] = []
    for entry in existing:
        path = root / str(entry["path"])
        if path.is_dir():
            shutil.rmtree(path)
        else:
            path.unlink()
        removed.append(str(entry["path"]))
    return {
        "status": "ok",
        "removed": sorted(removed),
        "refused": [],
    }


def _build_cache_entry(root: Path, spec: CacheSpec) -> dict[str, Any]:
    cache_path = root / spec.path
    manifest_path = root / spec.manifest if spec.manifest else None
    entry: dict[str, Any] = {
        "name": spec.name,
        "kind": spec.kind,
        "path": spec.path,
        "rebuildable": True,
        "required": spec.required,
        "rebuild_command": spec.rebuild_command,
        "ignored_by_policy": _ignored_by_policy(root, spec.ignore_rule),
        "exists": cache_path.exists(),
        "manifest": spec.manifest,
        "manifest_exists": bool(manifest_path and manifest_path.exists()),
        "stale": False,
        "stale_inputs": [],
        "inputs": 0,
    }

    if not spec.required:
        entry["status"] = "ready" if cache_path.exists() else "optional"
        return entry

    if not cache_path.exists() and not entry["manifest_exists"]:
        entry["status"] = "missing"
        return entry
    if not cache_path.exists():
        entry["status"] = "missing"
        return entry
    if manifest_path is None or not manifest_path.exists():
        entry["status"] = "missing-manifest"
        return entry

    manifest = _read_json(manifest_path)
    inputs = manifest.get("inputs", []) if isinstance(manifest, dict) else []
    stale_inputs = _stale_inputs(root, inputs)
    entry["inputs"] = len(inputs)
    entry["stale_inputs"] = stale_inputs
    entry["stale"] = bool(stale_inputs)
    entry["status"] = "stale" if stale_inputs else "ready"
    return entry


def _ignored_by_policy(root: Path, ignore_rule: str) -> bool:
    gitignore = root / "research" / ".gitignore"
    if not gitignore.exists():
        return False
    rules = [
        line.strip()
        for line in gitignore.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]
    return ignore_rule in rules


def _read_json(path: Path) -> dict[str, Any]:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    return data if isinstance(data, dict) else {}


def _stale_inputs(root: Path, inputs: list[Any]) -> list[str]:
    stale: list[str] = []
    for item in inputs:
        if not isinstance(item, dict):
            continue
        rel = item.get("path")
        expected_sha = item.get("sha256")
        if not isinstance(rel, str) or not isinstance(expected_sha, str):
            continue
        path = root / rel
        if not path.exists() or _sha(path) != expected_sha:
            stale.append(rel)
    return sorted(stale)


def _sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()
