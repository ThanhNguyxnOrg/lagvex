//go:build !linux && !darwin && !windows

package client

import (
	"context"
	"fmt"
	"net"
	"time"
)

// probeICMP reports that this platform has no ICMP echo implementation yet.
func probeICMP(ctx context.Context, ip net.IP, samples int, timeout time.Duration) (icmpProbeResult, error) {
	return icmpProbeResult{}, fmt.Errorf("icmp probing is not implemented on this platform")
}
