"""Guard the reproducible host used by the locked Playwright browser job."""

from pathlib import Path
import re
import unittest


ROOT = Path(__file__).resolve().parents[1]


class TestCIBrowserHost(unittest.TestCase):
    def setUp(self):
        workflow = (ROOT / ".github/workflows/ci.yml").read_text(encoding="utf-8")
        self.job = workflow.split("  ci:\n", 1)[1].split(
            "\n  external-validation:", 1
        )[0]

    def test_locked_browser_job_uses_supported_pinned_host(self):
        self.assertRegex(self.job, re.compile(r"^    runs-on: ubuntu-24\.04$", re.M))
        self.assertIn("npx playwright install --with-deps chromium firefox", self.job)

    def test_complete_browser_matrix_has_a_bounded_job_budget(self):
        match = re.search(r"^    timeout-minutes: (\d+)$", self.job, re.M)
        self.assertIsNotNone(match)
        self.assertEqual(int(match.group(1)), 60)

    def test_platform_guard_runs_in_the_browser_job(self):
        self.assertIn(
            "python3 -m unittest discover -s scripts -p 'test_ci_browser_host.py'",
            self.job,
        )


if __name__ == "__main__":
    unittest.main()
