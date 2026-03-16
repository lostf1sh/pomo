package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorkDuration != 25*time.Minute {
		t.Errorf("expected work duration 25m, got %v", cfg.WorkDuration)
	}
	if cfg.ShortBreakDuration != 5*time.Minute {
		t.Errorf("expected short break 5m, got %v", cfg.ShortBreakDuration)
	}
	if cfg.LongBreakDuration != 15*time.Minute {
		t.Errorf("expected long break 15m, got %v", cfg.LongBreakDuration)
	}
	if cfg.LongBreakInterval != 4 {
		t.Errorf("expected long break interval 4, got %d", cfg.LongBreakInterval)
	}
	if !cfg.NotifyDesktop {
		t.Error("expected notify desktop to be true")
	}
	if !cfg.NotifyBell {
		t.Error("expected notify bell to be true")
	}
}

func TestLoadConfigDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("POMO_CONFIG_DIR", dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkDuration != 25*time.Minute {
		t.Errorf("expected default work duration, got %v", cfg.WorkDuration)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("POMO_CONFIG_DIR", dir)

	cfg := DefaultConfig()
	cfg.WorkDuration = 30 * time.Minute
	cfg.NotifyDesktop = false

	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		t.Fatal("config file not created")
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.WorkDuration != 30*time.Minute {
		t.Errorf("expected work duration 30m, got %v", loaded.WorkDuration)
	}
	if loaded.NotifyDesktop {
		t.Error("expected notify desktop to be false")
	}
}

func TestDataDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("POMO_DATA_DIR", dir)

	if DataDir() != dir {
		t.Errorf("expected %s, got %s", dir, DataDir())
	}
}
