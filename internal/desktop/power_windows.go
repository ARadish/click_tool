//go:build windows

package desktop

var setThreadExecutionStateProc = kernel32.NewProc("SetThreadExecutionState")

func NewSystemPowerManager() *PowerManager {
	return newPowerManager(func(flags uint32) (uint32, error) {
		result, _, callErr := setThreadExecutionStateProc.Call(uintptr(flags))
		return uint32(result), callErr
	})
}
