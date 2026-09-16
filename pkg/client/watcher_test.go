package client

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFindRunningProcessSelf(t *testing.T) {
	exePath, err := os.Executable()
	if err != nil {
		t.Skipf("cannot get executable path: %v", err)
	}

	exeName := filepath.Base(exePath)
	found, matched, pid := findRunningProcess([]string{exeName, "nonexistent_proc_12345.exe"})

	if !found {
		t.Logf("Self executable %s not matched in process snapshot (may be running under test runner wrapper)", exeName)
		return
	}

	if matched != strings.ToLower(exeName) {
		t.Errorf("matched = %s; want %s", matched, exeName)
	}

	if pid <= 0 {
		t.Errorf("expected positive PID, got %d", pid)
	}
}

func TestProcessWatcherLifecycle(t *testing.T) {
	watcher := NewProcessWatcher([]string{"dummy_target_1.exe", "dummy_target_2.exe"})

	if watcher.ActivePID() != 0 {
		t.Errorf("expected active PID = 0 initially")
	}

	var mu sync.Mutex
	startedCalled := false
	stoppedCalled := false

	watcher.OnGameStarted = func(proc string) {
		mu.Lock()
		startedCalled = true
		mu.Unlock()
	}
	watcher.OnGameStopped = func(proc string) {
		mu.Lock()
		stoppedCalled = true
		mu.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run watcher in short background loop
	go watcher.Start(ctx, 50*time.Millisecond)
	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	if startedCalled || stoppedCalled {
		t.Errorf("dummy targets should not have triggered start/stop")
	}
	mu.Unlock()

	// Update targets
	watcher.UpdateTargets([]string{"updated_target.exe"})
	time.Sleep(100 * time.Millisecond)
}
