package cmd

import (
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
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show pomodoro statistics",
	Long:  "Display statistics about your pomodoro sessions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := config.DataDir()
		dbPath := filepath.Join(dataDir, "pomo.db")

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Println("No sessions found. Start your first pomodoro with: pomo start")
			return nil
		}

		s, err := store.New(dbPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer s.Close()

		from, to := periodRange(statsPeriodFlag)

		var sessions []timer.Session
		if statsTaskFlag != "" {
			sessions, err = s.GetSessionsByTask(statsTaskFlag, from, to)
		} else {
			sessions, err = s.GetSessions(from, to)
		}
		if err != nil {
			return fmt.Errorf("fetching sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found for the specified criteria.")
			return nil
		}

		computed := stats.Compute(sessions)

		fmt.Println()
		fmt.Printf("  Pomodoro Statistics (%s)\n", periodLabel(statsPeriodFlag))
		if statsTaskFlag != "" {
			fmt.Printf("  Task: %s\n", statsTaskFlag)
		}
		fmt.Println("  " + repeatStr("─", 40))
		fmt.Println(stats.FormatStats(computed))

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

func init() {
	statsCmd.Flags().StringVarP(&statsTaskFlag, "task", "t", "", "Filter by task name")
	statsCmd.Flags().StringVarP(&statsPeriodFlag, "period", "p", "all", "Time period: today, week, month, all")
	rootCmd.AddCommand(statsCmd)
}
