package client

import (
	"context"
	"strings"
	"sync"
	"time"
)

// ProcessWatcher monitors running system processes to detect game start and exit events.
type ProcessWatcher struct {
	mu          sync.Mutex
	targets     []string // Executable names to watch (lowercase)
	isRunning   bool
	activePID   int
	matchedProc string

	OnGameStarted func(procName string)
	OnGameStopped func(procName string)
}

// NewProcessWatcher creates a new watcher for the specified list of executable names.
func NewProcessWatcher(targets []string) *ProcessWatcher {
	lowered := make([]string, len(targets))
	for i, t := range targets {
		lowered[i] = strings.ToLower(strings.TrimSpace(t))
	}
	return &ProcessWatcher{
		targets: lowered,
	}
}

// UpdateTargets updates the watched executable names.
func (pw *ProcessWatcher) UpdateTargets(targets []string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	lowered := make([]string, len(targets))
	for i, t := range targets {
		lowered[i] = strings.ToLower(strings.TrimSpace(t))
	}
	pw.targets = lowered
}

// ActivePID returns the currently tracked game process ID, or 0 if not running.
func (pw *ProcessWatcher) ActivePID() int {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.activePID
}

// Start begins periodic process scanning.
func (pw *ProcessWatcher) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pw.scan()
		}
	}
}

func (pw *ProcessWatcher) scan() {
	pw.mu.Lock()
	targets := make([]string, len(pw.targets))
	copy(targets, pw.targets)
	wasRunning := pw.isRunning
	lastMatched := pw.matchedProc
	pw.mu.Unlock()

	if len(targets) == 0 {
		return
	}

	running, foundProc, pid := findRunningProcess(targets)

	pw.mu.Lock()
	if running && !wasRunning {
		pw.isRunning = true
		pw.matchedProc = foundProc
		pw.activePID = pid
		pw.mu.Unlock()
		if pw.OnGameStarted != nil {
			pw.OnGameStarted(foundProc)
		}
	} else if !running && wasRunning {
		pw.isRunning = false
		pw.matchedProc = ""
		pw.activePID = 0
		pw.mu.Unlock()
		if pw.OnGameStopped != nil {
			pw.OnGameStopped(lastMatched)
		}
	} else {
		pw.mu.Unlock()
	}
}
