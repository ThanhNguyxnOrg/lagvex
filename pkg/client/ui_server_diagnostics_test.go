package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

func newDiagnosticsTestServer(t *testing.T) *UIServer {
	t.Helper()
	server := NewUIServer(nil, nil, "")
	server.configStore = NewConfigStore(filepath.Join(t.TempDir(), "user_config.json"))
	return server
}

func TestDiagnosticsExportExcludesPSK(t *testing.T) {
	server := newDiagnosticsTestServer(t)
	relay := profiles.RelayEndpoint{
		ID:       "secret-relay",
		Name:     "Secret Relay",
		Endpoint: "203.0.113.9:4433",
		PSK:      "super-secret-psk-value",
		Tier:     "custom",
	}
	if err := server.profileMgr.AddRelay(relay); err != nil {
		t.Fatalf("add relay: %v", err)
	}
	server.configStore.AddCustomRelay(relay)

	handler := server.setupRoutes()
	req := httptest.NewRequest(http.MethodGet, "/api/diagnostics/export", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "super-secret-psk-value") {
		t.Fatalf("snapshot leaked a relay PSK")
	}
	if !strings.Contains(body, "secret-relay") {
		t.Fatalf("snapshot should list the relay without its PSK, got %s", body)
	}
	if !strings.Contains(body, `"psksIncluded":false`) {
		t.Fatalf("snapshot should declare that PSKs are excluded, got %s", body)
	}

	var snap struct {
		SnapshotVersion int `json:"snapshotVersion"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.SnapshotVersion != diagnosticsSnapshotVersion {
		t.Fatalf("expected snapshotVersion %d, got %d", diagnosticsSnapshotVersion, snap.SnapshotVersion)
	}
}

func TestDiagnosticsImportMergesSettings(t *testing.T) {
	server := newDiagnosticsTestServer(t)
	handler := server.setupRoutes()

	snapshot := map[string]any{
		"snapshotVersion": 1,
		"app":             "lagvex-client",
		"psksIncluded":    false,
		"config": map[string]any{
			"nickname": "Tester#4321",
			"tweaks": map[string]bool{
				"tcpNoDelay": true,
				"bogus":      true,
			},
			"customGames": []map[string]any{
				{
					"id": "mygame", "name": "My Game", "category": "FPS",
					"processNames": []string{"mygame.exe"},
					"regions": []map[string]any{{
						"id": "r1", "name": "Region One", "cidrs": []string{"203.0.113.0/24"},
					}},
				},
				{"id": "nocidr", "name": "No CIDR"},
			},
			"customRelays": []map[string]any{{
				"id": "r", "name": "R", "endpoint": "198.51.100.1:4433",
			}},
		},
	}
	payload, _ := json.Marshal(snapshot)
	req := httptest.NewRequest(http.MethodPost, "/api/diagnostics/import", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result diagnosticsImportResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode import result: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected warnings for the unknown tweak, missing CIDRs and skipped relays")
	}

	cfg := server.configStore.Get()
	if cfg.Nickname != "Tester#4321" {
		t.Fatalf("nickname not applied, got %q", cfg.Nickname)
	}
	if _, ok := cfg.Tweaks["bogus"]; ok {
		t.Fatalf("unknown tweak key must not be stored")
	}
	if _, ok := server.profileMgr.FindGameByID("mygame"); !ok {
		t.Fatalf("custom game was not added to the profile catalog")
	}
	if _, ok := server.profileMgr.FindGameByID("nocidr"); ok {
		t.Fatalf("a custom game without CIDRs must be skipped")
	}
}

func TestDiagnosticsImportRejectsUnknownVersion(t *testing.T) {
	server := newDiagnosticsTestServer(t)
	handler := server.setupRoutes()

	payload := []byte(`{"snapshotVersion":99,"config":{"nickname":"x"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/diagnostics/import", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unsupported version, got %d: %s", rec.Code, rec.Body.String())
	}
}
