//go:build windows

package client

import (
	"encoding/binary"
	"net"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

// connectionScanNote explains what this platform's socket tables expose.
const connectionScanNote = "TCP peers come from the Windows IP helper tables; " +
	"Windows does not expose the remote peer of another process's UDP socket, so UDP rows show the local endpoint only."

var (
	iphlpapiDLL             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedUDPTable = iphlpapiDLL.NewProc("GetExtendedUdpTable")
	procGetExtendedTCPTable = iphlpapiDLL.NewProc("GetExtendedTcpTable")
)

const (
	afINET  = 2
	afINET6 = 23

	udpTableOwnerPID = 1
	tcpTableOwnerPID = 5
)

// Windows IP helper row layouts (the *OwnerPid variants).
type mibUDPRowOwnerPID struct {
	LocalAddr uint32
	LocalPort uint32
	OwningPID uint32
}

type mibUDP6RowOwnerPID struct {
	LocalAddr    [16]byte
	LocalScopeID uint32
	LocalPort    uint32
	OwningPID    uint32
}

type mibTCPRowOwnerPID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPID  uint32
}

type mibTCP6RowOwnerPID struct {
	LocalAddr     [16]byte
	LocalScopeID  uint32
	LocalPort     uint32
	RemoteAddr    [16]byte
	RemoteScopeID uint32
	RemotePort    uint32
	State         uint32
	OwningPID     uint32
}

var windowsTCPStates = map[uint32]string{
	1: "closed", 2: "listen", 3: "syn-sent", 4: "syn-received",
	5: "established", 6: "fin-wait-1", 7: "fin-wait-2", 8: "close-wait",
	9: "closing", 10: "last-ack", 11: "time-wait", 12: "delete-tcb",
}

// scanProcessConnections reads the TCP and UDP endpoint tables via the IP
// helper API and keeps the rows owned by pid. Windows does not expose the
// remote peer of another process's UDP socket, so those rows are reported
// with an empty remote endpoint.
func scanProcessConnections(pid int) ([]ProcessConnection, bool) {
	target := uint32(pid)
	var conns []ProcessConnection

	for _, family := range []struct {
		value uint32
		ipv6  bool
	}{{afINET, false}, {afINET6, true}} {
		conns = append(conns, scanUDPTable(target, family.value, family.ipv6)...)
		conns = append(conns, scanTCPTable(target, family.value, family.ipv6)...)
	}
	return conns, true
}

// queryMIBTable performs the two-call buffer dance used by the IP helper
// table functions.
func queryMIBTable(proc *windows.LazyProc, family, class uint32) ([]byte, error) {
	var size uint32
	ret, _, callErr := proc.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), uintptr(class), 0)
	if ret == 0 && size == 0 {
		return nil, nil
	}
	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) && ret != 0 {
		return nil, callErr
	}
	if size == 0 {
		return nil, nil
	}

	buf := make([]byte, size)
	ret, _, callErr = proc.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0, uintptr(family), uintptr(class), 0,
	)
	if ret != 0 {
		return nil, callErr
	}
	return buf, nil
}

// copyTableRows copies the counted rows that follow the leading DWORD into an
// aligned slice of the row type.
func copyTableRows[T any](buf []byte) []T {
	if len(buf) < 4 {
		return nil
	}
	declared := binary.LittleEndian.Uint32(buf[:4])
	rowSize := int(unsafe.Sizeof(*new(T)))
	if rowSize == 0 {
		return nil
	}
	rows := int(declared)
	if maxRows := (len(buf) - 4) / rowSize; rows > maxRows {
		rows = maxRows
	}
	if rows <= 0 {
		return nil
	}
	out := make([]T, rows)
	src := unsafe.Slice((*byte)(unsafe.Pointer(&out[0])), rows*rowSize)
	copy(src, buf[4:4+rows*rowSize])
	return out
}

// winPort decodes the network-byte-order port stored in a DWORD.
func winPort(port uint32) int {
	return int((port&0xFF)<<8 | (port>>8)&0xFF)
}

// winIPv4 decodes an IPv4 address stored in network byte order in a DWORD.
func winIPv4(addr uint32) net.IP {
	return net.IPv4(byte(addr), byte(addr>>8), byte(addr>>16), byte(addr>>24))
}

func scanUDPTable(pid, family uint32, ipv6 bool) []ProcessConnection {
	buf, err := queryMIBTable(procGetExtendedUDPTable, family, udpTableOwnerPID)
	if err != nil || len(buf) == 0 {
		return nil
	}

	var conns []ProcessConnection
	if ipv6 {
		for _, row := range copyTableRows[mibUDP6RowOwnerPID](buf) {
			if row.OwningPID != pid {
				continue
			}
			ip := net.IP(row.LocalAddr[:])
			if ip.IsUnspecified() {
				continue
			}
			conns = append(conns, ProcessConnection{
				Protocol: "udp",
				Family:   "ipv6",
				Local:    net.JoinHostPort(ip.String(), strconv.Itoa(winPort(row.LocalPort))),
			})
		}
		return conns
	}

	for _, row := range copyTableRows[mibUDPRowOwnerPID](buf) {
		if row.OwningPID != pid {
			continue
		}
		ip := winIPv4(row.LocalAddr)
		if ip.IsUnspecified() {
			continue
		}
		conns = append(conns, ProcessConnection{
			Protocol: "udp",
			Family:   "ipv4",
			Local:    net.JoinHostPort(ip.String(), strconv.Itoa(winPort(row.LocalPort))),
		})
	}
	return conns
}

func scanTCPTable(pid, family uint32, ipv6 bool) []ProcessConnection {
	buf, err := queryMIBTable(procGetExtendedTCPTable, family, tcpTableOwnerPID)
	if err != nil || len(buf) == 0 {
		return nil
	}

	var conns []ProcessConnection
	if ipv6 {
		for _, row := range copyTableRows[mibTCP6RowOwnerPID](buf) {
			if row.OwningPID != pid {
				continue
			}
			remoteIP := net.IP(row.RemoteAddr[:])
			if remoteIP.IsUnspecified() {
				continue
			}
			conns = append(conns, ProcessConnection{
				Protocol: "tcp",
				Family:   "ipv6",
				Local:    net.JoinHostPort(net.IP(row.LocalAddr[:]).String(), strconv.Itoa(winPort(row.LocalPort))),
				Remote:   net.JoinHostPort(remoteIP.String(), strconv.Itoa(winPort(row.RemotePort))),
				State:    windowsTCPStates[row.State],
			})
		}
		return conns
	}

	for _, row := range copyTableRows[mibTCPRowOwnerPID](buf) {
		if row.OwningPID != pid {
			continue
		}
		remoteIP := winIPv4(row.RemoteAddr)
		if remoteIP.IsUnspecified() {
			continue
		}
		conns = append(conns, ProcessConnection{
			Protocol: "tcp",
			Family:   "ipv4",
			Local:    net.JoinHostPort(winIPv4(row.LocalAddr).String(), strconv.Itoa(winPort(row.LocalPort))),
			Remote:   net.JoinHostPort(remoteIP.String(), strconv.Itoa(winPort(row.RemotePort))),
			State:    windowsTCPStates[row.State],
		})
	}
	return conns
}
