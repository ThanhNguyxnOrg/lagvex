package relay

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Squad signaling is served on a public HTTP port with an open CORS policy,
// so the relay treats it as an untrusted surface:
//
//   - room creation and room polling are rate limited per source IP,
//   - the number of live rooms is capped,
//   - room codes are cryptographically random instead of a small counter
//     range, so codes cannot be enumerated.
const (
	// defaultSquadRoomsCap bounds the memory a hostile client can pin.
	defaultSquadRoomsCap = 64

	// squadCreateWindow / squadCreateLimit bound room creation per IP.
	squadCreateWindow = time.Minute
	squadCreateLimit  = 10

	// squadPollWindow / squadPollLimit bound room lookups per IP.
	squadPollWindow = time.Minute
	squadPollLimit  = 60

	// squadRateEntriesCap bounds the rate-limit bookkeeping itself.
	squadRateEntriesCap = 4096
)

// squadRateEntry tracks one source IP's recent requests.
type squadRateEntry struct {
	windowStart time.Time
	count       int
}

// roomCodeAlphabet omits characters that are easy to confuse when a code is
// read aloud or typed from a screenshot.
const roomCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// newRoomCode returns a cryptographically random room code such as
// "LGVX-K7P2QD". It falls back to a time-seeded code only if the system CSPRNG
// fails, which never happens on a supported platform.
func newRoomCode() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("LGVX-%d", time.Now().UnixNano()%900000+100000)
	}
	var sb strings.Builder
	sb.WriteString("LGVX-")
	for _, b := range buf {
		sb.WriteByte(roomCodeAlphabet[int(b)%len(roomCodeAlphabet)])
	}
	return sb.String()
}

// clientIP extracts the caller's address for rate limiting. Proxy headers are
// ignored on purpose: they are attacker-controlled, and the relay is expected
// to sit directly on the public internet.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// allowSquadRequest reports whether the source IP may perform another request
// of the given kind inside the window.
func (s *Server) allowSquadRequest(ip string, window time.Duration, limit int) bool {
	now := time.Now()

	s.squadRateMu.Lock()
	defer s.squadRateMu.Unlock()

	if s.squadRate == nil {
		s.squadRate = make(map[string]*squadRateEntry)
	}
	entry, ok := s.squadRate[ip]
	if !ok || now.Sub(entry.windowStart) >= window {
		s.squadRate[ip] = &squadRateEntry{windowStart: now, count: 1}
		s.pruneSquadRateLocked(now)
		return true
	}
	if entry.count >= limit {
		return false
	}
	entry.count++
	return true
}

// pruneSquadRateLocked drops stale bookkeeping entries so the map cannot grow
// without bound. The caller must hold squadRateMu.
func (s *Server) pruneSquadRateLocked(now time.Time) {
	if len(s.squadRate) <= squadRateEntriesCap {
		return
	}
	for ip, entry := range s.squadRate {
		if now.Sub(entry.windowStart) >= squadPollWindow {
			delete(s.squadRate, ip)
		}
	}
	// Still over the cap (many fresh entries): drop arbitrary oldest entries.
	for ip, entry := range s.squadRate {
		if len(s.squadRate) <= squadRateEntriesCap {
			return
		}
		if now.Sub(entry.windowStart) >= squadCreateWindow {
			delete(s.squadRate, ip)
		}
	}
}

// squadRoomCap returns the configured maximum number of live rooms.
func (s *Server) squadRoomCap() int {
	if s.squadRoomsCap > 0 {
		return s.squadRoomsCap
	}
	return defaultSquadRoomsCap
}

// insertRoomLocked inserts a room, evicting the oldest room when the cap is
// hit. The caller must hold squadRoomsMu.
func (s *Server) insertRoomLocked(room *RelaySquadRoom) {
	cap := s.squadRoomCap()
	for len(s.squadRooms) >= cap {
		var oldestCode string
		var oldest time.Time
		for code, existing := range s.squadRooms {
			existing.mu.Lock()
			created := existing.CreatedAt
			existing.mu.Unlock()
			if oldestCode == "" || created.Before(oldest) {
				oldestCode = code
				oldest = created
			}
		}
		if oldestCode == "" {
			break
		}
		delete(s.squadRooms, oldestCode)
	}
	s.squadRooms[room.Code] = room
}

// squadRateLimitMiddleware applies the per-IP limits to the signaling routes.
func (s *Server) squadRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		switch r.URL.Path {
		case "/squad/create":
			if !s.allowSquadRequest(ip, squadCreateWindow, squadCreateLimit) {
				http.Error(w, `{"error":"too many squad rooms created, try again later"}`, http.StatusTooManyRequests)
				return
			}
		case "/squad/room":
			if !s.allowSquadRequest(ip, squadPollWindow, squadPollLimit) {
				http.Error(w, `{"error":"too many room lookups, slow down"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
