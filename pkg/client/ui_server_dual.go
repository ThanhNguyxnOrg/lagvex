package client

import (
	"encoding/json"
	"net/http"
)

// handleDualPathToggle enables or disables dual-path redundancy (Redundant
// lite Phase 2). Enabling runs a real standby handshake, so the response
// reflects the measured outcome, not just the intent.
func (u *UIServer) handleDualPathToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if u.engine == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "engine is not running")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u.engine.SetDualPath(req.Enabled)

	stats := u.engine.Stats()
	resp := map[string]any{
		"enabled": u.engine.IsDualPathEnabled(),
		"active":  stats.DualPath != nil && stats.DualPath.Active,
	}
	if stats.DualPath != nil {
		resp["dualPath"] = stats.DualPath
	}
	_ = json.NewEncoder(w).Encode(resp)
}
