package edaexport

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	config "fecim-lattice-tools/module6-eda/pkg/config"
)

func TestEDAWebAdmission(t *testing.T) {
	for _, arch := range []string{compiler.ArchPassive, compiler.Arch1T1R, compiler.Arch2T1R} {
		for _, mode := range []compiler.OperationMode{compiler.ModeStorage, compiler.ModeMemory, compiler.ModeCompute} {
			t.Run(arch+mode.String(), func(t *testing.T) {
				c := compiler.NewArrayConfig(mode, 8, 8)
				if arch == compiler.Arch1T1R {
					c.With1T1R()
				}
				if arch == compiler.Arch2T1R {
					c.With2T1R()
				}
				b := Builder{Config: c, Cell: config.DefaultCellConfig()}
				b.Cell.CellType = arch
				r, e := os.OpenRoot(t.TempDir())
				if e != nil {
					t.Fatal(e)
				}
				defer r.Close()
				id := strings.Repeat("a", 64)
				if _, e = GenerateBuilderRoot(context.Background(), r, id, b); e != nil {
					t.Fatal(e)
				}
				if _, e = InspectRoot(r, id); e != nil {
					t.Fatal(e)
				}
				array, e := r.ReadFile("exports/" + id + "/design.v")
				if e != nil {
					t.Fatal(e)
				}
				cell, e := r.ReadFile("exports/" + id + "/cell.v")
				if e != nil {
					t.Fatal(e)
				}
				if strings.Contains(string(array), ".LEVEL(") && !strings.Contains(string(cell), "parameter LEVEL = 0;") {
					t.Fatal("compiler LEVEL overrides lack compatible cell parameter")
				}
				if _, e = GenerateBuilderRoot(context.Background(), r, id, b); e == nil {
					t.Fatal("overwrote completed export")
				}
			})
		}
	}
	for _, malformed := range []func(*Builder){
		func(b *Builder) { b.Config.Mode = 99 }, func(b *Builder) { b.Config.Architecture = "unsupported" }, func(b *Builder) { b.Config.Technology = "GF180MCU" }, func(b *Builder) { b.Cell.Name = "../../escape" }, func(b *Builder) { b.Cell.Width = math.Inf(1) },
	} {
		b := Defaults(compiler.ModeCompute, 2, 3, compiler.ArchPassive)
		malformed(&b)
		if b.Validate() == nil {
			t.Fatal("unsupported/hostile builder admitted")
		}
	}
	for _, n := range []int{0, 257, math.MaxInt} {
		c := compiler.NewComputeConfig(n, 2)
		if (Builder{Config: c, Cell: config.DefaultCellConfig()}).Validate() == nil {
			t.Fatalf("rows %d admitted", n)
		}
	}
	for _, n := range []int{1, 256} {
		if e := (Builder{Config: compiler.NewComputeConfig(n, 256), Cell: config.DefaultCellConfig()}).Validate(); e != nil {
			t.Fatal(e)
		}
	}
	c := compiler.NewComputeConfig(2, 2)
	c.ComputeConfig.InitialWeights = [][]float64{{1, 2}, {3}}
	if (Builder{Config: c, Cell: config.DefaultCellConfig()}).Validate() == nil {
		t.Fatal("ragged weights admitted")
	}
	c.ComputeConfig.InitialWeights = [][]float64{{1, 2}, {3, math.NaN()}}
	if (Builder{Config: c, Cell: config.DefaultCellConfig()}).Validate() == nil {
		t.Fatal("nonfinite weights admitted")
	}
	c.ComputeConfig.InitialWeights = nil
	c.Name = "evil; exec x"
	if (Builder{Config: c, Cell: config.DefaultCellConfig()}).Validate() == nil {
		t.Fatal("script identifier admitted")
	}
	if _, e := DecodeBuilder([]byte(strings.Repeat(" ", MaxConfigBytes+1))); e == nil {
		t.Fatal("oversize JSON admitted")
	}
	raw, _ := json.Marshal(map[string]any{"config": compiler.NewComputeConfig(1, 1), "cell": config.DefaultCellConfig(), "script": "x"})
	if _, e := DecodeBuilder(raw); e == nil {
		t.Fatal("unknown script admitted")
	}
	if e := CheckArtifacts(map[string]int64{"design.spice": MaxArtifactBytes}); e != nil {
		t.Fatal(e)
	}
	if CheckArtifacts(map[string]int64{"design.spice": MaxArtifactBytes, "design.v": MaxArtifactBytes, "design.def": MaxArtifactBytes, "cell.v": MaxArtifactBytes, "cell.lef": 1}) == nil {
		t.Fatal("total byte budget admitted")
	}
	for _, sizes := range []map[string]int64{{"../escape": 1}, {"design.spice": MaxArtifactBytes + 1}, {"a": -1}} {
		if CheckArtifacts(sizes) == nil {
			t.Fatal("invalid artifact limits")
		}
	}
	many := map[string]int64{}
	for i := 0; i < 65; i++ {
		many[strings.Repeat("a", i+1)] = 1
	}
	if CheckArtifacts(many) == nil {
		t.Fatal("artifact count")
	}
	t.Run("manifest cannot omit required formats", func(t *testing.T) {
		root, e := os.OpenRoot(t.TempDir())
		if e != nil {
			t.Fatal(e)
		}
		defer root.Close()
		id := strings.Repeat("b", 64)
		if _, e = GenerateBuilderRoot(context.Background(), root, id, Defaults(compiler.ModeCompute, 1, 1, compiler.ArchPassive)); e != nil {
			t.Fatal(e)
		}
		data, e := root.ReadFile("exports/" + id + "/manifest.json")
		if e != nil {
			t.Fatal(e)
		}
		var manifest Manifest
		if e = json.Unmarshal(data, &manifest); e != nil {
			t.Fatal(e)
		}
		delete(manifest.FilesSHA256, "design.spice")
		data, _ = json.Marshal(manifest)
		if e = root.WriteFile("exports/"+id+"/manifest.json", data, 0600); e != nil {
			t.Fatal(e)
		}
		if _, e = InspectRoot(root, id); e == nil {
			t.Fatal("incomplete manifest falsely completed")
		}
	})
	if CommandTimeout.String() != "10m0s" || JobTimeout.String() != "30m0s" || MaxStagingBytes != 512<<20 {
		t.Fatal("policy drift")
	}
}
