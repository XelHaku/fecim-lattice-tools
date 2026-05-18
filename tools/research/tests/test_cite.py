import io
import hashlib
import json
import tempfile
import unittest
from contextlib import redirect_stdout
from pathlib import Path

from fecim_research.cite import build_citation_packet, run_cite


class CiteTest(unittest.TestCase):
    def test_build_citation_packet_links_claim_sources_and_used_files(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_openalex_source(root, "doi_10_5555_new_paper")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["doi_10_5555_new_paper", "park2015_advmat_hzo"],
                used_in=["docs/TRUST.md"],
            )
            trust = root / "docs" / "TRUST.md"
            trust.parent.mkdir(parents=True)
            trust.write_text("[claim: hzo-remanent-polarization-range]\n", encoding="utf-8")

            packet = build_citation_packet(root, "hzo-remanent-polarization-range")

            self.assertEqual(packet["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(packet["claim"]["status"], "literature-backed")
            self.assertEqual(packet["claim"]["confidence"], "medium")
            self.assertEqual(
                [source["key"] for source in packet["sources"]],
                ["doi_10_5555_new_paper", "park2015_advmat_hzo"],
            )
            self.assertEqual(packet["sources"][0]["path"], "research/sources/doi_10_5555_new_paper.openalex.json")
            self.assertEqual(packet["sources"][0]["title"], "New open access FeCIM paper")
            self.assertEqual(packet["sources"][0]["openalex_id"], "https://openalex.org/W555")
            self.assertEqual(packet["sources"][1]["path"], "citations/papers/park2015_advmat_hzo.md")
            self.assertEqual(packet["sources"][1]["title"], "Park 2015")
            self.assertEqual(packet["missing_sources"], [])
            self.assertEqual(packet["used_in"][0]["path"], "docs/TRUST.md")
            self.assertTrue(packet["used_in"][0]["exists"])
            self.assertTrue(packet["used_in"][0]["references_claim"])

    def test_run_cite_writes_git_trackable_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["docs/TRUST.md"],
            )
            trust = root / "docs" / "TRUST.md"
            trust.parent.mkdir(parents=True)
            trust.write_text("[claim: hzo-remanent-polarization-range]\n", encoding="utf-8")

            out = io.StringIO()
            with redirect_stdout(out):
                code = run_cite(root, "hzo-remanent-polarization-range", json_output=True)

            self.assertEqual(code, 0)
            emitted = json.loads(out.getvalue())
            report_path = root / "research" / "reports" / "cite-latest.json"
            self.assertTrue(report_path.exists())
            report = json.loads(report_path.read_text())
            self.assertEqual(emitted["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(report["claim"]["id"], "hzo-remanent-polarization-range")
            self.assertEqual(report["sources"][0]["key"], "park2015_advmat_hzo")

    def test_run_cite_writes_content_addressed_history_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["docs/TRUST.md"],
            )
            trust = root / "docs" / "TRUST.md"
            trust.parent.mkdir(parents=True)
            trust.write_text("[claim: hzo-remanent-polarization-range]\n", encoding="utf-8")

            out = io.StringIO()
            with redirect_stdout(out):
                code = run_cite(root, "hzo-remanent-polarization-range", json_output=True)

            self.assertEqual(code, 0)
            latest_path = root / "research" / "reports" / "cite-latest.json"
            latest = json.loads(latest_path.read_text(encoding="utf-8"))
            self.assertIn("run_id", latest)
            self.assertIn("history_path", latest)
            report_body = {key: value for key, value in latest.items() if key not in {"run_id", "history_path"}}
            expected_run_id = hashlib.sha256(
                json.dumps(report_body, sort_keys=True, separators=(",", ":")).encode("utf-8")
            ).hexdigest()[:16]
            self.assertEqual(latest["run_id"], expected_run_id)
            self.assertEqual(latest["history_path"], f"research/reports/cites/{latest['run_id']}.json")
            history_path = root / latest["history_path"]
            self.assertTrue(history_path.exists())
            self.assertEqual(json.loads(history_path.read_text(encoding="utf-8")), latest)

    def test_run_cite_returns_nonzero_for_unknown_claim(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)

            out = io.StringIO()
            with redirect_stdout(out):
                code = run_cite(root, "missing-claim", json_output=False)

            self.assertEqual(code, 1)
            self.assertIn("unknown claim id missing-claim", out.getvalue())
            self.assertFalse((root / "research" / "reports" / "cite-latest.json").exists())

    def _write_paper(self, root: Path, key: str):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"**Key:** `{key}`\n"
            "**Title:** `Park 2015`\n"
            "**DOI:** `10.1002/adma.201404531`\n",
            encoding="utf-8",
        )

    def _write_openalex_source(self, root: Path, key: str):
        path = root / "research" / "sources" / f"{key}.openalex.json"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            json.dumps(
                {
                    "id": "https://openalex.org/W555",
                    "doi": "https://doi.org/10.5555/New.Paper",
                    "display_name": "New open access FeCIM paper",
                    "publication_year": 2026,
                }
            ),
            encoding="utf-8",
        )

    def _write_claim(self, root: Path, claim_id: str, sources: list[str], used_in: list[str]):
        path = root / "citations" / "claims" / f"{claim_id}.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"id: {claim_id}\n"
            "claim: HZO devices commonly report remanent polarization in a bounded literature range.\n"
            "status: literature-backed\n"
            "sources:\n"
            + "".join(f"  - {source}\n" for source in sources)
            + "used_in:\n"
            + "".join(f"  - {item}\n" for item in used_in)
            + "confidence: medium\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
