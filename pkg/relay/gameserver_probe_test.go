package relay

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// stubProbeDial replaces the relay's TCP dialer so handler tests never send a
// packet to a real host. Each "dial" completes in 5ms.
func stubProbeDial(t *testing.T) {
	t.Helper()
	original := probeDial
	probeDial = func(network, address string, timeout time.Duration) (net.Conn, error) {
		time.Sleep(5 * time.Millisecond)
		return &fakeConn{}, nil
	}
	t.Cleanup(func() { probeDial = original })
}

type fakeConn struct{}

func (c *fakeConn) Read(b []byte) (int, error)       { return 0, nil }
func (c *fakeConn) Write(b []byte) (int, error)      { return len(b), nil }
func (c *fakeConn) Close() error                     { return nil }
func (c *fakeConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (c *fakeConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (c *fakeConn) SetDeadline(time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(time.Time) error { return nil }

func postJSON(t *testing.T, srv *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/probe/gameserver", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.handleProbeGameServer(rec, req)
	return rec
}

// TestProbeGameServerRejectsNonPublicHosts: the relay must never dial private,
// loopback or bogon addresses on request from an anonymous caller.
func TestProbeGameServerRejectsNonPublicHosts(t *testing.T) {
	srv := &Server{}
	for _, body := range []string{
		`{"host":"127.0.0.1","port":443}`,
		`{"host":"10.0.0.5","port":443}`,
		`{"host":"192.168.1.10","port":443}`,
		`{"host":"169.254.169.254","port":80}`,
	} {
		rec := postJSON(t, srv, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", body, rec.Code)
		}
	}
	for _, body := range []string{
		`{"host":"203.0.113.7","port":0}`,
		`{"host":"203.0.113.7","port":70000}`,
		`{"host":"not-an-ip","port":443}`,
		`{}`,
	} {
		rec := postJSON(t, srv, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", body, rec.Code)
		}
	}
}

// TestProbeGameServerRateLimited: one probe per two seconds per source IP,
// enforced by a limiter of its own so squad polling cannot starve it.
func TestProbeGameServerRateLimited(t *testing.T) {
	stubProbeDial(t)
	srv := &Server{}
	first := postJSON(t, srv, `{"host":"203.0.113.7","port":443}`)
	if first.Code != http.StatusOK {
		t.Fatalf("first probe status = %d, want 200", first.Code)
	}
	if first.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q, want application/json", first.Header().Get("Content-Type"))
	}
	second := postJSON(t, srv, `{"host":"203.0.113.7","port":443}`)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second probe status = %d, want 429", second.Code)
	}
	if !strings.Contains(second.Body.String(), "wait 2 seconds") {
		t.Fatalf("body = %q, want an explicit retry hint", second.Body.String())
	}
}

// TestProbeGameServerResponseShape checks the payload names its own technique
// and reports the measured RTT from the (stubbed) handshakes.
func TestProbeGameServerResponseShape(t *testing.T) {
	stubProbeDial(t)
	srv := &Server{}
	rec := postJSON(t, srv, `{"host":"203.0.113.7","port":443}`)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["probeMethod"] != "tcp-handshake" {
		t.Fatalf("probeMethod = %v, want tcp-handshake", body["probeMethod"])
	}
	if body["measuredBy"] != "relay" {
		t.Fatalf("measuredBy = %v, want relay", body["measuredBy"])
	}
	if samples, _ := body["samples"].(float64); int(samples) != probeGameServerSamples {
		t.Fatalf("samples = %v, want %d", body["samples"], probeGameServerSamples)
	}
	// Every stubbed handshake succeeded, so an RTT must be reported.
	rtt, ok := body["rttMs"].(float64)
	if !ok || rtt <= 0 {
		t.Fatalf("rttMs = %v, want a positive measured value", body["rttMs"])
	}
	if loss, _ := body["lossPct"].(float64); loss != 0 {
		t.Fatalf("lossPct = %v, want 0 for successful handshakes", body["lossPct"])
	}
}

// TestProbeGameServerBodyLimit: a public endpoint must not buffer an
// unbounded request body. Anything past the middleware cap fails to decode,
// so the answer is a 400 rather than memory pressure.
func TestProbeGameServerBodyLimit(t *testing.T) {
	srv := &Server{}
	handler := srv.squadRateLimitMiddleware(http.HandlerFunc(srv.handleProbeGameServer))

	big := `{"host":"203.0.113.7","port":443,"pad":"` + strings.Repeat("a", 64<<10) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/probe/gameserver", strings.NewReader(big))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for an oversized body", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid JSON body") {
		t.Fatalf("body = %q, want the decode error", rec.Body.String())
	}
}

// TestProbeGameServerCORSAndOptions keeps the open-port contract intact: the
// body cap applies to requests with bodies, not to CORS preflights.
func TestProbeGameServerCORSAndOptions(t *testing.T) {
	srv := &Server{}
	handler := srv.squadRateLimitMiddleware(http.HandlerFunc(srv.handleProbeGameServer))

	req := httptest.NewRequest(http.MethodOptions, "/probe/gameserver", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("OPTIONS status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
}

// TestProbeRateLimiterIsPerEndpoint verifies the probe limiter is separate
// from the squad limiter: a squad-poll window must not consume probe slots.
func TestProbeRateLimiterIsPerEndpoint(t *testing.T) {
	l := probeRateLimiter{}
	if !l.allow("1.2.3.4", probeGameServerWindow, probeGameServerLimit) {
		t.Fatal("first probe request was denied")
	}
	if l.allow("1.2.3.4", probeGameServerWindow, probeGameServerLimit) {
		t.Fatal("second probe request inside the window was allowed")
	}
	// Different IP is unaffected.
	if !l.allow("5.6.7.8", probeGameServerWindow, probeGameServerLimit) {
		t.Fatal("another IP was denied by the first IP's window")
	}
	// Bookkeeping stays bounded.
	for i := 0; i < probeRateEntriesCap+50; i++ {
		l.allow(string(rune(i)), probeGameServerWindow, probeGameServerLimit)
	}
	l.mu.Lock()
	n := len(l.entries)
	l.mu.Unlock()
	if n > probeRateEntriesCap {
		t.Fatalf("limiter kept %d entries, want at most %d", n, probeRateEntriesCap)
	}
}
