import re
import unittest
from pathlib import Path


class MakefileResearchCITest(unittest.TestCase):
    def test_ci_target_runs_research_tests_and_audit(self):
        makefile = Path("Makefile").read_text(encoding="utf-8")

        ci_deps = self._target_dependencies(makefile, "ci")

        self.assertIn("test-research", ci_deps)
        self.assertIn("research-audit", ci_deps)

    def test_help_lists_research_test_target(self):
        makefile = Path("Makefile").read_text(encoding="utf-8")

        self.assertIn("make test-research", makefile)

    def _target_dependencies(self, makefile: str, target: str) -> list[str]:
        match = re.search(rf"^{re.escape(target)}:\s*(.*)$", makefile, flags=re.MULTILINE)
        self.assertIsNotNone(match, f"missing {target} target")
        return match.group(1).split()


if __name__ == "__main__":
    unittest.main()
