package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Violation is one direct UI mutation-like selector call from a goroutine that
// is not enclosed by fyne.Do/fyne.DoAndWait.
type Violation struct {
	Path    string
	Line    int
	Message string
}

// uiMutationMethods matches by method name only, with no receiver type
// information. Some shared/widgets/display types (OperationLog.Add,
// KeyStat.SetValue, ModeIndicator's Refresh path, ...) already wrap their
// own internal Fyne mutation in fyne.Do, making them safe to call from a
// goroutine directly - but this checker can't see across that boundary, so
// calls to those methods are still flagged. Verify each finding manually
// against the receiver's actual implementation before treating it as a
// real bug; a same-named method on a self-protecting wrapper widget is a
// known, expected false positive rather than a defect in the check itself.
var uiMutationMethods = map[string]bool{
	"Add": true, "AddObject": true, "Append": true,
	"Disable": true, "Enable": true,
	"Hide": true, "Move": true,
	"Refresh": true, "Remove": true, "RemoveAll": true, "Resize": true,
	"Select": true, "SetContent": true, "SetIcon": true, "SetImage": true,
	"SetPlaceHolder": true, "SetSelected": true, "SetText": true, "SetValue": true,
	"Show": true, "Unselect": true, "UnselectAll": true,
}

func main() {
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"cmd/fecim-lattice-tools", "module1-hysteresis/pkg/gui", "module2-crossbar/pkg/gui", "module3-mnist/pkg/gui", "module4-circuits/pkg/gui", "module5-comparison/pkg/gui", "module6-eda/pkg/gui", "module7-docs/pkg/gui", "shared/widgets"}
	}

	files, err := expandGoFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fyne-thread-check: %v\n", err)
		os.Exit(2)
	}
	violations, err := AnalyzeFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fyne-thread-check: %v\n", err)
		os.Exit(2)
	}
	for _, v := range violations {
		fmt.Printf("%s:%d: %s\n", v.Path, v.Line, v.Message)
	}
	if len(violations) > 0 {
		os.Exit(1)
	}
}

// funcInfo locates a top-level func or method declaration and the file it
// was parsed from, so a goroutine spawned by name (rather than an inline
// literal) can still be traced back to its body and reported at the right
// location.
type funcInfo struct {
	path string
	fset *token.FileSet
	decl *ast.FuncDecl
}

