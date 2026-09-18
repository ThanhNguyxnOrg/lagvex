package protocol

import (
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
)

// Secure header layout shared by all AEAD-protected messages:
//
//	[0]      wire header (version<<4 | message type)
//	[1:9]    session ID (plaintext, used for demux before decryption)
//	[9:17]   64-bit send counter (per direction, monotonic, starts at 1)
//	[17:..]  ciphertext = ChaCha20-Poly1305(payload) with a 16-byte Poly1305 tag
//
// The first 17 bytes double as the AEAD additional data, so any tampering
// with the header, session ID or counter fails authentication.
const (
	// SecureHeaderLen is the plaintext prefix of every AEAD-protected packet.
	SecureHeaderLen = 17
	// TagLen is the Poly1305 authentication tag length.
	TagLen = chacha20poly1305.Overhead
	// DataOverheadV2 is the total per-packet overhead of an encrypted Data
	// message compared to the raw IPv4 payload (header + counter + tag).
	DataOverheadV2 = SecureHeaderLen + TagLen
	// CipherKeyLen is the ChaCha20 key length.
	CipherKeyLen = chacha20poly1305.KeySize
)

var (
	ErrReplay            = errors.New("packet counter is duplicated or too old (replay)")
	ErrCipherClosed      = errors.New("session cipher not established")
	ErrBadControlPayload = errors.New("control message payload malformed")
)

// DeriveSessionKeys expands the PSK and the fresh handshake entropy into two
// direction keys: c2s (client→relay) and s2c (relay→client). Both parties
// derive identical keys from values they already share after the handshake
// (client nonce, server session ID, client ID), so no key material ever
// crosses the wire.
func DeriveSessionKeys(psk []byte, nonce, sessionID, clientID uint64) (c2s, s2c [CipherKeyLen]byte) {
	expand := func(label string, material []byte) [CipherKeyLen]byte {
		h := hmac.New(sha256.New, material)
		h.Write([]byte(label))
		var key [CipherKeyLen]byte
		copy(key[:], h.Sum(nil))
		return key
	}

	// keyMaterial = HMAC-PSK(nonce || sessionID || clientID || domain separator)
	h := hmac.New(sha256.New, psk)
	var scratch [24]byte
	binary.BigEndian.PutUint64(scratch[0:8], nonce)
	binary.BigEndian.PutUint64(scratch[8:16], sessionID)
	binary.BigEndian.PutUint64(scratch[16:24], clientID)
	h.Write(scratch[:])
	h.Write([]byte("lagvex-v2-session-keys"))
	material := h.Sum(nil)

	c2s = expand("lagvex-c2s", material)
	s2c = expand("lagvex-s2c", material)
	return c2s, s2c
}

// NewClientCrypto builds the client-side session cipher (sends with c2s key).
func NewClientCrypto(psk []byte, nonce, sessionID, clientID uint64) (*SessionCrypto, error) {
	return newSessionCrypto(psk, nonce, sessionID, clientID, true)
}

// NewRelayCrypto builds the relay-side session cipher (sends with s2c key).
func NewRelayCrypto(psk []byte, nonce, sessionID, clientID uint64) (*SessionCrypto, error) {
	return newSessionCrypto(psk, nonce, sessionID, clientID, false)
}

func newSessionCrypto(psk []byte, nonce, sessionID, clientID uint64, isClient bool) (*SessionCrypto, error) {
	c2s, s2c := DeriveSessionKeys(psk, nonce, sessionID, clientID)

	sendKey, recvKey := c2s, s2c
	if !isClient {
		sendKey, recvKey = s2c, c2s
	}

	sendAEAD, err := chacha20poly1305.New(sendKey[:])
	if err != nil {
		return nil, fmt.Errorf("init send cipher: %w", err)
	}
	recvAEAD, err := chacha20poly1305.New(recvKey[:])
	if err != nil {
		return nil, fmt.Errorf("init recv cipher: %w", err)
	}

	return &SessionCrypto{
		sendAEAD: sendAEAD,
		recvAEAD: recvAEAD,
	}, nil
}

// SessionCrypto encrypts and authenticates all post-handshake traffic of one
// session. It is safe for concurrent use: every outgoing packet consumes a
// unique monotonic counter (so ChaCha20 nonces are never reused), and every
// incoming packet is checked against a sliding replay window before being
// accepted.
type SessionCrypto struct {
	mu       sync.Mutex
	sendAEAD cipher.AEAD
	recvAEAD cipher.AEAD
	sendCtr  uint64
	recvWin  replayWindow
}

// SealPacket builds an AEAD-protected message of the given type. dst must
// have capacity for SecureHeaderLen+len(payload)+TagLen; the returned slice
// aliases dst.
func (c *SessionCrypto) SealPacket(dst []byte, msgType byte, sessionID uint64, payload []byte) []byte {
	c.mu.Lock()
	c.sendCtr++
	ctr := c.sendCtr
	c.mu.Unlock()

	out := dst[:SecureHeaderLen+len(payload)+TagLen]
	out[0] = HeaderByte(msgType)
	binary.BigEndian.PutUint64(out[1:9], sessionID)
	binary.BigEndian.PutUint64(out[9:17], ctr)

	var nonce [chacha20poly1305.NonceSize]byte
	binary.BigEndian.PutUint64(nonce[4:12], ctr)

	// AAD = plaintext header (version|type|sessionID|counter)
	return c.sendAEAD.Seal(out[:SecureHeaderLen], nonce[:], payload, out[:SecureHeaderLen])
}

