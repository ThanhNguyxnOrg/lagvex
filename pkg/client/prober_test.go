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

func TestRoutedLocalUDPAddr(t *testing.T) {
	// The helper only connects a throwaway datagram socket: no packet is sent.
	local, err := routedLocalUDPAddr(&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1})
	if err != nil {
		t.Fatalf("routedLocalUDPAddr loopback: %v", err)
	}
	if local == nil || local.IP == nil || !local.IP.IsLoopback() {
		t.Fatalf("routedLocalUDPAddr loopback = %+v; want a loopback source address", local)
	}

	if _, err := routedLocalUDPAddr(nil); err == nil {
		t.Fatal("routedLocalUDPAddr(nil) should fail, got nil error")
	}
}

func TestAdaptiveSampleTimeout(t *testing.T) {
	p := NewProber() // base SampleTimeout 800ms
	const endpoint = "127.0.0.1:4433"

	// finishProbe creates the history map lazily; seed it here.
	p.histMu.Lock()
	p.history = make(map[string][]probeHistoryEntry)
	p.histMu.Unlock()

	if got := p.adaptiveSampleTimeout(endpoint); got != p.cfg.SampleTimeout {
		t.Fatalf("no history: adaptiveSampleTimeout = %v; want base %v", got, p.cfg.SampleTimeout)
	}

	p.histMu.Lock()
	p.history[endpoint] = []probeHistoryEntry{
		{RTTMs: 200, Reached: true},
		{RTTMs: 220, Reached: true},
	}
	p.histMu.Unlock()

	// 3x the last measured RTT (220ms) = 660ms, still below the base.
	if got := p.adaptiveSampleTimeout(endpoint); got != p.cfg.SampleTimeout {
		t.Fatalf("slow-ish history: adaptiveSampleTimeout = %v; want base %v", got, p.cfg.SampleTimeout)
	}

	p.histMu.Lock()
	p.history[endpoint] = []probeHistoryEntry{
		{RTTMs: 900, Reached: true},
	}
	p.histMu.Unlock()
	// 3x 900ms = 2.7s: widened, but within the ceiling.
	if got := p.adaptiveSampleTimeout(endpoint); got != 2700*time.Millisecond {
		t.Fatalf("slow history: adaptiveSampleTimeout = %v; want 2.7s", got)
	}

	p.histMu.Lock()
	p.history[endpoint] = []probeHistoryEntry{
		{RTTMs: 4000, Reached: true},
	}
	p.histMu.Unlock()
	if got := p.adaptiveSampleTimeout(endpoint); got != probeSampleTimeoutCeil {
		t.Fatalf("very slow history: adaptiveSampleTimeout = %v; want ceiling %v", got, probeSampleTimeoutCeil)
	}

	p.histMu.Lock()
	p.history[endpoint] = []probeHistoryEntry{
		{RTTMs: 4000, Reached: false},
	}
	p.histMu.Unlock()
	if got := p.adaptiveSampleTimeout(endpoint); got != p.cfg.SampleTimeout {
		t.Fatalf("unreachable history: adaptiveSampleTimeout = %v; want base %v", got, p.cfg.SampleTimeout)
	}
}

func TestProbeResultCarriesMethodAndSource(t *testing.T) {
	// The probe never reaches a relay here; it must still report the method
	// it used and the local source address it bound.
	cfg := DefaultProberConfig()
	cfg.ICMPProbe = false
	cfg.SampleTimeout = 150 * time.Millisecond
	relay := profiles.RelayEndpoint{
		ID:       "local",
		Name:     "Local",
		Endpoint: "127.0.0.1:1",
		PSK:      "test-psk-0123456789abcdef",
	}
	result := NewProberWithConfig(cfg).ProbeEndpoint(context.Background(), relay, 1)
	if result.Error == "" {
		t.Fatal("probing a closed port should set an error")
	}
	if result.ProbeMethod != ProbeMethodTunnelUDP {
		t.Fatalf("ProbeMethod = %q; want %q", result.ProbeMethod, ProbeMethodTunnelUDP)
	}
	if result.SourceAddress == "" || net.ParseIP(result.SourceAddress) == nil {
		t.Fatalf("SourceAddress = %q; want an IP literal", result.SourceAddress)
	}
}

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
	return startMockRelayUDPServerMode(t, psk, simulatedDelay, true)
}

