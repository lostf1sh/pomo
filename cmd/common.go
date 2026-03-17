package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/store"
)

func openStore() (*store.Store, error) {
	dataDir := config.DataDir()
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "pomo.db")
	s, err := store.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	return s, nil
}
