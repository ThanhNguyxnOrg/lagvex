//go:build !linux

package tun

import (
	"errors"
	"sync"
)

// Device provides a userspace virtual network device for non-Linux platforms (e.g. Windows/macOS dev & hosting).
type Device struct {
	name      string
	mtu       int
	cidr      string
	closed    bool
	mu        sync.Mutex
	ch        chan []byte
	writeSink func([]byte)
}

// Open fails on non-Linux platforms. A userspace stub used to accept the TUN
// handle and silently discard every relayed packet, which made the relay look
// healthy while forwarding nothing. Hosting a relay requires a Linux TUN
// interface for kernel forwarding.
func Open(name string) (*Device, error) {
	return nil, errors.New("relay packet forwarding requires Linux /dev/net/tun; Windows and macOS relay hosting is not implemented")
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

// SetWriteSink mirrors the Linux userspace hook. Open always fails on
// non-Linux platforms, so no device ever exists to observe.
func (d *Device) SetWriteSink(sink func([]byte)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.writeSink = sink
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
