package relay

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/tun"
)

// RelaySquadMember represents a player connected to a Squad room on this relay.
type RelaySquadMember struct {
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	ISP      string    `json:"isp"`
	Ping     int       `json:"ping"`
	Game     string    `json:"game"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"lastSeen"`
}

// RelaySquadRoom stores real-time squad roster synced across all teammates.
type RelaySquadRoom struct {
	Code      string                       `json:"code"`
	Host      string                       `json:"host"`
	Game      string                       `json:"game"`
	Members   map[string]*RelaySquadMember `json:"members"`
	CreatedAt time.Time                    `json:"createdAt"`
	mu        sync.RWMutex
}

// Server implements the Lagvex Relay data plane and Squad Signaling Hub.
type Server struct {
	cfg    Config
	pool   *IPPool
	tunDev *tun.Device
	conn   *net.UDPConn

	mu           sync.RWMutex
	sessionsByID map[uint64]*Session
	sessionsByIP map[netip.Addr]*Session

	// dedup holds first-copy-wins state for dual-path connection groups.
	dedup *dualDedupRegistry
	obfs  *protocol.Obfuscator

	// tunAtomic mirrors tunDev so tests and tooling can fetch the device
	// without racing the assignment in Start.
	tunAtomic atomic.Pointer[tun.Device]

	// udpBound is the resolved UDP listen address, set under mu in Start.
	udpBound netip.AddrPort

	// httpBound is the resolved signaling HTTP listen address, set once the
	// listener is up. Keeping it allows callers (tests, tooling) to use a
	// port of 0 and still reach the signaling hub.
	httpBound atomic.Pointer[netip.AddrPort]

	squadRoomsMu sync.RWMutex
	squadRooms   map[string]*RelaySquadRoom

	// squadRateLimit bounds how often one source IP may create or poll
	// squad rooms. The signaling HTTP port is public and CORS-open, so
	// without a limit any web page could spam rooms or enumerate the
	// (short) room codes of a user's relay.
	squadRateMu   sync.Mutex
	squadRate     map[string]*squadRateEntry
	squadRoomsCap int

	// gameServerProbeLimiter bounds how often one source IP may ask the relay
	// to dial a game server. It is separate from squadRate so the two
	// endpoints cannot starve each other. The hourly limiter bounds sustained
	// scanning that the burst window alone would still allow.
	gameServerProbeLimiter  probeRateLimiter
	gameServerHourlyLimiter probeRateLimiter

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
		cfg:           cfg,
		pool:          pool,
		sessionsByID:  make(map[uint64]*Session),
		sessionsByIP:  make(map[netip.Addr]*Session),
		dedup:         newDualDedupRegistry(),
		obfs:          protocol.NewObfuscator(cfg.PSK),
		squadRooms:    make(map[string]*RelaySquadRoom),
		squadRate:     make(map[string]*squadRateEntry),
		squadRoomsCap: defaultSquadRoomsCap,
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
	s.tunAtomic.Store(dev)
	defer func() {
		if closeErr := s.tunDev.Close(); closeErr != nil {
			log.Printf("[Relay] close TUN device: %v", closeErr)
		}
	}()

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
	s.mu.Lock()
	s.udpBound = conn.LocalAddr().(*net.UDPAddr).AddrPort()
	s.mu.Unlock()
	defer func() {
		if closeErr := s.conn.Close(); closeErr != nil {
			log.Printf("[Relay] close UDP socket: %v", closeErr)
		}
	}()

	// Tune UDP socket buffers
	_ = s.conn.SetReadBuffer(8 * 1024 * 1024)
	_ = s.conn.SetWriteBuffer(8 * 1024 * 1024)

	log.Printf("[Relay] Listening for clients on UDP %s", s.conn.LocalAddr())

	var wg sync.WaitGroup
	wg.Add(4)

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

	go s.startSignalingHTTP(ctx, &wg)

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
		handled := false
		if ver == protocol.Version {
			switch mtype {
			case protocol.TypeHandshakeReq:
				handled = s.handleHandshake(buf[:n], remoteEndpoint, false)
			case protocol.TypeData, protocol.TypePing, protocol.TypeDisconnect, protocol.TypeFEC, protocol.TypeAnchorProbe:
				handled = s.handleSessionPacket(buf[:n], remoteEndpoint)
			}
		}

		// Transparent Anti-DPI fallback: if the packet was not valid plaintext,
		// attempt PSK-keyed deobfuscation for obfuscated clients on the same port.
		if !handled && s.obfs != nil {
			if unmaskedHS, ok := s.obfs.DeobfuscateHandshake(buf[:n]); ok {
				uVer, uType := protocol.ParseHeader(unmaskedHS[0])
				if uVer == protocol.Version && uType == protocol.TypeHandshakeReq {
					handled = s.handleHandshake(unmaskedHS, remoteEndpoint, true)
				}
			}
			if !handled && n >= protocol.SecureHeaderLen+protocol.TagLen {
				unmaskPtr := s.bufPool.Get().(*[]byte)
				unmasked := s.obfs.DeobfuscateSession(*unmaskPtr, buf[:n])
				uVer, uType := protocol.ParseHeader(unmasked[0])
				if uVer == protocol.Version {
					switch uType {
					case protocol.TypeData, protocol.TypePing, protocol.TypeDisconnect, protocol.TypeFEC, protocol.TypeAnchorProbe:
						s.handleSessionPacket(unmasked, remoteEndpoint)
					}
				}
				s.bufPool.Put(unmaskPtr)
			}
		}

		s.bufPool.Put(bufPtr)
	}
}

