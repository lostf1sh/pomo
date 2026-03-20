package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	WorkDuration       time.Duration `json:"work_duration"`
	ShortBreakDuration time.Duration `json:"short_break_duration"`
	LongBreakDuration  time.Duration `json:"long_break_duration"`
	LongBreakInterval  int           `json:"long_break_interval"`
	DailyGoalPomodoros int           `json:"daily_goal_pomodoros"`
	NotifyDesktop      bool          `json:"notify_desktop"`
	NotifyBell         bool          `json:"notify_bell"`
	Theme              string        `json:"theme"`
}

func DefaultConfig() Config {
	return Config{
		WorkDuration:       25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakInterval:  4,
		DailyGoalPomodoros: 0,
		NotifyDesktop:      true,
		NotifyBell:         true,
		Theme:              "default",
	}
}

func DataDir() string {
	if dir := os.Getenv("POMO_DATA_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "pomo")
}

func ConfigDir() string {
	if dir := os.Getenv("POMO_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pomo")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// configJSON is the on-disk representation with durations as strings.
type configJSON struct {
	WorkDuration       string `json:"work_duration"`
	ShortBreakDuration string `json:"short_break_duration"`
	LongBreakDuration  string `json:"long_break_duration"`
	LongBreakInterval  int    `json:"long_break_interval"`
	DailyGoalPomodoros int    `json:"daily_goal_pomodoros"`
	NotifyDesktop      bool   `json:"notify_desktop"`
	NotifyBell         bool   `json:"notify_bell"`
	Theme              string `json:"theme"`
}

func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	var raw configJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return cfg, err
	}

	if raw.WorkDuration != "" {
		if d, err := time.ParseDuration(raw.WorkDuration); err == nil {
			cfg.WorkDuration = d
		}
	}
	if raw.ShortBreakDuration != "" {
		if d, err := time.ParseDuration(raw.ShortBreakDuration); err == nil {
			cfg.ShortBreakDuration = d
		}
	}
	if raw.LongBreakDuration != "" {
		if d, err := time.ParseDuration(raw.LongBreakDuration); err == nil {
			cfg.LongBreakDuration = d
		}
	}
	if raw.LongBreakInterval > 0 {
		cfg.LongBreakInterval = raw.LongBreakInterval
	}
	if raw.DailyGoalPomodoros >= 0 {
		cfg.DailyGoalPomodoros = raw.DailyGoalPomodoros
	}
	cfg.NotifyDesktop = raw.NotifyDesktop
	cfg.NotifyBell = raw.NotifyBell
	cfg.Theme = raw.Theme
	if cfg.Theme == "" {
		cfg.Theme = "default"
	}

	return cfg, nil
}

func SaveConfig(cfg Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	raw := configJSON{
		WorkDuration:       cfg.WorkDuration.String(),
		ShortBreakDuration: cfg.ShortBreakDuration.String(),
		LongBreakDuration:  cfg.LongBreakDuration.String(),
		LongBreakInterval:  cfg.LongBreakInterval,
		DailyGoalPomodoros: cfg.DailyGoalPomodoros,
		NotifyDesktop:      cfg.NotifyDesktop,
		NotifyBell:         cfg.NotifyBell,
		Theme:              cfg.Theme,
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0o644)
}
