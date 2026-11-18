package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGameServerHistoryEndpoint(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	// Empty history for an unknown game.
	get := func(gameID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/game-server-history?gameId="+gameID, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	rec := get("unknown-game")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status %d", rec.Code)
	}
	var list []GameServerRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode GET: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty history, got %v", list)
	}

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/game-server-history", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	// A real measurement is recorded.
	if rec := post(`{"gameId":"valorant","host":"203.0.113.7","port":443,"avgMs":31.5,"lossPct":0,"samples":2}`); rec.Code != http.StatusOK {
		t.Fatalf("POST valid record: status %d body %s", rec.Code, rec.Body.String())
	}
	list = nil
	if err := json.Unmarshal(get("valorant").Body.Bytes(), &list); err != nil {
		t.Fatalf("decode GET after POST: %v", err)
	}
	if len(list) != 1 || list[0].Host != "203.0.113.7" || list[0].AvgMs != 31.5 {
		t.Fatalf("recorded history mismatch: %v", list)
	}

	// Failed probes are refused, so history can never be fabricated.
	if rec := post(`{"gameId":"valorant","host":"203.0.113.8","avgMs":0}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("POST avgMs=0 must fail, got %d", rec.Code)
	}
	// Hosts must be IP literals (no DNS on the API path).
	if rec := post(`{"gameId":"valorant","host":"example.com","avgMs":20}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("POST hostname must fail, got %d", rec.Code)
	}
	list = nil
	if err := json.Unmarshal(get("valorant").Body.Bytes(), &list); err != nil {
		t.Fatalf("decode GET after rejects: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("rejected records must not land in history: %v", list)
	}
}
