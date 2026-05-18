import json
import re
import sys
from collections import Counter
from pathlib import Path
from typing import Any

from .chunking import chunk_markdown
from .claims import ClaimRecord, load_claim_records
from .indexing import _sha, collect_chunk_files


STALE_INDEX_MESSAGE = "BM25 index is stale; rerun `fecim research index`"
INBOX_SEARCH_BACKEND = "inbox-local-jsonl"


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def load_chunk_lookup(root: Path) -> dict[str, dict[str, object]]:
    lookup: dict[str, dict[str, object]] = {}
    for path in collect_chunk_files(root):
        with path.open(encoding="utf-8") as f:
            for line in f:
                if not line.strip():
                    continue
                record = json.loads(line)
                record["chunk_file"] = _repo_relative(root, path)
                chunk_id = record.get("id")
                if isinstance(chunk_id, str):
                    lookup[chunk_id] = record
    return lookup


def load_inbox_chunk_lookup(root: Path) -> dict[str, dict[str, object]]:
    lookup: dict[str, dict[str, object]] = {}
    report = _read_inbox_report(root)
    local_only = report.get("local_only", [])
    if not isinstance(local_only, list):
        return lookup

    for item in local_only:
        if not isinstance(item, dict) or item.get("status") != "needs_promotion":
            continue
        path_value = item.get("path")
        if not isinstance(path_value, str) or not path_value:
            continue
        pdf_path = Path(path_value)
        source_pdf = pdf_path if pdf_path.is_absolute() else root / pdf_path
        markdown_path = source_pdf.with_suffix(".md")
        try:
            markdown = markdown_path.read_text(encoding="utf-8")
        except OSError:
            continue

        paper_key_value = item.get("paper_key")
        paper_key = paper_key_value if isinstance(paper_key_value, str) and paper_key_value else source_pdf.stem
        for chunk in chunk_markdown(paper_key, markdown):
            chunk_id = chunk.get("id")
            if not isinstance(chunk_id, str):
                continue
            inbox_chunk = dict(chunk)
            inbox_chunk["source_parser"] = "inbox-sidecar-markdown"
            inbox_chunk["source_path"] = _repo_relative(root, markdown_path)
            inbox_chunk["pdf_path"] = _repo_relative(root, source_pdf)
            inbox_chunk["citation_path"] = item.get("citation_path", "")
            inbox_chunk["review_status"] = item.get("status", "")
            inbox_chunk["trust_state"] = "unreviewed"
            inbox_chunk["review_required"] = True
            inbox_chunk["inbox"] = True
            inbox_chunk["inbox_sha256"] = item.get("sha256", "")
            lookup[chunk_id] = inbox_chunk
    return lookup


def _read_inbox_report(root: Path) -> dict[str, object]:
    path = root / "research" / "reports" / "local-inbox-pdfs.json"
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    return data if isinstance(data, dict) else {}


def index_is_stale(root: Path) -> bool:
    manifest_path = root / "research" / "manifests" / "index-latest.json"
    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return True

    inputs = manifest.get("inputs")
    if not isinstance(inputs, list):
        return True

    expected = []
    for item in inputs:
        if not isinstance(item, dict):
            return True
        path = item.get("path")
        sha256 = item.get("sha256")
        if not isinstance(path, str) or not isinstance(sha256, str):
            return True
        expected.append({"path": path, "sha256": sha256})

    current = [{"path": _repo_relative(root, path), "sha256": _sha(path)} for path in collect_chunk_files(root)]
    return sorted(current, key=lambda item: item["path"]) != sorted(expected, key=lambda item: item["path"])


def render_text_results(rows: list[dict[str, object]]) -> str:
    if not rows:
        return ""
    lines = []
    for row in rows:
        rank = row.get("rank", "")
        paper_key = row.get("paper_key", "")
        score = row.get("score", "")
        section = row.get("section", "")
        docid = row.get("docid", "")
        snippet = row.get("snippet", "")
        marker = " [UNREVIEWED]" if _row_is_unreviewed(row) else ""
        lines.append(f"{rank}. {paper_key}{marker} score={score} section={section} chunk={docid}")
        if snippet:
            lines.append(str(snippet))
    return "\n".join(lines) + "\n"


def _row_is_unreviewed(row: dict[str, object]) -> bool:
    return row.get("trust_state") == "unreviewed" or row.get("inbox") is True


def _snippet(text: str, limit: int = 240) -> str:
    compact = re.sub(r"\s+", " ", text).strip()
    if len(compact) <= limit:
        return compact
    return compact[: max(0, limit - 3)].rstrip() + "..."


def _tokens(text: str) -> list[str]:
    return re.findall(r"[a-z0-9]+", text.lower())


def _local_score(query_terms: list[str], record: dict[str, object]) -> float:
    text = " ".join(
        str(record.get(key, ""))
        for key in ["paper_key", "section", "contents"]
    )
    counts = Counter(_tokens(text))
    unique_terms = sorted(set(query_terms))
    hits = sum(counts[term] for term in unique_terms)
    if hits == 0:
        return 0.0
    coverage = sum(1 for term in unique_terms if counts[term] > 0)
    return float((coverage * 10) + hits)


