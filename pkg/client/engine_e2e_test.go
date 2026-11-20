package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/relay"
)

// startE2ERelay runs a real relay on loopback (userspace TUN fallback) so the
// client engine can be exercised against the actual data plane instead of
// mocks. The relay package does not import the client package, so there is no
// import cycle.
func startE2ERelay(t *testing.T, psk []byte) *relay.Server {
	t.Helper()

	if runtime.GOOS != "linux" {
		t.Skip("client/relay end-to-end test requires the Linux userspace TUN fallback")
	}

	srv, err := relay.NewServer(relay.Config{
		ListenAddr:  "127.0.0.1:0",
		HTTPAddr:    "127.0.0.1:0",
		TunName:     "lagvexe2e",
		Subnet:      netip.MustParsePrefix("10.77.0.0/24"),
		PSK:         psk,
		MTU:         1400,
		IdleTimeout: 30 * time.Second,
		MaxClients:  8,
	})
	if err != nil {
		t.Fatalf("new relay: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			_ = err
		case <-time.After(5 * time.Second):
			t.Log("relay did not stop within 5s")
		}
	})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if srv.UDPAddr().IsValid() && srv.HTTPAddr().IsValid() {
			return srv
		}
		select {
		case err := <-done:
			t.Fatalf("relay Start failed: %v", err)
		default:
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("relay did not bind UDP and HTTP in time")
	return nil
}

