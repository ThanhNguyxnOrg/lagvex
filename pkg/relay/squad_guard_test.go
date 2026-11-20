package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func newSquadTestServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	srv, err := NewServer(Config{
		ListenAddr:  "127.0.0.1:0",
		HTTPAddr:    "127.0.0.1:0",
		TunName:     "lagvexguard",
		Subnet:      netip.MustParsePrefix("10.78.0.0/24"),
		PSK:         []byte("guard-psk-0123456789ab"),
		MTU:         1400,
		IdleTimeout: 30,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/squad/create", srv.handleSquadCreate)
	mux.HandleFunc("/squad/join", srv.handleSquadJoin)
	mux.HandleFunc("/squad/room", srv.handleSquadRoom)
	mux.HandleFunc("/squad/heartbeat", srv.handleSquadHeartbeat)
	mux.HandleFunc("/squad/leave", srv.handleSquadLeave)
	return srv, srv.squadRateLimitMiddleware(mux)
}

// TestSquadRoomCodeIsUnguessable guards against the old counter-based code
// (LGVX-<4 digits>), which could be enumerated to read other squads' rooms.
func TestSquadRoomCodeIsUnguessable(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		code := newRoomCode()
		if !strings.HasPrefix(code, "LGVX-") {
			t.Fatalf("room code %q lost the LGVX prefix", code)
		}
		if len(code) != len("LGVX-")+6 {
			t.Fatalf("room code %q has an unexpected length", code)
		}
		if seen[code] {
			t.Fatalf("room code %q was generated twice", code)
		}
		seen[code] = true
	}
}

// TestSquadCreateIsRateLimited makes sure a single source cannot spam rooms.
func TestSquadCreateIsRateLimited(t *testing.T) {
	_, handler := newSquadTestServer(t)

	created := 0
	for i := 0; i < squadCreateLimit+5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/squad/create",
			strings.NewReader(fmt.Sprintf(`{"name":"player-%d","host":"127.0.0.1:4433"}`, i)))
		req.RemoteAddr = "203.0.113.9:5555"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			created++
			continue
		}
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: status %d body %s", i, rec.Code, rec.Body.String())
		}
	}
	if created != squadCreateLimit {
		t.Fatalf("created %d rooms, want exactly the limit %d", created, squadCreateLimit)
	}
}

// TestSquadRoomPollingIsRateLimited blocks code enumeration through the
// room lookup endpoint.
func TestSquadRoomPollingIsRateLimited(t *testing.T) {
	_, handler := newSquadTestServer(t)

	allowed := 0
	for i := 0; i < squadPollLimit+5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/squad/room?code=LGVX-ABCDEF", nil)
		req.RemoteAddr = "203.0.113.10:5555"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			continue
		}
		allowed++
	}
	if allowed != squadPollLimit {
		t.Fatalf("allowed %d lookups, want the limit %d", allowed, squadPollLimit)
	}
}

// TestSquadRoomsAreCapped keeps the room map bounded when a hostile client
// keeps creating rooms.
func TestSquadRoomsAreCapped(t *testing.T) {
	srv, handler := newSquadTestServer(t)
	srv.squadRoomsCap = 4

	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/squad/create",
			strings.NewReader(fmt.Sprintf(`{"name":"player-%d","host":"127.0.0.1:4433"}`, i)))
		req.RemoteAddr = fmt.Sprintf("198.51.100.%d:5555", i%200+1)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("create %d: status %d body %s", i, rec.Code, rec.Body.String())
		}
	}

	srv.squadRoomsMu.RLock()
	count := len(srv.squadRooms)
	srv.squadRoomsMu.RUnlock()
	if count > 4 {
		t.Fatalf("room count = %d, want at most 4", count)
	}
}

// TestSquadRoomAcceptsCodeFromBody keeps the endpoint usable from scripts
// that send JSON instead of a query parameter.
func TestSquadRoomAcceptsCodeFromBody(t *testing.T) {
	_, handler := newSquadTestServer(t)

	create := httptest.NewRequest(http.MethodPost, "/squad/create",
		strings.NewReader(`{"name":"host-player","host":"127.0.0.1:4433"}`))
	create.RemoteAddr = "203.0.113.20:5555"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, create)
	if rec.Code != http.StatusOK {
		t.Fatalf("create: status %d body %s", rec.Code, rec.Body.String())
	}
	var created struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Code == "" {
		t.Fatal("create did not return a room code")
	}

	room := httptest.NewRequest(http.MethodPost, "/squad/room",
		strings.NewReader(fmt.Sprintf(`{"code":%q}`, created.Code)))
	room.RemoteAddr = "203.0.113.21:5555"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, room)
	if rec.Code != http.StatusOK {
		t.Fatalf("room lookup: status %d body %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode room response: %v", err)
	}
	if got.Code != created.Code {
		t.Fatalf("room code = %q, want %q", got.Code, created.Code)
	}
}
