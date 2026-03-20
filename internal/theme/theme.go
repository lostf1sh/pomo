package theme

import (
	"sort"
)

// Theme holds lipgloss hex colors for the TUI.
type Theme struct {
	Name       string
	Work       string
	ShortBreak string
	LongBreak  string
	Muted      string
	Text       string
	Accent     string
	ProgressFG string
	ProgressBG string
}

// Themes is the built-in theme registry.
var Themes = map[string]Theme{
	"default": {
		Name:       "default",
		Work:       "#FF6B6B",
		ShortBreak: "#51CF66",
		LongBreak:  "#339AF0",
		Muted:      "#666666",
		Text:       "#FFFFFF",
		Accent:     "#FFD43B",
		ProgressFG: "#FF6B6B",
		ProgressBG: "#666666",
	},
	"catppuccin-mocha": {
		Name:       "catppuccin-mocha",
		Work:       "#F38BA8",
		ShortBreak: "#A6E3A1",
		LongBreak:  "#89B4FA",
		Muted:      "#585B70",
		Text:       "#CDD6F4",
		Accent:     "#F9E2AF",
		ProgressFG: "#F38BA8",
		ProgressBG: "#585B70",
	},
	"dracula": {
		Name:       "dracula",
		Work:       "#FF5555",
		ShortBreak: "#50FA7B",
		LongBreak:  "#8BE9FD",
		Muted:      "#6272A4",
		Text:       "#F8F8F2",
		Accent:     "#F1FA8C",
		ProgressFG: "#FF5555",
		ProgressBG: "#6272A4",
	},
	"gruvbox": {
		Name:       "gruvbox",
		Work:       "#FB4934",
		ShortBreak: "#B8BB26",
		LongBreak:  "#83A598",
		Muted:      "#665C54",
		Text:       "#EBDBB2",
		Accent:     "#FABD2F",
		ProgressFG: "#FB4934",
		ProgressBG: "#665C54",
	},
	"nord": {
		Name:       "nord",
		Work:       "#BF616A",
		ShortBreak: "#A3BE8C",
		LongBreak:  "#88C0D0",
		Muted:      "#4C566A",
		Text:       "#ECEFF4",
		Accent:     "#EBCB8B",
		ProgressFG: "#BF616A",
		ProgressBG: "#4C566A",
	},
	"tokyo-night": {
		Name:       "tokyo-night",
		Work:       "#F7768E",
		ShortBreak: "#9ECE6A",
		LongBreak:  "#7AA2F7",
		Muted:      "#565F89",
		Text:       "#C0CAF5",
		Accent:     "#E0AF68",
		ProgressFG: "#F7768E",
		ProgressBG: "#565F89",
	},
	"solarized": {
		Name:       "solarized",
		Work:       "#DC322F",
		ShortBreak: "#859900",
		LongBreak:  "#268BD2",
		Muted:      "#586E75",
		Text:       "#FDF6E3",
		Accent:     "#B58900",
		ProgressFG: "#DC322F",
		ProgressBG: "#586E75",
	},
}

// Get returns a theme by name.
func Get(name string) (Theme, bool) {
	t, ok := Themes[name]
	return t, ok
}

// Names returns sorted theme names.
func Names() []string {
	names := make([]string, 0, len(Themes))
	for n := range Themes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Default returns the built-in default theme.
func Default() Theme {
	return Themes["default"]
}
