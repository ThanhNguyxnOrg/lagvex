//go:build !windows

package client

import (
	"os"
	"os/exec"
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
		out, err := exec.Command("ps", "-A", "-c", "-o", "pid=,comm=").Output()
		if err != nil {
			return false, "", 0
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			comm := strings.ToLower(fields[1])
			for _, target := range targets {
				if comm == target || comm == strings.TrimSuffix(target, ".exe") {
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
			if procName == target || strings.TrimSuffix(target, ".exe") == procName {
				return true, target, pid
			}
		}
	}

	return false, "", 0
}
