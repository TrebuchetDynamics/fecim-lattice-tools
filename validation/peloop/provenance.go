// Package peloop provides offline, source-qualified P-E reference tooling.
package peloop

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
)

const (
	kimDOI       = "10.3390/ma13132968"
	kimArticle   = "https://www.mdpi.com/1996-1944/13/13/2968"
	kimAsset     = "https://pub.mdpi-res.com/materials/materials-13-02968/article_deploy/html/images/materials-13-02968-g003.png"
	kimSHA256    = "2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa"
	kimTrace     = "LEFT red circular Ref. markers; RTA after TE dep.; After wake-up"
	kimStack     = "TiN 90 nm / Hf0.5Zr0.5O2 10 nm / TiN 90 nm"
	kimAnnealing = "after top-electrode deposition; no additional bottom-electrode pre-anneal"
)

// SourceProfile schema 1 admits only the qualified Kim Figure 3(a) Ref. source.
// It contains provenance and calibration, never polarization samples or model defaults.
// A new source identity requires an explicit provenance/reader update.
type SourceProfile struct {
	Schema              int                   `json:"schema"`
	DatasetID           string                `json:"dataset_id"`
	DOI                 string                `json:"doi"`
	Article             string                `json:"article"`
	ArticleURL          string                `json:"article_url"`
	Figure              string                `json:"figure"`
	Trace               string                `json:"trace"`
	License             SourceLicense         `json:"license"`
	SourceAsset         SourceAsset           `json:"source_asset"`
	Conditions          MeasurementConditions `json:"conditions"`
	Calibration         Calibration           `json:"calibration"`
	PixelSelectionBound PixelSelectionBound   `json:"pixel_selection_bound"`
}

// SourceLicense records the actual rights evidence and derivative attribution.
type SourceLicense struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	EvidenceURL    string `json:"evidence_url"`
	ContentVersion string `json:"content_version"`
	Attribution    string `json:"attribution"`
}

