package relay

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/tun"
)

// Server implements the Lagvex Relay data plane on Linux.
type Server struct {
	cfg    Config
	pool   *IPPool
	tunDev *tun.Device
	conn   *net.UDPConn

	mu           sync.RWMutex
	sessionsByID map[uint64]*Session
	sessionsByIP map[netip.Addr]*Session

	bufPool sync.Pool

	activeSessions atomic.Int64
}

// NewServer creates a new Relay server instance.
func NewServer(cfg Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	pool, err := NewIPPool(cfg.Subnet)
	if err != nil {
		return nil, fmt.Errorf("init ip pool: %w", err)
	}

	return &Server{
		cfg:          cfg,
		pool:         pool,
		sessionsByID: make(map[uint64]*Session),
		sessionsByIP: make(map[netip.Addr]*Session),
		bufPool: sync.Pool{
			New: func() any {
				b := make([]byte, protocol.MaxPacketSize)
				return &b
			},
		},
	}, nil
}

// Start opens network interfaces and starts packet handling loops.
func (s *Server) Start(ctx context.Context) error {
	// 1. Open TUN device
	dev, err := tun.Open(s.cfg.TunName)
	if err != nil {
		return fmt.Errorf("open tun %s: %w", s.cfg.TunName, err)
	}
	s.tunDev = dev
	defer s.tunDev.Close()

	// 2. Configure TUN interface with gateway IP
	gwCidr := fmt.Sprintf("%s/%d", s.pool.Gateway(), s.cfg.Subnet.Bits())
	if err := s.tunDev.Configure(gwCidr, s.cfg.MTU); err != nil {
		return fmt.Errorf("configure tun: %w", err)
	}
	log.Printf("[Relay] TUN device %s up at %s (MTU %d)", s.tunDev.Name(), gwCidr, s.cfg.MTU)

	// 3. Listen on UDP port
	laddr, err := net.ResolveUDPAddr("udp4", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("resolve udp addr: %w", err)
	}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	s.conn = conn
	defer s.conn.Close()

	// Tune UDP socket buffers
	_ = s.conn.SetReadBuffer(8 * 1024 * 1024)
	_ = s.conn.SetWriteBuffer(8 * 1024 * 1024)

	log.Printf("[Relay] Listening for clients on UDP %s", s.conn.LocalAddr())

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		s.loopUDP(ctx)
	}()

	go func() {
		defer wg.Done()
		s.loopTUN(ctx)
	}()

	go func() {
		defer wg.Done()
		s.loopReaper(ctx)
	}()

	<-ctx.Done()
	log.Printf("[Relay] Shutting down...")
	_ = s.conn.Close()
	_ = s.tunDev.Close()
	wg.Wait()
	log.Printf("[Relay] Stopped cleanly.")
	return nil
}

// loopUDP handles incoming UDP packets from game booster clients.
func (s *Server) loopUDP(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		bufPtr := s.bufPool.Get().(*[]byte)
		buf := *bufPtr

		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			s.bufPool.Put(bufPtr)
			if ctx.Err() != nil {
				return
			}
			continue
		}

		if n < 1 {
			s.bufPool.Put(bufPtr)
			continue
		}

		remoteEndpoint := remoteAddr.AddrPort()
		ver, mtype := protocol.ParseHeader(buf[0])
		if ver != protocol.Version {
			s.bufPool.Put(bufPtr)
			continue
		}

		switch mtype {
		case protocol.TypeHandshakeReq:
			s.handleHandshake(buf[:n], remoteEndpoint)
		case protocol.TypeData, protocol.TypePing, protocol.TypeDisconnect:
			s.handleSessionPacket(buf[:n], remoteEndpoint)
		}

		s.bufPool.Put(bufPtr)
	}
}

func (s *Server) handleHandshake(raw []byte, remote netip.AddrPort) {
	req, err := protocol.DecodeHandshakeRequest(s.cfg.PSK, raw)
	if err != nil {
		return // Silent drop on auth / clock skew failure
	}

	// Check max clients
	if s.cfg.MaxClients > 0 && int(s.activeSessions.Load()) >= s.cfg.MaxClients {
		respBuf := protocol.EncodeHandshakeResponse(s.cfg.PSK, protocol.HandshakeResponse{
			Status: protocol.StatusServerBusy,
		})
		s.sendUDP(respBuf, remote)
		return
	}

	sessionID, err := protocol.RandomUint64()
	if err != nil {
		return
	}

	// Phase 0.1: Clean previous session if same client ID reconnected BEFORE allocating IP
	s.mu.Lock()
	for _, old := range s.sessionsByID {
		if old.ClientID == req.ClientID {
			delete(s.sessionsByID, old.ID)
			delete(s.sessionsByIP, old.InnerIP)
			s.pool.Release(old.InnerIP, old.ClientID, true) // Keep reservation for this client!
			s.activeSessions.Add(-1)
			log.Printf("[Relay] Reconnect: purged stale session=%016x for client=%016x", old.ID, req.ClientID)
			break
		}
	}
	s.mu.Unlock()

	innerIP, _, err := s.pool.Allocate(req.ClientID, sessionID)
	if err != nil {
		respBuf := protocol.EncodeHandshakeResponse(s.cfg.PSK, protocol.HandshakeResponse{
			Status: protocol.StatusPoolFull,
		})
		s.sendUDP(respBuf, remote)
		return
	}

	// Phase P0.1: Derive symmetric ChaCha20-Poly1305 session keys for relay
	sessCrypto, err := protocol.NewRelayCrypto(s.cfg.PSK, req.Nonce, sessionID, req.ClientID)
	if err != nil {
		log.Printf("[Relay] Failed to init session crypto: %v", err)
		return
	}

	sess := NewSession(sessionID, req.ClientID, innerIP, remote, sessCrypto)

	s.mu.Lock()
	s.sessionsByID[sessionID] = sess
	s.sessionsByIP[innerIP] = sess
	s.activeSessions.Add(1)
	s.mu.Unlock()

	log.Printf("[Relay] Handshake OK: client=%016x innerIP=%s remote=%s session=%016x (active=%d)",
		req.ClientID, innerIP, remote, sessionID, s.activeSessions.Load())

	respBuf := protocol.EncodeHandshakeResponse(s.cfg.PSK, protocol.HandshakeResponse{
		Status:    protocol.StatusOK,
		SessionID: sessionID,
		ClientIP:  innerIP,
		GatewayIP: s.pool.Gateway(),
		MTU:       uint16(s.cfg.MTU),
		NonceEcho: req.Nonce,
	})

	s.sendUDP(respBuf, remote)
}

