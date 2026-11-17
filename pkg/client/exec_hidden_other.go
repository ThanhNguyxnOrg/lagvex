//go:build !windows

package client

import (
	"context"
	"os/exec"
)

// execHidden mirrors the Windows helper: on Linux and macOS child processes
// have no console window to hide, so this is a plain exec.Command.
func execHidden(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// execHiddenContext is execHidden with a context.
func execHiddenContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
