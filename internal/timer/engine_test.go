package timer

import (
	"testing"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
)

func shortConfig() config.Config {
	return config.Config{
		WorkDuration:       3 * time.Second,
		ShortBreakDuration: 1 * time.Second,
		LongBreakDuration:  2 * time.Second,
		LongBreakInterval:  4,
		NotifyDesktop:      false,
		NotifyBell:         false,
	}
}

func TestNewEngine(t *testing.T) {
	cfg := shortConfig()
	e := NewEngine(cfg, "test task")

	if e.State != Idle {
		t.Errorf("expected Idle, got %v", e.State)
	}
	if e.CurrentType != Work {
		t.Errorf("expected Work, got %v", e.CurrentType)
	}
	if e.Task != "test task" {
		t.Errorf("expected 'test task', got %s", e.Task)
	}
	if e.Remaining != 3*time.Second {
		t.Errorf("expected 3s remaining, got %v", e.Remaining)
	}
}

func TestStartPause(t *testing.T) {
	e := NewEngine(shortConfig(), "test")

	e.Start()
	if e.State != Running {
		t.Errorf("expected Running, got %v", e.State)
	}

	e.Pause()
	if e.State != Paused {
		t.Errorf("expected Paused, got %v", e.State)
	}

	e.Start()
	if e.State != Running {
		t.Errorf("expected Running after unpause, got %v", e.State)
	}
}

func TestReset(t *testing.T) {
	cfg := shortConfig()
	e := NewEngine(cfg, "test")

	e.Start()
	time.Sleep(100 * time.Millisecond)
	e.Reset()

	if e.State != Idle {
		t.Errorf("expected Idle, got %v", e.State)
	}
	if e.Remaining != cfg.WorkDuration {
		t.Errorf("expected remaining to be reset to %v, got %v", cfg.WorkDuration, e.Remaining)
	}
}

func TestTickCompletesSession(t *testing.T) {
	cfg := config.Config{
		WorkDuration:       50 * time.Millisecond,
		ShortBreakDuration: 50 * time.Millisecond,
		LongBreakDuration:  50 * time.Millisecond,
		LongBreakInterval:  4,
	}
	e := NewEngine(cfg, "test")
	e.Start()

	time.Sleep(80 * time.Millisecond)

	sess := e.Tick()
	if sess == nil {
		t.Fatal("expected completed session")
	}
	if !sess.Completed {
		t.Error("expected session to be completed")
	}
	if sess.Type != Work {
		t.Errorf("expected Work session, got %v", sess.Type)
	}

	// After work completes, should advance to short break
	if e.CurrentType != ShortBreak {
		t.Errorf("expected ShortBreak, got %v", e.CurrentType)
	}
	if e.State != Idle {
		t.Errorf("expected Idle after advance, got %v", e.State)
	}
}

func TestSkip(t *testing.T) {
	e := NewEngine(shortConfig(), "test")
	e.Start()

	sess := e.Skip()
	if sess == nil {
		t.Fatal("expected incomplete session from skip")
	}
	if sess.Completed {
		t.Error("skipped session should not be completed")
	}
	if e.CurrentType != ShortBreak {
		t.Errorf("expected ShortBreak after skip, got %v", e.CurrentType)
	}
}

func TestLongBreakAfterInterval(t *testing.T) {
	cfg := config.Config{
		WorkDuration:       10 * time.Millisecond,
		ShortBreakDuration: 10 * time.Millisecond,
		LongBreakDuration:  10 * time.Millisecond,
		LongBreakInterval:  2, // Long break after 2 pomodoros
	}
	e := NewEngine(cfg, "test")

	// Complete 2 work sessions
	for i := 0; i < 2; i++ {
		e.Start()
		time.Sleep(20 * time.Millisecond)
		e.Tick()

		if i < 1 {
			// Start and complete the break
			e.Start()
			time.Sleep(20 * time.Millisecond)
			e.Tick()
		}
	}

	// After 2 work sessions, should be long break
	if e.CurrentType != LongBreak {
		t.Errorf("expected LongBreak after %d pomodoros, got %v", cfg.LongBreakInterval, e.CurrentType)
	}
}

func TestProgress(t *testing.T) {
	e := NewEngine(shortConfig(), "test")

	if p := e.Progress(); p != 0 {
		t.Errorf("expected 0 progress at start, got %f", p)
	}

	e.Start()
	time.Sleep(1500 * time.Millisecond)
	e.Tick()

	p := e.Progress()
	if p < 0.3 || p > 0.7 {
		t.Errorf("expected ~0.5 progress, got %f", p)
	}
}
