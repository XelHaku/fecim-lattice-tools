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
    duplicates: list[dict[str, object]] = []

    for group in _groups_by_sha(pdfs):
        selected_pdf, selected_match = _select_canonical_pdf(group, records)
        for pdf in group:
            if pdf != selected_pdf:
                duplicates.append(_duplicate_record(root, pdf, selected_pdf.path))

        pdf = selected_pdf
        match = selected_match
        if match.paper_key is None:
            unmatched.append(_unmatched_record(root, pdf))
            continue

        paper_key = match.paper_key
        _write_source(root, paper_key, pdf, match, records[paper_key])

        parsed_dir = root / "research" / "parsed" / paper_key
        marker_md = parsed_dir / "marker.md"
        existing_marker = marker_md.exists()
        parse_results: list[ParseResult] = [
            run_grobid_if_available(pdf.path, parsed_dir / "grobid.tei.xml")
        ]

        marker_result = copy_sidecar_markdown_if_present(pdf.path, marker_md)
        if marker_result.status != "ok":
            marker_result = run_marker_if_configured(pdf.path, marker_md)
        parse_results.append(marker_result)

        chunkable = marker_result.status == "ok"
        if not chunkable and existing_marker:
            reuse_result = ParseResult(
                paper_key=paper_key,
                parser="existing_markdown",
                status="ok",
                output_path=str(marker_md),
                message="reused existing parsed markdown",
            )
            parse_results.append(reuse_result)
            chunkable = True

        write_parse_manifest(
            parsed_dir / "manifest.json",
            [_normalize_parse_result(root, result, paper_key) for result in parse_results],
        )

        if chunkable and marker_md.exists():
            markdown = marker_md.read_text(encoding="utf-8", errors="replace")
            chunks = chunk_markdown(paper_key, markdown)
            write_chunks_jsonl(root / "research" / "chunks" / f"{paper_key}.jsonl", chunks)
            processed += 1

    _write_unmatched_report(root, unmatched)
    _write_duplicate_report(root, duplicates)
    _write_manifest(root, processed, len(unmatched), len(duplicates))
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


def _write_duplicate_report(root: Path, duplicates: list[dict[str, object]]) -> None:
    path = root / "research" / "reports" / "duplicate-pdfs.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {"duplicates": sorted(duplicates, key=lambda item: str(item["path"]))}
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _write_manifest(root: Path, processed: int, unmatched: int, duplicates: int) -> None:
    path = root / "research" / "manifests" / "ingest-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {"duplicates": duplicates, "processed": processed, "unmatched": unmatched}
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _unmatched_record(root: Path, pdf: DiscoveredPDF) -> dict[str, object]:
    return {
        "path": _display_path(root, pdf.path),
        "sha256": pdf.sha256,
        "size": pdf.size,
        "status": "unmatched",
    }


def _duplicate_record(root: Path, pdf: DiscoveredPDF, duplicate_of: Path) -> dict[str, object]:
    return {
        "duplicate_of": _display_path(root, duplicate_of),
        "path": _display_path(root, pdf.path),
        "sha256": pdf.sha256,
        "size": pdf.size,
        "status": "duplicate",
    }


def _groups_by_sha(pdfs: list[DiscoveredPDF]) -> list[list[DiscoveredPDF]]:
    groups: dict[str, list[DiscoveredPDF]] = {}
    for pdf in pdfs:
        groups.setdefault(pdf.sha256, []).append(pdf)
    return [sorted(groups[sha], key=lambda item: str(item.path)) for sha in sorted(groups)]


def _select_canonical_pdf(
    pdfs: list[DiscoveredPDF],
    records: dict[str, CitationRecord],
) -> tuple[DiscoveredPDF, PDFMatch]:
    matches = [(pdf, match_pdf_to_record(pdf, records)) for pdf in pdfs]
    return min(matches, key=lambda item: (_match_rank(item[1], item[0]), str(item[0].path)))


def _match_rank(match: PDFMatch, pdf: DiscoveredPDF) -> tuple[int, int]:
    if match.paper_key is None:
        return (1, 1)
    exact = pdf.path.stem.lower() == match.paper_key.lower()
    return (0, 0 if exact else 1)


def _normalize_parse_result(root: Path, result: ParseResult, paper_key: str) -> ParseResult:
    output_path = _display_path(root, Path(result.output_path))
    return ParseResult(
        paper_key=paper_key,
        parser=result.parser,
        status=result.status,
        output_path=output_path,
        message=_normalize_parse_message(root, result),
    )


def _normalize_parse_message(root: Path, result: ParseResult) -> str:
    if result.parser == "sidecar_markdown" and result.status == "ok":
        return f"copied {_display_path(root, Path(result.message.removeprefix('copied ')))}"
    if result.parser == "sidecar_markdown" and result.status == "skipped":
        return "sidecar markdown not found"
    root_text = str(root)
    return result.message.replace(root_text + "/", "").replace(root_text, ".")


def _display_path(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
