package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// handleTraceroute runs a PingPlotter-style hop-by-hop traceroute against a
// validated public IPv4 literal (no hostnames, private or bogon targets).
func (u *UIServer) handleTraceroute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var req struct {
		Host    string `json:"host"`
		MaxHops int    `json:"maxHops"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ip, err := netip.ParseAddr(strings.TrimSpace(req.Host))
	if err != nil || !ip.Is4() {
		writeJSONError(w, http.StatusBadRequest, "host must be a valid IPv4 literal")
		return
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		writeJSONError(w, http.StatusBadRequest, "traceroute only accepts public IPv4 addresses")
		return
	}
	oct := ip.As4()
	if oct[0] == 100 && oct[1] >= 64 && oct[1] <= 127 {
		writeJSONError(w, http.StatusBadRequest, "traceroute refuses carrier-grade NAT addresses")
		return
	}

	if wait, err := u.tcpPingLimiter.allow(); err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(wait)))
		writeJSONError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	maxHops := req.MaxHops
	if maxHops <= 0 {
		maxHops = 18
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result := runTraceroute(ctx, ip, maxHops)
	_ = json.NewEncoder(w).Encode(result)
}

// handleDNSBenchmark measures real UDP lookup latency to major public DNS
// resolvers (100% free local optimization, requires no relay or VPS).
func (u *UIServer) handleDNSBenchmark(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	if wait, err := u.tcpPingLimiter.allow(); err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(wait)))
		writeJSONError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()

	results := RunDNSBenchmark(ctx, nil, 3)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"measuredAt": time.Now().UTC().Format(time.RFC3339),
		"resolvers":  results,
	})
}

// handleDiagnoseMTU runs binary-search Path-MTU Discovery (PMTUD) against a
// validated public IPv4 address (default 1.1.1.1) and returns optimal Lagvex
// and WireGuard tunnel MTUs.
func (u *UIServer) handleDiagnoseMTU(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var req struct {
		Host string `json:"host"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req)
	target := strings.TrimSpace(req.Host)
	if target == "" {
		target = "1.1.1.1"
	}

	ip, err := validateTCPPingTarget(target)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	addr, ok := netip.AddrFromSlice(ip.To4())
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "MTU discovery requires an IPv4 address")
		return
	}

	if wait, err := u.tcpPingLimiter.allow(); err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(wait)))
		writeJSONError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	res := discoverPathMTU(ctx, addr)
	_ = json.NewEncoder(w).Encode(res)
}

// handleObfuscationToggle enables or disables Anti-DPI UDP obfuscation for
// native Lagvex relay sessions.
func (u *UIServer) handleObfuscationToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	if u.engine == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "engine is not running")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	u.engine.SetObfuscation(req.Enabled)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"obfuscated": u.engine.IsObfuscationEnabled(),
	})
}
