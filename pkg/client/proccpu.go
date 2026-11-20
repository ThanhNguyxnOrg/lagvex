package client

import (
	"sync"
	"time"
)

// processCPUSampler derives a per-process CPU percentage from two samples of
// the OS's own counters. It never injects into the target process and never
// estimates: a percentage is returned only after a second sample for the same
// PID, and only on platforms that expose those counters (Linux /proc, Windows
// GetProcessTimes). Elsewhere Sample reports unsupported so callers can say
// "not measured" instead of inventing a number.
type processCPUSampler struct {
	mu   sync.Mutex
	last map[int]cpuSample
}

type cpuSample struct {
	ticks float64 // process CPU time in seconds
	at    time.Time
}

// NewProcessCPUSampler returns a sampler with no prior observations.
func NewProcessCPUSampler() *processCPUSampler {
	return &processCPUSampler{last: make(map[int]cpuSample)}
}

// Sample returns the process CPU usage since the previous call for the same
// PID, as a percentage of one core (100% = one fully busy core). ok is false
// when the platform cannot measure it or when this is the first sample.
func (s *processCPUSampler) Sample(pid int) (pct float64, ok bool) {
	if pid <= 0 {
		return 0, false
	}
	now := time.Now()
	ticks, supported := readProcessCPUTicks(pid)
	if !supported {
		s.mu.Lock()
		delete(s.last, pid)
		s.mu.Unlock()
		return 0, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	prev, seen := s.last[pid]
	s.last[pid] = cpuSample{ticks: ticks, at: now}
	if !seen {
		return 0, false
	}
	wall := now.Sub(prev.at).Seconds()
	if wall <= 0.1 {
		// Too short a window for a meaningful ratio.
		return 0, false
	}
	used := ticks - prev.ticks
	if used < 0 {
		// A recycled PID can show decreasing totals; drop the stale sample.
		return 0, false
	}
	pct = used / wall * 100.0
	if pct < 0 {
		pct = 0
	}
	return pct, true
}

// Prune drops every PID not in the live set so exited or restarted processes
// are sampled from scratch and the map cannot grow without bound.
func (s *processCPUSampler) Prune(live map[int]bool) {
	s.mu.Lock()
	for pid := range s.last {
		if !live[pid] {
			delete(s.last, pid)
		}
	}
	s.mu.Unlock()
}

// Supported reports whether this OS exposes per-process CPU counters.
func (s *processCPUSampler) Supported() bool {
	return processCPUSupported()
}
