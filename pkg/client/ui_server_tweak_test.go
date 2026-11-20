package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type tweakResponse struct {
	Applied     []string `json:"applied"`
	Failed      []string `json:"failed"`
	Preferences []string `json:"savedPreferences"`
	Status      string   `json:"status"`
	Success     bool     `json:"success"`
}

func postTweak(t *testing.T, handler http.Handler, body string) tweakResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/tweak", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/tweak status = %d body %s", rec.Code, rec.Body.String())
	}
	var out tweakResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode tweak response: %v", err)
	}
	return out
}

func contains(list []string, needle string) bool {
	for _, item := range list {
		if strings.Contains(strings.ToLower(item), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

// TestTweakWithoutFlagsChangesNothing guards the honesty rule: a request that
// asks for nothing must not silently flip the Smart L4 filter off and then
// report an optimisation as applied.
//
// The DNS cache flush is an intentional always-on part of "apply tweaks", so
// the invariant checked here is about the L4 filter, which is a real state
// change the user did not ask for.
func TestTweakWithoutFlagsChangesNothing(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	server := NewUIServer(engine, nil, "")
	handler := server.setupRoutes()

	out := postTweak(t, handler, `{}`)
	if contains(out.Applied, "L4") {
		t.Fatalf("an empty request must not touch the L4 filter, got %v", out.Applied)
	}
	if contains(out.Preferences, "MTU") {
		t.Fatalf("an empty request must not record preferences, got %v", out.Preferences)
	}
}

// TestTweakOnlyTouchesTheFlagsItMentions makes sure an unrelated tweak
// request leaves the other preferences and the L4 filter alone.
func TestTweakOnlyTouchesTheFlagsItMentions(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	server := NewUIServer(engine, nil, "")
	handler := server.setupRoutes()

	out := postTweak(t, handler, `{"mtuClamping":true}`)
	if contains(out.Applied, "L4") {
		t.Fatalf("a request that did not mention the L4 filter must not change it: %v", out.Applied)
	}
	if !contains(out.Preferences, "MTU clamping") {
		t.Fatalf("MTU clamping preference was not recorded: %v", out.Preferences)
	}
}

// TestTweakSmartL4FilterIsExplicit checks the inverse: when the request does
// mention the filter, the change is applied and reported.
func TestTweakSmartL4FilterIsExplicit(t *testing.T) {
	engine, err := NewEngine(nil)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	server := NewUIServer(engine, nil, "")
	handler := server.setupRoutes()

	out := postTweak(t, handler, `{"smartL4Filter":true}`)
	if !contains(out.Applied, "L4") {
		t.Fatalf("an explicit L4 request must be reported as applied: %v", out.Applied)
	}
}
