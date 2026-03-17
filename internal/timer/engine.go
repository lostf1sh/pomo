package timer

import (
	"fmt"
	"time"

	"github.com/lostf1sh/pomo/internal/config"
)

type Engine struct {
	Config         config.Config
	State          TimerState
	CurrentType    SessionType
	Task           string
	Remaining      time.Duration
	TotalDuration  time.Duration
	PomodorosInSet int
	CompletedTotal int

	startTime      time.Time
	endTime        time.Time
	currentSession *Session
}

func NewEngine(cfg config.Config, task string) *Engine {
	return &Engine{
		Config:        cfg,
		State:         Idle,
		CurrentType:   Work,
		Task:          task,
		Remaining:     cfg.WorkDuration,
		TotalDuration: cfg.WorkDuration,
	}
}

func NewEngineFromSnapshot(cfg config.Config, snapshot *Snapshot) *Engine {
	e := &Engine{
		Config:         cfg,
		State:          snapshot.State,
		CurrentType:    snapshot.CurrentType,
		Task:           snapshot.Task,
		Remaining:      snapshot.Remaining,
		TotalDuration:  snapshot.TotalDuration,
		PomodorosInSet: snapshot.PomodorosInSet,
		CompletedTotal: snapshot.CompletedTotal,
	}

	if e.TotalDuration <= 0 {
		e.TotalDuration = e.durationForType(e.CurrentType)
	}

	if snapshot.CurrentSession != nil {
		sess := *snapshot.CurrentSession
		e.currentSession = &sess
		e.startTime = sess.StartTime
	}

	if e.State == Running {
		e.endTime = time.Now().Add(e.Remaining)
	}

	return e
}

func (e *Engine) Start() {
	switch e.State {
	case Idle:
		e.startTime = time.Now()
		e.endTime = e.startTime.Add(e.Remaining)
		e.State = Running
		e.currentSession = &Session{
			ID:        fmt.Sprintf("%d", e.startTime.UnixNano()),
			StartTime: e.startTime,
			Type:      e.CurrentType,
			Task:      e.Task,
		}
	case Paused:
		now := time.Now()
		e.endTime = now.Add(e.Remaining)
		e.State = Running
	}
}

func (e *Engine) Pause() {
	if e.State == Running {
		e.Remaining = time.Until(e.endTime)
		if e.Remaining < 0 {
			e.Remaining = 0
		}
		e.State = Paused
	}
}

func (e *Engine) Reset() {
	e.State = Idle
	e.Remaining = e.durationForType(e.CurrentType)
	e.TotalDuration = e.Remaining
	e.currentSession = nil
}

func (e *Engine) Skip() *Session {
	var completed *Session
	if e.CurrentType == Work && e.currentSession != nil {
		e.currentSession.EndTime = time.Now()
		e.currentSession.Completed = false
		completed = e.currentSession
	}
	e.advance()
	return completed
}

// Tick updates the timer. Returns a completed session when a segment finishes.
func (e *Engine) Tick() *Session {
	if e.State != Running {
		return nil
	}

	e.Remaining = time.Until(e.endTime)
	if e.Remaining < 0 {
		e.Remaining = 0
	}

	if e.Remaining <= 0 {
		var completed *Session
		if e.currentSession != nil {
			e.currentSession.EndTime = time.Now()
			e.currentSession.Completed = true
			completed = e.currentSession
		}

		if e.CurrentType == Work {
			e.PomodorosInSet++
			e.CompletedTotal++
		}

		e.advance()
		return completed
	}

	return nil
}

func (e *Engine) advance() {
	switch e.CurrentType {
	case Work:
		if e.PomodorosInSet >= e.Config.LongBreakInterval {
			e.CurrentType = LongBreak
			e.PomodorosInSet = 0
		} else {
			e.CurrentType = ShortBreak
		}
	case ShortBreak, LongBreak:
		e.CurrentType = Work
	}

	e.Remaining = e.durationForType(e.CurrentType)
	e.TotalDuration = e.Remaining
	e.State = Idle
	e.currentSession = nil
}

func (e *Engine) Progress() float64 {
	if e.TotalDuration == 0 {
		return 0
	}
	elapsed := e.TotalDuration - e.Remaining
	return float64(elapsed) / float64(e.TotalDuration)
}

func (e *Engine) Snapshot() *Snapshot {
	remaining := e.Remaining
	if e.State == Running {
		remaining = time.Until(e.endTime)
		if remaining < 0 {
			remaining = 0
		}
	}

	snapshot := &Snapshot{
		State:          e.State,
		CurrentType:    e.CurrentType,
		Task:           e.Task,
		Remaining:      remaining,
		TotalDuration:  e.TotalDuration,
		PomodorosInSet: e.PomodorosInSet,
		CompletedTotal: e.CompletedTotal,
	}

	if e.currentSession != nil {
		sess := *e.currentSession
		snapshot.CurrentSession = &sess
	}

	return snapshot
}

func (e *Engine) durationForType(st SessionType) time.Duration {
	switch st {
	case Work:
		return e.Config.WorkDuration
	case ShortBreak:
		return e.Config.ShortBreakDuration
	case LongBreak:
		return e.Config.LongBreakDuration
	default:
		return e.Config.WorkDuration
	}
}
