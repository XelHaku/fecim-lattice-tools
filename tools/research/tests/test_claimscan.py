import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.claimscan import run_claim_scan, scan_claims


class ClaimScanTest(unittest.TestCase):
    def test_scan_flags_numeric_scientific_claims_without_claim_marker(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            doc = root / "docs" / "note.md"
            doc.parent.mkdir(parents=True)
            doc.write_text(
                "HZO Pr is 24 uC/cm2 in Park 2015.\n"
                "HZO Ec is 1.0 MV/cm. [claim: hzo-remanent-polarization-range]\n"
                "Published DOI 10.1002/adma.201404531 supports this.\n",
                encoding="utf-8",
            )

            report = scan_claims(root, ["docs"])

            self.assertEqual(report.scanned_files, 1)
            self.assertEqual(len(report.findings), 2)
            self.assertEqual(report.findings[0].path, "docs/note.md")
            self.assertEqual(report.findings[0].line, 1)
            self.assertEqual(report.findings[0].reason, "numeric-unit")
            self.assertEqual(report.findings[1].reason, "doi")

    def test_scan_ignores_code_fences_and_inline_ignore_marker(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            doc = root / "README.md"
            doc.write_text(
                "```yaml\n"
                "pr_uC_cm2: 24\n"
                "```\n"
                "HZO Pr is 24 uC/cm2. <!-- claim-scan: ignore -->\n",
                encoding="utf-8",
            )

            report = scan_claims(root, ["README.md"])

            self.assertEqual(report.findings, [])

    def test_scan_treats_indented_child_lines_under_claim_marker_as_covered(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            facts = root / "citations" / "facts.md"
            facts.parent.mkdir(parents=True)
            facts.write_text(
                "- **Pr** = `24` `uC/cm2` [claim: hzo-remanent-polarization-range]\n"
                "  - Conditions: 10 nm film at 1 kHz\n"
                "- **Ec** = `1.0` `MV/cm`\n",
                encoding="utf-8",
            )

            report = scan_claims(root, ["citations/facts.md"])

            self.assertEqual(len(report.findings), 1)
            self.assertEqual(report.findings[0].text, "- **Ec** = `1.0` `MV/cm`")

    def test_run_claim_scan_writes_git_trackable_report_and_stays_report_only(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            readme = root / "README.md"
            readme.write_text("HZO Pr is 24 uC/cm2.\n", encoding="utf-8")

            code = run_claim_scan(root, ["README.md"], fail_on_findings=False)

            self.assertEqual(code, 0)
            report_path = root / "research" / "reports" / "claim-scan-latest.json"
            self.assertTrue(report_path.exists())
            payload = json.loads(report_path.read_text())
            self.assertEqual(payload["findings_count"], 1)
            self.assertEqual(payload["findings"][0]["path"], "README.md")

    def test_run_claim_scan_can_fail_on_findings(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            readme = root / "README.md"
            readme.write_text("HZO Ec is 1.0 MV/cm.\n", encoding="utf-8")

            code = run_claim_scan(root, ["README.md"], fail_on_findings=True)

            self.assertEqual(code, 1)


if __name__ == "__main__":
    unittest.main()
