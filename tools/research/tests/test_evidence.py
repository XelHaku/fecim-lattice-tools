import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.evidence import build_evidence_record, run_evidence


class EvidenceTest(unittest.TestCase):
    def test_build_evidence_record_preserves_claim_and_candidate_chunks(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_search_report(root, claim_id="hzo-remanent-polarization-range")

            record = build_evidence_record(root, "hzo-remanent-polarization-range")

            self.assertIsNotNone(record)
            assert record is not None
            self.assertEqual(record["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(record["status"], "candidate-evidence")
            self.assertEqual(record["review"]["state"], "needs-review")
            self.assertEqual(record["source_report"], "research/reports/search-latest.json")
            self.assertEqual(record["candidate_count"], 1)
            self.assertEqual(record["candidates"][0]["docid"], "park::sec-02::chunk-001")
            self.assertEqual(record["candidates"][0]["paper_key"], "park")

    def test_run_evidence_writes_git_trackable_evidence_and_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_search_report(root, claim_id="hzo-remanent-polarization-range")

            code = run_evidence(root, "hzo-remanent-polarization-range")

            self.assertEqual(code, 0)
            evidence_path = root / "research" / "evidence" / "hzo-remanent-polarization-range.json"
            report_path = root / "research" / "reports" / "evidence-latest.json"
            self.assertTrue(evidence_path.exists())
            self.assertTrue(report_path.exists())
            evidence = json.loads(evidence_path.read_text(encoding="utf-8"))
            report = json.loads(report_path.read_text(encoding="utf-8"))
            self.assertEqual(evidence["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(report["output"], "research/evidence/hzo-remanent-polarization-range.json")
            self.assertEqual(report["candidate_count"], 1)

    def test_run_evidence_rejects_search_report_for_different_claim(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_search_report(root, claim_id="other-claim")

            code = run_evidence(root, "hzo-remanent-polarization-range")

            self.assertEqual(code, 1)
            self.assertFalse((root / "research" / "evidence" / "hzo-remanent-polarization-range.json").exists())

    def test_run_evidence_requires_claim_linked_search_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            report = root / "research" / "reports" / "search-latest.json"
            report.parent.mkdir(parents=True, exist_ok=True)
            report.write_text(
                json.dumps({"ok": True, "backend": "local-jsonl", "query": "HZO", "results": []}),
                encoding="utf-8",
            )

            code = run_evidence(root, "hzo-remanent-polarization-range")

            self.assertEqual(code, 1)

    def _write_search_report(self, root: Path, claim_id: str) -> None:
        report = root / "research" / "reports" / "search-latest.json"
        report.parent.mkdir(parents=True, exist_ok=True)
        report.write_text(
            json.dumps(
                {
                    "ok": True,
                    "backend": "local-jsonl",
                    "query": "Park reports HZO remanent polarization.",
                    "claim": {
                        "id": claim_id,
                        "claim": "Park reports HZO remanent polarization.",
                        "status": "literature-backed",
                        "confidence": "medium",
                        "sources": ["park2015_advmat_hzo"],
                        "path": f"citations/claims/{claim_id}.yaml",
                    },
                    "result_count": 1,
                    "results": [
                        {
                            "rank": 1,
                            "score": 31.0,
                            "docid": "park::sec-02::chunk-001",
                            "paper_key": "park",
                            "section": "Results",
                            "snippet": "HZO remanent polarization evidence.",
                            "chunk_file": "research/chunks/park.jsonl",
                            "source_path": "research/parsed/park/marker.md",
                            "sha256": "abc123",
                        }
                    ],
                },
                sort_keys=True,
            )
            + "\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
