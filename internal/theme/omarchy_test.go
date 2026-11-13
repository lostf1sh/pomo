package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func writeOmarchyColors(t *testing.T, body string) {
	t.Helper()
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	dir := filepath.Join(state, "omarchy", "current", "theme")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "colors.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOmarchyMissing(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	if _, ok := Get(OmarchyName); ok {
		t.Fatal("expected omarchy theme to be unavailable without colors.toml")
	}
	for _, n := range Names() {
		if n == OmarchyName {
			t.Fatal("omarchy listed without colors.toml")
		}
	}
}

func TestOmarchyFromColors(t *testing.T) {
	writeOmarchyColors(t, `accent = "#d4a5ff"
foreground = "#d4a5ff"
color1 = "#ff6ec7"
color2 = "#5ffbf1"
color3 = "#f9f871"
color4 = "#8b9aff"
color8 = "#543a6e"
# comment = "#000000"
`)
	th, ok := Get(OmarchyName)
	if !ok {
		t.Fatal("expected omarchy theme")
	}
	want := Theme{Name: "omarchy", Work: "#D4A5FF", ShortBreak: "#5FFBF1", LongBreak: "#8B9AFF",
		Muted: "#946FB6", Text: "#D4A5FF", Accent: "#FF6EC7", ProgressFG: "#D4A5FF", ProgressBG: "#543A6E"}
	if th != want {
		t.Errorf("got %+v\nwant %+v", th, want)
	}
	found := false
	for _, n := range Names() {
		found = found || n == OmarchyName
	}
	if !found {
		t.Error("omarchy missing from Names()")
	}
}

func TestOmarchyFallsBack(t *testing.T) {
	writeOmarchyColors(t, `color1 = "#112233"`)
	th, _ := Get(OmarchyName)
	if th.Work != "#112233" || th.ShortBreak != Default().ShortBreak || th.ProgressBG != Default().Muted {
		t.Errorf("unexpected fallback: %+v", th)
	}
}
