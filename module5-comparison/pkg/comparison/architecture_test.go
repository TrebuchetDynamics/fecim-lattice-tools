package comparison

import "testing"

func TestArchitectures_ReturnsCanonicalSet(t *testing.T) {
	got := Architectures()
	if len(got) != 3 {
		t.Fatalf("Architectures() returned %d entries, want 3", len(got))
	}
	wantNames := []string{"Traditional CPU+DRAM", "GPU Accelerator", "FeCIM CIM"}
	for i, want := range wantNames {
		if got[i] == nil {
			t.Fatalf("Architectures()[%d] is nil", i)
		}
		if got[i].Name != want {
			t.Errorf("Architectures()[%d].Name = %q, want %q", i, got[i].Name, want)
		}
	}
}

func TestArchitectures_ReturnsFreshSliceEachCall(t *testing.T) {
	a := Architectures()
	b := Architectures()
	if &a[0] == &b[0] {
		t.Fatal("Architectures() returned shared backing array; callers could mutate the canonical set")
	}
}
