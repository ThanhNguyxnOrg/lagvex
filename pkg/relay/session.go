package relay

import (
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"net/netip"
	"sync/atomic"
	"time"
)

// Session represents an authenticated connected client.
type Session struct {
	ID        uint64
	ClientID  uint64
	InnerIP   netip.Addr
	RemoteUDP atomic.Pointer[netip.AddrPort]
	Crypto    *protocol.SessionCrypto

	lastSeen  atomic.Int64 // Unix nanoseconds
	createdAt time.Time

	BytesUp   atomic.Uint64
	BytesDown atomic.Uint64
}

// NewSession initializes a new client session.
func NewSession(id, clientID uint64, innerIP netip.Addr, remote netip.AddrPort, crypto *protocol.SessionCrypto) *Session {
	s := &Session{
		ID:        id,
		ClientID:  clientID,
		InnerIP:   innerIP,
		Crypto:    crypto,
		createdAt: time.Now(),
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
