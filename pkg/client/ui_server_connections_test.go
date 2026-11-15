package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
