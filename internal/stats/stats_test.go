package stats

import (
	"testing"
	"time"

	"github.com/lostf1sh/pomo/internal/timer"
)

func TestComputeEmpty(t *testing.T) {
	s := Compute(nil)
	if s.TotalPomodoros != 0 {
		t.Errorf("expected 0 pomodoros, got %d", s.TotalPomodoros)
	}
}

func TestComputeBasic(t *testing.T) {
	now := time.Now()
	sessions := []timer.Session{
		{
			ID:        "1",
			StartTime: now,
			EndTime:   now.Add(25 * time.Minute),
			Type:      timer.Work,
			Task:      "coding",
			Completed: true,
		},
		{
			ID:        "2",
			StartTime: now.Add(30 * time.Minute),
			EndTime:   now.Add(55 * time.Minute),
			Type:      timer.Work,
			Task:      "coding",
			Completed: true,
		},
		{
			ID:        "3",
			StartTime: now.Add(60 * time.Minute),
			EndTime:   now.Add(85 * time.Minute),
			Type:      timer.Work,
			Task:      "reading",
			Completed: true,
		},
		// Incomplete session should not count
		{
			ID:        "4",
			StartTime: now.Add(90 * time.Minute),
			EndTime:   now.Add(100 * time.Minute),
			Type:      timer.Work,
			Task:      "coding",
			Completed: false,
		},
		// Break sessions should not count
		{
			ID:        "5",
			StartTime: now.Add(110 * time.Minute),
			EndTime:   now.Add(115 * time.Minute),
			Type:      timer.ShortBreak,
			Task:      "",
			Completed: true,
		},
	}

	s := Compute(sessions)

	if s.TotalPomodoros != 3 {
		t.Errorf("expected 3 pomodoros, got %d", s.TotalPomodoros)
	}

	expectedWork := 75 * time.Minute
	if s.TotalWorkTime != expectedWork {
		t.Errorf("expected %v work time, got %v", expectedWork, s.TotalWorkTime)
	}

	if len(s.TaskBreakdown) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(s.TaskBreakdown))
	}

	// First task should be "coding" (2 pomodoros, sorted by count)
	if s.TaskBreakdown[0].Task != "coding" {
		t.Errorf("expected first task to be coding, got %s", s.TaskBreakdown[0].Task)
	}
	if s.TaskBreakdown[0].Pomodoros != 2 {
		t.Errorf("expected 2 coding pomodoros, got %d", s.TaskBreakdown[0].Pomodoros)
	}
}

func TestComputeStreak(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	sessions := []timer.Session{
		{ID: "1", StartTime: today.Add(10 * time.Hour), EndTime: today.Add(10*time.Hour + 25*time.Minute), Type: timer.Work, Task: "a", Completed: true},
		{ID: "2", StartTime: today.AddDate(0, 0, -1).Add(10 * time.Hour), EndTime: today.AddDate(0, 0, -1).Add(10*time.Hour + 25*time.Minute), Type: timer.Work, Task: "a", Completed: true},
		{ID: "3", StartTime: today.AddDate(0, 0, -2).Add(10 * time.Hour), EndTime: today.AddDate(0, 0, -2).Add(10*time.Hour + 25*time.Minute), Type: timer.Work, Task: "a", Completed: true},
		// Gap on day -3
		{ID: "4", StartTime: today.AddDate(0, 0, -5).Add(10 * time.Hour), EndTime: today.AddDate(0, 0, -5).Add(10*time.Hour + 25*time.Minute), Type: timer.Work, Task: "a", Completed: true},
		{ID: "5", StartTime: today.AddDate(0, 0, -4).Add(10 * time.Hour), EndTime: today.AddDate(0, 0, -4).Add(10*time.Hour + 25*time.Minute), Type: timer.Work, Task: "a", Completed: true},
	}

	s := Compute(sessions)

	if s.CurrentStreak != 3 {
		t.Errorf("expected current streak of 3, got %d", s.CurrentStreak)
	}
	if s.LongestStreak != 3 {
		t.Errorf("expected longest streak of 3, got %d", s.LongestStreak)
	}
}

func TestFormatStats(t *testing.T) {
	s := &Stats{
		TotalPomodoros: 10,
		TotalWorkTime:  250 * time.Minute,
		AvgPerDay:      3.3,
		CurrentStreak:  2,
		LongestStreak:  5,
		TaskBreakdown: []TaskStat{
			{Task: "coding", Pomodoros: 7, TotalTime: 175 * time.Minute},
			{Task: "reading", Pomodoros: 3, TotalTime: 75 * time.Minute},
		},
	}

	result := FormatStats(s)
	if result == "" {
		t.Error("expected non-empty formatted stats")
	}
}

func TestComputeGoalProgress(t *testing.T) {
	now := time.Now()
	progress := ComputeGoalProgress(3, []timer.Session{
		{ID: "1", StartTime: now, EndTime: now.Add(25 * time.Minute), Type: timer.Work, Completed: true},
		{ID: "2", StartTime: now.Add(time.Hour), EndTime: now.Add(85 * time.Minute), Type: timer.Work, Completed: true},
		{ID: "3", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(125 * time.Minute), Type: timer.ShortBreak, Completed: true},
	})

	if progress == nil {
		t.Fatal("expected progress")
	}
	if progress.Completed != 2 {
		t.Fatalf("expected completed 2, got %d", progress.Completed)
	}
	if progress.Remaining != 1 {
		t.Fatalf("expected remaining 1, got %d", progress.Remaining)
	}
	if progress.Met {
		t.Fatal("expected goal not met")
	}
}
