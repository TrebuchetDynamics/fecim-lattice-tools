package literature

import (
	"path/filepath"
	"testing"
)

func TestThresholdPolicyForDataset_ClassificationIsExplicit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		materialID string
		mode       string
		provenance string
	}{
		{materialID: "park2015_hzo_10nm", mode: "calibrated_relaxed", provenance: filepath.Join("data", "park2015_fig2a_hzo_10nm.provenance.json")},
		{materialID: "cheema2020_superlattice_5nm", mode: "strict", provenance: filepath.Join("data", "cheema2020_fig2c_hzo_superlattice_5nm.provenance.json")},
		{materialID: "mdpi2020_hzo_10nm_wakeup", mode: "strict", provenance: filepath.Join("data", "mdpi2020_ma13132968_fig3a_hzo_10nm_wakeup.provenance.json")},
		{materialID: "alscn2022_pmc9607415_fig6a_pt_200nm", mode: "placeholder_relaxed", provenance: filepath.Join("data", "alscn2022_pmc9607415_fig6a_pt_200nm.provenance.json")},
		{materialID: "alscn2022_pmc9607415_fig6b_mo_200nm", mode: "strict", provenance: filepath.Join("data", "alscn2022_pmc9607415_fig6b_mo_200nm.provenance.json")},
		{materialID: "pzt2024_nano14050432_fig2_thinfilm", mode: "placeholder_relaxed", provenance: filepath.Join("data", "pzt2024_nano14050432_fig2_thinfilm.provenance.json")},
		{materialID: "pzt2024_nano14050432_fig2_thinfilm_traceB", mode: "placeholder_relaxed", provenance: filepath.Join("data", "pzt2024_nano14050432_fig2_thinfilm_traceB.provenance.json")},
		{materialID: "bto2021_cryst11101192_hysteresis", mode: "placeholder_relaxed", provenance: filepath.Join("data", "bto2021_cryst11101192_hysteresis.provenance.json")},
		{materialID: "bto2021_cryst11101192_hysteresis_digitized", mode: "uncertainty_relaxed", provenance: filepath.Join("data", "bto2021_cryst11101192_hysteresis_digitized.provenance.json")},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.materialID, func(t *testing.T) {
			t.Parallel()
			ds := peLoopDataset{MaterialID: tc.materialID, Provenance: tc.provenance}
			decision, err := thresholdPolicyForDataset(ds, 10.0, 1.0, 10.0)
			if err != nil {
				t.Fatalf("thresholdPolicyForDataset returned error: %v", err)
			}
			if decision.Mode != tc.mode {
				t.Fatalf("policy mode mismatch for %s: got %q want %q", tc.materialID, decision.Mode, tc.mode)
			}

			if tc.mode == "strict" {
				if decision.PrPct != thPrPct || decision.EcPct != thEcPct || decision.RMSEps != thRMSEps || decision.AreaPct != thAreaPct {
					t.Fatalf("strict policy drift for %s: got pr=%v ec=%v rmse=%v area=%v", tc.materialID, decision.PrPct, decision.EcPct, decision.RMSEps, decision.AreaPct)
				}
				return
			}

			// Non-strict policies must relax at least one threshold versus baseline.
			relaxed := decision.PrPct > thPrPct || decision.EcPct > thEcPct || decision.RMSEps > thRMSEps || decision.AreaPct > thAreaPct
			if !relaxed {
				t.Fatalf("non-strict policy did not relax any threshold for %s", tc.materialID)
			}
		})
	}
}
