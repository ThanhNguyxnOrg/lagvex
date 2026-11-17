package protocol

import (
	"errors"
	"net/netip"
	"testing"
	"time"
)

func TestDualHandshakeExtensionRoundTrip(t *testing.T) {
	psk := []byte("relay-psk")
	req := HandshakeRequest{Nonce: 0x1122334455667788, Timestamp: time.Now().Unix(), ClientID: 0xAABB}
	var groupID [DualGroupIDLen]byte
	for i := range groupID {
		groupID[i] = byte(i + 1)
	}

	dual := EncodeHandshakeRequestDual(psk, req, groupID)
	if len(dual) != HandshakeReqLen+DualExtLen {
		t.Fatalf("dual handshake length = %d, want %d", len(dual), HandshakeReqLen+DualExtLen)
	}

	// A relay without dual support must still authenticate the base fields.
	base, err := DecodeHandshakeRequest(psk, dual)
	if err != nil {
		t.Fatalf("old decode of dual handshake: %v", err)
	}
	if base.Nonce != req.Nonce || base.ClientID != req.ClientID {
		t.Fatalf("old decode mismatch: %+v", base)
	}

	parsed, gotGroup, err := DecodeHandshakeRequestExt(psk, dual)
	if err != nil {
		t.Fatalf("ext decode: %v", err)
	}
	if gotGroup == nil || *gotGroup != groupID {
		t.Fatalf("group id mismatch: %v", gotGroup)
	}
	if parsed.Nonce != req.Nonce || parsed.ClientID != req.ClientID {
		t.Fatalf("request mismatch: %+v", parsed)
	}
}

func TestDualHandshakeExtensionOptional(t *testing.T) {
	psk := []byte("relay-psk")
	plain := EncodeHandshakeRequest(psk, HandshakeRequest{Nonce: 1, Timestamp: time.Now().Unix(), ClientID: 2})

	parsed, groupID, err := DecodeHandshakeRequestExt(psk, plain)
	if err != nil {
		t.Fatalf("plain decode: %v", err)
	}
	if groupID != nil {
		t.Fatalf("plain handshake must not carry a group, got %v", groupID)
	}
	if parsed.ClientID != 2 {
		t.Fatalf("client id mismatch: %+v", parsed)
	}

	// Wrong PSK must fail exactly as the base handshake does.
	if _, _, err := DecodeHandshakeRequestExt([]byte("other"), plain); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("wrong psk error = %v, want ErrAuthFailed", err)
	}
}

func TestDualHandshakeExtensionRejectedWhenMalformed(t *testing.T) {
	psk := []byte("relay-psk")
	dual := EncodeHandshakeRequestDual(psk,
		HandshakeRequest{Nonce: 5, Timestamp: time.Now().Unix(), ClientID: 6},
		[DualGroupIDLen]byte{0x0a})

	tampered := append([]byte(nil), dual...)
	tampered[HandshakeReqLen] = 0x02 // unknown extension type
	if _, _, err := DecodeHandshakeRequestExt(psk, tampered); !errors.Is(err, ErrUnknownHandshakeExtension) {
		t.Fatalf("unknown ext type error = %v", err)
	}

	truncated := dual[:len(dual)-1]
	if _, _, err := DecodeHandshakeRequestExt(psk, truncated); !errors.Is(err, ErrUnknownHandshakeExtension) {
		t.Fatalf("truncated ext error = %v", err)
	}
}

