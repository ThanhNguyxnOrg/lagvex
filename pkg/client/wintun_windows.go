//go:build windows

package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

var (
	modWintun   *syscall.LazyDLL
	modIphlpapi = syscall.NewLazyDLL("iphlpapi.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")

	procWintunCreateAdapter         *syscall.LazyProc
	procWintunOpenAdapter           *syscall.LazyProc
	procWintunCloseAdapter          *syscall.LazyProc
	procWintunStartSession          *syscall.LazyProc
	procWintunEndSession            *syscall.LazyProc
	procWintunGetReadWaitEvent      *syscall.LazyProc
	procWintunReceivePacket         *syscall.LazyProc
	procWintunReleaseReceivePacket  *syscall.LazyProc
	procWintunAllocateSendPacket    *syscall.LazyProc
	procWintunSendPacket            *syscall.LazyProc
	procWintunGetAdapterLUID        *syscall.LazyProc

	procConvertInterfaceLuidToIndex = modIphlpapi.NewProc("ConvertInterfaceLuidToIndex")
	procWaitForSingleObject         = modKernel32.NewProc("WaitForSingleObject")
)

func initWintunProcs(dllPath string) error {
	if modWintun != nil {
		return nil
	}
	if dllPath == "" {
		// Look in current exe dir, then bin/amd64
		exe, _ := os.Executable()
		candidates := []string{
			filepath.Join(filepath.Dir(exe), "wintun.dll"),
			filepath.Join(filepath.Dir(exe), "bin", "amd64", "wintun.dll"),
			`D:\Code\Lagvex\bin\amd64\wintun.dll`,
			"wintun.dll",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				dllPath = c
				break
			}
		}
		if dllPath == "" {
			dllPath = "wintun.dll"
		}
	}

	modWintun = syscall.NewLazyDLL(dllPath)
	procWintunCreateAdapter = modWintun.NewProc("WintunCreateAdapter")
	procWintunOpenAdapter = modWintun.NewProc("WintunOpenAdapter")
	procWintunCloseAdapter = modWintun.NewProc("WintunCloseAdapter")
	procWintunStartSession = modWintun.NewProc("WintunStartSession")
	procWintunEndSession = modWintun.NewProc("WintunEndSession")
	procWintunGetReadWaitEvent = modWintun.NewProc("WintunGetReadWaitEvent")
	procWintunReceivePacket = modWintun.NewProc("WintunReceivePacket")
	procWintunReleaseReceivePacket = modWintun.NewProc("WintunReleaseReceivePacket")
	procWintunAllocateSendPacket = modWintun.NewProc("WintunAllocateSendPacket")
	procWintunSendPacket = modWintun.NewProc("WintunSendPacket")
	procWintunGetAdapterLUID = modWintun.NewProc("WintunGetAdapterLUID")

	return nil
}

// WintunAdapter encapsulates an open WinTun adapter and session.
type WintunAdapter struct {
	mu           sync.Mutex
	adapterH     uintptr
	sessionH     uintptr
	readEvent    uintptr
	ifIndex      uint32
	name         string
	tunnelType   string
	closed       bool
}

