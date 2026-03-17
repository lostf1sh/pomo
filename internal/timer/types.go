package timer

import "time"

// SessionType represents the type of a pomodoro segment.
type SessionType int

const (
	Work SessionType = iota
	ShortBreak
	LongBreak
)

func (s SessionType) String() string {
	switch s {
	case Work:
		return "work"
	case ShortBreak:
		return "short-break"
	case LongBreak:
		return "long-break"
	default:
		return "unknown"
	}
}

// TimerState represents the current state of the timer.
type TimerState int

const (
	Idle TimerState = iota
	Running
	Paused
)

func (t TimerState) String() string {
	switch t {
	case Idle:
		return "idle"
	case Running:
		return "running"
	case Paused:
		return "paused"
	default:
		return "unknown"
	}
}

// Session represents a completed or in-progress pomodoro session.
type Session struct {
	ID        string      `json:"id"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Type      SessionType `json:"type"`
	Task      string      `json:"task"`
	Completed bool        `json:"completed"`
}

// Snapshot represents resumable timer state persisted between CLI runs.
type Snapshot struct {
	State          TimerState    `json:"state"`
	CurrentType    SessionType   `json:"current_type"`
	Task           string        `json:"task"`
	Remaining      time.Duration `json:"remaining"`
	TotalDuration  time.Duration `json:"total_duration"`
	PomodorosInSet int           `json:"pomodoros_in_set"`
	CompletedTotal int           `json:"completed_total"`
	CurrentSession *Session      `json:"current_session,omitempty"`
}
