package literature

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/validation/peloop"
)

// TestHZOIndependentDigitization catches substituted inputs, stale conversion,
// and truncated/reversed branches. It cannot certify visual marker fidelity.
func TestHZOIndependentDigitization(t *testing.T) {
	base := filepath.Join("data", "experimental", "kim2020_fig3a_ref_digitized")
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(base, name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	pixels := read("points.pixels.csv")
	profileBytes, physical := read("source.provenance.json"), read("loop.csv")
	profile, err := peloop.ReadSourceProfile(bytes.NewReader(profileBytes))
	if err != nil {
		t.Fatal(err)
	}
	const sourceSHA = "2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa"
	if profile.DatasetID != "kim2020_fig3a_ref_digitized" || profile.SourceAsset.SHA256 != sourceSHA {
		t.Fatal("reference must retain the independently pinned primary source")
	}
	regenerated, err := peloop.Digitize(profile, bytes.NewReader(pixels))
	if err != nil || !bytes.Equal(regenerated, physical) {
		t.Fatalf("physical CSV does not regenerate byte-for-byte: %v", err)
	}
	_, receipt, ok := strings.Cut(string(read("extraction.md")), "```json\n")
	if !ok {
		t.Fatal("missing extraction receipt")
	}
	receipt, _, ok = strings.Cut(receipt, "\n```")
	var manifest struct {
		DatasetID string            `json:"dataset_id"`
		SourceSHA string            `json:"source_sha256"`
		Hashes    map[string]string `json:"sha256"`
		Counts    map[string]int    `json:"counts"`
	}
	if !ok || json.Unmarshal([]byte(receipt), &manifest) != nil {
		t.Fatal("invalid extraction receipt")
	}
	if manifest.DatasetID != profile.DatasetID || manifest.SourceSHA != sourceSHA || len(manifest.Hashes) != 3 || len(manifest.Counts) != 2 {
		t.Fatal("receipt source identity or inventory mismatch")
	}
	for name, data := range map[string][]byte{"source.provenance.json": profileBytes, "points.pixels.csv": pixels, "loop.csv": physical} {
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != manifest.Hashes[name] {
			t.Errorf("%s SHA-256: got %s, receipt %s", name, got, manifest.Hashes[name])
		}
	}

	// ReadSamples enforces finite values, contiguous indices, distinct fields and
	// branch direction. Coverage additionally requires both signs, the visible
	// saturation spans and a sign-correct P=0 bracket on each branch, not a fit.
	validate := func(data []byte) ([]peloop.Sample, error) {
		samples, err := peloop.ReadSamples(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		for _, name := range []string{"ascending", "descending"} {
			var branch []peloop.Sample
			for _, s := range samples {
				if s.Branch == name {
					branch = append(branch, s)
				}
			}
			if len(branch) < 10 {
				return nil, fmt.Errorf("%s: fewer than 10 markers", name)
			}
			lo, hi := branch[0], branch[len(branch)-1]
			if name == "descending" {
				lo, hi = hi, lo
			}
			minE, maxE, minP, maxP := -2.2, 1.7, -30.0, 27.0
			if name == "descending" {
				minE, maxE, minP, maxP = -2.3, 2.4, -31, 31
			}
			if lo.FieldMVcm > minE || hi.FieldMVcm < maxE || lo.PolarizationUCcm2 > minP || hi.PolarizationUCcm2 < maxP {
				return nil, fmt.Errorf("%s: missing visible span/signs", name)
			}
			crossings := 0
			for i := 1; i < len(branch); i++ {
				a, b := branch[i-1], branch[i]
				if a.PolarizationUCcm2*b.PolarizationUCcm2 <= 0 {
					if (name == "ascending" && a.FieldMVcm > 0 && b.FieldMVcm > 0) || (name == "descending" && a.FieldMVcm < 0 && b.FieldMVcm < 0) {
						crossings++
					}
				}
			}
			if crossings != 1 {
				return nil, fmt.Errorf("%s: missing unique sign-correct zero crossing", name)
			}
		}
		return samples, nil
	}
	samples, err := validate(physical)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, s := range samples {
		counts[s.Branch]++
	}
	for name, count := range counts {
		if count != manifest.Counts[name] {
			t.Errorf("%s: %d markers, receipt %d", name, count, manifest.Counts[name])
		}
	}
	// Corrupt the admitted data at the same seam, without any legacy fixture.
	for _, corruption := range []string{"missing branch", "too few", "lost field signs", "lost P signs", "misplaced P crossing", "truncated saturation"} {
		t.Run(corruption, func(t *testing.T) {
			var invalid bytes.Buffer
			invalid.WriteString("branch,point_index,field_MV_cm,polarization_uC_cm2\n")
			for _, original := range samples {
				s := original
				if corruption == "truncated saturation" && s.Branch == "descending" {
					if s.PointIndex < 2 {
						continue
					}
					s.PointIndex -= 2
				}
				if s.Branch == "ascending" {
					switch corruption {
					case "missing branch":
						continue
					case "too few":
						if s.PointIndex >= 9 {
							continue
						}
					case "lost field signs":
						s.FieldMVcm += 4
					case "lost P signs":
						s.PolarizationUCcm2 += 40
					case "misplaced P crossing":
						if s.FieldMVcm > -1 {
							s.PolarizationUCcm2 = 30
						}
					}
				}
				fmt.Fprintf(&invalid, "%s,%d,%g,%g\n", s.Branch, s.PointIndex, s.FieldMVcm, s.PolarizationUCcm2)
			}
			if _, err := validate(invalid.Bytes()); err == nil {
				t.Fatal("invalid reference coverage admitted")
			}
		})
	}
}
