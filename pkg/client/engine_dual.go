package client

// Dual-path (Redundant-lite Phase 2) client integration.
//
// Data frames go to the primary relay only; the primary reports progress via
// DualAck, and frames that stay unacknowledged past a measured threshold are
// retransmitted — first to the primary relay (relay dedup makes that
// idempotent), then to the standby relay, which forwards them to the game
// server. The common case therefore adds no duplicate packets at the game
// server; every retransmission is counted and surfaced in /api/status.
//
// Dual mode requires both relays to share one PSK: the dedup scope is
// SHA-256(PSK || group ID), so different PSKs would mean different group
// identities per relay and the single trailer could not satisfy both.

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

const (
	dualKeepaliveInterval = 10 * time.Second
	dualStandbyTimeout    = 30 * time.Second // no standby pong → teardown
	dualRetransmitTick    = 10 * time.Millisecond
	dualRetryInterval     = 5 * time.Second
)

// DualPathStatus is the honest dual-path state for /api/status. Every value
// is measured; when the mode is inactive, Reason says exactly why.
type DualPathStatus struct {
	Enabled bool `json:"enabled"` // user intent
	Active  bool `json:"active"`  // standby session live

	// Reason explains an inactive state: "not connected", "no fresh standby
	// measurement", "relay does not support dual path", and so on.
	Reason string `json:"reason,omitempty"`

	StandbyRelayID   string `json:"standbyRelayId,omitempty"`
	StandbyRelayName string `json:"standbyRelayName,omitempty"`

	GroupHash    uint32 `json:"groupHash,omitempty"`
	AckedSeq     uint32 `json:"ackedSeq"`
	InFlight     int    `json:"inFlight"`
	Framed       uint64 `json:"framed"`
	Retransmits  uint64 `json:"retransmits"`
	AckPurged    uint64 `json:"ackPurged"`
	Expired      uint64 `json:"expired"`
	RingOverflow uint64 `json:"ringOverflow"`
	ThresholdMs  int    `json:"thresholdMs,omitempty"`
}

// SetDualPath enables or disables dual-path redundancy. When enabling while
// connected, activation runs synchronously (one dual handshake, ≤3 s) so the
// API response reflects the real state.
func (e *Engine) SetDualPath(enabled bool) {
	if !enabled {
		e.dualEnabled.Store(false)
		e.deactivateDual("turned off")
		return
	}

	e.dualEnabled.Store(true)
	e.mu.Lock()
	connected := e.state == StateConnected
	primaryPSK := append([]byte(nil), e.psk...)
	primaryEndpoint := e.relayAddr.String()
	pingMs := e.pingMs.Load()
	e.mu.Unlock()

	if !connected {
		e.setDualReason("not connected; activates when the tunnel connects")
		return
	}

	standby, fresh := e.freshStandby()
	if !fresh {
		e.setDualReason("no fresh standby measurement yet; retrying when the standby refreshes")
		return
	}
	if standby.Endpoint == primaryEndpoint || standby.Endpoint == "" {
		e.setDualReason("standby candidate is the active relay")
		return
	}

	psk, ok := e.relayPSKForEndpoint(standby.Endpoint)
	if !ok {
		e.setDualReason("standby relay has no PSK in the catalog")
		return
	}
	if primaryPSK == nil || string(psk) != string(primaryPSK) {
		e.setDualReason("dual path requires both relays to use the same PSK")
		return
	}

	e.activateDual(standby.Endpoint, psk, dualThresholdFor(pingMs), standby.RelayID, standby.Name)
}

// IsDualPathEnabled reports the user's dual-path intent.
func (e *Engine) IsDualPathEnabled() bool {
	return e.dualEnabled.Load()
}

// dualThresholdFor derives the retransmission threshold from the measured
// tunnel RTT: max(40 ms, 1.5× RTT), clamped by the sender's own bounds.
func dualThresholdFor(pingMs int64) time.Duration {
	threshold := 40 * time.Millisecond
	if pingMs > 0 {
		threshold = time.Duration(1.5 * float64(pingMs) * float64(time.Millisecond))
		if threshold < 40*time.Millisecond {
			threshold = 40 * time.Millisecond
		}
	}
	if threshold > DualMaxRetransmitThreshold {
		threshold = DualMaxRetransmitThreshold
	}
	if threshold < DualMinRetransmitThreshold {
		threshold = DualMinRetransmitThreshold
	}
	return threshold
}