// AnalyzeFiles scans Go source files for direct mutation-like Fyne calls
// inside goroutines unless those calls are nested under fyne.Do/DoAndWait.
// This covers both inline literals (go func() { ... }()) and goroutines
// spawned by name (go x.method() or go someFunc()) — the latter is the more
// common pattern for long-running loops (animation, polling, inference) and
// was previously invisible to this checker entirely.
func AnalyzeFiles(paths []string) ([]Violation, error) {
	type parsedFile struct {
		path string
		fset *token.FileSet
		file *ast.File
	}

	var parsed []parsedFile
	funcsByName := map[string][]funcInfo{}

	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, parsedFile{path: path, fset: fset, file: file})
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
				funcsByName[fn.Name.Name] = append(funcsByName[fn.Name.Name], funcInfo{path: path, fset: fset, decl: fn})
			}
		}
	}

	seen := map[string]bool{}
	var out []Violation
	addViolations := func(vs []Violation) {
		for _, v := range vs {
			key := fmt.Sprintf("%s:%d:%s", v.Path, v.Line, v.Message)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, v)
		}
	}

	analyzedTargets := map[*ast.FuncDecl]bool{}
	analyzeTarget := func(fi funcInfo) {
		if analyzedTargets[fi.decl] {
			return
		}
		analyzedTargets[fi.decl] = true
		addViolations(analyzeGoroutineBody(fi.path, fi.fset, fi.decl.Body))
	}

	for _, pf := range parsed {
		for _, decl := range pf.file.Decls {
			enclosing, _ := decl.(*ast.FuncDecl)
			ast.Inspect(decl, func(n ast.Node) bool {
				goStmt, ok := n.(*ast.GoStmt)
				if !ok {
					return true
				}
				switch fn := goStmt.Call.Fun.(type) {
				case *ast.FuncLit:
					if fn.Body != nil {
						addViolations(analyzeGoroutineBody(pf.path, pf.fset, fn.Body))
					}
				case *ast.Ident:
					// Plain function call: go someFunc().
					for _, fi := range funcsByName[fn.Name] {
						if fi.decl.Recv == nil {
							analyzeTarget(fi)
						}
					}
				case *ast.SelectorExpr:
					// Method call: go x.method(). Multiple unrelated types in
					// this codebase share method names (e.g. App.run and
					// HysteresisDataLogger.run), so matching by name alone
					// would misattribute a synchronously-called method's body
					// to some other type's actual goroutine target. Resolve
					// x's static type where cheaply derivable (receiver
					// self-reference or a local `x := &Type{...}` literal in
					// the enclosing function) and match precisely; only fall
					// back to matching every same-named candidate when the
					// type can't be determined this way.
					recvIdent, _ := fn.X.(*ast.Ident)
					wantType := resolveReceiverType(enclosing, recvIdent)
					for _, fi := range funcsByName[fn.Sel.Name] {
						if fi.decl.Recv == nil {
							continue
						}
						if wantType != "" && receiverTypeName(fi.decl) != wantType {
							continue
						}
						analyzeTarget(fi)
					}
				}
				return false
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Line < out[j].Line
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

// receiverTypeName returns the declared receiver type name of a method
// (e.g. "App" for both `func (a App)` and `func (a *App)`), or "" for a
// plain function.
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	return typeNameOf(fn.Recv.List[0].Type)
}

func typeNameOf(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeNameOf(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

// resolveReceiverType best-effort resolves the static type of identifier x
// within enclosing: either x is enclosing's own receiver (the common
// self-call pattern `go a.method()` inside a method on a), or x is a local
// variable assigned a composite literal (`x := &Type{...}` / `x := Type{...}`)
// somewhere in enclosing's body. Returns "" when the type can't be
// determined this way, signaling callers to fall back to a conservative
// name-only match.
func resolveReceiverType(enclosing *ast.FuncDecl, x *ast.Ident) string {
	if enclosing == nil || enclosing.Body == nil || x == nil {
		return ""
	}
	if enclosing.Recv != nil && len(enclosing.Recv.List) > 0 {
		for _, name := range enclosing.Recv.List[0].Names {
			if name.Name == x.Name {
				return typeNameOf(enclosing.Recv.List[0].Type)
			}
		}
	}
	var found string
	ast.Inspect(enclosing.Body, func(n ast.Node) bool {
		if found != "" {
			return false
		}
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.DEFINE {
			return true
		}
		for i, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name != x.Name || i >= len(assign.Rhs) {
				continue
			}
			if t := compositeLitTypeName(assign.Rhs[i]); t != "" {
				found = t
			}
		}
		return true
	})
	return found
}

func compositeLitTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return compositeLitTypeName(e.X)
		}
	case *ast.CompositeLit:
		return typeNameOf(e.Type)
	}
	return ""
}

func analyzeGoroutineBody(path string, fset *token.FileSet, body *ast.BlockStmt) []Violation {
	var violations []Violation
	var inspect func(ast.Node, bool) bool
	inspect = func(n ast.Node, protected bool) bool {
		if n == nil {
			return true
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if isSafeUICall(call) {
				for _, arg := range call.Args {
					ast.Inspect(arg, func(child ast.Node) bool { return inspect(child, true) })
				}
				return false
			}
			if !protected {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && uiMutationMethods[sel.Sel.Name] && !isKnownNonUIReceiver(sel.X) {
					pos := fset.Position(sel.Pos())
					violations = append(violations, Violation{Path: path, Line: pos.Line, Message: fmt.Sprintf("%s inside goroutine needs fyne.Do/fyne.DoAndWait", selectorString(sel))})
				}
			}
		}
		return true
	}
	ast.Inspect(body, func(n ast.Node) bool { return inspect(n, false) })
	return violations
}

func isSafeUICall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if pkg.Name == "fyne" && (sel.Sel.Name == "Do" || sel.Sel.Name == "DoAndWait") {
		return true
	}
	return sel.Sel.Name == "SafeDo" || sel.Sel.Name == "SafeUIUpdate" || sel.Sel.Name == "SafeRefresh"
}

func isKnownNonUIReceiver(expr ast.Expr) bool {
	name := ""
	switch x := expr.(type) {
	case *ast.Ident:
		name = x.Name
	case *ast.SelectorExpr:
		name = x.Sel.Name
	}
	switch strings.ToLower(name) {
	case "wg", "ticker", "timer", "ctx", "cancel", "file", "writer", "reader", "mu", "mutex", "operationlog", "keystat":
		return true
	default:
		return false
	}
}

func selectorString(sel *ast.SelectorExpr) string {
	return exprString(sel.X) + "." + sel.Sel.Name
}

func exprString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return selectorString(x)
	default:
		return "<expr>"
	}
}

func expandGoFiles(paths []string) ([]string, error) {
	seen := map[string]bool{}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
				files = append(files, p)
			}
			continue
		}
		err = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == ".worktrees" || name == "artifacts" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") && !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}
