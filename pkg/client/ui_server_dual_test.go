package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleDualPathToggle(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	server := &UIServer{engine: engine}

	// Wrong method.
	req := httptest.NewRequest(http.MethodGet, "/api/dual-path/toggle", nil)
	rec := httptest.NewRecorder()
	server.handleDualPathToggle(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", rec.Code)
	}

	// Enable while disconnected: honest inactive state.
	req = httptest.NewRequest(http.MethodPost, "/api/dual-path/toggle",
		strings.NewReader(`{"enabled":true}`))
	rec = httptest.NewRecorder()
	server.handleDualPathToggle(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST status = %d (body %s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Enabled  bool            `json:"enabled"`
		Active   bool            `json:"active"`
		DualPath *DualPathStatus `json:"dualPath"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Enabled || resp.Active {
		t.Fatalf("response = enabled %v active %v, want true/false", resp.Enabled, resp.Active)
	}
	if resp.DualPath == nil || !strings.Contains(resp.DualPath.Reason, "not connected") {
		t.Fatalf("dualPath = %+v, want an honest not-connected reason", resp.DualPath)
	}

	// Disable.
	req = httptest.NewRequest(http.MethodPost, "/api/dual-path/toggle",
		strings.NewReader(`{"enabled":false}`))
	rec = httptest.NewRecorder()
	server.handleDualPathToggle(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("disable status = %d", rec.Code)
	}
	if engine.IsDualPathEnabled() {
		t.Fatal("engine intent must clear after disable")
	}

	// No engine: explicit 503.
	bare := &UIServer{}
	req = httptest.NewRequest(http.MethodPost, "/api/dual-path/toggle", strings.NewReader(`{"enabled":true}`))
	rec = httptest.NewRecorder()
	bare.handleDualPathToggle(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("no-engine status = %d, want 503", rec.Code)
	}
}
