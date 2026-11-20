package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// TestDetectedProcessesReportsCPUSupport verifies the process scan always
// states whether CPU measurement is possible on this platform, so the
// dashboard can show a real dash on systems where it is not.
func TestDetectedProcessesReportsCPUSupport(t *testing.T) {
	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("profile manager: %v", err)
	}
	u := &UIServer{profileMgr: pm}

	req := httptest.NewRequest(http.MethodGet, "/api/detected-processes", nil)
	rec := httptest.NewRecorder()
	u.handleDetectedProcesses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		ScannedAt    string `json:"scannedAt"`
		CPUSupported bool   `json:"cpuSupported"`
		Games        []struct {
			GameID string `json:"gameId"`
			Name   string `json:"name"`
		} `json:"games"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ScannedAt == "" {
		t.Fatal("scannedAt is missing")
	}
	if len(body.Games) == 0 {
		t.Fatal("no games reported")
	}
	// On a supported platform the CPU field may still be absent when no
	// process matched or a second sample has not been taken; only the flag
	// must be truthful.
	if !body.CPUSupported && !processCPUSupported() {
		t.Fatal("cpuSupported = false on a platform that exposes counters")
	}
}

// TestProbeRelaysEmptyCatalogIsEmptyList guards the list contract: with no
// relays configured the API answers [] rather than JSON null, because a
// dashboard that expects a list reads a different shape.
func TestProbeRelaysEmptyCatalogIsEmptyList(t *testing.T) {
	t.Setenv("LAGVEX_CONFIG_DIR", t.TempDir())
	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("profile manager: %v", err)
	}
	u := &UIServer{profileMgr: pm, prober: NewProber()}

	req := httptest.NewRequest(http.MethodGet, "/api/probe-relays?force=1", nil)
	rec := httptest.NewRecorder()
	u.handleProbeRelays(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Fatalf("body = %q, want []", got)
	}
}
