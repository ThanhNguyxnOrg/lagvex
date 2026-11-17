package relay

import (
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

// Anchor-probe rate limits: they stop one client from turning a relay into a
// probe amplifier while still allowing occasional route-detail measurements.
const (
	anchorProbeMinInterval = 2 * time.Second
	anchorProbeMaxPerHour  = 60
)

// Session represents an authenticated connected client.
type Session struct {
	ID        uint64
	ClientID  uint64
	InnerIP   netip.Addr
	RemoteUDP atomic.Pointer[netip.AddrPort]
	Crypto    *protocol.SessionCrypto

	FECEncoder *protocol.FECEncoder
	FECDecoder *protocol.FECDecoder

	lastSeen  atomic.Int64 // Unix nanoseconds
	createdAt time.Time

	BytesUp          atomic.Uint64
	BytesDown        atomic.Uint64
	PacketsRecovered atomic.Uint64

	// Dual-path (Redundant-lite) state. GroupHash is 0 for single-path
	// sessions; non-zero means data frames carry the dual trailer and are
	// deduplicated against the shared connection group.
	GroupHash      uint32
	DualForwarded  atomic.Uint64
	DualSuppressed atomic.Uint64
	DualBadFrames  atomic.Uint64

	anchorProbeMu     sync.Mutex
	anchorProbeLast   time.Time
	anchorProbeWindow time.Time
	anchorProbeCount  int
}

// AllowAnchorProbe applies the per-session rate limit for relay-side probes.
func (s *Session) AllowAnchorProbe(now time.Time) bool {
	s.anchorProbeMu.Lock()
	defer s.anchorProbeMu.Unlock()

	if !s.anchorProbeLast.IsZero() && now.Sub(s.anchorProbeLast) < anchorProbeMinInterval {
		return false
	}
	if s.anchorProbeWindow.IsZero() || now.Sub(s.anchorProbeWindow) > time.Hour {
		s.anchorProbeWindow = now
		s.anchorProbeCount = 0
	}
	if s.anchorProbeCount >= anchorProbeMaxPerHour {
		return false
	}
	s.anchorProbeLast = now
	s.anchorProbeCount++
	return true
}

// NewSession initializes a new client session.
func NewSession(id, clientID uint64, innerIP netip.Addr, remote netip.AddrPort, crypto *protocol.SessionCrypto) *Session {
	s := &Session{
		ID:         id,
		ClientID:   clientID,
		InnerIP:    innerIP,
		Crypto:     crypto,
		FECEncoder: protocol.NewFECEncoder(protocol.DefaultFECEncoderConfig()),
		FECDecoder: protocol.NewFECDecoder(128),
		createdAt:  time.Now(),
	}
	s.RemoteUDP.Store(&remote)
	s.Touch()
	return s
}

// Touch refreshes the lastSeen timestamp.
func (s *Session) Touch() {
	s.lastSeen.Store(time.Now().UnixNano())
}

// LastSeenTime returns the last activity time.
func (s *Session) LastSeenTime() time.Time {
	return time.Unix(0, s.lastSeen.Load())
}

// UpdateRemote updates the client UDP endpoint if it changed (roaming support).
func (s *Session) UpdateRemote(newEndpoint netip.AddrPort) {
	curr := s.RemoteUDP.Load()
	if curr == nil || *curr != newEndpoint {
		s.RemoteUDP.Store(&newEndpoint)
	}
}
