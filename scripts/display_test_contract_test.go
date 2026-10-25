package scripts_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFullScreenshotSmokeRequiresDisplayBuildTag(t *testing.T) {
	root := repoRoot(t)
	for _, tags := range []string{"", "display"} {
		t.Run("tags="+tags, func(t *testing.T) {
			cmd := exec.Command("go", "list", "-json", "-tags="+tags, "./cmd/fecim-screenshotter-fyne")
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go list: %v: %s", err, out)
			}
			var pkg struct {
				Dir         string
				TestGoFiles []string
			}
			if err := json.Unmarshal(out, &pkg); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, name := range pkg.TestGoFiles {
				file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(pkg.Dir, name), nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "TestFullScreenshotterCapturesEveryModuleSequentially" {
						found = true
					}
				}
			}
			if found != (tags == "display") {
				t.Fatalf("full GUI smoke test selected=%v with tags=%q; must be opt-in with display", found, tags)
			}
		})
	}
}
