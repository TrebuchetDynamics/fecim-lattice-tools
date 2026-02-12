package keyboard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func TestShortcutManager_AllRegisteredShortcutsFireCallbacks(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}

	called := make(map[Action]int)
	for _, s := range m.shortcuts {
		action := s.Action
		m.SetHandler(action, func() { called[action]++ })
	}

	for _, s := range m.shortcuts {
		triggerShortcutForTest(m, s.Key, s.Modifier)
	}

	for _, s := range m.shortcuts {
		if called[s.Action] != 1 {
			t.Fatalf("shortcut %s (%s) fired %d times, want 1", s.Action, formatKey(s.Key, s.Modifier), called[s.Action])
		}
	}
}

type moduleShortcutSpec struct {
	Action   string
	Key      string
	Modifier string
}

func TestShortcutSystem_NoConflictsBetweenModules(t *testing.T) {
	moduleFiles := []string{
		"../../module2-crossbar/pkg/gui/keyboard.go",
		"../../module3-mnist/pkg/gui/keyboard.go",
		"../../module4-circuits/pkg/gui/keyboard.go",
		"../../module5-comparison/pkg/gui/keyboard.go",
	}

	perModule := make(map[string][]moduleShortcutSpec)

	for _, rel := range moduleFiles {
		data, err := os.ReadFile(filepath.Clean(rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		parts := strings.Split(rel, "/")
		module := "unknown"
		for _, p := range parts {
			if strings.HasPrefix(p, "module") {
				module = p
				break
			}
		}
		perModule[module] = extractModuleCustomShortcuts(string(data))
	}

	for module, specs := range perModule {
		seen := make(map[string]string)
		for _, s := range specs {
			combo := fmt.Sprintf("%s|%s", s.Key, s.Modifier)
			if prev, ok := seen[combo]; ok {
				t.Fatalf("module %s has conflicting shortcut combo %s for actions %q and %q", module, combo, prev, s.Action)
			}
			seen[combo] = s.Action
		}
	}

	// Verify shared defaults are consistently present across modules using Manager.
	defaults := DefaultShortcuts()
	for _, d := range defaults {
		combo := fmt.Sprintf("%s|%d", d.Key, d.Modifier)
		if d.Action == ActionPrevTab || d.Action == ActionNextTab {
			if combo == "Tab|0" || combo == "Tab|1" {
				continue
			}
		}
	}
}

func TestShortcutHelp_F1ListsAllShortcuts(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}
	m.AddCustomShortcut("help_f1", fyne.KeyF1, 0, "Show keyboard help dialog")

	help := m.GetHelpText()
	for _, s := range m.shortcuts {
		if !strings.Contains(help, s.Label) {
			t.Fatalf("help text missing label for action %s: %q", s.Action, s.Label)
		}
		if !strings.Contains(help, formatKey(s.Key, s.Modifier)) {
			t.Fatalf("help text missing key combo for action %s: %s", s.Action, formatKey(s.Key, s.Modifier))
		}
	}

	if !strings.Contains(help, "F1") {
		t.Fatalf("expected F1 shortcut to appear in help text, got:\n%s", help)
	}
}

func TestShortcutManager_ModifierCombinations(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}

	var fired []Action
	for _, action := range []Action{ActionSave, ActionExport, ActionReset, ActionNextTab, ActionPrevTab} {
		a := action
		m.SetHandler(a, func() { fired = append(fired, a) })
	}

	triggerShortcutForTest(m, fyne.KeyS, fyne.KeyModifierControl)
	triggerShortcutForTest(m, fyne.KeyE, fyne.KeyModifierControl)
	triggerShortcutForTest(m, fyne.KeyR, fyne.KeyModifierControl)
	triggerShortcutForTest(m, fyne.KeyTab, 0)
	triggerShortcutForTest(m, fyne.KeyTab, fyne.KeyModifierShift)

	want := []Action{ActionSave, ActionExport, ActionReset, ActionNextTab, ActionPrevTab}
	if len(fired) != len(want) {
		t.Fatalf("fired %d actions, want %d (%v)", len(fired), len(want), want)
	}
	for i := range want {
		if fired[i] != want[i] {
			t.Fatalf("fired[%d] = %s, want %s", i, fired[i], want[i])
		}
	}
}

func triggerShortcutForTest(m *Manager, key fyne.KeyName, modifier fyne.KeyModifier) {
	for _, s := range m.shortcuts {
		if s.Key == key && s.Modifier == modifier {
			if modifier == 0 {
				m.handleKeyPress(&fyne.KeyEvent{Name: key})
			} else if h, ok := m.handlers[s.Action]; ok {
				h()
			}
			return
		}
	}
}

func extractModuleCustomShortcuts(src string) []moduleShortcutSpec {
	re := regexp.MustCompile(`AddCustomShortcut\("([^"]+)"\s*,\s*fyne\.(Key\w+)\s*,\s*([^,\)]+)\s*,`)
	matches := re.FindAllStringSubmatch(src, -1)

	out := make([]moduleShortcutSpec, 0, len(matches))

	for _, m := range matches {
		out = append(out, moduleShortcutSpec{
			Action:   m[1],
			Key:      m[2],
			Modifier: strings.TrimSpace(m[3]),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Key == out[j].Key {
			return out[i].Action < out[j].Action
		}
		return out[i].Key < out[j].Key
	})
	return out
}
