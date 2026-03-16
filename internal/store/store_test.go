package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lostf1sh/pomo/internal/timer"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSaveAndGetSession(t *testing.T) {
	s := newTestStore(t)

	now := time.Now()
	sess := timer.Session{
		ID:        "test-1",
		StartTime: now,
		EndTime:   now.Add(25 * time.Minute),
		Type:      timer.Work,
		Task:      "coding",
		Completed: true,
	}

	if err := s.SaveSession(sess); err != nil {
		t.Fatal(err)
	}

	sessions, err := s.GetSessions(now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "test-1" {
		t.Errorf("expected ID test-1, got %s", sessions[0].ID)
	}
	if sessions[0].Task != "coding" {
		t.Errorf("expected task coding, got %s", sessions[0].Task)
	}
}

func TestGetSessionsByTask(t *testing.T) {
	s := newTestStore(t)

	now := time.Now()
	sessions := []timer.Session{
		{ID: "1", StartTime: now, EndTime: now.Add(25 * time.Minute), Type: timer.Work, Task: "coding", Completed: true},
		{ID: "2", StartTime: now.Add(30 * time.Minute), EndTime: now.Add(55 * time.Minute), Type: timer.Work, Task: "reading", Completed: true},
		{ID: "3", StartTime: now.Add(60 * time.Minute), EndTime: now.Add(85 * time.Minute), Type: timer.Work, Task: "coding", Completed: true},
	}

	for _, sess := range sessions {
		if err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	result, err := s.GetSessionsByTask("coding", now.Add(-time.Hour), now.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 coding sessions, got %d", len(result))
	}
}

func TestGetSessionsRangeFilter(t *testing.T) {
	s := newTestStore(t)

	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		start := base.Add(time.Duration(i) * time.Hour)
		sess := timer.Session{
			ID:        string(rune('a' + i)),
			StartTime: start,
			EndTime:   start.Add(25 * time.Minute),
			Type:      timer.Work,
			Task:      "test",
			Completed: true,
		}
		if err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	// Get sessions from hour 1 to hour 3 (should get 3 sessions: index 1, 2, 3)
	from := base.Add(1 * time.Hour)
	to := base.Add(3 * time.Hour)
	result, err := s.GetSessions(from, to)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 sessions in range, got %d", len(result))
	}
}

func TestGetAllSessions(t *testing.T) {
	s := newTestStore(t)

	now := time.Now()
	for i := 0; i < 3; i++ {
		start := now.Add(time.Duration(i) * time.Hour)
		sess := timer.Session{
			ID:        string(rune('a' + i)),
			StartTime: start,
			EndTime:   start.Add(25 * time.Minute),
			Type:      timer.Work,
			Task:      "test",
			Completed: true,
		}
		if err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	result, err := s.GetAllSessions()
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result))
	}
}
