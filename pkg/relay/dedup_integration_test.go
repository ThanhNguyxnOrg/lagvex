package relay

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"runtime"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

// testClient is one authenticated client session against the loopback relay.
type testClient struct {
	conn      *net.UDPConn
	relayAddr netip.AddrPort
	crypto    *protocol.SessionCrypto
	sessionID uint64
	innerIP   netip.Addr
	dualOK    bool
	groupHash uint32
}

// startTestRelay boots a relay on loopback UDP with a userspace TUN sink and
// returns the server, its UDP address and the forwarded-packet channel.
func startTestRelay(t *testing.T, psk []byte) (*Server, netip.AddrPort, chan []byte) {
	t.Helper()

	if runtime.GOOS != "linux" {
		t.Skip("relay integration test requires the Linux userspace TUN fallback")
	}

	cfg := Config{
		ListenAddr:  "127.0.0.1:0",
		HTTPAddr:    "127.0.0.1:0",
		TunName:     "lagvexddt",
		Subnet:      netip.MustParsePrefix("10.77.0.0/24"),
		PSK:         psk,
		MTU:         1400,
		IdleTimeout: 30 * time.Second,
		MaxClients:  8,
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	forwarded := make(chan []byte, 64)
	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if dev := srv.TUN(); dev != nil {
				dev.SetWriteSink(func(p []byte) {
					select {
					case forwarded <- p:
					default:
					}
				})
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	startErr := make(chan error, 1)
	go func() {
		startErr <- srv.Start(ctx)
	}()
	t.Cleanup(cancel)

	var udpAddr netip.AddrPort
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if addr := srv.UDPAddr(); addr.IsValid() {
			udpAddr = addr
			break
		}
		select {
		case err := <-startErr:
			t.Fatalf("relay Start failed: %v", err)
		default:
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !udpAddr.IsValid() {
		t.Fatal("relay did not bind UDP in time")
	}
	return srv, udpAddr, forwarded
}

func dialTestClient(t *testing.T, udpAddr netip.AddrPort, psk []byte, clientID uint64, groupID *[protocol.DualGroupIDLen]byte) *testClient {
	t.Helper()

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		t.Fatalf("client socket: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	nonce, err := protocol.RandomUint64()
	if err != nil {
		t.Fatalf("nonce: %v", err)
	}
	req := protocol.HandshakeRequest{Nonce: nonce, Timestamp: time.Now().Unix(), ClientID: clientID}

	var reqBuf []byte
	if groupID != nil {
		reqBuf = protocol.EncodeHandshakeRequestDual(psk, req, *groupID)
	} else {
		reqBuf = protocol.EncodeHandshakeRequest(psk, req)
	}

	if _, err := conn.WriteToUDP(reqBuf, net.UDPAddrFromAddrPort(udpAddr)); err != nil {
		t.Fatalf("send handshake: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 256)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read handshake response: %v", err)
	}

	resp, dualOK, err := protocol.DecodeHandshakeResponseExt(psk, buf[:n], nonce)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != protocol.StatusOK {
		t.Fatalf("handshake status = %d", resp.Status)
	}

	crypto, err := protocol.NewClientCrypto(psk, nonce, resp.SessionID, clientID)
	if err != nil {
		t.Fatalf("client crypto: %v", err)
	}

	c := &testClient{
		conn:      conn,
		relayAddr: udpAddr,
		crypto:    crypto,
		sessionID: resp.SessionID,
		innerIP:   resp.ClientIP,
		dualOK:    dualOK,
	}
	if groupID != nil {
		c.groupHash = protocol.GroupHash(psk, groupID[:])
	}
	return c
}

// send writes one authenticated data frame carrying the given payload.
func (c *testClient) send(t *testing.T, payload []byte) {
	t.Helper()
	buf := make([]byte, protocol.MaxPacketSize)
	pkt := c.crypto.EncodeData(buf, c.sessionID, payload)
	if _, err := c.conn.WriteToUDP(pkt, net.UDPAddrFromAddrPort(c.relayAddr)); err != nil {
		t.Fatalf("send data: %v", err)
	}
}

// expectAck reads one DualAck control message and returns its sequence.
func (c *testClient) expectAck(t *testing.T, wantMin uint32) uint32 {
	t.Helper()
	buf := make([]byte, 256)
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Second))
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read dual ack: %v", err)
	}
	mtype, payload, err := c.crypto.OpenPacket(make([]byte, 256), buf[:n])
	if err != nil {
		t.Fatalf("open dual ack: %v", err)
	}
	if mtype != protocol.TypeDualAck {
		t.Fatalf("expected DualAck, got %s", protocol.TypeString(mtype))
	}
	seq, err := protocol.DecodeDualAckPayload(payload)
	if err != nil {
		t.Fatalf("decode dual ack: %v", err)
	}
	if seq < wantMin {
		t.Fatalf("ack seq = %d, want >= %d", seq, wantMin)
	}
	return seq
}

// buildInnerPacket builds a minimal valid IPv4/UDP packet the relay accepts:
// correct source (session IP) and a public destination.
func buildInnerPacket(t *testing.T, src, dst netip.Addr, marker byte) []byte {
	t.Helper()

	pkt := make([]byte, 28) // IPv4 header (20) + UDP header (8)
	pkt[0] = 0x45
	binary.BigEndian.PutUint16(pkt[2:4], uint16(len(pkt)))
	pkt[8] = 64
	pkt[9] = 17 // UDP
	src4 := src.As4()
	dst4 := dst.As4()
	copy(pkt[12:16], src4[:])
	copy(pkt[16:20], dst4[:])
	binary.BigEndian.PutUint16(pkt[20:22], 40000)
	binary.BigEndian.PutUint16(pkt[22:24], 7777)
	binary.BigEndian.PutUint16(pkt[24:26], 8)
	pkt[27] = marker
	return pkt
}

// expectForwarded reads one forwarded inner packet, failing on timeout.
func expectForwarded(t *testing.T, forwarded chan []byte, want []byte) {
	t.Helper()
	select {
	case got := <-forwarded:
		if string(got) != string(want) {
			t.Fatalf("forwarded packet differs from the inner payload (len %d vs %d)", len(got), len(want))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected packet was not forwarded to the TUN sink")
	}
}

// expectNoForwarded asserts nothing else arrives shortly after.
func expectNoForwarded(t *testing.T, forwarded chan []byte, why string) {
	t.Helper()
	select {
	case extra := <-forwarded:
		t.Fatalf("%s: unexpected forwarded packet (%d bytes)", why, len(extra))
	case <-time.After(300 * time.Millisecond):
	}
}

func activeSession(t *testing.T, srv *Server) *Session {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.RLock()
		var found *Session
		for _, sess := range srv.sessionsByID {
			found = sess
			break
		}
		srv.mu.RUnlock()
		if found != nil {
			return found
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no active session found")
	return nil
}

func TestDualDedupEndToEnd(t *testing.T) {
	psk := []byte("integration-psk-0123456789ab")
	srv, udpAddr, forwarded := startTestRelay(t, psk)

	var groupID [protocol.DualGroupIDLen]byte
	for i := range groupID {
		groupID[i] = byte(0x40 + i)
	}
	client := dialTestClient(t, udpAddr, psk, 0xD011, &groupID)
	if !client.dualOK {
		t.Fatal("relay did not advertise dual-path support for a dual handshake")
	}
	if client.groupHash == 0 {
		t.Fatal("group hash must not be zero")
	}

	// 1. The same dual frame twice: first copy forwarded, duplicate suppressed.
	inner := buildInnerPacket(t, client.innerIP, netip.MustParseAddr("8.8.8.8"), 1)
	framed := protocol.AppendDualTrailer(inner, 1, client.groupHash)
	client.send(t, framed)
	expectForwarded(t, forwarded, inner)

	// The relay reports progress so the client knows seq 1 needs no
	// retransmission; this is the Phase 2 ARQ groundwork.
	client.expectAck(t, 1)

	client.send(t, framed)
	expectNoForwarded(t, forwarded, "duplicate copy")

	sess := activeSession(t, srv)
	if sess.DualForwarded.Load() != 1 || sess.DualSuppressed.Load() != 1 {
		t.Fatalf("counters = forwarded %d suppressed %d, want 1/1",
			sess.DualForwarded.Load(), sess.DualSuppressed.Load())
	}

	// 2. A frame with a mismatched group hash is dropped, never forwarded.
	badFramed := protocol.AppendDualTrailer(inner, 2, client.groupHash^0xFFFFFFFF)
	client.send(t, badFramed)
	expectNoForwarded(t, forwarded, "wrong group hash")
	sess = activeSession(t, srv)
	if sess.DualBadFrames.Load() != 1 {
		t.Fatalf("bad frames = %d, want 1", sess.DualBadFrames.Load())
	}

	// 3. A sequence repeated after the dedup TTL is forwarded again: a
	// retransmission of a previously lost frame must survive.
	time.Sleep(dualSeqTTL + 150*time.Millisecond)
	client.send(t, protocol.AppendDualTrailer(inner, 1, client.groupHash))
	expectForwarded(t, forwarded, inner)

	// 4. A plain (non-dual) session is never deduplicated: two different
	// outer packets carrying identical inner bytes both reach the TUN.
	plain := dialTestClient(t, udpAddr, psk, 0xF00D, nil)
	if plain.dualOK {
		t.Fatal("legacy handshake must not receive a dual capability echo")
	}
	plainInner := buildInnerPacket(t, plain.innerIP, netip.MustParseAddr("8.8.4.4"), 2)
	plain.send(t, plainInner)
	expectForwarded(t, forwarded, plainInner)
	// Second, distinct outer packet with the same inner content.
	plain.send(t, plainInner)
	expectForwarded(t, forwarded, plainInner)
}
