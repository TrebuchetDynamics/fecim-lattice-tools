//go:build legacy_fyne

package uilog

import "testing"

func TestLogger_IsStableSingleton(t *testing.T) {
	first := Logger()
	second := Logger()
	if first == nil {
		t.Fatal("expected logger")
	}
	if first != second {
		t.Fatal("expected stable logger singleton")
	}
}
