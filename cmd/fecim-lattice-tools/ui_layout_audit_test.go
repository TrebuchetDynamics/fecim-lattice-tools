//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

// NOTE: moduleLifecycle interface already exists in e2e_gui_test.go (same package).
// Reuse it here.

type moduleFactory struct {
	name   string
	create func() (moduleLifecycle, error)
}

func TestLayoutAudit_AllModulesTabsAndSizes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping layout audit in short mode")
	}
	if isHeadlessEnvironment() {
		t.Skip("Skipping layout audit: requires a display. Try: xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit")
	}

	// Use a real app (fonts + layout closer to production). We do not call app.Run();
	// we only build content, show, resize, and capture.
	fy := app.New()
	defer fy.Quit()

	sizes := []struct {
		w, h float32
	}{
		{1200, 800},
		{390, 844},
	}

	modules := []moduleFactory{
		{"hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }},
		{"crossbar", func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() }},
		{"mnist", func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil }},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }},
		{"comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }},
		{"eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }},
	}

	for _, m := range modules {
		m := m
		t.Run(m.name, func(t *testing.T) {
			mod, err := m.create()
			if err != nil {
				t.Fatalf("Failed to create %s module: %v", m.name, err)
			}
			if mod == nil {
				t.Fatalf("%s module is nil", m.name)
			}

			w := fy.NewWindow("LayoutAudit - " + m.name)
			defer w.Close()

			content := mod.BuildContent(fy, w)
			w.SetContent(container.NewMax(content))
			w.Show()

			// IMPORTANT: avoid calling Start() here—some modules spin simulation/animation loops.
			time.Sleep(200 * time.Millisecond)

			for _, sz := range sizes {
				sz := sz
				w.Resize(fyne.NewSize(sz.w, sz.h))
				time.Sleep(200 * time.Millisecond)

				baseName := fmt.Sprintf("layout_%s_%dx%d_base", m.name, int(sz.w), int(sz.h))
				img := captureWindow(w)
				saveTestScreenshot(t, img, baseName)
				verifyImageNotEmpty(t, img, baseName)
				captureOverlays(t, w, content, m.name, int(sz.w), int(sz.h), "base")

				// Traverse all AppTabs (including nested). For each tab set, capture each tab.
				tabSets := findAllAppTabs(content)
				for k, tabs := range tabSets {
					for i := 0; i < len(tabs.Items); i++ {
						tabs.SelectIndex(i)
						time.Sleep(150 * time.Millisecond)
						name := fmt.Sprintf("layout_%s_%dx%d_tabs%d_i%d", m.name, int(sz.w), int(sz.h), k, i)
						img := captureWindow(w)
						saveTestScreenshot(t, img, name)
						verifyImageNotEmpty(t, img, name)
						captureOverlays(t, w, content, m.name, int(sz.w), int(sz.h), fmt.Sprintf("tabs%d_i%d", k, i))
					}
				}
			}
		})
	}
}

// findAllAppTabs walks a CanvasObject tree and returns all *container.AppTabs found.
// It handles:
// - *fyne.Container children
// - *container.AppTabs items' content
// - common wrappers that store a single child in a field named Content/content (via reflection)
func findAllAppTabs(root fyne.CanvasObject) []*container.AppTabs {
	seenObj := map[uintptr]bool{}
	seenTabs := map[uintptr]bool{}
	var out []*container.AppTabs

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		ptr := ptrID(o)
		if ptr != 0 {
			if seenObj[ptr] {
				return
			}
			seenObj[ptr] = true
		}

		if tabs, ok := o.(*container.AppTabs); ok {
			tid := ptrID(tabs)
			if tid == 0 || !seenTabs[tid] {
				if tid != 0 {
					seenTabs[tid] = true
				}
				out = append(out, tabs)
			}
			for _, it := range tabs.Items {
				walk(it.Content)
			}
			return
		}

		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child)
			}
			return
		}

		// Reflection-based fallback for wrappers (e.g., Scroll) that hold a single child.
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content"} {
				f := v.FieldByName(fieldName)
				if f.IsValid() && f.CanInterface() {
					if child, ok := f.Interface().(fyne.CanvasObject); ok {
						walk(child)
					}
				}
			}
		}
	}

	walk(root)
	return out
}

func captureOverlays(t *testing.T, win fyne.Window, root fyne.CanvasObject, module string, w, h int, phase string) {
	t.Helper()

	// Conservative allow-list: text labels that typically open popups/modals.
	allow := map[string]bool{
		"about":    true,
		"help":     true,
		"docs":     true,
		"glossary": true,
		"info":     true,
		"learn":    true,
	}
	closeWords := map[string]bool{
		"close":   true,
		"back":    true,
		"dismiss": true,
		"ok":      true,
		"done":    true,
	}

	buttons := findAllButtons(root)
	seenLabel := map[string]int{}

	for _, b := range buttons {
		label := strings.TrimSpace(b.Text)
		if label == "" {
			continue
		}
		norm := strings.ToLower(label)
		if !allow[norm] {
			continue
		}
		if b.OnTapped == nil {
			continue
		}

		// Trigger overlay
		b.OnTapped()
		time.Sleep(150 * time.Millisecond)

		idx := seenLabel[norm]
		seenLabel[norm] = idx + 1
		name := fmt.Sprintf("layout_%s_%dx%d_overlay_%s_%s_%d", module, w, h, safeName(norm), safeName(phase), idx)
		img := captureWindow(win)
		saveTestScreenshot(t, img, name)
		verifyImageNotEmpty(t, img, name)

		// Best-effort close: look for a close/back/dismiss button and tap it.
		for _, cb := range findAllButtons(root) {
			cl := strings.ToLower(strings.TrimSpace(cb.Text))
			if closeWords[cl] && cb.OnTapped != nil {
				cb.OnTapped()
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func findAllButtons(root fyne.CanvasObject) []*widget.Button {
	seenObj := map[uintptr]bool{}
	seenBtn := map[uintptr]bool{}
	var out []*widget.Button

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		ptr := ptrID(o)
		if ptr != 0 {
			if seenObj[ptr] {
				return
			}
			seenObj[ptr] = true
		}

		if b, ok := o.(*widget.Button); ok {
			bid := ptrID(b)
			if bid == 0 || !seenBtn[bid] {
				if bid != 0 {
					seenBtn[bid] = true
				}
				out = append(out, b)
			}
			return
		}

		if tabs, ok := o.(*container.AppTabs); ok {
			for _, it := range tabs.Items {
				walk(it.Content)
			}
			return
		}

		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child)
			}
			return
		}

		// Reflection-based fallback for wrappers (e.g., Scroll) that hold a single child.
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content"} {
				f := v.FieldByName(fieldName)
				if f.IsValid() && f.CanInterface() {
					if child, ok := f.Interface().(fyne.CanvasObject); ok {
						walk(child)
					}
				}
			}
		}
	}

	walk(root)
	return out
}

func safeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, ":", "-")
	return s
}

func ptrID(o any) uintptr {
	v := reflect.ValueOf(o)
	if !v.IsValid() {
		return 0
	}
	if v.Kind() != reflect.Pointer {
		return 0
	}
	return v.Pointer()
}
