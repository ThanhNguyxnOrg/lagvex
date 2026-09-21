//go:build !windows

package client

import (
	"net"
	"syscall"
)

// ApplySocketQoS sets IP_TOS DSCP EF-46 on non-Windows platforms.
func ApplySocketQoS(conn *net.UDPConn) error {
	if conn == nil {
		return nil
	}
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	return rawConn.Control(func(fd uintptr) {
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TOS, 0xB8)
	})
}
