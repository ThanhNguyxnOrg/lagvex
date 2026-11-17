//go:build windows

package client

import (
	"context"
	"os/exec"
	"syscall"
)

// createNoWindow is the Windows CREATE_NO_WINDOW process creation flag. Every
// helper the client runs (ip, netsh, route, tasklist, powershell) would
// otherwise pop a visible console window on Windows — the "terminal flashes
// when I click" report.
const createNoWindow = 0x08000000

// execHidden runs a child process without a visible console window.
func execHidden(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}

// execHiddenContext is execHidden with a context.
func execHiddenContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}
