package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/stats"
	"github.com/lostf1sh/pomo/internal/store"
	"github.com/lostf1sh/pomo/internal/timer"
	"github.com/spf13/cobra"
)

var (
	statsTaskFlag   string
	statsPeriodFlag string
	statsJSONFlag   bool
)

type statsJSONResponse struct {
	Period    string              `json:"period"`
	Task      string              `json:"task,omitempty"`
	Stats     statsJSONSummary    `json:"stats"`
	DailyGoal *stats.GoalProgress `json:"daily_goal,omitempty"`
}

type statsJSONSummary struct {
	TotalPomodoros   int                      `json:"total_pomodoros"`
	TotalWorkTime    string                   `json:"total_work_time"`
	TotalWorkTimeMin int                      `json:"total_work_time_minutes"`
	AvgPerDay        float64                  `json:"avg_per_day"`
	CurrentStreak    int                      `json:"current_streak"`
	LongestStreak    int                      `json:"longest_streak"`
	TaskBreakdown    []statsJSONTaskBreakdown `json:"task_breakdown"`
}

type statsJSONTaskBreakdown struct {
	Task             string `json:"task"`
	Pomodoros        int    `json:"pomodoros"`
	TotalTime        string `json:"total_time"`
	TotalTimeMinutes int    `json:"total_time_minutes"`
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show pomodoro statistics",
	Long:  "Display statistics about your pomodoro sessions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		dataDir := config.DataDir()
		dbPath := filepath.Join(dataDir, "pomo.db")

		var sessions []timer.Session
		var todaySessions []timer.Session

		if _, err := os.Stat(dbPath); err == nil {
			s, err := store.New(dbPath)
			if err != nil {
				return fmt.Errorf("opening database: %w", err)
			}
			defer s.Close()

			from, to := periodRange(statsPeriodFlag)

			if statsTaskFlag != "" {
				sessions, err = s.GetSessionsByTask(statsTaskFlag, from, to)
			} else {
				sessions, err = s.GetSessions(from, to)
			}
			if err != nil {
				return fmt.Errorf("fetching sessions: %w", err)
			}

			todayFrom, todayTo := periodRange("today")
			todaySessions, err = s.GetSessions(todayFrom, todayTo)
			if err != nil {
				return fmt.Errorf("fetching today's sessions: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("checking database: %w", err)
		}

		if len(sessions) == 0 {
			if statsJSONFlag {
				return printStatsJSON(stats.Compute(nil), stats.ComputeGoalProgress(cfg.DailyGoalPomodoros, todaySessions))
			}
			fmt.Println("No sessions found for the specified criteria.")
			return nil
		}

		computed := stats.Compute(sessions)
		goal := stats.ComputeGoalProgress(cfg.DailyGoalPomodoros, todaySessions)

		if statsJSONFlag {
			return printStatsJSON(computed, goal)
		}

		fmt.Println()
		fmt.Printf("  Pomodoro Statistics (%s)\n", periodLabel(statsPeriodFlag))
		if statsTaskFlag != "" {
			fmt.Printf("  Task: %s\n", statsTaskFlag)
		}
		fmt.Println("  " + repeatStr("─", 40))
		fmt.Println(stats.FormatStats(computed))
		if goal != nil {
			fmt.Printf("  Daily Goal:         %d/%d", goal.Completed, goal.Target)
			if goal.Met {
				fmt.Print(" (met)")
			}
			fmt.Println()
		}

		return nil
	},
}

func periodRange(period string) (time.Time, time.Time) {
	now := time.Now()
	to := now

	switch period {
	case "today":
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return from, to
	case "week":
		from := now.AddDate(0, 0, -7)
		return from, to
	case "month":
		from := now.AddDate(0, -1, 0)
		return from, to
	default: // "all"
		from := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		return from, to
	}
}

func periodLabel(period string) string {
	switch period {
	case "today":
		return "Today"
	case "week":
		return "Last 7 days"
	case "month":
		return "Last 30 days"
	default:
		return "All time"
	}
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func printStatsJSON(computed *stats.Stats, goal *stats.GoalProgress) error {
	response := statsJSONResponse{
		Period:    statsPeriodFlag,
		Task:      statsTaskFlag,
		Stats:     buildStatsJSONSummary(computed),
		DailyGoal: goal,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func buildStatsJSONSummary(computed *stats.Stats) statsJSONSummary {
	summary := statsJSONSummary{
		TotalPomodoros:   computed.TotalPomodoros,
		TotalWorkTime:    formatDurationJSON(computed.TotalWorkTime),
		TotalWorkTimeMin: int(computed.TotalWorkTime.Minutes()),
		AvgPerDay:        computed.AvgPerDay,
		CurrentStreak:    computed.CurrentStreak,
		LongestStreak:    computed.LongestStreak,
		TaskBreakdown:    make([]statsJSONTaskBreakdown, 0, len(computed.TaskBreakdown)),
	}

	for _, task := range computed.TaskBreakdown {
		summary.TaskBreakdown = append(summary.TaskBreakdown, statsJSONTaskBreakdown{
			Task:             task.Task,
			Pomodoros:        task.Pomodoros,
			TotalTime:        formatDurationJSON(task.TotalTime),
			TotalTimeMinutes: int(task.TotalTime.Minutes()),
		})
	}

	return summary
}

func formatDurationJSON(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func init() {
	statsCmd.Flags().StringVarP(&statsTaskFlag, "task", "t", "", "Filter by task name")
	statsCmd.Flags().StringVarP(&statsPeriodFlag, "period", "p", "all", "Time period: today, week, month, all")
	statsCmd.Flags().BoolVar(&statsJSONFlag, "json", false, "Output statistics as JSON")
	rootCmd.AddCommand(statsCmd)
}
