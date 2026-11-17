package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"github.com/ThanhNguyxnOrg/lagvex/web"
)

// UIServer serves the modern Gaming Booster Dashboard and REST API.
type UIServer struct {
	engine      *Engine
	profileMgr  *profiles.Manager
	webDir      string
	server      *http.Server
	prober      *Prober
	configStore *ConfigStore
}

// NewUIServer creates a new dashboard UI server.
func NewUIServer(engine *Engine, pm *profiles.Manager, webDir string) *UIServer {
	if webDir == "" {
		exe, _ := os.Executable()
		candidates := []string{
			filepath.Join(filepath.Dir(exe), "web"),
			"web",
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				webDir = c
				break
			}
		}
	}

	if pm == nil {
		pm, _ = profiles.NewManager("")
	}

	cs := NewConfigStore("")
	if pm != nil {
		cfg := cs.Get()
		for _, g := range cfg.CustomGames {
			_ = pm.AddGame(g)
		}
		for _, r := range cfg.CustomRelays {
			_ = pm.AddRelay(r)
		}
	}

	return &UIServer{
		engine:      engine,
		profileMgr:  pm,
		webDir:      webDir,
		prober:      NewProber(),
		configStore: cs,
	}
}

// setupRoutes registers all dashboard API routes and web filesystem handlers.
func (u *UIServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/status", u.handleStatus)
	mux.HandleFunc("/api/games", u.handleGames)
	mux.HandleFunc("/api/detected-processes", u.handleDetectedProcesses)
	mux.HandleFunc("/api/game-connections", u.handleGameConnections)
	mux.HandleFunc("/api/route-detail", u.handleRouteDetail)
	mux.HandleFunc("/api/peer-routing/toggle", u.handlePeerRoutingToggle)
	mux.HandleFunc("/api/full-tunnel/toggle", u.handleFullTunnelToggle)
	mux.HandleFunc("/api/relays", u.handleRelays)
	mux.HandleFunc("/api/probe-relays", u.handleProbeRelays)
	mux.HandleFunc("/api/best-relay", u.handleBestRelay)
	mux.HandleFunc("/api/connect", u.handleConnect)
	mux.HandleFunc("/api/disconnect", u.handleDisconnect)
	mux.HandleFunc("/api/test-relay", u.handleTestRelay)
	mux.HandleFunc("/api/ping-relay", u.handleTestRelay)
	mux.HandleFunc("/api/add-game", u.handleAddGame)
	mux.HandleFunc("/api/add-relay", u.handleAddRelay)
	mux.HandleFunc("/api/tweak", u.handleTweak)
	mux.HandleFunc("/api/advisor", u.handleAdvisor)
	mux.HandleFunc("/api/failover/toggle", u.handleFailoverToggle)
	mux.HandleFunc("/api/failover/history", u.handleFailoverHistory)
	mux.HandleFunc("/api/fec/toggle", u.handleFECToggle)
	mux.HandleFunc("/api/optimize-profile", u.handleOptimizeProfile)
	mux.HandleFunc("/api/system-info", u.handleSystemInfo)
	mux.HandleFunc("/api/relaunch-admin", u.handleRelaunchAdmin)
	mux.HandleFunc("/api/shutdown", u.handleShutdown)
	mux.HandleFunc("/api/diagnose-network", u.handleNetworkDiagnostics)
	mux.HandleFunc("/api/diagnose-bufferbloat", u.handleDiagnoseBufferbloat)
	mux.HandleFunc("/api/diagnostics/export", u.handleDiagnosticsExport)
	mux.HandleFunc("/api/diagnostics/import", u.handleDiagnosticsImport)
	mux.HandleFunc("/api/game-telemetry", u.handleGameTelemetry)
	mux.HandleFunc("/api/squad/create", u.handleSquadCreate)
	mux.HandleFunc("/api/squad/join", u.handleSquadJoin)
	mux.HandleFunc("/api/squad/room", u.handleSquadRoom)
	mux.HandleFunc("/api/squad/heartbeat", u.handleSquadHeartbeat)
	mux.HandleFunc("/api/squad/leave", u.handleSquadLeave)

	// Static Web Assets (disk prioritization with embedded binary fallback)
	fs := http.FileServer(web.GetFileSystem(u.webDir))
	mux.Handle("/", fs)

	return mux
}

// Start listens on the specified address (e.g. "127.0.0.1:18888").
func (u *UIServer) Start(addr string) error {
	mux := u.setupRoutes()

	u.server = &http.Server{
		Addr:         addr,
		Handler:      withLocalRequestGuard(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("[Dashboard] Web Dashboard active at http://%s", addr)
	return u.server.ListenAndServe()
}

func (u *UIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var stats TunnelStats
	if u.engine != nil {
		stats = u.engine.Stats()
	} else {
		stats = TunnelStats{State: StateDisconnected}
	}

	resp := map[string]any{
		"state":            stats.State,
		"active":           stats.State == StateConnected,
		"relayAddr":        stats.RelayAddr,
		"clientIP":         stats.ClientIP,
		"gatewayIP":        stats.GatewayIP,
		"pingMs":           stats.PingMs,
		"bytesUp":          stats.BytesUp,
		"bytesDown":        stats.BytesDown,
		"upRateBps":        stats.UpRateBps,
		"downRateBps":      stats.DownRateBps,
		"activeGame":       stats.ActiveGame,
		"activeRegion":     stats.ActiveRegion,
		"activeRelayId":    stats.ActiveRelayID,
		"activeRelayName":  stats.ActiveRelayName,
		"autoFailover":     stats.AutoFailover,
		"routeCount":       stats.RouteCount,
		"gameRunning":      stats.GameRunning,
		"fecActive":        stats.FECActive,
		"fecRatio":         stats.FECRatio,
		"packetsRecovered": stats.PacketsRecovered,
		"fecParitySent":    stats.FECParitySent,
		"fecParityRecv":    stats.FECParityRecv,
		"driverMode":       stats.DriverMode,
		"lastError":        stats.LastError,
		"smartL4Filter":    stats.SmartL4Filter,
		"packetLoss":       stats.PacketLoss,
		"packetLossPct":    stats.PacketLossPct,
		"peerRouting":      stats.PeerRouting,
		"peerRouteCount":   stats.PeerRouteCount,
		"peerRoutes":       stats.PeerRouteIPs,
		"fullTunnel":       stats.FullTunnel,
		"failovers":        stats.Failovers,
		"standbyRelayId":   stats.StandbyRelayID,
		"standbyRelayName": stats.StandbyRelayName,
		"standbyRttMs":     stats.StandbyRttMs,
		"standbyTotalMs":   stats.StandbyTotalMs,
		"standbyAgeSec":    stats.StandbyAgeSec,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (u *UIServer) handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	_ = json.NewEncoder(w).Encode(catalog.Games)
}

// handleDetectedProcesses reports which configured game processes are running
// right now, using a single local process-table scan. It is a detection aid
// only: it does not change routing and does not attach to any game process.
func (u *UIServer) handleDetectedProcesses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()

	type gameProcessStatus struct {
		GameID         string   `json:"gameId"`
		Name           string   `json:"name"`
		Processes      []string `json:"processes"`
		Running        bool     `json:"running"`
		MatchedProcess string   `json:"matchedProcess,omitempty"`
		PID            int      `json:"pid,omitempty"`
	}

	var targets []string
	for _, g := range catalog.Games {
		targets = append(targets, g.ProcessNames...)
	}
	matches := findRunningProcesses(targets)

	statuses := make([]gameProcessStatus, 0, len(catalog.Games))
	for _, g := range catalog.Games {
		status := gameProcessStatus{
			GameID:    g.ID,
			Name:      g.Name,
			Processes: g.ProcessNames,
		}
		for _, p := range g.ProcessNames {
			if pid, ok := matches[p]; ok {
				status.Running = true
				status.MatchedProcess = p
				status.PID = pid
				break
			}
		}
		statuses = append(statuses, status)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"scannedAt": time.Now().UTC().Format(time.RFC3339),
		"games":     statuses,
	})
}

func (u *UIServer) handleRelays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	// Never expose relay pre-shared keys to the browser. The engine resolves
	// them from the local catalog when a connection is requested.
	relays := make([]profiles.RelayEndpoint, len(catalog.Relays))
	copy(relays, catalog.Relays)
	for i := range relays {
		relays[i].PSK = ""
	}
	_ = json.NewEncoder(w).Encode(relays)
}

