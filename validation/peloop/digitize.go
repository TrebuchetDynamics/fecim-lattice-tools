package peloop

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
)

const (
	maxPixelBytes = 2 << 20
	maxPixelRows  = 10000
)

// Digitize converts original-raster pixel CSV into physical-unit CSV using the
// admitted profile's independent affine axes. It never extrapolates, fits,
// smooths or fills samples. Rows are emitted ascending then descending, in
// zero-based, contiguous point_index order within each supplied branch.
// At most 10000 data rows and 2MiB of input (including the header) are accepted.
// On any failure it returns no output bytes.
func Digitize(p SourceProfile, r io.Reader) ([]byte, error) {
	if err := validateSourceProfile(p); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(r, maxPixelBytes+1))
	if err != nil {
		return nil, fmt.Errorf("pixels: %w", err)
	}
	if len(data) > maxPixelBytes {
		return nil, fmt.Errorf("pixels: input exceeds 2MiB")
	}
	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil || !slices.Equal(header, []string{"branch", "point_index", "x_px", "y_px"}) {
		return nil, fmt.Errorf("pixels: expected branch,point_index,x_px,y_px header")
	}
	type point struct{ field, polarization float64 }
	branches := map[string]map[int]point{
		"ascending": {}, "descending": {},
	}
	fields := map[string]map[float64]bool{
		"ascending": {}, "descending": {},
	}
	rows := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("pixels: row %d: %w", rows+1, err)
		}
		rows++
		if rows > maxPixelRows {
			return nil, fmt.Errorf("pixels: more than 10000 data rows")
		}
		branch, ok := branches[record[0]]
		if !ok {
			return nil, fmt.Errorf("pixels: row %d: expected ascending or descending branch", rows)
		}
		index, err := strconv.Atoi(record[1])
		if err != nil || index < 0 || index >= maxPixelRows {
			return nil, fmt.Errorf("pixels: row %d: invalid point_index", rows)
		}
		if _, exists := branch[index]; exists {
			return nil, fmt.Errorf("pixels: row %d: repeated point_index", rows)
		}
		x, err := calibratedPixel(record[2], p.Calibration.X)
		if err != nil {
			return nil, fmt.Errorf("pixels: row %d: x: %w", rows, err)
		}
		y, err := calibratedPixel(record[3], p.Calibration.Y)
		if err != nil {
			return nil, fmt.Errorf("pixels: row %d: y: %w", rows, err)
		}
		if fields[record[0]][x] {
			return nil, fmt.Errorf("pixels: row %d: duplicate within-branch field", rows)
		}
		fields[record[0]][x] = true
		branch[index] = point{x, y}
	}
	if rows == 0 {
		return nil, fmt.Errorf("pixels: no points")
	}
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	// bytes.Buffer writes cannot fail; the returned CSV is withheld until all
	// branches have also passed the contiguous-index check.
	_ = writer.Write([]string{"branch", "point_index", "field_MV_cm", "polarization_uC_cm2"})
	for _, name := range []string{"ascending", "descending"} {
		branch := branches[name]
		for i := 0; i < len(branch); i++ {
			pt, ok := branch[i]
			if !ok {
				return nil, fmt.Errorf("pixels: %s: missing point_index %d", name, i)
			}
			_ = writer.Write([]string{name, strconv.Itoa(i), strconv.FormatFloat(pt.field, 'g', -1, 64), strconv.FormatFloat(pt.polarization, 'g', -1, 64)})
		}
	}
	writer.Flush()
	return output.Bytes(), nil
}

func calibratedPixel(raw string, axis AxisCalibration) (float64, error) {
	pixel, err := strconv.ParseFloat(raw, 64)
	a, b := axis.Anchors[0], axis.Anchors[1]
	if err != nil || !finite(pixel) || pixel < math.Min(a.Pixel, b.Pixel) || pixel > math.Max(a.Pixel, b.Pixel) {
		return 0, fmt.Errorf("nonfinite, invalid or outside calibrated ticks")
	}
	value := a.Value + (pixel-a.Pixel)/(b.Pixel-a.Pixel)*(b.Value-a.Value)
	if !finite(value) {
		return 0, fmt.Errorf("nonfinite calibrated value")
	}
	return value, nil
}