// OpenPacket authenticates and decrypts a sealed message. The plaintext
// payload is appended to dst (which may be nil) and returned together with
// the message type. Replay-detection failures and tag failures both return
// an error and the packet must be dropped.
func (c *SessionCrypto) OpenPacket(dst []byte, buf []byte) (msgType byte, payload []byte, err error) {
	if len(buf) < SecureHeaderLen+TagLen {
		return 0, nil, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, nil, ErrInvalidVersion
	}

	ctr := binary.BigEndian.Uint64(buf[9:17])
	var nonce [chacha20poly1305.NonceSize]byte
	binary.BigEndian.PutUint64(nonce[4:12], ctr)

	plain, err := c.recvAEAD.Open(dst[:0] /* cap preserved */, nonce[:], buf[SecureHeaderLen:], buf[:SecureHeaderLen])
	if err != nil {
		return 0, nil, ErrAuthFailed
	}

	c.mu.Lock()
	fresh := c.recvWin.accept(ctr)
	c.mu.Unlock()
	if !fresh {
		return 0, nil, ErrReplay
	}

	return mtype, plain, nil
}

// EncodeData seals a raw IPv4 packet as a Data message.
func (c *SessionCrypto) EncodeData(dst []byte, sessionID uint64, ipPacket []byte) []byte {
	return c.SealPacket(dst, TypeData, sessionID, ipPacket)
}

// EncodePing seals a keepalive Ping carrying the client monotonic timestamp.
func (c *SessionCrypto) EncodePing(dst []byte, sessionID uint64, tsNano uint64) []byte {
	var p [8]byte
	binary.BigEndian.PutUint64(p[:], tsNano)
	return c.SealPacket(dst, TypePing, sessionID, p[:])
}

// EncodePong seals a keepalive Pong echoing the ping timestamp.
func (c *SessionCrypto) EncodePong(dst []byte, sessionID uint64, tsNano uint64) []byte {
	var p [8]byte
	binary.BigEndian.PutUint64(p[:], tsNano)
	return c.SealPacket(dst, TypePong, sessionID, p[:])
}

// EncodeDisconnect seals an empty Disconnect message.
func (c *SessionCrypto) EncodeDisconnect(dst []byte, sessionID uint64) []byte {
	return c.SealPacket(dst, TypeDisconnect, sessionID, nil)
}

// DecodeControlPayload extracts the 8-byte timestamp carried by Ping/Pong.
func DecodeControlPayload(payload []byte) (uint64, error) {
	if len(payload) != 8 {
		return 0, ErrBadControlPayload
	}
	return binary.BigEndian.Uint64(payload), nil
}

// ReadPacketHeader extracts the message type and plaintext session ID from a
// sealed packet so the relay can demux the session before decryption.
func ReadPacketHeader(buf []byte) (msgType byte, sessionID uint64, err error) {
	if len(buf) < SecureHeaderLen {
		return 0, 0, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, 0, ErrInvalidVersion
	}
	return mtype, binary.BigEndian.Uint64(buf[1:9]), nil
}

// --- Sliding replay window (RFC 6479 style) ---

const replayWindowBits = 1024

// replayWindow tracks accepted packet counters. A counter is accepted if it
// is new (greater than the highest seen) or falls within the trailing
// replayWindowBits window and has not been observed yet. Out-of-order UDP
// delivery within the window therefore still works, while exact replays and
// anything older than the window are rejected.
type replayWindow struct {
	last uint64
	bits [replayWindowBits / 64]uint64
}

func (w *replayWindow) get(seq uint64) bool {
	pos := seq % replayWindowBits
	return w.bits[pos/64]&(1<<(pos%64)) != 0
}

func (w *replayWindow) set(seq uint64) {
	pos := seq % replayWindowBits
	w.bits[pos/64] |= 1 << (pos % 64)
}

func (w *replayWindow) clear(seq uint64) {
	pos := seq % replayWindowBits
	w.bits[pos/64] &^= 1 << (pos % 64)
}

// accept reports whether seq is fresh. Callers must hold the owning mutex.
func (w *replayWindow) accept(seq uint64) bool {
	if seq == 0 {
		return false
	}
	if seq > w.last {
		if seq-w.last >= replayWindowBits {
			// Window jumped entirely: reset every bit (positions alias mod 1024).
			w.bits = [replayWindowBits / 64]uint64{}
		} else {
			// Clear positions the advanced window now covers so that late,
			// legitimately out-of-order packets in (old last, new last] are not
			// mistaken for replays of stale bits from seq-1024.
			for s := w.last + 1; s <= seq; s++ {
				w.clear(s)
			}
		}
		w.last = seq
		w.set(seq)
		return true
	}
	if w.last-seq >= replayWindowBits {
		return false // older than the window
	}
	if w.get(seq) {
		return false // duplicate
	}
	w.set(seq)
	return true
}