func (s *Server) handleSessionPacket(raw []byte, remote netip.AddrPort) {
	mtype, sessionID, err := protocol.ReadPacketHeader(raw)
	if err != nil {
		return
	}

	s.mu.RLock()
	sess, exists := s.sessionsByID[sessionID]
	s.mu.RUnlock()

	if !exists || sess.Crypto == nil {
		return
	}

	plainBuf := make([]byte, protocol.MaxPacketSize)
	decType, payload, err := sess.Crypto.OpenPacket(plainBuf, raw)
	if err != nil || decType != mtype {
		return // Dropped: AEAD auth failure or anti-replay violation
	}

	// Phase P0.3: Authenticated Roaming — update endpoint strictly AFTER Poly1305 authentication succeeds
	sess.Touch()
	sess.UpdateRemote(remote)

	switch decType {
	case protocol.TypeData:
		if len(payload) < 20 || (payload[0]>>4) != 4 {
			return
		}
		// Anti-spoofing: Verify inner source IPv4 matches assigned session IP
		srcIP := netip.AddrFrom4([4]byte{payload[12], payload[13], payload[14], payload[15]})
		if srcIP != sess.InnerIP {
			return
		}

		// Security: Prevent accessing private / bogon destinations
		dstIP := netip.AddrFrom4([4]byte{payload[16], payload[17], payload[18], payload[19]})
		if isForbiddenDestination(dstIP) {
			return
		}

		sess.BytesUp.Add(uint64(len(payload)))

		// Push packet into Linux TUN device
		_, _ = s.tunDev.Write(payload)

	case protocol.TypePing:
		ts, err := protocol.DecodeControlPayload(payload)
		if err != nil {
			return
		}
		pongBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
		pong := sess.Crypto.EncodePong(pongBuf, sessionID, ts)
		s.sendUDP(pong, remote)

	case protocol.TypeDisconnect:
		s.mu.Lock()
		if curSess, ok := s.sessionsByID[sessionID]; ok {
			delete(s.sessionsByID, sessionID)
			delete(s.sessionsByIP, curSess.InnerIP)
			s.pool.Release(curSess.InnerIP, curSess.ClientID, false)
			s.activeSessions.Add(-1)
			log.Printf("[Relay] Authenticated Disconnect: session=%016x innerIP=%s", sessionID, curSess.InnerIP)
		}
		s.mu.Unlock()
	}
}

// loopTUN reads return packets from the Linux TUN device and sends them to clients.
func (s *Server) loopTUN(ctx context.Context) {
	outBuf := make([]byte, protocol.MaxPacketSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		bufPtr := s.bufPool.Get().(*[]byte)
		buf := *bufPtr

		n, err := s.tunDev.Read(buf)
		if err != nil {
			s.bufPool.Put(bufPtr)
			if ctx.Err() != nil {
				return
			}
			continue
		}

		if n < 20 || (buf[0]>>4) != 4 { // must be valid IPv4
			s.bufPool.Put(bufPtr)
			continue
		}

		// Inner destination IPv4
		dstIP := netip.AddrFrom4([4]byte{buf[16], buf[17], buf[18], buf[19]})

		s.mu.RLock()
		sess, ok := s.sessionsByIP[dstIP]
		s.mu.RUnlock()

		if ok {
			remote := sess.RemoteUDP.Load()
			if remote != nil && sess.Crypto != nil {
				sess.Touch()
				sess.BytesDown.Add(uint64(n))

				packet := sess.Crypto.EncodeData(outBuf, sess.ID, buf[:n])
				s.sendUDP(packet, *remote)
			}
		}

		s.bufPool.Put(bufPtr)
	}
}

// loopReaper cleans up expired idle sessions.
func (s *Server) loopReaper(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for id, sess := range s.sessionsByID {
				if now.Sub(sess.LastSeenTime()) > s.cfg.IdleTimeout {
					delete(s.sessionsByID, id)
					delete(s.sessionsByIP, sess.InnerIP)
					s.pool.Release(sess.InnerIP, sess.ClientID, true) // keep reservation
					s.activeSessions.Add(-1)
					log.Printf("[Relay] Session timed out: session=%016x innerIP=%s", id, sess.InnerIP)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Server) sendUDP(data []byte, to netip.AddrPort) {
	udpAddr := net.UDPAddrFromAddrPort(to)
	_, _ = s.conn.WriteToUDP(data, udpAddr)
}

// isForbiddenDestination returns true if destination IP is private, bogon, or loopback.
func isForbiddenDestination(ip netip.Addr) bool {
	if !ip.Is4() {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	// Specifically guard AWS/Cloud metadata service 169.254.169.254
	b := ip.As4()
	if b[0] == 169 && b[1] == 254 {
		return true
	}
	return false
}
