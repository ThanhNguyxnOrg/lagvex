package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

const (
	shareMaxBodyBytes        = 64 << 10  // 64 KiB for URI / WireGuard .conf bodies
	subscriptionMaxRespBytes = 256 << 10 // 256 KiB cap for community relay lists
	subscriptionMaxRelays    = 64
)

// handleRelayImportURI parses a lagvex:// share URI and saves the relay locally.
func (u *UIServer) handleRelayImportURI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, shareMaxBodyBytes)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	relay, err := profiles.ParseRelayShareURI(req.URI)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	u.saveRelay(relay)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"relay":   relay.Redacted(),
	})
}

// handleRelayExportURI encodes a configured native Lagvex relay into a
// copyable lagvex:// share URI. POST-only so the X-Lagvex-Request CSRF guard
// protects relay PSKs from cross-site reads.
func (u *UIServer) handleRelayExportURI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ID       string `json:"id"`
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if u.profileMgr == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "profile catalog unavailable")
		return
	}
	var found *profiles.RelayEndpoint
	for _, rel := range u.profileMgr.Catalog().Relays {
		if (req.ID != "" && strings.EqualFold(rel.ID, req.ID)) ||
			(req.Endpoint != "" && strings.EqualFold(rel.Endpoint, req.Endpoint)) {
			c := rel
			found = &c
			break
		}
	}
	if found == nil {
		writeJSONError(w, http.StatusNotFound, "relay not found")
		return
	}

	uri, err := profiles.EncodeRelayShareURI(*found)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"id":      found.ID,
		"name":    found.Name,
		"uri":     uri,
	})
}

// handleRelayImportWireGuard parses a WireGuard / Cloudflare WARP .conf INI
// file and adds it as a split-tunnel transport endpoint.
func (u *UIServer) handleRelayImportWireGuard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Config    string `json:"config"`
		Name      string `json:"name"`
		Location  string `json:"location"`
		Continent string `json:"continent"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, shareMaxBodyBytes)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	relay, err := profiles.ParseWireGuardConfig(req.Config, req.Name, req.Location, req.Continent)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	u.saveRelay(relay)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"relay":   relay.Redacted(),
	})
}

// handleRelayRegisterWARP registers a free Cloudflare WARP WireGuard profile
// and stores it locally so users can boost games at zero VPS cost.
func (u *UIServer) handleRelayRegisterWARP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	relay, err := RegisterFreeWARP(ctx)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	u.saveRelay(relay)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"relay":   relay.Redacted(),
	})
}

// handleRelayImportSubscription fetches a user-provided HTTPS URL returning a
// JSON array of relays (or newline-separated lagvex:// URIs) and adds valid
// entries to the local catalog. Private/loopback hosts are rejected to prevent
// SSRF against local services.
func (u *UIServer) handleRelayImportSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	parsed, err := validateSubscriptionURL(req.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("fetch subscription: %v", err))
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("subscription returned HTTP %d", resp.StatusCode))
		return
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, subscriptionMaxRespBytes))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("read subscription: %v", err))
		return
	}

	imported, skipped := parseSubscriptionPayload(raw)
	if len(imported) == 0 {
		writeJSONError(w, http.StatusBadRequest, "no valid relays found in subscription payload")
		return
	}

	redacted := make([]profiles.RelayEndpoint, 0, len(imported))
	for _, rel := range imported {
		u.saveRelay(rel)
		redacted = append(redacted, rel.Redacted())
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"imported": len(redacted),
		"skipped":  skipped,
		"relays":   redacted,
	})
}

// handleRelayDelete removes a configured relay node by ID.
func (u *UIServer) handleRelayDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	if req.ID == "" {
		writeJSONError(w, http.StatusBadRequest, "relay id is required")
		return
	}

	removed := false
	if u.profileMgr != nil {
		removed = u.profileMgr.RemoveRelay(req.ID)
	}
	if u.configStore != nil && u.configStore.RemoveCustomRelay(req.ID) {
		removed = true
	}
	if !removed {
		writeJSONError(w, http.StatusNotFound, "relay not found")
		return
	}

	u.invalidateSweepCache()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"id":      req.ID,
	})
}

func (u *UIServer) saveRelay(relay profiles.RelayEndpoint) {
	if u.profileMgr != nil {
		_ = u.profileMgr.AddRelay(relay)
	}
	if u.configStore != nil {
		u.configStore.AddCustomRelay(relay)
	}
	u.invalidateSweepCache()
}

func (u *UIServer) invalidateSweepCache() {
	u.sweepMu.Lock()
	u.sweepKey = ""
	u.sweepResults = nil
	u.sweepAt = time.Time{}
	u.sweepMu.Unlock()
}

// allowLocalSubscriptionForTests lets unit tests point subscriptions at httptest servers.
var allowLocalSubscriptionForTests bool

func validateSubscriptionURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" && !(allowLocalSubscriptionForTests && u.Scheme == "http") {
		return nil, fmt.Errorf("subscription URL must use https://")
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("subscription URL must include a host")
	}
	if allowLocalSubscriptionForTests {
		return u, nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".local") {
		return nil, fmt.Errorf("subscription URL cannot target loopback or local hosts")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if !isPublicIPv4OrV6(ip) {
			return nil, fmt.Errorf("subscription URL cannot target private or loopback addresses")
		}
		return u, nil
	}
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, fmt.Errorf("resolve subscription host: %w", err)
	}
	for _, a := range addrs {
		if ip, err := netip.ParseAddr(a); err == nil && !isPublicIPv4OrV6(ip) {
			return nil, fmt.Errorf("subscription host resolves to a private or loopback address")
		}
	}
	return u, nil
}

func isPublicIPv4OrV6(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast()
}

func parseSubscriptionPayload(raw []byte) (imported []profiles.RelayEndpoint, skipped int) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, 0
	}

	// 1. JSON array of RelayEndpoint (or object with {"relays": [...]})
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		var candidates []profiles.RelayEndpoint
		if strings.HasPrefix(trimmed, "[") {
			_ = json.Unmarshal([]byte(trimmed), &candidates)
		} else {
			var wrapper struct {
				Relays []profiles.RelayEndpoint `json:"relays"`
			}
			_ = json.Unmarshal([]byte(trimmed), &wrapper)
			candidates = wrapper.Relays
		}
		for _, c := range candidates {
			if len(imported) >= subscriptionMaxRelays {
				skipped++
				continue
			}
			if c.IsWireGuard() {
				if err := profiles.ValidateWireGuardConfig(c.WireGuard, c.Endpoint); err != nil {
					skipped++
					continue
				}
				c.Transport = "wireguard"
			} else {
				if strings.TrimSpace(c.Endpoint) == "" || strings.TrimSpace(c.PSK) == "" {
					skipped++
					continue
				}
				if _, _, err := net.SplitHostPort(strings.TrimSpace(c.Endpoint)); err != nil {
					skipped++
					continue
				}
				c.Transport = "lagvex"
			}
			if strings.TrimSpace(c.ID) == "" {
				c.ID = fmt.Sprintf("sub-%d", len(imported)+1)
			}
			if strings.TrimSpace(c.Name) == "" {
				c.Name = c.Endpoint
			}
			imported = append(imported, c)
		}
		return imported, skipped
	}

	// 2. Plain text list of lagvex:// URIs (one per line)
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(imported) >= subscriptionMaxRelays {
			skipped++
			continue
		}
		rel, err := profiles.ParseRelayShareURI(line)
		if err != nil {
			skipped++
			continue
		}
		imported = append(imported, rel)
	}
	return imported, skipped
}