// startMockRelayUDPServerMode additionally controls whether the mock answers
// anchor probes, so tests can cover the graceful "no reply" path.
func startMockRelayUDPServerMode(t *testing.T, psk []byte, simulatedDelay time.Duration, answerAnchors bool) (*net.UDPConn, string) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}

	go func() {
		defer func() { _ = conn.Close() }()
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
					switch mtype {
					case protocol.TypePing:
						ts, _ := protocol.DecodeControlPayload(payload)
						pongBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
						pong := sessCrypto.EncodePong(pongBuf, sessionID, ts)
						_, _ = conn.WriteToUDP(pong, remote)
					case protocol.TypeAnchorProbe:
						if !answerAnchors {
							continue
						}
						packed, _ := protocol.DecodeControlPayload(payload)
						seq, _ := protocol.DecodeAnchorProbeRequest(packed)
						replyBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
						reply := sessCrypto.EncodeAnchorProbeReply(replyBuf, sessionID,
							protocol.EncodeAnchorProbeReply(seq, 12_345, 0))
						_, _ = conn.WriteToUDP(reply, remote)
					case protocol.TypeDisconnect:
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

func TestCalculateRouteAdvice(t *testing.T) {
	relay := &ProbeResult{
		RelayID:     "relay-sg",
		Name:        "Singapore #1",
		RTTMedianMs: 32.0,
		JitterMs:    1.2,
		PacketLoss:  0.0,
		Reachable:   true,
	}

	// 1. Direct path is 15ms (faster by >5ms) and 0% loss -> VerdictDirect
	advice := CalculateRouteAdvice(15.0, 0.0, relay)
	if advice.Verdict != VerdictDirect {
		t.Errorf("expected VerdictDirect for 15ms direct vs 32ms relay, got %s", advice.Verdict)
	}

	// 2. Direct path is 60ms (slower by >5ms) -> VerdictBoost
	advice = CalculateRouteAdvice(60.0, 0.0, relay)
	if advice.Verdict != VerdictBoost {
		t.Errorf("expected VerdictBoost for 60ms direct vs 32ms relay, got %s", advice.Verdict)
	}
	if advice.SavingsMs != 28.0 {
		t.Errorf("expected 28.0ms savings, got %f", advice.SavingsMs)
	}

	// 3. Direct path has 5% packet loss -> VerdictBoost
	advice = CalculateRouteAdvice(32.0, 5.0, relay)
	if advice.Verdict != VerdictBoost {
		t.Errorf("expected VerdictBoost when direct has 5%% packet loss, got %s", advice.Verdict)
	}

	// 4. Comparable paths (33ms direct vs 32ms relay) -> VerdictComparable
	advice = CalculateRouteAdvice(33.0, 0.0, relay)
	if advice.Verdict != VerdictComparable {
		t.Errorf("expected VerdictComparable for 33ms vs 32ms, got %s", advice.Verdict)
	}

	// 5. Relay unreachable -> VerdictDirect
	unreachableRelay := &ProbeResult{Reachable: false}
	advice = CalculateRouteAdvice(40.0, 0.0, unreachableRelay)
	if advice.Verdict != VerdictDirect {
		t.Errorf("expected VerdictDirect when relay is unreachable, got %s", advice.Verdict)
	}
}

func TestProbeHistoryAccumulates(t *testing.T) {
	psk := []byte("test-psk-prober-secret-123")
	_, endpoint := startMockRelayUDPServer(t, psk, 5*time.Millisecond)

	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 2,
		SampleTimeout:  1 * time.Second,
		SampleInterval: 5 * time.Millisecond,
	})
	relay := profiles.RelayEndpoint{
		ID:       "history-relay",
		Name:     "History Relay",
		Endpoint: endpoint,
		PSK:      string(psk),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	first := prober.ProbeEndpoint(ctx, relay, 2)
	if first.HistorySamples != 1 {
		t.Fatalf("expected 1 history sample after first probe, got %d", first.HistorySamples)
	}
	if first.RTTAvgMs <= 0 {
		t.Fatalf("expected a positive average RTT, got %f", first.RTTAvgMs)
	}
	if len(first.RTTTrend) != 1 {
		t.Fatalf("expected 1 trend point, got %d", len(first.RTTTrend))
	}
	if first.LastProbeAt == "" {
		t.Fatalf("expected lastProbeAt to be set")
	}

	second := prober.ProbeEndpoint(ctx, relay, 2)
	if second.HistorySamples != 2 {
		t.Fatalf("expected 2 history samples after second probe, got %d", second.HistorySamples)
	}
	if len(second.RTTTrend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(second.RTTTrend))
	}
	if second.RTTStdMs < 0 {
		t.Fatalf("expected a non-negative RTT standard deviation, got %f", second.RTTStdMs)
	}
}

func TestProbeHistorySkipsMissingPSK(t *testing.T) {
	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 1,
		SampleTimeout:  50 * time.Millisecond,
		SampleInterval: 5 * time.Millisecond,
	})
	result := prober.ProbeEndpoint(context.Background(), profiles.RelayEndpoint{
		ID:       "no-psk",
		Endpoint: "127.0.0.1:59997",
	}, 1)
	if result.HistorySamples != 0 {
		t.Fatalf("a missing PSK must not be recorded as a network sample, got %d", result.HistorySamples)
	}
}

