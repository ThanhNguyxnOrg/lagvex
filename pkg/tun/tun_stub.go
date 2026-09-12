//go:build !linux

package tun

import (
	"errors"
)

var errNotSupported = errors.New("native Linux TUN device is not supported on this operating system")

// Device stub for non-Linux platforms.
type Device struct {
	name string
}

// Open stub.
func Open(name string) (*Device, error) {
	return nil, errNotSupported
}

func (d *Device) Name() string                { return d.name }
func (d *Device) Read(p []byte) (int, error)  { return 0, errNotSupported }
func (d *Device) Write(p []byte) (int, error) { return 0, errNotSupported }
func (d *Device) Close() error                { return nil }
func (d *Device) Configure(cidr string, mtu int) error {
	return errNotSupported
}
