package tui

import "github.com/charmbracelet/lipgloss"

var (
	workColor      = lipgloss.Color("#FF6B6B")
	shortBreakColor = lipgloss.Color("#51CF66")
	longBreakColor  = lipgloss.Color("#339AF0")
	mutedColor     = lipgloss.Color("#666666")
	whiteColor     = lipgloss.Color("#FFFFFF")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(whiteColor).
			MarginBottom(1)

	taskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD43B")).
			Bold(true)

	timerStyle = lipgloss.NewStyle().
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	stateStyle = lipgloss.NewStyle().
			Italic(true).
			MarginBottom(1)

	pomodoroCountStyle = lipgloss.NewStyle().
				MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	progressBarStyle = lipgloss.NewStyle().
				MarginBottom(1)
)

func colorForSessionType(st int) lipgloss.Color {
	switch st {
	case 0: // Work
		return workColor
	case 1: // ShortBreak
		return shortBreakColor
	case 2: // LongBreak
		return longBreakColor
	default:
		return whiteColor
	}
}
