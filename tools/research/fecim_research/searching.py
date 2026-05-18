import json
import re
import sys
from collections import Counter
from pathlib import Path
from typing import Any

from .indexing import _sha, collect_chunk_files


STALE_INDEX_MESSAGE = "BM25 index is stale; rerun `fecim research index`"


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
        lines.append(f"{rank}. {paper_key} score={score} section={section} chunk={docid}")
        if snippet:
            lines.append(str(snippet))
    return "\n".join(lines) + "\n"


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
    return {
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


def search_chunks_locally(root: Path, query: str, limit: int) -> list[dict[str, object]]:
    query_terms = _tokens(query)
    if not query_terms:
        return []

    lookup = load_chunk_lookup(root)
    scored: list[tuple[float, str, dict[str, Any]]] = []
    for docid in sorted(lookup):
        record = lookup[docid]
        score = _local_score(query_terms, record)
        if score > 0:
            scored.append((score, docid, record))

    scored.sort(key=lambda item: (-item[0], str(item[2].get("paper_key", "")), item[1]))
    return [_row(rank, score, docid, record) for rank, (score, docid, record) in enumerate(scored[:limit], start=1)]


def write_search_report(root: Path, query: str, backend: str, rows: list[dict[str, object]]) -> Path:
    report_path = root / "research" / "reports" / "search-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report = {
        "ok": True,
        "backend": backend,
        "query": query,
        "result_count": len(rows),
        "results": rows,
    }
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return report_path


def run_search(root: Path, query: str, limit: int, json_output: bool, local: bool = False) -> int:
    if local:
        rows = search_chunks_locally(root, query, limit)
        write_search_report(root, query, "local-jsonl", rows)
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

    write_search_report(root, query, "pyserini", rows)
    if json_output:
        print(json.dumps(rows, indent=2, sort_keys=True))
    else:
        sys.stdout.write(render_text_results(rows))
    return 0
