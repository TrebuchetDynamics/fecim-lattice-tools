package physics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAllConfigYAMLParse validates that every YAML under config/ parses cleanly.
func TestAllConfigYAMLParse(t *testing.T) {
	// Go tests run in the package directory. Source paths from runtime.Caller
	// are module-relative under -trimpath and cannot locate on-disk fixtures.
	configDir := ".."

	err := filepath.WalkDir(configDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".yaml") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}

		var v any
		if unmarshalErr := yaml.Unmarshal(data, &v); unmarshalErr != nil {
			t.Fatalf("parse %s: %v", path, unmarshalErr)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk config dir %s: %v", configDir, err)
	}
}
