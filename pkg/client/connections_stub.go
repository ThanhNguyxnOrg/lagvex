//go:build !linux && !darwin && !windows

package client

// connectionScanNote explains why this platform reports no peers.
const connectionScanNote = "Socket enumeration is not implemented on this platform."

// scanProcessConnections reports that this platform has no socket-table
// reader yet. The caller surfaces "unsupported" instead of guessing.
func scanProcessConnections(pid int) ([]ProcessConnection, bool) {
	return nil, false
}
