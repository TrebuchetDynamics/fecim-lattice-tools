// pkg/gui/tabs/conductance_heatmap.go
// canvas.Raster-based conductance heatmap for the Builder tab's Array Map preview.
// Pattern reuses the hot-colormap approach from module1-hysteresis/pkg/gui/forc_panel.go.
package tabs

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// conductanceMatrix holds normalized conductance values in [0,1] for each cell.
type conductanceMatrix struct {
	Rows, Cols int
	Values     [][]float64
}

// heatColorCIM maps a normalized value in [0,1] to a scientific hot colormap color.
// Gradient: dark-blue (low) → cyan → green → yellow (high).
// Matches the FORC density colormap in module1-hysteresis/pkg/gui/forc_panel.go.
func heatColorCIM(v float64) color.RGBA {
	if v <= 0 {
		return color.RGBA{0, 0, 80, 255}
	}
	if v >= 1 {
		return color.RGBA{255, 255, 0, 255}
	}
	switch {
	case v < 0.25:
		t := v / 0.25
		return color.RGBA{0, uint8(t * 255), 255, 255}
	case v < 0.5:
		t := (v - 0.25) / 0.25
		return color.RGBA{0, 255, uint8((1 - t) * 255), 255}
	case v < 0.75:
		t := (v - 0.5) / 0.25
		return color.RGBA{uint8(t * 255), 255, 0, 255}
	default:
		t := (v - 0.75) / 0.25
		return color.RGBA{255, uint8((1 - t) * 255), 0, 255}
	}
}

func makeConductanceUniform(rows, cols int, level float64) conductanceMatrix {
	m := conductanceMatrix{Rows: rows, Cols: cols, Values: make([][]float64, rows)}
	for i := range m.Values {
		m.Values[i] = make([]float64, cols)
		for j := range m.Values[i] {
			m.Values[i][j] = level
		}
	}
	return m
}

func makeConductanceGradient(rows, cols int) conductanceMatrix {
	m := conductanceMatrix{Rows: rows, Cols: cols, Values: make([][]float64, rows)}
	cx := float64(cols-1) / 2
	cy := float64(rows-1) / 2
	maxDist := math.Sqrt(cx*cx + cy*cy)
	if maxDist == 0 {
		maxDist = 1
	}
	for i := range m.Values {
		m.Values[i] = make([]float64, cols)
		for j := range m.Values[i] {
			dx := float64(j) - cx
			dy := float64(i) - cy
			dist := math.Sqrt(dx*dx+dy*dy) / maxDist
			m.Values[i][j] = 1.0 - dist
		}
	}
	return m
}

func makeConductanceRandom(rows, cols int, seed int64) conductanceMatrix {
	rng := rand.New(rand.NewSource(seed))
	m := conductanceMatrix{Rows: rows, Cols: cols, Values: make([][]float64, rows)}
	for i := range m.Values {
		m.Values[i] = make([]float64, cols)
		for j := range m.Values[i] {
			m.Values[i][j] = rng.Float64()
		}
	}
	return m
}

func makeConductanceChecker(rows, cols int) conductanceMatrix {
	m := conductanceMatrix{Rows: rows, Cols: cols, Values: make([][]float64, rows)}
	for i := range m.Values {
		m.Values[i] = make([]float64, cols)
		for j := range m.Values[i] {
			if (i+j)%2 == 0 {
				m.Values[i][j] = 0.9
			} else {
				m.Values[i][j] = 0.1
			}
		}
	}
	return m
}

// newConductanceRaster creates a canvas.Raster rendering the conductance matrix
// using the hot colormap (blue=low, yellow=high). Row 0 is at the bottom.
func newConductanceRaster(m conductanceMatrix) *canvas.Raster {
	if m.Rows == 0 || m.Cols == 0 {
		return canvas.NewRaster(func(w, h int) image.Image {
			return image.NewRGBA(image.Rect(0, 0, w, h))
		})
	}
	return canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for py := 0; py < h; py++ {
			row := m.Rows - 1 - int(float64(py)*float64(m.Rows)/float64(h))
			if row < 0 {
				row = 0
			}
			if row >= m.Rows {
				row = m.Rows - 1
			}
			for px := 0; px < w; px++ {
				col := int(float64(px) * float64(m.Cols) / float64(w))
				if col >= m.Cols {
					col = m.Cols - 1
				}
				img.Set(px, py, heatColorCIM(m.Values[row][col]))
			}
		}
		return img
	})
}

// newHeatmapLegend creates a horizontal color scale legend bar.
func newHeatmapLegend() *canvas.Raster {
	r := canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for px := 0; px < w; px++ {
			c := heatColorCIM(float64(px) / math.Max(float64(w-1), 1))
			for py := 0; py < h; py++ {
				img.Set(px, py, c)
			}
		}
		return img
	})
	r.SetMinSize(fyne.NewSize(280, 14))
	return r
}

// MakeConductanceHeatmapPanel creates an interactive conductance heatmap panel.
// Shows simulated conductance distributions across the crossbar to help designers
// visualize spatial variation patterns. Patterns are synthetic — not from device state.
func MakeConductanceHeatmapPanel(cfg *config.ArrayConfig) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 8, Cols: 8}
	}

	patterns := []string{"Gradient", "Random", "Checkerboard", "Uniform Hi", "Uniform Lo"}
	patternSelect := widget.NewSelect(patterns, nil)
	patternSelect.SetSelected("Gradient")

	infoLabel := widget.NewLabel("")

	// Stack container: swap rasters on each refresh without rebuilding the layout.
	placeholder := newConductanceRaster(conductanceMatrix{})
	placeholder.SetMinSize(fyne.NewSize(360, 270))
	rasterContainer := container.NewStack(placeholder)

	updateRaster := func() {
		rows := cfg.Rows
		cols := cfg.Cols
		if rows <= 0 {
			rows = 8
		}
		if cols <= 0 {
			cols = 8
		}

		var m conductanceMatrix
		switch patternSelect.Selected {
		case "Random":
			m = makeConductanceRandom(rows, cols, time.Now().UnixNano())
		case "Checkerboard":
			m = makeConductanceChecker(rows, cols)
		case "Uniform Hi":
			m = makeConductanceUniform(rows, cols, 0.9)
		case "Uniform Lo":
			m = makeConductanceUniform(rows, cols, 0.1)
		default: // Gradient
			m = makeConductanceGradient(rows, cols)
		}

		r := newConductanceRaster(m)
		r.SetMinSize(fyne.NewSize(360, 270))
		rasterContainer.Objects = []fyne.CanvasObject{r}
		rasterContainer.Refresh()

		infoLabel.SetText(fmt.Sprintf(
			"Array: %d × %d  |  Total cells: %d  |  Pattern: %s",
			rows, cols, rows*cols, patternSelect.Selected,
		))
	}

	patternSelect.OnChanged = func(string) { updateRaster() }
	refreshBtn := widget.NewButton("Refresh", func() { updateRaster() })

	updateRaster()

	legend := newHeatmapLegend()
	legendRow := container.NewHBox(
		widget.NewLabel("G_min"),
		legend,
		widget.NewLabel("G_max"),
	)

	descLabel := widget.NewLabelWithStyle(
		"Note: Patterns are illustrative (not from device state). Blue = low conductance, Yellow = high.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)
	descLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Pattern:"), patternSelect, refreshBtn),
		infoLabel,
		container.NewCenter(rasterContainer),
		legendRow,
		descLabel,
	)
}