// relayActiveUsers reads the live session count from the relay's health
// endpoint. It is the relay's own view, not a client-side guess.
func relayActiveUsers(t *testing.T, httpAddr netip.AddrPort) int {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/health", httpAddr))
	if err != nil {
		t.Fatalf("relay health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var out struct {
		ActiveUsers int `json:"activeUsers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode relay health: %v", err)
	}
	return out.ActiveUsers
}

// waitActiveUsers polls the relay until it reports the expected session count.
func waitActiveUsers(t *testing.T, httpAddr netip.AddrPort, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if relayActiveUsers(t, httpAddr) == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("relay never reported %d active session(s)", want)
}

// TestProbeAnchorViaRelayRefusals covers the validation that keeps the relay
// probe honest: unknown anchors are rejected locally and a disconnected
// tunnel cannot claim to measure anything.
func TestProbeAnchorViaRelayRefusals(t *testing.T) {
	eng, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if _, _, err := eng.ProbeAnchorViaRelay(context.Background(), 200); err == nil {
		t.Fatal("unknown anchor id was accepted")
	} else if !strings.Contains(err.Error(), "unknown anchor id") {
		t.Fatalf("error = %v, want an unknown-anchor error", err)
	}

	for id := range map[byte]struct{}{1: {}, 5: {}, 16: {}} {
		if _, _, err := eng.ProbeAnchorViaRelay(context.Background(), id); err == nil {
			t.Fatalf("anchor %d probe was accepted while disconnected", id)
		} else if !strings.Contains(err.Error(), "not connected") {
			t.Fatalf("anchor %d error = %v, want a not-connected error", id, err)
		}
	}
}

// TestEngineConnectKeepaliveDisconnect is the client's end-to-end test: a real
// handshake over UDP, the authenticated keepalive, and a clean disconnect that
// the relay observes. It also covers the pumped goroutines for a few ticks.
func TestEngineConnectKeepaliveDisconnect(t *testing.T) {
	psk := []byte("e2e-psk-0123456789abcdef")
	srv := startE2ERelay(t, psk)
	udpAddr := srv.UDPAddr()
	httpAddr := srv.HTTPAddr()

	eng, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	if err := eng.Connect(udpAddr.String(), psk, "", "", false); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = eng.Disconnect() })

	stats := eng.Stats()
	if stats.State != StateConnected {
		t.Fatalf("state after connect = %s, want connected", stats.State)
	}
	if !strings.HasPrefix(stats.ClientIP, "10.77.") {
		t.Fatalf("client IP = %q, want an address from the relay subnet", stats.ClientIP)
	}
	if stats.GatewayIP != "10.77.0.1" {
		t.Fatalf("gateway IP = %q, want 10.77.0.1", stats.GatewayIP)
	}
	if stats.RelayAddr == "" {
		t.Fatal("relay address is empty after connect")
	}
	if stats.LastError != "" {
		t.Fatalf("unexpected last error after connect: %q", stats.LastError)
	}
	if stats.PingSampled {
		t.Fatal("ping was reported as sampled before any keepalive pong")
	}
	if stats.ActiveGame != "" || stats.ActiveRegion != "" {
		t.Fatalf("connect without a game reported activeGame=%q activeRegion=%q", stats.ActiveGame, stats.ActiveRegion)
	}
	// The engine must name the adapter it actually used. On a host without a
	// usable TUN device this is the userspace fallback, which the dashboard
	// labels as control-only rather than acceleration.
	if !strings.HasPrefix(stats.DriverMode, "Kernel (") && stats.DriverMode != "Userspace QoS" {
		t.Fatalf("driver mode = %q, want a kernel mode or Userspace QoS", stats.DriverMode)
	}

	// The relay must see exactly one session. This proves the handshake went
	// over the wire and was accepted; a client-local state flip would not.
	waitActiveUsers(t, httpAddr, 1)

	// Keepalive: the engine sends an authenticated Ping every 2s and records
	// a Pong received through the encrypted data plane. lastPong is seeded at
	// connect time, so we wait for it to move forward: only a real Pong from
	// the relay can update it.
	initialPong := eng.lastPong.Load()
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) && eng.lastPong.Load() == initialPong {
		time.Sleep(25 * time.Millisecond)
	}
	if eng.lastPong.Load() == initialPong {
		t.Fatal("no authenticated pong received from the relay")
	}
	// Loopback RTT may legitimately round down to 0ms, so only reject a
	// negative value (which would mean fabricated telemetry).
	stats = eng.Stats()
	if !stats.PingSampled {
		t.Fatal("pong arrived but pingSampled stayed false")
	}
	if got := stats.PingMs; got < 0 {
		t.Fatalf("ping = %dms, want a measured value >= 0", got)
	}

	// Connecting while connected must be refused without disturbing the
	// live tunnel.
	if err := eng.Connect(udpAddr.String(), psk, "", "", false); err == nil {
		t.Fatal("second connect while connected was accepted")
	}
	if got := eng.Stats().State; got != StateConnected {
		t.Fatalf("state after refused second connect = %s, want connected", got)
	}

	if err := eng.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	stats = eng.Stats()
	if stats.State != StateDisconnected {
		t.Fatalf("state after disconnect = %s, want disconnected", stats.State)
	}
	if stats.PingMs != 0 || stats.BytesUp != 0 || stats.BytesDown != 0 {
		t.Fatalf("telemetry was not reset after disconnect: ping=%d up=%d down=%d",
			stats.PingMs, stats.BytesUp, stats.BytesDown)
	}
	if stats.PingSampled {
		t.Fatal("pingSampled stayed true after disconnect")
	}
	if stats.DriverMode != "" {
		t.Fatalf("driverMode = %q after disconnect, want it cleared with the session", stats.DriverMode)
	}
	// The relay destroys the session on the authenticated Disconnect frame.
	waitActiveUsers(t, httpAddr, 0)

	// Disconnecting again is a documented no-op.
	if err := eng.Disconnect(); err != nil {
		t.Fatalf("second disconnect: %v", err)
	}
}

// TestEngineConnectRefusals covers the validation paths that must fail before
// any tunnel is reported as connected.
func TestEngineConnectRefusals(t *testing.T) {
	psk := []byte("e2e-psk-0123456789abcdef")
	srv := startE2ERelay(t, psk)
	endpoint := srv.UDPAddr().String()

	t.Run("game without region", func(t *testing.T) {
		eng, err := NewEngine(nil)
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}
		err = eng.Connect(endpoint, psk, "valorant", "", false)
		if err == nil || !strings.Contains(err.Error(), "requires a server region") {
			t.Fatalf("connect error = %v, want a server-region error", err)
		}
		if got := eng.Stats().State; got != StateDisconnected {
			t.Fatalf("state = %s, want disconnected after a pre-flight failure", got)
		}
	})

	t.Run("unknown game with no profiles", func(t *testing.T) {
		eng, err := NewEngine(nil)
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}
		err = eng.Connect(endpoint, psk, "valorant", "eu-west", false)
		if err == nil || !strings.Contains(err.Error(), "profiles are unavailable") {
			t.Fatalf("connect error = %v, want a profiles-unavailable error", err)
		}
	})

	t.Run("wrong PSK", func(t *testing.T) {
		eng, err := NewEngine(nil)
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}
		start := time.Now()
		err = eng.Connect(endpoint, []byte("wrong-psk-000000000000"), "", "", false)
		if err == nil || !strings.Contains(err.Error(), "handshake failed") {
			t.Fatalf("connect error = %v, want a handshake failure", err)
		}
		if elapsed := time.Since(start); elapsed < 2*time.Second {
			t.Fatalf("handshake gave up after %v; the relay silently drops bad PSKs, so three 1s attempts are expected", elapsed)
		}
		stats := eng.Stats()
		if stats.State != StateDisconnected {
			t.Fatalf("state = %s, want disconnected after a bad handshake", stats.State)
		}
		if !strings.Contains(stats.LastError, "handshake failed") {
			t.Fatalf("last error = %q, want the handshake failure", stats.LastError)
		}
	})
}
