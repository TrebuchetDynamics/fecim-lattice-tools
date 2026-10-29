//go:build linux

package eda

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

func TestEDAWebToolBoundary(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	id := strings.Repeat("c", 64)
	if _, err = edaexport.GenerateBuilderRoot(context.Background(), root, id, edaexport.Defaults(compiler.ModeCompute, 2, 3, compiler.ArchPassive)); err != nil {
		t.Fatal(err)
	}
	fixture := func(name, script string) string {
		p := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"+script), 0700); err != nil {
			t.Fatal(err)
		}
		return p
	}
	missing := toolRunner{executables: map[string]string{}, commandTimeout: time.Second, jobTimeout: time.Second, stagingLimit: edaexport.MaxStagingBytes}
	result, err := missing.run(context.Background(), root, id, "iverilog")
	if err != nil || result.State != "unavailable" {
		t.Fatalf("missing tool falsely passed: %+v %v", result, err)
	}
	for _, name := range []string{"/bin/sh", "iverilog;echo x", "../yosys"} {
		if _, err = missing.run(context.Background(), root, id, name); err == nil {
			t.Fatal("client command admitted")
		}
	}
	runner := missing
	runner.executables = map[string]string{"iverilog": fixture("fixture-iverilog", "if [ \"$1\" = '-V' ]; then echo 'PRIVATE fixture, not real Icarus'; exit 0; fi\nyes '<script>not HTML</script>' | head -c 1200000\n")}
	result, err = runner.run(context.Background(), root, id, "iverilog")
	if err != nil || !result.Truncated || result.LogBytes < 1200000 || len(result.Log) > MaxLogBytes+256 || !strings.Contains(result.Log, "truncated") || !strings.Contains(result.Version, "PRIVATE fixture") {
		t.Fatalf("log/evidence bounds: %+v %v", result, err)
	}
	runner.executables["iverilog"] = fixture("fixture-combined-budget", "if [ \"$1\" = '-V' ]; then yes version | head -c 700000; exit 0; fi\nyes command | head -c 700000\n")
	result, err = runner.run(context.Background(), root, id, "iverilog")
	if err != nil || len(result.Version)+len(result.Log) > MaxLogBytes+256 || !result.Truncated {
		t.Fatal("version and command logs exceed combined budget", err)
	}
	runner.executables["iverilog"] = fixture("fixture-stall", "if [ \"$1\" = '-V' ]; then echo fixture; exit; fi\nsleep 30 &\nwait\n")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err = runner.run(ctx, root, id, "iverilog"); err == nil || time.Since(start) > 3*time.Second {
		t.Fatal("cancel did not clean process group", err)
	}
	runner.executables["iverilog"] = fixture("fixture-disk", "if [ \"$1\" = '-V' ]; then echo fixture; exit; fi\nhead -c 10000 /dev/zero > excess\nsleep 30\n")
	initialA, _ := edaexport.ReadArtifactRoot(root, id, "design.v")
	initialB, _ := edaexport.ReadArtifactRoot(root, id, "cell.v")
	runner.stagingLimit = int64(len(initialA)+len(initialB)) + 1000
	if _, err = runner.run(context.Background(), root, id, "iverilog"); err == nil {
		t.Fatal("staging excess not aborted")
	}
	// A failed public job must preserve the same execution identity as a success.
	fake := fixture("iverilog", "if [ \"$1\" = '-V' ]; then echo 'PRIVATE failed-tool fixture'; exit; fi\necho syntax-failed\nexit 23\n")
	t.Setenv("PATH", filepath.Dir(fake))
	p := New()
	defer p.Close()
	p.last, err = edaexport.InspectRoot(root, id)
	if err != nil {
		t.Fatal(err)
	}
	p.project = moduleapi.ResourceReference{ID: "opened", Kind: "project"}
	ctx = native.WithProjects(context.Background(), weightGrant{root: root})
	job, err := p.Start(ctx, moduleapi.Command{Action: "tool", Input: map[string]string{"tool": "iverilog"}})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for job.State == "queued" || job.State == "running" {
		if time.Now().After(deadline) {
			t.Fatal("failed tool job stalled")
		}
		time.Sleep(time.Millisecond)
		job, _ = p.Job(ctx, job.ID)
	}
	snapshot, err := p.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var evidence ToolResult
	for _, f := range snapshot.Fields {
		if f.ID == "tool" {
			if err = json.Unmarshal([]byte(f.Input), &evidence); err != nil {
				t.Fatal(err)
			}
		}
	}
	if job.State != "failed" || evidence.Tool != "iverilog" || !strings.Contains(evidence.Version, "PRIVATE failed-tool") || !strings.Contains(evidence.Command, "iverilog -g2012") {
		t.Fatalf("failed tool lost provenance: %+v", evidence)
	}
	dir, e := root.Open(".eda-staging")
	if e == nil {
		defer dir.Close()
		names, _ := dir.Readdirnames(-1)
		if len(names) != 0 {
			t.Fatalf("staging not cleaned: %v", names)
		}
	}
}
