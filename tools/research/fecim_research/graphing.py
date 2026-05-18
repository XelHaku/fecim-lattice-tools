from __future__ import annotations

from pathlib import Path
import json

from .claims import CLAIM_REF_RE, load_claim_records
from .citations import load_citation_records
from .reporting import write_content_addressed_report


def run_graph(root: Path) -> int:
    graph = build_provenance_graph(root)
    report = _graph_report(graph)
    addressed_graph = write_content_addressed_report(
        root,
        "research/graphs/provenance-graph.json",
        "research/graphs/history",
        graph,
    )
    report["graph_run_id"] = str(addressed_graph["run_id"])
    report["graph_history_path"] = str(addressed_graph["history_path"])
    report_path = root / "research" / "reports" / "graph-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"research graph complete: nodes={report['nodes']} edges={report['edges']} claims={report['claims']}")
    return 0


def build_provenance_graph(root: Path) -> dict[str, list[dict[str, object]]]:
    nodes: dict[str, dict[str, object]] = {}
    edges: dict[tuple[str, str, str], dict[str, str]] = {}
    citations = load_citation_records(root)
    claims = load_claim_records(root)

    for key, record in sorted(citations.items()):
        _add_node(
            nodes,
            {
                "id": f"paper:{key}",
                "type": "paper",
                "key": key,
                "title": record.title,
                "path": _rel(root, record.path),
            },
        )

    for claim_id, record in sorted(claims.items()):
        claim_node = f"claim:{claim_id}"
        _add_node(
            nodes,
            {
                "id": claim_node,
                "type": "claim",
                "key": claim_id,
                "status": record.status,
                "confidence": record.confidence,
                "claim": record.claim,
                "path": _rel(root, record.path),
            },
        )
        for source in record.sources:
            source_node = f"paper:{source}"
            if source_node not in nodes:
                _add_node(nodes, {"id": source_node, "type": "source", "key": source})
            _add_edge(edges, source_node, claim_node, "supports")
        for used_path in record.used_in:
            file_node = f"file:{used_path}"
            _add_node(nodes, {"id": file_node, "type": "file", "path": used_path})
            _add_edge(edges, claim_node, file_node, "used_in")

    for path in _reference_files(root):
        file_node = f"file:{_rel(root, path)}"
        refs = _claim_refs(path)
        if not refs:
            continue
        _add_node(nodes, {"id": file_node, "type": "file", "path": _rel(root, path)})
        for claim_id in refs:
            claim_node = f"claim:{claim_id}"
            if claim_node not in nodes:
                _add_node(nodes, {"id": claim_node, "type": "claim", "key": claim_id})
            _add_edge(edges, file_node, claim_node, "references")

    return {
        "nodes": [nodes[key] for key in sorted(nodes)],
        "edges": [edges[key] for key in sorted(edges)],
    }


def _reference_files(root: Path) -> list[Path]:
    candidates: list[Path] = [
        root / "citations" / "facts.md",
        root / "citations" / "disputed.md",
        root / "docs" / "TRUST.md",
    ]
    config = root / "config"
    if config.exists():
        candidates.extend(sorted(config.glob("*.yaml")))
    return [path for path in candidates if path.exists()]


def _claim_refs(path: Path) -> list[str]:
    return sorted(set(CLAIM_REF_RE.findall(path.read_text(encoding="utf-8", errors="replace"))))


def _add_node(nodes: dict[str, dict[str, object]], node: dict[str, object]) -> None:
    existing = nodes.get(str(node["id"]))
    if existing is None:
        nodes[str(node["id"])] = node
        return
    existing.update({key: value for key, value in node.items() if value not in {"", None}})


def _add_edge(edges: dict[tuple[str, str, str], dict[str, str]], from_node: str, to_node: str, relation: str) -> None:
    edges[(from_node, to_node, relation)] = {
        "from": from_node,
        "to": to_node,
        "relation": relation,
    }


def _graph_report(graph: dict[str, list[dict[str, object]]]) -> dict[str, int]:
    nodes = graph["nodes"]
    return {
        "nodes": len(nodes),
        "edges": len(graph["edges"]),
        "claims": sum(1 for node in nodes if node.get("type") == "claim"),
        "sources": sum(1 for node in nodes if node.get("type") in {"paper", "source"}),
        "files": sum(1 for node in nodes if node.get("type") == "file"),
    }


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
