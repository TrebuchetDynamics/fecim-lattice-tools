package design

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestCompositionSnapshot(t *testing.T) {
	c := &Composition{}
	ds := c.Snapshot()
	if ds.Summary == "" {
		t.Error("Snapshot().Summary is empty")
	}
}

func TestCompositionExportWithoutEDA(t *testing.T) {
	c := &Composition{}
	err := c.ExportDesign()
	if err == nil {
		t.Error("ExportDesign without EDA should return error")
	}
}

func TestDesignSnapshotDefaults(t *testing.T) {
	c := &Composition{
		Hysteresis: viewmodel.NewStaticModule(viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleHysteresis,
		}, []viewmodel.Section{}),
	}
	ds := c.Snapshot()
	if ds.Material != "" {
		t.Error("empty hysteresis should produce empty material")
	}
}