func TestProbeEndpointWithAnchorMeasuresRelaySegment(t *testing.T) {
	psk := []byte("test-psk-prober-secret-123")
	_, endpoint := startMockRelayUDPServer(t, psk, 5*time.Millisecond)

	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 2,
		SampleTimeout:  1 * time.Second,
		SampleInterval: 5 * time.Millisecond,
	})
	relay := profiles.RelayEndpoint{
		ID:       "anchor-relay",
		Endpoint: endpoint,
		PSK:      string(psk),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := prober.ProbeEndpointWithAnchor(ctx, relay, 2, 1)
	if !result.Reachable {
		t.Fatalf("expected a reachable relay, got error %q", result.Error)
	}
	if result.RelayAnchorRTTMs <= 0 {
		t.Fatalf("expected a relay anchor RTT, got %f", result.RelayAnchorRTTMs)
	}
	if result.TotalEstimateMs < result.RTTMedianMs {
		t.Fatalf("total estimate %f must include the client->relay RTT %f", result.TotalEstimateMs, result.RTTMedianMs)
	}

	// A plain probe must not carry anchor values.
	plain := prober.ProbeEndpoint(ctx, relay, 2)
	if plain.RelayAnchorRTTMs != 0 || plain.TotalEstimateMs != 0 {
		t.Fatalf("plain probe must not report anchor values: %+v", plain)
	}
}

func TestProbeEndpointWithoutAnchorReplyDegradesGracefully(t *testing.T) {
	psk := []byte("test-psk-prober-secret-123")
	_, endpoint := startMockRelayUDPServerMode(t, psk, 5*time.Millisecond, false)

	prober := NewProberWithConfig(ProberConfig{
		DefaultSamples: 1,
		SampleTimeout:  1 * time.Second,
		SampleInterval: 5 * time.Millisecond,
	})
	relay := profiles.RelayEndpoint{
		ID:       "silent-anchor-relay",
		Endpoint: endpoint,
		PSK:      string(psk),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	result := prober.ProbeEndpointWithAnchor(ctx, relay, 1, 1)
	if !result.Reachable {
		t.Fatalf("relay should stay reachable without an anchor reply, got error %q", result.Error)
	}
	if result.RelayAnchorRTTMs != 0 || result.TotalEstimateMs != 0 {
		t.Fatalf("no anchor reply must mean no anchor values: %+v", result)
	}
}

func TestUnstableRelayMarking(t *testing.T) {
	now := time.Now()

	lossy := NewProber()
	lossy.history = map[string][]probeHistoryEntry{
		"relay:1": {
			{RTTMs: 30, Loss: 8, Reached: true, At: now},
			{RTTMs: 31, Loss: 10, Reached: true, At: now},
			{RTTMs: 30.5, Loss: 9, Reached: true, At: now},
		},
	}
	lossyResult := ProbeResult{Endpoint: "relay:1", Reachable: true, RTTMedianMs: 30.5, PacketLoss: 9}
	lossy.finishProbe(&lossyResult)
	if !lossyResult.Unstable {
		t.Fatalf("a relay with >=5%% average loss must be marked unstable: %+v", lossyResult)
	}

	steady := NewProber()
	steady.history = map[string][]probeHistoryEntry{
		"relay:2": {
			{RTTMs: 30, Loss: 0, Reached: true, At: now},
			{RTTMs: 30.4, Loss: 0, Reached: true, At: now},
			{RTTMs: 30.2, Loss: 0, Reached: true, At: now},
		},
	}
	steadyResult := ProbeResult{Endpoint: "relay:2", Reachable: true, RTTMedianMs: 30.2, PacketLoss: 0}
	steady.finishProbe(&steadyResult)
	if steadyResult.Unstable {
		t.Fatalf("a steady relay must not be marked unstable: %+v", steadyResult)
	}

	fresh := NewProber()
	firstRun := ProbeResult{Endpoint: "relay:3", Reachable: true, RTTMedianMs: 40, PacketLoss: 0}
	fresh.finishProbe(&firstRun)
	if firstRun.Unstable {
		t.Fatalf("one sample must not be enough to call a relay unstable: %+v", firstRun)
	}
}

func TestRankProbeResultsPrefersStable(t *testing.T) {
	results := []ProbeResult{
		{RelayID: "fast-but-flapping", Reachable: true, Score: 10, RTTMedianMs: 10, Unstable: true},
		{RelayID: "steady-slower", Reachable: true, Score: 30, RTTMedianMs: 30},
		{RelayID: "down", Reachable: false, Score: 999999},
	}

	rankProbeResults(results)

	if results[0].RelayID != "steady-slower" || !results[0].IsOptimal {
		t.Fatalf("a steady relay must rank first and be optimal: %+v", results)
	}
	if results[2].RelayID != "down" || results[2].IsOptimal {
		t.Fatalf("unreachable relays must stay last and never optimal: %+v", results)
	}
}

func TestRankProbeResultsUsesTotalsWhenMeasured(t *testing.T) {
	results := []ProbeResult{
		{RelayID: "near-but-slow-relay", Reachable: true, Score: 5, RTTMedianMs: 5, TotalEstimateMs: 120},
		{RelayID: "far-but-fast-region", Reachable: true, Score: 40, RTTMedianMs: 40, TotalEstimateMs: 60},
	}

	rankProbeResults(results)

	if results[0].RelayID != "far-but-fast-region" || !results[0].IsOptimal {
		t.Fatalf("per-segment totals must decide the ranking when measured: %+v", results)
	}
}
