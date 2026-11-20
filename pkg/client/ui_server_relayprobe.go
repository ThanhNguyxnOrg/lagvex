package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleRelayGameServerProbe asks the connected relay to measure the TCP
// handshake RTT to a game server, so the Boost page can show the
// relay->server segment instead of only the client->relay half (ExitLag's
// "PingToDS" equivalent). The relay answers; this endpoint only validates the
// target and forwards the verdict.
func (u *UIServer) handleRelayGameServerProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Port <= 0 || req.Port > 65535 {
		writeJSONError(w, http.StatusBadRequest, "port must be between 1 and 65535")
		return
	}
	// Same rule as the local TCP ping tool: explicit public IP literals only,
	// so the same value the client refused to dial is never sent onward.
	ip, err := validateTCPPingTarget(strings.TrimSpace(req.Host))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	relayURL := u.getRelaySignalingURL()
	if relayURL == "" {
		writeJSONError(w, http.StatusConflict, "no connected relay: the relay segment can only be measured through a tunnel")
		return
	}

	body, _ := json.Marshal(map[string]any{"host": ip.String(), "port": req.Port})
	// proxySquadHTTP forwards whatever the relay answers, including its 429
	// rate limit, and reports an unreachable relay as an explicit 502.
	_ = proxySquadHTTP(w, relayURL+"/probe/gameserver", http.MethodPost, body)
}
