package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/theme"
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "List, preview, or set color themes",
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available color themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, name := range theme.Names() {
			fmt.Println(name)
		}
		return nil
	},
}

var themeSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set the active color theme (same as pomo config set --theme)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, ok := theme.Get(name); !ok {
			return fmt.Errorf("unknown theme %q; available: %s", name, strings.Join(theme.Names(), ", "))
		}
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		cfg.Theme = name
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Theme set to %q.\n", name)
		return nil
	},
}

var themePreviewCmd = &cobra.Command{
	Use:   "preview <name>",
	Short: "Preview theme colors in the terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		th, ok := theme.Get(args[0])
		if !ok {
			return fmt.Errorf("unknown theme %q; available: %s", args[0], strings.Join(theme.Names(), ", "))
		}
		fmt.Printf("Theme: %s\n\n", th.Name)

		line := func(label, hex string) {
			st := lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
			fmt.Printf("%-14s %s\n", label+":", st.Render(hex))
		}
		line("Work", th.Work)
		line("Short break", th.ShortBreak)
		line("Long break", th.LongBreak)
		line("Muted", th.Muted)
		line("Text", th.Text)
		line("Accent", th.Accent)
		line("Progress FG", th.ProgressFG)
		line("Progress BG", th.ProgressBG)

		width := 24
		filled := int(0.45 * float64(width))
		bar := lipgloss.NewStyle().Foreground(lipgloss.Color(th.ProgressFG)).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(th.ProgressBG)).Render(strings.Repeat("░", width-filled))
		fmt.Printf("\nSample bar:  %s 45%%\n", bar)
		return nil
	},
}

func init() {
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeSetCmd)
	themeCmd.AddCommand(themePreviewCmd)
	rootCmd.AddCommand(themeCmd)
}
