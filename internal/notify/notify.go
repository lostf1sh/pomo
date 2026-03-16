package notify

import (
	"fmt"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/gen2brain/beeep"
)

type Notifier struct {
	desktop bool
	bell    bool
}

func New(cfg config.Config) *Notifier {
	return &Notifier{
		desktop: cfg.NotifyDesktop,
		bell:    cfg.NotifyBell,
	}
}

func (n *Notifier) Notify(title, message string) {
	if n.bell {
		fmt.Print("\a")
	}
	if n.desktop {
		_ = beeep.Notify(title, message, "")
	}
}

func (n *Notifier) NotifyWorkComplete(task string) {
	msg := "Work session complete! Time for a break."
	if task != "" {
		msg = fmt.Sprintf("Work session for '%s' complete! Time for a break.", task)
	}
	n.Notify("Pomodoro Complete", msg)
}

func (n *Notifier) NotifyBreakComplete() {
	n.Notify("Break Over", "Break is over! Ready to work?")
}

func (n *Notifier) NotifyLongBreakComplete() {
	n.Notify("Long Break Over", "Long break is over! Ready for a new set?")
}
