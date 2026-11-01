package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/lostf1sh/pomo/internal/config"
)

type Notifier struct {
	desktop bool
	bell    bool
	sound   string
}

func New(cfg config.Config) *Notifier {
	return &Notifier{
		desktop: cfg.NotifyDesktop,
		bell:    cfg.NotifyBell,
		sound:   cfg.Sound,
	}
}

func (n *Notifier) Notify(title, message string) {
	if n.bell {
		fmt.Print("\a")
	}
	if n.desktop {
		_ = beeep.Alert(title, message, "")
	}
	n.playSound()
}

func (n *Notifier) playSound() {
	if strings.EqualFold(n.sound, "off") || strings.EqualFold(n.sound, "none") || strings.EqualFold(n.sound, "false") {
		return
	}

	sound := n.sound
	go func() {
		switch runtime.GOOS {
		case "linux":
			filePath := resolveSoundPath(sound)
			if filePath != "" {
				if exec.Command("paplay", filePath).Run() == nil {
					return
				}
				if exec.Command("pw-play", filePath).Run() == nil {
					return
				}
				if exec.Command("mpv", "--no-video", filePath).Run() == nil {
					return
				}
				if exec.Command("aplay", filePath).Run() == nil {
					return
				}
				if exec.Command("canberra-gtk-play", "-f", filePath).Run() == nil {
					return
				}
			}

			// Fallback defaults
			defaultSounds := []string{
				"/usr/share/sounds/freedesktop/stereo/complete.oga",
				"/usr/share/sounds/freedesktop/stereo/bell.oga",
				"/usr/share/sounds/gnome/default/alerts/glass.ogg",
			}
			for _, sf := range defaultSounds {
				if _, err := os.Stat(sf); err == nil {
					if exec.Command("paplay", sf).Run() == nil {
						return
					}
					if exec.Command("pw-play", sf).Run() == nil {
						return
					}
					if exec.Command("aplay", sf).Run() == nil {
						return
					}
				}
			}
			_ = exec.Command("canberra-gtk-play", "-i", "complete").Run()
		case "darwin":
			target := "/System/Library/Sounds/Glass.aiff"
			if resolved := resolveCustomFile(sound); resolved != "" {
				target = resolved
			}
			_ = exec.Command("afplay", target).Run()
		case "windows":
			if resolved := resolveCustomFile(sound); resolved != "" {
				_ = exec.Command("powershell", "-c", fmt.Sprintf(`(New-Object Media.SoundPlayer "%s").PlaySync()`, resolved)).Run()
				return
			}
			_ = exec.Command("powershell", "-c", "[System.Media.SystemSounds]::Asterisk.Play()").Run()
		}
	}()
}

func resolveCustomFile(sound string) string {
	if sound == "" || sound == "default" {
		return ""
	}
	if strings.HasPrefix(sound, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			sound = filepath.Join(home, sound[2:])
		}
	}
	if _, err := os.Stat(sound); err == nil {
		return sound
	}
	return ""
}

func resolveSoundPath(sound string) string {
	if sound == "" || sound == "default" {
		return ""
	}

	if resolved := resolveCustomFile(sound); resolved != "" {
		return resolved
	}

	presets := map[string]string{
		"complete": "/usr/share/sounds/freedesktop/stereo/complete.oga",
		"alarm":    "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga",
		"bell":     "/usr/share/sounds/freedesktop/stereo/bell.oga",
		"message":  "/usr/share/sounds/freedesktop/stereo/message.oga",
		"info":     "/usr/share/sounds/freedesktop/stereo/dialog-information.oga",
	}

	if p, ok := presets[strings.ToLower(sound)]; ok {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	namedPath := filepath.Join("/usr/share/sounds/freedesktop/stereo", sound+".oga")
	if _, err := os.Stat(namedPath); err == nil {
		return namedPath
	}

	return ""
}

// ValidateSound checks if the sound is a recognized keyword, preset, or existing file.
func ValidateSound(sound string) error {
	if sound == "" || sound == "default" || strings.EqualFold(sound, "off") || strings.EqualFold(sound, "none") || strings.EqualFold(sound, "false") {
		return nil
	}
	if resolveSoundPath(sound) != "" {
		return nil
	}
	return fmt.Errorf("sound %q not found (must be a preset like alarm/bell/complete/message, or a valid audio file path)", sound)
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