func (u *UIServer) handleProbeRelays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	relays := catalog.Relays

	continent := r.URL.Query().Get("continent")
	if continent != "" {
		relays = FilterRelaysByContinent(relays, continent)
	}

	samples := 3
	if sStr := r.URL.Query().Get("samples"); sStr != "" {
		if s, err := strconv.Atoi(sStr); err == nil && s > 0 && s <= 10 {
			samples = s
		}
	}

	anchorID := u.resolveAnchorProbeID(r.URL.Query().Get("gameId"), r.URL.Query().Get("regionId"))

	sweepTimeout := 5 * time.Second
	if anchorID != 0 {
		// Each relay also performs an allowlisted anchor probe during the sweep.
		sweepTimeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), sweepTimeout)
	defer cancel()

	prober := u.prober
	if samples != prober.cfg.DefaultSamples {
		cfg := prober.cfg
		cfg.DefaultSamples = samples
		prober = NewProberWithConfig(cfg)
	}

	results := prober.ProbeAllWithAnchor(ctx, relays, 8, anchorID)

	// Mark recently used relays from the local config so the dashboard can
	// offer the same node-history shortcut as GearUP/Mudfish.
	if u.configStore != nil {
		for _, endpoint := range u.configStore.Get().RecentRelays {
			for i := range results {
				if strings.EqualFold(results[i].Endpoint, endpoint) {
					results[i].Recent = true
				}
			}
		}
	}

	_ = json.NewEncoder(w).Encode(results)
}

func (u *UIServer) handleBestRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	relays := catalog.Relays

	continent := r.URL.Query().Get("continent")
	if continent != "" {
		relays = FilterRelaysByContinent(relays, continent)
	}

	anchorID := u.resolveAnchorProbeID(r.URL.Query().Get("gameId"), r.URL.Query().Get("regionId"))

	sweepTimeout := 5 * time.Second
	if anchorID != 0 {
		sweepTimeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), sweepTimeout)
	defer cancel()

	best, err := u.prober.SelectBestWithAnchor(ctx, relays, anchorID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(best)
}

type connectReq struct {
	RelayEndpoint string `json:"relayEndpoint"`
	RelayAddr     string `json:"relayAddr"`
	PSK           string `json:"psk"`
	GameID        string `json:"gameId"`
	RegionID      string `json:"regionId"`
	ForceNow      bool   `json:"forceNow"`
	AutoNode      bool   `json:"autoNode"`
	FullTunnel    bool   `json:"fullTunnel"`
}

func (u *UIServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req connectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.RelayEndpoint == "" && req.RelayAddr != "" {
		req.RelayEndpoint = req.RelayAddr
	}

	if req.GameID != "" {
		if req.RegionID == "" {
			writeJSONError(w, http.StatusBadRequest, "regionId is required when gameId is set")
			return
		}
		if u.profileMgr == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "game profile catalog is unavailable")
			return
		}
		if _, err := u.profileMgr.GetRegionCIDRs(req.GameID, req.RegionID); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	catalog := u.profileMgr.Catalog()

	// Auto Node Selection if requested or endpoint is "auto"
	if req.AutoNode || req.RelayEndpoint == "auto" || req.RelayEndpoint == "" {
		anchorID := u.resolveAnchorProbeID(req.GameID, req.RegionID)
		autoTimeout := 3 * time.Second
		if anchorID != 0 {
			autoTimeout = 15 * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), autoTimeout)
		defer cancel()

		best, err := u.prober.SelectBestWithAnchor(ctx, catalog.Relays, anchorID)
		if err == nil && best != nil {
			req.RelayEndpoint = best.Endpoint
			for _, rel := range catalog.Relays {
				if rel.Endpoint == best.Endpoint {
					req.PSK = rel.PSK
					break
				}
			}
			log.Printf("[Dashboard] Auto-selected optimal relay: %s (%s, score=%.1f)", best.Name, best.Endpoint, best.Score)
		} else {
			writeJSONError(w, http.StatusServiceUnavailable, "no reachable relays found in catalog")
			return
		}
	} else if req.PSK == "" {
		// Lookup PSK from catalog if not explicitly provided
		for _, rel := range catalog.Relays {
			if rel.Endpoint == req.RelayEndpoint {
				req.PSK = rel.PSK
				break
			}
		}
	}

	if req.RelayEndpoint == "" {
		writeJSONError(w, http.StatusBadRequest, "relayEndpoint is required")
		return
	}
	if req.PSK == "" {
		writeJSONError(w, http.StatusBadRequest, "relay PSK is required; add the relay with its PSK before connecting")
		return
	}

	if strings.Contains(req.RelayEndpoint, "127.0.0.1") || strings.Contains(req.RelayEndpoint, "localhost") {
		if err := EnsureLocalRelayRunning(req.RelayEndpoint, []byte(req.PSK)); err != nil {
			writeJSONError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	if u.engine == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "engine not initialized")
		return
	}

	// Connect synchronously so the dashboard only shows an active boost after
	// the relay handshake, TUN setup and route validation actually succeeded.
	// Store the full-tunnel preference first: an active session switches over
	// immediately, a fresh connect installs the half-default routes as it comes
	// up.
	u.engine.SetFullTunnel(req.FullTunnel)

	if err := u.engine.Connect(req.RelayEndpoint, []byte(req.PSK), req.GameID, req.RegionID, req.ForceNow); err != nil {
		log.Printf("[Dashboard] Connect error: %v", err)
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	// Remember the relay for the recent-node shortcut only after a real
	// successful connect.
	if u.configStore != nil {
		u.configStore.AddRecentRelay(req.RelayEndpoint)
	}

	stats := u.engine.Stats()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":     true,
		"status":      "connected",
		"state":       stats.State,
		"endpoint":    req.RelayEndpoint,
		"gameId":      stats.ActiveGame,
		"regionId":    stats.ActiveRegion,
		"routeCount":  stats.RouteCount,
		"gameRunning": stats.GameRunning,
		"pingMs":      stats.PingMs,
		"driverMode":  stats.DriverMode,
	})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"status":  "error",
		"error":   message,
	})
}

