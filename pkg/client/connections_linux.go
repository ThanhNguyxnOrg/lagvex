//go:build linux

package client

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// connectionScanNote explains what this platform's socket tables expose.
const connectionScanNote = "Peers are read from the /proc socket tables for connected UDP and TCP sockets."

// scanProcessConnections enumerates the remote sockets held by pid using
// /proc/<pid>/fd and /proc/net/{tcp,udp}{,6}. It never touches the process.
func scanProcessConnections(pid int) ([]ProcessConnection, bool) {
	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		// The process may have exited between detection and this scan.
		return nil, true
	}

	inodes := make(map[string]struct{})
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
		if err != nil || !strings.HasPrefix(target, "socket:[") {
			continue
		}
		inode := strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")
		if inode != "" {
			inodes[inode] = struct{}{}
		}
	}
	if len(inodes) == 0 {
		return nil, true
	}

	tables := []struct {
		path     string
		protocol string
		ipv6     bool
	}{
		{"/proc/net/tcp", "tcp", false},
		{"/proc/net/tcp6", "tcp", true},
		{"/proc/net/udp", "udp", false},
		{"/proc/net/udp6", "udp", true},
	}

	var conns []ProcessConnection
	for _, table := range tables {
		file, err := os.Open(table.path)
		if err != nil {
			continue
		}
		rows := parseProcNetTable(file, table.protocol, table.ipv6)
		_ = file.Close()
		for inode := range inodes {
			if conn, ok := rows[inode]; ok {
				conns = append(conns, conn)
			}
		}
	}
	return conns, true
}
