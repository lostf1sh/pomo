package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
	"github.com/lostf1sh/pomo/internal/timer"
)

const exportVersion = 1

// ExportData is the JSON format for backup / restore.
type ExportData struct {
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exported_at"`
	Config     config.Config   `json:"config"`
	Sessions   []timer.Session `json:"sessions"`
}

// Export writes config and sessions as JSON to w.
func Export(cfg config.Config, sessions []timer.Session, w io.Writer) error {
	data := ExportData{
		Version:    exportVersion,
		ExportedAt: time.Now().UTC(),
		Config:     cfg,
		Sessions:   sessions,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}

// Import reads and validates export JSON.
func Import(r io.Reader) (*ExportData, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var data ExportData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse export: %w", err)
	}
	if data.Version != exportVersion {
		return nil, fmt.Errorf("unsupported export version %d (supported: %d)", data.Version, exportVersion)
	}
	if data.Sessions == nil {
		data.Sessions = []timer.Session{}
	}
	return &data, nil
}
