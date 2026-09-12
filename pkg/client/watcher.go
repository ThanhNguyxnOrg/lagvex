package client

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ProcessWatcher monitors running system processes to detect game start and exit events.
type ProcessWatcher struct {
	mu           sync.Mutex
	targets      []string // Executable names to watch (lowercase)
	isRunning    bool
	activePID    int
	matchedProc  string

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

	running, foundProc := isAnyProcessRunning(targets)

	pw.mu.Lock()
	if running && !wasRunning {
		pw.isRunning = true
		pw.matchedProc = foundProc
		pw.mu.Unlock()
		if pw.OnGameStarted != nil {
			pw.OnGameStarted(foundProc)
		}
	} else if !running && wasRunning {
		pw.isRunning = false
		pw.matchedProc = ""
		pw.mu.Unlock()
		if pw.OnGameStopped != nil {
			pw.OnGameStopped(lastMatched)
		}
	} else {
		pw.mu.Unlock()
	}
}

// isAnyProcessRunning checks running Windows tasks using tasklist.
func isAnyProcessRunning(targets []string) (bool, string) {
	// Query tasklist with CSV format
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			procName := strings.ToLower(strings.Trim(parts[0], "\" \r\t"))
			for _, target := range targets {
				if procName == target {
					return true, target
				}
			}
		}
	}

	return false, ""
}
