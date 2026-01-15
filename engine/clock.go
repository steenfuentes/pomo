package engine

import "time"

// Clock abstracts time operations for testability.
type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) Ticker
	Sleep(d time.Duration)
}

// Ticker abstracts time.Ticker for testability.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// RealClock uses the system monotonic clock.
type RealClock struct{}

func (RealClock) Now() time.Time                   { return time.Now() }
func (RealClock) NewTicker(d time.Duration) Ticker { return &realTicker{time.NewTicker(d)} }
func (RealClock) Sleep(d time.Duration)            { time.Sleep(d) }

type realTicker struct{ *time.Ticker }

func (t *realTicker) C() <-chan time.Time { return t.Ticker.C }

// MockClock provides manual time control for testing.
type MockClock struct {
	current time.Time
	tickers []*MockTicker
}

// NewMockClock creates a MockClock starting at the given time.
func NewMockClock(start time.Time) *MockClock {
	return &MockClock{current: start}
}

func (m *MockClock) Now() time.Time { return m.current }

func (m *MockClock) NewTicker(d time.Duration) Ticker {
	t := &MockTicker{
		interval: d,
		ch:       make(chan time.Time, 1),
		nextTick: m.current.Add(d),
		stopped:  false,
	}
	m.tickers = append(m.tickers, t)
	return t
}

func (m *MockClock) Sleep(d time.Duration) {
	m.Advance(d)
}

// Advance moves time forward and fires any due tickers.
func (m *MockClock) Advance(d time.Duration) {
	target := m.current.Add(d)
	for m.current.Before(target) {
		var earliest *MockTicker
		for _, t := range m.tickers {
			if t.stopped {
				continue
			}
			if earliest == nil || t.nextTick.Before(earliest.nextTick) {
				earliest = t
			}
		}

		if earliest == nil || earliest.nextTick.After(target) {
			m.current = target
			break
		}

		m.current = earliest.nextTick
		select {
		case earliest.ch <- m.current:
		default:
		}
		earliest.nextTick = earliest.nextTick.Add(earliest.interval)
	}
}

// MockTicker is a controllable ticker for testing.
type MockTicker struct {
	interval time.Duration
	ch       chan time.Time
	nextTick time.Time
	stopped  bool
}

func (t *MockTicker) C() <-chan time.Time { return t.ch }
func (t *MockTicker) Stop()               { t.stopped = true }
