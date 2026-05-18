import io
import unittest
from contextlib import redirect_stdout

from fecim_research.cli import main


class CLITest(unittest.TestCase):
    def test_help_lists_core_commands(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["--help"])
        self.assertEqual(ctx.exception.code, 0)
        text = out.getvalue()
        self.assertIn("acquire", text)
        self.assertIn("audit", text)
        self.assertIn("cache", text)
        self.assertIn("cite", text)
        self.assertIn("claim-scan", text)
        self.assertIn("graph", text)
        self.assertIn("ingest", text)
        self.assertIn("index", text)
        self.assertIn("register-pdfs", text)
        self.assertIn("rebuild", text)
        self.assertIn("search", text)

    def test_acquire_help_lists_doi_option(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["acquire", "--help"])
        self.assertEqual(ctx.exception.code, 0)
        self.assertIn("--doi", out.getvalue())

    def test_cache_help_mentions_rebuildable_status_report(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["cache", "--help"])
        self.assertEqual(ctx.exception.code, 0)
        self.assertIn("rebuildable", out.getvalue())
        self.assertIn("--clean", out.getvalue())

    def test_rebuild_help_lists_skip_index_option(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["rebuild", "--help"])
        self.assertEqual(ctx.exception.code, 0)
        self.assertIn("--skip-index", out.getvalue())

    def test_search_help_lists_local_option(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["search", "--help"])
        self.assertEqual(ctx.exception.code, 0)
        self.assertIn("--local", out.getvalue())
        self.assertIn("--claim", out.getvalue())

    def test_register_pdfs_help_lists_write_stubs_option(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["register-pdfs", "--help"])
        self.assertEqual(ctx.exception.code, 0)
        self.assertIn("--write-stubs", out.getvalue())

    def test_unknown_command_fails(self):
        with self.assertRaises(SystemExit) as ctx:
            main(["unknown"])
        self.assertNotEqual(ctx.exception.code, 0)

    def test_core_commands_import_without_optional_dependencies(self):
        import fecim_research.acquisition
        import fecim_research.cache
        import fecim_research.cite
        import fecim_research.claims
        import fecim_research.claimscan
        import fecim_research.graphing
        import fecim_research.ingest
        import fecim_research.indexing
        import fecim_research.registration
        import fecim_research.rebuild
        import fecim_research.searching

        self.assertTrue(hasattr(fecim_research.acquisition, "run_acquire"))
        self.assertTrue(hasattr(fecim_research.cache, "run_cache"))
        self.assertTrue(hasattr(fecim_research.cite, "run_cite"))
        self.assertTrue(hasattr(fecim_research.claims, "run_audit"))
        self.assertTrue(hasattr(fecim_research.claimscan, "run_claim_scan"))
        self.assertTrue(hasattr(fecim_research.graphing, "run_graph"))
        self.assertTrue(hasattr(fecim_research.ingest, "run_ingest"))
        self.assertTrue(hasattr(fecim_research.indexing, "run_index"))
        self.assertTrue(hasattr(fecim_research.registration, "run_register_pdfs"))
        self.assertTrue(hasattr(fecim_research.rebuild, "run_rebuild"))
        self.assertTrue(hasattr(fecim_research.searching, "run_search"))


if __name__ == "__main__":
    unittest.main()
