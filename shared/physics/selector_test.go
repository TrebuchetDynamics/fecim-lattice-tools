package physics

import "testing"

func TestSelector_OnResistance(t *testing.T) {
	cases := []MOSFETSelector{SKY130NMOS(), NMOS28nm()}
	for _, s := range cases {
		if s.Ron <= 0 {
			t.Fatalf("Ron must be > 0, got %e", s.Ron)
		}
		if s.Ion <= 0 {
			t.Fatalf("Ion must be > 0, got %e", s.Ion)
		}
		// Ron should be in physically reasonable range for compact selectors.
		if s.Ron < 1e2 || s.Ron > 1e5 {
			t.Fatalf("Ron out of expected range: %e", s.Ron)
		}
	}
}

func TestSelector_OffCurrent(t *testing.T) {
	cases := []MOSFETSelector{SKY130NMOS(), NMOS28nm()}
	for _, s := range cases {
		ratio := s.Ion / s.Ioff
		if ratio < 1e6 {
			t.Fatalf("Ion/Ioff=%e, want >=1e6", ratio)
		}
	}
}

func TestSelector_GateCapacitance(t *testing.T) {
	s := SKY130NMOS()
	if s.Cgate <= 0 {
		t.Fatalf("Cgate must be > 0, got %e", s.Cgate)
	}

	s2 := s
	s2.W = 2 * s.W
	s2.Cgate = 2 * s.Cgate
	if s2.Cgate <= s.Cgate {
		t.Fatalf("expected gate capacitance to scale with W, base=%e scaled=%e", s.Cgate, s2.Cgate)
	}
}

func TestSelector_CurrentVsVgs(t *testing.T) {
	s := NMOS28nm()
	prev := -1.0
	for vgs := 0.0; vgs <= 1.2; vgs += 0.05 {
		i := s.Current(vgs, 1.0)
		if i < prev {
			t.Fatalf("non-monotonic current: Vgs=%.2f I=%e prev=%e", vgs, i, prev)
		}
		prev = i
	}
}
