package sync

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/timer"
)

func TestExportImportRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDuration = 30 * time.Minute
	cfg.Theme = "dracula"

	now := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	sessions := []timer.Session{
		{
			ID:        "a",
			StartTime: now,
			EndTime:   now.Add(25 * time.Minute),
			Type:      timer.Work,
			Task:      "code",
			Completed: true,
		},
	}

	var buf bytes.Buffer
	if err := Export(cfg, sessions, &buf); err != nil {
		t.Fatal(err)
	}

	data, err := Import(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if data.Version != 1 {
		t.Errorf("version %d", data.Version)
	}
	if data.Config.WorkDuration != cfg.WorkDuration {
		t.Errorf("config work: got %v", data.Config.WorkDuration)
	}
	if data.Config.Theme != "dracula" {
		t.Errorf("theme: got %q", data.Config.Theme)
	}
	if len(data.Sessions) != 1 {
		t.Fatalf("sessions len %d", len(data.Sessions))
	}
	if data.Sessions[0].Task != "code" {
		t.Errorf("task %q", data.Sessions[0].Task)
	}
	if !data.Sessions[0].StartTime.Equal(now) {
		t.Errorf("start time mismatch")
	}
}

func TestImportUnknownVersion(t *testing.T) {
	r := strings.NewReader(`{"version": 99, "exported_at": "2025-01-01T00:00:00Z", "config": {}, "sessions": []}`)
	_, err := Import(r)
	if err == nil || !strings.Contains(err.Error(), "unsupported export version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestExportImportEmptySessions(t *testing.T) {
	cfg := config.DefaultConfig()
	var buf bytes.Buffer
	if err := Export(cfg, nil, &buf); err != nil {
		t.Fatal(err)
	}
	data, err := Import(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if data.Sessions == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(data.Sessions) != 0 {
		t.Fatalf("len %d", len(data.Sessions))
	}
}
