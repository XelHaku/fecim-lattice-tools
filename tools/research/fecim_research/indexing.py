import hashlib
import json
import shutil
import subprocess
import sys
from pathlib import Path

from .semantic import (
    EMBEDDING_PROVIDER,
    VECTOR_DIMENSION,
    build_vector_records,
    effective_embedding_model,
    write_vector_cache,
)


LATEST_INDEX_MANIFEST = "research/manifests/index-latest.json"
PYSERINI_INDEX_MANIFEST = "research/manifests/index-pyserini.json"
LANCEDB_INDEX_MANIFEST = "research/manifests/index-lancedb.json"


def collect_chunk_files(root: Path) -> list[Path]:
    chunk_dir = root / "research" / "chunks"
    if not chunk_dir.is_dir():
        return []
    return sorted(chunk_dir.glob("*.jsonl"))


def _sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def write_index_manifest(
    root: Path,
    backend: str,
    inputs: list[Path],
    semantic: bool,
    embedding_model: str,
    vector_dimension: int | None = None,
    embedding_provider: str = "",
    lancedb_index: str = "",
    external_ai: bool = False,
) -> Path:
    latest_path = root / LATEST_INDEX_MANIFEST
    backend_path = root / index_manifest_for_semantic(semantic)
    latest_path.parent.mkdir(parents=True, exist_ok=True)
    data = {
        "backend": backend,
        "semantic": semantic,
        "embedding_model": embedding_model,
        "embedding_provider": embedding_provider,
        "external_ai": external_ai,
        "inputs": [
            {
                "path": _repo_relative(root, path),
                "sha256": _sha(path),
            }
            for path in sorted(inputs)
        ],
        "pyserini_index": "research/index/pyserini",
    }
    if semantic:
        data["lancedb_index"] = lancedb_index or "research/index/lancedb"
        data["vector_dimension"] = vector_dimension or VECTOR_DIMENSION
    payload = json.dumps(data, indent=2, sort_keys=True) + "\n"
    backend_path.write_text(payload, encoding="utf-8")
    latest_path.write_text(payload, encoding="utf-8")
    return latest_path


def index_manifest_for_semantic(semantic: bool) -> str:
    return LANCEDB_INDEX_MANIFEST if semantic else PYSERINI_INDEX_MANIFEST


def run_index(root: Path, semantic: bool, embedding_model: str) -> int:
    if semantic:
        return _run_semantic_index(root, embedding_model)

    chunks = collect_chunk_files(root)
    if not chunks:
        print("no chunk files found under research/chunks", file=sys.stderr)
        return 1

    index_dir = root / "research" / "index" / "pyserini"
    if index_dir.exists():
        shutil.rmtree(index_dir)
    index_dir.parent.mkdir(parents=True, exist_ok=True)

    command = [
        sys.executable,
        "-m",
        "pyserini.index.lucene",
        "--collection",
        "JsonCollection",
        "--input",
        str(root / "research" / "chunks"),
        "--index",
        str(index_dir),
        "--generator",
        "DefaultLuceneDocumentGenerator",
        "--threads",
        "1",
        "--storePositions",
        "--storeDocvectors",
        "--storeRaw",
    ]
    result = subprocess.run(command, check=False)
    if result.returncode != 0:
        print(
            "Pyserini indexing failed; install pyserini and a compatible Java runtime, then rerun `fecim research index`.",
            file=sys.stderr,
        )
        return result.returncode

    write_index_manifest(root, "pyserini", chunks, semantic=False, embedding_model=embedding_model)
    print(f"indexed {len(chunks)} chunk file(s) into research/index/pyserini")
    return 0


def _run_semantic_index(root: Path, embedding_model: str) -> int:
    chunks = collect_chunk_files(root)
    if not chunks:
        print("no chunk files found under research/chunks", file=sys.stderr)
        return 1

    index_dir = root / "research" / "index" / "lancedb"
    if index_dir.exists():
        shutil.rmtree(index_dir)
    index_dir.mkdir(parents=True, exist_ok=True)

    model = effective_embedding_model(embedding_model)
    records = build_vector_records(root, chunks, model)
    write_vector_cache(root, records)
    backend = "local-vector-jsonl"

    try:
        import lancedb

        db = lancedb.connect(str(index_dir))
        db.create_table("chunks", data=records)
        backend = "lancedb"
    except Exception:
        backend = "local-vector-jsonl"

    write_index_manifest(
        root,
        backend,
        chunks,
        semantic=True,
        embedding_model=model,
        vector_dimension=VECTOR_DIMENSION,
        embedding_provider=EMBEDDING_PROVIDER,
        lancedb_index="research/index/lancedb",
        external_ai=False,
    )
    print(f"indexed {len(chunks)} chunk file(s) into research/index/lancedb backend={backend}")
    return 0
