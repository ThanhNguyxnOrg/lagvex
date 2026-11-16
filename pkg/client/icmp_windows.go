//go:build windows

package client

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	iphlpapiICMP        = windows.NewLazySystemDLL("iphlpapi.dll")
	procIcmpCreateFile  = iphlpapiICMP.NewProc("IcmpCreateFile")
	procIcmpCloseHandle = iphlpapiICMP.NewProc("IcmpCloseHandle")
	procIcmpSendEcho    = iphlpapiICMP.NewProc("IcmpSendEcho")
)

// icmpUnavailable is set once when the IP helper refuses to create an ICMP
// handle, so later probes fail fast instead of retrying every relay.
var icmpUnavailable atomic.Bool

type icmpOptions struct {
	Ttl         uint8
	Tos         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       icmpOptions
}

// probeICMP measures ICMP echo RTT through the Windows IP helper API
// (IcmpSendEcho), which does not require administrator rights.
func probeICMP(ctx context.Context, ip net.IP, samples int, timeout time.Duration) (icmpProbeResult, error) {
	if icmpUnavailable.Load() {
		return icmpProbeResult{}, fmt.Errorf("icmp handles are unavailable on this host")
	}
	if samples <= 0 {
		samples = 2
	}
	if timeout <= 0 {
		timeout = 700 * time.Millisecond
	}

	v4 := ip.To4()
	if v4 == nil {
		return icmpProbeResult{}, fmt.Errorf("icmp probe needs an IPv4 address")
	}

	handle, _, callErr := procIcmpCreateFile.Call()
	if handle == 0 || handle == ^uintptr(0) {
		icmpUnavailable.Store(true)
		return icmpProbeResult{}, fmt.Errorf("icmp create file: %v", callErr)
	}
	defer procIcmpCloseHandle.Call(handle)

	// IcmpSendEcho takes the address in network byte order.
	addr := uint32(v4[0]) | uint32(v4[1])<<8 | uint32(v4[2])<<16 | uint32(v4[3])<<24
	payload := []byte("lagvex\x01\x02")
	replyBuf := make([]byte, int(unsafe.Sizeof(icmpEchoReply{}))+len(payload)+64)

	var rtts []float64
	for seq := 0; seq < samples; seq++ {
		if ctx.Err() != nil {
			break
		}
		t0 := time.Now()
		ret, _, _ := procIcmpSendEcho.Call(
			handle,
			uintptr(addr),
			uintptr(unsafe.Pointer(&payload[0])),
			uintptr(len(payload)),
			0,
			uintptr(unsafe.Pointer(&replyBuf[0])),
			uintptr(len(replyBuf)),
			uintptr(timeout.Milliseconds()),
		)
		if ret == 0 {
			continue
		}
		reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuf[0]))
		if reply.Status != 0 { // IP_SUCCESS
			continue
		}
		rtts = append(rtts, float64(time.Since(t0).Microseconds())/1000.0)
	}

	lost := samples - len(rtts)
	result := icmpProbeResult{
		LossPct: float64(lost) / float64(samples) * 100.0,
		Samples: samples,
	}
	if len(rtts) > 0 {
		result.RTTMs = calculateMedian(rtts)
	}
	return result, nil
}
