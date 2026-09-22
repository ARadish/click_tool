package desktop

import "fmt"

const (
	executionStateSystemRequired  uint32 = 0x00000001
	executionStateDisplayRequired uint32 = 0x00000002
	executionStateContinuous      uint32 = 0x80000000
)

type executionStateSetter func(flags uint32) (uint32, error)

type PowerManager struct {
	setExecutionState executionStateSetter
}

func newPowerManager(setter executionStateSetter) *PowerManager {
	return &PowerManager{setExecutionState: setter}
}

func (p *PowerManager) Acquire() error {
	return p.set(executionStateContinuous | executionStateDisplayRequired | executionStateSystemRequired)
}

func (p *PowerManager) Release() error {
	return p.set(executionStateContinuous)
}

func (p *PowerManager) set(flags uint32) error {
	previous, err := p.setExecutionState(flags)
	if previous == 0 {
		return fmt.Errorf("SetThreadExecutionState(0x%08x): %w", flags, err)
	}
	return nil
}
