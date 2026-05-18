import json
from pathlib import Path

from .chunking import chunk_markdown, write_chunks_jsonl
from .citations import CitationRecord, load_citation_records
from .discovery import DiscoveredPDF, PDFMatch, discover_pdfs, match_pdf_to_record
from .parsing import (
    ParseResult,
    copy_sidecar_markdown_if_present,
    run_grobid_if_available,
    run_marker_if_configured,
    write_parse_manifest,
)
from .yamlio import dumps_yaml


def run_ingest(root: Path, extra_paths: list[Path]) -> int:
    records = load_citation_records(root)
    pdfs = discover_pdfs(root, extra_paths)
    processed = 0
    unmatched: list[dict[str, object]] = []

    for pdf in pdfs:
        match = match_pdf_to_record(pdf, records)
        if match.paper_key is None:
            unmatched.append(_unmatched_record(root, pdf))
            continue

        paper_key = match.paper_key
        _write_source(root, paper_key, pdf, match, records[paper_key])

        parsed_dir = root / "research" / "parsed" / paper_key
        marker_md = parsed_dir / "marker.md"
        parse_results: list[ParseResult] = [
            run_grobid_if_available(pdf.path, parsed_dir / "grobid.tei.xml")
        ]

        marker_result = copy_sidecar_markdown_if_present(pdf.path, marker_md)
        if marker_result.status != "ok":
            marker_result = run_marker_if_configured(pdf.path, marker_md)
        parse_results.append(marker_result)
        write_parse_manifest(parsed_dir / "manifest.json", parse_results)

        if marker_md.exists():
            markdown = marker_md.read_text(encoding="utf-8", errors="replace")
            chunks = chunk_markdown(paper_key, markdown)
            write_chunks_jsonl(root / "research" / "chunks" / f"{paper_key}.jsonl", chunks)
            processed += 1

    _write_unmatched_report(root, unmatched)
    _write_manifest(root, processed, len(unmatched))
    print(f"ingest complete: processed={processed} unmatched={len(unmatched)}")
    return 0


def _write_source(
    root: Path,
    paper_key: str,
    pdf: DiscoveredPDF,
    match: PDFMatch,
    record: CitationRecord,
) -> None:
    path = root / "research" / "sources" / f"{paper_key}.yaml"
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {
        "paper_key": paper_key,
        "citation_path": _display_path(root, record.path),
        "doi": record.doi,
        "match": {
            "confidence": match.confidence,
            "method": match.method,
            "status": match.status,
        },
        "pdf": {
            "path": _display_path(root, pdf.path),
            "sha256": pdf.sha256,
            "size": pdf.size,
        },
        "title": record.title,
    }
    path.write_text(dumps_yaml(data), encoding="utf-8")


def _write_unmatched_report(root: Path, unmatched: list[dict[str, object]]) -> None:
    path = root / "research" / "reports" / "unmatched-pdfs.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {"unmatched": sorted(unmatched, key=lambda item: str(item["path"]))}
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _write_manifest(root: Path, processed: int, unmatched: int) -> None:
    path = root / "research" / "manifests" / "ingest-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {"processed": processed, "unmatched": unmatched}
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _unmatched_record(root: Path, pdf: DiscoveredPDF) -> dict[str, object]:
    return {
        "path": _display_path(root, pdf.path),
        "sha256": pdf.sha256,
        "size": pdf.size,
        "status": "unmatched",
    }


def _display_path(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
