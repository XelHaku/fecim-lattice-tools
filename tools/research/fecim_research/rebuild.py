from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass
from pathlib import Path
import json


@dataclass(frozen=True)
class RebuildStages:
    ingest: Callable[[], int]
    missing: Callable[[], int]
    index: Callable[[], int]
    cache: Callable[[], int]
    audit: Callable[[], int]
    graph: Callable[[], int]


def run_rebuild(
    root: Path,
    extra_paths: list[Path],
    semantic: bool,
    embedding_model: str,
    skip_index: bool,
    stages: RebuildStages | None = None,
) -> int:
    if stages is None:
        stages = _default_stages(root, extra_paths, semantic, embedding_model)

    stage_results: list[dict[str, object]] = []
    stage_results.append(_run_stage("ingest", stages.ingest))
    stage_results.append(_run_stage("missing", stages.missing))
    if skip_index:
        stage_results.append(_skipped_stage("index"))
    else:
        stage_results.append(_run_stage("index", stages.index))
    stage_results.append(_run_stage("cache", stages.cache, warning_only=True))
    stage_results.append(_run_stage("audit", stages.audit))
    stage_results.append(_run_stage("graph", stages.graph))

    failed = sum(1 for result in stage_results if result["status"] == "failed")
    warnings = sum(1 for result in stage_results if result["status"] == "warning")
    report = {
        "ok": failed == 0,
        "failed": failed,
        "warnings": warnings,
        "skipped": sum(1 for result in stage_results if result["status"] == "skipped"),
        "semantic": semantic,
        "embedding_model": embedding_model,
        "extra_paths": [_path_text(path) for path in extra_paths],
        "stages": stage_results,
    }
    report_path = root / "research" / "reports" / "rebuild-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    print(
        "research rebuild complete: "
        f"stages={len(stage_results)} failed={report['failed']} "
        f"warnings={report['warnings']} skipped={report['skipped']}"
    )
    if failed == 0:
        return 0
    for result in stage_results:
        if result["status"] == "failed":
            return int(result["exit_code"])
    return 1


def _default_stages(root: Path, extra_paths: list[Path], semantic: bool, embedding_model: str) -> RebuildStages:
    from .cache import run_cache
    from .claims import run_audit
    from .graphing import run_graph
    from .indexing import run_index
    from .ingest import run_ingest
    from .missing import run_missing

    return RebuildStages(
        ingest=lambda: run_ingest(root=root, extra_paths=extra_paths),
        missing=lambda: run_missing(root=root),
        index=lambda: run_index(root=root, semantic=semantic, embedding_model=embedding_model),
        cache=lambda: run_cache(root=root),
        audit=lambda: run_audit(root=root),
        graph=lambda: run_graph(root=root),
    )


def _run_stage(stage: str, runner: Callable[[], int], warning_only: bool = False) -> dict[str, object]:
    code = runner()
    if warning_only and code != 0:
        status = "warning"
    else:
        status = "ok" if code == 0 else "failed"
    return {
        "stage": stage,
        "status": status,
        "exit_code": code,
        "artifacts": _stage_artifacts(stage),
    }


def _skipped_stage(stage: str) -> dict[str, object]:
    return {
        "stage": stage,
        "status": "skipped",
        "exit_code": 0,
        "artifacts": _stage_artifacts(stage),
    }


def _stage_artifacts(stage: str) -> list[str]:
    artifacts = {
        "ingest": [
            "research/manifests/ingest-latest.json",
            "research/reports/duplicate-pdfs.json",
            "research/reports/unmatched-pdfs.json",
        ],
        "missing": [
            "research/reports/missing-papers-latest.json",
        ],
        "index": [
            "research/manifests/index-latest.json",
            "research/manifests/index-pyserini.json",
            "research/manifests/index-lancedb.json",
            "research/index/pyserini",
            "research/index/lancedb",
        ],
        "cache": [
            "research/reports/cache-latest.json",
        ],
        "audit": [
            "research/reports/claim-audit-latest.json",
        ],
        "graph": [
            "research/graphs/provenance-graph.json",
            "research/reports/graph-latest.json",
        ],
    }
    return artifacts.get(stage, [])


def _path_text(path: Path) -> str:
    return path.as_posix()
