import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.ingest import run_ingest


class IngestTest(unittest.TestCase):
    def test_ingest_writes_source_chunk_manifest_and_unmatched_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n**DOI:** `10.1002/adma.201404531`\n", encoding="utf-8")

            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")
            pdf.with_suffix(".md").write_text("## Results\n\nHZO coercive field evidence.", encoding="utf-8")

            unknown = root / "research" / "papers" / "unknown.pdf"
            unknown.write_bytes(b"%PDF unknown")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            self.assertTrue((root / "research" / "sources" / "park2015_advmat_hzo.yaml").exists())
            self.assertTrue((root / "research" / "parsed" / "park2015_advmat_hzo" / "marker.md").exists())
            self.assertTrue((root / "research" / "chunks" / "park2015_advmat_hzo.jsonl").exists())
            report = json.loads((root / "research" / "reports" / "unmatched-pdfs.json").read_text(encoding="utf-8"))
            self.assertEqual(len(report["unmatched"]), 1)
            self.assertIn("unknown.pdf", report["unmatched"][0]["path"])


if __name__ == "__main__":
    unittest.main()
