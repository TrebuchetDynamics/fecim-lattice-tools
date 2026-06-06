package progress

import (
	"os/exec"
	"strings"
	"testing"
)

func TestDefaultProgressPackageImportsFyneWidgetSurface(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list shared/progress imports failed: %v\n%s", err, out)
	}
	if !containsImportPrefix(strings.Fields(string(out)), "fyne.io/fyne/v2") {
		t.Fatalf("shared/progress default build must import Fyne widget surface after Fyne rollback; imports=%s", out)
	}
}

func containsImportPrefix(imports []string, prefix string) bool {
	for _, imp := range imports {
		if imp == prefix || strings.HasPrefix(imp, prefix+"/") {
			return true
		}
	}
	return false
}
