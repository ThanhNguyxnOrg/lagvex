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

	// Retained measurements from earlier probe runs for the same endpoint. They
	// let the dashboard show an average, spread and trend that a single sweep
	// cannot provide. All values are measured locally by this client.
	RTTAvgMs       float64   `json:"rttAvgMs,omitempty"`
	RTTStdMs       float64   `json:"rttStdMs,omitempty"`
	LossAvgPct     float64   `json:"lossAvgPct,omitempty"`
	RTTTrend       []float64 `json:"rttTrend,omitempty"`
	HistorySamples int       `json:"historySamples,omitempty"`
	LastProbeAt    string    `json:"lastProbeAt,omitempty"`

	// Per-segment anchor measurement: the relay's own RTT to one allowlisted
	// public anchor (relay -> region), and the client->relay + relay->anchor
	// estimate used to score per-segment paths.
	RelayAnchorRTTMs float64 `json:"relayAnchorRttMs,omitempty"`
	TotalEstimateMs  float64 `json:"totalEstimateMs,omitempty"`

	// Recent marks a relay the user connected through recently (set by the UI
	// layer from the local config; the prober itself never sets it).
	Recent bool `json:"recent,omitempty"`

	// Optional ICMP echo measurement to the relay host, when the OS allows
	// unprivileged ICMP. Absent means "not measured", never zero.
	ICMPRTTMs   float64 `json:"icmpRttMs,omitempty"`
	ICMPLossPct float64 `json:"icmpLossPct,omitempty"`

	// ProbeMethod names the measurement that produced the primary RTT
	// ("tunnel-udp" = authenticated UDP handshake + ping/pong, the same path
	// game traffic takes). The dashboard labels values with it so a number is
	// never mistaken for a different measurement technique.
	ProbeMethod string `json:"probeMethod,omitempty"`

	// SourceAddress is the local address the probe socket was bound to when
	// the routed source could be pinned (multi-homed machines). Absent means
	// the operating system chose the source address.
	SourceAddress string `json:"sourceAddress,omitempty"`

	// Unstable is derived from retained history: high average loss or a large
	// RTT spread across at least three probe runs.
	Unstable bool `json:"unstable,omitempty"`
}

// anchorProbeTimeout bounds the wait for a relay anchor-probe reply during a
// sweep. The relay itself dials two TCP samples with 1.5s timeouts each.
const anchorProbeTimeout = 3500 * time.Millisecond

// Probe method labels: they name the technique behind an RTT value so the
// dashboard never implies a number was measured a different way.
const (
	ProbeMethodTunnelUDP  = "tunnel-udp"
	ProbeMethodUDPEcho    = "udp-echo"
	ProbeMethodICMPEcho   = "icmp-echo"
	ProbeMethodTCPConnect = "tcp-connect"
)

// probeSampleTimeoutCeil caps the adaptive probe timeout so a relay that has
// gone dark cannot stall a whole sweep.
const probeSampleTimeoutCeil = 3 * time.Second

// routedLocalUDPAddr returns the local UDP address the OS routes to target.
// The socket is only connected, never written to, so no packet leaves the
// machine. Both shipped competitors bind their probe sockets this way: GearUP
// picks the source with GetBestInterface/GetIpAddrTable, ExitLag keeps a
// bridge->local-endpoint cache — both so a measurement leaves through the same
// interface the real traffic would use on multi-homed machines.
func routedLocalUDPAddr(target *net.UDPAddr) (*net.UDPAddr, error) {
	conn, err := net.DialUDP("udp4", nil, target)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("unexpected local address type %T", conn.LocalAddr())
	}
	return local, nil
}

// adaptiveSampleTimeout widens the per-sample timeout for endpoints whose last
// measured RTT is slower than the configured base. It never shrinks below the
// configured value and never exceeds probeSampleTimeoutCeil. GearUP does the
// same with its adjust_timeout knob, so a slow but healthy relay is not scored
// as loss.
func (p *Prober) adaptiveSampleTimeout(endpoint string) time.Duration {
	base := p.cfg.SampleTimeout
	p.histMu.Lock()
	entries := p.history[endpoint]
	p.histMu.Unlock()
	if len(entries) == 0 {
		return base
	}
	last := entries[len(entries)-1]
	if !last.Reached || last.RTTMs <= 0 {
		return base
	}
	scaled := time.Duration(last.RTTMs * 3 * float64(time.Millisecond))
	if scaled <= base {
		return base
	}
	if scaled > probeSampleTimeoutCeil {
		return probeSampleTimeoutCeil
	}
	return scaled
}

