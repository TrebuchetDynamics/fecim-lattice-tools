package gui

import "testing"

func TestAccuracySweepRequiresLabelledDataset(t *testing.T) {
	p := &AccuracySweepPanel{app: &DualModeApp{networkCtrl: &NetworkController{}}}
	images, labels := p.getTestSubset(500)
	if len(images) != 0 || len(labels) != 0 {
		t.Fatal("missing dataset must not synthesize accuracy samples")
	}
}
