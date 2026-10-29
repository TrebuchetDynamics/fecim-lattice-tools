package eda

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/edaexport"
)

const MaxLogBytes = 1 << 20

type ToolResult struct {
	State     string `json:"state"`
	Summary   string `json:"summary"`
	Tool      string `json:"tool,omitempty"`
	Version   string `json:"version,omitempty"`
	Command   string `json:"command,omitempty"`
	Log       string `json:"log,omitempty"`
	LogBytes  int64  `json:"log_bytes"`
	Truncated bool   `json:"truncated"`
}
type toolRunner struct {
	executables                map[string]string
	commandTimeout, jobTimeout time.Duration
	stagingLimit               int64
}

func installedRunner() toolRunner {
	paths := map[string]string{}
	for _, name := range []string{"iverilog", "yosys"} {
		if p, e := exec.LookPath(name); e == nil {
			if abs, e := filepath.Abs(p); e == nil {
				paths[name] = abs
			}
		}
	}
	return toolRunner{paths, edaexport.CommandTimeout, edaexport.JobTimeout, edaexport.MaxStagingBytes}
}
func DiscoverTools() ToolResult {
	r := installedRunner()
	summary := "Local structural profile: "
	for _, name := range []string{"iverilog", "yosys"} {
		state := "unavailable (not installed)"
		if r.executables[name] != "" {
			state = "installed; not run or qualified"
		}
		summary += name + " " + state + ". "
	}
	summary += "ngspice, OpenROAD, KLayout, Magic, Netgen and OpenLane are unavailable in this initial bounded web profile; need separate model/PDK/resource authorization. No installation, Docker pull or network acquisition."
	return ToolResult{State: "discovered-not-validated", Summary: summary}
}
func RunTool(ctx context.Context, root *os.Root, id, tool string) (ToolResult, error) {
	return installedRunner().run(ctx, root, id, tool)
}
func (r toolRunner) run(ctx context.Context, root *os.Root, id, tool string) (ToolResult, error) {
	result := ToolResult{Tool: tool, State: "unavailable", Summary: "Optional tool unavailable; no validation passed"}
	if tool != "iverilog" && tool != "yosys" {
		return result, errors.New("tool is not allowlisted")
	}
	executable := r.executables[tool]
	if executable == "" {
		return result, nil
	}
	if _, err := edaexport.InspectRoot(root, id); err != nil {
		return result, err
	}
	files := map[string][]byte{}
	for _, name := range []string{"design.v", "cell.v"} {
		data, err := edaexport.ReadArtifactRoot(root, id, name)
		if err != nil {
			result.Summary = "This bundle lacks the freeform cell Verilog required for installed-tool checks; skipped, not passed"
			return result, nil
		}
		files[name] = data
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, min(r.jobTimeout, edaexport.JobTimeout))
	defer cancel()
	if err := root.MkdirAll(".eda-staging", 0700); err != nil {
		return result, err
	}
	parent, err := root.OpenRoot(".eda-staging")
	if err != nil {
		return result, err
	}
	defer parent.Close()
	stageID := randomID()
	if err = parent.Mkdir(stageID, 0700); err != nil {
		return result, err
	}
	stage, err := parent.OpenRoot(stageID)
	if err != nil {
		return result, err
	}
	defer stage.Close()
	defer func() {
		owned, e := stage.Stat(".")
		entry, e2 := parent.Lstat(stageID)
		if e == nil && e2 == nil && os.SameFile(owned, entry) {
			_ = parent.RemoveAll(stageID)
		}
	}()
	for name, data := range files {
		if err = stage.WriteFile(name, data, 0600); err != nil {
			return result, err
		}
	}
	log := &boundedLog{limit: MaxLogBytes}
	versionLog := &boundedLog{limit: MaxLogBytes}
	if err = runProcess(ctx, stage, executable, []string{"-V"}, versionLog, min(r.commandTimeout, edaexport.CommandTimeout), r.stagingLimit); err != nil {
		return result, fmt.Errorf("tool version failed: %w", err)
	}
	result.Version = versionLog.text()
	if versionLog.truncated() {
		return result, errors.New("truncated tool version evidence")
	}
	// Version and command stdout/stderr share one retained-byte budget.
	log.limit = MaxLogBytes - len(result.Version)
	var args []string
	if tool == "iverilog" {
		args = []string{"-g2012", "-t", "null", "design.v", "cell.v"}
		result.Command = "iverilog -g2012 -t null design.v cell.v (syntax/elaboration only)"
	} else {
		args = []string{"-p", "read_verilog design.v cell.v; hierarchy -check; check"}
		result.Command = "yosys: read_verilog design.v cell.v; hierarchy -check; check (structural only)"
	}
	err = runProcess(ctx, stage, executable, args, log, min(r.commandTimeout, edaexport.CommandTimeout), r.stagingLimit)
	result.Log = log.text()
	result.LogBytes = versionLog.count + log.count
	result.Truncated = log.truncated()
	if err != nil {
		result.State = "failed"
		result.Summary = "Tool failed/cancelled; no validation passed"
		return result, err
	}
	result.State = "structural-only"
	result.Summary = "Installed tool exited zero for server-owned syntax/structure command; Placeholder model, not science, signoff, LVS or foundry qualification"
	if result.Truncated {
		result.State = "incomplete"
		result.Summary = "Tool exited zero but logs are truncated; incomplete evidence, not a validation pass"
	}
	return result, nil
}

type boundedLog struct {
	mu    sync.Mutex
	data  []byte
	count int64
	limit int
}

func (b *boundedLog) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.count += int64(len(p))
	remaining := b.limit - len(b.data)
	if remaining > 0 {
		b.data = append(b.data, p[:min(remaining, len(p))]...)
	}
	return len(p), nil
}
func (b *boundedLog) truncated() bool { return b.count > int64(len(b.data)) }
func (b *boundedLog) text() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := string(b.data)
	if b.truncated() {
		s += fmt.Sprintf("\n[truncated: retained %d of %d drained bytes]\n", len(b.data), b.count)
	}
	return s
}
