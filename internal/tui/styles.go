package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lostf1sh/pomo/internal/theme"
)

// ThemeColors holds lipgloss colors derived from a theme.Theme.
type ThemeColors struct {
	Work       lipgloss.Color
	ShortBreak lipgloss.Color
	LongBreak  lipgloss.Color
	Muted      lipgloss.Color
	Text       lipgloss.Color
	Accent     lipgloss.Color
	ProgressBG lipgloss.Color
}

// NewThemeColors converts hex strings from theme.Theme to lipgloss colors.
func NewThemeColors(t theme.Theme) ThemeColors {
	return ThemeColors{
		Work:       lipgloss.Color(t.Work),
		ShortBreak: lipgloss.Color(t.ShortBreak),
		LongBreak:  lipgloss.Color(t.LongBreak),
		Muted:      lipgloss.Color(t.Muted),
		Text:       lipgloss.Color(t.Text),
		Accent:     lipgloss.Color(t.Accent),
		ProgressBG: lipgloss.Color(t.ProgressBG),
	}
}

var (
	timerStyle = lipgloss.NewStyle().
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	stateStyle = lipgloss.NewStyle().
			Italic(true).
			MarginBottom(1)

	pomodoroCountStyle = lipgloss.NewStyle().
				MarginBottom(1)

	progressBarStyle = lipgloss.NewStyle().
				MarginBottom(1)
)

func colorForSessionType(colors ThemeColors, st int) lipgloss.Color {
	switch st {
	case 0: // Work
		return colors.Work
	case 1: // ShortBreak
		return colors.ShortBreak
	case 2: // LongBreak
		return colors.LongBreak
	default:
		return colors.Text
	}
}
