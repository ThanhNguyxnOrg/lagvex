//go:build windows

package main

import (
	"log"
	"os"
	"syscall"
)

// attachParentConsole attaches the process to its parent console (PowerShell / CMD) if one exists.
// When compiled with -ldflags="-H=windowsgui", Windows does not create a console window on double-click.
// If run from a command prompt, this attaches stdout/stderr back to that prompt.
func attachParentConsole() bool {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole := modkernel32.NewProc("AttachConsole")
	const attachParentProcess = ^uintptr(0) // (DWORD)-1

	r1, _, _ := procAttachConsole.Call(attachParentProcess)
	if r1 != 0 {
		// Successfully attached to parent console
		if stdout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
			os.Stdout = stdout
		}
		if stderr, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
			os.Stderr = stderr
			log.SetOutput(stderr)
		}
		return true
	}
	return false
}