// OpenOrCreateWintunAdapter creates a WinTun adapter (or opens if exists).
func OpenOrCreateWintunAdapter(name, tunnelType, dllPath string) (*WintunAdapter, error) {
	if err := initWintunProcs(dllPath); err != nil {
		return nil, err
	}

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	typePtr, err := syscall.UTF16PtrFromString(tunnelType)
	if err != nil {
		return nil, err
	}

	// Try create adapter
	r1, _, errSys := procWintunCreateAdapter.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(typePtr)),
		0,
	)

	adapterH := r1
	if adapterH == 0 {
		// If creation failed, try open existing
		r1, _, errSys = procWintunOpenAdapter.Call(uintptr(unsafe.Pointer(namePtr)))
		adapterH = r1
		if adapterH == 0 {
			return nil, fmt.Errorf("create or open wintun adapter %q: %w", name, errSys)
		}
	}

	// Query LUID and interface index
	var luid uint64
	procWintunGetAdapterLUID.Call(adapterH, uintptr(unsafe.Pointer(&luid)))

	var ifIndex uint32
	r1, _, _ = procConvertInterfaceLuidToIndex.Call(
		uintptr(unsafe.Pointer(&luid)),
		uintptr(unsafe.Pointer(&ifIndex)),
	)
	if r1 != 0 {
		procWintunCloseAdapter.Call(adapterH)
		return nil, fmt.Errorf("convert LUID to index failed: error code %d", r1)
	}

	// Start ring-buffer session with 0x400000 capacity (4 MB)
	const sessionRingCapacity = 0x400000
	sessionH, _, errSys := procWintunStartSession.Call(adapterH, uintptr(sessionRingCapacity))
	if sessionH == 0 {
		procWintunCloseAdapter.Call(adapterH)
		return nil, fmt.Errorf("start wintun session: %w", errSys)
	}

	eventH, _, _ := procWintunGetReadWaitEvent.Call(sessionH)

	return &WintunAdapter{
		adapterH:   adapterH,
		sessionH:   sessionH,
		readEvent:  eventH,
		ifIndex:    ifIndex,
		name:       name,
		tunnelType: tunnelType,
	}, nil
}

// InterfaceIndex returns the Windows interface index of the WinTun adapter.
func (w *WintunAdapter) InterfaceIndex() uint32 {
	return w.ifIndex
}

// ReadPacket reads a raw IP packet from the WinTun ring buffer, waiting if necessary.
func (w *WintunAdapter) ReadPacket(buf []byte) (int, error) {
	for {
		if w.closed {
			return 0, errors.New("adapter closed")
		}

		var pktSize uint32
		r1, _, _ := procWintunReceivePacket.Call(w.sessionH, uintptr(unsafe.Pointer(&pktSize)))
		if r1 != 0 {
			packetPtr := (*byte)(*(*unsafe.Pointer)(unsafe.Pointer(&r1)))
			size := int(pktSize)
			if size > len(buf) {
				size = len(buf)
			}

			// Copy bytes into caller's buffer
			srcSlice := unsafe.Slice(packetPtr, size)
			copy(buf, srcSlice)

			// Release packet in ring buffer
			procWintunReleaseReceivePacket.Call(w.sessionH, r1)
			return size, nil
		}

		// Wait for read event (timeout 1000ms to allow checking closed flag)
		const waitTimeout = 1000
		r1, _, _ = procWaitForSingleObject.Call(w.readEvent, uintptr(waitTimeout))
		if r1 != 0 && r1 != 0x102 { // 0x102 = WAIT_TIMEOUT
			return 0, fmt.Errorf("wait event error: 0x%x", r1)
		}
	}
}

// WritePacket writes a raw IP packet into the WinTun send ring buffer.
func (w *WintunAdapter) WritePacket(packet []byte) error {
	if w.closed {
		return errors.New("adapter closed")
	}

	r1, _, errSys := procWintunAllocateSendPacket.Call(w.sessionH, uintptr(len(packet)))
	if r1 == 0 {
		return fmt.Errorf("allocate send packet: %w", errSys)
	}

	dstSlice := unsafe.Slice((*byte)(*(*unsafe.Pointer)(unsafe.Pointer(&r1))), len(packet))
	copy(dstSlice, packet)

	procWintunSendPacket.Call(w.sessionH, r1)
	return nil
}

// Close closes the session and adapter.
func (w *WintunAdapter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true

	if w.sessionH != 0 {
		procWintunEndSession.Call(w.sessionH)
		w.sessionH = 0
	}
	if w.adapterH != 0 {
		procWintunCloseAdapter.Call(w.adapterH)
		w.adapterH = 0
	}
	return nil
}