// relayPSKForEndpoint looks up the configured PSK for a relay endpoint.
func (e *Engine) relayPSKForEndpoint(endpoint string) ([]byte, bool) {
	if e.profileMgr == nil {
		return nil, false
	}
	for _, relay := range e.profileMgr.Catalog().Relays {
		if relay.Endpoint == endpoint {
			if relay.PSK == "" {
				return nil, false
			}
			return []byte(relay.PSK), true
		}
	}
	return nil, false
}

// ensureGroupID returns the process-lifetime connection group ID, generating
// it on first use.
func (e *Engine) ensureGroupID() ([16]byte, error) {
	e.dualMu.Lock()
	defer e.dualMu.Unlock()
	if e.groupID == ([16]byte{}) {
		if _, err := rand.Read(e.groupID[:]); err != nil {
			return [16]byte{}, fmt.Errorf("generate group id: %w", err)
		}
	}
	return e.groupID, nil
}

// activateDual opens the standby dual session and starts its loops. It is
// synchronous: the caller learns the real result from dualActive/dualReason.
func (e *Engine) activateDual(endpoint string, psk []byte, threshold time.Duration, relayID, relayName string) {
	groupID, err := e.ensureGroupID()
	if err != nil {
		e.setDualReason(err.Error())
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", endpoint)
	if err != nil {
		e.setDualReason(fmt.Sprintf("resolve standby endpoint: %v", err))
		return
	}
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		e.setDualReason(fmt.Sprintf("standby socket: %v", err))
		return
	}

	nonce, err := protocol.RandomUint64()
	if err != nil {
		_ = conn.Close()
		e.setDualReason("entropy failure during standby handshake")
		return
	}
	req := protocol.HandshakeRequest{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		ClientID:  e.clientID,
	}
	reqBuf := protocol.EncodeHandshakeRequestDual(psk, req, groupID)

	var resp *protocol.HandshakeResponse
	var dualOK bool
	recvBuf := make([]byte, 256)
	for attempt := 1; attempt <= 3; attempt++ {
		if _, err := conn.WriteToUDP(reqBuf, udpAddr); err != nil {
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		n, _, readErr := conn.ReadFromUDP(recvBuf)
		if readErr != nil {
			continue
		}
		r, dual, decErr := protocol.DecodeHandshakeResponseExt(psk, recvBuf[:n], nonce)
		if decErr == nil && r.Status == protocol.StatusOK {
			resp, dualOK = r, dual
			break
		}
	}
	_ = conn.SetReadDeadline(time.Time{})

	if resp == nil || resp.Status != protocol.StatusOK {
		_ = conn.Close()
		e.setDualReason("standby relay handshake failed (unreachable or PSK mismatch)")
		return
	}
	if !dualOK {
		_ = conn.Close()
		e.setDualReason("standby relay does not support dual path (relay build predates Redundant-lite)")
		return
	}

	crypto, err := protocol.NewClientCrypto(psk, nonce, resp.SessionID, e.clientID)
	if err != nil {
		_ = conn.Close()
		e.setDualReason("standby session crypto failed")
		return
	}

	tunnelCtx := e.cancelCtx
	if tunnelCtx == nil {
		tunnelCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(tunnelCtx)

	groupHash := protocol.GroupHash(psk, groupID[:])
	sender := newDualSender(groupHash, nil, threshold)

	// Replace any previous standby session.
	e.dualMu.Lock()
	oldConn := e.dualStandbyConn
	oldCancel := e.dualCancel
	e.dualStandbyAddr = udpAddr.AddrPort()
	e.dualStandbyConn = conn
	e.dualStandbyCrypto = crypto
	e.dualStandbySession = resp.SessionID
	e.dualStandbyID = relayID
	e.dualStandbyName = relayName
	e.dualSender = sender
	e.dualCancel = cancel
	e.dualMu.Unlock()

	if oldConn != nil {
		_ = oldConn.Close()
	}
	if oldCancel != nil {
		oldCancel()
	}

	e.dualLastPong.Store(time.Now().UnixNano())
	e.dualActive.Store(true)
	e.dualReason.Store(nil)
	e.dualRetryAt.Store(0)

	e.wg.Add(2)
	go e.pumpDualStandby(ctx, conn)
	go e.loopDualMaintain(ctx)

	log.Printf("[Engine] Dual path active: standby=%s (%s) session=%016x group=%08x threshold=%s",
		relayName, endpoint, resp.SessionID, groupHash, threshold)
}

// pumpDualStandby reads the standby relay socket: pongs refresh liveness, a
// server disconnect tears the standby down. Downstream game traffic stays on
// the primary relay.
func (e *Engine) pumpDualStandby(ctx context.Context, conn *net.UDPConn) {
	defer e.wg.Done()
	buf := make([]byte, protocol.MaxPacketSize)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			// Socket closed by teardown or an unrecoverable error.
			if e.dualActive.Load() {
				e.deactivateDual("standby socket error")
			}
			return
		}
		if n < protocol.SecureHeaderLen+protocol.TagLen {
			continue
		}

		e.dualMu.Lock()
		crypto := e.dualStandbyCrypto
		e.dualMu.Unlock()
		if crypto == nil {
			return
		}
		mtype, _, err := crypto.OpenPacket(nil, buf[:n])
		if err != nil {
			continue
		}
		switch mtype {
		case protocol.TypePong:
			e.dualLastPong.Store(time.Now().UnixNano())
		case protocol.TypeDisconnect:
			if e.dualActive.Load() {
				e.deactivateDual("standby relay disconnected the session")
			}
			return
		}
	}
}

