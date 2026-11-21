package client

import (
	"context"
	"net/netip"
	"testing"
)

func TestDiscoverPathMTUBinarySearch(t *testing.T) {
	oldProbe := dfPingProbe
	defer func() { dfPingProbe = oldProbe }()

	// Simulate a 1492-byte PPPoE line (max ICMP payload = 1492 - 28 = 1464).
	dfPingProbe = func(_ context.Context, _ netip.Addr, payloadBytes int) bool {
		return payloadBytes <= 1464
	}

	res := discoverPathMTU(context.Background(), netip.MustParseAddr("1.1.1.1"))
	if !res.Measured {
		t.Fatalf("expected Measured=true, got %+v", res)
	}
	if res.PathMTU != 1492 {
		t.Fatalf("PathMTU = %d, want 1492", res.PathMTU)
	}
	if res.RecommendedMTU != 1400 {
		t.Fatalf("RecommendedMTU = %d, want 1400", res.RecommendedMTU)
	}
	if res.WireGuardMTU != 1412 {
		t.Fatalf("WireGuardMTU = %d, want 1412", res.WireGuardMTU)
	}
}

func TestDiscoverPathMTUBlockedICMP(t *testing.T) {
	oldProbe := dfPingProbe
	defer func() { dfPingProbe = oldProbe }()

	dfPingProbe = func(_ context.Context, _ netip.Addr, _ int) bool { return false }

	res := discoverPathMTU(context.Background(), netip.MustParseAddr("1.1.1.1"))
	if res.Measured {
		t.Fatalf("expected Measured=false when ICMP is blocked, got %+v", res)
	}
	if res.PathMTU != 0 {
		t.Fatalf("expected PathMTU=0 when unmeasured, got %d", res.PathMTU)
	}
}