func (u *UIServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if u.engine != nil {
		go func() {
			_ = u.engine.Disconnect()
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "disconnecting"})
}

func (u *UIServer) handleTestRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var endpoint, psk string

	switch r.Method {
	case http.MethodGet:
		endpoint = r.URL.Query().Get("endpoint")
		psk = r.URL.Query().Get("psk")
	case http.MethodPost:
		var req struct {
			Endpoint string `json:"endpoint"`
			PSK      string `json:"psk"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			endpoint = req.Endpoint
			psk = req.PSK
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if endpoint == "" {
		http.Error(w, "endpoint is required", http.StatusBadRequest)
		return
	}

	catalog := u.profileMgr.Catalog()
	var relay profiles.RelayEndpoint
	found := false
	for _, rel := range catalog.Relays {
		if rel.Endpoint == endpoint {
			relay = rel
			found = true
			break
		}
	}
	if !found {
		relay = profiles.RelayEndpoint{
			ID:       "custom",
			Name:     "Custom Endpoint",
			Endpoint: endpoint,
			PSK:      psk,
		}
	} else if psk != "" {
		relay.PSK = psk
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result := u.prober.ProbeEndpoint(ctx, relay, 3)

	resp := map[string]any{
		"success":    result.Reachable,
		"reachable":  result.Reachable,
		"latencyMs":  result.RTTMedianMs,
		"rttMs":      result.RTTMedianMs,
		"jitterMs":   result.JitterMs,
		"packetLoss": result.PacketLoss,
		"score":      result.Score,
		"endpoint":   result.Endpoint,
		"name":       result.Name,
		"isOptimal":  result.IsOptimal,
	}
	if result.Error != "" {
		resp["error"] = result.Error
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (u *UIServer) handleAddGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var game profiles.GameDefinition
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if game.ID == "" || game.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "id and name are required")
		return
	}
	if len(game.Regions) == 0 {
		writeJSONError(w, http.StatusBadRequest, "at least one game region with CIDRs is required")
		return
	}
	for ri := range game.Regions {
		region := &game.Regions[ri]
		region.ID = strings.TrimSpace(region.ID)
		region.Name = strings.TrimSpace(region.Name)
		if region.ID == "" || region.Name == "" {
			writeJSONError(w, http.StatusBadRequest, "each region needs an id and name")
			return
		}
		if len(region.CIDRs) == 0 {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("region %q has no CIDRs", region.ID))
			return
		}
		valid := make([]string, 0, len(region.CIDRs))
		for _, raw := range region.CIDRs {
			cidr := strings.TrimSpace(raw)
			if cidr == "" {
				continue
			}
			if _, err := netip.ParsePrefix(cidr); err != nil {
				writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid CIDR %q", cidr))
				return
			}
			valid = append(valid, cidr)
		}
		if len(valid) == 0 {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("region %q has no valid CIDRs", region.ID))
			return
		}
		region.CIDRs = valid
	}

	if err := u.profileMgr.AddGame(game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if u.configStore != nil {
		u.configStore.AddCustomGame(game)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "game": game})
}

func (u *UIServer) handleAddRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var relay profiles.RelayEndpoint
	if err := json.NewDecoder(r.Body).Decode(&relay); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if relay.ID == "" || relay.Endpoint == "" {
		http.Error(w, "id and endpoint are required", http.StatusBadRequest)
		return
	}
	if relay.Name == "" {
		relay.Name = relay.Endpoint
	}
	if relay.PSK == "" {
		http.Error(w, "psk is required", http.StatusBadRequest)
		return
	}
	if relay.Tier == "" {
		relay.Tier = "custom"
	}

	if err := u.profileMgr.AddRelay(relay); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if u.configStore != nil {
		u.configStore.AddCustomRelay(relay)
	}

	publicRelay := relay
	publicRelay.PSK = ""
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "relay": publicRelay})
}

type tweakReq struct {
	TCPNoDelay    bool `json:"tcpNoDelay"`
	DisableNagle  bool `json:"disableNagle"`
	MMCSSPriority bool `json:"mmcssPriority"`
	MTUClamping   bool `json:"mtuClamping"`
	SmartL4Filter bool `json:"smartL4Filter"`
}

func (u *UIServer) handleTweak(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req tweakReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	applied := []string{}
	failed := []string{}
	preferences := []string{}

	switch runtime.GOOS {
	case "windows":
		// 1. Flush DNS cache
		if err := exec.Command("ipconfig", "/flushdns").Run(); err == nil {
			applied = append(applied, "DNS resolver cache flushed")
		} else {
			failed = append(failed, "DNS resolver cache flush")
		}

		// 2. TCP Window Auto-Tuning
		if err := exec.Command("netsh", "int", "tcp", "set", "global", "autotuninglevel=normal").Run(); err == nil {
			applied = append(applied, "TCP window auto-tuning set to normal")
		} else {
			failed = append(failed, "TCP window auto-tuning (requires administrator)")
		}

		// 3. MMCSS Gaming Priority & Network Throttling
		if req.MMCSSPriority {
			throttleErr := exec.Command("reg", "add", `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Multimedia\SystemProfile`, "/v", "NetworkThrottlingIndex", "/t", "REG_DWORD", "/d", "0xffffffff", "/f").Run()
			responseErr := exec.Command("reg", "add", `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Multimedia\SystemProfile`, "/v", "SystemResponsiveness", "/t", "REG_DWORD", "/d", "0", "/f").Run()
			if throttleErr == nil && responseErr == nil {
				applied = append(applied, "MMCSS network throttling disabled")
			} else {
				failed = append(failed, "MMCSS registry tuning (requires administrator)")
			}
		}
	case "darwin":
		if err := exec.Command("dscacheutil", "-flushcache").Run(); err == nil {
			applied = append(applied, "macOS DNS cache flushed")
		} else {
			failed = append(failed, "macOS DNS cache flush")
		}
		_ = exec.Command("killall", "-HUP", "mDNSResponder").Run()

		if req.TCPNoDelay || req.DisableNagle {
			if err := exec.Command("sysctl", "-w", "net.inet.tcp.delayed_ack=0").Run(); err == nil {
				applied = append(applied, "TCP delayed ACK disabled")
			} else {
				failed = append(failed, "TCP delayed ACK change (requires root)")
			}
		}
	case "linux":
		if resolvePath, err := exec.LookPath("resolvectl"); err == nil {
			if err := exec.Command(resolvePath, "flush-caches").Run(); err == nil {
				applied = append(applied, "systemd-resolved DNS cache flushed")
			} else {
				failed = append(failed, "systemd-resolved DNS cache flush")
			}
		} else if resolvePath, err := exec.LookPath("systemd-resolve"); err == nil {
			if err := exec.Command(resolvePath, "--flush-caches").Run(); err == nil {
				applied = append(applied, "systemd-resolve DNS cache flushed")
			} else {
				failed = append(failed, "systemd-resolve DNS cache flush")
			}
		}

		if req.TCPNoDelay || req.DisableNagle {
			lowLatencyErr := exec.Command("sysctl", "-w", "net.ipv4.tcp_low_latency=1").Run()
			slowStartErr := exec.Command("sysctl", "-w", "net.ipv4.tcp_slow_start_after_idle=0").Run()
			if lowLatencyErr == nil && slowStartErr == nil {
				applied = append(applied, "Linux TCP low-latency sysctls applied")
			} else {
				failed = append(failed, "Linux TCP sysctls (requires root)")
			}
		}
	}

	// TCP_NODELAY and MTU clamping are not global OS switches. Record them as
	// preferences and describe exactly where they take effect.
	if req.TCPNoDelay || req.DisableNagle {
		preferences = append(preferences, "TCP_NODELAY preference saved; it applies to TCP sockets owned by the client, not system-wide.")
	}
	if req.MTUClamping {
		preferences = append(preferences, "MTU clamping preference saved; the virtual adapter uses the relay-negotiated MTU when a tunnel connects.")
	}

	if u.engine != nil {
		u.engine.SetL4FilterEnabled(req.SmartL4Filter)
		if req.SmartL4Filter {
			applied = append(applied, "Smart L4 game-tick filter enabled (UDP prioritized)")
		} else {
			applied = append(applied, "Smart L4 game-tick filter disabled")
		}
	}

	if u.configStore != nil {
		u.configStore.SetTweaks(map[string]bool{
			"tcpNoDelay":    req.TCPNoDelay,
			"disableNagle":  req.DisableNagle,
			"mmcssPriority": req.MMCSSPriority,
			"mtuClamping":   req.MTUClamping,
			"smartL4Filter": req.SmartL4Filter,
		})
	}

	status := "No changes requested"
	switch {
	case len(applied) > 0 && len(failed) == 0:
		status = "Optimizations applied"
	case len(applied) > 0:
		status = "Partially applied"
	case len(preferences) > 0:
		status = "Preferences saved; no system commands succeeded"
	case len(failed) > 0:
		status = "No system command succeeded"
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":          true,
		"status":           status,
		"applied":          applied,
		"failed":           failed,
		"savedPreferences": preferences,
	})
}

func (u *UIServer) handleAdvisor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameID := r.URL.Query().Get("gameId")
	regionID := r.URL.Query().Get("regionId")

	if gameID == "" || regionID == "" {
		if u.engine != nil {
			stats := u.engine.Stats()
			if gameID == "" {
				gameID = stats.ActiveGame
			}
			if regionID == "" {
				regionID = stats.ActiveRegion
			}
		}
	}

	catalog := u.profileMgr.Catalog()
	if len(catalog.Relays) == 0 {
		http.Error(w, `{"error":"no relays configured in catalog"}`, http.StatusServiceUnavailable)
		return
	}

	anchorID := u.resolveAnchorProbeID(gameID, regionID)
	sweepTimeout := 3 * time.Second
	if anchorID != 0 {
		sweepTimeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), sweepTimeout)
	defer cancel()

	// 1. Probe candidate relays to find optimal relay
	bestRelay, err := u.prober.SelectBestWithAnchor(ctx, catalog.Relays, anchorID)
	if err != nil {
		for _, rel := range catalog.Relays {
			res := u.prober.ProbeEndpoint(ctx, rel, 2)
			if res.Reachable {
				bestRelay = &res
				break
			}
		}
	}
	if bestRelay == nil {
		bestRelay = &ProbeResult{
			RelayID:   "none",
			Name:      "No Relay Reachable",
			Reachable: false,
		}
	}

	// 2. Determine target direct gateway
	targetHost := "8.8.8.8:53"
	if gameID != "" {
		if game, ok := u.profileMgr.FindGameByID(gameID); ok {
			if len(game.LobbyAddresses) > 0 {
				targetHost = game.LobbyAddresses[0]
				if !strings.Contains(targetHost, ":") {
					targetHost += ":443"
				}
			} else {
				for _, reg := range game.Regions {
					if (regionID == "" || strings.EqualFold(reg.ID, regionID)) && len(reg.CIDRs) > 0 {
						if p, err := netip.ParsePrefix(reg.CIDRs[0]); err == nil {
							targetHost = net.JoinHostPort(p.Addr().Next().String(), "443")
						} else {
							prefix := strings.Split(reg.CIDRs[0], "/")[0]
							targetHost = prefix + ":443"
						}
						break
					}
				}
			}
		}
	}

	// 3. Probe Direct Gateway
	directRTT, directLoss, err := u.prober.ProbeDirectGateway(ctx, targetHost)
	if err != nil || directRTT <= 0 {
		directRTT, directLoss, _ = u.prober.ProbeDirectGateway(ctx, "1.1.1.1:53")
	}
	if directRTT <= 0 {
		http.Error(w, `{"error":"direct route probe failed; no advice can be produced without a measured RTT"}`, http.StatusServiceUnavailable)
		return
	}

	advice := CalculateRouteAdvice(directRTT, directLoss, bestRelay)
	_ = json.NewEncoder(w).Encode(advice)
}

func (u *UIServer) handleFailoverToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if u.engine == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"autoFailover": false})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Enabled != nil {
			u.engine.SetAutoFailover(*req.Enabled)
		} else {
			current := u.engine.IsAutoFailoverEnabled()
			u.engine.SetAutoFailover(!current)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"autoFailover": u.engine.IsAutoFailoverEnabled(),
	})
}

func (u *UIServer) handleFailoverHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var history []FailoverEvent
	if u.engine != nil {
		history = u.engine.FailoverHistory()
	}
	if history == nil {
		history = []FailoverEvent{}
	}
	_ = json.NewEncoder(w).Encode(history)
}

func (u *UIServer) handleFECToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if u.engine == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"fecEnabled": false})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Enabled != nil {
			u.engine.SetFECEnabled(*req.Enabled)
		} else {
			current := u.engine.IsFECEnabled()
			u.engine.SetFECEnabled(!current)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"fecEnabled": u.engine.IsFECEnabled(),
	})
}

func (u *UIServer) handleOptimizeProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if u.profileMgr == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "game profile catalog is unavailable")
		return
	}

	gameID := strings.TrimSpace(r.URL.Query().Get("gameId"))
	if gameID == "" {
		writeJSONError(w, http.StatusBadRequest, "gameId is required")
		return
	}

	game, ok := u.profileMgr.FindGameByID(gameID)
	if !ok {
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("game %q is not in the profile catalog", gameID))
		return
	}

	subnetsCount := 0
	for _, reg := range game.Regions {
		subnetsCount += len(reg.CIDRs)
	}

	fecEnabled := false
	fecBlockSize := 0
	if u.engine != nil {
		fecEnabled = u.engine.IsFECEnabled()
		fecBlockSize = u.engine.FECBlockSize()
	}

	details := []string{
		fmt.Sprintf("%d CIDR subnets across %d region(s) will be routed through the tunnel on connect.", subnetsCount, len(game.Regions)),
		"The tunnel MTU is negotiated with the relay and applied to the virtual adapter during connect.",
	}
	if fecEnabled && fecBlockSize > 0 {
		details = append(details, fmt.Sprintf("Systematic FEC is enabled with %d data packets per parity packet.", fecBlockSize))
	} else {
		details = append(details, "Systematic FEC is currently disabled.")
	}
	if runtime.GOOS == "windows" {
		details = append(details, "DSCP EF marking is requested through the Windows QoS API.")
	} else {
		details = append(details, "DSCP EF marking is requested with the IP_TOS socket option (best effort).")
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"gameId":       game.ID,
		"gameTitle":    game.Name,
		"genre":        game.Category,
		"status":       "preview",
		"applied":      false,
		"note":         "Nothing is changed by this request. Connect to install routes and start the tunnel.",
		"subnetsCount": subnetsCount,
		"regionCount":  len(game.Regions),
		"fecEnabled":   fecEnabled,
		"fecBlockSize": fecBlockSize,
		"qosSupported": true, // the relay socket is tagged on every supported OS
		"details":      details,
	})
}

func (u *UIServer) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	hostname, _ := os.Hostname()
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	if username == "" {
		username = "Player"
	}

	// Compute deterministic 4-digit discriminator from hostname & username
	hashInput := fmt.Sprintf("%s:%s", hostname, username)
	crc := crc32.ChecksumIEEE([]byte(hashInput))
	discriminator := 1000 + int(crc%9000)

	// Clean suggested nickname avoiding generic collisions
	suggestedNickname := fmt.Sprintf("%s#%d", username, discriminator)
	uLower := strings.ToLower(username)
	if uLower == "admin" || uLower == "administrator" || uLower == "user" || uLower == "pc" || uLower == "owner" || uLower == "gamer" || uLower == "player" {
		if hostname != "" {
			cleanHost := strings.TrimPrefix(hostname, "DESKTOP-")
			if len(cleanHost) > 8 {
				cleanHost = cleanHost[:8]
			}
			suggestedNickname = fmt.Sprintf("%s#%d", cleanHost, discriminator)
		}
	}

	if u.configStore != nil {
		savedNick := u.configStore.Get().Nickname
		if savedNick != "" && savedNick != "Gamer" && savedNick != "Player" {
			suggestedNickname = savedNick
		}
	}

	driverName := "WinTun NDIS"
	switch runtime.GOOS {
	case "darwin":
		driverName = "macOS utun"
	case "linux":
		driverName = "Linux dev/net/tun"
	}

	res := map[string]any{
		"hostname":          hostname,
		"username":          username,
		"suggestedNickname": suggestedNickname,
		"discriminator":     discriminator,
		"isAdmin":           isProcessElevated(),
		"os":                runtime.GOOS,
		"arch":              runtime.GOARCH,
		"driverName":        driverName,
	}
	if u.configStore != nil {
		res["savedTweaks"] = u.configStore.Get().Tweaks
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (u *UIServer) handleNetworkDiagnostics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Measure real RTT to public DNS resolvers via TCP dial.
	var ispPingMs float64
	for _, target := range []string{"1.1.1.1:53", "8.8.8.8:53"} {
		t0 := time.Now()
		conn, err := net.DialTimeout("tcp", target, 1200*time.Millisecond)
		if err == nil {
			ispPingMs = float64(time.Since(t0).Microseconds()) / 1000.0
			_ = conn.Close()
			break
		}
	}

	// 2. Measure the real default-gateway LAN RTT. No fabricated fallback.
	gateway := detectDefaultGateway()
	var lanPingMs float64
	if gateway != "" {
		tG := time.Now()
		gConn, gErr := net.DialTimeout("tcp", net.JoinHostPort(gateway, "80"), 400*time.Millisecond)
		if gErr == nil {
			lanPingMs = float64(time.Since(tG).Microseconds()) / 1000.0
			_ = gConn.Close()
		}
	}

	res := map[string]any{
		"gateway":     gateway,
		"lanPingMs":   math.Round(lanPingMs*10) / 10,
		"ispPingMs":   math.Round(ispPingMs*10) / 10,
		"dnsServer":   detectDNSServer(),
		"networkType": "",
		"measured":    ispPingMs > 0,
	}
	_ = json.NewEncoder(w).Encode(res)
}

// detectDefaultGateway returns the host's LAN default gateway, or "" when it
// cannot be determined. It never invents an address.
func detectDefaultGateway() string {
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	switch runtime.GOOS {
	case "linux":
		out, err := exec.CommandContext(ctx, "ip", "route", "show", "default").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			for i, field := range fields {
				if field == "via" && i+1 < len(fields) {
					return fields[i+1]
				}
			}
		}
	case "darwin":
		out, err := exec.CommandContext(ctx, "route", "-n", "get", "default").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "gateway:") {
					return strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
				}
			}
		}
	case "windows":
		out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
			"(Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Sort-Object RouteMetric | Select-Object -First 1).NextHop").Output()
		if err == nil {
			gw := strings.TrimSpace(string(out))
			if net.ParseIP(gw) != nil {
				return gw
			}
		}
	}
	return ""
}

// detectDNSServer reads the configured resolver without assuming a provider.
func detectDNSServer() string {
	if runtime.GOOS == "windows" {
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()
		out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
			"(Get-DnsClientServerAddress -AddressFamily IPv4 | Where-Object { $_.ServerAddresses } | Select-Object -First 1).ServerAddresses[0]").Output()
		if err == nil {
			if dns := strings.TrimSpace(string(out)); dns != "" {
				return dns
			}
		}
		return ""
	}

	if runtime.GOOS == "darwin" {
		// macOS does not use /etc/resolv.conf; scutil reads the live
		// SystemConfiguration resolver list.
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()
		out, err := exec.CommandContext(ctx, "scutil", "--dns").Output()
		if err != nil {
			return ""
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			// Lines look like: nameserver[0] : 192.168.1.1
			if !strings.HasPrefix(line, "nameserver[") {
				continue
			}
			if _, value, found := strings.Cut(line, ":"); found {
				if dns := strings.TrimSpace(value); dns != "" {
					return dns
				}
			}
		}
		return ""
	}

	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			return fields[1]
		}
	}
	return ""
}

type RegionTelemetry struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Anchor      string  `json:"anchor,omitempty"`
	DirectRTTMs float64 `json:"directRttMs"`
	Measured    bool    `json:"measured"`
}

type GameTelemetryItem struct {
	GameID       string            `json:"gameId"`
	Region       string            `json:"region"`
	BaselinePing float64           `json:"baselinePing"`
	AccelPing    float64           `json:"accelPing"`
	Trend        string            `json:"trend"`
	Measured     bool              `json:"measured"`
	Source       string            `json:"source,omitempty"`
	Anchor       string            `json:"anchor,omitempty"`
	Regions      []RegionTelemetry `json:"regions,omitempty"`
}

type telemetryAnchor struct {
	id      string
	label   string
	probeID byte // key in protocol.AnchorProbeTargets (shared with the relay)
	keys    []string
}

// telemetryAnchors are public cloud endpoints used only as direct-route RTT
// references. They are not game servers, so the API never presents their RTT
// as a game-server ping. The probeID lets the relay measure the same target
// from its side without accepting an arbitrary host from a client.
var telemetryAnchors = []telemetryAnchor{
	{"sg", "Singapore", 1, []string{"singapore", "asia-sg", "sg-"}},
	{"jp", "Tokyo", 2, []string{"tokyo", "japan", "asia-jp"}},
	{"kr", "Seoul", 3, []string{"seoul", "korea", "asia-kr"}},
	{"hk", "Hong Kong", 4, []string{"hong kong", "hongkong", "asia-hk"}},
	{"in", "Mumbai", 5, []string{"mumbai", "india", "asia-mumbai"}},
	{"au", "Sydney", 6, []string{"sydney", "oceania", "oc-"}},
	{"us-east", "US East", 7, []string{"us-east", "virginia", "chicago", "na-us-east", "na-chicago"}},
	{"us-west", "US West", 8, []string{"us-west", "california", "na-us-west"}},
	{"eu", "Frankfurt", 9, []string{"frankfurt", "london", "stockholm", "europe", "eu-"}},
	{"sa", "Sao Paulo", 10, []string{"sao paulo", "são paulo", "brazil", "south america", "sa-"}},
	{"af", "Johannesburg", 11, []string{"johannesburg", "africa", "af-"}},
}

// anchorTarget resolves the allowlisted probe target for an anchor.
func anchorTarget(anchor telemetryAnchor) string {
	return protocol.AnchorProbeTargets[anchor.probeID]
}

// resolveAnchorProbeID maps a game/region selection to the shared anchor
// allowlist id, or 0 when the region has no anchor.
func (u *UIServer) resolveAnchorProbeID(gameID, regionID string) byte {
	return anchorProbeIDForRegion(u.profileMgr, gameID, regionID)
}

func matchTelemetryAnchor(regionID, regionName string) (telemetryAnchor, bool) {
	haystack := strings.ToLower(regionID + " " + regionName)
	for _, anchor := range telemetryAnchors {
		for _, key := range anchor.keys {
			if strings.Contains(haystack, key) {
				return anchor, true
			}
		}
	}
	return telemetryAnchor{}, false
}

var (
	gameTelemMu    sync.Mutex
	gameTelemCache map[string]GameTelemetryItem
	gameTelemTime  time.Time
)

func (u *UIServer) handleGameTelemetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if u.profileMgr == nil {
		_ = json.NewEncoder(w).Encode(map[string]GameTelemetryItem{})
		return
	}

	gameTelemMu.Lock()
	if time.Since(gameTelemTime) < 30*time.Second && len(gameTelemCache) > 0 {
		cached := gameTelemCache
		gameTelemMu.Unlock()
		_ = json.NewEncoder(w).Encode(cached)
		return
	}
	gameTelemMu.Unlock()

	// Measure direct-route RTT to public regional anchors concurrently.
	rtts := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(r.Context(), 2500*time.Millisecond)
	defer cancel()

	for _, anchor := range telemetryAnchors {
		target := anchorTarget(anchor)
		if target == "" {
			continue
		}
		wg.Add(1)
		go func(a telemetryAnchor, target string) {
			defer wg.Done()
			vals, err := probeTCPGateway(ctx, target, 2)
			if err == nil && len(vals) > 0 {
				median := calculateMedian(vals)
				mu.Lock()
				rtts[a.id] = math.Round(median*10) / 10
				mu.Unlock()
			}
		}(anchor, target)
	}
	wg.Wait()

	var engineStats TunnelStats
	if u.engine != nil {
		engineStats = u.engine.Stats()
	}

	result := make(map[string]GameTelemetryItem)
	anyMeasured := false

	for _, g := range u.profileMgr.Catalog().Games {
		var baseline, accel float64
		var regionName, anchorLabel, anchorID string
		regions := make([]RegionTelemetry, 0, len(g.Regions))

		for _, region := range g.Regions {
			rt := RegionTelemetry{ID: region.ID, Name: region.Name}
			if anchor, ok := matchTelemetryAnchor(region.ID, region.Name); ok {
				rt.Anchor = anchor.id
				if rtt, measured := rtts[anchor.id]; measured && rtt > 0 {
					rt.DirectRTTMs = rtt
					rt.Measured = true
					if baseline == 0 || rtt < baseline {
						baseline = rtt
						regionName = region.Name
						anchorLabel = anchor.label
						anchorID = anchor.id
					}
				}
			}
			regions = append(regions, rt)
		}
		if regionName == "" && len(g.Regions) > 0 {
			regionName = g.Regions[0].Name
		}
		if regionName == "" {
			regionName = "Global"
		}

		source := "No direct anchor measured"
		if baseline > 0 {
			source = fmt.Sprintf("Lowest direct TCP RTT across regional cloud anchors (%s); anchors are not game servers", anchorLabel)
		}

		if engineStats.State == StateConnected && strings.EqualFold(engineStats.ActiveGame, g.ID) && engineStats.PingMs > 0 {
			accel = float64(engineStats.PingMs)
			source = "Live tunnel RTT measured by the engine"
			for _, region := range g.Regions {
				if region.ID == engineStats.ActiveRegion {
					regionName = region.Name
					break
				}
			}
		}

		trend := "Not measured"
		switch {
		case baseline > 0 && accel > 0:
			if accel < baseline-1.0 {
				pct := int(math.Round((1.0 - (accel / baseline)) * 100))
				trend = fmt.Sprintf("%d%% lower than the direct anchor", pct)
			} else {
				trend = "No measured improvement over the direct anchor"
			}
		case baseline > 0:
			trend = "Direct anchor measured; connect to compare tunnel RTT"
		case accel > 0:
			trend = "Tunnel RTT measured; no matching direct anchor"
		}

		measured := baseline > 0 || accel > 0
		if measured {
			anyMeasured = true
		}
		result[g.ID] = GameTelemetryItem{
			GameID:       g.ID,
			Region:       regionName,
			BaselinePing: math.Round(baseline*10) / 10,
			AccelPing:    math.Round(accel*10) / 10,
			Trend:        trend,
			Measured:     measured,
			Source:       source,
			Anchor:       anchorID,
			Regions:      regions,
		}
	}

	if anyMeasured {
		gameTelemMu.Lock()
		gameTelemCache = result
		gameTelemTime = time.Now()
		gameTelemMu.Unlock()
	}

	_ = json.NewEncoder(w).Encode(result)
}

func (u *UIServer) handleRelaunchAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if isProcessElevated() {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "Already Elevated", "isAdmin": true})
		return
	}

	if runtime.GOOS == "windows" {
		exe, err := os.Executable()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Relaunch with Start-Process -Verb RunAs to invoke Windows UAC modal
		cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", exe))
		if err := cmd.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "UAC Prompt Triggered", "isAdmin": false})
		return
	}

	if runtime.GOOS == "darwin" {
		exe, err := os.Executable()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// On macOS, trigger native graphical TouchID / Password prompt via AppleScript
		script := fmt.Sprintf("do shell script \"%s\" with administrator privileges", exe)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "macOS Authorization Prompt Triggered", "isAdmin": false})
		return
	}

	if runtime.GOOS == "linux" {
		exe, err := os.Executable()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// On Linux, trigger PolicyKit graphical elevation modal (pkexec)
		if pkPath, err := exec.LookPath("pkexec"); err == nil {
			cmd := exec.Command(pkPath, exe)
			if err := cmd.Start(); err == nil {
				_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "PolicyKit Elevation Prompt Triggered", "isAdmin": false})
				return
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "status": "Please run with 'sudo ./lagvex-client' in terminal", "isAdmin": false})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "Unknown Platform"})
}

func (u *UIServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": "Shutting down..."})

	go func() {
		time.Sleep(300 * time.Millisecond)
		if u.engine != nil {
			_ = u.engine.Disconnect()
		}
		os.Exit(0)
	}()
}

// ── SQUAD ROOM REAL-TIME SIGNALING ──

type SquadMemberInfo struct {
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	ISP      string    `json:"isp"`
	Ping     int       `json:"ping"`
	Game     string    `json:"game"`
	Status   string    `json:"status"`
	IsSelf   bool      `json:"isSelf,omitempty"`
	LastSeen time.Time `json:"lastSeen"`
}

type SquadRoom struct {
	Code      string                      `json:"code"`
	Host      string                      `json:"host"`
	RelayID   string                      `json:"relayId"`
	RelayName string                      `json:"relayName"`
	Members   map[string]*SquadMemberInfo `json:"members"`
	CreatedAt time.Time                   `json:"createdAt"`
	mu        sync.Mutex
}

var (
	squadRoomsMu sync.Mutex
	squadRooms   = make(map[string]*SquadRoom)
)

func (u *UIServer) getRelaySignalingURL() string {
	if u.engine == nil {
		return ""
	}
	stats := u.engine.Stats()
	if stats.State != StateConnected || stats.RelayAddr == "" {
		return ""
	}
	host, portStr, err := net.SplitHostPort(stats.RelayAddr)
	if err != nil {
		return ""
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", host, port+1)
}

func (u *UIServer) proxyRelaySquad(w http.ResponseWriter, path, method string, body []byte) bool {
	relayURL := u.getRelaySignalingURL()
	if relayURL == "" {
		return false
	}
	targetURL := relayURL + path
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	var req *http.Request
	var err error
	if len(body) > 0 {
		req, err = http.NewRequest(method, targetURL, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, targetURL, nil)
	}
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return true
	}
	return false
}

func (u *UIServer) handleSquadCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code      string `json:"code"`
		Host      string `json:"host"`
		Name      string `json:"name"`
		RelayID   string `json:"relayId"`
		RelayName string `json:"relayName"`
		Game      string `json:"game"`
		ISP       string `json:"isp"`
		Ping      int    `json:"ping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Host == "" {
		req.Host = req.Name
	}
	req.Host = strings.TrimSpace(req.Host)
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if req.Host == "" {
		writeJSONError(w, http.StatusBadRequest, "host name is required")
		return
	}
	if req.Code == "" {
		req.Code = fmt.Sprintf("LGVX-%04d", time.Now().UnixNano()%9000+1000)
	}

	bodyBytes, _ := json.Marshal(req)
	if u.proxyRelaySquad(w, "/squad/create", "POST", bodyBytes) {
		return
	}

	squadRoomsMu.Lock()
	room := &SquadRoom{
		Code:      req.Code,
		Host:      req.Host,
		RelayID:   req.RelayID,
		RelayName: req.RelayName,
		Members:   make(map[string]*SquadMemberInfo),
		CreatedAt: time.Now(),
	}
	room.Members[req.Host] = &SquadMemberInfo{
		Name:     req.Host,
		Role:     "Host",
		ISP:      req.ISP,
		Ping:     req.Ping,
		Game:     req.Game,
		Status:   "In room",
		LastSeen: time.Now(),
	}
	squadRooms[req.Code] = room
	squadRoomsMu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"room":    room,
	})
}

