package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

// TestProbeRelaysSweepsContendersAndCaches covers the ExitLag-style sweep
// behaviour: an ids= subset probes only the named relays, an identical
// follow-up inside the TTL is served from the completed sweep, and force=1
// re-measures even when the cached sweep is fresh.
func TestProbeRelaysSweepsContendersAndCaches(t *testing.T) {
	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("profile manager: %v", err)
	}
	for _, rel := range []profiles.RelayEndpoint{
		{ID: "r1", Name: "R1", Endpoint: "127.0.0.1:1", Tier: "custom"},
		{ID: "r2", Name: "R2", Endpoint: "127.0.0.1:2", Tier: "custom"},
		{ID: "r3", Name: "R3", Endpoint: "127.0.0.1:3", Tier: "custom"},
	} {
		if err := pm.AddRelay(rel); err != nil {
			t.Fatalf("add relay %s: %v", rel.ID, err)
		}
	}
	server := NewUIServer(nil, pm, "")
	handler := server.setupRoutes()

	do := func(query string) []ProbeResult {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/probe-relays"+query, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
		}
		var results []ProbeResult
		if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		return results
	}

	// Relays without a PSK fail before any network I/O, so this sweep is
	// instant and never leaves the machine.
	first := do("?ids=r1,r2&samples=1")
	if len(first) != 2 {
		t.Fatalf("expected exactly the two named relays, got %d", len(first))
	}
	seen := map[string]bool{}
	for _, r := range first {
		if r.Error == "" {
			t.Errorf("relay %s: expected a PSK-required error, got none", r.RelayID)
		}
		seen[r.RelayID] = true
	}
	if !seen["r1"] || !seen["r2"] {
		t.Fatalf("ids filter missed a relay: %v", seen)
	}

	second := do("?ids=r1,r2&samples=1")
	if len(second) != 2 {
		t.Fatalf("cached sweep returned %d relays", len(second))
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("an identical sweep inside the TTL must be served from cache unchanged")
	}

	atCache := server.sweepAt
	time.Sleep(2 * time.Millisecond)
	forced := do("?ids=r1,r2&samples=1&force=1")
	if len(forced) != 2 {
		t.Fatalf("forced sweep returned %d relays", len(forced))
	}
	if !server.sweepAt.After(atCache) {
		t.Fatal("force=1 must re-run the sweep instead of serving the cached one")
	}
}