// SourceAsset identifies original raster bytes, not a resized display or CDN rewrite.
type SourceAsset struct {
	URL         string `json:"url"`
	SHA256      string `json:"sha256"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Bytes       int    `json:"bytes"`
	Acquisition string `json:"acquisition"`
}

// UnreportedField keeps unknown measurement conditions explicit and explained.
// Neither fabrication temperatures nor model constructor defaults fill these fields.
type UnreportedField struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// MeasurementConditions separates fabrication annealing from measurement conditions.
type MeasurementConditions struct {
	Stack                  string          `json:"stack"`
	FrequencyHz            float64         `json:"frequency_hz"`
	WakeUpCycles           int             `json:"wake_up_cycles"`
	WakeUpFieldMVcm        float64         `json:"wake_up_field_MV_cm"`
	Anneal                 Anneal          `json:"anneal"`
	MeasurementTemperature UnreportedField `json:"measurement_temperature"`
	Waveform               UnreportedField `json:"waveform"`
	DeviceDiameter         UnreportedField `json:"device_diameter"`
	Evidence               string          `json:"evidence"`
}

// Anneal is a fabrication step, not a P-E measurement temperature.
type Anneal struct {
	TemperatureC float64 `json:"temperature_C"`
	DurationS    float64 `json:"duration_s"`
	Atmosphere   string  `json:"atmosphere"`
	Sequence     string  `json:"sequence"`
}

// AxisAnchor associates one original-raster tick centre with its printed value.
// Pixel is x for the field axis and y for the polarization axis, using zero-based
// coordinates from the top-left of the original image (y increases downwards).
type AxisAnchor struct {
	Pixel float64 `json:"pixel"`
	Value float64 `json:"value"`
}

// AxisCalibration holds exactly two labelled ticks, not plot-border positions.
type AxisCalibration struct {
	Unit    string       `json:"unit"`
	Anchors []AxisAnchor `json:"anchors"`
}

// Calibration defines independent affine x and y axes; y has a negative slope.
type Calibration struct {
	X AxisCalibration `json:"x"`
	Y AxisCalibration `json:"y"`
}

// PixelSelectionBound is a conservative extraction bound in original pixels.
// It is not a fitted standard deviation or an experimental measurement uncertainty.
type PixelSelectionBound struct {
	MaxErrorPX float64 `json:"max_error_px"`
	Basis      string  `json:"basis"`
}

// ReadSourceProfile reads exactly one schema-1 JSON object and fails closed on
// incomplete, unsupported or inconsistent provenance. It performs no network I/O
// and cannot certify visual trace fidelity or verify externally supplied image bytes.
// On failure it returns a zero profile, never a partially admitted source.
func ReadSourceProfile(r io.Reader) (SourceProfile, error) {
	var p SourceProfile
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return SourceProfile{}, fmt.Errorf("source profile: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return SourceProfile{}, fmt.Errorf("source profile: expected end of JSON input")
	}
	if err := validateSourceProfile(p); err != nil {
		return SourceProfile{}, err
	}
	return p, nil
}

func validateSourceProfile(p SourceProfile) error {
	if p.Schema != 1 || p.DatasetID != "kim2020_fig3a_ref_digitized" || p.DOI != kimDOI || p.ArticleURL != kimArticle || strings.TrimSpace(p.Article) == "" || p.Figure != "3(a)" || p.Trace != kimTrace {
		return fmt.Errorf("source profile: unsupported or missing schema/article/figure/trace identity")
	}
	l := p.License
	if l.Name != "CC BY 4.0" || l.URL != "https://creativecommons.org/licenses/by/4.0/" || l.EvidenceURL != "https://api.crossref.org/works/"+kimDOI || l.ContentVersion != "vor" || strings.TrimSpace(l.Attribution) == "" {
		return fmt.Errorf("source profile: missing or unsupported license evidence/attribution")
	}
	s := p.SourceAsset
	if s.URL != kimAsset || s.SHA256 != kimSHA256 || s.Width != 3629 || s.Height != 1210 || s.Bytes != 961494 || strings.TrimSpace(s.Acquisition) == "" {
		return fmt.Errorf("source profile: original source asset URL/SHA/dimensions/acquisition required")
	}
	c := p.Conditions
	if c.Stack != kimStack || c.FrequencyHz != 10000 || c.WakeUpCycles != 100000 || c.WakeUpFieldMVcm != 2.5 || c.Anneal.TemperatureC != 400 || c.Anneal.DurationS != 60 || c.Anneal.Atmosphere != "N2" || c.Anneal.Sequence != kimAnnealing || strings.TrimSpace(c.Evidence) == "" {
		return fmt.Errorf("source profile: missing or inconsistent source conditions")
	}
	for _, f := range []UnreportedField{c.MeasurementTemperature, c.Waveform, c.DeviceDiameter} {
		if f.Status != "not reported" || strings.TrimSpace(f.Detail) == "" {
			return fmt.Errorf("source profile: measurement temperature/waveform/exact diameter must remain explicitly not reported")
		}
	}
	if err := validateAxis(p.Calibration.X, "MV/cm", -3, 3, s.Width, false); err != nil {
		return err
	}
	if err := validateAxis(p.Calibration.Y, "uC/cm2", -45, 45, s.Height, true); err != nil {
		return err
	}
	b := p.PixelSelectionBound
	if !finite(b.MaxErrorPX) || b.MaxErrorPX <= 0 || strings.TrimSpace(b.Basis) == "" {
		return fmt.Errorf("source profile: documented finite positive pixel-selection bound required")
	}
	return nil
}

func validateAxis(a AxisCalibration, unit string, lo, hi float64, extent int, inverted bool) error {
	if a.Unit != unit || len(a.Anchors) != 2 {
		return fmt.Errorf("source profile: %s requires exactly two labelled tick anchors", unit)
	}
	for _, t := range a.Anchors {
		if !finite(t.Pixel) || !finite(t.Value) || t.Pixel <= 0 || t.Pixel >= float64(extent-1) {
			return fmt.Errorf("source profile: nonfinite or out-of-raster %s tick", unit)
		}
	}
	a0, a1 := a.Anchors[0], a.Anchors[1]
	if a0.Value != lo || a1.Value != hi || a0.Pixel == a1.Pixel || (a1.Pixel < a0.Pixel) != inverted {
		return fmt.Errorf("source profile: degenerate, unlabelled or reversed %s anchors", unit)
	}
	return nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
