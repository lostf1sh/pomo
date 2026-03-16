package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/store"
	"github.com/lostf1sh/pomo/internal/tui"
	"github.com/spf13/cobra"
)

var (
	taskFlag string
	workFlag time.Duration
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a pomodoro session",
	Long:  "Start a pomodoro timer with an interactive TUI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if workFlag > 0 {
			cfg.WorkDuration = workFlag
		}

		dataDir := config.DataDir()
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			return fmt.Errorf("creating data directory: %w", err)
		}

		dbPath := filepath.Join(dataDir, "pomo.db")
		s, err := store.New(dbPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer s.Close()

		m := tui.NewModel(cfg, s, taskFlag)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}

		return nil
	},
}

func init() {
	startCmd.Flags().StringVarP(&taskFlag, "task", "t", "", "Task name for this pomodoro session")
	startCmd.Flags().DurationVarP(&workFlag, "work", "w", 0, "Override work duration (e.g., 30m)")
	rootCmd.AddCommand(startCmd)
}
