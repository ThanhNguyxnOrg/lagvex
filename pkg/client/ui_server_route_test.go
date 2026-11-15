package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteDetailEndpointShape(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/route-detail?gameId=valorant&regionId=asia-sg", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		MeasuredAt string `json:"measuredAt"`
		RegionName string `json:"regionName"`
		Connected  bool   `json:"connected"`
		Hops       []struct {
			ID       string `json:"id"`
			Measured bool   `json:"measured"`
			Reason   string `json:"reason"`
		} `json:"hops"`
		Notes []string `json:"notes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if res.MeasuredAt == "" {
		t.Fatalf("expected measuredAt to be set")
	}
	if res.RegionName == "" {
		t.Fatalf("expected the region name to be echoed back")
	}
	if len(res.Notes) == 0 {
		t.Fatalf("expected measurement notes")
	}

	seen := map[string]bool{}
	for _, hop := range res.Hops {
		seen[hop.ID] = true
		if !hop.Measured && hop.Reason == "" {
			t.Fatalf("unmeasured hop %q must explain why", hop.ID)
		}
	}
	for _, id := range []string{"pc", "gateway", "relay", "tunnel", "game-server"} {
		if !seen[id] {
			t.Fatalf("expected hop %q in the route detail, got %+v", id, res.Hops)
		}
	}

	// Without a connection the tunnel hop must be explicitly unmeasured.
	for _, hop := range res.Hops {
		if hop.ID == "tunnel" && hop.Measured {
			t.Fatalf("tunnel hop must not be measured while disconnected")
		}
	}
}

func TestRouteDetailRejectsUnknownRegion(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/route-detail?gameId=valorant&regionId=not-a-region", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown region, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRouteDetailRequiresRegionWithGame(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/route-detail?gameId=valorant", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a region, got %d: %s", rec.Code, rec.Body.String())
	}
}