func (s *Server) handleHandshake(raw []byte, remote netip.AddrPort, obfuscated bool) bool {
	req, groupID, err := protocol.DecodeHandshakeRequestExt(s.cfg.PSK, raw)
	if err != nil {
		return false // Silent drop on auth / clock skew failure
	}

	sendResp := func(respBuf []byte) {
		if obfuscated && s.obfs != nil {
			respBuf = s.obfs.ObfuscateHandshake(respBuf)
		}
		s.sendUDP(respBuf, remote)
	}

	// Check max clients
	if s.cfg.MaxClients > 0 && int(s.activeSessions.Load()) >= s.cfg.MaxClients {
		respBuf := protocol.EncodeHandshakeResponse(s.cfg.PSK, protocol.HandshakeResponse{
			Status: protocol.StatusServerBusy,
		})
		sendResp(respBuf)
		return true
	}

	sessionID, err := protocol.RandomUint64()
	if err != nil {
		return true
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
		sendResp(respBuf)
		return true
	}

	// Phase P0.1: Derive symmetric ChaCha20-Poly1305 session keys for relay
	sessCrypto, err := protocol.NewRelayCrypto(s.cfg.PSK, req.Nonce, sessionID, req.ClientID)
	if err != nil {
		log.Printf("[Relay] Failed to init session crypto: %v", err)
		return true
	}

	sess := NewSession(sessionID, req.ClientID, innerIP, remote, sessCrypto)
	sess.Obfuscated = obfuscated

	// Dual-path registration (Redundant-lite Phase 1). The group hash is
	// derived from the PSK, so only authenticated clients can join a group.
	if groupID != nil {
		gh := protocol.GroupHash(s.cfg.PSK, groupID[:])
		sess.GroupHash = gh
		s.dedup.touch(gh)
		log.Printf("[Relay] Dual-path group registered: client=%016x group=%08x", req.ClientID, gh)
	}

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
	if sess.GroupHash != 0 {
		// Advertise dual-path support so a Phase 2 client only duplicates
		// frames when this relay actually deduplicates them.
		respBuf = protocol.EncodeHandshakeResponseDual(s.cfg.PSK, protocol.HandshakeResponse{
			Status:    protocol.StatusOK,
			SessionID: sessionID,
			ClientIP:  innerIP,
			GatewayIP: s.pool.Gateway(),
			MTU:       uint16(s.cfg.MTU),
			NonceEcho: req.Nonce,
		})
	}

	sendResp(respBuf)
	return true
}

