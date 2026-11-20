package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postRelayProbe(t *testing.T, u *UIServer, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/relay-tcping", strings.NewReader(body))
	rec := httptest.NewRecorder()
	u.handleRelayGameServerProbe(rec, req)
	return rec
}

// TestRelayGameServerProbeNeedsTunnel: without a connected relay the segment
// cannot exist, so the handler must refuse instead of measuring nothing and
// calling it 0 ms.
func TestRelayGameServerProbeNeedsTunnel(t *testing.T) {
	u := &UIServer{}
	rec := postRelayProbe(t, u, `{"host":"203.0.113.7","port":443}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "no connected relay") {
		t.Fatalf("body = %q, want an explicit reason", rec.Body.String())
	}
}

// TestRelayGameServerProbeValidatesTarget reuses the local tool's rules:
// no hostnames, no private or loopback addresses, port in range.
func TestRelayGameServerProbeValidatesTarget(t *testing.T) {
	u := &UIServer{}
	for _, body := range []string{
		`{"host":"example.com","port":443}`,
		`{"host":"127.0.0.1","port":443}`,
		`{"host":"10.0.0.4","port":443}`,
		`{"host":"203.0.113.7","port":70000}`,
		`{"host":"203.0.113.7","port":0}`,
		`{"host":"","port":443}`,
		`not-json`,
	} {
		rec := postRelayProbe(t, u, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", body, rec.Code)
		}
		var out struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s: body not JSON: %v", body, err)
		}
		if out.Error == "" {
			t.Fatalf("%s: error message missing", body)
		}
	}
}

// TestRelayGameServerProbeRejectsNonPOST keeps the API surface consistent.
func TestRelayGameServerProbeRejectsNonPOST(t *testing.T) {
	u := &UIServer{}
	req := httptest.NewRequest(http.MethodGet, "/api/relay-tcping", nil)
	rec := httptest.NewRecorder()
	u.handleRelayGameServerProbe(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
