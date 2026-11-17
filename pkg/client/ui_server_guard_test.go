package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalRequestGuard(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := withLocalRequestGuard(next)

	tests := []struct {
		name       string
		method     string
		origin     string
		header     string
		wantStatus int
	}{
		{name: "state change without header", method: http.MethodPost, wantStatus: http.StatusForbidden},
		{name: "state change with local origin but no header", method: http.MethodPost, origin: "http://127.0.0.1:18888", wantStatus: http.StatusForbidden},
		{name: "state change with local origin and header", method: http.MethodPost, origin: "http://127.0.0.1:18888", header: "1", wantStatus: http.StatusNoContent},
		{name: "localhost dev origin with header", method: http.MethodPost, origin: "http://localhost:5173", header: "1", wantStatus: http.StatusNoContent},
		{name: "internet origin is rejected even with header", method: http.MethodPost, origin: "https://evil.example", header: "1", wantStatus: http.StatusForbidden},
		{name: "opaque origin is rejected", method: http.MethodPost, origin: "null", header: "1", wantStatus: http.StatusForbidden},
		{name: "read-only request is not blocked", method: http.MethodGet, origin: "https://evil.example", wantStatus: http.StatusNoContent},
		{name: "cli request with header and no origin", method: http.MethodPost, header: "1", wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.header != "" {
				req.Header.Set(dashboardRequestHeader, tt.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
