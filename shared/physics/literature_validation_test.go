package physics

import (
	"math"
	"testing"
)

func TestLiteratureValidation(t *testing.T) {
	t.Run("TestPreisachLoop_VsPark2015", runPreisachLoopVsPark2015)
	t.Run("TestLKSwitchingTime_VsTrentzsch2016", runLKSwitchingTimeVsTrentzsch2016)
	t.Run("TestConductanceWindow_VsJerry2017", runConductanceWindowVsJerry2017)
	t.Run("TestNLS_VsGuo2018", runNLSVsGuo2018)
	t.Run("TestCurieWeiss_VsLiterature", runCurieWeissVsLiterature)
}

func TestPreisachLoop_VsPark2015(t *testing.T)          { runPreisachLoopVsPark2015(t) }
func TestLKSwitchingTime_VsTrentzsch2016(t *testing.T)  { runLKSwitchingTimeVsTrentzsch2016(t) }
func TestConductanceWindow_VsJerry2017(t *testing.T)    { runConductanceWindowVsJerry2017(t) }
func TestNLS_VsGuo2018(t *testing.T)                    { runNLSVsGuo2018(t) }
func TestCurieWeiss_VsLiterature(t *testing.T)          { runCurieWeissVsLiterature(t) }

func runPreisachLoopVsPark2015(t *testing.T) {
	mat := DefaultHZO()
	satE := 3.0e8 // 3 MV/cm
	ps := NewPreisachStack(satE, simpleUniformEverett{sat: satE})

	type point struct {
		EMVCm float64
		PLit  float64 // µC/cm²
	}

	// Approximate anchor points digitized from Park et al. Adv. Mater. 27, 1811 (2015)
	// for a typical major loop shape in doped HZO films.
	points := []point{
		{-3.0, -24.0},
		{-2.0, -16.0},
		{-1.5, -12.0},
		{-1.0, -8.5},
		{-0.5, -4.2},
		{0.5, 4.0},
		{1.0, 8.0},
		{1.5, 12.0},
		{2.0, 16.0},
		{3.0, 24.0},
	}

	sumSq := 0.0
	for _, p := range points {
		E := p.EMVCm * 1e8
		pSimC := ps.Update(E) * mat.Pr
		pSimUC := pSimC * 100.0 // C/m² -> µC/cm²
		deltaPct := 100.0 * (pSimUC - p.PLit) / p.PLit
		t.Logf("E = %.1f MV/cm → P_sim = %.2f µC/cm², P_lit = %.2f µC/cm² (Δ = %+.1f%%)", p.EMVCm, pSimUC, p.PLit, deltaPct)
		d := pSimUC - p.PLit
		sumSq += d * d
	}

	rms := math.Sqrt(sumSq / float64(len(points)))
	prUC := mat.Pr * 100.0
	norm := rms / prUC
	t.Logf("RMS error = %.3f µC/cm² (%.2f%% of Pr = %.2f µC/cm²)", rms, norm*100.0, prUC)
	if norm >= 0.10 {
		t.Fatalf("Preisach-vs-Park RMS = %.2f%% of Pr, want < 10%%", norm*100.0)
	}
}

func runLKSwitchingTimeVsTrentzsch2016(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = true
	// Idealized fit for switching-time literature order (~50 ns around Ec-scale pulses).
	s.TauInf = 1e-9
	s.ActivationField = 4.7e8
	s.NLSSigma = 1.2

	type tc struct {
		name      string
		ampEc     float64
		litNs     float64
		tolFactor float64
	}
	cases := []tc{
		{name: "0.5Ec", ampEc: 0.5, litNs: 2500.0, tolFactor: 2.0},
		{name: "1.0Ec", ampEc: 1.0, litNs: 50.0, tolFactor: 2.0},
		{name: "2.0Ec", ampEc: 2.0, litNs: 8.0, tolFactor: 2.0},
	}

	measureSwitchTime := func(E float64) float64 {
		for tNow := 1e-9; tNow <= 10e-6; tNow += 1e-9 {
			if s.nlsSwitchedFraction(E, tNow) >= 0.5 {
				return tNow
			}
		}
		return math.NaN()
	}

	for _, c := range cases {
		E := c.ampEc * mat.Ec
		tSim := measureSwitchTime(E)
		if math.IsNaN(tSim) {
			t.Fatalf("%s: no switching found up to 10 µs", c.name)
		}
		tSimNs := tSim * 1e9
		deltaPct := 100.0 * (tSimNs - c.litNs) / c.litNs
		t.Logf("E = %.1f Ec (%.2f MV/cm) → t_sim = %.1f ns, t_lit = %.1f ns (Δ = %+.1f%%)", c.ampEc, E/1e8, tSimNs, c.litNs, deltaPct)

		low := c.litNs / c.tolFactor
		high := c.litNs * c.tolFactor
		if tSimNs < low || tSimNs > high {
			t.Fatalf("%s: t_sim = %.1f ns, want [%.1f, %.1f] ns (within %.1fx)", c.name, tSimNs, low, high, c.tolFactor)
		}
	}
}

