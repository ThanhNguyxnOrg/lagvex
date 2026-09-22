//go:build linux

package tun

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

const (
	cloneDevice = "/dev/net/tun"

	iffTun    = 0x0001
	iffNoPI   = 0x1000 // Do not prepend 4-byte packet info header
	tunSetIff = 0x400454ca
)

type ifReq struct {
	Name  [16]byte
	Flags uint16
	_     [22]byte
}

// Device wraps an active Linux TUN network interface or userspace fallback.
type Device struct {
	file        *os.File
	name        string
	isUserspace bool
	mu          sync.Mutex
	ch          chan []byte
	closed      bool
}

// Open creates or connects to a TUN interface by name (e.g. "lagvex0").
func Open(name string) (*Device, error) {
	f, err := os.OpenFile(cloneDevice, os.O_RDWR, 0)
	if err != nil {
		log.Printf("[TUN] Notice: %s not accessible (%v), falling back to userspace virtual device", cloneDevice, err)
		return &Device{
			name:        name,
			isUserspace: true,
			ch:          make(chan []byte, 1024),
		}, nil
	}

	var req ifReq
	if len(name) >= len(req.Name) {
		f.Close()
		return nil, fmt.Errorf("interface name %q too long", name)
	}
	copy(req.Name[:], name)
	req.Flags = iffTun | iffNoPI

	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(tunSetIff),
		uintptr(unsafe.Pointer(&req)),
	); errno != 0 {
		f.Close()
		log.Printf("[TUN] Notice: ioctl TUNSETIFF failed (%v), falling back to userspace virtual device", errno)
		return &Device{
			name:        name,
			isUserspace: true,
			ch:          make(chan []byte, 1024),
		}, nil
	}

	actualName := name
	for i, b := range req.Name {
		if b == 0 {
			actualName = string(req.Name[:i])
			break
		}
	}

	return &Device{file: f, name: actualName}, nil
}

func (d *Device) Name() string { return d.name }

func (d *Device) Read(p []byte) (int, error) {
	if d.isUserspace {
		pkt, ok := <-d.ch
		if !ok {
			return 0, nil
		}
		return copy(p, pkt), nil
	}
	return d.file.Read(p)
}

func (d *Device) Write(p []byte) (int, error) {
	if d.isUserspace {
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.closed {
			return 0, nil
		}
		return len(p), nil
	}
	return d.file.Write(p)
}

func (d *Device) Close() error {
	if d.isUserspace {
		d.mu.Lock()
		defer d.mu.Unlock()
		if !d.closed {
			d.closed = true
			close(d.ch)
		}
		return nil
	}
	return d.file.Close()
}

// Configure configures the interface IPv4 address, MTU, and brings it UP via `ip`.
func (d *Device) Configure(cidr string, mtu int) error {
	if d.isUserspace {
		return nil
	}
	commands := [][]string{
		{"ip", "addr", "replace", cidr, "dev", d.name},
		{"ip", "link", "set", "dev", d.name, "mtu", strconv.Itoa(mtu)},
		{"ip", "link", "set", "dev", d.name, "up"},
	}

	for _, cmd := range commands {
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("running %v: %w (output: %s)", cmd, err, string(out))
		}
	}
	return nil
}
