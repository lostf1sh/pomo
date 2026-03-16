package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pomo",
	Short: "A terminal-based pomodoro timer",
	Long:  "Pomodoro CLI - A terminal-based pomodoro timer with TUI, session tracking, and statistics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to start when no subcommand is given
		return startCmd.RunE(cmd, args)
	},
}

func Execute() error {
	return rootCmd.Execute()
}
