//go:build !windows && !darwin && !linux

package client

import "errors"

var errWintunNotSupported = errors.New("WinTun is only supported on Windows")

type WintunAdapter struct{}

func OpenOrCreateWintunAdapter(name, tunnelType, dllPath string) (*WintunAdapter, error) {
	return nil, errWintunNotSupported
}

func (w *WintunAdapter) InterfaceIndex() uint32 { return 0 }
func (w *WintunAdapter) ReadPacket(buf []byte) (int, error) { return 0, errWintunNotSupported }
func (w *WintunAdapter) WritePacket(packet []byte) error { return errWintunNotSupported }
func (w *WintunAdapter) Close() error { return nil }
