package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/timer"
	"github.com/lostf1sh/pomo/internal/tui"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the last unfinished pomodoro session",
	Long:  "Resume the last unfinished pomodoro session that was saved when the TUI exited.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		s, err := openStore()
		if err != nil {
			return err
		}
		defer s.Close()

		snapshot, err := s.GetActiveState()
		if err != nil {
			return fmt.Errorf("loading active session: %w", err)
		}
		if snapshot == nil {
			fmt.Println("No resumable session found. Start one with: pomo start")
			return nil
		}

		engine := timer.NewEngineFromSnapshot(cfg, snapshot)
		m := tui.NewModelWithEngine(cfg, s, engine)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
