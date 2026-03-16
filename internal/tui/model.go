package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/notify"
	"github.com/lostf1sh/pomo/internal/store"
	"github.com/lostf1sh/pomo/internal/timer"
)

type Model struct {
	engine   *timer.Engine
	store    *store.Store
	notifier *notify.Notifier
	config   config.Config
	width    int
	height   int
	quitting bool
	showHelp bool
}

func NewModel(cfg config.Config, s *store.Store, task string) Model {
	return Model{
		engine:   timer.NewEngine(cfg, task),
		store:    s,
		notifier: notify.New(cfg),
		config:   cfg,
	}
}

func (m Model) Init() tea.Cmd {
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
		return m, doTick()
	case timer.Running:
		m.engine.Pause()
		return m, nil
	case timer.Paused:
		m.engine.Start()
		return m, doTick()
	}
	return m, nil
}

func (m Model) handleSkip() (tea.Model, tea.Cmd) {
	sess := m.engine.Skip()
	if sess != nil && m.store != nil {
		_ = m.store.SaveSession(*sess)
	}
	return m, nil
}

func (m Model) handleTick() (tea.Model, tea.Cmd) {
	sess := m.engine.Tick()

	if sess != nil {
		// Save completed session
		if m.store != nil {
			_ = m.store.SaveSession(*sess)
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

		return m, nil
	}

	if m.engine.State == timer.Running {
		return m, doTick()
	}
	return m, nil
}

func (m *Model) handleQuit() {
	// Save incomplete session on quit
	if m.engine.State == timer.Running || m.engine.State == timer.Paused {
		sess := m.engine.Skip()
		if sess != nil && m.store != nil {
			_ = m.store.SaveSession(*sess)
		}
	}
}
