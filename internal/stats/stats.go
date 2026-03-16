package stats

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lostf1sh/pomo/internal/timer"
)

type TaskStat struct {
	Task       string
	Pomodoros  int
	TotalTime  time.Duration
}

type Stats struct {
	TotalPomodoros int
	TotalWorkTime  time.Duration
	AvgPerDay      float64
	CurrentStreak  int
	LongestStreak  int
	TaskBreakdown  []TaskStat
}

func Compute(sessions []timer.Session) *Stats {
	s := &Stats{}

	taskMap := make(map[string]*TaskStat)
	daySet := make(map[string]bool)

	for _, sess := range sessions {
		if sess.Type != timer.Work || !sess.Completed {
			continue
		}
		s.TotalPomodoros++
		duration := sess.EndTime.Sub(sess.StartTime)
		s.TotalWorkTime += duration

		day := sess.StartTime.Format("2006-01-02")
		daySet[day] = true

		task := sess.Task
		if task == "" {
			task = "(no task)"
		}
		if ts, ok := taskMap[task]; ok {
			ts.Pomodoros++
			ts.TotalTime += duration
		} else {
			taskMap[task] = &TaskStat{
				Task:      task,
				Pomodoros: 1,
				TotalTime: duration,
			}
		}
	}

	if len(daySet) > 0 {
		s.AvgPerDay = float64(s.TotalPomodoros) / float64(len(daySet))
	}

	for _, ts := range taskMap {
		s.TaskBreakdown = append(s.TaskBreakdown, *ts)
	}
	sort.Slice(s.TaskBreakdown, func(i, j int) bool {
		return s.TaskBreakdown[i].Pomodoros > s.TaskBreakdown[j].Pomodoros
	})

	s.CurrentStreak, s.LongestStreak = computeStreaks(daySet)

	return s
}

func computeStreaks(daySet map[string]bool) (current, longest int) {
	if len(daySet) == 0 {
		return 0, 0
	}

	var days []time.Time
	for d := range daySet {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		days = append(days, t)
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})

	// Compute longest streak
	streak := 1
	longest = 1
	for i := 1; i < len(days); i++ {
		diff := days[i].Sub(days[i-1]).Hours() / 24
		if diff <= 1.5 { // Account for DST
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 1
		}
	}

	// Current streak: count backwards from today
	today := time.Now().Truncate(24 * time.Hour)
	current = 0
	checkDay := today
	for {
		dayStr := checkDay.Format("2006-01-02")
		if daySet[dayStr] {
			current++
			checkDay = checkDay.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return current, longest
}

func FormatStats(s *Stats) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("  Total Pomodoros:   %d\n", s.TotalPomodoros))
	b.WriteString(fmt.Sprintf("  Total Work Time:   %s\n", formatDuration(s.TotalWorkTime)))
	b.WriteString(fmt.Sprintf("  Avg Per Day:       %.1f\n", s.AvgPerDay))
	b.WriteString(fmt.Sprintf("  Current Streak:    %d day(s)\n", s.CurrentStreak))
	b.WriteString(fmt.Sprintf("  Longest Streak:    %d day(s)\n", s.LongestStreak))

	if len(s.TaskBreakdown) > 0 {
		b.WriteString("\n  Task Breakdown:\n")
		for _, ts := range s.TaskBreakdown {
			b.WriteString(fmt.Sprintf("    %-20s %3d pomodoros  %s\n",
				ts.Task, ts.Pomodoros, formatDuration(ts.TotalTime)))
		}
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
