package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

// ProbeResult holds latency, jitter, loss, and composite score metrics for a relay node.
type ProbeResult struct {
	RelayID     string  `json:"relayId"`
	Name        string  `json:"name"`
	Endpoint    string  `json:"endpoint"`
	Continent   string  `json:"continent"`
	Location    string  `json:"location"`
	Tier        string  `json:"tier"`
	RTTMedianMs float64 `json:"rttMedianMs"`
	JitterMs    float64 `json:"jitterMs"`
	PacketLoss  float64 `json:"packetLoss"` // Percentage: 0.0 to 100.0%
	Score       float64 `json:"score"`      // Composite score: lower is better
	Reachable   bool    `json:"reachable"`
	IsOptimal   bool    `json:"isOptimal"`
	Samples     int     `json:"samples"`
	Error       string  `json:"error,omitempty"`
}

// ProberConfig defines configuration options for probing relay health and latency.
type ProberConfig struct {
	DefaultSamples     int
	SampleTimeout      time.Duration
	SampleInterval     time.Duration
	DefaultConcurrency int
}

// DefaultProberConfig returns recommended defaults for fast, non-blocking sweeps.
func DefaultProberConfig() ProberConfig {
	return ProberConfig{
		DefaultSamples:     3,
		SampleTimeout:      800 * time.Millisecond,
		SampleInterval:     40 * time.Millisecond,
		DefaultConcurrency: 8,
	}
}

// Prober evaluates relay endpoints via active UDP handshakes and AEAD ping exchanges.
type Prober struct {
	cfg ProberConfig
}

// NewProber initializes a Prober with default configuration.
func NewProber() *Prober {
	return NewProberWithConfig(DefaultProberConfig())
}

// NewProberWithConfig initializes a Prober with custom tuning.
func NewProberWithConfig(cfg ProberConfig) *Prober {
	if cfg.DefaultSamples <= 0 {
		cfg.DefaultSamples = 3
	}
	if cfg.SampleTimeout <= 0 {
		cfg.SampleTimeout = 800 * time.Millisecond
	}
	if cfg.SampleInterval <= 0 {
		cfg.SampleInterval = 40 * time.Millisecond
	}
	if cfg.DefaultConcurrency <= 0 {
		cfg.DefaultConcurrency = 8
	}
	return &Prober{cfg: cfg}
}

