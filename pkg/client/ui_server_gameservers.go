package client

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleGameServerHistory serves the persisted per-game game-server history
// (GET) and records a new measurement (POST). Only genuinely measured RTTs are
// accepted: the API refuses records without a positive avgMs, so a failed
// probe can never fabricate history. Hosts are IP literals only — the same
// rule as the TCP ping tool — so recording never performs a DNS lookup.
func (u *UIServer) handleGameServerHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		gameID := strings.TrimSpace(r.URL.Query().Get("gameId"))
		if u.configStore == nil || gameID == "" {
			_ = json.NewEncoder(w).Encode([]GameServerRecord{})
			return
		}
		list := u.configStore.GameServersFor(gameID)
		if list == nil {
			list = []GameServerRecord{}
		}
		_ = json.NewEncoder(w).Encode(list)

	case http.MethodPost:
		var req struct {
			GameID  string  `json:"gameId"`
			Host    string  `json:"host"`
			Port    int     `json:"port"`
			AvgMs   float64 `json:"avgMs"`
			LossPct float64 `json:"lossPct"`
			Samples int     `json:"samples"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Host = strings.TrimSpace(req.Host)
		if req.GameID == "" || req.Host == "" || req.AvgMs <= 0 {
			http.Error(w, "gameId, host and a positive avgMs are required; failed probes are never recorded", http.StatusBadRequest)
			return
		}
		if net.ParseIP(req.Host) == nil {
			http.Error(w, "host must be an IP literal", http.StatusBadRequest)
			return
		}
		if req.Port <= 0 || req.Port > 65535 {
			req.Port = 443
		}
		if req.Samples <= 0 {
			req.Samples = 1
		}
		if u.configStore == nil {
			http.Error(w, "config store unavailable", http.StatusServiceUnavailable)
			return
		}
		u.configStore.RecordGameServer(req.GameID, GameServerRecord{
			Host:     req.Host,
			Port:     req.Port,
			AvgMs:    req.AvgMs,
			LossPct:  req.LossPct,
			Samples:  req.Samples,
			LastSeen: time.Now().UTC(),
		})
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "recorded": strconv.Quote(req.Host)})

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
