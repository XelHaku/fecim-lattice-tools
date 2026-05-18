import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.rebuild import RebuildStages, run_rebuild


class RebuildTest(unittest.TestCase):
    def test_run_rebuild_runs_stages_in_order_and_writes_trackable_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            calls: list[str] = []

            stages = RebuildStages(
                ingest=lambda: self._stage(calls, "ingest", 0),
                index=lambda: self._stage(calls, "index", 0),
                audit=lambda: self._stage(calls, "audit", 0),
                graph=lambda: self._stage(calls, "graph", 0),
            )

            code = run_rebuild(
                root=root,
                extra_paths=[Path("seed-pdfs")],
                semantic=False,
                embedding_model="",
                skip_index=False,
                stages=stages,
            )

            self.assertEqual(code, 0)
            self.assertEqual(calls, ["ingest", "index", "audit", "graph"])
            report = json.loads((root / "research" / "reports" / "rebuild-latest.json").read_text())
            self.assertTrue(report["ok"])
            self.assertEqual(report["extra_paths"], ["seed-pdfs"])
            self.assertEqual([stage["stage"] for stage in report["stages"]], ["ingest", "index", "audit", "graph"])
            self.assertEqual([stage["status"] for stage in report["stages"]], ["ok", "ok", "ok", "ok"])
            self.assertIn("research/manifests/ingest-latest.json", report["stages"][0]["artifacts"])
            self.assertIn("research/manifests/index-latest.json", report["stages"][1]["artifacts"])
            self.assertIn("research/reports/claim-audit-latest.json", report["stages"][2]["artifacts"])
            self.assertIn("research/graphs/provenance-graph.json", report["stages"][3]["artifacts"])

    def test_run_rebuild_can_skip_index_for_file_only_audits(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            calls: list[str] = []

            stages = RebuildStages(
                ingest=lambda: self._stage(calls, "ingest", 0),
                index=lambda: self._stage(calls, "index", 99),
                audit=lambda: self._stage(calls, "audit", 0),
                graph=lambda: self._stage(calls, "graph", 0),
            )

            code = run_rebuild(
                root=root,
                extra_paths=[],
                semantic=False,
                embedding_model="",
                skip_index=True,
                stages=stages,
            )

            self.assertEqual(code, 0)
            self.assertEqual(calls, ["ingest", "audit", "graph"])
            report = json.loads((root / "research" / "reports" / "rebuild-latest.json").read_text())
            self.assertTrue(report["ok"])
            self.assertEqual([stage["stage"] for stage in report["stages"]], ["ingest", "index", "audit", "graph"])
            self.assertEqual(report["stages"][1]["status"], "skipped")
            self.assertEqual(report["stages"][1]["exit_code"], 0)

    def test_run_rebuild_reports_failures_but_continues_later_file_stages(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            calls: list[str] = []

            stages = RebuildStages(
                ingest=lambda: self._stage(calls, "ingest", 0),
                index=lambda: self._stage(calls, "index", 7),
                audit=lambda: self._stage(calls, "audit", 0),
                graph=lambda: self._stage(calls, "graph", 0),
            )

            code = run_rebuild(
                root=root,
                extra_paths=[],
                semantic=False,
                embedding_model="",
                skip_index=False,
                stages=stages,
            )

            self.assertEqual(code, 7)
            self.assertEqual(calls, ["ingest", "index", "audit", "graph"])
            report = json.loads((root / "research" / "reports" / "rebuild-latest.json").read_text())
            self.assertFalse(report["ok"])
            self.assertEqual(report["stages"][1]["status"], "failed")
            self.assertEqual(report["stages"][1]["exit_code"], 7)
            self.assertEqual(report["failed"], 1)

    def _stage(self, calls: list[str], name: str, code: int) -> int:
        calls.append(name)
        return code


if __name__ == "__main__":
    unittest.main()
