//go:build linux

package client

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	once   sync.Once
	hzVal  float64
	hzRead bool
)

// clockTicksPerSecond measures the kernel's USER_HZ instead of assuming 100:
// /proc/self/stat's starttime is in ticks since boot and /proc/uptime is in
// seconds since boot, so their quotient is the tick frequency the kernel uses
// for the very stat fields this sampler reads.
func clockTicksPerSecond() (float64, bool) {
	once.Do(func() {
		upBytes, err := os.ReadFile("/proc/uptime")
		if err != nil {
			return
		}
		upFields := strings.Fields(string(upBytes))
		if len(upFields) < 1 {
			return
		}
		uptime, err := strconv.ParseFloat(upFields[0], 64)
		if err != nil || uptime <= 0 {
			return
		}
		data, err := os.ReadFile(filepath.Join("/proc", "self", "stat"))
		if err != nil {
			return
		}
		s := string(data)
		idx := strings.LastIndexByte(s, ')')
		if idx < 0 || idx+2 > len(s) {
			return
		}
		fields := strings.Fields(s[idx+2:])
		// fields[0] is state (field 3); starttime is field 22 -> fields[19].
		if len(fields) < 22 {
			return
		}
		starttime, err := strconv.ParseFloat(fields[19], 64)
		if err != nil {
			return
		}
		if hz := starttime / uptime; hz > 0 {
			hzVal = hz
			hzRead = true
		}
	})
	return hzVal, hzRead
}

// processCPUSupported reports whether this Linux build can read /proc.
func processCPUSupported() bool {
	_, ok := clockTicksPerSecond()
	return ok
}

// readProcessCPUTicks returns the process CPU time in seconds (user+system).
// It reads only /proc/<pid>/stat, the same file the OS itself uses to schedule.
func readProcessCPUTicks(pid int) (float64, bool) {
	clk, ok := clockTicksPerSecond()
	if !ok || clk <= 0 {
		return 0, false
	}
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0, false
	}
	s := string(data)
	// Field 14 (utime) and 15 (stime), after the comm field which may itself
	// contain spaces and parentheses, so parse from the last ')'.
	idx := strings.LastIndexByte(s, ')')
	if idx < 0 || idx+2 > len(s) {
		return 0, false
	}
	fields := strings.Fields(s[idx+2:])
	if len(fields) < 13 {
		return 0, false
	}
	utime, err1 := strconv.ParseFloat(fields[11], 64)
	stime, err2 := strconv.ParseFloat(fields[12], 64)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	return (utime + stime) / clk, true
}
