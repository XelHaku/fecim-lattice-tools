import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.chunking import chunk_markdown, write_chunks_jsonl


class ChunkingTest(unittest.TestCase):
    def test_chunks_markdown_by_heading_and_size(self):
        text = "# Title\n\n## Results\n\nHZO coercive field text.\n\nMore remanent polarization text."
        chunks = chunk_markdown("park2015_advmat_hzo", text, max_chars=40)
        self.assertGreaterEqual(len(chunks), 2)
        self.assertEqual(chunks[0]["paper_key"], "park2015_advmat_hzo")
        self.assertIn("contents", chunks[0])
        self.assertTrue(chunks[0]["id"].startswith("park2015_advmat_hzo::sec-"))

    def test_write_chunks_jsonl_is_deterministic(self):
        chunks = chunk_markdown("park2015_advmat_hzo", "## Results\n\nHZO coercive field text.", max_chars=100)
        with tempfile.TemporaryDirectory() as td:
            path = Path(td) / "chunks.jsonl"
            write_chunks_jsonl(path, chunks)
            lines = path.read_text(encoding="utf-8").splitlines()
            self.assertEqual(len(lines), 1)
            record = json.loads(lines[0])
            self.assertEqual(record["paper_key"], "park2015_advmat_hzo")
            self.assertEqual(record["chunk_number"], 1)


if __name__ == "__main__":
    unittest.main()
