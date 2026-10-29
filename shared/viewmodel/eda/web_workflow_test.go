package eda_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	config "fecim-lattice-tools/module6-eda/pkg/config"
	export "fecim-lattice-tools/module6-eda/pkg/export"
	legacy "fecim-lattice-tools/shared/viewmodel/eda"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

type projectGrant struct{ root *os.Root }

func (g projectGrant) WithProject(ctx context.Context, ref moduleapi.ResourceReference, _ bool, fn func(*os.Root) error) error {
	if ref.ID != "opened" || ref.Kind != "project" {
		return os.ErrPermission
	}
	return fn(g.root)
}
func TestEDAWebWorkflow(t *testing.T) {
	provider, ok := any(legacy.New()).(moduleapi.Provider)
	if !ok {
		t.Fatal("EDA generate actions are successful no-ops: no real web Provider")
	}
	port := provider.WebPort()
	defer port.Close()
	jobs, ok := port.(moduleapi.JobPort)
	if !ok {
		t.Fatal("missing real EDA jobs")
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	ctx := native.WithProjects(context.Background(), projectGrant{root})
	cfg := compiler.NewComputeConfig(2, 3)
	payload, _ := json.Marshal(map[string]any{"config": cfg, "cell": config.DefaultCellConfig()})
	before, _ := port.Snapshot(ctx)
	job, err := jobs.Start(ctx, moduleapi.Command{Action: "generate_all", Revision: before.Revision, Payload: payload, Resources: []moduleapi.ResourceReference{{ID: "opened", Kind: "project"}}})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for job.State == "queued" || job.State == "running" {
		if time.Now().After(deadline) {
			t.Fatal("job did not finish")
		}
		time.Sleep(time.Millisecond)
		job, err = jobs.Job(ctx, job.ID)
		if err != nil {
			t.Fatal(err)
		}
	}
	if job.State != "finished" {
		t.Fatalf("not generated: %+v", job)
	}
	dir := filepath.Join("exports", job.ID)
	data, err := root.ReadFile(filepath.Join(dir, "design.spice"))
	if err != nil {
		t.Fatal(err)
	}
	design, _ := compiler.GenerateDesign(cfg)
	if string(data) != export.GenerateSPICE(design, cfg.Peripherals.VDD) {
		t.Fatal("SPICE differs from Go oracle")
	}
	manifest, err := root.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Origin  string            `json:"origin"`
		Warning string            `json:"warning"`
		Files   map[string]string `json:"files_sha256"`
	}
	if err = json.Unmarshal(manifest, &m); err != nil {
		t.Fatal(err)
	}
	if m.Origin != "freeform" || !strings.Contains(m.Warning, "ESTIMATED_SCALAR") || !strings.Contains(m.Warning, "Placeholder") {
		t.Fatalf("false provenance: %s", manifest)
	}
	for name, want := range m.Files {
		b, e := root.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatalf("hash %s", name)
		}
	}
	snapshot, _ := port.Snapshot(ctx)
	snapshot, err = port.Apply(ctx, moduleapi.Command{Action: "preview", Revision: snapshot.Revision, Input: map[string]string{"artifact": "design.v"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = port.Apply(ctx, moduleapi.Command{Action: "preview", Revision: snapshot.Revision, Input: map[string]string{"artifact": "../project.yaml"}}); err == nil {
		t.Fatal("hostile preview admitted")
	}
	if err = root.WriteFile(filepath.Join(dir, "design.spice"), []byte("tamper"), 0600); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = port.Snapshot(ctx)
	if _, err = port.Apply(ctx, moduleapi.Command{Action: "validate", Revision: snapshot.Revision}); err == nil {
		t.Fatal("tampered export falsely validated")
	}
	snapshot, _ = port.Snapshot(ctx)
	job, err = jobs.Start(ctx, moduleapi.Command{Action: "generate_all", Revision: snapshot.Revision, Payload: payload, Resources: []moduleapi.ResourceReference{{ID: "opened", Kind: "project"}}})
	if err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(3 * time.Second)
	for job.State == "queued" || job.State == "running" {
		if time.Now().After(deadline) {
			t.Fatal("second generation stalled")
		}
		time.Sleep(time.Millisecond)
		job, err = jobs.Job(ctx, job.ID)
		if err != nil {
			t.Fatal(err)
		}
	}
	if job.State != "finished" {
		t.Fatal(job)
	}
	snapshot, _ = port.Snapshot(ctx)
	for _, f := range snapshot.Fields {
		if f.ID == "preview" && (f.Input != "" || f.Label != "") {
			t.Fatal("new export retained old preview under new provenance")
		}
	}
}