// probeHistoryEntry is one completed probe run for an endpoint.
type probeHistoryEntry struct {
	RTTMs   float64
	Loss    float64
	Reached bool
	At      time.Time
}

// probeHistoryCap bounds retained probe runs per endpoint so long sessions do
// not grow without limit (60 runs x 3 samples each at default settings).
const probeHistoryCap = 60

// ProberConfig defines configuration options for probing relay health and latency.
type ProberConfig struct {
	DefaultSamples     int
	SampleTimeout      time.Duration
	SampleInterval     time.Duration
	DefaultConcurrency int

	// Optional ICMP echo measurement alongside the tunnel probe. It uses
	// unprivileged ICMP sockets where the OS allows them and is skipped
	// silently otherwise (never faked).
	ICMPProbe   bool
	ICMPSamples int
	ICMPTimeout time.Duration
}

// DefaultProberConfig returns recommended defaults for fast, non-blocking sweeps.
func DefaultProberConfig() ProberConfig {
	return ProberConfig{
		DefaultSamples:     3,
		SampleTimeout:      800 * time.Millisecond,
		SampleInterval:     40 * time.Millisecond,
		DefaultConcurrency: 8,
		ICMPProbe:          true,
		ICMPSamples:        2,
		ICMPTimeout:        700 * time.Millisecond,
	}
}

// Prober evaluates relay endpoints via active UDP handshakes and AEAD ping exchanges.
type Prober struct {
	cfg ProberConfig

	histMu  sync.Mutex
	history map[string][]probeHistoryEntry
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
	if cfg.ICMPSamples <= 0 {
		cfg.ICMPSamples = 2
	}
	if cfg.ICMPTimeout <= 0 {
		cfg.ICMPTimeout = 700 * time.Millisecond
	}
	return &Prober{cfg: cfg}
}

// ProbeEndpoint tests an individual relay node and returns detailed telemetry.
func (p *Prober) ProbeEndpoint(ctx context.Context, relay profiles.RelayEndpoint, samples int) ProbeResult {
	return p.ProbeEndpointWithAnchor(ctx, relay, samples, 0)
}

