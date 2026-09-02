package openlane

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Result contains the output of a tool execution
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Runner executes OpenLane tools in Docker or native mode
type Runner struct {
	manager *Manager
	config  *Config
}

// NewRunner creates a new runner
func NewRunner(manager *Manager, config *Config) *Runner {
	return &Runner{
		manager: manager,
		config:  config,
	}
}

// RunOpenROAD executes an OpenROAD TCL script
// workDir should contain the script file and design files
// scriptName is the name of the TCL script file in workDir
func (r *Runner) RunOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerOpenROAD(scriptName, workDir, envVars)
	case ModeNative:
		return r.runNativeOpenROAD(scriptName, workDir, envVars)
	default:
		return nil, fmt.Errorf("no OpenROAD execution mode available (install Docker or OpenROAD)")
	}
}

const (
	xvfbPrefix            = `Xvfb :99 -screen 0 1024x768x24 -nolisten tcp >/dev/null 2>&1 & sleep 1; export DISPLAY=:99; `
	dockerOpenROADProgram = xvfbPrefix + `exec openroad -no_splash -exit "$1"`
	dockerKLayoutProgram  = xvfbPrefix + `exec klayout "$@"`
)

func dockerDesignScriptPath(scriptPath string) string {
	const invalidPath = "/design/__invalid_script__"
	if scriptPath == "" {
		return invalidPath
	}

	cleaned := filepath.Clean(scriptPath)
	if filepath.IsAbs(cleaned) {
		base := filepath.Base(cleaned)
		if base == "." || base == ".." || base == string(filepath.Separator) {
			return invalidPath
		}
		return "/design/" + filepath.ToSlash(base)
	}
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return invalidPath
	}
	return "/design/" + filepath.ToSlash(cleaned)
}

func dockerOpenROADArgs(image, absWorkDir, scriptName string, envVars map[string]string) []string {
	args := []string{
		"run", "--rm",
		"--entrypoint", "sh",
		"-v", absWorkDir + ":/design",
		"-w", "/design",
	}
	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "-e", key+"="+envVars[key])
	}
	return append(args,
		image,
		"-c", dockerOpenROADProgram,
		"fecim-openroad", dockerDesignScriptPath(scriptName),
	)
}

// runDockerOpenROAD runs OpenROAD in Docker container with Xvfb for headless image export.
func (r *Runner) runDockerOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}
	args := dockerOpenROADArgs(r.manager.GetDockerImage(), absWorkDir, scriptName, envVars)
	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutPlacement)
}

// runNativeOpenROAD runs OpenROAD directly
func (r *Runner) runNativeOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	scriptPath := filepath.Join(workDir, scriptName)
	args := []string{"-no_splash", "-exit", scriptPath}

	// Set up environment from caller-provided vars
	// For FeCIM validation, caller provides CELL_LEF path - no external PDK required
	env := os.Environ()

	// Add user-provided env vars (CELL_LEF, DEF_FILE, etc.)
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return r.runWithTimeoutEnv("openroad", args, workDir, r.config.TimeoutPlacement, env)
}

// RunYosys executes a Yosys command
// workDir should contain the Verilog files
// yosysCmd is the yosys command string (e.g., "read_verilog file.v; hierarchy -check")
func (r *Runner) RunYosys(yosysCmd string, workDir string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerYosys(yosysCmd, workDir)
	case ModeNative:
		return r.runNativeYosys(yosysCmd, workDir)
	default:
		return nil, fmt.Errorf("no Yosys execution mode available (install Docker with OpenLane image or native yosys)")
	}
}

// runDockerYosys runs Yosys in Docker container
func (r *Runner) runDockerYosys(yosysCmd string, workDir string) (*Result, error) {
	// Docker requires absolute paths for volume mounts
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Build Docker command with --entrypoint yosys
	args := []string{
		"run", "--rm",
		"--entrypoint", "yosys",
		"-v", fmt.Sprintf("%s:/design", absWorkDir),
		"-w", "/design",
	}

	// Add image and Yosys command
	args = append(args, r.manager.GetDockerImage())
	args = append(args, "-p", yosysCmd)

	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutSynthesis)
}

// runNativeYosys runs Yosys directly
func (r *Runner) runNativeYosys(yosysCmd string, workDir string) (*Result, error) {
	args := []string{"-p", yosysCmd}
	return r.runWithTimeout("yosys", args, workDir, r.config.TimeoutSynthesis)
}

// RunKLayout executes a KLayout script for layout visualization
// workDir should contain the DEF and LEF files
// scriptPath is the path to the Python/Ruby script
func (r *Runner) RunKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerKLayout(scriptPath, workDir, envVars)
	case ModeNative:
		return r.runNativeKLayout(scriptPath, workDir, envVars)
	default:
		return nil, fmt.Errorf("no KLayout execution mode available (install Docker with OpenLane image or native klayout)")
	}
}

func dockerKLayoutArgs(image, absWorkDir, scriptPath string, envVars map[string]string) []string {
	args := []string{
		"run", "--rm",
		"--entrypoint", "sh",
		"-v", absWorkDir + ":/design",
		"-w", "/design",
		image,
		"-c", dockerKLayoutProgram,
		"fecim-klayout", "-z",
	}
	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, "-rd", key+"="+envVars[key])
	}
	return append(args, "-r", dockerDesignScriptPath(scriptPath))
}

// runDockerKLayout runs KLayout in Docker container with Xvfb for headless image export.
func (r *Runner) runDockerKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}
	args := dockerKLayoutArgs(r.manager.GetDockerImage(), absWorkDir, scriptPath, envVars)
	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutPlacement)
}

// runNativeKLayout runs KLayout directly
// Uses -rd flags to pass variables to scripts (standard KLayout pattern)
func (r *Runner) runNativeKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	// KLayout flags: -z for batch mode with main window (required for image export)
	args := []string{"-z"}

	// Add variables using -rd (standard KLayout pattern per docs)
	for k, v := range envVars {
		args = append(args, "-rd", fmt.Sprintf("%s=%s", k, v))
	}

	// Add script
	args = append(args, "-r", scriptPath)

	return r.runWithTimeout("klayout", args, workDir, r.config.TimeoutPlacement)
}

// runWithTimeout executes a command with timeout
func (r *Runner) runWithTimeout(command string, args []string, workDir string, timeout time.Duration) (*Result, error) {
	return r.runWithTimeoutEnv(command, args, workDir, timeout, nil)
}

// runWithTimeoutEnv executes a command with timeout and custom environment
func (r *Runner) runWithTimeoutEnv(command string, args []string, workDir string, timeout time.Duration, env []string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	if env != nil {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %v", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}
