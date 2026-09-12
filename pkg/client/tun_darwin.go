//go:build darwin

package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

// WintunAdapter on macOS implements the TUN interface via native utun.
type WintunAdapter struct {
	file   *os.File
	name   string
	closed bool
}

const (
	sysprotoControl = 2
	afSysControl    = 32
	ctlIOCGINFO     = 0xc0644e03
	utunOptIfname   = 2
	utunControlName = "com.apple.net.utun_control"
)

type ctlInfo struct {
	ctlID   uint32
	ctlName [96]byte
}

type sockaddrCtl struct {
	scLen      uint8
	scFamily   uint8
	ssSysaddr  uint16
	scID       uint32
	scUnit     uint32
	scReserved [5]uint32
}

// OpenOrCreateWintunAdapter on macOS creates a native utun interface.
func OpenOrCreateWintunAdapter(name, tunnelType, dllPath string) (*WintunAdapter, error) {
	fd, err := syscall.Socket(afSysControl, syscall.SOCK_DGRAM, sysprotoControl)
	if err != nil {
		return nil, fmt.Errorf("socket(SYSPROTO_CONTROL): %w", err)
	}

	var info ctlInfo
	copy(info.ctlName[:], utunControlName)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(ctlIOCGINFO), uintptr(unsafe.Pointer(&info))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("ioctl(CTLIOCGINFO): %w", errno)
	}

	sc := sockaddrCtl{
		scLen:     uint8(unsafe.Sizeof(sockaddrCtl{})),
		scFamily:  afSysControl,
		ssSysaddr: 1, // AF_SYS_CONTROL
		scID:      info.ctlID,
		scUnit:    0, // Dynamically allocate utunX
	}

	if _, _, errno := syscall.Syscall(syscall.SYS_CONNECT, uintptr(fd), uintptr(unsafe.Pointer(&sc)), uintptr(unsafe.Sizeof(sc))); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("connect(AF_SYS_CONTROL): %w", errno)
	}

	var ifNameBuf [64]byte
	ifNameLen := uint32(len(ifNameBuf))
	if _, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		sysprotoControl,
		utunOptIfname,
		uintptr(unsafe.Pointer(&ifNameBuf[0])),
		uintptr(unsafe.Pointer(&ifNameLen)),
		0,
	); errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("getsockopt(UTUN_OPT_IFNAME): %w", errno)
	}

	ifName := string(ifNameBuf[:ifNameLen-1])
	file := os.NewFile(uintptr(fd), ifName)

	return &WintunAdapter{
		file: file,
		name: ifName,
	}, nil
}

func (w *WintunAdapter) InterfaceIndex() uint32 {
	iface, err := net.InterfaceByName(w.name)
	if err != nil {
		return 0
	}
	return uint32(iface.Index)
}

// ReadPacket reads from utun (skipping the 4-byte macOS AF header).
func (w *WintunAdapter) ReadPacket(buf []byte) (int, error) {
	if w.closed {
		return 0, errors.New("adapter closed")
	}

	tempBuf := make([]byte, len(buf)+4)
	n, err := w.file.Read(tempBuf)
	if err != nil {
		return 0, err
	}
	if n <= 4 {
		return 0, nil
	}

	copy(buf, tempBuf[4:n])
	return n - 4, nil
}

// WritePacket writes to utun (prepending 4-byte AF_INET header).
func (w *WintunAdapter) WritePacket(packet []byte) error {
	if w.closed {
		return errors.New("adapter closed")
	}

	tempBuf := make([]byte, len(packet)+4)
	tempBuf[3] = syscall.AF_INET // Protocol family IPv4 in network byte order
	copy(tempBuf[4:], packet)

	_, err := w.file.Write(tempBuf)
	return err
}

func (w *WintunAdapter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	return w.file.Close()
}
