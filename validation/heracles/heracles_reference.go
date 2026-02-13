package heracles

import "math"

// PEPoint is a digitized point from a published P-E branch.
// E is in MV/cm, P is in µC/cm².
type PEPoint struct {
	E_MVcm float64 `json:"E_MV_cm"`
	P_uCcm float64 `json:"P_uC_cm2"`
}

// ReferenceMetrics stores key loop figures-of-merit from the reference curve.
type ReferenceMetrics struct {
	Pr_uCcm2      float64 `json:"Pr_uC_cm2"`
	Ec_MVcm       float64 `json:"Ec_MV_cm"`
	LoopArea_Jm3  float64 `json:"loop_area_J_m3"`
	DigitizeNotes string  `json:"digitize_notes"`
}

// HeraclesReferenceDataset captures approximate 10 nm HfO2, 300 K P-E loop
// points digitized from Heracles paper figures (arXiv:2410.07791).
type HeraclesReferenceDataset struct {
	SourceCitation string           `json:"source_citation"`
	Material       string           `json:"material"`
	TemperatureK   float64          `json:"temperature_K"`
	ThicknessNm    float64          `json:"thickness_nm"`
	Ascending      []PEPoint        `json:"ascending_branch"`
	Descending     []PEPoint        `json:"descending_branch"`
	Metrics        ReferenceMetrics `json:"metrics"`
}

func Reference10nmHfO2_300K() HeraclesReferenceDataset {
	asc := []PEPoint{
		{-3.0, -30.0}, {-2.4, -29.0}, {-1.8, -27.5}, {-1.2, -24.0}, {-0.6, -18.0},
		{0.0, -11.0}, {0.6, -3.0}, {1.2, 8.0}, {1.8, 20.0}, {2.4, 28.0}, {3.0, 31.0},
	}
	desc := []PEPoint{
		{3.0, 30.5}, {2.4, 29.0}, {1.8, 26.0}, {1.2, 21.0}, {0.6, 15.0},
		{0.0, 10.5}, {-0.6, 2.0}, {-1.2, -9.0}, {-1.8, -20.5}, {-2.4, -28.0}, {-3.0, -30.5},
	}

	pr, ec := estimatePrEc(asc, desc)
	area := loopAreaJm3(asc, desc)

	return HeraclesReferenceDataset{
		SourceCitation: "Heracles ferroelectric compact-model publication, arXiv:2410.07791 (digitized from published P-E figures; approximate).",
		Material:       "HfO2 (10 nm)",
		TemperatureK:   300,
		ThicknessNm:    10,
		Ascending:      asc,
		Descending:     desc,
		Metrics: ReferenceMetrics{
			Pr_uCcm2:      pr,
			Ec_MVcm:       ec,
			LoopArea_Jm3:  area,
			DigitizeNotes: "Approximate figure digitization for comparator harness only; not raw Heracles simulator output.",
		},
	}
}

func estimatePrEc(asc, desc []PEPoint) (prUCcm2, ecMVcm float64) {
	prVals := []float64{}
	ecVals := []float64{}

	// Pr from E=0 intersections on both branches.
	for i := 1; i < len(asc); i++ {
		if y0, ok := interpYAtX0(asc[i-1].E_MVcm, asc[i-1].P_uCcm, asc[i].E_MVcm, asc[i].P_uCcm); ok {
			prVals = append(prVals, math.Abs(y0))
		}
	}
	for i := 1; i < len(desc); i++ {
		if y0, ok := interpYAtX0(desc[i-1].E_MVcm, desc[i-1].P_uCcm, desc[i].E_MVcm, desc[i].P_uCcm); ok {
			prVals = append(prVals, math.Abs(y0))
		}
	}

	// Ec from P=0 intersections.
	for i := 1; i < len(asc); i++ {
		if x0, ok := interpXAtY0(asc[i-1].P_uCcm, asc[i-1].E_MVcm, asc[i].P_uCcm, asc[i].E_MVcm); ok {
			ecVals = append(ecVals, math.Abs(x0))
		}
	}
	for i := 1; i < len(desc); i++ {
		if x0, ok := interpXAtY0(desc[i-1].P_uCcm, desc[i-1].E_MVcm, desc[i].P_uCcm, desc[i].E_MVcm); ok {
			ecVals = append(ecVals, math.Abs(x0))
		}
	}

	for _, v := range prVals {
		prUCcm2 += v
	}
	if len(prVals) > 0 {
		prUCcm2 /= float64(len(prVals))
	}
	for _, v := range ecVals {
		ecMVcm += v
	}
	if len(ecVals) > 0 {
		ecMVcm /= float64(len(ecVals))
	}
	return prUCcm2, ecMVcm
}

func loopAreaJm3(asc, desc []PEPoint) float64 {
	path := make([]PEPoint, 0, len(asc)+len(desc))
	path = append(path, asc...)
	path = append(path, desc...)
	if len(path) < 2 {
		return 0
	}
	var area float64
	for i := 1; i < len(path); i++ {
		e1 := path[i-1].E_MVcm * 1e8
		e2 := path[i].E_MVcm * 1e8
		p1 := path[i-1].P_uCcm * 1e-2
		p2 := path[i].P_uCcm * 1e-2
		area += 0.5 * (p1 + p2) * (e2 - e1)
	}
	return math.Abs(area)
}

func interpXAtY0(y1, x1, y2, x2 float64) (float64, bool) {
	dy := y2 - y1
	if dy == 0 {
		return 0, false
	}
	f := -y1 / dy
	if f < 0 || f > 1 {
		return 0, false
	}
	return x1 + f*(x2-x1), true
}

func interpYAtX0(x1, y1, x2, y2 float64) (float64, bool) {
	dx := x2 - x1
	if dx == 0 {
		return 0, false
	}
	f := -x1 / dx
	if f < 0 || f > 1 {
		return 0, false
	}
	return y1 + f*(y2-y1), true
}
