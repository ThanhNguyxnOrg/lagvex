//go:build windows

package client

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// findRunningProcess scans active Windows processes in-memory using Toolhelp32 snapshots.
// This is anti-cheat safe, sub-millisecond, and avoids spawning tasklist.exe subprocesses.
func findRunningProcess(targets []string) (bool, string, int) {
	if len(targets) == 0 {
		return false, "", 0
	}

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false, "", 0
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		procName := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		for _, target := range targets {
			if procName == target {
				return true, target, int(entry.ProcessID)
			}
		}
		err = windows.Process32Next(snapshot, &entry)
	}

	return false, "", 0
}
