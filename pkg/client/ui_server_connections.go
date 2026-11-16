package client

import (
	"encoding/json"
	"net/http"
	"time"
)

type gameConnectionStatus struct {
	GameID      string              `json:"gameId"`
	Name        string              `json:"name"`
	Running     bool                `json:"running"`
	Process     string              `json:"process,omitempty"`
	PID         int                 `json:"pid,omitempty"`
	Connections []ProcessConnection `json:"connections"`
}

// handleGameConnections lists the remote endpoints held by running game
// processes, read from the local OS socket tables only. It never attaches to,
// injects into, or reads memory of a game process.
func (u *UIServer) handleGameConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	catalog := u.profileMgr.Catalog()

	var targets []string
	for _, game := range catalog.Games {
		targets = append(targets, game.ProcessNames...)
	}
	matches := findRunningProcesses(targets)

	games := make([]gameConnectionStatus, 0, len(catalog.Games))
	supported := true
	for _, game := range catalog.Games {
		status := gameConnectionStatus{
			GameID:      game.ID,
			Name:        game.Name,
			Connections: []ProcessConnection{},
		}
		for _, proc := range game.ProcessNames {
			pid, ok := matches[proc]
			if !ok {
				continue
			}
			status.Running = true
			status.Process = proc
			status.PID = pid
			conns, ok := scanProcessConnections(pid)
			if !ok {
				supported = false
			}
			status.Connections = mergeConnections(status.Connections, conns)
			break
		}
		games = append(games, status)
	}

	resp := map[string]any{
		"scannedAt": time.Now().UTC().Format(time.RFC3339),
		"supported": supported,
		"note":      connectionScanNote,
		"games":     games,
	}
	if !supported {
		resp["reason"] = "socket enumeration is unavailable on this system"
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// handlePeerRoutingToggle enables or disables opt-in /32 routing for the
// public peers a running game process is observed talking to.
func (u *UIServer) handlePeerRoutingToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if u.engine == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "engine is not running")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u.engine.SetPeerRouting(req.Enabled)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"enabled":    u.engine.PeerRoutingEnabled(),
		"peerRoutes": u.engine.ActivePeerRoutes(),
	})
}

// handleFullTunnelToggle enables or disables routing all IPv4 traffic through
// the relay. This is an explicit user choice: while it is on, a slow or down
// relay slows down or breaks normal browsing too.
func (u *UIServer) handleFullTunnelToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if u.engine == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "engine is not running")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u.engine.SetFullTunnel(req.Enabled)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"enabled": u.engine.FullTunnelEnabled(),
	})
}
