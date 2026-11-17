//go:build darwin

package client

import (
	"errors"
	"os/exec"
	"strconv"
)

// connectionScanNote explains what this platform's socket tables expose.
const connectionScanNote = "Peers are read with lsof for sockets owned by this user; listening sockets are skipped."

// scanProcessConnections enumerates the connected sockets held by pid with
// lsof, which ships with macOS. It never touches the process; lsof only reads
// the kernel socket tables for the given pid.
func scanProcessConnections(pid int) ([]ProcessConnection, bool) {
	for _, candidate := range []string{"/usr/sbin/lsof", "lsof"} {
		cmd := execHidden(candidate, "-nP", "-i", "-a", "-p", strconv.Itoa(pid))
		out, err := cmd.Output()
		if err == nil {
			return parseLsofConnections(string(out)), true
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// lsof exits 1 when the process has no matching sockets.
			return nil, true
		}
		if errors.Is(err, exec.ErrNotFound) {
			continue
		}
		// lsof exists but failed (for example, a permission problem).
		return nil, false
	}
	return nil, false
}
