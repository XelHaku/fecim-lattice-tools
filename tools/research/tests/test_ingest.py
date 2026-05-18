import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from fecim_research.ingest import run_ingest
from fecim_research.parsing import run_grobid_if_available


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

    def test_ingest_reports_duplicates_without_reprocessing(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")

            papers = root / "research" / "papers"
            papers.mkdir(parents=True)
            canonical = papers / "park2015_advmat_hzo.pdf"
            duplicate = papers / "park2015_advmat_hzo_copy.pdf"
            canonical.write_bytes(b"%PDF same")
            duplicate.write_bytes(b"%PDF same")
            canonical.with_suffix(".md").write_text("## Results\n\nCanonical evidence.", encoding="utf-8")
            duplicate.with_suffix(".md").write_text("## Results\n\nDuplicate evidence.", encoding="utf-8")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            duplicate_report = json.loads(
                (root / "research" / "reports" / "duplicate-pdfs.json").read_text(encoding="utf-8")
            )
            self.assertEqual(
                duplicate_report,
                {
                    "duplicates": [
                        {
                            "duplicate_of": "research/papers/park2015_advmat_hzo.pdf",
                            "path": "research/papers/park2015_advmat_hzo_copy.pdf",
                            "sha256": duplicate_report["duplicates"][0]["sha256"],
                            "size": 9,
                            "status": "duplicate",
                        }
                    ]
                },
            )
            manifest = json.loads((root / "research" / "manifests" / "ingest-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(manifest["processed"], 1)
            self.assertEqual(manifest["duplicates"], 1)

            chunk_text = (root / "research" / "chunks" / "park2015_advmat_hzo.jsonl").read_text(encoding="utf-8")
            self.assertIn("Canonical evidence", chunk_text)
            self.assertNotIn("Duplicate evidence", chunk_text)

    def test_duplicate_canonical_prefers_matched_pdf_over_first_sorted_path(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")

            papers = root / "research" / "papers"
            papers.mkdir(parents=True)
            unmatched = papers / "a_scan.pdf"
            matched = papers / "park2015_advmat_hzo.pdf"
            unmatched.write_bytes(b"%PDF same scan")
            matched.write_bytes(b"%PDF same scan")
            matched.with_suffix(".md").write_text("## Results\n\nMatched sidecar evidence.", encoding="utf-8")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            manifest = json.loads((root / "research" / "manifests" / "ingest-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(manifest["processed"], 1)
            self.assertEqual(manifest["unmatched"], 0)
            self.assertEqual(manifest["duplicates"], 1)

            chunk_text = (root / "research" / "chunks" / "park2015_advmat_hzo.jsonl").read_text(encoding="utf-8")
            self.assertIn("Matched sidecar evidence", chunk_text)

            duplicate_report = json.loads(
                (root / "research" / "reports" / "duplicate-pdfs.json").read_text(encoding="utf-8")
            )
            self.assertEqual(len(duplicate_report["duplicates"]), 1)
            self.assertEqual(duplicate_report["duplicates"][0]["path"], "research/papers/a_scan.pdf")
            self.assertEqual(
                duplicate_report["duplicates"][0]["duplicate_of"],
                "research/papers/park2015_advmat_hzo.pdf",
            )

            unmatched_report = json.loads(
                (root / "research" / "reports" / "unmatched-pdfs.json").read_text(encoding="utf-8")
            )
            self.assertEqual(unmatched_report["unmatched"], [])

    def test_parse_manifest_uses_repo_relative_paths_and_messages(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")

            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")
            pdf.with_suffix(".md").write_text("## Results\n\nHZO evidence.", encoding="utf-8")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            manifest_text = (root / "research" / "parsed" / "park2015_advmat_hzo" / "manifest.json").read_text(
                encoding="utf-8"
            )
            self.assertNotIn(td, manifest_text)
            manifest = json.loads(manifest_text)
            output_paths = [result["output_path"] for result in manifest["results"]]
            self.assertIn("research/parsed/park2015_advmat_hzo/marker.md", output_paths)

    def test_existing_marker_reuse_is_explicit(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")

            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")
            marker = root / "research" / "parsed" / "park2015_advmat_hzo" / "marker.md"
            marker.parent.mkdir(parents=True)
            marker.write_text("## Prior Parse\n\nReusable evidence.", encoding="utf-8")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            manifest = json.loads(
                (root / "research" / "parsed" / "park2015_advmat_hzo" / "manifest.json").read_text(encoding="utf-8")
            )
            self.assertIn(
                {
                    "message": "reused existing parsed markdown",
                    "output_path": "research/parsed/park2015_advmat_hzo/marker.md",
                    "paper_key": "park2015_advmat_hzo",
                    "parser": "existing_markdown",
                    "status": "ok",
                },
                manifest["results"],
            )
            self.assertTrue((root / "research" / "chunks" / "park2015_advmat_hzo.jsonl").exists())

    def test_grobid_multipart_filename_is_sanitized(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / 'bad"name\n.pdf'
            pdf.write_bytes(b"%PDF fixture")
            captured = {}

            class Response:
                def __enter__(self):
                    return self

                def __exit__(self, exc_type, exc, traceback):
                    return False

                def read(self):
                    return b"<tei/>"

            def fake_urlopen(request, timeout):
                captured["body"] = request.data
                captured["timeout"] = timeout
                return Response()

            with mock.patch.dict(os.environ, {"FECIM_GROBID_URL": "http://grobid.example"}, clear=False):
                with mock.patch("urllib.request.urlopen", fake_urlopen):
                    result = run_grobid_if_available(pdf, root / "out.tei.xml")

            self.assertEqual(result.status, "ok")
            body = captured["body"].decode("utf-8", errors="replace")
            disposition = body.split("\r\n", 2)[1]
            self.assertIn('filename="bad_name_.pdf"', disposition)
            self.assertNotIn('"name', disposition)
            self.assertNotIn("\n", disposition)
            self.assertEqual(captured["timeout"], 20)


if __name__ == "__main__":
    unittest.main()
