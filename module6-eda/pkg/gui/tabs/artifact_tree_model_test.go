package tabs

import (
	"reflect"
	"testing"
)

func TestArtifactTreeModelUsesStableFullPathIDs(t *testing.T) {
	model := NewArtifactTreeModel([]string{
		"runs/a/fecim_array.v",
		"runs/b/fecim_array.v",
		"runs/a/reports/timing.rpt",
	})

	if got, want := model.ChildUIDs(""), []string{"runs"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("root children = %#v, want %#v", got, want)
	}

	if !model.IsBranch("runs/a") {
		t.Fatal("runs/a should be a branch")
	}
	if model.IsBranch("runs/a/fecim_array.v") {
		t.Fatal("runs/a/fecim_array.v should be a leaf")
	}

	wantRunA := []string{"runs/a/fecim_array.v", "runs/a/reports"}
	if got := model.ChildUIDs("runs/a"); !reflect.DeepEqual(got, wantRunA) {
		t.Fatalf("runs/a children = %#v, want %#v", got, wantRunA)
	}

	if got := model.Label("runs/a/fecim_array.v"); got != "fecim_array.v" {
		t.Fatalf("leaf label = %q, want fecim_array.v", got)
	}
	if got := model.Label("runs/b/fecim_array.v"); got != "fecim_array.v" {
		t.Fatalf("duplicate leaf label = %q, want fecim_array.v", got)
	}
	if model.ID("runs/a/fecim_array.v") == model.ID("runs/b/fecim_array.v") {
		t.Fatal("duplicate basenames must keep unique full-path IDs")
	}
}

func TestArtifactTreeModelNormalizesAndSortsPaths(t *testing.T) {
	model := NewArtifactTreeModel([]string{
		"./exports/z.spice",
		"exports/a.v",
		"exports/lib/fecim.lib",
		"",
	})

	want := []string{"exports/a.v", "exports/lib", "exports/z.spice"}
	if got := model.ChildUIDs("exports"); !reflect.DeepEqual(got, want) {
		t.Fatalf("exports children = %#v, want %#v", got, want)
	}
}
