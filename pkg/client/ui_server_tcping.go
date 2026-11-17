package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// handleTCPPing measures the TCP handshake RTT to one public game-server IP.
// It is a POST so the dashboard CSRF guard covers it: a random web page must
// not be able to use the local client as a port scanner.
func (u *UIServer) handleTCPPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Target    string `json:"target"`
		Port      int    `json:"port"`
		Attempts  int    `json:"attempts"`
		TimeoutMs int    `json:"timeoutMs"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ip, err := validateTCPPingTarget(req.Target)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	port, attempts, timeout, err := clampTCPPingDefaults(req.Port, req.Attempts, req.TimeoutMs)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if wait, err := u.tcpPingLimiter.allow(); err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(wait)))
		writeJSONError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	budget := time.Duration(attempts)*(timeout+tcpPingDefaultInterval) + 2*time.Second
	ctx, cancel := context.WithTimeout(r.Context(), budget)
	defer cancel()

	result := runTCPPing(ctx, ip, port, attempts, timeout, tcpPingDefaultInterval)
	_ = json.NewEncoder(w).Encode(result)
}
