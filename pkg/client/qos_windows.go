//go:build windows

package client

import (
	"log"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modQwave               = windows.NewLazySystemDLL("qwave.dll")
	procQOSCreateHandle    = modQwave.NewProc("QOSCreateHandle")
	procQOSAddSocketToFlow = modQwave.NewProc("QOSAddSocketToFlow")
)

type qosVersion struct {
	MajorVersion uint16
	MinorVersion uint16
}

// ApplySocketQoS tags the UDP socket with DSCP Expedited Forwarding (EF-46) priority.
func ApplySocketQoS(conn *net.UDPConn) error {
	if conn == nil {
		return nil
	}

	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	return rawConn.Control(func(fd uintptr) {
		// 1. IP_TOS DSCP EF-46 (0xB8 = 46 << 2)
		_ = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_TOS, 0xB8)

		// 2. Windows Native qWAVE API (Hardware queue priority for games)
		if err := modQwave.Load(); err == nil {
			var qosHandle windows.Handle
			ver := qosVersion{MajorVersion: 1, MinorVersion: 0}
			r1, _, _ := procQOSCreateHandle.Call(uintptr(unsafe.Pointer(&ver)), uintptr(unsafe.Pointer(&qosHandle)))
			if r1 != 0 && qosHandle != 0 {
				flowType := uint32(2) // QOSTrafficTypeAudioVideo (interactive game packet priority)
				var flowID uint32
				r2, _, _ := procQOSAddSocketToFlow.Call(uintptr(qosHandle), fd, 0, uintptr(flowType), 0, uintptr(unsafe.Pointer(&flowID)))
				if r2 != 0 {
					log.Printf("[QoS] Successfully tagged socket flow with Windows qWAVE DSCP EF-46 (FlowID: %d)", flowID)
				}
			}
		}
	})
}
