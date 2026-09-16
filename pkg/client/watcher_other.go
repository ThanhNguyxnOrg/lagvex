//go:build !windows

package client

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// findRunningProcess checks running processes via /proc on Linux or basic scan.
func findRunningProcess(targets []string) (bool, string, int) {
	if len(targets) == 0 {
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
			if procName == target || strings.TrimSuffix(target, ".exe") == procName {
				return true, target, pid
			}
		}
	}

	return false, "", 0
}
