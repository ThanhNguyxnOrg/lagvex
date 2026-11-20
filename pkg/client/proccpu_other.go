//go:build !linux && !windows

package client

// processCPUSupported: nil. macOS and other platforms have no portable,
// injection-free per-process CPU counter in pure Go (proc_pid_rusage needs cgo),
// so callers must report "not measured" rather than estimate.
func processCPUSupported() bool { return false }

// readProcessCPUTicks always reports unsupported on this platform.
func readProcessCPUTicks(pid int) (float64, bool) { return 0, false }
