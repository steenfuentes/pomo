package engine

import (
	"context"
	"time"
)

// DefaultTickInterval is the frequency at which timer events are emitted.
const DefaultTickInterval = 200 * time.Millisecond

// TimerEvent is sent on each tick or phase completion.
type TimerEvent struct {
	Phase         Phase
	Elapsed       time.Duration
	Remaining     time.Duration
	Total         time.Duration
	Fraction      float64
	PhaseComplete bool
	CycleNum      int
	TotalCycles   int
	PhaseNum      int
	TotalPhases   int
}

// Timer runs a pomodoro session, emitting events on each tick.
type Timer struct {
	clock        Clock
	tickInterval time.Duration
	session      *Session
}

// NewTimer creates a timer with the real system clock.
func NewTimer(cfg Config) *Timer {
	return NewTimerWithClock(cfg, RealClock{}, DefaultTickInterval)
}

// NewTimerWithClock creates a timer with a custom clock for testing.
func NewTimerWithClock(cfg Config, clock Clock, tickInterval time.Duration) *Timer {
	return &Timer{
		clock:        clock,
		tickInterval: tickInterval,
		session:      NewSession(cfg),
	}
}

// Session returns the underlying session.
func (t *Timer) Session() *Session { return t.session }

// Run executes the full session, sending events to the provided channel.
// It blocks until session completes or context is cancelled.
func (t *Timer) Run(ctx context.Context, events chan<- TimerEvent) error {
	defer close(events)

	for t.session.CurrentPhase() != PhaseDone {
		if err := t.runPhase(ctx, events); err != nil {
			return err
		}
		t.session.NextPhase()
	}

	return nil
}

func (t *Timer) runPhase(ctx context.Context, events chan<- TimerEvent) error {
	duration := t.session.PhaseDuration()
	if duration == 0 {
		return nil
	}

	start := t.clock.Now()
	ticker := t.clock.NewTicker(t.tickInterval)
	defer ticker.Stop()

	for {
		elapsed := t.clock.Now().Sub(start)
		remaining := duration - elapsed
		if remaining < 0 {
			remaining = 0
		}

		event := TimerEvent{
			Phase:         t.session.CurrentPhase(),
			Elapsed:       elapsed,
			Remaining:     remaining,
			Total:         duration,
			Fraction:      float64(elapsed) / float64(duration),
			PhaseComplete: elapsed >= duration,
			CycleNum:      t.session.CyclesComplete() + 1,
			TotalCycles:   t.session.TotalCycles(),
			PhaseNum:      t.session.PhasesComplete() + 1,
			TotalPhases:   t.session.TotalPhases(),
		}

		if event.Fraction > 1.0 {
			event.Fraction = 1.0
		}

		select {
		case events <- event:
		case <-ctx.Done():
			return ctx.Err()
		}

		if event.PhaseComplete {
			return nil
		}

		select {
		case <-ticker.C():
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
