from __future__ import annotations

from collections.abc import Iterable
from pathlib import Path

_SOURCE_LEDGER_EXCLUDED_SUFFIXES = (
    ".acquisition.yaml",
    ".promotion.yaml",
)


def sources_dir(root: Path) -> Path:
    return root / "research" / "sources"


def source_ledger_paths(root: Path) -> list[Path]:
    return _sorted_existing(
        path
        for path in sources_dir(root).rglob("*.yaml")
        if not path.name.endswith(_SOURCE_LEDGER_EXCLUDED_SUFFIXES)
    )


def openalex_record_paths(root: Path) -> list[Path]:
    return _sorted_existing(sources_dir(root).rglob("*.openalex.json"))


def acquisition_ledger_paths(root: Path) -> list[Path]:
    return _sorted_existing(sources_dir(root).rglob("*.acquisition.yaml"))


def promotion_ledger_paths(root: Path) -> list[Path]:
    return _sorted_existing(sources_dir(root).rglob("*.promotion.yaml"))


def acquisition_ledger_path(root: Path, paper_key: str, near: Path | None = None) -> Path:
    return _sidecar_path(root, paper_key, ".acquisition.yaml", near)


def _sidecar_path(root: Path, paper_key: str, suffix: str, near: Path | None = None) -> Path:
    filename = f"{paper_key}{suffix}"
    if near is not None:
        candidate = near.parent / filename
        if candidate.is_file():
            return candidate
    root_candidate = sources_dir(root) / filename
    if root_candidate.is_file():
        return root_candidate
    matches = sorted(sources_dir(root).rglob(filename), key=lambda path: path.as_posix())
    if matches:
        return matches[0]
    return root_candidate


def _sorted_existing(paths: Iterable[Path]) -> list[Path]:
    return sorted((path for path in paths if path.is_file()), key=lambda path: path.as_posix())
