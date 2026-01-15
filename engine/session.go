// Package engine implements the core pomodoro timing logic.
package engine

import "time"

// Phase represents the current phase of the pomodoro session.
type Phase int

const (
	PhaseWork Phase = iota
	PhaseShortBreak
	PhaseLongBreak
	PhaseDone
)

func (p Phase) String() string {
	switch p {
	case PhaseWork:
		return "Work"
	case PhaseShortBreak:
		return "Short Break"
	case PhaseLongBreak:
		return "Long Break"
	case PhaseDone:
		return "Done"
	default:
		return "Unknown"
	}
}

// Config holds session configuration.
type Config struct {
	WorkDuration       time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
	LongBreakEvery     int
	TotalCycles        int
}

// Session manages pomodoro state transitions.
type Session struct {
	config         Config
	currentPhase   Phase
	cyclesComplete int
	totalPhases    int
	phasesComplete int
}

// NewSession creates a session with the given configuration.
func NewSession(cfg Config) *Session {
	s := &Session{
		config:       cfg,
		currentPhase: PhaseWork,
	}
	s.totalPhases = s.calculateTotalPhases()
	return s
}

func (s *Session) calculateTotalPhases() int {
	if s.config.TotalCycles == 0 {
		return 0
	}

	cycles := s.config.TotalCycles
	phases := cycles

	if s.config.LongBreakEvery > 0 {
		longBreaks := (cycles - 1) / s.config.LongBreakEvery
		shortBreaks := cycles - 1 - longBreaks
		phases += longBreaks + shortBreaks
	} else {
		phases += cycles - 1
	}

	return phases
}

// CurrentPhase returns the current phase.
func (s *Session) CurrentPhase() Phase { return s.currentPhase }

// CyclesComplete returns completed work cycles.
func (s *Session) CyclesComplete() int { return s.cyclesComplete }

// TotalCycles returns total cycles configured.
func (s *Session) TotalCycles() int { return s.config.TotalCycles }

// TotalPhases returns total phases in session.
func (s *Session) TotalPhases() int { return s.totalPhases }

// PhasesComplete returns completed phases.
func (s *Session) PhasesComplete() int { return s.phasesComplete }

// PhaseDuration returns the duration of the current phase.
func (s *Session) PhaseDuration() time.Duration {
	switch s.currentPhase {
	case PhaseWork:
		return s.config.WorkDuration
	case PhaseShortBreak:
		return s.config.ShortBreakDuration
	case PhaseLongBreak:
		return s.config.LongBreakDuration
	default:
		return 0
	}
}

// NextPhase transitions to the next phase and returns it.
func (s *Session) NextPhase() Phase {
	if s.currentPhase == PhaseDone {
		return PhaseDone
	}

	s.phasesComplete++

	switch s.currentPhase {
	case PhaseWork:
		s.cyclesComplete++

		if s.config.TotalCycles > 0 && s.cyclesComplete >= s.config.TotalCycles {
			s.currentPhase = PhaseDone
			return s.currentPhase
		}

		if s.config.LongBreakEvery > 0 && s.cyclesComplete%s.config.LongBreakEvery == 0 {
			s.currentPhase = PhaseLongBreak
		} else {
			s.currentPhase = PhaseShortBreak
		}

	case PhaseShortBreak, PhaseLongBreak:
		s.currentPhase = PhaseWork
	}

	return s.currentPhase
}
