package protocol

import (
	"bytes"
	"net/netip"
	"testing"
	"time"
)

func TestHandshakeSuccess(t *testing.T) {
	psk := []byte("super-secret-pre-shared-key-32chars!!")
	nonce := uint64(1234567890123456)
	clientID := uint64(9876543210987654)
	ts := time.Now().Unix()

	req := HandshakeRequest{
		Nonce:     nonce,
		Timestamp: ts,
		ClientID:  clientID,
	}

	encoded := EncodeHandshakeRequest(psk, req)
	if len(encoded) != HandshakeReqLen {
		t.Fatalf("expected length %d, got %d", HandshakeReqLen, len(encoded))
	}

	decoded, err := DecodeHandshakeRequest(psk, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.Nonce != nonce || decoded.ClientID != clientID || decoded.Timestamp != ts {
		t.Fatalf("decoded mismatch: %+v", decoded)
	}

	// Test Response
	sessionID := uint64(8888999911112222)
	cip := netip.MustParseAddr("10.88.0.5")
	gip := netip.MustParseAddr("10.88.0.1")
	resp := HandshakeResponse{
		Status:    StatusOK,
		SessionID: sessionID,
		ClientIP:  cip,
		GatewayIP: gip,
		MTU:       1400,
		NonceEcho: nonce,
	}

	respEncoded := EncodeHandshakeResponse(psk, resp)
	if len(respEncoded) != HandshakeRespLen {
		t.Fatalf("expected resp len %d, got %d", HandshakeRespLen, len(respEncoded))
	}

	respDecoded, err := DecodeHandshakeResponse(psk, respEncoded, nonce)
	if err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if respDecoded.Status != StatusOK || respDecoded.SessionID != sessionID ||
		respDecoded.ClientIP != cip || respDecoded.GatewayIP != gip || respDecoded.MTU != 1400 {
		t.Fatalf("decoded response mismatch: %+v", respDecoded)
	}
}

func TestHandshakeAuthFailure(t *testing.T) {
	psk := []byte("correct-key")
	wrongPsk := []byte("wrong-key-attacker")

	req := HandshakeRequest{
		Nonce:     1,
		Timestamp: time.Now().Unix(),
		ClientID:  2,
	}

	encoded := EncodeHandshakeRequest(wrongPsk, req)
	_, err := DecodeHandshakeRequest(psk, encoded)
	if err != ErrAuthFailed {
		t.Fatalf("expected ErrAuthFailed, got %v", err)
	}
}

func TestSessionCryptoData(t *testing.T) {
	psk := []byte("secret-key-for-session-crypto-32b")
	nonce := uint64(999888777)
	sessionID := uint64(0x1122334455667788)
	clientID := uint64(1001)

	clientCrypto, err := NewClientCrypto(psk, nonce, sessionID, clientID)
	if err != nil {
		t.Fatalf("failed to create client crypto: %v", err)
	}
	relayCrypto, err := NewRelayCrypto(psk, nonce, sessionID, clientID)
	if err != nil {
		t.Fatalf("failed to create relay crypto: %v", err)
	}

	// Minimal IPv4 packet (28 bytes)
	rawIPv4 := make([]byte, 28)
	rawIPv4[0] = 0x45
	rawIPv4[9] = 17
	copy(rawIPv4[20:], []byte("DATA"))

	buf := make([]byte, 1024)
	sealed := clientCrypto.EncodeData(buf, sessionID, rawIPv4)

	mtype, sid, err := ReadPacketHeader(sealed)
	if err != nil || mtype != TypeData || sid != sessionID {
		t.Fatalf("ReadPacketHeader failed: err=%v, mtype=%d, sid=%x", err, mtype, sid)
	}

	plainBuf := make([]byte, 1024)
	decType, payload, err := relayCrypto.OpenPacket(plainBuf, sealed)
	if err != nil {
		t.Fatalf("OpenPacket failed: %v", err)
	}
	if decType != TypeData {
		t.Fatalf("expected TypeData, got %d", decType)
	}
	if !bytes.Equal(payload, rawIPv4) {
		t.Fatalf("payload mismatch")
	}
}

func TestSessionCryptoPingPong(t *testing.T) {
	psk := []byte("secret-key-for-session-crypto-32b")
	nonce := uint64(111222333)
	sessionID := uint64(0xaabbccddeeff0011)
	clientID := uint64(2002)

	clientCrypto, _ := NewClientCrypto(psk, nonce, sessionID, clientID)
	relayCrypto, _ := NewRelayCrypto(psk, nonce, sessionID, clientID)

	ts := uint64(time.Now().UnixNano())
	buf := make([]byte, 256)

	// Client -> Relay: Ping
	pingSealed := clientCrypto.EncodePing(buf, sessionID, ts)
	mtype, pingPayload, err := relayCrypto.OpenPacket(nil, pingSealed)
	if err != nil || mtype != TypePing {
		t.Fatalf("relay open ping failed: err=%v, mtype=%d", err, mtype)
	}
	decTs, err := DecodeControlPayload(pingPayload)
	if err != nil || decTs != ts {
		t.Fatalf("relay decode ping ts failed: %v, decTs=%d, orig=%d", err, decTs, ts)
	}

	// Relay -> Client: Pong
	pongSealed := relayCrypto.EncodePong(buf, sessionID, ts)
	mtype2, pongPayload, err := clientCrypto.OpenPacket(nil, pongSealed)
	if err != nil || mtype2 != TypePong {
		t.Fatalf("client open pong failed: err=%v, mtype=%d", err, mtype2)
	}
	decTs2, err := DecodeControlPayload(pongPayload)
	if err != nil || decTs2 != ts {
		t.Fatalf("client decode pong ts failed: %v, decTs2=%d, orig=%d", err, decTs2, ts)
	}
}

func TestSessionCryptoReplayAndTamper(t *testing.T) {
	psk := []byte("secret-key-for-session-crypto-32b")
	nonce := uint64(555666777)
	sessionID := uint64(0x3344556677889900)
	clientID := uint64(3003)

	clientCrypto, _ := NewClientCrypto(psk, nonce, sessionID, clientID)
	relayCrypto, _ := NewRelayCrypto(psk, nonce, sessionID, clientID)

	data := []byte("important game packet payload")
	buf := make([]byte, 256)
	sealed := clientCrypto.EncodeData(buf, sessionID, data)

	// 1. Legitimate decrypt
	_, _, err := relayCrypto.OpenPacket(nil, sealed)
	if err != nil {
		t.Fatalf("first decrypt failed: %v", err)
	}

	// 2. Replay attack: duplicate packet should fail
	_, _, err = relayCrypto.OpenPacket(nil, sealed)
	if err != ErrReplay {
		t.Fatalf("expected ErrReplay for duplicate packet, got %v", err)
	}

	// 3. Tamper attack: modify payload byte
	sealed2 := clientCrypto.EncodeData(buf, sessionID, data)
	tampered := make([]byte, len(sealed2))
	copy(tampered, sealed2)
	tampered[len(tampered)-1] ^= 0xff // flip bit in Poly1305 tag

	_, _, err = relayCrypto.OpenPacket(nil, tampered)
	if err != ErrAuthFailed {
		t.Fatalf("expected ErrAuthFailed for tampered tag, got %v", err)
	}
}

func TestSessionCryptoOutOfOrderWithinWindow(t *testing.T) {
	psk := []byte("secret-key-for-session-crypto-32b")
	nonce := uint64(444333222)
	sessionID := uint64(0x778899aabbccddee)
	clientID := uint64(4004)

	clientCrypto, _ := NewClientCrypto(psk, nonce, sessionID, clientID)
	relayCrypto, _ := NewRelayCrypto(psk, nonce, sessionID, clientID)

	p1 := clientCrypto.EncodeData(make([]byte, 128), sessionID, []byte("p1"))
	p2 := clientCrypto.EncodeData(make([]byte, 128), sessionID, []byte("p2"))
	p3 := clientCrypto.EncodeData(make([]byte, 128), sessionID, []byte("p3"))

	// Receive p3 first, then p1, then p2
	if _, _, err := relayCrypto.OpenPacket(nil, p3); err != nil {
		t.Fatalf("open p3 failed: %v", err)
	}
	if _, _, err := relayCrypto.OpenPacket(nil, p1); err != nil {
		t.Fatalf("open p1 (out of order) failed: %v", err)
	}
	if _, _, err := relayCrypto.OpenPacket(nil, p2); err != nil {
		t.Fatalf("open p2 (out of order) failed: %v", err)
	}

	// Now replay p1 -> must fail
	if _, _, err := relayCrypto.OpenPacket(nil, p1); err != ErrReplay {
		t.Fatalf("expected ErrReplay for replayed p1, got %v", err)
	}
}
