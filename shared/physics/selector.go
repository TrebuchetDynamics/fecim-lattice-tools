package physics

import "math"

const thermalVoltage300K = 0.02585 // V

// MOSFETSelector is a compact CMOS selector model for array-level estimations.
type MOSFETSelector struct {
	W, L  float64 // gate width/length (m)
	Vth   float64 // threshold voltage (V)
	Ion   float64 // on-current (A) at Vgs=Vdd, Vds=Vdd
	Ioff  float64 // off-current (A) at Vgs=0
	Cgate float64 // gate capacitance (F)
	Ron   float64 // on-resistance (ohm) = Vdd/Ion
}

func newMOSFETSelector(w, l, vth, ion, ioff, cgate, vdd float64) MOSFETSelector {
	s := MOSFETSelector{
		W:     w,
		L:     l,
		Vth:   vth,
		Ion:   ion,
		Ioff:  ioff,
		Cgate: cgate,
	}
	if ion > 0 && vdd > 0 {
		s.Ron = vdd / ion
	}
	return s
}

// SKY130NMOS returns a SKY130-like NMOS selector preset.
func SKY130NMOS() MOSFETSelector {
	return newMOSFETSelector(0.42e-6, 0.15e-6, 0.45, 200e-6, 10e-12, 0.3e-15, 1.8)
}

// NMOS28nm returns a 28nm-like NMOS selector preset.
func NMOS28nm() MOSFETSelector {
	return newMOSFETSelector(0.30e-6, 0.03e-6, 0.35, 400e-6, 100e-12, 0.1e-15, 1.0)
}

// Current returns drain current magnitude (A) at the provided terminal biases.
func (s *MOSFETSelector) Current(Vgs, Vds float64) float64 {
	if s == nil {
		return 0
	}
	if Vds == 0 {
		return 0
	}
	vds := math.Abs(Vds)

	if Vgs <= 0 {
		return s.Ioff
	}
	if Vgs <= s.Vth {
		n := 1.5
		iSub := s.Ioff * math.Exp((Vgs-s.Vth)/(n*thermalVoltage300K))
		if iSub < s.Ioff {
			iSub = s.Ioff
		}
		return iSub
	}

	vov := Vgs - s.Vth
	vovRef := 0.7 // calibrated so Vgs~Vth+0.7 gives Ion
	if s.Ion <= 0 || vovRef <= 0 {
		return 0
	}
	k := 2.0 * s.Ion / (vovRef * vovRef)

	var id float64
	if vds < vov {
		id = k * (vov*vds - 0.5*vds*vds)
	} else {
		id = 0.5 * k * vov * vov
	}
	if id < s.Ioff {
		id = s.Ioff
	}
	if id > s.Ion {
		id = s.Ion
	}
	return id
}

// Conductance returns small-signal-equivalent conductance (S) at (Vgs, Vds).
func (s *MOSFETSelector) Conductance(Vgs, Vds float64) float64 {
	if s == nil {
		return 0
	}
	vds := math.Abs(Vds)
	if vds < 1e-9 {
		vds = 1e-9
	}
	return s.Current(Vgs, Vds) / vds
}

// SeriesConductance combines two conductances in series.
func SeriesConductance(cellG, selectorG float64) float64 {
	if cellG <= 0 || selectorG <= 0 {
		return 0
	}
	return 1.0 / (1.0/cellG + 1.0/selectorG)
}
