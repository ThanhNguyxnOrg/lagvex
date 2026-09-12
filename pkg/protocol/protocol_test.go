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

func TestDataPacket(t *testing.T) {
	sessionID := uint64(0x1122334455667788)
	// Sample minimal IPv4 packet (20 bytes header)
	rawIPv4 := make([]byte, 28)
	rawIPv4[0] = 0x45 // IPv4, IHL 5
	rawIPv4[9] = 17   // UDP
	copy(rawIPv4[20:], []byte("DATA"))

	buf := make([]byte, 1024)
	packet := EncodeDataPacket(buf, sessionID, rawIPv4)

	if len(packet) != DataHeaderLen+len(rawIPv4) {
		t.Fatalf("expected packet length %d, got %d", DataHeaderLen+len(rawIPv4), len(packet))
	}

	decSession, decPayload, err := DecodeDataPacket(packet)
	if err != nil {
		t.Fatalf("decode data packet failed: %v", err)
	}
	if decSession != sessionID {
		t.Fatalf("expected session ID %d, got %d", sessionID, decSession)
	}
	if !bytes.Equal(decPayload, rawIPv4) {
		t.Fatalf("payload mismatch")
	}
}

func TestPingPong(t *testing.T) {
	sessionID := uint64(555)
	ts := uint64(time.Now().UnixNano())

	ping := EncodePing(sessionID, ts)
	sID, rts, err := DecodePing(ping)
	if err != nil || sID != sessionID || rts != ts {
		t.Fatalf("ping failed: %v, sID=%d, rts=%d", err, sID, rts)
	}

	pong := EncodePong(sessionID, ts)
	sID2, rts2, err := DecodePong(pong)
	if err != nil || sID2 != sessionID || rts2 != ts {
		t.Fatalf("pong failed: %v, sID=%d, rts=%d", err, sID2, rts2)
	}
}