// ProbeEndpoint tests an individual relay node and returns detailed telemetry.
func (p *Prober) ProbeEndpoint(ctx context.Context, relay profiles.RelayEndpoint, samples int) ProbeResult {
	if samples <= 0 {
		samples = p.cfg.DefaultSamples
	}

	result := ProbeResult{
		RelayID:   relay.ID,
		Name:      relay.Name,
		Endpoint:  relay.Endpoint,
		Continent: relay.Continent,
		Location:  relay.Location,
		Tier:      relay.Tier,
		Samples:   samples,
		Score:     999999.0,
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", relay.Endpoint)
	if err != nil {
		result.Error = fmt.Sprintf("dns resolve failed: %v", err)
		return result
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		result.Error = fmt.Sprintf("udp listen failed: %v", err)
		return result
	}
	defer conn.Close()

	psk := []byte(relay.PSK)
	if len(psk) == 0 {
		psk = []byte("lagvex-community-us-free-public-psk-2026")
	}

	var rtts []float64

	// Sample 1: Handshake (validates PSK, protocol version, and measures connect RTT)
	nonce, _ := protocol.RandomUint64()
	clientID, _ := protocol.RandomUint64()
	handshakeReq := protocol.HandshakeRequest{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		ClientID:  clientID,
	}
	reqBuf := protocol.EncodeHandshakeRequest(psk, handshakeReq)

	deadline := time.Now().Add(p.cfg.SampleTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	start := time.Now()
	if _, err := conn.WriteToUDP(reqBuf, udpAddr); err != nil {
		result.Error = fmt.Sprintf("udp write failed: %v", err)
		return result
	}

	recvBuf := make([]byte, protocol.MaxPacketSize)
	n, _, err := conn.ReadFromUDP(recvBuf)
	if err != nil {
		result.Error = "timeout reaching relay"
		result.PacketLoss = 100.0
		return result
	}

	hsRTT := float64(time.Since(start).Microseconds()) / 1000.0
	resp, err := protocol.DecodeHandshakeResponse(psk, recvBuf[:n], nonce)
	if err != nil {
		result.Error = "invalid handshake response (bad PSK?)"
		result.PacketLoss = 100.0
		return result
	}
	if resp.Status != protocol.StatusOK {
		result.Error = fmt.Sprintf("relay status: %d", resp.Status)
		result.PacketLoss = 100.0
		return result
	}

	rtts = append(rtts, hsRTT)

	// Derive AEAD crypto for remaining Ping samples and clean session release
	sessCrypto, err := protocol.NewClientCrypto(psk, nonce, resp.SessionID, clientID)
	if err == nil && sessCrypto != nil {
		defer func() {
			// Clean teardown: Release allocated IP from relay's pool
			discBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
			disc := sessCrypto.EncodeDisconnect(discBuf, resp.SessionID)
			_ = conn.SetWriteDeadline(time.Now().Add(150 * time.Millisecond))
			_, _ = conn.WriteToUDP(disc, udpAddr)
		}()

		// Remaining samples: Ping / Pong RTT over established AEAD session
		pingBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
		for s := 1; s < samples; s++ {
			select {
			case <-ctx.Done():
				break
			case <-time.After(p.cfg.SampleInterval):
			}

			pingTs := uint64(time.Now().UnixNano())
			pingPacket := sessCrypto.EncodePing(pingBuf, resp.SessionID, pingTs)

			sDeadline := time.Now().Add(p.cfg.SampleTimeout)
			if d, ok := ctx.Deadline(); ok && d.Before(sDeadline) {
				sDeadline = d
			}
			_ = conn.SetDeadline(sDeadline)

			pStart := time.Now()
			if _, err := conn.WriteToUDP(pingPacket, udpAddr); err != nil {
				continue // lost packet
			}

			pn, _, err := conn.ReadFromUDP(recvBuf)
			if err != nil {
				continue // lost packet
			}

			plainBuf := make([]byte, protocol.MaxPacketSize)
			mtype, payload, err := sessCrypto.OpenPacket(plainBuf, recvBuf[:pn])
			if err != nil || mtype != protocol.TypePong {
				continue // corrupted or wrong type
			}

			echoTs, err := protocol.DecodeControlPayload(payload)
			if err != nil || echoTs != pingTs {
				continue
			}

			pRtt := float64(time.Since(pStart).Microseconds()) / 1000.0
			rtts = append(rtts, pRtt)
		}
	}

	// Calculate loss
	lost := samples - len(rtts)
	lossPct := (float64(lost) / float64(samples)) * 100.0
	result.PacketLoss = roundTo(lossPct, 1)

	if len(rtts) == 0 {
		result.Reachable = false
		result.Score = 999999.0
		return result
	}

	result.Reachable = true
	result.RTTMedianMs = roundTo(calculateMedian(rtts), 2)
	result.JitterMs = roundTo(calculateJitter(rtts), 2)

	// Composite score: Score = RTT_median + (2 * Jitter) + (PacketLoss * 50)
	result.Score = roundTo(result.RTTMedianMs+(2.0*result.JitterMs)+(result.PacketLoss*50.0), 2)

	return result
}

// ProbeAll concurrently probes all provided relays and returns sorted results.
// Reachable relays are sorted with lowest composite score first.
// The top reachable relay is marked with IsOptimal = true.
func (p *Prober) ProbeAll(ctx context.Context, relays []profiles.RelayEndpoint, concurrency int) []ProbeResult {
	if len(relays) == 0 {
		return nil
	}
	if concurrency <= 0 {
		concurrency = p.cfg.DefaultConcurrency
	}
	if concurrency > len(relays) {
		concurrency = len(relays)
	}

	results := make([]ProbeResult, len(relays))
	type probeTask struct {
		index int
		relay profiles.RelayEndpoint
	}

	taskCh := make(chan probeTask, len(relays))
	for i, r := range relays {
		taskCh <- probeTask{index: i, relay: r}
	}
	close(taskCh)

	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					results[task.index] = ProbeResult{
						RelayID:   task.relay.ID,
						Name:      task.relay.Name,
						Endpoint:  task.relay.Endpoint,
						Continent: task.relay.Continent,
						Location:  task.relay.Location,
						Tier:      task.relay.Tier,
						Reachable: false,
						Score:     999999.0,
						Error:     "context cancelled",
					}
				default:
					results[task.index] = p.ProbeEndpoint(ctx, task.relay, p.cfg.DefaultSamples)
				}
			}
		}()
	}

	wg.Wait()

	// Sort: Reachable first sorted by Score ascending, then unreachable
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Reachable != results[j].Reachable {
			return results[i].Reachable
		}
		return results[i].Score < results[j].Score
	})

	// Mark the best reachable node as optimal
	for i := range results {
		if results[i].Reachable {
			results[i].IsOptimal = true
			break
		}
	}

	return results
}

// SelectBest evaluates the given candidate relays and returns the optimal node.
func (p *Prober) SelectBest(ctx context.Context, relays []profiles.RelayEndpoint) (*ProbeResult, error) {
	if len(relays) == 0 {
		return nil, errors.New("no relays provided")
	}

	results := p.ProbeAll(ctx, relays, p.cfg.DefaultConcurrency)
	for _, res := range results {
		if res.Reachable && res.IsOptimal {
			best := res
			return &best, nil
		}
	}

	return nil, errors.New("no reachable relay found")
}

// FilterRelaysByContinent filters endpoints by continent name (case-insensitive).
func FilterRelaysByContinent(relays []profiles.RelayEndpoint, continent string) []profiles.RelayEndpoint {
	cLower := strings.TrimSpace(strings.ToLower(continent))
	if cLower == "" || cLower == "all" || cLower == "global" {
		return relays
	}

	var filtered []profiles.RelayEndpoint
	for _, r := range relays {
		if strings.ToLower(r.Continent) == cLower || strings.Contains(strings.ToLower(r.Name), cLower) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// calculateMedian returns the median value of a float slice.
func calculateMedian(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2.0
}

// calculateJitter computes the mean absolute difference of consecutive RTTs.
func calculateJitter(vals []float64) float64 {
	if len(vals) <= 1 {
		return 0
	}
	var diffSum float64
	for i := 1; i < len(vals); i++ {
		diffSum += math.Abs(vals[i] - vals[i-1])
	}
	return diffSum / float64(len(vals)-1)
}

// roundTo rounds a float64 to the specified number of decimal places.
func roundTo(val float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(val*factor) / factor
}
