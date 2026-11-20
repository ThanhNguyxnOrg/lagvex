package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// postSquad runs one squad request against the handler and fails the test if
// the handler does not answer in time. This catches lock mistakes that would
// otherwise only show up as a hung dashboard.
func postSquad(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		done <- rec
	}()
	select {
	case rec := <-done:
		return rec
	case <-time.After(5 * time.Second):
		t.Fatalf("squad request %s did not return (lock deadlock?)", path)
		return nil
	}
}

// TestSquadCreateDoesNotDeadlock covers the local fallback path (no relay
// connected): creating several rooms must not block on the room mutex.
func TestSquadCreateDoesNotDeadlock(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	for i := 0; i < 3; i++ {
		rec := postSquad(t, handler, "/api/squad/create",
			`{"host":"HostPlayer","game":"Valorant","ping":20}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("create %d: status %d body %s", i, rec.Code, rec.Body.String())
		}
		var out struct {
			Success bool `json:"success"`
			Room    struct {
				Code string `json:"code"`
			} `json:"room"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode create response: %v", err)
		}
		if !out.Success || out.Room.Code == "" {
			t.Fatalf("create %d returned no room code: %s", i, rec.Body.String())
		}
		if !strings.HasPrefix(out.Room.Code, "LGVX-") {
			t.Fatalf("room code %q lost the LGVX prefix", out.Room.Code)
		}
	}
}

// TestSquadRoomCapIsEnforced keeps the local fallback map bounded.
func TestSquadRoomCapIsEnforced(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	for i := 0; i < squadRoomsCap+10; i++ {
		rec := postSquad(t, handler, "/api/squad/create",
			`{"host":"HostPlayer","game":"Valorant"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("create %d: status %d body %s", i, rec.Code, rec.Body.String())
		}
	}

	squadRoomsMu.Lock()
	count := len(squadRooms)
	squadRoomsMu.Unlock()
	if count > squadRoomsCap {
		t.Fatalf("local room count = %d, want at most %d", count, squadRoomsCap)
	}
}
