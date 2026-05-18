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
        self.assertIn("ingest", text)
        self.assertIn("index", text)
        self.assertIn("search", text)

    def test_unknown_command_fails(self):
        with self.assertRaises(SystemExit) as ctx:
            main(["unknown"])
        self.assertNotEqual(ctx.exception.code, 0)


if __name__ == "__main__":
    unittest.main()
