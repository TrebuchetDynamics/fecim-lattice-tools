package main

import "testing"

func TestParseCSVLoopAcceptsSyntheticHeader(t *testing.T) {
	e, p, err := parseCSVLoop([]byte("# SYNTHETIC, not experimental\n# generator: Preisach sweep\nE_MV_cm,P_uC_cm2\n-3,-19.4\n3,19.4\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(e) != 2 || len(p) != 2 || e[0] != -3 || p[1] != 19.4 {
		t.Fatalf("unexpected points: %v %v", e, p)
	}
}
