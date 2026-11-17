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

// dualAckInterval bounds how often a relay reports dual-path progress to one
// client: at most 50 ACKs per second per session.
const dualAckInterval = 20 * time.Millisecond

// dualAckHolesCap bounds the out-of-order set behind the contiguous ACK
// watermark. When it overflows, the oldest hole is evicted; the resulting
// stall self-heals because the client retransmits that sequence and the
// suppressed duplicate re-enters tracking.
const dualAckHolesCap = 512

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

	// Dual ACK bookkeeping: the highest *contiguous* dual sequence forwarded
	// to the game server, reported to the client at most once per interval.
	// Contiguity matters: a highest-seq ACK with gaps below it would let the
	// client purge frames it still needs to retransmit.
	dualAckMu     sync.Mutex
	dualAckContig uint32
	dualAckHoles  map[uint32]struct{}
	dualAckHoleQ  []uint32
	dualAckLast   time.Time

	anchorProbeMu     sync.Mutex
	anchorProbeLast   time.Time
	anchorProbeWindow time.Time
	anchorProbeCount  int
}

// TrackDualForwarded records a dual sequence that reached the game server:
// the first forwarded copy, or a suppressed duplicate (proof it was forwarded
// earlier). It advances the contiguous ACK watermark when possible.
func (s *Session) TrackDualForwarded(seq uint32) {
	if seq == 0 {
		return // sequence 0 is never assigned
	}
	s.dualAckMu.Lock()
	defer s.dualAckMu.Unlock()

	if seq <= s.dualAckContig {
		return
	}
	if seq == s.dualAckContig+1 {
		s.dualAckContig = seq
		for {
			next := s.dualAckContig + 1
			if _, ok := s.dualAckHoles[next]; !ok {
				break
			}
			delete(s.dualAckHoles, next)
			s.removeHoleQ(next)
			s.dualAckContig = next
		}
		return
	}
	if _, ok := s.dualAckHoles[seq]; ok {
		return
	}
	if len(s.dualAckHoles) >= dualAckHolesCap {
		oldest := s.dualAckHoleQ[0]
		s.dualAckHoleQ = s.dualAckHoleQ[1:]
		delete(s.dualAckHoles, oldest)
	}
	s.dualAckHoles[seq] = struct{}{}
	s.dualAckHoleQ = append(s.dualAckHoleQ, seq)
}

// removeHoleQ drops one sequence from the FIFO hole queue. Callers hold
// dualAckMu.
func (s *Session) removeHoleQ(seq uint32) {
	for i, candidate := range s.dualAckHoleQ {
		if candidate == seq {
			s.dualAckHoleQ = append(s.dualAckHoleQ[:i], s.dualAckHoleQ[i+1:]...)
			return
		}
	}
}

// TakeDualAck returns the contiguous forwarded dual sequence when an ACK may
// be emitted now, respecting dualAckInterval. The value is 0 before the first
// forwarded dual frame; callers must not send an ACK in that case.
func (s *Session) TakeDualAck(now time.Time) (contiguous uint32, ok bool) {
	s.dualAckMu.Lock()
	defer s.dualAckMu.Unlock()
	if s.dualAckContig == 0 {
		return 0, false
	}
	if !s.dualAckLast.IsZero() && now.Sub(s.dualAckLast) < dualAckInterval {
		return 0, false
	}
	s.dualAckLast = now
	return s.dualAckContig, true
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
		ID:           id,
		ClientID:     clientID,
		InnerIP:      innerIP,
		Crypto:       crypto,
		FECEncoder:   protocol.NewFECEncoder(protocol.DefaultFECEncoderConfig()),
		FECDecoder:   protocol.NewFECDecoder(128),
		dualAckHoles: make(map[uint32]struct{}),
		createdAt:    time.Now(),
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
