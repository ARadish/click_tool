package keeper

import (
	"fmt"
	"sync"
	"time"
)

type Point struct {
	X int
	Y int
}

type MouseMover interface {
	Position() (Point, error)
	MoveRelative(dx, dy int) error
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type TickerFactory func(interval time.Duration) Ticker

type State struct {
	Running   bool
	Interval  time.Duration
	LastError string
}

type commandKind uint8

const (
	commandStart commandKind = iota
	commandStop
	commandToggle
	commandSetInterval
	commandState
	commandClose
)

type command struct {
	kind     commandKind
	interval time.Duration
	state    chan State
	err      chan error
}

type Keeper struct {
	commands chan command
	done     chan struct{}
	close    sync.Once
}

type timeTicker struct{ *time.Ticker }

func (t timeTicker) C() <-chan time.Time { return t.Ticker.C }

func New(mover MouseMover, interval time.Duration, factory TickerFactory, onChange func(State)) *Keeper {
	if factory == nil {
		factory = func(interval time.Duration) Ticker { return timeTicker{time.NewTicker(interval)} }
	}
	k := &Keeper{commands: make(chan command), done: make(chan struct{})}
	go k.run(mover, interval, factory, onChange)
	return k
}

func (k *Keeper) Start()                            { k.change(commandStart) }
func (k *Keeper) Stop()                             { k.change(commandStop) }
func (k *Keeper) Toggle()                           { k.change(commandToggle) }
func (k *Keeper) SetInterval(d time.Duration) error { return k.changeWithInterval(d) }

func (k *Keeper) State() State {
	reply := make(chan State, 1)
	k.commands <- command{kind: commandState, state: reply}
	return <-reply
}

func (k *Keeper) Close() {
	k.close.Do(func() {
		reply := make(chan error, 1)
		k.commands <- command{kind: commandClose, err: reply}
		<-reply
		<-k.done
	})
}

func (k *Keeper) change(kind commandKind) {
	reply := make(chan error, 1)
	k.commands <- command{kind: kind, err: reply}
	<-reply
}

func (k *Keeper) changeWithInterval(interval time.Duration) error {
	reply := make(chan error, 1)
	k.commands <- command{kind: commandSetInterval, interval: interval, err: reply}
	return <-reply
}

func (k *Keeper) run(mover MouseMover, interval time.Duration, factory TickerFactory, onChange func(State)) {
	defer close(k.done)
	state := State{Interval: interval}
	direction := 1
	var ticker Ticker
	var ticks <-chan time.Time

	notify := func() {
		if onChange != nil {
			onChange(state)
		}
	}
	start := func() {
		if state.Running {
			return
		}
		ticker = factory(state.Interval)
		ticks = ticker.C()
		state.Running = true
		state.LastError = ""
		notify()
	}
	stop := func() {
		if !state.Running {
			return
		}
		ticker.Stop()
		ticker = nil
		ticks = nil
		state.Running = false
		notify()
	}

	for {
		select {
		case cmd := <-k.commands:
			switch cmd.kind {
			case commandStart:
				start()
				cmd.err <- nil
			case commandStop:
				stop()
				cmd.err <- nil
			case commandToggle:
				if state.Running {
					stop()
				} else {
					start()
				}
				cmd.err <- nil
			case commandSetInterval:
				if cmd.interval <= 0 {
					cmd.err <- fmt.Errorf("interval must be positive")
					continue
				}
				if state.Interval != cmd.interval {
					wasRunning := state.Running
					stop()
					state.Interval = cmd.interval
					if wasRunning {
						start()
					} else {
						notify()
					}
				}
				cmd.err <- nil
			case commandState:
				cmd.state <- state
			case commandClose:
				stop()
				cmd.err <- nil
				return
			}
		case <-ticks:
			actualDirection, err := moveOnce(mover, direction)
			if err != nil {
				if ticker != nil {
					ticker.Stop()
				}
				ticker = nil
				ticks = nil
				state.Running = false
				state.LastError = err.Error()
				notify()
				continue
			}
			direction = -actualDirection
		}
	}
}

func moveOnce(mover MouseMover, direction int) (int, error) {
	before, err := mover.Position()
	if err != nil {
		return direction, fmt.Errorf("read mouse position: %w", err)
	}
	if err := mover.MoveRelative(direction, 0); err != nil {
		return direction, fmt.Errorf("move mouse: %w", err)
	}
	after, err := mover.Position()
	if err != nil {
		return direction, fmt.Errorf("verify mouse position: %w", err)
	}
	if after == before {
		direction = -direction
		if err := mover.MoveRelative(direction, 0); err != nil {
			return direction, fmt.Errorf("move mouse from boundary: %w", err)
		}
	}
	return direction, nil
}
