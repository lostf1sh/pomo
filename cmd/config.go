package cmd

import (
	"fmt"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or modify configuration",
	Long:  "Display current configuration or modify settings.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Println()
		fmt.Println("  Pomodoro Configuration")
		fmt.Println("  " + repeatStr("─", 40))
		fmt.Printf("  Work Duration:        %s\n", cfg.WorkDuration)
		fmt.Printf("  Short Break Duration: %s\n", cfg.ShortBreakDuration)
		fmt.Printf("  Long Break Duration:  %s\n", cfg.LongBreakDuration)
		fmt.Printf("  Long Break Interval:  %d pomodoros\n", cfg.LongBreakInterval)
		if cfg.DailyGoalPomodoros > 0 {
			fmt.Printf("  Daily Goal:           %d pomodoros\n", cfg.DailyGoalPomodoros)
		} else {
			fmt.Println("  Daily Goal:           disabled")
		}
		fmt.Printf("  Desktop Notifications: %v\n", cfg.NotifyDesktop)
		fmt.Printf("  Bell Notifications:    %v\n", cfg.NotifyBell)
		fmt.Println()
		fmt.Printf("  Config file: %s\n", config.ConfigDir())
		fmt.Printf("  Data file:   %s\n", config.DataDir())
		fmt.Println()

		return nil
	},
}

var (
	setWorkFlag       time.Duration
	setShortBreakFlag time.Duration
	setLongBreakFlag  time.Duration
	setIntervalFlag   int
	setDailyGoalFlag  int
	setDesktopFlag    string
	setBellFlag       string
)

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Modify configuration",
	Long:  "Set configuration values. Example: pomo config set --work 30m --short-break 10m",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		changed := false

		if cmd.Flags().Changed("work") {
			cfg.WorkDuration = setWorkFlag
			changed = true
		}
		if cmd.Flags().Changed("short-break") {
			cfg.ShortBreakDuration = setShortBreakFlag
			changed = true
		}
		if cmd.Flags().Changed("long-break") {
			cfg.LongBreakDuration = setLongBreakFlag
			changed = true
		}
		if cmd.Flags().Changed("interval") {
			cfg.LongBreakInterval = setIntervalFlag
			changed = true
		}
		if cmd.Flags().Changed("daily-goal") {
			if setDailyGoalFlag < 0 {
				return fmt.Errorf("daily goal must be 0 or greater")
			}
			cfg.DailyGoalPomodoros = setDailyGoalFlag
			changed = true
		}
		if cmd.Flags().Changed("desktop") {
			cfg.NotifyDesktop = setDesktopFlag == "true" || setDesktopFlag == "on"
			changed = true
		}
		if cmd.Flags().Changed("bell") {
			cfg.NotifyBell = setBellFlag == "true" || setBellFlag == "on"
			changed = true
		}

		if !changed {
			fmt.Println("No changes specified. Use --help to see available options.")
			return nil
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Println("Configuration updated successfully.")
		return nil
	},
}

func init() {
	configSetCmd.Flags().DurationVar(&setWorkFlag, "work", 0, "Work duration (e.g., 25m)")
	configSetCmd.Flags().DurationVar(&setShortBreakFlag, "short-break", 0, "Short break duration (e.g., 5m)")
	configSetCmd.Flags().DurationVar(&setLongBreakFlag, "long-break", 0, "Long break duration (e.g., 15m)")
	configSetCmd.Flags().IntVar(&setIntervalFlag, "interval", 0, "Pomodoros before long break")
	configSetCmd.Flags().IntVar(&setDailyGoalFlag, "daily-goal", 0, "Daily pomodoro goal (0 disables)")
	configSetCmd.Flags().StringVar(&setDesktopFlag, "desktop", "", "Desktop notifications (true/false)")
	configSetCmd.Flags().StringVar(&setBellFlag, "bell", "", "Bell notifications (true/false)")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
