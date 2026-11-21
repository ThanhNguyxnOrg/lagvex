package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
)

// Anti-DPI UDP Obfuscation (AmneziaWG-inspired).
//
// Plain Lagvex packets have two static DPI fingerprints:
//   1. Byte 0 is always 0x11..0x1A (Version=1 | MsgType), followed by a 16-byte
//      plaintext SessionID + Counter header.
//   2. Handshake packets have fixed lengths (57B / 60B / 74B / 64B).
//
// When obfuscation is enabled:
//   - Sealed session packets (Data, Ping, Pong, FEC, DualAck, AnchorProbe,
//     Disconnect) XOR bytes[0:17] with SHA-256("lagvex-obfs-v1" || PSK || tag[0:8]),
//     where tag is the cryptographically random 16-byte Poly1305 tag at the
//     tail of the packet. Overhead: 0 bytes. Every byte on the wire looks like
//     uniform random noise.
//   - Handshake packets XOR their non-HMAC prefix with
//     SHA-256("lagvex-obfs-hs" || PSK || salt) and append 4..35 random junk
//     bytes plus a 1-byte masked pad length and a 4-byte random salt so
//     handshake sizes vary across packets.
//
// Relays transparently accept both plain and obfuscated clients on the same
// UDP port: plain validation runs first, and obfuscated unmasking runs only if
// the packet is not valid plaintext.

const (
	obfsMaxHandshakePad = 32
	obfsMinHandshakePad = 4
	obfsHandshakeTail   = 5 // 4-byte salt + 1-byte masked padLen
)

// Obfuscator caches the PSK-derived key material for fast per-packet masking.
type Obfuscator struct {
	key [32]byte
}

// NewObfuscator derives an obfuscation key from the relay PSK.
func NewObfuscator(psk []byte) *Obfuscator {
	h := sha256.New()
	h.Write([]byte("lagvex-obfs-key-v1:"))
	h.Write(psk)
	var o Obfuscator
	copy(o.key[:], h.Sum(nil))
	return &o
}

func (o *Obfuscator) headerMask(salt []byte) [32]byte {
	h := sha256.New()
	h.Write(o.key[:])
	h.Write(salt)
	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum
}

// ObfuscateSession masks the 17-byte plaintext header of an AEAD-sealed packet
// in place (returning a slice aliasing dst or a new copy if dst is nil) with
// zero wire overhead.
func (o *Obfuscator) ObfuscateSession(dst, pkt []byte) []byte {
	if len(pkt) < SecureHeaderLen+TagLen {
		return pkt
	}
	var out []byte
	if cap(dst) >= len(pkt) {
		out = dst[:len(pkt)]
	} else {
		out = make([]byte, len(pkt))
	}
	copy(out, pkt)
	// Use the first 8 bytes of the trailing Poly1305 tag as per-packet salt.
	salt := out[len(out)-TagLen : len(out)-TagLen+8]
	mask := o.headerMask(salt)
	for i := 0; i < SecureHeaderLen; i++ {
		out[i] ^= mask[i]
	}
	return out
}

// DeobfuscateSession reverses ObfuscateSession (XOR is self-inverse).
func (o *Obfuscator) DeobfuscateSession(dst, pkt []byte) []byte {
	return o.ObfuscateSession(dst, pkt)
}

// ObfuscateHandshake masks a HandshakeReq or HandshakeResp and appends random
// junk padding (4..35 bytes) so DPI cannot match fixed handshake packet sizes.
func (o *Obfuscator) ObfuscateHandshake(pkt []byte) []byte {
	var saltAndPad [5]byte
	_, _ = rand.Read(saltAndPad[:])
	padLen := obfsMinHandshakePad + int(saltAndPad[4])%(obfsMaxHandshakePad-obfsMinHandshakePad+1)

	total := len(pkt) + padLen + obfsHandshakeTail
	out := make([]byte, total)
	copy(out, pkt)

	// Random junk padding
	_, _ = rand.Read(out[len(pkt) : len(pkt)+padLen])

	// 4-byte random salt
	saltOff := len(pkt) + padLen
	copy(out[saltOff:saltOff+4], saltAndPad[:4])

	// Derive streaming mask from (key, salt)
	mask := o.headerMask(saltAndPad[:4])
	// XOR every payload byte with repeating keyed stream (counter-mixed per 32B block)
	xorStream(o.key[:], saltAndPad[:4], out[:len(pkt)])

	// Final byte stores padLen masked with mask[0]
	out[total-1] = byte(padLen) ^ mask[0]
	return out
}

// DeobfuscateHandshake reverses ObfuscateHandshake and returns the inner
// handshake bytes, or (nil, false) if lengths are inconsistent.
func (o *Obfuscator) DeobfuscateHandshake(pkt []byte) ([]byte, bool) {
	if len(pkt) < HandshakeReqLen+obfsMinHandshakePad+obfsHandshakeTail {
		return nil, false
	}
	salt := pkt[len(pkt)-obfsHandshakeTail : len(pkt)-1]
	mask := o.headerMask(salt)
	padLen := int(pkt[len(pkt)-1] ^ mask[0])
	if padLen < obfsMinHandshakePad || padLen > obfsMaxHandshakePad {
		return nil, false
	}
	innerLen := len(pkt) - obfsHandshakeTail - padLen
	if innerLen < HandshakeReqLen {
		return nil, false
	}

	out := make([]byte, innerLen)
	copy(out, pkt[:innerLen])
	xorStream(o.key[:], salt, out)
	return out, true
}

func xorStream(key, salt, buf []byte) {
	var ctr [4]byte
	for off := 0; off < len(buf); off += 32 {
		binary.BigEndian.PutUint32(ctr[:], uint32(off/32))
		h := sha256.New()
		h.Write(key)
		h.Write(salt)
		h.Write(ctr[:])
		block := h.Sum(nil)
		limit := len(buf) - off
		if limit > 32 {
			limit = 32
		}
		for i := 0; i < limit; i++ {
			buf[off+i] ^= block[i]
		}
	}
}
