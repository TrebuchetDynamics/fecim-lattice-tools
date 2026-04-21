package validation

import "path/filepath"

func validationArtifactsRoot(repoRoot string) string {
	return filepath.Join(repoRoot, "validation", "testdata", "validation_artifacts")
}

func validationArtifactPath(repoRoot string, elems ...string) string {
	parts := append([]string{validationArtifactsRoot(repoRoot)}, elems...)
	return filepath.Join(parts...)
}

func validationArtifactGlob(repoRoot string, elems ...string) string {
	return validationArtifactPath(repoRoot, elems...)
}
