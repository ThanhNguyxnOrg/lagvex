//go:build linux || darwin

package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"
)

// icmpUnavailable is set once when the OS refuses a datagram ICMP socket, so
// later probes fail fast instead of paying a socket error per relay.
var icmpUnavailable atomic.Bool

// probeICMP measures ICMP echo RTT with an unprivileged datagram ICMP socket.
// Linux requires net.ipv4.ping_group_range to include this process's group;
// macOS permits datagram ICMP for the current user. No raw socket or root is
// used here, and permission failures are reported, never faked.
func probeICMP(ctx context.Context, ip net.IP, samples int, timeout time.Duration) (icmpProbeResult, error) {
	if icmpUnavailable.Load() {
		return icmpProbeResult{}, fmt.Errorf("icmp sockets are unavailable on this host")
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

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_ICMP)
	if err != nil {
		if err == unix.EACCES || err == unix.EPERM || err == unix.EPROTONOSUPPORT {
			icmpUnavailable.Store(true)
		}
		return icmpProbeResult{}, fmt.Errorf("icmp socket: %w", err)
	}
	defer func() { _ = unix.Close(fd) }()

	sa := &unix.SockaddrInet4{Addr: [4]byte{v4[0], v4[1], v4[2], v4[3]}}
	id := uint16(os.Getpid() & 0xffff)

	var rtts []float64
	buf := make([]byte, 1500)

	for seq := 0; seq < samples; seq++ {
		if ctx.Err() != nil {
			break
		}
		deadline := time.Now().Add(timeout)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}

		req := buildICMPEchoRequest(id, uint16(seq))
		start := time.Now()
		if err := unix.Sendto(fd, req, 0, sa); err != nil {
			continue
		}

	sample:
		for {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				break sample
			}
			pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
			ready, err := unix.Poll(pfd, int(remaining.Milliseconds())+1)
			if err != nil {
				if err == unix.EINTR {
					continue
				}
				break sample
			}
			if ready == 0 {
				break sample
			}

			n, _, err := unix.Recvfrom(fd, buf, 0)
			if err != nil {
				break sample
			}
			if matchICMPEchoReply(buf[:n], uint16(seq)) {
				rtts = append(rtts, float64(time.Since(start).Microseconds())/1000.0)
				break sample
			}
		}
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
