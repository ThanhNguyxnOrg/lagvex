package client

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateTCPPingTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{name: "public IPv4", target: "1.1.1.1"},
		{name: "public IPv6", target: "2606:4700:4700::1111"},
		{name: "IPv4-mapped public IPv6", target: "::ffff:8.8.8.8"},
		{name: "empty", target: "", wantErr: true},
		{name: "hostname", target: "example.com", wantErr: true},
		{name: "IP with port", target: "1.1.1.1:443", wantErr: true},
		{name: "private RFC1918", target: "192.168.1.10", wantErr: true},
		{name: "private 10/8", target: "10.1.2.3", wantErr: true},
		{name: "loopback", target: "127.0.0.1", wantErr: true},
		{name: "IPv6 loopback", target: "::1", wantErr: true},
		{name: "link local", target: "169.254.10.10", wantErr: true},
		{name: "carrier NAT", target: "100.64.0.1", wantErr: true},
		{name: "benchmark range", target: "198.18.0.1", wantErr: true},
		{name: "reserved", target: "240.0.0.1", wantErr: true},
		{name: "multicast", target: "224.0.0.1", wantErr: true},
		{name: "unspecified", target: "0.0.0.0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateTCPPingTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTCPPingTarget(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestRunTCPPingAgainstLoopbackListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	result := runTCPPing(context.Background(), net.ParseIP("127.0.0.1"), port, 3, 500*time.Millisecond, 10*time.Millisecond)

	if !result.Measured {
		t.Fatalf("expected a measured result, got %+v", result)
	}
	if len(result.Samples) != 3 {
		t.Fatalf("samples = %d, want 3", len(result.Samples))
	}
	for _, sample := range result.Samples {
		if sample.Outcome != "open" {
			t.Fatalf("sample outcome = %q, want open (%+v)", sample.Outcome, sample)
		}
		if sample.RTTMs <= 0 {
			t.Fatalf("sample RTT = %v, want > 0", sample.RTTMs)
		}
	}
	if result.LossPct != 0 {
		t.Fatalf("loss = %v, want 0", result.LossPct)
	}
	if result.MinMs <= 0 || result.AvgMs < result.MinMs || result.MaxMs < result.AvgMs {
		t.Fatalf("inconsistent min/avg/max: %+v", result)
	}
}

func TestRunTCPPingClosedPortCountsAsMeasured(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	result := runTCPPing(context.Background(), net.ParseIP("127.0.0.1"), port, 1, 500*time.Millisecond, 10*time.Millisecond)
	if len(result.Samples) != 1 {
		t.Fatalf("samples = %d, want 1", len(result.Samples))
	}
	sample := result.Samples[0]
	if sample.Outcome != "closed" {
		t.Fatalf("outcome = %q, want closed (%+v)", sample.Outcome, sample)
	}
	if !result.Measured || sample.RTTMs <= 0 {
		t.Fatalf("RST response was not treated as a measurement: %+v", result)
	}
	if !strings.Contains(result.Note, "closed") {
		t.Fatalf("note does not explain the closed port: %q", result.Note)
	}
}

func TestClampTCPPingDefaults(t *testing.T) {
	port, attempts, timeout, err := clampTCPPingDefaults(0, 0, 0)
	if err != nil {
		t.Fatalf("defaults rejected: %v", err)
	}
	if port != tcpPingDefaultPort || attempts != tcpPingDefaultAttempts || timeout != tcpPingDefaultTimeout {
		t.Fatalf("defaults = %d/%d/%s", port, attempts, timeout)
	}

	bad := []struct{ port, attempts, timeout int }{
		{port: -1},
		{port: 70000},
		{attempts: -1},
		{attempts: tcpPingMaxAttempts + 1},
		{timeout: int(tcpPingMinTimeout/time.Millisecond) - 1},
		{timeout: int(tcpPingMaxTimeout/time.Millisecond) + 1},
	}
	for _, tt := range bad {
		if _, _, _, err := clampTCPPingDefaults(tt.port, tt.attempts, tt.timeout); err == nil {
			t.Fatalf("expected error for %+v", tt)
		}
	}
}

func TestTCPPingLimiter(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	limiter := tcpPingLimiter{now: func() time.Time { return now }}

	if _, err := limiter.allow(); err != nil {
		t.Fatalf("first run rejected: %v", err)
	}
	if wait, err := limiter.allow(); err == nil || wait <= 0 {
		t.Fatalf("cooldown not enforced: wait=%v err=%v", wait, err)
	}

	for i := 0; i < tcpPingHourlyLimit-1; i++ {
		now = now.Add(tcpPingCooldown + time.Second)
		if _, err := limiter.allow(); err != nil {
			t.Fatalf("run %d rejected: %v", i+2, err)
		}
	}
	if wait, err := limiter.allow(); err == nil || wait <= 0 {
		t.Fatalf("hourly limit not enforced: wait=%v err=%v", wait, err)
	}

	now = now.Add(time.Hour)
	if _, err := limiter.allow(); err != nil {
		t.Fatalf("new window rejected: %v", err)
	}
}

func TestHandleTCPPingRejectsUnsafeRequests(t *testing.T) {
	server := &UIServer{}

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{name: "wrong method", method: http.MethodGet, body: `{}`, wantStatus: http.StatusMethodNotAllowed},
		{name: "invalid JSON", method: http.MethodPost, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "private target", method: http.MethodPost, body: `{"target":"192.168.1.1"}`, wantStatus: http.StatusBadRequest},
		{name: "hostname target", method: http.MethodPost, body: `{"target":"example.com"}`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/tcping", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			server.handleTCPPing(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestTCPPingRouteIsGuarded(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := withLocalRequestGuard(server.setupRoutes())

	// The browser CSRF guard must cover the TCP ping endpoint: without the
	// intent header a cross-site page could otherwise drive local dials.
	req := httptest.NewRequest(http.MethodPost, "/api/tcping", strings.NewReader(`{"target":"1.1.1.1"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status without guard header = %d, want 403", rec.Code)
	}

	// With the header the request reaches the handler, which still refuses
	// non-public targets before any socket is opened.
	req = httptest.NewRequest(http.MethodPost, "/api/tcping", strings.NewReader(`{"target":"192.168.1.1"}`))
	req.Header.Set(dashboardRequestHeader, "1")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status for private target = %d, want 400 (body %s)", rec.Code, rec.Body.String())
	}
}

func TestHandleTCPPingRateLimit(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	server := &UIServer{}
	server.tcpPingLimiter.now = func() time.Time { return now }
	if _, err := server.tcpPingLimiter.allow(); err != nil {
		t.Fatalf("seed run rejected: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/tcping", strings.NewReader(`{"target":"1.1.1.1","attempts":1}`))
	rec := httptest.NewRecorder()
	server.handleTCPPing(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 (body %s)", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header")
	}
}