def _row(rank: int, score: float, docid: str, record: dict[str, Any]) -> dict[str, object]:
    contents = str(record.get("contents", ""))
    row: dict[str, object] = {
        "rank": rank,
        "score": score,
        "docid": docid,
        "paper_key": record.get("paper_key", ""),
        "section": record.get("section", ""),
        "contents": contents,
        "snippet": _snippet(contents),
        "chunk_file": record.get("chunk_file", ""),
        "source_parser": record.get("source_parser", ""),
        "source_path": record.get("source_path", ""),
        "section_number": record.get("section_number"),
        "chunk_number": record.get("chunk_number"),
        "page_start": record.get("page_start"),
        "page_end": record.get("page_end"),
        "char_start": record.get("char_start"),
        "char_end": record.get("char_end"),
        "sha256": record.get("sha256", ""),
    }
    for key in [
        "trust_state",
        "review_required",
        "inbox",
        "pdf_path",
        "citation_path",
        "review_status",
        "inbox_sha256",
    ]:
        if key in record:
            row[key] = record.get(key)
    return row


def search_chunks_locally(root: Path, query: str, limit: int) -> list[dict[str, object]]:
    return _search_lookup(load_chunk_lookup(root), query, limit)


def search_inbox_chunks_locally(root: Path, query: str, limit: int) -> list[dict[str, object]]:
    return _search_lookup(load_inbox_chunk_lookup(root), query, limit)


def _search_lookup(lookup: dict[str, dict[str, object]], query: str, limit: int) -> list[dict[str, object]]:
    query_terms = _tokens(query)
    if not query_terms:
        return []

    scored: list[tuple[float, str, dict[str, Any]]] = []
    for docid in sorted(lookup):
        record = lookup[docid]
        score = _local_score(query_terms, record)
        if score > 0:
            scored.append((score, docid, record))

    scored.sort(key=lambda item: (-item[0], str(item[2].get("paper_key", "")), item[1]))
    return [_row(rank, score, docid, record) for rank, (score, docid, record) in enumerate(scored[:limit], start=1)]


def _claim_context(root: Path, record: ClaimRecord | None) -> dict[str, object] | None:
    if record is None:
        return None
    return {
        "id": record.id,
        "claim": record.claim,
        "status": record.status,
        "confidence": record.confidence,
        "sources": record.sources,
        "path": _repo_relative(root, record.path),
    }


def write_search_report(
    root: Path,
    query: str,
    backend: str,
    rows: list[dict[str, object]],
    claim: dict[str, object] | None = None,
    trust_state: str | None = None,
    review_required: bool | None = None,
) -> Path:
    report_path = root / "research" / "reports" / "search-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report = {
        "ok": True,
        "backend": backend,
        "query": query,
        "result_count": len(rows),
        "results": rows,
    }
    if claim is not None:
        report["claim"] = claim
    if trust_state is not None:
        report["trust_state"] = trust_state
    if review_required is not None:
        report["review_required"] = review_required
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return report_path


def run_search(
    root: Path,
    query: str,
    limit: int,
    json_output: bool,
    local: bool = False,
    inbox: bool = False,
    claim_id: str = "",
) -> int:
    if inbox and claim_id:
        print("inbox search cannot be claim-linked; promote the PDF before evidence search", file=sys.stderr)
        return 1

    claim: dict[str, object] | None = None
    if claim_id:
        record = load_claim_records(root).get(claim_id)
        if record is None:
            print(f"unknown claim id {claim_id}", file=sys.stderr)
            return 1
        query = record.claim
        claim = _claim_context(root, record)

    if inbox:
        rows = search_inbox_chunks_locally(root, query, limit)
        write_search_report(
            root,
            query,
            INBOX_SEARCH_BACKEND,
            rows,
            trust_state="unreviewed",
            review_required=True,
        )
        if json_output:
            print(json.dumps(rows, indent=2, sort_keys=True))
        else:
            sys.stdout.write(render_text_results(rows))
        return 0

    if local:
        rows = search_chunks_locally(root, query, limit)
        write_search_report(root, query, "local-jsonl", rows, claim=claim)
        if json_output:
            print(json.dumps(rows, indent=2, sort_keys=True))
        else:
            sys.stdout.write(render_text_results(rows))
        return 0

    index_dir = root / "research" / "index" / "pyserini"
    if not index_dir.is_dir():
        print("missing BM25 index; run `fecim research index` first", file=sys.stderr)
        return 1
    if index_is_stale(root):
        print(STALE_INDEX_MESSAGE, file=sys.stderr)
        return 1

    try:
        from pyserini.search.lucene import LuceneSearcher
    except ImportError:
        print("Pyserini is not installed; install pyserini to run BM25 evidence search.", file=sys.stderr)
        return 1

    searcher = LuceneSearcher(str(index_dir))
    hits = searcher.search(query, k=limit)
    lookup = load_chunk_lookup(root)
    rows = []
    for rank, hit in enumerate(hits, start=1):
        docid = hit.docid
        record = lookup.get(docid, {"id": docid, "contents": ""})
        rows.append(_row(rank, hit.score, docid, record))

    write_search_report(root, query, "pyserini", rows, claim=claim)
    if json_output:
        print(json.dumps(rows, indent=2, sort_keys=True))
    else:
        sys.stdout.write(render_text_results(rows))
    return 0
