package theme

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// OmarchyName is the theme that follows the active Omarchy desktop theme.
const OmarchyName = "omarchy"

var omarchyLine = regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*"(#[0-9A-Fa-f]{6})"`)

// omarchyColorsPath is the colors.toml of the active Omarchy theme.
func omarchyColorsPath() string {
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		state = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(state, "omarchy", "current", "theme", "colors.toml")
}

// LoadOmarchy builds a theme from the active Omarchy theme's colors.toml.
// It is read every time, so pomo picks up desktop theme changes on its next start.
func LoadOmarchy() (Theme, bool) {
	f, err := os.Open(omarchyColorsPath())
	if err != nil {
		return Theme{}, false
	}
	defer f.Close()

	c := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := omarchyLine.FindStringSubmatch(sc.Text()); m != nil {
			c[strings.ToLower(m[1])] = strings.ToUpper(m[2])
		}
	}

	pick := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := c[k]; ok {
				return v
			}
		}
		return ""
	}

	fallback := Default()
	t := Theme{
		Name:       OmarchyName,
		Work:       pick("accent", "color1"),
		ShortBreak: pick("color2"),
		LongBreak:  pick("color4"),
		Muted:      pick("color8"),
		Text:       pick("foreground", "color7"),
		Accent:     pick("color1", "color3"),
	}
	for _, p := range []struct {
		dst *string
		def string
	}{
		{&t.Work, fallback.Work},
		{&t.ShortBreak, fallback.ShortBreak},
		{&t.LongBreak, fallback.LongBreak},
		{&t.Muted, fallback.Muted},
		{&t.Text, fallback.Text},
		{&t.Accent, fallback.Accent},
	} {
		if *p.dst == "" {
			*p.dst = p.def
		}
	}
	t.ProgressFG = t.Work
	t.ProgressBG = t.Muted
	// color8 is a dark "bright black" in most themes: fine for the progress track,
	// too dim for the key hints, so lift those halfway towards the text color.
	t.Muted = mix(t.Muted, t.Text)
	return t, true
}

// mix returns the midpoint of two #RRGGBB colors.
func mix(a, b string) string {
	var ra, ga, ba, rb, gb, bb int
	if _, err := fmt.Sscanf(a, "#%02x%02x%02x", &ra, &ga, &ba); err != nil {
		return a
	}
	if _, err := fmt.Sscanf(b, "#%02x%02x%02x", &rb, &gb, &bb); err != nil {
		return a
	}
	return fmt.Sprintf("#%02X%02X%02X", (ra+rb)/2, (ga+gb)/2, (ba+bb)/2)
}
