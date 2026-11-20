//go:build windows

package client

import (
	"golang.org/x/sys/windows"
)

// processCPUSupported reports whether this Windows build can query process times.
func processCPUSupported() bool { return true }

// readProcessCPUTicks returns the process CPU time in seconds (user+kernel)
// from GetProcessTimes — the same counters treble tools report, no injection.
func readProcessCPUTicks(pid int) (float64, bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, false
	}
	defer windows.CloseHandle(h)

	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &creation, &exit, &kernel, &user); err != nil {
		return 0, false
	}
	// FILETIME is in 100-nanosecond intervals.
	hundredNanos := uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime)
	hundredNanos += uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime)
	return float64(hundredNanos) / 1e7, true
}