func TestDualResponseCapability(t *testing.T) {
	psk := []byte("relay-psk")
	resp := HandshakeResponse{
		Status:    StatusOK,
		SessionID: 9,
		ClientIP:  netip.MustParseAddr("10.99.0.5"),
		GatewayIP: netip.MustParseAddr("10.99.0.1"),
		MTU:       1400,
		NonceEcho: 5,
	}

	dual := EncodeHandshakeResponseDual(psk, resp)
	if len(dual) != HandshakeRespLen+DualRespExtLen {
		t.Fatalf("dual response length = %d, want %d", len(dual), HandshakeRespLen+DualRespExtLen)
	}

	// An old client decoding a dual-capable response must still succeed.
	base, err := DecodeHandshakeResponse(psk, dual, 5)
	if err != nil {
		t.Fatalf("old decode of dual response: %v", err)
	}
	if base.SessionID != 9 || base.MTU != 1400 {
		t.Fatalf("old decode mismatch: %+v", base)
	}

	parsed, dualOK, err := DecodeHandshakeResponseExt(psk, dual, 5)
	if err != nil {
		t.Fatalf("ext decode: %v", err)
	}
	if !dualOK {
		t.Fatal("dual response must advertise dual support")
	}
	if parsed.SessionID != resp.SessionID || parsed.ClientIP != resp.ClientIP {
		t.Fatalf("response mismatch: %+v", parsed)
	}

	plain := EncodeHandshakeResponse(psk, resp)
	if _, dualOK, err := DecodeHandshakeResponseExt(psk, plain, 5); err != nil || dualOK {
		t.Fatalf("plain response dual=%v err=%v, want false/nil", dualOK, err)
	}

	tampered := append([]byte(nil), dual...)
	tampered[HandshakeRespLen] = 0x09
	if _, _, err := DecodeHandshakeResponseExt(psk, tampered, 5); !errors.Is(err, ErrUnknownHandshakeExtension) {
		t.Fatalf("tampered ext error = %v", err)
	}
}

func TestDualAckRoundTrip(t *testing.T) {
	psk := []byte("relay-psk")
	cl, err := NewClientCrypto(psk, 0xABCD, 42, 7)
	if err != nil {
		t.Fatalf("client crypto: %v", err)
	}
	rl, err := NewRelayCrypto(psk, 0xABCD, 42, 7)
	if err != nil {
		t.Fatalf("relay crypto: %v", err)
	}

	buf := make([]byte, 256)
	pkt := cl.EncodeDualAck(buf, 42, 12345)
	mtype, payload, err := rl.OpenPacket(make([]byte, 256), pkt)
	if err != nil {
		t.Fatalf("open ack: %v", err)
	}
	if mtype != TypeDualAck {
		t.Fatalf("type = %s, want DualAck", TypeString(mtype))
	}
	seq, err := DecodeDualAckPayload(payload)
	if err != nil {
		t.Fatalf("decode ack: %v", err)
	}
	if seq != 12345 {
		t.Fatalf("ack seq = %d, want 12345", seq)
	}

	if _, err := DecodeDualAckPayload(payload[:4]); !errors.Is(err, ErrPacketTooShort) {
		t.Fatalf("short ack error = %v", err)
	}
}

func TestGroupHashBindsPSKAndGroup(t *testing.T) {
	psk := []byte("relay-psk")
	var groupID [DualGroupIDLen]byte

	h1 := GroupHash(psk, groupID[:])
	if h1 != GroupHash(psk, groupID[:]) {
		t.Fatal("group hash must be deterministic")
	}
	if h1 == GroupHash([]byte("other-psk"), groupID[:]) {
		t.Fatal("group hash must depend on the PSK")
	}
	other := groupID
	other[0] ^= 0xFF
	if h1 == GroupHash(psk, other[:]) {
		t.Fatal("group hash must depend on the group ID")
	}
}

func TestDualTrailerRoundTrip(t *testing.T) {
	inner := make([]byte, 40)
	inner[0] = 0x45
	framed := AppendDualTrailer(inner, 7, 0xDEADBEEF)
	if len(framed) != len(inner)+DualTrailerLen {
		t.Fatalf("framed length = %d", len(framed))
	}

	got, seq, groupHash, ok := SplitDualTrailer(framed)
	if !ok {
		t.Fatal("split failed on a valid frame")
	}
	if seq != 7 || groupHash != 0xDEADBEEF {
		t.Fatalf("trailer values = seq %d group %08x", seq, groupHash)
	}
	if string(got) != string(inner) {
		t.Fatal("inner payload was modified")
	}

	if _, _, _, ok := SplitDualTrailer(make([]byte, DualMinInnerLen+DualTrailerLen-1)); ok {
		t.Fatal("split must fail on a short payload")
	}
}
