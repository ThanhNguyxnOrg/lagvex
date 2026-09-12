//go:build linux

package client

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

type WintunAdapter struct {
	file   *os.File
	name   string
	closed bool
}

const (
	cloneDevice = "/dev/net/tun"
	iffTun      = 0x0001
	iffNoPI     = 0x1000
	tunSetIff   = 0x400454ca
)

type ifReq struct {
	Name  [16]byte
	Flags uint16
	_     [22]byte
}

func OpenOrCreateWintunAdapter(name, tunnelType, dllPath string) (*WintunAdapter, error) {
	if name == "" {
		name = "lagvex-tun0"
	}

	f, err := os.OpenFile(cloneDevice, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cloneDevice, err)
	}

	var req ifReq
	copy(req.Name[:], name)
	req.Flags = iffTun | iffNoPI

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(tunSetIff), uintptr(unsafe.Pointer(&req))); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("ioctl TUNSETIFF: %w", errno)
	}

	actualName := name
	for i, b := range req.Name {
		if b == 0 {
			actualName = string(req.Name[:i])
			break
		}
	}

	return &WintunAdapter{
		file: f,
		name: actualName,
	}, nil
}

func (w *WintunAdapter) InterfaceIndex() uint32 {
	iface, err := net.InterfaceByName(w.name)
	if err != nil {
		return 0
	}
	return uint32(iface.Index)
}

func (w *WintunAdapter) ReadPacket(buf []byte) (int, error) {
	return w.file.Read(buf)
}

func (w *WintunAdapter) WritePacket(packet []byte) error {
	_, err := w.file.Write(packet)
	return err
}

func (w *WintunAdapter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	return w.file.Close()
}
