import argparse
from pathlib import Path

from .paths import repo_root


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="fecim research")
    parser.add_argument("--repo-root", type=Path, default=None)
    sub = parser.add_subparsers(dest="command", required=True)

    acquire = sub.add_parser("acquire", help="plan or download legal OpenAlex OA PDFs for missing papers")
    acquire.add_argument("keys", nargs="*", help="optional citation keys to acquire")
    acquire.add_argument("--doi", dest="dois", action="append", default=[], help="acquire a new paper by DOI")
    acquire.add_argument("--download", action="store_true", help="download OA PDFs into ignored research/papers")

    sub.add_parser("audit", help="validate reviewed claim registry and claim references")

    ingest = sub.add_parser("ingest", help="discover, parse, and chunk local papers")
    ingest.add_argument("paths", nargs="*", help="optional extra PDF roots")

    index = sub.add_parser("index", help="build rebuildable search indexes")
    index.add_argument("--semantic", action="store_true", help="build local semantic index")
    index.add_argument("--embedding-model", default="", help="local embedding model name")

    search = sub.add_parser("search", help="search evidence chunks")
    search.add_argument("query", help="search query")
    search.add_argument("--json", action="store_true", help="emit JSON results")
    search.add_argument("--limit", type=int, default=10)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    root = args.repo_root.resolve() if args.repo_root else repo_root()

    if args.command == "acquire":
        from .acquisition import run_acquire

        return run_acquire(root=root, keys=args.keys, dois=args.dois, download=args.download)
    if args.command == "audit":
        from .claims import run_audit

        return run_audit(root=root)
    if args.command == "ingest":
        from .ingest import run_ingest

        return run_ingest(root=root, extra_paths=[Path(p) for p in args.paths])
    if args.command == "index":
        from .indexing import run_index

        return run_index(root=root, semantic=args.semantic, embedding_model=args.embedding_model)
    if args.command == "search":
        from .searching import run_search

        return run_search(root=root, query=args.query, limit=args.limit, json_output=args.json)
    parser.error(f"unknown command {args.command}")
    return 2
