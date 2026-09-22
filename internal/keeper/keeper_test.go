package keeper

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type manualTicker struct {
	ch      chan time.Time
	mu      sync.Mutex
	stopped bool
}

func (t *manualTicker) C() <-chan time.Time { return t.ch }
func (t *manualTicker) Stop() {
	t.mu.Lock()
	t.stopped = true
	t.mu.Unlock()
}
func (t *manualTicker) isStopped() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.stopped
}

type tickerFactory struct {
	mu        sync.Mutex
	durations []time.Duration
	tickers   []*manualTicker
}

func (f *tickerFactory) new(interval time.Duration) Ticker {
	f.mu.Lock()
	defer f.mu.Unlock()
	ticker := &manualTicker{ch: make(chan time.Time, 4)}
	f.durations = append(f.durations, interval)
	f.tickers = append(f.tickers, ticker)
	return ticker
}

type fakeMover struct {
	mu        sync.Mutex
	x         int
	minX      int
	maxX      int
	moves     []int
	moveError error
}

type fakeWakeLock struct {
	acquireCount int
	releaseCount int
	acquireError error
}

func (w *fakeWakeLock) Acquire() error {
	w.acquireCount++
	return w.acquireError
}

func (w *fakeWakeLock) Release() error {
	w.releaseCount++
	return nil
}

func (m *fakeMover) Position() (Point, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Point{X: m.x, Y: 5}, nil
}

func (m *fakeMover) MoveRelative(dx, _ int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moves = append(m.moves, dx)
	if m.moveError != nil {
		return m.moveError
	}
	m.x += dx
	if m.x > m.maxX {
		m.x = m.maxX
	}
	if m.x < m.minX {
		m.x = m.minX
	}
	return nil
}

func (m *fakeMover) moveSnapshot() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]int(nil), m.moves...)
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

func newTestKeeper(mover MouseMover, factory *tickerFactory) *Keeper {
	return New(mover, &fakeWakeLock{}, 30*time.Second, factory.new, nil)
}

func TestStartAcquiresWakeLockAndStopReleasesIt(t *testing.T) {
	factory := &tickerFactory{}
	wake := &fakeWakeLock{}
	k := New(&fakeMover{maxX: 10}, wake, 30*time.Second, factory.new, nil)
	defer k.Close()

	k.Start()
	k.Stop()

	if wake.acquireCount != 1 {
		t.Fatalf("acquire count = %d, want 1", wake.acquireCount)
	}
	if wake.releaseCount != 1 {
		t.Fatalf("release count = %d, want 1", wake.releaseCount)
	}
}

func TestWakeLockFailureLeavesKeeperStopped(t *testing.T) {
	factory := &tickerFactory{}
	wake := &fakeWakeLock{acquireError: errors.New("power request failed")}
	k := New(&fakeMover{maxX: 10}, wake, 30*time.Second, factory.new, nil)
	defer k.Close()

	k.Start()

	state := k.State()
	if state.Running {
		t.Fatal("Running = true after wake lock failure")
	}
	if state.LastError == "" {
		t.Fatal("LastError is empty after wake lock failure")
	}
}

func TestStartIsIdempotent(t *testing.T) {
	factory := &tickerFactory{}
	k := newTestKeeper(&fakeMover{maxX: 10}, factory)
	defer k.Close()

	k.Start()
	k.Start()

	if len(factory.tickers) != 1 {
		t.Fatalf("ticker count = %d, want 1", len(factory.tickers))
	}
	if !k.State().Running {
		t.Fatal("Running = false after Start")
	}
}

func TestStartMovesFiftyPixelsImmediatelyThenAlternatesOnTicks(t *testing.T) {
	mover := &fakeMover{x: 100, maxX: 1000}
	factory := &tickerFactory{}
	k := newTestKeeper(mover, factory)
	defer k.Close()
	k.Start()

	factory.tickers[0].ch <- time.Now()
	waitFor(t, func() bool { return len(mover.moveSnapshot()) == 2 })

	moves := mover.moveSnapshot()
	if moves[0] != 50 || moves[1] != -50 {
		t.Fatalf("moves = %v, want [50 -50]", moves)
	}
}

func TestStartReversesAtScreenBoundary(t *testing.T) {
	mover := &fakeMover{x: 100, maxX: 100}
	factory := &tickerFactory{}
	k := newTestKeeper(mover, factory)
	defer k.Close()
	k.Start()

	waitFor(t, func() bool { return len(mover.moveSnapshot()) == 2 })

	moves := mover.moveSnapshot()
	if moves[0] != 50 || moves[1] != -50 {
		t.Fatalf("boundary moves = %v, want [50 -50]", moves)
	}
}

func TestImmobileCursorStopsKeeperAndReportsError(t *testing.T) {
	mover := &fakeMover{minX: 0, maxX: 0}
	factory := &tickerFactory{}
	k := newTestKeeper(mover, factory)
	defer k.Close()

	k.Start()

	state := k.State()
	if state.Running {
		t.Fatal("Running = true when cursor did not move in either direction")
	}
	if state.LastError == "" {
		t.Fatal("LastError is empty when cursor did not move")
	}
}

func TestSetIntervalReplacesRunningTicker(t *testing.T) {
	factory := &tickerFactory{}
	k := newTestKeeper(&fakeMover{maxX: 10}, factory)
	defer k.Close()
	k.Start()

	if err := k.SetInterval(15 * time.Second); err != nil {
		t.Fatalf("SetInterval returned error: %v", err)
	}

	if len(factory.tickers) != 2 || factory.durations[1] != 15*time.Second {
		t.Fatalf("ticker durations = %v, want [30s 15s]", factory.durations)
	}
	if !factory.tickers[0].isStopped() {
		t.Fatal("old ticker was not stopped")
	}
}

func TestMoveFailureStopsKeeperAndReportsError(t *testing.T) {
	mover := &fakeMover{x: 5, maxX: 10, moveError: errors.New("move failed")}
	factory := &tickerFactory{}
	states := make(chan State, 4)
	wake := &fakeWakeLock{}
	k := New(mover, wake, 30*time.Second, factory.new, func(state State) { states <- state })
	defer k.Close()
	k.Start()

	state := k.State()
	if state.LastError == "" {
		t.Fatal("LastError is empty after movement failure")
	}
	if len(factory.tickers) != 0 {
		t.Fatal("ticker was created after immediate movement failure")
	}
	if wake.releaseCount != 1 {
		t.Fatalf("release count = %d after movement failure, want 1", wake.releaseCount)
	}
}

func TestStopIsIdempotent(t *testing.T) {
	factory := &tickerFactory{}
	k := newTestKeeper(&fakeMover{maxX: 10}, factory)
	defer k.Close()
	k.Start()
	k.Stop()
	k.Stop()

	if k.State().Running {
		t.Fatal("Running = true after Stop")
	}
	if !factory.tickers[0].isStopped() {
		t.Fatal("ticker was not stopped")
	}
}
