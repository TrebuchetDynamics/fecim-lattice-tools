"""Keep policy-private paths ignored without hiding public source or fixtures."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

from check_public_release_boundary import DISALLOWED_PATHS


ROOT = Path(__file__).resolve().parents[2]


class TestPrivateIgnorePolicy(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.env = {
            key: value for key, value in os.environ.items()
            if not key.startswith("GIT_")
        }
        self.env.update(GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1")
        subprocess.run(
            ["git", "init", "--quiet", str(self.root)],
            env=self.env, check=True, capture_output=True,
        )
        shutil.copyfile(ROOT / ".gitignore", self.root / ".gitignore")

    def ignored(self, path):
        result = subprocess.run(
            ["git", "check-ignore", "--no-index", "--", path],
            cwd=self.root, env=self.env, capture_output=True, text=True,
        )
        self.assertIn(result.returncode, (0, 1), result.stderr)
        return result.returncode == 0

    def test_public_boundary_private_paths_are_ignored(self):
        for path in (*DISALLOWED_PATHS,
                     "docs/internals/release/FILTER_REPO_PATHS.txt"):
            with self.subTest(path=path):
                candidate = path + "local-notes.md" if path.endswith("/") else path
                self.assertTrue(self.ignored(candidate), candidate)

    def test_public_source_fixtures_and_templates_remain_visible(self):
        for path in (
            "README.md",
            "docs/guides/installation.md",
            "docs/internals/adr/0005-unified-react-hosted-and-local.md",
            "docs/internals/testing/all-modules-web-checkpoint.md",
            "docs/internals/superpowers-public/architecture.md",
            "docs/research/papers/approved-reference.pdf",
            "tasks/items/WEB-WORKBENCH.md",
            "shared/physics/nls.go",
            "shared/crossbar/array.go",
            "validation/testdata/physics_regression/example.json",
            "internal/browserui/assets/README.txt",
            "scripts/public-release/check_public_release_boundary.py",
            ".env.example",
            ".env.local.example",
        ):
            with self.subTest(path=path):
                self.assertFalse(self.ignored(path), path)


if __name__ == "__main__":
    unittest.main()
