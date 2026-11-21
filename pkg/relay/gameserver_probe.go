package relay

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The relay-hosted game-server probe (ExitLag's "PingToDS" equivalent): the
// dashboard asks the relay itself to measure the TCP handshake RTT to a game
// server, so the Boost page can show the relay->server segment instead of
// only the client->relay half of the path.
//
// The endpoint is open (no PSK) because it performs the same measurement the
// client's own TCP ping tool performs locally and reveals nothing about the
// relay; it is rate limited per source IP and refuses non-public targets.

const (
	// probeGameServerWindow / probeGameServerLimit bound how often one IP may
	// ask the relay to dial a game server.
	probeGameServerWindow = 2 * time.Second
	probeGameServerLimit  = 1

	// probeGameServerSamples is the number of TCP handshakes per request.
	probeGameServerSamples = 3
	// probeGameServerTimeout bounds each individual handshake.
	probeGameServerTimeout = 1500 * time.Millisecond

	// probeRateEntriesCap bounds the probe limiter's own bookkeeping.
	probeRateEntriesCap = 4096

	// probeHourlyWindow / probeHourlyLimit bound sustained abuse. The 2 s
	// window already stops bursts; this one bounds a scanner to a fraction of
	// its reach per hour. A legitimate dashboard needs about one probe per
	// minute per game while boosting, so 600/hour is 10x headroom.
	probeHourlyWindow = time.Hour
	probeHourlyLimit  = 600
)

// probeDial dials one TCP target. It is a variable so tests can replace it
// and never send a packet to a real host.
var probeDial = net.DialTimeout

// probeRateEntry tracks one source IP's recent probe requests. The limiter is
// deliberately separate from the squad limiter: the two endpoints share a port
// and an audience, and a shared counter would let one starve the other.
type probeRateEntry struct {
	windowStart time.Time
	count       int
}

type probeRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*probeRateEntry
}

func (l *probeRateLimiter) allow(ip string, window time.Duration, limit int) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.entries == nil {
		l.entries = make(map[string]*probeRateEntry)
	}
	entry, ok := l.entries[ip]
	if !ok || now.Sub(entry.windowStart) >= window {
		l.entries[ip] = &probeRateEntry{windowStart: now, count: 1}
		if len(l.entries) > probeRateEntriesCap {
			l.pruneLocked(now)
		}
		return true
	}
	if entry.count >= limit {
		return false
	}
	entry.count++
	return true
}

func (l *probeRateLimiter) pruneLocked(now time.Time) {
	for ip, entry := range l.entries {
		if now.Sub(entry.windowStart) >= probeGameServerWindow {
			delete(l.entries, ip)
		}
	}
	// Fresh entries can still exceed the cap. Map order is random, so the
	// survivors are arbitrary rather than the oldest; the cap is about memory,
	// not fairness.
	for ip := range l.entries {
		if len(l.entries) <= probeRateEntriesCap {
			return
		}
		delete(l.entries, ip)
	}
}

// gameServerProbeRequest is the JSON body accepted by /probe/gameserver.
type gameServerProbeRequest struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// gameServerProbeResponse is the JSON body returned by /probe/gameserver.
type gameServerProbeResponse struct {
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	RTTMs       float64 `json:"rttMs,omitempty"`
	LossPct     float64 `json:"lossPct"`
	Samples     int     `json:"samples"`
	MeasuredBy  string  `json:"measuredBy"`
	ProbeMethod string  `json:"probeMethod"`
	Error       string  `json:"error,omitempty"`
}

// handleProbeGameServer measures the relay->game-server TCP handshake RTT.
func (s *Server) handleProbeGameServer(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
		return
	}

	var req gameServerProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.Port <= 0 || req.Port > 65535 {
		http.Error(w, `{"error":"port must be between 1 and 65535"}`, http.StatusBadRequest)
		return
	}
	ip, err := netip.ParseAddr(strings.TrimSpace(req.Host))
	if err != nil {
		http.Error(w, `{"error":"host must be an IP literal"}`, http.StatusBadRequest)
		return
	}
	if isForbiddenDestination(ip) {
		http.Error(w, `{"error":"relay probes only public game-server addresses"}`, http.StatusBadRequest)
		return
	}

	// Validation failures never consume quota: a typo must not make the next
	// correct request wait out the window.
	if !s.gameServerProbeLimiter.allow(clientIP(r), probeGameServerWindow, probeGameServerLimit) {
		http.Error(w, `{"error":"too many relay probe requests, wait 2 seconds"}`, http.StatusTooManyRequests)
		return
	}
	if !s.gameServerHourlyLimiter.allow(clientIP(r), probeHourlyWindow, probeHourlyLimit) {
		http.Error(w, `{"error":"relay probe hourly limit reached for this address"}`, http.StatusTooManyRequests)
		return
	}

	res := probeGameServer(ip, req.Port)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// probeGameServer dials the target TCP port several times and reports the
// median handshake RTT and measured loss. A closed port can still return a
// real RTT when the host answers with RST, exactly like the client's tool.
func probeGameServer(ip netip.Addr, port int) gameServerProbeResponse {
	res := gameServerProbeResponse{
		Host:        ip.String(),
		Port:        port,
		MeasuredBy:  "relay",
		ProbeMethod: "tcp-handshake",
		Samples:     probeGameServerSamples,
	}

	target := netip.AddrPortFrom(ip, uint16(port)).String()
	var rtts []float64
	failures := 0
	for i := 0; i < probeGameServerSamples; i++ {
		start := time.Now()
		conn, err := probeDial("tcp", target, probeGameServerTimeout)
		elapsed := time.Since(start)
		if err != nil {
			// A refused connection is a real network answer: the host replied
			// with RST, so the handshake RTT is valid even if no service
			// accepted the connection.
			if errors.Is(err, syscall.ECONNREFUSED) {
				rtts = append(rtts, float64(elapsed.Microseconds())/1000.0)
			} else {
				failures++
			}
		} else {
			rtts = append(rtts, float64(elapsed.Microseconds())/1000.0)
			_ = conn.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}

	res.LossPct = float64(failures) / float64(probeGameServerSamples) * 100.0
	if len(rtts) > 0 {
		sort.Float64s(rtts)
		res.RTTMs = medianRTT(rtts)
	} else if failures == probeGameServerSamples {
		res.Error = "no answer from the game server"
	}
	return res
}

func medianRTT(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return round2(sorted[n/2])
	}
	return round2((sorted[n/2-1] + sorted[n/2]) / 2.0)
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
