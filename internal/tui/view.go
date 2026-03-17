package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lostf1sh/pomo/internal/timer"
)

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	color := colorForSessionType(int(m.engine.CurrentType))

	var b strings.Builder

	// Header: session type
	header := sessionTypeLabel(m.engine.CurrentType)
	b.WriteString(titleStyle.Foreground(color).Render(header))
	b.WriteString("\n")

	// Task name
	if m.engine.Task != "" {
		b.WriteString(taskStyle.Render("Task: " + m.engine.Task))
		b.WriteString("\n")
	}

	// Big timer display
	minutes := int(m.engine.Remaining.Minutes())
	seconds := int(m.engine.Remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)
	bigTimer := timerStyle.Foreground(color).Render(renderBigDigits(timeStr))
	b.WriteString(bigTimer)
	b.WriteString("\n")

	// State
	stateStr := stateLabel(m.engine.State)
	b.WriteString(stateStyle.Foreground(color).Render(stateStr))
	b.WriteString("\n")

	// Progress bar
	progress := m.engine.Progress()
	progressBar := renderProgressBar(progress, 40, color)
	b.WriteString(progressBarStyle.Render(progressBar))
	b.WriteString("\n")

	// Pomodoro counter
	counter := renderPomodoroCounter(m.engine.PomodorosInSet, m.engine.Config.LongBreakInterval, m.engine.CompletedTotal)
	b.WriteString(pomodoroCountStyle.Foreground(color).Render(counter))
	b.WriteString("\n")

	if m.config.DailyGoalPomodoros > 0 {
		goal := renderDailyGoal(m.dailyCompleted, m.config.DailyGoalPomodoros)
		b.WriteString(pomodoroCountStyle.Foreground(color).Render(goal))
		b.WriteString("\n")
	}

	// Help footer
	help := helpText(m.engine.State, m.showHelp)
	b.WriteString(helpStyle.Render(help))

	// Center everything
	content := b.String()
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content)
}

func sessionTypeLabel(st timer.SessionType) string {
	switch st {
	case timer.Work:
		return "WORK"
	case timer.ShortBreak:
		return "SHORT BREAK"
	case timer.LongBreak:
		return "LONG BREAK"
	default:
		return "POMODORO"
	}
}

func stateLabel(s timer.TimerState) string {
	switch s {
	case timer.Idle:
		return "Press [s] to start"
	case timer.Running:
		return "Running..."
	case timer.Paused:
		return "Paused"
	default:
		return ""
	}
}

func renderProgressBar(progress float64, width int, color lipgloss.Color) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(mutedColor)

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	pct := fmt.Sprintf(" %3d%%", int(progress*100))
	return bar + pct
}

func renderPomodoroCounter(current, total, completedTotal int) string {
	var dots []string
	for i := 0; i < total; i++ {
		if i < current {
			dots = append(dots, "●")
		} else {
			dots = append(dots, "○")
		}
	}
	counter := strings.Join(dots, " ")
	return fmt.Sprintf("Pomodoros: [%s]  Total: %d", counter, completedTotal)
}

func renderDailyGoal(completed, target int) string {
	if completed > target {
		return fmt.Sprintf("Daily Goal: %d/%d completed", completed, target)
	}
	return fmt.Sprintf("Daily Goal: %d/%d", completed, target)
}

func helpText(state timer.TimerState, showHelp bool) string {
	if showHelp {
		return `Keyboard Shortcuts:
  s  - Start / Pause
  r  - Reset current segment
  k  - Skip to next segment
  q  - Quit
  ?  - Toggle this help`
	}

	switch state {
	case timer.Idle:
		return "s: start  r: reset  k: skip  q: quit  ?: help"
	case timer.Running:
		return "s: pause  r: reset  k: skip  q: quit  ?: help"
	case timer.Paused:
		return "s: resume  r: reset  k: skip  q: quit  ?: help"
	default:
		return "q: quit  ?: help"
	}
}

var bigDigits = map[byte][5]string{
	'0': {" ██ ", "█  █", "█  █", "█  █", " ██ "},
	'1': {" █  ", "██  ", " █  ", " █  ", "███ "},
	'2': {" ██ ", "█  █", "  █ ", " █  ", "████"},
	'3': {"███ ", "   █", " ██ ", "   █", "███ "},
	'4': {"█  █", "█  █", "████", "   █", "   █"},
	'5': {"████", "█   ", "███ ", "   █", "███ "},
	'6': {" ██ ", "█   ", "███ ", "█  █", " ██ "},
	'7': {"████", "   █", "  █ ", " █  ", " █  "},
	'8': {" ██ ", "█  █", " ██ ", "█  █", " ██ "},
	'9': {" ██ ", "█  █", " ███", "   █", " ██ "},
	':': {"    ", " ██ ", "    ", " ██ ", "    "},
}

func renderBigDigits(s string) string {
	var lines [5]string
	for i := 0; i < 5; i++ {
		for _, c := range []byte(s) {
			if d, ok := bigDigits[c]; ok {
				lines[i] += d[i] + " "
			}
		}
	}
	return strings.Join(lines[:], "\n")
}
