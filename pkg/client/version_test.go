package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestBuildVersionIsReported guards dashboard honesty: the footer version must
// come from the running binary, not a hardcoded release number.
func TestBuildVersionIsReported(t *testing.T) {
	if got := BuildVersion(); got == "" {
		t.Fatal("default build version is empty")
	}

	SetBuildVersion("")
	if got := BuildVersion(); got != "dev" {
		t.Fatalf("empty SetBuildVersion changed version to %q", got)
	}

	SetBuildVersion("1.8.0-demo")
	t.Cleanup(func() { SetBuildVersion("dev") })

	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/system-info", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("system-info status = %d, want 200", rec.Code)
	}

	var res struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode system-info: %v", err)
	}
	if res.Version != "1.8.0-demo" {
		t.Fatalf("system-info version = %q, want the running build version", res.Version)
	}
}