func (s *Server) sendSessionUDP(sess *Session, data []byte, to netip.AddrPort) {
	if sess != nil && sess.Obfuscated && s.obfs != nil {
		data = s.obfs.ObfuscateSession(nil, data)
	}
	s.sendUDP(data, to)
}

func (s *Server) handleSessionPacket(raw []byte, remote netip.AddrPort) bool {
	mtype, sessionID, err := protocol.ReadPacketHeader(raw)
	if err != nil {
		return false
	}

	s.mu.RLock()
	sess, exists := s.sessionsByID[sessionID]
	s.mu.RUnlock()

	if !exists || sess.Crypto == nil {
		return false
	}

	bufPtr := s.bufPool.Get().(*[]byte)
	defer s.bufPool.Put(bufPtr)
	plainBuf := *bufPtr
	decType, payload, err := sess.Crypto.OpenPacket(plainBuf, raw)
	if err != nil || decType != mtype {
		return false // Dropped: AEAD auth failure or anti-replay violation
	}

	// Phase P0.3: Authenticated Roaming — update endpoint strictly AFTER Poly1305 authentication succeeds
	sess.Touch()
	sess.UpdateRemote(remote)

	switch decType {
	case protocol.TypeData:
		if len(payload) < 20 || (payload[0]>>4) != 4 {
			return true
		}
		// Anti-spoofing: Verify inner source IPv4 matches assigned session IP
		srcIP := netip.AddrFrom4([4]byte{payload[12], payload[13], payload[14], payload[15]})
		if srcIP != sess.InnerIP {
			return true
		}

		// Security: Prevent accessing private / bogon destinations
		dstIP := netip.AddrFrom4([4]byte{payload[16], payload[17], payload[18], payload[19]})
		if isForbiddenDestination(dstIP) {
			return true
		}

		// Dual-path dedup (Redundant-lite): a dual session's frames carry a
		// sequence trailer; the first copy seen by this relay is forwarded
		// and later copies of the same (group, seq) are dropped.
		inner := payload
		if sess.GroupHash != 0 {
			var seq, frameGroup uint32
			var ok bool
			inner, seq, frameGroup, ok = protocol.SplitDualTrailer(payload)
			if !ok || frameGroup != sess.GroupHash {
				sess.DualBadFrames.Add(1)
				return true
			}
			if s.dedup.observe(sess.GroupHash, seq) {
				sess.DualSuppressed.Add(1)
				// A suppressed duplicate proves this sequence reached the
				// game server earlier; it still counts toward the
				// contiguous ACK watermark (a retransmission can fill a
				// hole that overflowed the tracking set).
				sess.TrackDualForwarded(seq)
				return true
			}
			sess.DualForwarded.Add(1)
			sess.TrackDualForwarded(seq)

			// Dual ACK (Phase 2 groundwork): tell the client the highest
			// sequence forwarded, rate-limited per session, so its
			// retransmit logic only fires for frames this relay may have
			// missed. Single-path sessions never receive ACKs.
			if ackSeq, ok := sess.TakeDualAck(time.Now()); ok {
				ackBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
				ack := sess.Crypto.EncodeDualAck(ackBuf, sessionID, ackSeq)
				s.sendSessionUDP(sess, ack, remote)
			}
		}

		sess.BytesUp.Add(uint64(len(inner)))
		if sess.FECDecoder != nil {
			seq := binary.BigEndian.Uint64(raw[9:17])
			sess.FECDecoder.RecordPacket(seq, inner)
		}

		// Push packet into Linux TUN device
		_, _ = s.tunDev.Write(inner)

	case protocol.TypeFEC:
		if sess.FECDecoder == nil {
			return true
		}
		fecPayload, err := protocol.DecodeFECPayload(payload)
		if err != nil {
			return true
		}
		for _, rec := range sess.FECDecoder.ProcessParityMulti(fecPayload) {
			recovered := rec.Data
			if len(recovered) >= 20 && (recovered[0]>>4) == 4 {
				srcIP := netip.AddrFrom4([4]byte{recovered[12], recovered[13], recovered[14], recovered[15]})
				dstIP := netip.AddrFrom4([4]byte{recovered[16], recovered[17], recovered[18], recovered[19]})
				if srcIP == sess.InnerIP && !isForbiddenDestination(dstIP) {
					sess.BytesUp.Add(uint64(len(recovered)))
					sess.PacketsRecovered.Add(1)
					_, _ = s.tunDev.Write(recovered)
				}
			}
		}

	case protocol.TypePing:
		ts, err := protocol.DecodeControlPayload(payload)
		if err != nil {
			return true
		}
		pongBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
		pong := sess.Crypto.EncodePong(pongBuf, sessionID, ts)
		s.sendSessionUDP(sess, pong, remote)

	case protocol.TypeAnchorProbe:
		packed, err := protocol.DecodeControlPayload(payload)
		if err != nil {
			return true
		}
		seq, anchorID := protocol.DecodeAnchorProbeRequest(packed)
		target, ok := protocol.AnchorProbeTargets[anchorID]
		if !ok {
			// Unknown ids are never probed; the allowlist is the only source of
			// targets so the relay cannot be pointed at arbitrary hosts.
			log.Printf("[Relay] Rejected anchor probe for unknown id %d from session %016x", anchorID, sessionID)
			return true
		}
		if !sess.AllowAnchorProbe(time.Now()) {
			return true // Rate limited; silently drop to avoid amplification.
		}
		rttUs, lossPct := probeAnchorTarget(target)
		replyBuf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
		reply := sess.Crypto.EncodeAnchorProbeReply(replyBuf, sessionID,
			protocol.EncodeAnchorProbeReply(seq, rttUs, lossPct))
		s.sendSessionUDP(sess, reply, remote)

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
	return true
}

// loopTUN reads return packets from the Linux TUN device and sends them to clients.
func (s *Server) loopTUN(ctx context.Context) {
	outBuf := make([]byte, protocol.MaxPacketSize)
	fecBuf := make([]byte, protocol.MaxPacketSize)

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
				s.sendSessionUDP(sess, packet, *remote)

				if sess.FECEncoder != nil {
					seq := binary.BigEndian.Uint64(packet[9:17])
					for _, fecPayloadBytes := range sess.FECEncoder.AddPacketMulti(seq, buf[:n]) {
						fecPacket := sess.Crypto.EncodeFEC(fecBuf, sess.ID, fecPayloadBytes)
						s.sendSessionUDP(sess, fecPacket, *remote)
					}
				}
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

			// Expire dual-path dedup windows and idle groups.
			s.dedup.sweep()

			// Reap stale squad rooms (> 24h old or inactive > 2m)
			s.squadRoomsMu.Lock()
			for code, room := range s.squadRooms {
				room.mu.Lock()
				empty := len(room.Members) == 0
				expired := now.Sub(room.CreatedAt) > 24*time.Hour
				allStale := true
				for _, m := range room.Members {
					if now.Sub(m.LastSeen) < 2*time.Minute {
						allStale = false
						break
					}
				}
				room.mu.Unlock()
				if empty || expired || allStale {
					delete(s.squadRooms, code)
				}
			}
			s.squadRoomsMu.Unlock()
		}
	}
}

