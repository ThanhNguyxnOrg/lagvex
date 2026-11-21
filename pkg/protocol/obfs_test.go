package protocol

import (
	"bytes"
	"testing"
	"time"
)

func TestObfuscatorSessionRoundTrip(t *testing.T) {
	psk := []byte("anti-dpi-test-psk-0123456789")
	obfs := NewObfuscator(psk)

	cCrypto, err := NewClientCrypto(psk, 11, 22, 33)
	if err != nil {
		t.Fatalf("NewClientCrypto: %v", err)
	}
	rCrypto, err := NewRelayCrypto(psk, 11, 22, 33)
	if err != nil {
		t.Fatalf("NewRelayCrypto: %v", err)
	}

	ipPkt := make([]byte, 48)
	ipPkt[0] = 0x45
	plain := cCrypto.EncodeData(make([]byte, MaxPacketSize), 22, ipPkt)

	masked := obfs.ObfuscateSession(nil, plain)
	if len(masked) != len(plain) {
		t.Fatalf("expected 0-byte overhead: got %d want %d", len(masked), len(plain))
	}
	if bytes.Equal(masked[:SecureHeaderLen], plain[:SecureHeaderLen]) {
		t.Fatal("obfuscated header must differ from plaintext header")
	}

	unmasked := obfs.DeobfuscateSession(nil, masked)
	if !bytes.Equal(unmasked, plain) {
		t.Fatal("deobfuscated session packet mismatch")
	}

	mtype, dec, err := rCrypto.OpenPacket(nil, unmasked)
	if err != nil || mtype != TypeData || !bytes.Equal(dec, ipPkt) {
		t.Fatalf("OpenPacket failed after deobfuscation: mtype=%d err=%v", mtype, err)
	}
}

func TestObfuscatorHandshakeRoundTripAndLengthVariance(t *testing.T) {
	psk := []byte("anti-dpi-handshake-psk")
	obfs := NewObfuscator(psk)

	req := HandshakeRequest{
		Nonce:     0x123456789abcdef0,
		Timestamp: time.Now().Unix(),
		ClientID:  0xdeadbeef,
	}
	plain := EncodeHandshakeRequest(psk, req)

	lengths := map[int]bool{}
	for i := 0; i < 16; i++ {
		masked := obfs.ObfuscateHandshake(plain)
		lengths[len(masked)] = true
		if len(masked) <= len(plain) {
			t.Fatalf("expected random padding >0, got len=%d", len(masked))
		}
		unmasked, ok := obfs.DeobfuscateHandshake(masked)
		if !ok {
			t.Fatalf("DeobfuscateHandshake failed on iteration %d", i)
		}
		if !bytes.Equal(unmasked, plain) {
			t.Fatalf("handshake round-trip mismatch on iteration %d", i)
		}
		parsed, err := DecodeHandshakeRequest(psk, unmasked)
		if err != nil || parsed.Nonce != req.Nonce {
			t.Fatalf("DecodeHandshakeRequest failed: %v", err)
		}
	}
	if len(lengths) < 2 {
		t.Fatalf("expected variable handshake lengths for anti-DPI padding, saw only %v", lengths)
	}
}
