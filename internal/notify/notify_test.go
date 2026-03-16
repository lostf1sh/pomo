package notify

import (
	"testing"

	"github.com/lostf1sh/pomo/internal/config"
)

func TestNewNotifier(t *testing.T) {
	cfg := config.Config{
		NotifyDesktop: true,
		NotifyBell:    false,
	}

	n := New(cfg)
	if !n.desktop {
		t.Error("expected desktop notifications enabled")
	}
	if n.bell {
		t.Error("expected bell notifications disabled")
	}
}

func TestNotifyDoesNotPanic(t *testing.T) {
	// Just verify that calling notify methods doesn't panic
	n := &Notifier{desktop: false, bell: false}
	n.Notify("test", "test message")
	n.NotifyWorkComplete("task")
	n.NotifyBreakComplete()
	n.NotifyLongBreakComplete()
}
