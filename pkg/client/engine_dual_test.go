package client

import (
	"context"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/relay"
)

func TestSetDualPathHonestStates(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	// Off by default: no dualPath object at all.
	if stats := engine.Stats(); stats.DualPath != nil {
		t.Fatalf("dualPath present before any toggle: %+v", stats.DualPath)
	}

	// Enabling while disconnected records the honest reason.
	engine.SetDualPath(true)
	if !engine.IsDualPathEnabled() {
		t.Fatal("intent must be stored")
	}
	stats := engine.Stats()
	if stats.DualPath == nil {
		t.Fatal("dualPath missing after enabling")
	}
	if stats.DualPath.Active {
		t.Fatal("dual path must not be active while disconnected")
	}
	if !strings.Contains(stats.DualPath.Reason, "not connected") {
		t.Fatalf("reason = %q, want a not-connected explanation", stats.DualPath.Reason)
	}

	// Disabling tears down and clears the intent.
	engine.SetDualPath(false)
	if engine.IsDualPathEnabled() {
		t.Fatal("intent must clear")
	}
	if stats := engine.Stats(); stats.DualPath != nil {
		t.Fatalf("dualPath present after disabling: %+v", stats.DualPath)
	}
}

// seedDualFixture prepares a "connected" engine with a fresh standby and two
// catalog relays so SetDualPath gets past its guards.
func seedDualFixture(t *testing.T, standbyPSK, activePSK string) *Engine {
	t.Helper()

	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("profiles: %v", err)
	}
	_ = pm.AddRelay(profiles.RelayEndpoint{
		ID: "primary", Name: "Primary", Endpoint: "198.51.100.10:4433", PSK: activePSK,
	})
	_ = pm.AddRelay(profiles.RelayEndpoint{
		ID: "standby", Name: "Standby", Endpoint: "203.0.113.10:4433", PSK: standbyPSK,
	})

	engine, err := NewEngine(pm)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	engine.mu.Lock()
	engine.state = StateConnected
	engine.clientID = 0x1234
	engine.psk = []byte(activePSK)
	engine.relayAddr = netip.MustParseAddrPort("198.51.100.10:4433")
	engine.pingMs.Store(30)
	engine.mu.Unlock()

	engine.standbyMu.Lock()
	engine.standby = &ProbeResult{
		RelayID:  "standby",
		Name:     "Standby",
		Endpoint: "203.0.113.10:4433",
	}
	engine.standbyAt = time.Now()
	engine.standbyMu.Unlock()

	return engine
}

func TestSetDualPathRefusesDifferentPSK(t *testing.T) {
	engine := seedDualFixture(t, "standby-psk", "primary-psk")

	engine.SetDualPath(true)
	stats := engine.Stats()
	if stats.DualPath == nil || stats.DualPath.Active {
		t.Fatalf("dual path must stay inactive: %+v", stats.DualPath)
	}
	if !strings.Contains(stats.DualPath.Reason, "same PSK") {
		t.Fatalf("reason = %q, want a same-PSK explanation", stats.DualPath.Reason)
	}
}

func TestSetDualPathRefusesMissingStandbyPSK(t *testing.T) {
	engine := seedDualFixture(t, "", "primary-psk")

	engine.SetDualPath(true)
	stats := engine.Stats()
	if stats.DualPath == nil || stats.DualPath.Active {
		t.Fatalf("dual path must stay inactive: %+v", stats.DualPath)
	}
	if !strings.Contains(stats.DualPath.Reason, "no PSK") {
		t.Fatalf("reason = %q, want a missing-PSK explanation", stats.DualPath.Reason)
	}
}

func TestDualThresholdFor(t *testing.T) {
	if got := dualThresholdFor(0); got != 40*time.Millisecond {
		t.Fatalf("no-RTT threshold = %s, want 40ms", got)
	}
	if got := dualThresholdFor(30); got != 45*time.Millisecond {
		t.Fatalf("30ms RTT threshold = %s, want 45ms", got)
	}
	if got := dualThresholdFor(400); got != DualMaxRetransmitThreshold {
		t.Fatalf("400ms RTT threshold = %s, want clamped", got)
	}
}

func TestActivateDualAgainstLoopbackRelay(t *testing.T) {
	if testing.Short() {
		t.Skip("network integration test")
	}

	psk := []byte("dual-loopback-psk-0123456789")
	cfg := relay.Config{
		ListenAddr:  "127.0.0.1:0",
		HTTPAddr:    "127.0.0.1:0",
		TunName:     "lagvexdualcl",
		Subnet:      netip.MustParsePrefix("10.78.0.0/24"),
		PSK:         psk,
		MTU:         1400,
		IdleTimeout: 30 * time.Second,
	}
	srv, err := relay.NewServer(cfg)
	if err != nil {
		t.Fatalf("new relay: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()

	var udpAddr netip.AddrPort
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.UDPAddr(); addr.IsValid() {
			udpAddr = addr
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !udpAddr.IsValid() {
		t.Fatal("relay did not bind UDP in time")
	}

	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.mu.Lock()
	engine.state = StateConnected
	engine.clientID = 0xCAFE
	engine.mu.Unlock()

	ctxTun, cancelTun := context.WithCancel(context.Background())
	engine.cancelCtx = ctxTun
	engine.cancelFunc = cancelTun
	t.Cleanup(cancelTun)

	endpoint := udpAddr.String()

	// The user asked for dual mode; activation happens directly here.
	engine.dualEnabled.Store(true)

	// Activate synchronously against the loopback relay.
	engine.activateDual(endpoint, psk, 40*time.Millisecond, "loopback", "Loopback Relay")

	if !engine.dualActive.Load() {
		reason := ""
		if r := engine.dualReason.Load(); r != nil {
			reason = *r
		}
		t.Fatalf("dual activation failed: %s", reason)
	}

	engine.dualMu.Lock()
	session := engine.dualStandbySession
	sender := engine.dualSender
	engine.dualMu.Unlock()
	if session == 0 || sender == nil {
		t.Fatal("standby session or sender missing after activation")
	}
	if snap := sender.stats(); snap.GroupHash == 0 {
		t.Fatal("sender group hash must be set")
	}

	status := engine.dualStatusSnapshot()
	if status == nil || !status.Active || status.StandbyRelayName != "Loopback Relay" {
		t.Fatalf("status = %+v", status)
	}

	// Teardown is honest and complete.
	engine.deactivateDual("test teardown")
	if engine.dualActive.Load() {
		t.Fatal("dual active after teardown")
	}
	stats := engine.Stats()
	if stats.DualPath == nil || stats.DualPath.Active || stats.DualPath.Reason != "test teardown" {
		t.Fatalf("post-teardown status = %+v", stats.DualPath)
	}
}

func TestDualHandshakeRefusedByCapabilityMissing(t *testing.T) {
	// A relay that predates dual support would answer the base handshake
	// without the capability extension. Simulate at the protocol level: a
	// plain response must yield dualOK=false.
	psk := []byte("dual-psk")
	resp := protocol.HandshakeResponse{
		Status:    protocol.StatusOK,
		SessionID: 7,
		ClientIP:  netip.MustParseAddr("10.78.0.2"),
		GatewayIP: netip.MustParseAddr("10.78.0.1"),
		NonceEcho: 9,
	}
	plain := protocol.EncodeHandshakeResponse(psk, resp)
	if _, dualOK, err := protocol.DecodeHandshakeResponseExt(psk, plain, 9); err != nil || dualOK {
		t.Fatalf("plain response dual=%v err=%v, want false/nil", dualOK, err)
	}
}
