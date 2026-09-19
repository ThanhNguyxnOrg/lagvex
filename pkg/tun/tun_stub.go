//go:build !linux

package tun

import (
	"sync"
)

// Device provides a userspace virtual network device for non-Linux platforms (e.g. Windows/macOS dev & hosting).
type Device struct {
	name   string
	mtu    int
	cidr   string
	closed bool
	mu     sync.Mutex
	ch     chan []byte
}

// Open creates a userspace virtual tunnel device.
func Open(name string) (*Device, error) {
	return &Device{
		name: name,
		mtu:  1400,
		ch:   make(chan []byte, 1024),
	}, nil
}

func (d *Device) Name() string { return d.name }

func (d *Device) Read(p []byte) (int, error) {
	pkt, ok := <-d.ch
	if !ok {
		return 0, nil
	}
	n := copy(p, pkt)
	return n, nil
}

func (d *Device) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return 0, nil
	}
	return len(p), nil
}

func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.closed {
		d.closed = true
		close(d.ch)
	}
	return nil
}

func (d *Device) Configure(cidr string, mtu int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cidr = cidr
	d.mtu = mtu
	return nil
}

