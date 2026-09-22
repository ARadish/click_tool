package desktop

import "testing"

func TestPowerManagerAcquiresAndReleasesExecutionState(t *testing.T) {
	var calls []uint32
	manager := newPowerManager(func(flags uint32) (uint32, error) {
		calls = append(calls, flags)
		return 1, nil
	})

	if err := manager.Acquire(); err != nil {
		t.Fatalf("Acquire returned error: %v", err)
	}
	if err := manager.Release(); err != nil {
		t.Fatalf("Release returned error: %v", err)
	}

	want := []uint32{0x80000003, 0x80000000}
	if len(calls) != len(want) || calls[0] != want[0] || calls[1] != want[1] {
		t.Fatalf("execution state calls = %#v, want %#v", calls, want)
	}
}
