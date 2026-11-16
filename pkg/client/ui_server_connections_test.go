package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGameConnectionsEndpointShape(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/game-connections", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		ScannedAt string `json:"scannedAt"`
		Supported bool   `json:"supported"`
		Note      string `json:"note"`
		Reason    string `json:"reason"`
		Games     []struct {
			GameID      string              `json:"gameId"`
			Name        string              `json:"name"`
			Connections []ProcessConnection `json:"connections"`
		} `json:"games"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.ScannedAt == "" {
		t.Fatalf("expected scannedAt to be set")
	}
	if res.Note == "" {
		t.Fatalf("expected the platform note to explain what is scanned")
	}
	if !res.Supported && res.Reason == "" {
		t.Fatalf("an unsupported platform must carry a reason")
	}
	if len(res.Games) == 0 {
		t.Fatalf("expected the game catalog to be reflected in the response")
	}
	for _, game := range res.Games {
		if game.GameID == "" || game.Name == "" {
			t.Fatalf("game entry missing identifiers: %+v", game)
		}
	}
}

func TestPeerRoutingToggleEndpoint(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	server := NewUIServer(engine, nil, "")
	handler := server.setupRoutes()

	for _, want := range []bool{true, false} {
		body, _ := json.Marshal(map[string]bool{"enabled": want})
		req := httptest.NewRequest(http.MethodPost, "/api/peer-routing/toggle", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var res struct {
			Success bool `json:"success"`
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !res.Success || res.Enabled != want {
			t.Fatalf("toggle wanted %v, got %+v", want, res)
		}
	}
}

func TestPeerRoutingToggleRequiresEngine(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodPost, "/api/peer-routing/toggle", bytes.NewReader([]byte(`{"enabled":true}`)))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without an engine, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFullTunnelToggleEndpoint(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("create engine: %v", err)
	}
	server := NewUIServer(engine, nil, "")
	handler := server.setupRoutes()

	for _, want := range []bool{true, false} {
		body, _ := json.Marshal(map[string]bool{"enabled": want})
		req := httptest.NewRequest(http.MethodPost, "/api/full-tunnel/toggle", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var res struct {
			Success bool `json:"success"`
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !res.Success || res.Enabled != want {
			t.Fatalf("full tunnel toggle wanted %v, got %+v", want, res)
		}
	}

	// /api/status carries the flag even while disconnected.
	statusReq := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusRec := httptest.NewRecorder()
	handler.ServeHTTP(statusRec, statusReq)
	if !strings.Contains(statusRec.Body.String(), "fullTunnel") {
		t.Fatalf("status must expose fullTunnel, got %s", statusRec.Body.String())
	}
}