// loopDualMaintain sends standby keepalives and drives retransmissions.
func (e *Engine) loopDualMaintain(ctx context.Context) {
	defer e.wg.Done()

	retransmit := time.NewTicker(dualRetransmitTick)
	defer retransmit.Stop()
	keepalive := time.NewTicker(dualKeepaliveInterval)
	defer keepalive.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-retransmit.C:
			e.pumpDualRetransmits()
		case <-keepalive.C:
			e.pingDualStandby()
			if last := e.dualLastPong.Load(); last > 0 && time.Since(time.Unix(0, last)) > dualStandbyTimeout {
				if e.dualActive.Load() {
					e.deactivateDual("standby relay stopped responding")
				}
				return
			}
		}
	}
}

// pumpDualRetransmits sends every due retransmission. Attempt 1 goes to the
// primary relay (its dedup window absorbs it if the original arrived);
// attempt 2 goes to the standby relay, which then forwards to the game
// server.
func (e *Engine) pumpDualRetransmits() {
	e.dualMu.Lock()
	sender := e.dualSender
	standbyCrypto := e.dualStandbyCrypto
	standbySession := e.dualStandbySession
	standbyConn := e.dualStandbyConn
	standbyAddr := e.dualStandbyAddr
	e.dualMu.Unlock()
	if sender == nil {
		return
	}

	due := sender.pendingRetransmits()
	if len(due) == 0 {
		return
	}

	e.mu.Lock()
	primaryCrypto := e.crypto
	primarySession := e.sessionID
	primaryConn := e.udpConn
	primaryAddr := e.relayAddr
	e.mu.Unlock()

	primaryBuf := make([]byte, protocol.MaxPacketSize)
	standbyBuf := make([]byte, protocol.MaxPacketSize)
	for _, item := range due {
		switch item.Attempt {
		case 1:
			if primaryCrypto == nil || primaryConn == nil {
				continue
			}
			pkt := primaryCrypto.EncodeData(primaryBuf, primarySession, item.Frame)
			_, _ = primaryConn.WriteToUDP(pkt, net.UDPAddrFromAddrPort(primaryAddr))
		case 2:
			if standbyCrypto == nil || standbyConn == nil {
				continue
			}
			pkt := standbyCrypto.EncodeData(standbyBuf, standbySession, item.Frame)
			_, _ = standbyConn.WriteToUDP(pkt, net.UDPAddrFromAddrPort(standbyAddr))
		}
	}
}