// ProbeEndpointWithAnchor also asks the relay for its own RTT to an allowlisted
// anchor (protocol.AnchorProbeTargets) when anchorID is not zero, so callers
// can score the whole client->relay->anchor segment instead of one hop.
func (p *Prober) ProbeEndpointWithAnchor(ctx context.Context, relay profiles.RelayEndpoint, samples int, anchorID byte) (result ProbeResult) {
	if samples <= 0 {
		samples = p.cfg.DefaultSamples
	}

	result = ProbeResult{
		RelayID:   relay.ID,
		Name:      relay.Name,
		Endpoint:  relay.Endpoint,
		Continent: relay.Continent,
		Location:  relay.Location,
		Tier:      relay.Tier,
		Samples:   samples,
		Score:     999999.0,
	}
	// Every completed run feeds the per-endpoint history used for average,
	// spread and trend values in the dashboard.
	defer func() { p.finishProbe(&result) }()

	udpAddr, err := net.ResolveUDPAddr("udp4", relay.Endpoint)
	if err != nil {
		result.Error = fmt.Sprintf("dns resolve failed: %v", err)
		result.PacketLoss = 100.0
		return result
	}

	// Pin the probe socket to the routed local address (see
	// routedLocalUDPAddr) and fall back to an OS-chosen source if that address
	// disappears between the route lookup and the bind.
	var bindAddr *net.UDPAddr
	if local, bindErr := routedLocalUDPAddr(udpAddr); bindErr == nil {
		bindAddr = local
	}
	conn, err := net.ListenUDP("udp4", bindAddr)
	if err != nil && bindAddr != nil {
		bindAddr = nil
		conn, err = net.ListenUDP("udp4", nil)
	}
	if err != nil {
		result.Error = fmt.Sprintf("udp listen failed: %v", err)
		return result
	}
	defer func() { _ = conn.Close() }()

	result.ProbeMethod = ProbeMethodTunnelUDP
	if bindAddr != nil && bindAddr.IP != nil {
		result.SourceAddress = bindAddr.IP.String()
	}

	if relay.PSK == "" {
		result.Error = "relay PSK is required"
		result.PacketLoss = 100.0
		return result
	}
	psk := []byte(relay.PSK)

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

	// Scale the per-sample timeout with the endpoint's last measured RTT.
	sampleTimeout := p.adaptiveSampleTimeout(relay.Endpoint)
	deadline := time.Now().Add(sampleTimeout)
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
	sampleLoop:
		for s := 1; s < samples; s++ {
			select {
			case <-ctx.Done():
				break sampleLoop
			case <-time.After(p.cfg.SampleInterval):
			}

			pingTs := uint64(time.Now().UnixNano())
			pingPacket := sessCrypto.EncodePing(pingBuf, resp.SessionID, pingTs)

			sDeadline := time.Now().Add(sampleTimeout)
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

		// Optional per-segment measurement: the relay's own RTT to an
		// allowlisted anchor. The relay rejects unknown ids, so this can never
		// be pointed at an arbitrary host.
		if anchorID != 0 {
			seq := uint32(time.Now().UnixNano())
			probeBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
			probePacket := sessCrypto.EncodeAnchorProbe(probeBuf, resp.SessionID,
				protocol.EncodeAnchorProbeRequest(seq, anchorID))

			probeDeadline := time.Now().Add(anchorProbeTimeout)
			if d, ok := ctx.Deadline(); ok && d.Before(probeDeadline) {
				probeDeadline = d
			}
			_ = conn.SetDeadline(probeDeadline)

			if _, err := conn.WriteToUDP(probePacket, udpAddr); err == nil {
			waitAnchorReply:
				for {
					an, _, err := conn.ReadFromUDP(recvBuf)
					if err != nil {
						break
					}
					plainA := make([]byte, protocol.MaxPacketSize)
					mtypeA, payloadA, err := sessCrypto.OpenPacket(plainA, recvBuf[:an])
					if err != nil || mtypeA != protocol.TypeAnchorProbeReply {
						continue
					}
					packed, err := protocol.DecodeControlPayload(payloadA)
					if err != nil {
						continue
					}
					replySeq, rttUs, _ := protocol.DecodeAnchorProbeReply(packed)
					if replySeq != seq {
						continue
					}
					if rttUs > 0 {
						result.RelayAnchorRTTMs = roundTo(float64(rttUs)/1000.0, 1)
					}
					break waitAnchorReply
				}
			}
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
	if result.RelayAnchorRTTMs > 0 {
		result.TotalEstimateMs = roundTo(result.RTTMedianMs+result.RelayAnchorRTTMs, 1)
	}

	// Optional ICMP echo to the same host, when the OS allows it. Failures are
	// simply "not measured" so the sweep never invents a value.
	if p.cfg.ICMPProbe {
		if icmpRes, err := probeICMP(ctx, udpAddr.IP, p.cfg.ICMPSamples, p.cfg.ICMPTimeout); err == nil && icmpRes.Samples > 0 {
			if icmpRes.RTTMs > 0 {
				result.ICMPRTTMs = roundTo(icmpRes.RTTMs, 1)
			}
			result.ICMPLossPct = roundTo(icmpRes.LossPct, 0)
		}
	}

	return result
}

// finishProbe records a completed probe run and fills the history-derived
// average, spread and trend fields from earlier runs to the same endpoint.
func (p *Prober) finishProbe(result *ProbeResult) {
	// A missing PSK is a configuration error, not a network measurement, so it
	// must not pollute the endpoint's loss history.
	if result.Error == "relay PSK is required" {
		return
	}

	entry := probeHistoryEntry{
		RTTMs:   result.RTTMedianMs,
		Loss:    result.PacketLoss,
		Reached: result.Reachable,
		At:      time.Now().UTC(),
	}

	p.histMu.Lock()
	if p.history == nil {
		p.history = make(map[string][]probeHistoryEntry)
	}
	key := strings.ToLower(result.Endpoint)
	entries := append(p.history[key], entry)
	if len(entries) > probeHistoryCap {
		entries = entries[len(entries)-probeHistoryCap:]
	}
	stored := make([]probeHistoryEntry, len(entries))
	copy(stored, entries)
	p.history[key] = stored
	p.histMu.Unlock()

	avgMs, stdMs, lossAvgPct, trend, lastAt := summarizeHistory(stored)
	result.HistorySamples = len(stored)
	if !lastAt.IsZero() {
		result.LastProbeAt = lastAt.Format(time.RFC3339)
	}
	if len(trend) > 0 {
		result.RTTAvgMs = avgMs
		result.RTTStdMs = stdMs
		result.RTTTrend = trend
	}
	if len(stored) > 0 {
		result.LossAvgPct = lossAvgPct
	}

	// Mark relays whose retained history shows instability, so ranking can
	// prefer a slightly slower but steady path.
	if result.HistorySamples >= 3 && result.Reachable {
		unstableLoss := result.LossAvgPct >= 5.0
		unstableSpread := result.RTTAvgMs > 0 && result.RTTStdMs >= result.RTTAvgMs*0.35
		result.Unstable = unstableLoss || unstableSpread
	}
}

// summarizeHistory computes measured statistics from retained probe runs.
func summarizeHistory(entries []probeHistoryEntry) (avgMs, stdMs, lossAvgPct float64, trend []float64, lastAt time.Time) {
	var rtts, lossPcts []float64
	trend = make([]float64, 0, len(entries))
	for _, e := range entries {
		if e.Reached && e.RTTMs > 0 {
			rtts = append(rtts, e.RTTMs)
			trend = append(trend, roundTo(e.RTTMs, 2))
		}
		lossPcts = append(lossPcts, e.Loss)
		if e.At.After(lastAt) {
			lastAt = e.At
		}
	}
	if len(rtts) > 0 {
		avgMs = roundTo(calculateMean(rtts), 2)
		stdMs = roundTo(calculateStdDev(rtts), 2)
	}
	if len(lossPcts) > 0 {
		lossAvgPct = roundTo(calculateMean(lossPcts), 1)
	}
	return avgMs, stdMs, lossAvgPct, trend, lastAt
}

// ProbeHistoryReport summarizes the measured history retained for one endpoint.
type ProbeHistoryReport struct {
	Endpoint       string    `json:"endpoint"`
	RTTAvgMs       float64   `json:"rttAvgMs,omitempty"`
	RTTStdMs       float64   `json:"rttStdMs,omitempty"`
	LossAvgPct     float64   `json:"lossAvgPct,omitempty"`
	HistorySamples int       `json:"historySamples"`
	LastProbeAt    string    `json:"lastProbeAt,omitempty"`
	RTTTrend       []float64 `json:"rttTrend,omitempty"`
}

// HistoryFor returns the retained measured history for one endpoint.
func (p *Prober) HistoryFor(endpoint string) (ProbeHistoryReport, bool) {
	key := strings.ToLower(endpoint)

	p.histMu.Lock()
	entries, ok := p.history[key]
	if !ok {
		p.histMu.Unlock()
		return ProbeHistoryReport{}, false
	}
	stored := make([]probeHistoryEntry, len(entries))
	copy(stored, entries)
	p.histMu.Unlock()

	avgMs, stdMs, lossAvgPct, trend, lastAt := summarizeHistory(stored)
	report := ProbeHistoryReport{
		Endpoint:       key,
		RTTAvgMs:       avgMs,
		RTTStdMs:       stdMs,
		LossAvgPct:     lossAvgPct,
		HistorySamples: len(stored),
		RTTTrend:       trend,
	}
	if !lastAt.IsZero() {
		report.LastProbeAt = lastAt.Format(time.RFC3339)
	}
	return report, true
}

// HistorySnapshot returns a copy of the measured history for every endpoint
// this prober has probed, without performing any new probes.
func (p *Prober) HistorySnapshot() []ProbeHistoryReport {
	p.histMu.Lock()
	defer p.histMu.Unlock()

	reports := make([]ProbeHistoryReport, 0, len(p.history))
	for endpoint, entries := range p.history {
		stored := make([]probeHistoryEntry, len(entries))
		copy(stored, entries)
		avgMs, stdMs, lossAvgPct, trend, lastAt := summarizeHistory(stored)

		report := ProbeHistoryReport{
			Endpoint:       endpoint,
			RTTAvgMs:       avgMs,
			RTTStdMs:       stdMs,
			LossAvgPct:     lossAvgPct,
			HistorySamples: len(stored),
			RTTTrend:       trend,
		}
		if !lastAt.IsZero() {
			report.LastProbeAt = lastAt.Format(time.RFC3339)
		}
		reports = append(reports, report)
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Endpoint < reports[j].Endpoint })
	return reports
}

func calculateMean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func calculateStdDev(vals []float64) float64 {
	if len(vals) <= 1 {
		return 0
	}
	mean := calculateMean(vals)
	var sq float64
	for _, v := range vals {
		d := v - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(len(vals)))
}

// ProbeAll concurrently probes all provided relays and returns sorted results.
// Reachable relays are sorted with lowest composite score first.
// The top reachable relay is marked with IsOptimal = true.
func (p *Prober) ProbeAll(ctx context.Context, relays []profiles.RelayEndpoint, concurrency int) []ProbeResult {
	return p.ProbeAllWithAnchor(ctx, relays, concurrency, 0)
}

// ProbeAllWithAnchor additionally asks every reachable relay for its own RTT to
// one allowlisted anchor and sorts by the measured client->relay->anchor total
// when those per-segment values exist.
func (p *Prober) ProbeAllWithAnchor(ctx context.Context, relays []profiles.RelayEndpoint, concurrency int, anchorID byte) []ProbeResult {
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
					results[task.index] = p.ProbeEndpointWithAnchor(ctx, task.relay, p.cfg.DefaultSamples, anchorID)
				}
			}
		}()
	}

	wg.Wait()
	rankProbeResults(results)
	return results
}

