package client

import (
	"context"
	"math"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		input    []float64
		expected float64
	}{
		{[]float64{}, 0},
		{[]float64{15.0}, 15.0},
		{[]float64{30.0, 10.0, 20.0}, 20.0},
		{[]float64{10.0, 20.0, 30.0, 40.0}, 25.0},
		{[]float64{5.5, 9.2, 1.1, 7.8, 3.4}, 5.5},
	}

	for i, tc := range tests {
		got := calculateMedian(tc.input)
		if math.Abs(got-tc.expected) > 0.001 {
			t.Errorf("[%d] calculateMedian(%v) = %f; want %f", i, tc.input, got, tc.expected)
		}
	}
}

func TestCalculateJitter(t *testing.T) {
	tests := []struct {
		input    []float64
		expected float64
	}{
		{[]float64{}, 0},
		{[]float64{20.0}, 0},
		// |15-10| = 5, |12-15| = 3 -> mean = 4.0
		{[]float64{10.0, 15.0, 12.0}, 4.0},
		// constant latency -> 0 jitter
		{[]float64{25.0, 25.0, 25.0}, 0.0},
	}

	for i, tc := range tests {
		got := calculateJitter(tc.input)
		if math.Abs(got-tc.expected) > 0.001 {
			t.Errorf("[%d] calculateJitter(%v) = %f; want %f", i, tc.input, got, tc.expected)
		}
	}
}

func TestFilterRelaysByContinent(t *testing.T) {
	relays := []profiles.RelayEndpoint{
		{ID: "r1", Continent: "Asia-Pacific", Name: "Singapore #1"},
		{ID: "r2", Continent: "Europe", Name: "Frankfurt #1"},
		{ID: "r3", Continent: "North America", Name: "Virginia #1"},
	}

	all := FilterRelaysByContinent(relays, "all")
	if len(all) != 3 {
		t.Fatalf("expected 3 relays, got %d", len(all))
	}

	asia := FilterRelaysByContinent(relays, "asia-pacific")
	if len(asia) != 1 || asia[0].ID != "r1" {
		t.Fatalf("expected 1 asia relay, got %v", asia)
	}

	eu := FilterRelaysByContinent(relays, "Europe")
	if len(eu) != 1 || eu[0].ID != "r2" {
		t.Fatalf("expected 1 eu relay, got %v", eu)
	}
}

// startMockRelayUDPServer starts a lightweight mock Lagvex UDP relay for testing prober.
func startMockRelayUDPServer(t *testing.T, psk []byte, simulatedDelay time.Duration) (*net.UDPConn, string) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}

	go func() {
		defer conn.Close()
		buf := make([]byte, protocol.MaxPacketSize)
		var sessCrypto *protocol.SessionCrypto
		var sessionID uint64

		for {
			n, remote, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			if simulatedDelay > 0 {
				time.Sleep(simulatedDelay)
			}

			if n == protocol.HandshakeReqLen {
				req, err := protocol.DecodeHandshakeRequest(psk, buf[:n])
				if err == nil {
					sessionID = 0x9988776655443322
					sessCrypto, _ = protocol.NewRelayCrypto(psk, req.Nonce, sessionID, req.ClientID)

					resp := protocol.HandshakeResponse{
						Status:    protocol.StatusOK,
						SessionID: sessionID,
						ClientIP:  netip.MustParseAddr("10.88.0.2"),
						GatewayIP: netip.MustParseAddr("10.88.0.1"),
						MTU:       1400,
						NonceEcho: req.Nonce,
					}
					respBuf := protocol.EncodeHandshakeResponse(psk, resp)
					_, _ = conn.WriteToUDP(respBuf, remote)
					continue
				}
			}

			// Handle AEAD session packets (Ping / Disconnect)
			if sessCrypto != nil && n >= protocol.SecureHeaderLen+protocol.TagLen {
				plain := make([]byte, protocol.MaxPacketSize)
				mtype, payload, err := sessCrypto.OpenPacket(plain, buf[:n])
				if err == nil {
					if mtype == protocol.TypePing {
						ts, _ := protocol.DecodeControlPayload(payload)
						pongBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
						pong := sessCrypto.EncodePong(pongBuf, sessionID, ts)
						_, _ = conn.WriteToUDP(pong, remote)
					} else if mtype == protocol.TypeDisconnect {
						sessCrypto = nil
						sessionID = 0
					}
				}
			}
		}
	}()

	return conn, conn.LocalAddr().String()
}

