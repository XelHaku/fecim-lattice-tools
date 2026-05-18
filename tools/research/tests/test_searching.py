import hashlib
import json
import sys
import tempfile
import types
import unittest
from contextlib import redirect_stderr, redirect_stdout
from io import StringIO
from pathlib import Path
from unittest.mock import patch

from fecim_research.searching import (
    _row,
    load_chunk_lookup,
    render_text_results,
    run_search,
    search_chunks_locally,
)


class SearchingTest(unittest.TestCase):
    def test_load_chunk_lookup_reads_jsonl_chunks(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps({"id": "park::sec-01::chunk-001", "paper_key": "park", "contents": "HZO coercive field evidence"}) + "\n",
                encoding="utf-8",
            )
            lookup = load_chunk_lookup(root)
            self.assertIn("park::sec-01::chunk-001", lookup)

    def test_render_text_results_includes_score_key_and_snippet(self):
        rows = [
            {
                "rank": 1,
                "score": 7.5,
                "docid": "park::sec-01::chunk-001",
                "paper_key": "park",
                "section": "Results",
                "snippet": "HZO coercive field evidence",
            }
        ]
        text = render_text_results(rows)
        self.assertIn("1. park", text)
        self.assertIn("score=7.5", text)
        self.assertIn("HZO coercive field evidence", text)

    def test_render_text_results_marks_inbox_results_unreviewed(self):
        rows = [
            {
                "rank": 1,
                "score": 7.5,
                "docid": "park::sec-01::chunk-001",
                "paper_key": "park",
                "section": "Results",
                "snippet": "HZO coercive field evidence",
                "trust_state": "unreviewed",
                "inbox": True,
            }
        ]
        text = render_text_results(rows)
        self.assertIn("UNREVIEWED", text)

    def test_run_search_rejects_stale_index_before_pyserini_import(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps({"id": "park::sec-01::chunk-001", "paper_key": "park", "contents": "new chunk contents"}) + "\n",
                encoding="utf-8",
            )
            (root / "research" / "index" / "pyserini").mkdir(parents=True)
            manifest = root / "research" / "manifests" / "index-latest.json"
            manifest.parent.mkdir(parents=True)
            manifest.write_text(
                json.dumps(
                    {
                        "backend": "pyserini",
                        "semantic": False,
                        "embedding_model": "",
                        "pyserini_index": "research/index/pyserini",
                        "inputs": [{"path": "research/chunks/park.jsonl", "sha256": "stale"}],
                    },
                    sort_keys=True,
                )
                + "\n",
                encoding="utf-8",
            )

            err = StringIO()
            with redirect_stderr(err):
                code = run_search(root, "HZO", 3, json_output=False)

            self.assertEqual(code, 1)
            self.assertIn("BM25 index is stale; rerun `fecim research index`", err.getvalue())

    def test_run_search_uses_pyserini_manifest_even_when_latest_manifest_is_semantic(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps({"id": "park::sec-01::chunk-001", "paper_key": "park", "contents": "HZO evidence"}) + "\n",
                encoding="utf-8",
            )
            from fecim_research.indexing import write_index_manifest

            write_index_manifest(root, "pyserini", [chunk], semantic=False, embedding_model="")
            latest = root / "research" / "manifests" / "index-latest.json"
            latest.write_text(
                json.dumps(
                    {
                        "backend": "local-vector-jsonl",
                        "semantic": True,
                        "embedding_model": "fecim-hashing-bow-v1",
                        "inputs": [{"path": "research/chunks/park.jsonl", "sha256": "stale-semantic"}],
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            (root / "research" / "index" / "pyserini").mkdir(parents=True)
            lucene = types.ModuleType("pyserini.search.lucene")

            class FakeHit:
                docid = "park::sec-01::chunk-001"
                score = 3.0

            class FakeSearcher:
                def __init__(self, index_dir: str):
                    self.index_dir = index_dir

                def search(self, query: str, k: int):
                    return [FakeHit()]

            lucene.LuceneSearcher = FakeSearcher

            out = StringIO()
            with patch.dict(
                sys.modules,
                {
                    "pyserini": types.ModuleType("pyserini"),
                    "pyserini.search": types.ModuleType("pyserini.search"),
                    "pyserini.search.lucene": lucene,
                },
            ), redirect_stdout(out):
                code = run_search(root, "HZO", 3, json_output=True)

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park::sec-01::chunk-001")

    def test_search_chunks_locally_ranks_jsonl_chunks_without_index(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "papers.jsonl"
            chunk.parent.mkdir(parents=True)
            records = [
                {
                    "id": "materlik::sec-01::chunk-001",
                    "paper_key": "materlik",
                    "section": "Background",
                    "contents": "HZO ferroelectric phase discussion.",
                },
                {
                    "id": "park::sec-02::chunk-001",
                    "paper_key": "park",
                    "section": "Results",
                    "contents": "HZO coercive field and Preisach switching evidence.",
                },
                {
                    "id": "control::sec-01::chunk-001",
                    "paper_key": "control",
                    "section": "Methods",
                    "contents": "Silicon capacitor baseline.",
                },
            ]
            chunk.write_text("\n".join(json.dumps(record) for record in records) + "\n", encoding="utf-8")

            rows = search_chunks_locally(root, "HZO coercive Preisach", 5)

            self.assertEqual([row["docid"] for row in rows], ["park::sec-02::chunk-001", "materlik::sec-01::chunk-001"])
            self.assertGreater(rows[0]["score"], rows[1]["score"])

    def test_run_search_local_writes_git_trackable_report_without_pyserini_index(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-02::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "HZO coercive field and Preisach switching evidence.",
                    }
                )
                + "\n",
                encoding="utf-8",
            )

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(root, "HZO coercive", 3, json_output=True, local=True)

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park::sec-02::chunk-001")
            report_path = root / "research" / "reports" / "search-latest.json"
            self.assertTrue(report_path.exists())
            report = json.loads(report_path.read_text(encoding="utf-8"))
            self.assertTrue(report["ok"])
            self.assertEqual(report["backend"], "local-jsonl")
            self.assertEqual(report["query"], "HZO coercive")
            self.assertEqual(report["results"][0]["docid"], "park::sec-02::chunk-001")

    def test_run_search_writes_content_addressed_history_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-02::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "HZO coercive field and Preisach switching evidence.",
                    }
                )
                + "\n",
                encoding="utf-8",
            )

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(root, "HZO coercive", 3, json_output=True, local=True)

            self.assertEqual(code, 0)
            latest_path = root / "research" / "reports" / "search-latest.json"
            latest = json.loads(latest_path.read_text(encoding="utf-8"))
            self.assertIn("run_id", latest)
            self.assertIn("history_path", latest)
            report_body = {key: value for key, value in latest.items() if key not in {"run_id", "history_path"}}
            expected_run_id = hashlib.sha256(
                json.dumps(report_body, sort_keys=True, separators=(",", ":")).encode("utf-8")
            ).hexdigest()[:16]
            self.assertEqual(latest["run_id"], expected_run_id)
            self.assertEqual(latest["history_path"], f"research/reports/searches/{latest['run_id']}.json")
            history_path = root / latest["history_path"]
            self.assertTrue(history_path.exists())
            self.assertEqual(json.loads(history_path.read_text(encoding="utf-8")), latest)

    def test_run_search_semantic_reads_rebuildable_vector_cache(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "papers.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-02::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "HZO coercive field and Preisach switching evidence.",
                        "source_path": "research/parsed/park/marker.md",
                    }
                )
                + "\n"
                + json.dumps(
                    {
                        "id": "control::sec-01::chunk-001",
                        "paper_key": "control",
                        "section": "Methods",
                        "contents": "Silicon capacitor baseline.",
                        "source_path": "research/parsed/control/marker.md",
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            from fecim_research.indexing import run_index

            with patch.dict(sys.modules, {"lancedb": None}):
                code = run_index(root, semantic=True, embedding_model="")
            self.assertEqual(code, 0)

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(root, "HZO coercive", 3, json_output=True, semantic=True)

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park::sec-02::chunk-001")
            report = json.loads((root / "research" / "reports" / "search-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(report["backend"], "local-vector-jsonl")
            self.assertTrue(report["semantic"])
            self.assertEqual(report["embedding_model"], "fecim-hashing-bow-v1")
            self.assertEqual(report["results"][0]["docid"], "park::sec-02::chunk-001")

    def test_run_search_semantic_uses_vector_manifest_even_when_latest_manifest_is_pyserini(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "papers.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-02::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "HZO coercive field and Preisach switching evidence.",
                        "source_path": "research/parsed/park/marker.md",
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            from fecim_research.indexing import run_index

            with patch.dict(sys.modules, {"lancedb": None}):
                code = run_index(root, semantic=True, embedding_model="")
            self.assertEqual(code, 0)
            latest = root / "research" / "manifests" / "index-latest.json"
            latest.write_text(
                json.dumps(
                    {
                        "backend": "pyserini",
                        "semantic": False,
                        "embedding_model": "",
                        "inputs": [{"path": "research/chunks/papers.jsonl", "sha256": "stale-pyserini"}],
                    }
                )
                + "\n",
                encoding="utf-8",
            )

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(root, "HZO coercive", 3, json_output=True, semantic=True)

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park::sec-02::chunk-001")

    def test_run_search_inbox_reads_unreviewed_sidecar_markdown_and_marks_results(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF-1.4")
            pdf.with_suffix(".md").write_text(
                "## Results\n\nLocal inbox HZO coercive field evidence.",
                encoding="utf-8",
            )
            inbox_report = root / "research" / "reports" / "local-inbox-pdfs.json"
            inbox_report.parent.mkdir(parents=True, exist_ok=True)
            inbox_report.write_text(
                json.dumps(
                    {
                        "ok": True,
                        "local_only": [
                            {
                                "action": "promote_pdf_first",
                                "citation_path": "citations/papers/park2015_advmat_hzo.md",
                                "paper_key": "park2015_advmat_hzo",
                                "path": "research/papers/park2015_advmat_hzo.pdf",
                                "sha256": "abc123",
                                "size": 8,
                                "status": "needs_promotion",
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(root, "HZO coercive", 3, json_output=True, inbox=True)

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park2015_advmat_hzo::sec-01::chunk-001")
            self.assertEqual(printed[0]["source_path"], "research/papers/park2015_advmat_hzo.md")
            self.assertEqual(printed[0]["pdf_path"], "research/papers/park2015_advmat_hzo.pdf")
            self.assertEqual(printed[0]["trust_state"], "unreviewed")
            self.assertTrue(printed[0]["review_required"])
            self.assertTrue(printed[0]["inbox"])
            report = json.loads((root / "research" / "reports" / "search-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(report["backend"], "inbox-local-jsonl")
            self.assertEqual(report["trust_state"], "unreviewed")
            self.assertTrue(report["review_required"])
            self.assertTrue(report["results"][0]["inbox"])

    def test_run_search_rejects_claim_linked_inbox_search(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_claim(root, "hzo-remanent-polarization-range")

            err = StringIO()
            with redirect_stderr(err):
                code = run_search(
                    root,
                    "",
                    3,
                    json_output=False,
                    inbox=True,
                    claim_id="hzo-remanent-polarization-range",
                )

            self.assertEqual(code, 1)
            self.assertIn("inbox search cannot be claim-linked", err.getvalue())
            self.assertFalse((root / "research" / "reports" / "search-latest.json").exists())

    def test_run_search_claim_uses_claim_text_and_records_claim_context(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_claim(root, "hzo-remanent-polarization-range")
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-02::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "Park reports Si-doped HZO remanent polarization near 24 uC/cm2.",
                    }
                )
                + "\n",
                encoding="utf-8",
            )

            out = StringIO()
            with redirect_stdout(out):
                code = run_search(
                    root,
                    "",
                    3,
                    json_output=True,
                    local=True,
                    claim_id="hzo-remanent-polarization-range",
                )

            self.assertEqual(code, 0)
            printed = json.loads(out.getvalue())
            self.assertEqual(printed[0]["docid"], "park::sec-02::chunk-001")
            report = json.loads((root / "research" / "reports" / "search-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(report["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(report["claim"]["status"], "literature-backed")
            self.assertEqual(report["claim"]["sources"], ["park2015_advmat_hzo"])
            self.assertEqual(report["query"], "Park reports Si-doped HZO remanent polarization near 24 uC/cm2.")

    def test_run_search_claim_returns_nonzero_for_unknown_claim(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)

            err = StringIO()
            with redirect_stderr(err):
                code = run_search(root, "", 3, json_output=False, local=True, claim_id="missing-claim")

            self.assertEqual(code, 1)
            self.assertIn("unknown claim id missing-claim", err.getvalue())
            self.assertFalse((root / "research" / "reports" / "search-latest.json").exists())

    def test_row_copies_source_and_span_fields(self):
        record = {
            "paper_key": "park",
            "contents": "HZO coercive field evidence with enough detail for audit linkage.",
            "source_parser": "marker",
            "source_path": "research/parsed/park/marker.md",
            "section": "Results",
            "section_number": 2,
            "chunk_number": 3,
            "page_start": 4,
            "page_end": 5,
            "char_start": 100,
            "char_end": 220,
            "sha256": "abc123",
        }

        row = _row(1, 7.5, "park::sec-02::chunk-003", record)

        self.assertEqual(row["contents"], "HZO coercive field evidence with enough detail for audit linkage.")
        self.assertEqual(row["source_path"], "research/parsed/park/marker.md")
        self.assertEqual(row["section_number"], 2)
        self.assertEqual(row["chunk_number"], 3)
        self.assertEqual(row["page_start"], 4)
        self.assertEqual(row["page_end"], 5)
        self.assertEqual(row["char_start"], 100)
        self.assertEqual(row["char_end"], 220)
        self.assertEqual(row["sha256"], "abc123")

    def _write_claim(self, root: Path, claim_id: str) -> None:
        path = root / "citations" / "claims" / f"{claim_id}.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"id: {claim_id}\n"
            "claim: Park reports Si-doped HZO remanent polarization near 24 uC/cm2.\n"
            "status: literature-backed\n"
            "sources:\n"
            "  - park2015_advmat_hzo\n"
            "used_in:\n"
            "  - docs/TRUST.md\n"
            "confidence: medium\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
