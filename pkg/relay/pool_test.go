package relay

import (
	"net/netip"
	"testing"
)

func TestIPPoolAllocationAndReservation(t *testing.T) {
	subnet := netip.MustParsePrefix("10.88.0.0/24")
	pool, err := NewIPPool(subnet)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	gw := pool.Gateway()
	if gw.String() != "10.88.0.1" {
		t.Fatalf("expected gateway 10.88.0.1, got %s", gw)
	}

	// First allocation
	clientID1 := uint64(1001)
	sessionID1 := uint64(2001)
	ip1, resumed, err := pool.Allocate(clientID1, sessionID1)
	if err != nil || resumed {
		t.Fatalf("expected fresh ip allocation, got %v, resumed=%v", ip1, resumed)
	}
	if ip1.String() != "10.88.0.2" {
		t.Fatalf("expected 10.88.0.2, got %s", ip1)
	}

	// Second client
	clientID2 := uint64(1002)
	sessionID2 := uint64(2002)
	ip2, _, err := pool.Allocate(clientID2, sessionID2)
	if err != nil || ip2.String() != "10.88.0.3" {
		t.Fatalf("expected 10.88.0.3, got %s", ip2)
	}

	// Release client 1 with reservation
	pool.Release(ip1, clientID1, true)

	// Reconnect client 1 with new session
	sessionID1New := uint64(2003)
	ip1Reconnected, resumed, err := pool.Allocate(clientID1, sessionID1New)
	if err != nil || !resumed {
		t.Fatalf("expected resumed allocation, got ip=%v, resumed=%v", ip1Reconnected, resumed)
	}
	if ip1Reconnected != ip1 {
		t.Fatalf("expected reconnected ip %s, got %s", ip1, ip1Reconnected)
	}
}

func TestIPPoolReconnectStorm(t *testing.T) {
	// Subnet /30: exactly 1 usable client IP (.2)
	subnet := netip.MustParsePrefix("10.88.0.0/30")
	pool, err := NewIPPool(subnet)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	clientID := uint64(0xcafe)

	// Simulate rapid reconnects: each release then allocate
	var lastIP netip.Addr
	for i := 1; i <= 50; i++ {
		sessID := uint64(1000 + i)
		if lastIP.IsValid() {
			pool.Release(lastIP, clientID, true)
		}
		ip, _, err := pool.Allocate(clientID, sessID)
		if err != nil {
			t.Fatalf("reconnect #%d failed: %v (pool exhausted unexpectedly)", i, err)
		}
		if ip.String() != "10.88.0.2" {
			t.Fatalf("expected IP 10.88.0.2, got %s", ip)
		}
		lastIP = ip
	}
}