func TestProbeEndpointSuccess(t *testing.T) {
	psk := []byte("test-psk-prober-secret-123")
	_, endpoint := startMockRelayUDPServer(t, psk, 5*time.Millisecond)

	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 3,
		SampleTimeout:  1 * time.Second,
		SampleInterval: 10 * time.Millisecond,
	})

	relay := profiles.RelayEndpoint{
		ID:        "mock-sg-1",
		Name:      "Mock Singapore",
		Endpoint:  endpoint,
		Continent: "Asia-Pacific",
		PSK:       string(psk),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := prober.ProbeEndpoint(ctx, relay, 3)
	if !result.Reachable {
		t.Fatalf("expected relay to be reachable, got error: %s", result.Error)
	}

	if result.PacketLoss != 0.0 {
		t.Errorf("expected 0%% packet loss, got %f", result.PacketLoss)
	}

	if result.RTTMedianMs <= 0 {
		t.Errorf("expected positive median RTT, got %f", result.RTTMedianMs)
	}

	if result.Score <= 0 || result.Score >= 999999 {
		t.Errorf("unexpected score: %f", result.Score)
	}
}

func TestProbeEndpointUnreachable(t *testing.T) {
	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 2,
		SampleTimeout:  100 * time.Millisecond,
		SampleInterval: 10 * time.Millisecond,
	})

	relay := profiles.RelayEndpoint{
		ID:        "dead-relay",
		Name:      "Dead Relay",
		Endpoint:  "127.0.0.1:59999", // Unused port
		Continent: "Europe",
		PSK:       "some-psk",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := prober.ProbeEndpoint(ctx, relay, 2)
	if result.Reachable {
		t.Fatalf("expected dead relay to be unreachable")
	}

	if result.PacketLoss != 100.0 {
		t.Errorf("expected 100%% packet loss, got %f", result.PacketLoss)
	}

	if result.Score != 999999.0 {
		t.Errorf("expected score 999999.0 for unreachable node, got %f", result.Score)
	}
}

func TestProbeAllAndSelectBest(t *testing.T) {
	psk := []byte("fast-vs-slow-psk")
	_, fastEndpoint := startMockRelayUDPServer(t, psk, 2*time.Millisecond)
	_, slowEndpoint := startMockRelayUDPServer(t, psk, 40*time.Millisecond)

	relays := []profiles.RelayEndpoint{
		{ID: "slow-node", Name: "Slow Relay", Endpoint: slowEndpoint, Continent: "Europe", PSK: string(psk)},
		{ID: "dead-node", Name: "Dead Relay", Endpoint: "127.0.0.1:59998", Continent: "US", PSK: string(psk)},
		{ID: "fast-node", Name: "Fast Relay", Endpoint: fastEndpoint, Continent: "Asia", PSK: string(psk)},
	}

	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples:     2,
		SampleTimeout:      300 * time.Millisecond,
		SampleInterval:     10 * time.Millisecond,
		DefaultConcurrency: 4,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results := prober.ProbeAll(ctx, relays, 4)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Fastest node should be ranked #1 and marked IsOptimal
	if !results[0].IsOptimal {
		t.Errorf("expected top result to be IsOptimal = true")
	}
	if results[0].RelayID != "fast-node" {
		t.Errorf("expected fast-node to be top ranked, got %s (score %f vs %f)", results[0].RelayID, results[0].Score, results[1].Score)
	}

	// Dead node should be last
	if results[2].Reachable {
		t.Errorf("expected last node to be dead/unreachable")
	}

	// Test SelectBest
	best, err := prober.SelectBest(ctx, relays)
	if err != nil {
		t.Fatalf("SelectBest returned error: %v", err)
	}
	if best.RelayID != "fast-node" {
		t.Fatalf("SelectBest picked %s; want fast-node", best.RelayID)
	}
}
