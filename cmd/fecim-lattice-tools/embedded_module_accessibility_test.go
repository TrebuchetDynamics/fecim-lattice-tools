//go:build cgo

package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
	"fecim-lattice-tools/shared/themes"
)

func TestEmbeddedModulesLayoutAccessibilitySmoke(t *testing.T) {
	t.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	cases := []struct {
		name  string
		theme fyne.Theme
	}{
		{name: "dark", theme: themes.GetTheme(themes.ThemeDark)},
		{name: "high-contrast-large-text", theme: themes.NewScaledTheme(themes.GetTheme(themes.ThemeHighContrast), 1.35)},
	}

	sizes := []fyne.Size{
		fyne.NewSize(1024, 768),
		fyne.NewSize(1400, 900),
	}

	for _, module := range newLifecycleModulesForTest(t) {
		for _, themeCase := range cases {
			for _, size := range sizes {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", module.name, themeCase.name, int(size.Width), int(size.Height)), func(t *testing.T) {
					app := fyneTest.NewTempApp(t)
					app.Settings().SetTheme(themeCase.theme)
					window := app.NewWindow("layout-" + module.name)
					defer window.Close()

					content := module.module.BuildContent(app, window)
					if content == nil {
						t.Fatal("BuildContent returned nil")
					}
					// Match the production shell: size the window before content layout so
					// large-text themes do not force first layout through a zero-size canvas.
					window.Resize(size)
					window.SetContent(content)
					window.Show()
					fyne.DoAndWait(func() {
						window.Canvas().Refresh(content)
					})
					time.Sleep(25 * time.Millisecond)

					for _, node := range flattenLayoutNodes(content) {
						if node.obj == nil || !node.obj.Visible() {
							continue
						}
						if _, ok := node.obj.(fyne.Widget); ok {
							s := node.obj.Size()
							if s.Width <= 0 || s.Height <= 0 {
								t.Fatalf("zero-size visible widget %T", node.obj)
							}
						}
						if label, ok := node.obj.(*widget.Label); ok && isLongUnboundedLabel(label) {
							t.Fatalf("long visible label lacks wrapping/truncation: %q", label.Text)
						}
					}
				})
			}
		}
	}
}

type testLifecycleModule struct {
	name   string
	module lifecycleModule
}

func newLifecycleModulesForTest(t *testing.T) []testLifecycleModule {
	t.Helper()
	crossbar, err := demo2gui.NewEmbeddedCrossbarApp()
	if err != nil {
		t.Fatalf("create crossbar module: %v", err)
	}
	return []testLifecycleModule{
		{name: "hysteresis", module: demo1gui.NewEmbeddedApp()},
		{name: "crossbar", module: crossbar},
		{name: "mnist", module: demo3gui.NewEmbeddedDualModeApp()},
		{name: "circuits", module: demo4gui.NewEmbeddedCircuitsApp()},
		{name: "comparison", module: demo5gui.NewEmbeddedComparisonApp()},
		{name: "eda", module: demo6gui.NewEmbeddedEDAApp()},
		{name: "docs", module: demo7gui.NewEmbeddedDocsApp()},
	}
}

type layoutNode struct {
	obj fyne.CanvasObject
}

func flattenLayoutNodes(root fyne.CanvasObject) []layoutNode {
	seen := map[uintptr]bool{}
	out := make([]layoutNode, 0, 256)
	var walk func(fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		if p := canvasObjectPointer(o); p != 0 {
			if seen[p] {
				return
			}
			seen[p] = true
		}
		out = append(out, layoutNode{obj: o})
		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child)
			}
			return
		}
		if tabs, ok := o.(*container.AppTabs); ok {
			if selected := tabs.Selected(); selected != nil {
				walk(selected.Content)
			}
			return
		}
		if scroll, ok := o.(*container.Scroll); ok {
			walk(scroll.Content)
			return
		}
		if accordion, ok := o.(*widget.Accordion); ok {
			for _, item := range accordion.Items {
				walk(item.Detail)
			}
			return
		}
		if split, ok := o.(*container.Split); ok {
			walk(split.Leading)
			walk(split.Trailing)
			return
		}
	}
	walk(root)
	return out
}

func canvasObjectPointer(o fyne.CanvasObject) uintptr {
	v := reflect.ValueOf(o)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return 0
	}
	return v.Pointer()
}

func isLongUnboundedLabel(label *widget.Label) bool {
	return len(label.Text) > 80 && label.Wrapping == fyne.TextWrapOff && label.Truncation == fyne.TextTruncateOff
}
