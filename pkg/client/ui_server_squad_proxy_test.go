package client

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProxySquadHTTPForwardsRelayErrors guards dashboard honesty: when the
// relay rejects a squad request (rate limit, full room, missing room), the
// dashboard must see the relay's real answer instead of silently creating a
// local-only room that no teammate can join.
func TestProxySquadHTTPForwardsRelayErrors(t *testing.T) {
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"too many squad rooms created, try again later"}`))
	}))
	t.Cleanup(relay.Close)

	rec := httptest.NewRecorder()
	handled := proxySquadHTTP(rec, relay.URL+"/squad/create", http.MethodPost, []byte(`{"host":"player"}`))
	if !handled {
		t.Fatal("a relay answer was not treated as handled")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want the relay's 429", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "too many squad rooms") {
		t.Fatalf("body = %q, want the relay error text", rec.Body.String())
	}
}

// TestProxySquadHTTPForwardsSuccess checks the normal path keeps the relay's
// status and payload untouched.
func TestProxySquadHTTPForwardsSuccess(t *testing.T) {
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"LGVX-K7P2QD","success":true}`))
	}))
	t.Cleanup(relay.Close)

	rec := httptest.NewRecorder()
	if !proxySquadHTTP(rec, relay.URL+"/squad/create", http.MethodPost, []byte(`{}`)) {
		t.Fatal("relay response was not handled")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "LGVX-K7P2QD" {
		t.Fatalf("code = %q, want the relay's code", body.Code)
	}
}

// TestProxySquadHTTPReportsUnreachableRelay makes sure a connected but
// unreachable signaling port is an explicit error, never a silent local room.
func TestProxySquadHTTPReportsUnreachableRelay(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	rec := httptest.NewRecorder()
	if !proxySquadHTTP(rec, "http://"+addr+"/squad/create", http.MethodPost, []byte(`{}`)) {
		t.Fatal("unreachable relay was not treated as handled")
	}
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !strings.Contains(body.Error, "unreachable") {
		t.Fatalf("error = %q, want an explicit unreachable message", body.Error)
	}
}
