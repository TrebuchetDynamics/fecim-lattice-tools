//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fecim-lattice-tools/module4-circuits/pkg/gui/visual"
)

type vcLegendSpec = visual.VCellLegendSpec

func (ca *CircuitsApp) currentVCLegendSpec() vcLegendSpec {
	maxAbs := 1.0
	readMax := 0.2
	mode := OpModeRead
	if ca != nil && ca.deviceState != nil {
		wr := ca.deviceState.GetWriteRange()
		rr := ca.deviceState.GetReadRange()
		mode = ca.deviceState.GetOperationMode()
		if wr.Max > maxAbs {
			maxAbs = wr.Max
		}
		if rr.Max > 0 {
			readMax = rr.Max
		}
		if mode == OpModeRead || mode == OpModeCompute {
			maxAbs = rr.Max
			if maxAbs <= 0 {
				maxAbs = 1.0
			}
		}
	}
	if maxAbs <= 0 {
		maxAbs = 1.0
	}
	if readMax > maxAbs {
		readMax = maxAbs
	}
	if readMax < 0 {
		readMax = 0
	}

	_ = readMax
	_ = mode

	return visual.NewVCellLegendSpec(maxAbs)
}

func vcOverlayColor(voltage, maxAbs float64) color.RGBA {
	return visual.VCellOverlayColor(voltage, maxAbs)
}

func lerpRGBA(a, b color.RGBA, t float64) color.RGBA {
	return visual.LerpRGBA(a, b, t)
}

func blendRGBA(base, overlay color.RGBA, alpha float64) color.RGBA {
	return visual.BlendRGBA(base, overlay, alpha)
}

func drawVCLegend(img *image.RGBA, x, y, w, h int, spec vcLegendSpec) {
	if w < 40 || h < 6 {
		return
	}

	drawSimpleText(img, spec.Title, x, y-12, color.RGBA{220, 225, 240, 230})

	for i := 0; i < w; i++ {
		v := spec.Min + (spec.Max-spec.Min)*(float64(i)/float64(w-1))
		c := vcOverlayColor(v, math.Max(math.Abs(spec.Min), math.Abs(spec.Max)))
		drawRect(img, x+i, y, 1, h, c)
	}
	drawRectBorder(img, x, y, w, h, color.RGBA{210, 210, 220, 230})

	for i := range spec.Ticks {
		tv := spec.Ticks[i]
		if spec.Max == spec.Min {
			continue
		}
		n := (tv - spec.Min) / (spec.Max - spec.Min)
		tx := x + int(n*float64(w-1))
		drawRect(img, tx, y-2, 1, h+4, color.RGBA{255, 255, 255, 200})
		if i < len(spec.TickText) {
			label := spec.TickText[i]
			drawSimpleText(img, label, tx-len(label)*3, y+h+2, color.RGBA{200, 210, 230, 210})
		}
	}

	drawSimpleText(img, fmt.Sprintf("%+.2fV", spec.Min), x, y+h+12, color.RGBA{140, 180, 255, 220})
	right := fmt.Sprintf("%+.2fV", spec.Max)
	drawSimpleText(img, right, x+w-len(right)*6, y+h+12, color.RGBA{255, 160, 140, 220})
	drawSimpleText(img, spec.SignText, x, y+h+22, color.RGBA{170, 190, 220, 210})
}
