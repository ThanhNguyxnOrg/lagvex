//go:build linux || darwin

package client

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestProbeICMPLocalhost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := probeICMP(ctx, net.ParseIP("127.0.0.1"), 2, 700*time.Millisecond)
	if err != nil {
		t.Skipf("ICMP sockets are not permitted in this environment: %v", err)
	}
	if result.Samples != 2 {
		t.Fatalf("samples = %d, want 2", result.Samples)
	}
	// The loopback interface answers echo requests, so RTT must be measured.
	if result.RTTMs <= 0 {
		t.Fatalf("expected a measured loopback RTT, got %+v", result)
	}
	if result.LossPct != 0 {
		t.Fatalf("loss = %v%%, want 0 on loopback", result.LossPct)
	}
}

func TestProbeICMPRefusesIPv6(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := probeICMP(ctx, net.ParseIP("::1"), 1, 200*time.Millisecond); err == nil {
		t.Fatalf("the IPv4-only ICMP probe must reject an IPv6 address")
	}
}
