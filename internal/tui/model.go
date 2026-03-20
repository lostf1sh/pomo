package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/notify"
	"github.com/lostf1sh/pomo/internal/stats"
	"github.com/lostf1sh/pomo/internal/store"
	"github.com/lostf1sh/pomo/internal/theme"
	"github.com/lostf1sh/pomo/internal/timer"
)

type Model struct {
	engine         *timer.Engine
	store          *store.Store
	notifier       *notify.Notifier
	config         config.Config
	colors         ThemeColors
	dailyCompleted int
	width          int
	height         int
	quitting       bool
	showHelp       bool
}

func NewModel(cfg config.Config, s *store.Store, task string) Model {
	return NewModelWithEngine(cfg, s, timer.NewEngine(cfg, task))
}

func NewModelWithEngine(cfg config.Config, s *store.Store, engine *timer.Engine) Model {
	th := theme.Default()
	if cfg.Theme != "" {
		if t, ok := theme.Get(cfg.Theme); ok {
			th = t
		}
	}
	return Model{
		engine:         engine,
		store:          s,
		notifier:       notify.New(cfg),
		config:         cfg,
		colors:         NewThemeColors(th),
		dailyCompleted: loadTodayCompleted(s),
	}
}

func (m Model) Init() tea.Cmd {
	if m.engine != nil && m.engine.State == timer.Running {
		return doTick()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.handleQuit()
			m.quitting = true
			return m, tea.Quit

		case "s", " ":
			return m.handleStartPause()

		case "r":
			m.engine.Reset()
			m.clearActiveState()
			return m, nil

		case "k":
			return m.handleSkip()

		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	case tickMsg:
		return m.handleTick()
	}

	return m, nil
}

func (m Model) handleStartPause() (tea.Model, tea.Cmd) {
	switch m.engine.State {
	case timer.Idle:
		m.engine.Start()
		m.saveActiveState()
		return m, doTick()
	case timer.Running:
		m.engine.Pause()
		m.saveActiveState()
		return m, nil
	case timer.Paused:
		m.engine.Start()
		m.saveActiveState()
		return m, doTick()
	}
	return m, nil
}

func (m Model) handleSkip() (tea.Model, tea.Cmd) {
	sess := m.engine.Skip()
	if sess != nil && m.store != nil {
		_ = m.store.SaveSession(*sess)
	}
	m.clearActiveState()
	return m, nil
}

func (m Model) handleTick() (tea.Model, tea.Cmd) {
	sess := m.engine.Tick()

	if sess != nil {
		// Save completed session
		if m.store != nil {
			_ = m.store.SaveSession(*sess)
		}
		if sess.Type == timer.Work && sess.Completed {
			m.dailyCompleted++
		}

		// Send notification
		if m.notifier != nil {
			switch sess.Type {
			case timer.Work:
				m.notifier.NotifyWorkComplete(sess.Task)
			case timer.ShortBreak:
				m.notifier.NotifyBreakComplete()
			case timer.LongBreak:
				m.notifier.NotifyLongBreakComplete()
			}
		}

		m.clearActiveState()
		return m, nil
	}

	if m.engine.State == timer.Running {
		m.saveActiveState()
		return m, doTick()
	}
	return m, nil
}

func (m *Model) handleQuit() {
	if m.engine.State == timer.Running || m.engine.State == timer.Paused {
		m.saveActiveState()
		return
	}
	m.clearActiveState()
}

func (m *Model) saveActiveState() {
	if m.store == nil || m.engine == nil {
		return
	}
	_ = m.store.SaveActiveState(*m.engine.Snapshot())
}

func (m *Model) clearActiveState() {
	if m.store == nil {
		return
	}
	_ = m.store.ClearActiveState()
}

func loadTodayCompleted(s *store.Store) int {
	if s == nil {
		return 0
	}

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sessions, err := s.GetSessions(from, now)
	if err != nil {
		return 0
	}

	return stats.Compute(sessions).TotalPomodoros
}
