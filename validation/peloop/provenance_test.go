package peloop_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/validation/peloop"
)

func TestSourceProfileAdmission(t *testing.T) {
	data, err := os.ReadFile("../literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	p, err := peloop.ReadSourceProfile(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if p.Schema != 1 || p.DatasetID != "kim2020_fig3a_ref_digitized" || p.DOI != "10.3390/ma13132968" || p.Figure != "3(a)" || p.Trace != "LEFT red circular Ref. markers; RTA after TE dep.; After wake-up" {
		t.Fatalf("wrong source identity: %+v", p)
	}
	if p.SourceAsset.SHA256 != "2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa" || p.SourceAsset.Width != 3629 || p.SourceAsset.Height != 1210 || p.SourceAsset.Bytes != 961494 {
		t.Fatalf("wrong raster: %+v", p.SourceAsset)
	}
	if p.License.URL != "https://creativecommons.org/licenses/by/4.0/" || p.License.EvidenceURL != "https://api.crossref.org/works/10.3390/ma13132968" || p.License.ContentVersion != "vor" || p.License.Attribution == "" {
		t.Fatalf("missing license evidence: %+v", p.License)
	}
	c := p.Conditions
	if c.Stack != "TiN 90 nm / Hf0.5Zr0.5O2 10 nm / TiN 90 nm" || c.FrequencyHz != 10000 || c.WakeUpCycles != 100000 || c.WakeUpFieldMVcm != 2.5 || c.Anneal.TemperatureC != 400 || c.Anneal.DurationS != 60 || c.Anneal.Atmosphere != "N2" || c.Anneal.Sequence != "after top-electrode deposition; no additional bottom-electrode pre-anneal" {
		t.Fatalf("wrong conditions: %+v", c)
	}
	for _, field := range []peloop.UnreportedField{c.MeasurementTemperature, c.Waveform, c.DeviceDiameter} {
		if field.Status != "not reported" || strings.TrimSpace(field.Detail) == "" {
			t.Fatalf("unreported measurement replaced by a default: %+v", field)
		}
	}
	if p.Calibration.X.Unit != "MV/cm" || p.Calibration.Y.Unit != "uC/cm2" || !reflect.DeepEqual(p.Calibration.X.Anchors, []peloop.AxisAnchor{{Pixel: 272.5, Value: -3}, {Pixel: 905.5, Value: 3}}) || !reflect.DeepEqual(p.Calibration.Y.Anchors, []peloop.AxisAnchor{{Pixel: 939.5, Value: -45}, {Pixel: 217.5, Value: 45}}) || p.PixelSelectionBound.MaxErrorPX != 2 || p.PixelSelectionBound.Basis == "" {
		t.Fatalf("wrong calibration/bound: %+v / %+v", p.Calibration, p.PixelSelectionBound)
	}
	// Every admitted metadata field must survive the public reader, not just the assertions above.
	var want map[string]any
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil || !reflect.DeepEqual(want, got) {
		t.Fatalf("profile fields lost: %s (%v)", encoded, err)
	}

	for _, tc := range []struct {
		name, path string
		value      any
	}{
		{"schema", "schema", 2},
		{"legacy dataset", "dataset_id", "mdpi2020_hzo_10nm_wakeup"},
		{"wrong DOI", "doi", "10.1002/adma.201404531"},
		{"missing article", "article_url", nil},
		{"missing title", "article", nil},
		{"wrong figure", "figure", "3(b)"},
		{"missing trace", "trace", nil},
		{"wrong trace", "trace", "Annealed BE"},
		{"missing license", "license", nil},
		{"missing license evidence", "license.evidence_url", nil},
		{"missing attribution", "license.attribution", nil},
		{"missing hash", "source_asset.sha256", nil},
		{"wrong hash", "source_asset.sha256", strings.Repeat("a", 64)},
		{"missing asset URL", "source_asset.url", nil},
		{"resized raster", "source_asset.width", 2000},
		{"missing conditions", "conditions", nil},
		{"missing frequency", "conditions.frequency_hz", nil},
		{"wrong stack", "conditions.stack", "TiN/HZO/TiN"},
		{"missing wake-up", "conditions.wake_up_cycles", nil},
		{"missing anneal", "conditions.anneal", nil},
		{"missing temperature status", "conditions.measurement_temperature", nil},
		{"anneal as measurement", "conditions.measurement_temperature", map[string]any{"status": "reported", "detail": "400 C"}},
		{"invented waveform", "conditions.waveform", map[string]any{"status": "reported", "detail": "triangular"}},
		{"invented diameter", "conditions.device_diameter", map[string]any{"status": "reported", "detail": "100 um"}},
		{"missing calibration", "calibration", nil},
		{"missing x anchors", "calibration.x.anchors", nil},
		{"one y anchor", "calibration.y.anchors", []any{map[string]any{"pixel": 939.5, "value": -45}}},
		{"degenerate pixels", "calibration.x.anchors", []any{map[string]any{"pixel": 272.5, "value": -3}, map[string]any{"pixel": 272.5, "value": 3}}},
		{"degenerate values", "calibration.y.anchors", []any{map[string]any{"pixel": 939.5, "value": -45}, map[string]any{"pixel": 217.5, "value": -45}}},
		{"outside raster", "calibration.x.anchors", []any{map[string]any{"pixel": -1, "value": -3}, map[string]any{"pixel": 905.5, "value": 3}}},
		{"wrong x units", "calibration.x.unit", "V/m"},
		{"wrong y units", "calibration.y.unit", "C/m2"},
		{"missing bound", "pixel_selection_bound", nil},
		{"undocumented bound", "pixel_selection_bound.basis", nil},
		{"zero bound", "pixel_selection_bound.max_error_px", 0},
		{"unknown field", "fitted_uncertainty", 0.02},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc map[string]any
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatal(err)
			}
			keys := strings.Split(tc.path, ".")
			obj := doc
			for _, key := range keys[:len(keys)-1] {
				obj = obj[key].(map[string]any)
			}
			obj[keys[len(keys)-1]] = tc.value
			bad, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if result, err := peloop.ReadSourceProfile(bytes.NewReader(bad)); err == nil || !reflect.DeepEqual(result, peloop.SourceProfile{}) {
				t.Fatalf("admitted invalid profile: %+v (%v)", result, err)
			}
		})
	}
	for _, bad := range []string{
		strings.Replace(string(data), "272.5", "1e999", 1),
		strings.Replace(string(data), "939.5", "NaN", 1),
		strings.Replace(string(data), "217.5", "-Infinity", 1),
		string(data) + " {}",
		string(data) + " garbage",
		"null",
	} {
		if _, err := peloop.ReadSourceProfile(strings.NewReader(bad)); err == nil {
			t.Fatalf("admitted malformed/nonfinite/trailing input: %s", bad)
		}
	}
}
