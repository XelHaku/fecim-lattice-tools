from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
import hashlib
import json
import re
import shutil

from .citations import load_citation_records
from .yamlio import dumps_yaml


@dataclass(frozen=True)
class PromotionResult:
    paper_key: str
    status: str
    citation_path: str = ""
    source_path: str = ""
    destination_path: str = ""
    sha256: str = ""
    size: int = 0
    license: str = ""
    license_url: str = ""
    review_note: str = ""
    promotion_ledger_path: str = ""
    message: str = ""


def run_promote_pdf(
    root: Path,
    key: str,
    destination: str,
    source: str,
    license_name: str = "",
    license_url: str = "",
    review_note: str = "",
) -> int:
    license_name = license_name.strip()
    license_url = license_url.strip()
    review_note = review_note.strip()

    records = load_citation_records(root)
    record = records.get(key)
    if record is None:
        result = PromotionResult(key, "failed", message=f"unknown citation key {key}")
        _write_report(root, result)
        print(f"pdf promotion failed: {result.message}")
        return 1

    source_path = source.strip() or _local_pdf_from_citation(record.path)
    validation_error = _validate_paths(source_path, destination)
    if validation_error:
        result = PromotionResult(
            key,
            "failed",
            citation_path=_rel(root, record.path),
            source_path=source_path,
            destination_path=destination,
            message=validation_error,
        )
        _write_report(root, result)
        print(f"pdf promotion failed: {result.message}")
        return 1

    metadata_error = _validate_review_metadata(license_name, license_url, review_note)
    if metadata_error:
        result = PromotionResult(
            key,
            "failed",
            citation_path=_rel(root, record.path),
            source_path=source_path,
            destination_path=destination,
            license=license_name,
            license_url=license_url,
            review_note=review_note,
            message=metadata_error,
        )
        _write_report(root, result)
        print(f"pdf promotion failed: {result.message}")
        return 1

    source_file = root / source_path
    if not source_file.is_file():
        result = PromotionResult(
            key,
            "failed",
            citation_path=_rel(root, record.path),
            source_path=source_path,
            destination_path=destination,
            message=f"source PDF does not exist: {source_path}",
        )
        _write_report(root, result)
        print(f"pdf promotion failed: {result.message}")
        return 1

    destination_file = root / destination
    digest = _sha256_file(source_file)
    size = source_file.stat().st_size
    if destination_file.exists() and _sha256_file(destination_file) != digest:
        result = PromotionResult(
            key,
            "failed",
            citation_path=_rel(root, record.path),
            source_path=source_path,
            destination_path=destination,
            sha256=digest,
            size=size,
            license=license_name,
            license_url=license_url,
            review_note=review_note,
            message=f"destination PDF already exists with different content: {destination}",
        )
        _write_report(root, result)
        print(f"pdf promotion failed: {result.message}")
        return 1

    destination_file.parent.mkdir(parents=True, exist_ok=True)
    if not destination_file.exists():
        shutil.copy2(source_file, destination_file)
    _update_citation(record.path, destination, digest, size)

    promotion_ledger_path = f"research/sources/{key}.promotion.yaml"
    result = PromotionResult(
        key,
        "promoted",
        citation_path=_rel(root, record.path),
        source_path=source_path,
        destination_path=destination,
        sha256=digest,
        size=size,
        license=license_name,
        license_url=license_url,
        review_note=review_note,
        promotion_ledger_path=promotion_ledger_path,
        message="promoted reviewed PDF to tracked canonical path",
    )
    _write_promotion_ledger(root, result)
    _write_report(root, result)
    _refresh_missing_report(root)
    print(f"pdf promotion complete: paper_key={key} destination={destination}")
    return 0


def _validate_paths(source: str, destination: str) -> str:
    if not source:
        return "source PDF path is required; pass --source or add a Local PDF field"
    if not _is_repo_relative_path(source):
        return "source PDF path must be repo-relative"
    if not destination:
        return "destination PDF path is required"
    if not _is_repo_relative_path(destination):
        return "destination PDF path must be repo-relative"
    if not destination.lower().endswith(".pdf"):
        return "destination path must end with .pdf"
    if _is_ignored_pdf_inbox_path(destination):
        return "destination must be a tracked canonical path, not ignored research/papers"
    if not _is_tracked_pdf_collection(destination):
        return "destination must be under a tracked canonical PDF collection"
    return ""


def _validate_review_metadata(license_name: str, license_url: str, review_note: str) -> str:
    if not license_name or not license_url or not review_note:
        return "license, license_url, and review_note are required before promotion"
    if not license_url.startswith(("http://", "https://")):
        return "license_url must be an http or https URL"
    return ""


def _local_pdf_from_citation(path: Path) -> str:
    return _field_value(path.read_text(encoding="utf-8", errors="replace"), "Local PDF")


def _field_value(text: str, name: str) -> str:
    pattern = re.compile(rf"^\*\*{re.escape(name)}:\*\*\s*`([^`]+)`", re.MULTILINE)
    match = pattern.search(text)
    if match is None:
        return ""
    return match.group(1).strip()


def _update_citation(path: Path, pdf_path: str, sha256: str, size: int) -> None:
    text = path.read_text(encoding="utf-8", errors="replace")
    text = _set_field(text, "PDF", pdf_path)
    text = _set_field(text, "SHA256", sha256)
    text = _set_field(text, "Size", str(size))
    path.write_text(text, encoding="utf-8")


def _set_field(text: str, name: str, value: str) -> str:
    line = f"**{name}:** `{value}`"
    pattern = re.compile(rf"^\*\*{re.escape(name)}:\*\*.*$", re.MULTILINE)
    if pattern.search(text):
        return pattern.sub(line, text, count=1)
    separator = "\n---\n"
    if separator in text:
        return text.replace(separator, f"{line}\n{separator}", 1)
    suffix = "" if text.endswith("\n") else "\n"
    return f"{text}{suffix}{line}\n"


def _write_report(root: Path, result: PromotionResult) -> None:
    path = root / "research" / "reports" / "pdf-promotion-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(asdict(result), indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _write_promotion_ledger(root: Path, result: PromotionResult) -> None:
    path = root / result.promotion_ledger_path
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(dumps_yaml(asdict(result)), encoding="utf-8")


def _refresh_missing_report(root: Path) -> None:
    from .missing import run_missing

    run_missing(root=root)


def _sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def _is_repo_relative_path(path: str) -> bool:
    rel = Path(path)
    return not rel.is_absolute() and ".." not in rel.parts


def _is_ignored_pdf_inbox_path(path: str) -> bool:
    return path == "research/papers" or path.startswith("research/papers/")


def _is_tracked_pdf_collection(path: str) -> bool:
    return path.startswith("docs/4-research/papers/") or path.startswith("citations/pdfs/")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