// TUN returns the tunnel device once Start has opened it, or nil before that.
func (s *Server) TUN() *tun.Device {
	return s.tunAtomic.Load()
}

// UDPAddr returns the bound UDP listen address once Start has bound it.
func (s *Server) UDPAddr() netip.AddrPort {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.udpBound
}

// HTTPAddr returns the bound squad-signaling HTTP address once Start has
// opened it, or an invalid address before that.
func (s *Server) HTTPAddr() netip.AddrPort {
	if bound := s.httpBound.Load(); bound != nil {
		return *bound
	}
	return netip.AddrPort{}
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

// ── CENTRALIZED SQUAD SIGNALING HUB ──

func (s *Server) startSignalingHTTP(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	httpAddr := s.cfg.HTTPAddr
	if httpAddr == "" {
		host, portStr, err := net.SplitHostPort(s.cfg.ListenAddr)
		if err == nil {
			if port, pErr := strconv.Atoi(portStr); pErr == nil {
				httpAddr = net.JoinHostPort(host, strconv.Itoa(port+1))
			}
		}
		if httpAddr == "" {
			httpAddr = ":51821"
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/squad/create", s.handleSquadCreate)
	mux.HandleFunc("/squad/join", s.handleSquadJoin)
	mux.HandleFunc("/squad/room", s.handleSquadRoom)
	mux.HandleFunc("/squad/heartbeat", s.handleSquadHeartbeat)
	mux.HandleFunc("/squad/leave", s.handleSquadLeave)
	mux.HandleFunc("/probe/gameserver", s.handleProbeGameServer)

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: s.squadRateLimitMiddleware(mux),
	}

	// Listen explicitly so the bound address is known even when the config
	// asked for port 0. Tools and tests rely on HTTPAddr().
	listener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Printf("[Relay] Warning: signaling HTTP listen on %s failed: %v", httpAddr, err)
		return
	}
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		if addr, ok := netip.AddrFromSlice(tcpAddr.IP); ok {
			bound := netip.AddrPortFrom(addr.Unmap(), uint16(tcpAddr.Port))
			s.httpBound.Store(&bound)
		}
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("[Relay] Squad Signaling Hub active on HTTP %s", listener.Addr())
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Printf("[Relay] Warning: signaling HTTP server closed: %v", err)
	}
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	s.mu.RLock()
	active := len(s.sessionsByID)
	dualSessions := 0
	for _, sess := range s.sessionsByID {
		if sess.GroupHash != 0 {
			dualSessions++
		}
	}
	s.mu.RUnlock()

	s.squadRoomsMu.RLock()
	roomsCount := len(s.squadRooms)
	s.squadRoomsMu.RUnlock()

	var dualForwarded, dualSuppressed uint64
	for _, stat := range s.dedup.stats() {
		dualForwarded += stat.Forwarded
		dualSuppressed += stat.Suppressed
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         "online",
		"version":        "1.0.0",
		"activeUsers":    active,
		"squadRooms":     roomsCount,
		"listen":         s.cfg.ListenAddr,
		"dualSessions":   dualSessions,
		"dualGroups":     len(s.dedup.stats()),
		"dualForwarded":  dualForwarded,
		"dualSuppressed": dualSuppressed,
	})
}

