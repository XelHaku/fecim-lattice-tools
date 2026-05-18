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

    cache = sub.add_parser(
        "cache",
        help="write rebuildable cache status report",
        description="Write a rebuildable cache status report.",
    )
    cache.add_argument("--clean", action="store_true", help="remove ignored rebuildable cache directories")

    cite = sub.add_parser("cite", help="build a git-trackable packet for one reviewed claim")
    cite.add_argument("claim_id", help="claim id from citations/claims")
    cite.add_argument("--json", action="store_true", help="emit JSON citation packet")

    claim_scan = sub.add_parser("claim-scan", help="report likely uncited scientific claims")
    claim_scan.add_argument("paths", nargs="*", help="files or directories to scan")
    claim_scan.add_argument("--fail-on-findings", action="store_true", help="return nonzero when findings exist")

    sub.add_parser("graph", help="build git-trackable citation and claim provenance graph")

    ingest = sub.add_parser("ingest", help="discover, parse, and chunk local papers")
    ingest.add_argument("paths", nargs="*", help="optional extra PDF roots")

    index = sub.add_parser("index", help="build rebuildable search indexes")
    index.add_argument("--semantic", action="store_true", help="build local semantic index")
    index.add_argument("--embedding-model", default="", help="local embedding model name")

    register = sub.add_parser("register-pdfs", help="report or create reviewed stubs for local PDFs")
    register.add_argument("paths", nargs="*", help="optional extra PDF roots")
    register.add_argument("--write-stubs", action="store_true", help="write needs-review citation stubs")

    rebuild = sub.add_parser("rebuild", help="run ingestion, indexing, audit, and graph rebuild stages")
    rebuild.add_argument("paths", nargs="*", help="optional extra PDF roots")
    rebuild.add_argument("--skip-index", action="store_true", help="skip rebuildable search index generation")
    rebuild.add_argument("--semantic", action="store_true", help="build local semantic index")
    rebuild.add_argument("--embedding-model", default="", help="local embedding model name")

    search = sub.add_parser("search", help="search evidence chunks")
    search.add_argument("query", help="search query")
    search.add_argument("--json", action="store_true", help="emit JSON results")
    search.add_argument("--local", action="store_true", help="search tracked JSONL chunks without a rebuildable BM25 cache")
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
    if args.command == "cache":
        from .cache import run_cache

        return run_cache(root=root, clean=args.clean)
    if args.command == "cite":
        from .cite import run_cite

        return run_cite(root=root, claim_id=args.claim_id, json_output=args.json)
    if args.command == "claim-scan":
        from .claimscan import run_claim_scan

        return run_claim_scan(root=root, paths=args.paths, fail_on_findings=args.fail_on_findings)
    if args.command == "graph":
        from .graphing import run_graph

        return run_graph(root=root)
    if args.command == "ingest":
        from .ingest import run_ingest

        return run_ingest(root=root, extra_paths=[Path(p) for p in args.paths])
    if args.command == "index":
        from .indexing import run_index

        return run_index(root=root, semantic=args.semantic, embedding_model=args.embedding_model)
    if args.command == "register-pdfs":
        from .registration import run_register_pdfs

        return run_register_pdfs(
            root=root,
            extra_paths=[Path(p) for p in args.paths],
            write_stubs=args.write_stubs,
        )
    if args.command == "rebuild":
        from .rebuild import run_rebuild

        return run_rebuild(
            root=root,
            extra_paths=[Path(p) for p in args.paths],
            semantic=args.semantic,
            embedding_model=args.embedding_model,
            skip_index=args.skip_index,
        )
    if args.command == "search":
        from .searching import run_search

        return run_search(root=root, query=args.query, limit=args.limit, json_output=args.json, local=args.local)
    parser.error(f"unknown command {args.command}")
    return 2
