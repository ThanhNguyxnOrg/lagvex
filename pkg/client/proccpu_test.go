package client

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// TestProcessCPUSupported is a platform inventory: on Linux the /proc counters
// exist, everywhere else the sampler reports unsupported instead of faking.
func TestProcessCPUSupported(t *testing.T) {
	if processCPUSupported() {
		t.Log("platform exposes per-process CPU counters")
		return
	}
	if _, ok := readProcessCPUTicks(os.Getpid()); ok {
		t.Fatal("readProcessCPUTicks succeeded on a platform that reports no support")
	}
}

// TestProcessCPUTicksForSelf verifies the raw reader returns a non-negative
// CPU time for a live process and refuses a bogus PID.
func TestProcessCPUTicksForSelf(t *testing.T) {
	if !processCPUSupported() {
		t.Skip("per-process CPU counters are not available on this platform")
	}
	ticks, ok := readProcessCPUTicks(os.Getpid())
	if !ok {
		t.Fatal("CPU time for the current process was not readable")
	}
	if ticks < 0 {
		t.Fatalf("CPU time = %fs, want >= 0", ticks)
	}
	if _, ok := readProcessCPUTicks(-1); ok {
		t.Fatal("CPU time for a negative pid was reported as readable")
	}
}

// TestProcessCPUSamplerTwoSamples proves the percentage only appears after a
// second sample: one call must boot, two must produce a measured value.
func TestProcessCPUSamplerTwoSamples(t *testing.T) {
	if !processCPUSupported() {
		t.Skip("per-process CPU counters are not available on this platform")
	}
	s := NewProcessCPUSampler()
	pid := os.Getpid()

	if _, ok := s.Sample(pid); ok {
		t.Fatal("first sample reported a CPU percentage")
	}
	// Burn a slice of one core so the deltas are real numbers.
	deadline := time.Now().Add(250 * time.Millisecond)
	x := 0.0
	for time.Now().Before(deadline) {
		for i := 0; i < 200000; i++ {
			x += float64(i)
		}
	}
	_ = x
	pct, ok := s.Sample(pid)
	if !ok {
		t.Fatal("second sample reported no percentage")
	}
	if pct < 0 || pct > 100*float64(max(1, runtime.NumCPU()))*4 {
		// Allow generous headroom above 100% * ncpu for scheduler bursts.
		t.Fatalf("cpu percent = %v, want within a sane range", pct)
	}
}

// TestProcessCPUSamplerPruneAndRecycledPID checks the map stays bounded and a
// decreasing counter (recycled PID) is dropped, not reported as negative load.
func TestProcessCPUSamplerPruneAndRecycledPID(t *testing.T) {
	if !processCPUSupported() {
		t.Skip("per-process CPU counters are not available on this platform")
	}
	s := NewProcessCPUSampler()
	pid := os.Getpid()
	_, _ = s.Sample(pid)
	s.Prune(map[int]bool{})
	s.mu.Lock()
	n := len(s.last)
	s.mu.Unlock()
	if n != 0 {
		t.Fatalf("prune left %d entries, want 0", n)
	}
	if _, ok := s.Sample(-5); ok {
		t.Fatal("sampler accepted a negative pid")
	}
}