func (s *Server) handleSquadCreate(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Host string `json:"host"`
		Game string `json:"game"`
		Ping int    `json:"ping"`
		ISP  string `json:"isp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = req.Host
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if req.Name == "" {
		http.Error(w, `{"error":"host name is required"}`, http.StatusBadRequest)
		return
	}
	if req.Code == "" {
		req.Code = newRoomCode()
	}

	room := &RelaySquadRoom{
		Code:      req.Code,
		Host:      req.Name,
		Game:      req.Game,
		Members:   make(map[string]*RelaySquadMember),
		CreatedAt: time.Now(),
	}
	room.Members[req.Name] = &RelaySquadMember{
		Name:     req.Name,
		Role:     "Host",
		ISP:      req.ISP,
		Ping:     req.Ping,
		Game:     req.Game,
		Status:   "In room",
		LastSeen: time.Now(),
	}

	s.squadRoomsMu.Lock()
	s.insertRoomLocked(room)
	s.squadRoomsMu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"code":    room.Code,
		"host":    room.Host,
		"game":    room.Game,
	})
}

func (s *Server) handleSquadJoin(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Game string `json:"game"`
		Ping int    `json:"ping"`
		ISP  string `json:"isp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" {
		http.Error(w, `{"error":"room code is required"}`, http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, `{"error":"player name is required"}`, http.StatusBadRequest)
		return
	}

	s.squadRoomsMu.Lock()
	room, ok := s.squadRooms[req.Code]
	s.squadRoomsMu.Unlock()
	if !ok {
		http.Error(w, `{"error":"room not found; the host must create it first"}`, http.StatusNotFound)
		return
	}

	room.mu.Lock()
	if len(room.Members) >= 5 && room.Members[req.Name] == nil {
		room.mu.Unlock()
		http.Error(w, `{"error":"squad room is full (5/5 players)"}`, http.StatusConflict)
		return
	}
	role := "Member"
	if req.Name == room.Host {
		role = "Host"
	}
	room.Members[req.Name] = &RelaySquadMember{
		Name:     req.Name,
		Role:     role,
		ISP:      req.ISP,
		Ping:     req.Ping,
		Game:     req.Game,
		Status:   "In room",
		LastSeen: time.Now(),
	}

	memberList := make([]RelaySquadMember, 0, len(room.Members))
	for _, m := range room.Members {
		memberList = append(memberList, *m)
	}
	room.mu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"code":    room.Code,
		"members": memberList,
	})
}

