package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDigitizeCommand(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "pe-reference")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, out)
	}
	profile, err := filepath.Abs("../../validation/literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	const header = "branch,point_index,x_px,y_px\n"
	const valid = header + "descending,0,905.5,217.5\nascending,1,589,578.5\nascending,0,272.5,939.5\n"
	const want = "branch,point_index,field_MV_cm,polarization_uC_cm2\nascending,0,-3,-45\nascending,1,0,0\ndescending,0,3,45\n"
	for _, tc := range []struct {
		name, pixels string
		mode         string
		fail         bool
	}{
		{"valid", valid, "", false},
		{"invalid points", valid + "ascending,2,NaN,578.5\n", "", true},
		{"existing output", valid, "existing", true},
		{"missing pixels", valid, "missing pixels", true},
		{"missing profile", valid, "missing profile", true},
		{"invalid profile", valid, "invalid profile", true},
		{"missing out flag", valid, "missing out", true},
		{"unknown flag", valid, "unknown flag", true},
		{"unexpected argument", valid, "extra argument", true},
		{"unknown command", valid, "unknown command", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			pixels, output := filepath.Join(dir, "pixels.csv"), filepath.Join(dir, "physical.csv")
			if err := os.WriteFile(pixels, []byte(tc.pixels), 0600); err != nil {
				t.Fatal(err)
			}
			profilePath := profile
			switch tc.mode {
			case "existing":
				if err := os.WriteFile(output, []byte("preserve me"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing pixels":
				pixels += ".missing"
			case "missing profile":
				profilePath += ".missing"
			case "invalid profile":
				profilePath = filepath.Join(dir, "invalid.json")
				if err := os.WriteFile(profilePath, []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"digitize", "--profile", profilePath, "--pixels", pixels, "--out", output}
			switch tc.mode {
			case "missing out":
				args = args[:len(args)-2]
			case "unknown flag":
				args = append(args, "--online")
			case "extra argument":
				args = append(args, "unexpected")
			case "unknown command":
				args[0] = "fetch"
			}
			// Real subprocess: a command failure must be observable as a nonzero exit.
			attempts := 1
			if !tc.fail {
				attempts = 2 // Repeated successful runs must produce identical bytes.
			}
			for attempt := 0; attempt < attempts; attempt++ {
				cmd := exec.Command(binary, args...)
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if (err != nil) != tc.fail || stdout.Len() != 0 || (tc.fail && strings.TrimSpace(stderr.String()) == "") || (!tc.fail && stderr.Len() != 0) {
					t.Fatalf("exit=%v stdout=%q stderr=%q; fail=%v", err, &stdout, &stderr, tc.fail)
				}
				got, readErr := os.ReadFile(output)
				switch {
				case tc.mode == "existing":
					if readErr != nil || string(got) != "preserve me" {
						t.Fatalf("existing output changed: %q, %v", got, readErr)
					}
				case tc.fail:
					if !os.IsNotExist(readErr) {
						t.Fatalf("failed conversion created output: %q, %v", got, readErr)
					}
				default:
					if readErr != nil || string(got) != want {
						t.Fatalf("output=%q, %v; want %q", got, readErr, want)
					}
					if err := os.Remove(output); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