func (u *UIServer) handleSquadJoin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Game string `json:"game"`
		ISP  string `json:"isp"`
		Ping int    `json:"ping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "room code is required")
		return
	}
	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "player name is required")
		return
	}

	bodyBytes, _ := json.Marshal(req)
	if u.proxyRelaySquad(w, "/squad/join", "POST", bodyBytes) {
		return
	}

	squadRoomsMu.Lock()
	room, ok := squadRooms[req.Code]
	squadRoomsMu.Unlock()
	if !ok {
		writeJSONError(w, http.StatusNotFound, "room not found; the host must create it first")
		return
	}

	room.mu.Lock()
	if len(room.Members) >= 5 && room.Members[req.Name] == nil {
		room.mu.Unlock()
		writeJSONError(w, http.StatusConflict, "squad room is full (5/5 players)")
		return
	}
	role := "Member"
	if req.Name == room.Host {
		role = "Host"
	}
	room.Members[req.Name] = &SquadMemberInfo{
		Name:     req.Name,
		Role:     role,
		ISP:      req.ISP,
		Ping:     req.Ping,
		Game:     req.Game,
		Status:   "In room",
		LastSeen: time.Now(),
	}
	memberList := make([]SquadMemberInfo, 0, len(room.Members))
	for _, m := range room.Members {
		memberList = append(memberList, *m)
	}
	room.mu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"code":    room.Code,
		"relay":   room.RelayName,
		"members": memberList,
	})
}

func (u *UIServer) handleSquadRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("code")))
	if code == "" {
		http.Error(w, `{"error":"room code required"}`, http.StatusBadRequest)
		return
	}

	if u.proxyRelaySquad(w, "/squad/room?code="+code, "GET", nil) {
		return
	}

	squadRoomsMu.Lock()
	room, ok := squadRooms[code]
	squadRoomsMu.Unlock()

	if !ok {
		http.Error(w, `{"error":"room not found"}`, http.StatusNotFound)
		return
	}

	room.mu.Lock()
	now := time.Now()
	// Purge stale members that haven't heartbeated in > 30s
	for name, m := range room.Members {
		if now.Sub(m.LastSeen) > 30*time.Second && m.Role != "Host" {
			delete(room.Members, name)
		}
	}
	memberList := make([]SquadMemberInfo, 0, len(room.Members))
	for _, m := range room.Members {
		memberList = append(memberList, *m)
	}
	hostName := room.Host
	relayName := room.RelayName
	room.mu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":      code,
		"host":      hostName,
		"relayName": relayName,
		"members":   memberList,
		"count":     len(memberList),
	})
}

func (u *UIServer) handleSquadHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Ping int    `json:"ping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	bodyBytes, _ := json.Marshal(req)
	if u.proxyRelaySquad(w, "/squad/heartbeat", "POST", bodyBytes) {
		return
	}

	squadRoomsMu.Lock()
	room, ok := squadRooms[req.Code]
	squadRoomsMu.Unlock()

	if ok {
		room.mu.Lock()
		if m, exists := room.Members[req.Name]; exists {
			m.LastSeen = time.Now()
			// A zero RTT is a real state (not boosting/no sample yet), so it
			// replaces the previous value instead of leaving a stale reading.
			m.Ping = req.Ping
		}
		room.mu.Unlock()
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (u *UIServer) handleSquadLeave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	leaveBytes, _ := json.Marshal(req)
	if u.proxyRelaySquad(w, "/squad/leave", "POST", leaveBytes) {
		return
	}

	squadRoomsMu.Lock()
	room, ok := squadRooms[req.Code]
	if ok {
		room.mu.Lock()
		delete(room.Members, req.Name)
		empty := len(room.Members) == 0
		room.mu.Unlock()
		if empty {
			delete(squadRooms, req.Code)
		}
	}
	squadRoomsMu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// ── BUFFERBLOAT & LATENCY UNDER LOAD DIAGNOSTIC ──

func (u *UIServer) handleDiagnoseBufferbloat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Measure baseline idle RTT. If this fails, report that no measurement
	// was possible instead of inventing a value.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	basePings, _ := probeTCPGateway(ctx, "1.1.1.1:53", 4)
	if len(basePings) == 0 {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"available": false,
			"error":     "baseline RTT probe failed; no bufferbloat measurement was taken",
		})
		return
	}
	baseline := calculateMedian(basePings)

	// 2. Generate a short saturated network burst in parallel.
	burstDone := make(chan struct{})
	go func() {
		defer close(burstDone)
		tr := &http.Transport{DisableKeepAlives: true}
		client := &http.Client{Transport: tr, Timeout: 1200 * time.Millisecond}
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get("https://speed.cloudflare.com/__down?bytes=5000000")
				if err == nil && resp != nil {
					buf := make([]byte, 32768)
					for {
						n, rErr := resp.Body.Read(buf)
						if n == 0 || rErr != nil {
							break
						}
					}
					_ = resp.Body.Close()
				}
			}()
		}
		wg.Wait()
	}()

	// 3. Measure loaded RTT during the burst.
	time.Sleep(150 * time.Millisecond)
	loadedPings, _ := probeTCPGateway(ctx, "1.1.1.1:53", 4)
	<-burstDone

	if len(loadedPings) == 0 {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"available":  false,
			"baselineMs": math.Round(baseline*10) / 10,
			"error":      "loaded RTT probe failed; bufferbloat could not be measured",
		})
		return
	}
	loaded := calculateMedian(loadedPings)
	diff := loaded - baseline
	gradeDiff := diff
	if gradeDiff < 0 {
		gradeDiff = 0
	}

	var grade, rating, advice string
	switch {
	case gradeDiff <= 6.0:
		grade = "A+"
		rating = "Ultra Low Latency (Zero Queuing)"
		advice = "The connection kept latency flat while the test saturated the link."
	case gradeDiff <= 15.0:
		grade = "A"
		rating = "Minimal Bufferbloat"
		advice = "Negligible queue delay appeared during background downloads."
	case gradeDiff <= 30.0:
		grade = "B"
		rating = "Moderate Bufferbloat"
		advice = "Minor queue delay detected. Enabling Lagvex QoS can prioritize game UDP traffic."
	case gradeDiff <= 60.0:
		grade = "C"
		rating = "Noticeable Bufferbloat"
		advice = "The router queues traffic under load. Lagvex QoS and MTU tuning may help."
	default:
		grade = "D"
		rating = "Severe Bufferbloat Detected"
		advice = "Substantial queuing delay under load. Limit background downloads and enable QoS."
	}

	res := map[string]any{
		"available":  true,
		"baselineMs": math.Round(baseline*10) / 10,
		"loadedMs":   math.Round(loaded*10) / 10,
		"diffMs":     math.Round(diff*10) / 10,
		"grade":      grade,
		"rating":     rating,
		"advice":     advice,
	}
	_ = json.NewEncoder(w).Encode(res)
}
