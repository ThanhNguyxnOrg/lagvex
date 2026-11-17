//go:build !windows

package client

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// findRunningProcess checks running processes via ps on macOS or /proc on Linux.
func findRunningProcess(targets []string) (bool, string, int) {
	if len(targets) == 0 {
		return false, "", 0
	}

	if runtime.GOOS == "darwin" {
		out, err := execHidden("ps", "-A", "-c", "-o", "pid=,comm=").Output()
		if err != nil {
			return false, "", 0
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			comm := strings.ToLower(strings.Join(fields[1:], " "))
			for _, target := range targets {
				if strings.EqualFold(comm, target) || strings.EqualFold(comm, strings.TrimSuffix(target, ".exe")) {
					return true, target, pid
				}
			}
		}
		return false, "", 0
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, "", 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commBytes, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		procName := strings.ToLower(strings.TrimSpace(string(commBytes)))
		for _, target := range targets {
			if strings.EqualFold(procName, target) || strings.EqualFold(procName, strings.TrimSuffix(target, ".exe")) {
				return true, target, pid
			}
		}
	}

	return false, "", 0
}

// findRunningProcesses returns every watched target that is currently running,
// mapped to the first PID seen for it. One process-table scan serves all
// targets, unlike findRunningProcess which stops at the first match.
func findRunningProcesses(targets []string) map[string]int {
	matches := make(map[string]int)
	if len(targets) == 0 {
		return matches
	}

	match := func(procName string, pid int) {
		lowered := strings.ToLower(strings.TrimSpace(procName))
		for _, target := range targets {
			if _, seen := matches[target]; seen {
				continue
			}
			if strings.EqualFold(lowered, target) || strings.EqualFold(lowered, strings.TrimSuffix(target, ".exe")) {
				matches[target] = pid
			}
		}
	}

	if runtime.GOOS == "darwin" {
		out, err := execHidden("ps", "-A", "-c", "-o", "pid=,comm=").Output()
		if err != nil {
			return matches
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			match(strings.Join(fields[1:], " "), pid)
		}
		return matches
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return matches
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		commBytes, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		match(string(commBytes), pid)
	}

	return matches
}