// pingDualStandby keeps the standby session alive so retransmissions never
// pay a handshake cost.
func (e *Engine) pingDualStandby() {
	e.dualMu.Lock()
	crypto := e.dualStandbyCrypto
	session := e.dualStandbySession
	conn := e.dualStandbyConn
	addr := e.dualStandbyAddr
	e.dualMu.Unlock()

	if crypto == nil || conn == nil {
		return
	}
	buf := make([]byte, protocol.SecureHeaderLen+8+protocol.TagLen)
	ping := crypto.EncodePing(buf, session, uint64(time.Now().UnixNano()))
	_, _ = conn.WriteToUDP(ping, net.UDPAddrFromAddrPort(addr))
}

// deactivateDual tears the standby session down and records an honest reason.
func (e *Engine) deactivateDual(reason string) {
	e.dualMu.Lock()
	cancel := e.dualCancel
	conn := e.dualStandbyConn
	crypto := e.dualStandbyCrypto
	session := e.dualStandbySession
	addr := e.dualStandbyAddr
	e.dualStandbyConn = nil
	e.dualStandbyCrypto = nil
	e.dualStandbySession = 0
	e.dualStandbyID = ""
	e.dualStandbyName = ""
	e.dualSender = nil
	e.dualCancel = nil
	e.dualMu.Unlock()

	wasActive := e.dualActive.Swap(false)
	e.setDualReason(reason)

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		if crypto != nil && session != 0 {
			buf := make([]byte, protocol.SecureHeaderLen+protocol.TagLen)
			disc := crypto.EncodeDisconnect(buf, session)
			_, _ = conn.WriteToUDP(disc, net.UDPAddrFromAddrPort(addr))
		}
		_ = conn.Close()
	}
	if wasActive {
		log.Printf("[Engine] Dual path deactivated: %s", reason)
	}
}

// maybeRetryDual re-attempts activation when the user asked for dual mode but
// the last attempt could not (no fresh standby, transient handshake failure).
func (e *Engine) maybeRetryDual() {
	if !e.dualEnabled.Load() || e.dualActive.Load() {
		return
	}
	e.mu.Lock()
	connected := e.state == StateConnected
	e.mu.Unlock()
	if !connected {
		return
	}
	now := time.Now().UnixNano()
	last := e.dualRetryAt.Load()
	if last != 0 && now-last < int64(dualRetryInterval) {
		return
	}
	if !e.dualRetryAt.CompareAndSwap(last, now) {
		return
	}
	go e.SetDualPath(true)
}

func (e *Engine) setDualReason(reason string) {
	e.dualReason.Store(&reason)
}

// dualStatusSnapshot builds the honest status object for /api/status.
func (e *Engine) dualStatusSnapshot() *DualPathStatus {
	enabled := e.dualEnabled.Load()
	active := e.dualActive.Load()
	if !enabled && !active {
		return nil
	}

	status := &DualPathStatus{
		Enabled: enabled,
		Active:  active,
	}
	if reason := e.dualReason.Load(); reason != nil {
		status.Reason = *reason
	}

	e.dualMu.Lock()
	status.StandbyRelayID = e.dualStandbyID
	status.StandbyRelayName = e.dualStandbyName
	sender := e.dualSender
	e.dualMu.Unlock()

	if sender != nil {
		snap := sender.stats()
		status.GroupHash = snap.GroupHash
		status.AckedSeq = snap.AckedSeq
		status.InFlight = snap.InFlight
		status.Framed = snap.Framed
		status.Retransmits = snap.Retransmits
		status.AckPurged = snap.AckPurged
		status.Expired = snap.Expired
		status.RingOverflow = snap.RingOverflow
		status.ThresholdMs = snap.ThresholdMs
	}
	return status
}

// dualSenderSnapshot returns the live sender, or nil when dual is inactive.
func (e *Engine) dualSenderSnapshot() *dualSender {
	e.dualMu.Lock()
	defer e.dualMu.Unlock()
	return e.dualSender
}

// applyDualAck feeds a primary-relay DualAck into the sender ring.
func (e *Engine) applyDualAck(contiguous uint32) {
	if sender := e.dualSenderSnapshot(); sender != nil {
		sender.onAck(contiguous)
	}
}