// rankProbeResults orders reachable relays (stable before unstable, then by
// the measured per-segment total when available, else the composite score) and
// marks the best reachable one as optimal.
func rankProbeResults(results []ProbeResult) {
	anyTotals := false
	for i := range results {
		if results[i].Reachable && results[i].TotalEstimateMs > 0 {
			anyTotals = true
			break
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Reachable != results[j].Reachable {
			return results[i].Reachable
		}
		// A steady path beats a flapping one even when its current sample is
		// slightly faster.
		if results[i].Unstable != results[j].Unstable {
			return !results[i].Unstable
		}
		if anyTotals {
			ti, tj := results[i].TotalEstimateMs, results[j].TotalEstimateMs
			if ti == 0 {
				ti = math.MaxFloat64
			}
			if tj == 0 {
				tj = math.MaxFloat64
			}
			if ti != tj {
				return ti < tj
			}
		}
		return results[i].Score < results[j].Score
	})

	for i := range results {
		results[i].IsOptimal = false
	}
	for i := range results {
		if results[i].Reachable {
			results[i].IsOptimal = true
			break
		}
	}
}

// SelectBest evaluates the given candidate relays and returns the optimal node.
func (p *Prober) SelectBest(ctx context.Context, relays []profiles.RelayEndpoint) (*ProbeResult, error) {
	return p.SelectBestWithAnchor(ctx, relays, 0)
}

