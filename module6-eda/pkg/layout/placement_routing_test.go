package layout

import "testing"

func TestPlaceForceDirected_Basic(t *testing.T) {
	macros := []MacroBlock{
		{Name: "XBAR0", Width: 200, Height: 200},
		{Name: "XBAR1", Width: 200, Height: 200},
		{Name: "DAC", Width: 150, Height: 150},
		{Name: "ADC", Width: 150, Height: 150},
	}
	nets := []Net{
		{Name: "n0", Nodes: []string{"XBAR0", "DAC"}},
		{Name: "n1", Nodes: []string{"XBAR0", "ADC"}},
		{Name: "n2", Nodes: []string{"XBAR0", "XBAR1"}},
	}

	placed := PlaceForceDirected(macros, nets, 2000, 2000, 10, 10, 60)
	if len(placed) != len(macros) {
		t.Fatalf("expected %d placed macros, got %d", len(macros), len(placed))
	}

	for _, m := range macros {
		p, ok := placed[m.Name]
		if !ok {
			t.Fatalf("missing macro placement for %s", m.Name)
		}
		if p.X < 0 || p.Y < 0 {
			t.Fatalf("invalid negative placement for %s: %+v", m.Name, p)
		}
		if p.X+m.Width > 2000 || p.Y+m.Height > 2000 {
			t.Fatalf("placement out of bounds for %s: %+v", m.Name, p)
		}
	}

	for i := 0; i < len(macros); i++ {
		for j := i + 1; j < len(macros); j++ {
			ma, mb := macros[i], macros[j]
			pa, pb := placed[ma.Name], placed[mb.Name]
			if rectOverlap(pa.X, pa.Y, ma.Width, ma.Height, pb.X, pb.Y, mb.Width, mb.Height) {
				t.Fatalf("overlap detected between %s and %s", ma.Name, mb.Name)
			}
		}
	}
}

func TestRouteManhattan_Basic(t *testing.T) {
	macros := []MacroBlock{
		{Name: "A", Width: 200, Height: 200},
		{Name: "B", Width: 200, Height: 200},
		{Name: "OBS", Width: 200, Height: 200},
	}
	placements := map[string]Placement{
		"A":   {X: 0, Y: 0},
		"B":   {X: 1200, Y: 0},
		"OBS": {X: 500, Y: 0},
	}
	nets := []Net{{Name: "nAB", Nodes: []string{"A", "B"}}}

	routes, err := RouteManhattan(macros, placements, nets, 100)
	if err != nil {
		t.Fatalf("RouteManhattan failed: %v", err)
	}
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if len(routes[0].Segments) == 0 {
		t.Fatal("expected non-empty route segments")
	}

	// Ensure every segment is Manhattan.
	for _, s := range routes[0].Segments {
		if s.X1 != s.X2 && s.Y1 != s.Y2 {
			t.Fatalf("non-Manhattan segment found: %+v", s)
		}
	}
}
