from __future__ import annotations

from pathlib import Path
import json

from .claims import load_claim_records
from .citations import CitationRecord, load_citation_records
from .reporting import write_content_addressed_report


def run_cite(root: Path, claim_id: str, json_output: bool = False) -> int:
    packet = build_citation_packet(root, claim_id)
    if packet is None:
        print(f"unknown claim id {claim_id}")
        return 1

    write_content_addressed_report(
        root,
        "research/reports/cite-latest.json",
        "research/reports/cites",
        packet,
    )

    if json_output:
        print(json.dumps(packet, indent=2, sort_keys=True))
    else:
        _print_text_packet(packet)
    return 0


def build_citation_packet(root: Path, claim_id: str) -> dict[str, object] | None:
    claims = load_claim_records(root)
    record = claims.get(claim_id)
    if record is None:
        return None

    citations = load_citation_records(root)
    openalex_records = _load_openalex_records(root)
    sources: list[dict[str, object]] = []
    missing_sources: list[str] = []
    for source_key in record.sources:
        source = _source_packet(root, source_key, citations.get(source_key), openalex_records.get(source_key))
        if source is None:
            missing_sources.append(source_key)
            continue
        sources.append(source)

    return {
        "claim": {
            "id": record.id,
            "claim": record.claim,
            "status": record.status,
            "confidence": record.confidence,
            "path": _rel(root, record.path),
        },
        "sources": sources,
        "missing_sources": missing_sources,
        "used_in": [_used_in_packet(root, claim_id, path) for path in record.used_in],
    }


def _source_packet(
    root: Path,
    key: str,
    citation: CitationRecord | None,
    openalex: tuple[Path, dict[str, object]] | None,
) -> dict[str, object] | None:
    if citation is not None:
        packet = {
            "key": key,
            "type": "citation",
            "path": _rel(root, citation.path),
            "title": citation.title,
            "doi": citation.doi,
            "arxiv_id": citation.arxiv_id,
        }
        if openalex is not None:
            packet["openalex_path"] = _rel(root, openalex[0])
            packet["openalex_id"] = str(openalex[1].get("id", ""))
        return packet

    if openalex is None:
        return None

    path, data = openalex
    return {
        "key": key,
        "type": "openalex",
        "path": _rel(root, path),
        "title": str(data.get("display_name", "")),
        "doi": str(data.get("doi", "")),
        "openalex_id": str(data.get("id", "")),
        "publication_year": data.get("publication_year", ""),
    }


def _used_in_packet(root: Path, claim_id: str, rel_path: str) -> dict[str, object]:
    path = root / rel_path
    exists = path.exists()
    references_claim = False
    if exists:
        references_claim = f"[claim: {claim_id}]" in path.read_text(encoding="utf-8", errors="replace")
    return {
        "path": rel_path,
        "exists": exists,
        "references_claim": references_claim,
    }


def _load_openalex_records(root: Path) -> dict[str, tuple[Path, dict[str, object]]]:
    records: dict[str, tuple[Path, dict[str, object]]] = {}
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return records
    for path in sorted(sources_dir.glob("*.openalex.json")):
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            continue
        if isinstance(data, dict):
            records[path.name.removesuffix(".openalex.json")] = (path, data)
    return records


def _print_text_packet(packet: dict[str, object]) -> None:
    claim = packet["claim"]
    assert isinstance(claim, dict)
    print(f"claim: {claim['id']}")
    print(f"status: {claim['status']}")
    print(f"confidence: {claim['confidence']}")
    print(f"text: {claim['claim']}")
    print("sources:")
    for source in packet["sources"]:
        assert isinstance(source, dict)
        title = source.get("title") or "(untitled)"
        print(f"- {source['key']}: {title} ({source['path']})")
    missing_sources = packet["missing_sources"]
    assert isinstance(missing_sources, list)
    if missing_sources:
        print("missing_sources:")
        for source in missing_sources:
            print(f"- {source}")
    print("used_in:")
    for item in packet["used_in"]:
        assert isinstance(item, dict)
        marker = "references claim" if item["references_claim"] else "missing claim marker"
        if not item["exists"]:
            marker = "missing file"
        print(f"- {item['path']}: {marker}")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
