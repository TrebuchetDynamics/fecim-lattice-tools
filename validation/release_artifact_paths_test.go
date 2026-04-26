package validation

import "path/filepath"

func releaseArtifactPath(parts ...string) string {
	repoRoot := filepath.Clean("..")
	all := append([]string{repoRoot, "validation", "testdata", "release_artifacts"}, parts...)
	return filepath.Join(all...)
}
