package project

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A disk fallback, invented input hash, permissive decoder or allocating a huge
// range before preflight breaks the byte-only loader contract.
func TestLoadBytes(t *testing.T) {
	documents := func() map[string][]byte {
		return map[string][]byte{
			"project.yaml": []byte(validProjectYAML), "design.yaml": []byte(validDesignYAML), "sweep.yaml": []byte(validSweepYAML),
		}
	}
	t.Run("native hashes and defaults", func(t *testing.T) {
		docs := documents()
		input := []byte("voltage,current\n1,2\n")
		digest := fmt.Sprintf("%x", sha256.Sum256(input))
		docs["data/input.csv"] = input
		docs["project.yaml"] = append(docs["project.yaml"], []byte(fmt.Sprintf("inputs:\n  - path: data/input.csv\n    sha256: %s\n    citation: park2015_advmat_hzo\n    evidence: experiment-calibrated\n", digest))...)
		docs["sweep.yaml"] = []byte(strings.Replace(string(docs["sweep.yaml"]), "max_points: 32\n", "", 1))
		got, err := LoadBytes(docs)
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		for name, data := range docs {
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
		native, err := Load(dir, LoadOptions{})
		if err != nil {
			t.Fatal(err)
		}
		native.Root = ""
		if !reflect.DeepEqual(got, native) || got.Sweep.MaxPoints != 10000 || got.Inputs[0].SHA256 != digest {
			t.Fatalf("byte/native mismatch: %+v / %+v", got, native)
		}
		delete(docs, "data/input.csv")
		if _, err := LoadBytes(docs); err == nil {
			t.Fatal("missing input accepted")
		}
		docs["data/input.csv"] = []byte("tampered")
		if _, err := LoadBytes(docs); err == nil {
			t.Fatal("hash mismatch accepted")
		}
	})
	for _, tt := range []struct{ name, document, body string }{
		{"unknown field", "design.yaml", validDesignYAML + "unknown: true\n"},
		{"multiple YAML", "project.yaml", validProjectYAML + "---\n" + validProjectYAML},
		{"missing schema", "project.yaml", strings.Replace(validProjectYAML, "schema_version: 1", "schema_version: 2", 1)},
		{"oversize", "project.yaml", validProjectYAML + strings.Repeat("#", 1<<20)},
		{"huge range", "sweep.yaml", "schema_version: 1\nmax_points: 32\nparameters:\n  - path: array.rows\n    range: {start: 1, stop: 32, count: 9223372036854775807}\n"},
		{"integer overflow", "sweep.yaml", "schema_version: 1\nparameters:\n  - path: array.rows\n    values: [1e100]\n"},
		{"bad citation", "project.yaml", strings.Replace(validProjectYAML, "park2015_advmat_hzo", "../outside", 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			docs := documents()
			docs[tt.document] = []byte(tt.body)
			if _, err := LoadBytes(docs); err == nil {
				t.Fatal("invalid bundle accepted")
			}
		})
	}
	t.Run("hosted limits before expansion", func(t *testing.T) {
		docs := documents()
		docs["sweep.yaml"] = []byte("schema_version: 1\nmax_points: 100000\nparameters:\n  - path: array.rows\n    range: {start: 1, stop: 32, count: 100000}\n")
		if _, err := LoadBytesLimited(docs, 256, 32); err == nil {
			t.Fatal("hosted point cap not enforced")
		}
		docs = documents()
		docs["sweep.yaml"] = []byte("schema_version: 1\nparameters:\n  - path: array.rows\n    values: [33]\n")
		if _, err := LoadBytesLimited(docs, 256, 32); err == nil {
			t.Fatal("hosted dimension cap not enforced")
		}
	})
	t.Run("rooted regular bounded reads", func(t *testing.T) {
		for _, target := range []string{"project.yaml", "input.bin"} {
			t.Run(target+" FIFO", func(t *testing.T) {
				if runtime.GOOS != "linux" {
					t.Skip("FIFO regression uses Linux named pipes")
				}
				docs := documents()
				dir := t.TempDir()
				if target == "input.bin" {
					docs["project.yaml"] = append(docs["project.yaml"], []byte(fmt.Sprintf("inputs:\n  - path: input.bin\n    sha256: %064x\n    evidence: simulation-default\n", 0))...)
				}
				for name, data := range docs {
					if name != target {
						if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				fifo := filepath.Join(dir, target)
				if err := exec.Command("mkfifo", fifo).Run(); err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				go func() { _, err := Load(dir, LoadOptions{}); done <- err }()
				select {
				case err := <-done:
					if err == nil {
						t.Fatal("FIFO accepted")
					}
					if !strings.Contains(err.Error(), "regular") {
						t.Fatalf("nonregular rejection: %v", err)
					}
				case <-time.After(300 * time.Millisecond):
					file, err := os.OpenFile(fifo, os.O_RDWR, 0600)
					if err != nil {
						t.Fatal(err)
					}
					file.WriteString("invalid\n")
					file.Close()
					<-done
					t.Fatal("rooted load blocked opening FIFO")
				}
			})
		}
		for _, sizes := range [][]int{{(8 << 20) + 1}, {5 << 20, 5 << 20}} {
			t.Run(fmt.Sprint(sizes), func(t *testing.T) {
				docs := documents()
				dir := t.TempDir()
				refs := "inputs:\n"
				for i, size := range sizes {
					data := make([]byte, size)
					name := fmt.Sprintf("input-%d.bin", i)
					if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
						t.Fatal(err)
					}
					refs += fmt.Sprintf("  - path: %s\n    sha256: %x\n    evidence: simulation-default\n", name, sha256.Sum256(data))
				}
				docs["project.yaml"] = append(docs["project.yaml"], refs...)
				for name, data := range docs {
					if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := Load(dir, LoadOptions{}); err == nil {
					t.Fatal("unbounded native inputs accepted")
				}
			})
		}
		if runtime.GOOS == "linux" {
			root, err := os.OpenRoot("/dev")
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			if err := decodeRoot(root, "null", new(Project)); err == nil || !strings.Contains(err.Error(), "regular") {
				t.Fatalf("device was opened as configuration: %v", err)
			}
		}
	})
	t.Run("rooted range preflight", func(t *testing.T) {
		docs := documents()
		docs["sweep.yaml"] = []byte("schema_version: 1\nmax_points: 32\nparameters:\n  - path: array.rows\n    range: {start: 1, stop: 32, count: 9223372036854775807}\n")
		dir := t.TempDir()
		for name, data := range docs {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
		defer func() {
			if recover() != nil {
				t.Error("rooted load allocated range before preflight")
			}
		}()
		if _, err := Load(dir, LoadOptions{}); err == nil {
			t.Fatal("oversized range accepted")
		}
	})
	t.Run("unsafe and unreferenced documents", func(t *testing.T) {
		for _, name := range []string{"../outside", "/outside", "data/../outside", "extra.yaml"} {
			docs := documents()
			docs[name] = []byte("extra")
			if _, err := LoadBytes(docs); err == nil {
				t.Fatalf("accepted %s", name)
			}
		}
		docs := documents()
		delete(docs, "design.yaml")
		if _, err := LoadBytes(docs); err == nil {
			t.Fatal("missing design accepted")
		}
	})
}