// SelectBestWithAnchor ranks candidates by the measured client->relay->anchor
// path when an anchor id is given, falling back to the client->relay score.
func (p *Prober) SelectBestWithAnchor(ctx context.Context, relays []profiles.RelayEndpoint, anchorID byte) (*ProbeResult, error) {
	if len(relays) == 0 {
		return nil, errors.New("no relays provided")
	}

	results := p.ProbeAllWithAnchor(ctx, relays, p.cfg.DefaultConcurrency, anchorID)
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

// AdvisorVerdict represents the high-level recommendation for the user's route.
type AdvisorVerdict string

const (
	VerdictBoost      AdvisorVerdict = "BOOST_RECOMMENDED"
	VerdictDirect     AdvisorVerdict = "DIRECT_OPTIMAL"
	VerdictComparable AdvisorVerdict = "COMPARABLE"
)

// RouteAdvice encapsulates the comparative telemetry and recommendation between Direct ISP and Relay.
type RouteAdvice struct {
	Verdict          AdvisorVerdict `json:"verdict"`
	DirectRTTMs      float64        `json:"directRttMs"`
	DirectLossPct    float64        `json:"directLossPct"`
	RelayRTTMs       float64        `json:"relayRttMs"`
	RelayJitterMs    float64        `json:"relayJitterMs"`
	RelayLossPct     float64        `json:"relayLossPct"`
	OptimalRelayID   string         `json:"optimalRelayId,omitempty"`
	OptimalRelayName string         `json:"optimalRelayName,omitempty"`
	SavingsMs        float64        `json:"savingsMs"`
	Headline         string         `json:"headline"`
	Summary          string         `json:"summary"`
	Timestamp        time.Time      `json:"timestamp"`
}

// CalculateRouteAdvice evaluates direct ISP RTT against the best available relay node.
func CalculateRouteAdvice(directRTT, directLoss float64, bestRelay *ProbeResult) RouteAdvice {
	now := time.Now()

	// 1. If best relay is nil or unreachable
	if bestRelay == nil || !bestRelay.Reachable {
		return RouteAdvice{
			Verdict:       VerdictDirect,
			DirectRTTMs:   roundTo(directRTT, 1),
			DirectLossPct: roundTo(directLoss, 1),
			Headline:      "DIRECT ISP ACTIVE",
			Summary:       "No relay nodes available or selected; direct domestic connection is active.",
			Timestamp:     now,
		}
	}

	relayRTT := bestRelay.RTTMedianMs
	relayLoss := bestRelay.PacketLoss
	relayJitter := bestRelay.JitterMs

	// 2. If direct path is unreachable or has high loss
	if directRTT <= 0 || directLoss >= 15.0 {
		savings := 0.0
		if directRTT > 0 {
			savings = directRTT - relayRTT
		}
		return RouteAdvice{
			Verdict:          VerdictBoost,
			DirectRTTMs:      roundTo(directRTT, 1),
			DirectLossPct:    roundTo(directLoss, 1),
			RelayRTTMs:       roundTo(relayRTT, 1),
			RelayJitterMs:    roundTo(relayJitter, 1),
			RelayLossPct:     roundTo(relayLoss, 1),
			OptimalRelayID:   bestRelay.RelayID,
			OptimalRelayName: bestRelay.Name,
			SavingsMs:        roundTo(savings, 1),
			Headline:         "BOOST STRONGLY RECOMMENDED",
			Summary:          fmt.Sprintf("Direct path has %.0f%% loss. %s provides a stable tunnel with %.1fms latency.", directLoss, bestRelay.Name, relayRTT),
			Timestamp:        now,
		}
	}

	// 3. Direct path is significantly better (at least 5ms faster with 0% loss)
	if directRTT <= (relayRTT-5.0) && directLoss <= 0.5 {
		return RouteAdvice{
			Verdict:          VerdictDirect,
			DirectRTTMs:      roundTo(directRTT, 1),
			DirectLossPct:    roundTo(directLoss, 1),
			RelayRTTMs:       roundTo(relayRTT, 1),
			RelayJitterMs:    roundTo(relayJitter, 1),
			RelayLossPct:     roundTo(relayLoss, 1),
			OptimalRelayID:   bestRelay.RelayID,
			OptimalRelayName: bestRelay.Name,
			SavingsMs:        roundTo(relayRTT-directRTT, 1),
			Headline:         "DIRECT ISP PATH OPTIMAL",
			Summary:          fmt.Sprintf("Your direct domestic connection is currently optimal (%.1fms vs %.1fms via %s). Booster recommended: OFF.", directRTT, relayRTT, bestRelay.Name),
			Timestamp:        now,
		}
	}

	// 4. Relay path is significantly better (at least 5ms faster or cuts packet loss)
	if (relayRTT <= directRTT-5.0) || (directLoss > 2.0 && relayLoss <= 0.5) {
		savings := directRTT - relayRTT
		return RouteAdvice{
			Verdict:          VerdictBoost,
			DirectRTTMs:      roundTo(directRTT, 1),
			DirectLossPct:    roundTo(directLoss, 1),
			RelayRTTMs:       roundTo(relayRTT, 1),
			RelayJitterMs:    roundTo(relayJitter, 1),
			RelayLossPct:     roundTo(relayLoss, 1),
			OptimalRelayID:   bestRelay.RelayID,
			OptimalRelayName: bestRelay.Name,
			SavingsMs:        roundTo(savings, 1),
			Headline:         "BOOST RECOMMENDED",
			Summary:          fmt.Sprintf("%s accelerates your connection: saving ~%.1fms with %.1fms jitter and 0%% packet loss.", bestRelay.Name, savings, relayJitter),
			Timestamp:        now,
		}
	}

	// 5. Comparable paths (within +/- 5ms)
	return RouteAdvice{
		Verdict:          VerdictComparable,
		DirectRTTMs:      roundTo(directRTT, 1),
		DirectLossPct:    roundTo(directLoss, 1),
		RelayRTTMs:       roundTo(relayRTT, 1),
		RelayJitterMs:    roundTo(relayJitter, 1),
		RelayLossPct:     roundTo(relayLoss, 1),
		OptimalRelayID:   bestRelay.RelayID,
		OptimalRelayName: bestRelay.Name,
		SavingsMs:        roundTo(directRTT-relayRTT, 1),
		Headline:         "COMPARABLE LATENCY",
		Summary:          fmt.Sprintf("Direct path (%.1fms) and %s (%.1fms) provide equivalent speeds. Boost is optional.", directRTT, bestRelay.Name, relayRTT),
		Timestamp:        now,
	}
}

// ProbeDirectGateway measures round-trip time directly to a game server or public gateway without using the relay.
func (p *Prober) ProbeDirectGateway(ctx context.Context, hostPort string) (float64, float64, error) {
	if hostPort == "" {
		return 0, 0, errors.New("empty target host")
	}

	var rtts []float64
	attempts := 3
	lossCount := 0

	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		default:
		}

		start := time.Now()
		d := net.Dialer{Timeout: 800 * time.Millisecond}
		conn, err := d.DialContext(ctx, "tcp", hostPort)
		if err != nil {
			// Fallback to UDP check
			uConn, uErr := d.DialContext(ctx, "udp", hostPort)
			if uErr != nil {
				lossCount++
				continue
			}
			_ = uConn.Close()
			lossCount++
			continue
		}
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		_ = conn.Close()
		rtts = append(rtts, elapsed)
		time.Sleep(20 * time.Millisecond)
	}

	lossPct := (float64(lossCount) / float64(attempts)) * 100.0
	if len(rtts) == 0 {
		return 0, lossPct, errors.New("all direct probes failed")
	}

	medianRTT := calculateMedian(rtts)
	return roundTo(medianRTT, 1), roundTo(lossPct, 1), nil
}

// probeTCPGateway measures real microsecond round-trip TCP connect latency to a target host.
func probeTCPGateway(ctx context.Context, hostPort string, samples int) ([]float64, error) {
	if samples <= 0 {
		samples = 3
	}
	var rtts []float64
probeLoop:
	for i := 0; i < samples; i++ {
		select {
		case <-ctx.Done():
			break probeLoop
		default:
		}
		t0 := time.Now()
		d := net.Dialer{Timeout: 1200 * time.Millisecond}
		conn, err := d.DialContext(ctx, "tcp", hostPort)
		if err != nil {
			continue
		}
		dur := float64(time.Since(t0).Microseconds()) / 1000.0
		_ = conn.Close()
		rtts = append(rtts, dur)
		time.Sleep(20 * time.Millisecond)
	}
	if len(rtts) == 0 {
		return nil, errors.New("fallback gateway probe timed out")
	}
	return rtts, nil
}
