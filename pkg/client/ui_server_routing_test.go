package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

func TestConnectRejectsMissingRegion(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	body := map[string]any{
		"gameId":        "valorant",
		"relayEndpoint": "127.0.0.1:1",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/connect", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing region, got %d: %s", rec.Code, rec.Body.String())
	}
	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["success"] != false {
		t.Fatalf("expected success=false, got %+v", res)
	}
}

func TestConnectRejectsUnknownRegion(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	body := map[string]any{
		"gameId":        "valorant",
		"regionId":      "not-a-real-region",
		"relayEndpoint": "127.0.0.1:1",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/connect", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown region, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConnectWithoutEngineReportsFailure(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	body := map[string]any{
		"gameId":        "valorant",
		"regionId":      "asia-sg",
		"relayEndpoint": "127.0.0.1:1",
		"psk":           "test",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/connect", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 with nil engine, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConnectRequiresRelayPSK(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	body := map[string]any{
		"gameId":        "valorant",
		"regionId":      "asia-sg",
		"relayEndpoint": "203.0.113.77:4433",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/connect", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a PSK, got %d: %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("PSK")) {
		t.Fatalf("expected a PSK error, got %s", rec.Body.String())
	}
}

func TestAddRelayRequiresPSK(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	body := map[string]any{
		"id":       "no-psk-relay",
		"name":     "No PSK Relay",
		"endpoint": "203.0.113.78:4433",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/add-relay", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a PSK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRelaysNeverExposePSK(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	addBody := map[string]any{
		"id":       "test-secret-relay",
		"name":     "Test Secret Relay",
		"endpoint": "198.51.100.10:4433",
		"psk":      "super-secret-psk",
	}
	b, _ := json.Marshal(addBody)
	addReq := httptest.NewRequest(http.MethodPost, "/api/add-relay", bytes.NewReader(b))
	addRec := httptest.NewRecorder()
	handler.ServeHTTP(addRec, addReq)
	if addRec.Code != http.StatusOK {
		t.Fatalf("add relay failed: %d %s", addRec.Code, addRec.Body.String())
	}
	if bytes.Contains(addRec.Body.Bytes(), []byte("super-secret-psk")) {
		t.Fatalf("add-relay response leaked PSK: %s", addRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/relays", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("relays failed: %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("super-secret-psk")) {
		t.Fatalf("/api/relays leaked PSK: %s", rec.Body.String())
	}
}

func TestAddGameValidatesCIDRs(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	invalid := map[string]any{
		"id":           "custom-bad",
		"name":         "Bad Custom Game",
		"processNames": []string{"bad.exe"},
		"regions": []map[string]any{
			{"id": "global", "name": "Global", "cidrs": []string{"not-a-cidr"}},
		},
	}
	b, _ := json.Marshal(invalid)
	req := httptest.NewRequest(http.MethodPost, "/api/add-game", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid CIDR rejection, got %d: %s", rec.Code, rec.Body.String())
	}

	valid := map[string]any{
		"id":           "custom-good",
		"name":         "Good Custom Game",
		"processNames": []string{"good.exe"},
		"regions": []map[string]any{
			{"id": "global", "name": "Global", "cidrs": []string{"203.0.113.0/24"}},
		},
	}
	b, _ = json.Marshal(valid)
	req = httptest.NewRequest(http.MethodPost, "/api/add-game", bytes.NewReader(b))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected valid game acceptance, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestOptimizeProfileIsPreviewOnly(t *testing.T) {
	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("new profile manager: %v", err)
	}
	server := NewUIServer(nil, pm, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/optimize-profile?gameId=valorant", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["applied"] != false || res["status"] != "preview" {
		t.Fatalf("optimize-profile must not claim changes: %+v", res)
	}
	if count, ok := res["subnetsCount"].(float64); !ok || count <= 0 {
		t.Fatalf("expected real CIDR count, got %+v", res["subnetsCount"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/optimize-profile", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without gameId, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/optimize-profile?gameId=not-a-game", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown game, got %d", rec.Code)
	}
}

func TestGameTelemetryNeverFabricatesMeasurements(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/game-telemetry", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var res map[string]struct {
		BaselinePing float64 `json:"baselinePing"`
		AccelPing    float64 `json:"accelPing"`
		Measured     bool    `json:"measured"`
		Source       string  `json:"source"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res) == 0 {
		t.Fatal("expected telemetry entries for the embedded catalog")
	}
	for gameID, item := range res {
		if !item.Measured {
			if item.BaselinePing != 0 || item.AccelPing != 0 {
				t.Fatalf("%s reported unmeasured ping values: %+v", gameID, item)
			}
			continue
		}
		if item.Source == "" {
			t.Fatalf("%s reported a measurement without a source", gameID)
		}
		if item.BaselinePing <= 0 && item.AccelPing <= 0 {
			t.Fatalf("%s marked measured without a positive value: %+v", gameID, item)
		}
	}
}

func TestGameTelemetryWithoutCatalogReturnsEmpty(t *testing.T) {
	server := &UIServer{}
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/game-telemetry", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected empty telemetry without a catalog, got %+v", res)
	}
}

func TestDetectedProcessesReportsCatalogGames(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/detected-processes", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		ScannedAt string `json:"scannedAt"`
		Games     []struct {
			GameID    string   `json:"gameId"`
			Processes []string `json:"processes"`
			Running   bool     `json:"running"`
		} `json:"games"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res.Games) == 0 {
		t.Fatalf("expected game statuses, got none")
	}
	if res.ScannedAt == "" {
		t.Fatalf("expected scannedAt to be set")
	}
	for _, g := range res.Games {
		if g.GameID == "" {
			t.Fatalf("game status missing gameId: %+v", g)
		}
	}
}
