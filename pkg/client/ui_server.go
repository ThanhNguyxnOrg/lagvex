package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/web"
)

// UIServer serves the modern Gaming Booster Dashboard and REST API.
type UIServer struct {
	engine     *Engine
	profileMgr *profiles.Manager
	webDir     string
	server     *http.Server
	prober     *Prober
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

	return &UIServer{
		engine:     engine,
		profileMgr: pm,
		webDir:     webDir,
		prober:     NewProber(),
	}
}

// Start listens on the specified address (e.g. "127.0.0.1:18888").
func (u *UIServer) Start(addr string) error {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/status", u.handleStatus)
	mux.HandleFunc("/api/games", u.handleGames)
	mux.HandleFunc("/api/relays", u.handleRelays)
	mux.HandleFunc("/api/probe-relays", u.handleProbeRelays)
	mux.HandleFunc("/api/best-relay", u.handleBestRelay)
	mux.HandleFunc("/api/connect", u.handleConnect)
	mux.HandleFunc("/api/disconnect", u.handleDisconnect)
	mux.HandleFunc("/api/test-relay", u.handleTestRelay)
	mux.HandleFunc("/api/ping-relay", u.handleTestRelay)
	mux.HandleFunc("/api/add-game", u.handleAddGame)
	mux.HandleFunc("/api/advisor", u.handleAdvisor)
	mux.HandleFunc("/api/failover/toggle", u.handleFailoverToggle)
	mux.HandleFunc("/api/failover/history", u.handleFailoverHistory)
	mux.HandleFunc("/api/fec/toggle", u.handleFECToggle)

	// Static Web Assets (disk prioritization with embedded binary fallback)
	fs := http.FileServer(web.GetFileSystem(u.webDir))
	mux.Handle("/", fs)

	u.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("[Dashboard] Web Dashboard active at http://%s", addr)
	return u.server.ListenAndServe()
}

func (u *UIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := u.engine.Stats()
	_ = json.NewEncoder(w).Encode(stats)
}

func (u *UIServer) handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	_ = json.NewEncoder(w).Encode(catalog.Games)
}

func (u *UIServer) handleRelays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()
	_ = json.NewEncoder(w).Encode(catalog.Relays)
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	prober := u.prober
	if samples != prober.cfg.DefaultSamples {
		cfg := prober.cfg
		cfg.DefaultSamples = samples
		prober = NewProberWithConfig(cfg)
	}

	results := prober.ProbeAll(ctx, relays, 8)
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	best, err := u.prober.SelectBest(ctx, relays)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(best)
}

type connectReq struct {
	RelayEndpoint string `json:"relayEndpoint"`
	PSK           string `json:"psk"`
	GameID        string `json:"gameId"`
	RegionID      string `json:"regionId"`
	ForceNow      bool   `json:"forceNow"`
	AutoNode      bool   `json:"autoNode"`
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

	catalog := u.profileMgr.Catalog()

	// Auto Node Selection if requested or endpoint is "auto"
	if req.AutoNode || req.RelayEndpoint == "auto" || req.RelayEndpoint == "" {
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		best, err := u.prober.SelectBest(ctx, catalog.Relays)
		if err != nil {
			http.Error(w, "auto node selection failed: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		req.RelayEndpoint = best.Endpoint
		// Lookup PSK from catalog
		for _, rel := range catalog.Relays {
			if rel.Endpoint == best.Endpoint {
				req.PSK = rel.PSK
				break
			}
		}
		log.Printf("[Dashboard] Auto-selected optimal relay: %s (%s, score=%.1f)", best.Name, best.Endpoint, best.Score)
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
		http.Error(w, "relayEndpoint is required", http.StatusBadRequest)
		return
	}

	go func() {
		if err := u.engine.Connect(req.RelayEndpoint, []byte(req.PSK), req.GameID, req.RegionID, req.ForceNow); err != nil {
			log.Printf("[Dashboard] Connect error: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "connecting",
		"endpoint": req.RelayEndpoint,
	})
}

func (u *UIServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		_ = u.engine.Disconnect()
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "disconnecting"})
}

func (u *UIServer) handleTestRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var endpoint, psk string

	if r.Method == http.MethodGet {
		endpoint = r.URL.Query().Get("endpoint")
		psk = r.URL.Query().Get("psk")
	} else if r.Method == http.MethodPost {
		var req struct {
			Endpoint string `json:"endpoint"`
			PSK      string `json:"psk"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			endpoint = req.Endpoint
			psk = req.PSK
		}
	} else {
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
		http.Error(w, "id and name are required", http.StatusBadRequest)
		return
	}

	if err := u.profileMgr.AddGame(game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "game": game})
}

func (u *UIServer) handleAdvisor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameID := r.URL.Query().Get("gameId")
	regionID := r.URL.Query().Get("regionId")

	if gameID == "" || regionID == "" {
		stats := u.engine.Stats()
		if gameID == "" {
			gameID = stats.ActiveGame
		}
		if regionID == "" {
			regionID = stats.ActiveRegion
		}
	}

	catalog := u.profileMgr.Catalog()
	if len(catalog.Relays) == 0 {
		http.Error(w, `{"error":"no relays configured in catalog"}`, http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// 1. Probe candidate relays to find optimal relay
	bestRelay, err := u.prober.SelectBest(ctx, catalog.Relays)
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
						prefix := strings.Split(reg.CIDRs[0], "/")[0]
						targetHost = prefix + ":443"
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
		if directRTT <= 0 {
			directRTT = 35.0
			directLoss = 0.0
		}
	}

	advice := CalculateRouteAdvice(directRTT, directLoss, bestRelay)
	_ = json.NewEncoder(w).Encode(advice)
}

func (u *UIServer) handleFailoverToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
	history := u.engine.FailoverHistory()
	if history == nil {
		history = []FailoverEvent{}
	}
	_ = json.NewEncoder(w).Encode(history)
}

func (u *UIServer) handleFECToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