func (s *Server) handleSquadRoom(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	// The room code may arrive as a query parameter (what the dashboard uses)
	// or inside a JSON body (what a script would naturally send).
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("code")))
	if code == "" && r.Method == http.MethodPost {
		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err == nil {
			code = strings.ToUpper(strings.TrimSpace(body.Code))
		}
	}
	if code == "" {
		http.Error(w, `{"error":"room code required"}`, http.StatusBadRequest)
		return
	}

	s.squadRoomsMu.Lock()
	room, ok := s.squadRooms[code]
	s.squadRoomsMu.Unlock()

	if !ok {
		http.Error(w, `{"error":"room not found"}`, http.StatusNotFound)
		return
	}

	room.mu.Lock()
	now := time.Now()
	for name, m := range room.Members {
		if now.Sub(m.LastSeen) > 30*time.Second && m.Role != "Host" {
			delete(room.Members, name)
		}
	}
	memberList := make([]RelaySquadMember, 0, len(room.Members))
	for _, m := range room.Members {
		memberList = append(memberList, *m)
	}
	hostName := room.Host
	gameName := room.Game
	room.mu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"host":    hostName,
		"game":    gameName,
		"members": memberList,
		"count":   len(memberList),
	})
}

func (s *Server) handleSquadHeartbeat(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Ping int    `json:"ping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	s.squadRoomsMu.Lock()
	room, ok := s.squadRooms[req.Code]
	s.squadRoomsMu.Unlock()

	if ok {
		room.mu.Lock()
		if m, exists := room.Members[req.Name]; exists {
			m.LastSeen = time.Now()
			// Zero means "no live RTT sample", so clear stale readings.
			m.Ping = req.Ping
		}
		room.mu.Unlock()
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (s *Server) handleSquadLeave(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	// A body that does not decode must not be reported as a successful
	// leave: nothing changed, and the caller deserves to know why.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	s.squadRoomsMu.Lock()
	room, ok := s.squadRooms[req.Code]
	if ok {
		room.mu.Lock()
		delete(room.Members, req.Name)
		empty := len(room.Members) == 0
		room.mu.Unlock()
		if empty {
			delete(s.squadRooms, req.Code)
		}
	}
	s.squadRoomsMu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}
