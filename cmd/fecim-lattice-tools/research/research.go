// Package research adapts the fecim-lattice-tools Go command to the Python
// research CLI under tools/research.
package research

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RunTool locates the repository research CLI and runs it with args.
func RunTool(args []string) error {
	python := os.Getenv("FECIM_RESEARCH_PYTHON")
	if python == "" {
		python = "python3"
	}
	root, err := RepoRoot(args)
	if err != nil {
		return err
	}
	script := filepath.Join(root, "tools", "research", "research_cli.py")
	cmdArgs := append([]string{script}, NormalizeRepoRootArg(args, root)...)
	cmd := exec.Command(python, cmdArgs...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("research tool: %w", err)
	}
	return nil
}

// RepoRoot returns the repository root for the research CLI.
func RepoRoot(args []string) (string, error) {
	if root := RepoRootFromArgs(args); root != "" {
		return ValidateRepoRoot(root)
	}
	if cwd, err := os.Getwd(); err == nil {
		if root, ok := FindRepoRoot(cwd); ok {
			return root, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		if root, ok := FindRepoRoot(filepath.Dir(exe)); ok {
			return root, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if root, ok := FindRepoRoot(filepath.Dir(file)); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("research tool: could not locate repository root containing tools/research/research_cli.py")
}

// RepoRootFromArgs extracts --repo-root from args, if present.
func RepoRootFromArgs(args []string) string {
	for i, arg := range args {
		if arg == "--repo-root" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--repo-root=") {
			return strings.TrimPrefix(arg, "--repo-root=")
		}
	}
	return ""
}

// NormalizeRepoRootArg rewrites --repo-root args to the validated absolute root.
func NormalizeRepoRootArg(args []string, root string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--repo-root" && i+1 < len(args) {
			out = append(out, arg, root)
			i++
			continue
		}
		if strings.HasPrefix(arg, "--repo-root=") {
			out = append(out, "--repo-root="+root)
			continue
		}
		out = append(out, arg)
	}
	return out
}

// ValidateRepoRoot returns an absolute, clean repo root containing the research CLI.
func ValidateRepoRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("research tool: resolve repo root %q: %w", root, err)
	}
	abs = filepath.Clean(abs)
	if ScriptExists(abs) {
		return abs, nil
	}
	return "", fmt.Errorf("research tool: could not find tools/research/research_cli.py under %s", abs)
}

// FindRepoRoot walks up from start until it finds tools/research/research_cli.py.
func FindRepoRoot(start string) (string, bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	current = filepath.Clean(current)
	for {
		if ScriptExists(current) {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

// ScriptExists reports whether root contains tools/research/research_cli.py.
func ScriptExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, "tools", "research", "research_cli.py"))
	return err == nil && !info.IsDir()
}