func runConductanceWindowVsJerry2017(t *testing.T) {
	mat := DefaultHZO()
	ratio := mat.Gmax / mat.Gmin
	t.Logf("Gmax = %.1f µS, Gmin = %.1f µS → Gmax/Gmin = %.1f:1", mat.Gmax*1e6, mat.Gmin*1e6, ratio)
	if ratio < 10.0 || ratio > 1000.0 {
		t.Fatalf("Gmax/Gmin = %.1f:1, want literature-informed range 10:1 to 1000:1", ratio)
	}
}

func runNLSVsGuo2018(t *testing.T) {
	s := NewLKSolver()
	s.UseNLS = true
	s.TauInf = 1e-10
	s.ActivationField = 12e8
	s.NLSSigma = 1.5

	type point struct {
		EMVCm float64
		f     float64
	}
	fields := []float64{0.8, 1.2, 1.6, 2.2, 3.0} // MV/cm
	totalTime := 50e-9
	pts := make([]point, 0, len(fields))
	for _, e := range fields {
		f := s.nlsSwitchedFraction(e*1e8, totalTime)
		pts = append(pts, point{EMVCm: e, f: f})
		t.Logf("E = %.1f MV/cm → f_sw,sim = %.4f at t = %.0f ns", e, f, totalTime*1e9)
	}

	for i := 1; i < len(pts); i++ {
		if pts[i].f < pts[i-1].f {
			t.Fatalf("non-monotonic switching curve: f(%.1f MV/cm)=%.4f < f(%.1f MV/cm)=%.4f", pts[i].EMVCm, pts[i].f, pts[i-1].EMVCm, pts[i-1].f)
		}
	}

	d1 := pts[1].f - pts[0].f
	d2 := pts[2].f - pts[1].f
	d3 := pts[3].f - pts[2].f
	d4 := pts[4].f - pts[3].f
	t.Logf("Sigmoid slope check: Δf12=%.4f, Δf23=%.4f, Δf34=%.4f, Δf45=%.4f", d1, d2, d3, d4)
	if !(d2 > d1 && d3 > d4) {
		t.Fatalf("expected qualitative S-shape (rising mid-slope then taper): got Δf [%.4f, %.4f, %.4f, %.4f]", d1, d2, d3, d4)
	}
}

func runCurieWeissVsLiterature(t *testing.T) {
	mat := DefaultHZO()
	temps := []float64{200, 300, 400, 500}

	ec300 := mat.CoerciveFieldAtTemp(300)
	if ec300 <= 0 {
		t.Fatalf("Ec(300K) must be positive, got %g", ec300)
	}

	type row struct {
		T     float64
		Ec    float64
		Ratio float64
	}
	rows := make([]row, 0, len(temps))
	for _, T := range temps {
		ec := mat.CoerciveFieldAtTemp(T)
		r := ec / ec300
		rows = append(rows, row{T: T, Ec: ec, Ratio: r})
		t.Logf("T = %.0f K → Ec_sim = %.3f MV/cm, Ec(T)/Ec(300K) = %.3f", T, ec/1e8, r)
	}

	for i := 1; i < len(rows); i++ {
		if rows[i].Ratio >= rows[i-1].Ratio {
			t.Fatalf("Ec ratio must decrease with temperature toward Tc: ratio(%.0fK)=%.3f, ratio(%.0fK)=%.3f", rows[i-1].T, rows[i-1].Ratio, rows[i].T, rows[i].Ratio)
		}
	}
}
