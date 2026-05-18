import json
import tempfile
import unittest
from contextlib import redirect_stderr
from io import StringIO
from pathlib import Path

from fecim_research.searching import _row, load_chunk_lookup, render_text_results, run_search


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

    def test_row_copies_source_and_span_fields(self):
        record = {
            "paper_key": "park",
            "contents": "HZO coercive field evidence",
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

        self.assertEqual(row["source_path"], "research/parsed/park/marker.md")
        self.assertEqual(row["section_number"], 2)
        self.assertEqual(row["chunk_number"], 3)
        self.assertEqual(row["page_start"], 4)
        self.assertEqual(row["page_end"], 5)
        self.assertEqual(row["char_start"], 100)
        self.assertEqual(row["char_end"], 220)
        self.assertEqual(row["sha256"], "abc123")


if __name__ == "__main__":
    unittest.main()
