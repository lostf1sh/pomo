package theme

import (
	"regexp"
	"testing"
)

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func TestThemesValidHex(t *testing.T) {
	for name, th := range Themes {
		for _, field := range []struct {
			label string
			val   string
		}{
			{"Work", th.Work},
			{"ShortBreak", th.ShortBreak},
			{"LongBreak", th.LongBreak},
			{"Muted", th.Muted},
			{"Text", th.Text},
			{"Accent", th.Accent},
			{"ProgressFG", th.ProgressFG},
			{"ProgressBG", th.ProgressBG},
		} {
			if !hexColor.MatchString(field.val) {
				t.Errorf("theme %q %s: invalid hex %q", name, field.label, field.val)
			}
		}
	}
}

func TestGet(t *testing.T) {
	th, ok := Get("dracula")
	if !ok {
		t.Fatal("expected dracula theme")
	}
	if th.Name != "dracula" {
		t.Errorf("Name: got %q", th.Name)
	}

	_, ok = Get("nonexistent-theme-xyz")
	if ok {
		t.Error("expected miss for unknown theme")
	}
}

func TestNamesSortedAndComplete(t *testing.T) {
	names := Names()
	if len(names) != len(Themes) {
		t.Fatalf("Names() len %d, want %d", len(names), len(Themes))
	}
	for i := 1; i < len(names); i++ {
		if names[i] <= names[i-1] {
			t.Errorf("not sorted: %q before %q", names[i-1], names[i])
		}
	}
	for n := range Themes {
		found := false
		for _, x := range names {
			if x == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("theme %q missing from Names()", n)
		}
	}
}

func TestDefault(t *testing.T) {
	d := Default()
	if d.Name != "default" {
		t.Errorf("Default().Name = %q, want default", d.Name)
	}
	if got, ok := Get("default"); !ok || got.Work != d.Work {
		t.Error("Default() should match Themes[\"default\"]")
	}
}
