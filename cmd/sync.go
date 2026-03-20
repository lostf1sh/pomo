package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/sync"
	"github.com/spf13/cobra"
)

var (
	syncOutputPath string
	syncReplace    bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Export, import, or inspect session backup data",
}

var syncExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export config and sessions as JSON (stdout unless --output)",
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

		sessions, err := s.GetAllSessions()
		if err != nil {
			return fmt.Errorf("loading sessions: %w", err)
		}

		w := os.Stdout
		if syncOutputPath != "" {
			f, err := os.Create(syncOutputPath)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()
			w = f
		}

		if err := sync.Export(cfg, sessions, w); err != nil {
			return fmt.Errorf("export: %w", err)
		}
		if syncOutputPath != "" {
			fmt.Fprintf(os.Stderr, "Wrote export to %s\n", syncOutputPath)
		}
		return nil
	},
}

var syncImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import sessions from a JSON export (merge by default)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		export, err := sync.Import(bytes.NewReader(data))
		if err != nil {
			return err
		}

		s, err := openStore()
		if err != nil {
			return err
		}
		defer s.Close()

		total := len(export.Sessions)
		var added, skipped int

		if syncReplace {
			if err := s.ReplaceAllSessions(export.Sessions); err != nil {
				return fmt.Errorf("replace sessions: %w", err)
			}
			added = total
			skipped = 0
		} else {
			var impErr error
			added, skipped, impErr = s.ImportSessions(export.Sessions)
			if impErr != nil {
				return fmt.Errorf("import sessions: %w", impErr)
			}
		}

		fmt.Printf("Imported %d sessions (%d new, %d skipped)\n", total, added, skipped)
		return nil
	},
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database path, session count, and date range",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := filepath.Join(config.DataDir(), "pomo.db")

		s, err := openStore()
		if err != nil {
			return err
		}
		defer s.Close()

		n, err := s.CountSessions()
		if err != nil {
			return fmt.Errorf("count sessions: %w", err)
		}

		var rangeStr string
		oldest, newest, err := s.DateRange()
		if err != nil {
			return fmt.Errorf("date range: %w", err)
		}
		if n == 0 {
			rangeStr = "—"
		} else {
			rangeStr = fmt.Sprintf("%s — %s", oldest.Format("2006-01-02"), newest.Format("2006-01-02"))
		}

		var sizeStr string
		if fi, err := os.Stat(dbPath); err == nil {
			sizeStr = formatBytes(fi.Size())
		} else {
			sizeStr = "unknown"
		}

		fmt.Printf("Database: %s\n", dbPath)
		fmt.Printf("Total sessions: %d\n", n)
		fmt.Printf("Date range: %s\n", rangeStr)
		fmt.Printf("Database size: %s\n", sizeStr)
		return nil
	},
}

func formatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%d KB", n/1024)
	}
	return fmt.Sprintf("%d MB", n/(1024*1024))
}

func init() {
	syncExportCmd.Flags().StringVar(&syncOutputPath, "output", "", "Write export to file instead of stdout")
	syncImportCmd.Flags().BoolVar(&syncReplace, "replace", false, "Replace all sessions instead of merging")

	syncCmd.AddCommand(syncExportCmd)
	syncCmd.AddCommand(syncImportCmd)
	syncCmd.AddCommand(syncStatusCmd)
	rootCmd.AddCommand(syncCmd)
}
