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

func TestActiveStateRoundTrip(t *testing.T) {
	s := newTestStore(t)

	snapshot := timer.Snapshot{
		State:          timer.Paused,
		CurrentType:    timer.Work,
		Task:           "writing",
		Remaining:      12 * time.Minute,
		TotalDuration:  25 * time.Minute,
		PomodorosInSet: 2,
		CompletedTotal: 5,
		CurrentSession: &timer.Session{
			ID:        "resume-1",
			StartTime: time.Now().Add(-13 * time.Minute),
			Type:      timer.Work,
			Task:      "writing",
		},
	}

	if err := s.SaveActiveState(snapshot); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.GetActiveState()
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("expected active state")
	}
	if loaded.Task != "writing" {
		t.Fatalf("expected task writing, got %s", loaded.Task)
	}
	if loaded.CurrentSession == nil || loaded.CurrentSession.ID != "resume-1" {
		t.Fatal("expected current session to round trip")
	}

	if err := s.ClearActiveState(); err != nil {
		t.Fatal(err)
	}

	cleared, err := s.GetActiveState()
	if err != nil {
		t.Fatal(err)
	}
	if cleared != nil {
		t.Fatal("expected cleared active state to be nil")
	}
}

func TestImportSessionsMerge(t *testing.T) {
	s := newTestStore(t)
	base := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	existing := timer.Session{
		ID:        "e1",
		StartTime: base,
		EndTime:   base.Add(25 * time.Minute),
		Type:      timer.Work,
		Task:      "a",
		Completed: true,
	}
	if err := s.SaveSession(existing); err != nil {
		t.Fatal(err)
	}

	incoming := []timer.Session{
		existing, // duplicate key
		{
			ID:        "e2",
			StartTime: base.Add(time.Hour),
			EndTime:   base.Add(time.Hour + 25*time.Minute),
			Type:      timer.Work,
			Task:      "b",
			Completed: true,
		},
	}

	added, skipped, err := s.ImportSessions(incoming)
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 || skipped != 1 {
		t.Fatalf("added=%d skipped=%d, want 1,1", added, skipped)
	}

	all, err := s.GetAllSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}
}

func TestReplaceAllSessions(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()
	for i := 0; i < 3; i++ {
		sess := timer.Session{
			ID:        string(rune('x' + i)),
			StartTime: now.Add(time.Duration(i) * time.Hour),
			EndTime:   now.Add(time.Duration(i)*time.Hour + 25*time.Minute),
			Type:      timer.Work,
			Task:      "old",
			Completed: true,
		}
		if err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	replacement := []timer.Session{
		{
			ID:        "only",
			StartTime: now.Add(100 * time.Hour),
			EndTime:   now.Add(100*time.Hour + 25*time.Minute),
			Type:      timer.Work,
			Task:      "new",
			Completed: true,
		},
	}
	if err := s.ReplaceAllSessions(replacement); err != nil {
		t.Fatal(err)
	}

	all, err := s.GetAllSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Task != "new" {
		t.Fatalf("got %+v", all)
	}
}

func TestCountSessions(t *testing.T) {
	s := newTestStore(t)
	n, err := s.CountSessions()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("count %d", n)
	}
	now := time.Now()
	if err := s.SaveSession(timer.Session{
		ID: "1", StartTime: now, EndTime: now.Add(time.Minute), Type: timer.Work, Task: "t", Completed: true,
	}); err != nil {
		t.Fatal(err)
	}
	n, err = s.CountSessions()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count %d", n)
	}
}

func TestDateRange(t *testing.T) {
	s := newTestStore(t)
	a := time.Date(2023, 1, 10, 12, 0, 0, 0, time.UTC)
	b := time.Date(2024, 2, 20, 12, 0, 0, 0, time.UTC)
	for _, st := range []time.Time{b, a} {
		sess := timer.Session{
			ID:        st.Format(time.RFC3339Nano),
			StartTime: st,
			EndTime:   st.Add(time.Minute),
			Type:      timer.Work,
			Task:      "t",
			Completed: true,
		}
		if err := s.SaveSession(sess); err != nil {
			t.Fatal(err)
		}
	}

	oldest, newest, err := s.DateRange()
	if err != nil {
		t.Fatal(err)
	}
	if !oldest.Equal(a) || !newest.Equal(b) {
		t.Fatalf("range %v — %v, want %v — %v", oldest, newest, a, b)
	}
}

func TestDateRangeEmpty(t *testing.T) {
	s := newTestStore(t)
	oldest, newest, err := s.DateRange()
	if err != nil {
		t.Fatal(err)
	}
	if !oldest.IsZero() || !newest.IsZero() {
		t.Fatalf("expected zero times, got %v %v", oldest, newest)
	}
}
