package protocol

// Dual-path (Redundant-lite) wire support. A dual-mode client sends the same
// data frame to two relays; the relay that sees a sequence number first
// forwards it and every later copy is dropped. See docs/PROTOCOL.md section 9.
//
// The group identity is a random 16-byte connection group ID chosen by the
// client. Relays derive a dedup scope hash from (PSK, group ID), so only
// holders of the relay's PSK can land in a given group: an attacker who can
// compute the hash can already authenticate to the relay.

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

const (
	// DualExtType marks the dual-path handshake extension.
	DualExtType byte = 0x01

	// DualGroupIDLen is the size of the random connection group ID.
	DualGroupIDLen = 16

	// DualExtLen is the appended handshake extension size: one type byte plus
	// the group ID. It follows the 57-byte base handshake request, whose PSK
	// HMAC is unchanged so old relays still authenticate the base fields.
	DualExtLen = 1 + DualGroupIDLen

	// DualTrailerLen is the per-frame trailer appended to dual-mode data
	// payloads: sequence (u32) + group hash (u32).
	DualTrailerLen = 8

	// DualMinInnerLen is the smallest IPv4 packet a data payload may carry.
	DualMinInnerLen = 20
)

// ErrUnknownHandshakeExtension is returned when a handshake request carries
// bytes the decoder does not understand.
var ErrUnknownHandshakeExtension = errors.New("unknown handshake extension")

// EncodeHandshakeRequestDual serializes a handshake request plus the dual-path
// extension. The base 57 bytes and their HMAC are identical to
// EncodeHandshakeRequest, so relays without dual-path support still validate
// the handshake and simply ignore the trailing bytes.
func EncodeHandshakeRequestDual(psk []byte, req HandshakeRequest, groupID [DualGroupIDLen]byte) []byte {
	buf := make([]byte, HandshakeReqLen+DualExtLen)
	copy(buf, EncodeHandshakeRequest(psk, req))
	buf[HandshakeReqLen] = DualExtType
	copy(buf[HandshakeReqLen+1:], groupID[:])
	return buf
}

// DecodeHandshakeRequestExt parses a handshake request and, when present, the
// dual-path extension. A request without the extension returns a nil group ID.
// The extension is not covered by the PSK HMAC, which is why the relay never
// trusts the raw group ID: it derives the dedup scope via GroupHash, keyed by
// the PSK.
func DecodeHandshakeRequestExt(psk, buf []byte) (*HandshakeRequest, *[DualGroupIDLen]byte, error) {
	req, err := DecodeHandshakeRequest(psk, buf)
	if err != nil {
		return nil, nil, err
	}

	switch len(buf) {
	case HandshakeReqLen:
		return req, nil, nil
	case HandshakeReqLen + DualExtLen:
		if buf[HandshakeReqLen] != DualExtType {
			return nil, nil, ErrUnknownHandshakeExtension
		}
		var groupID [DualGroupIDLen]byte
		copy(groupID[:], buf[HandshakeReqLen+1:])
		return req, &groupID, nil
	default:
		return nil, nil, ErrUnknownHandshakeExtension
	}
}

// GroupHash derives the 32-bit dedup scope for a connection group:
// SHA-256(PSK || group ID) truncated to 4 bytes, big-endian.
func GroupHash(psk, groupID []byte) uint32 {
	h := sha256.New()
	h.Write(psk)
	h.Write(groupID)
	sum := h.Sum(nil)
	return binary.BigEndian.Uint32(sum[:4])
}

// DualRespExtLen is the dual-path capability extension appended to a
// handshake response: one type byte plus three reserved flag bytes. The base
// 60-byte response and its HMAC are unchanged, so old clients ignore it.
const DualRespExtLen = 4

// EncodeHandshakeResponseDual serializes a handshake response plus the
// dual-path capability extension. Relays send it only when the client asked
// for a dual-path group in its request.
func EncodeHandshakeResponseDual(psk []byte, resp HandshakeResponse) []byte {
	out := make([]byte, HandshakeRespLen+DualRespExtLen)
	copy(out, EncodeHandshakeResponse(psk, resp))
	out[HandshakeRespLen] = DualExtType
	// Three reserved capability flag bytes stay zero in Phase 1.
	return out
}

// DecodeHandshakeResponseExt parses a handshake response and reports whether
// the relay advertised dual-path support. Responses without the extension
// (old relays) return dual=false, which is the client's cue to stay
// single-path.
func DecodeHandshakeResponseExt(psk, buf []byte, expectedNonce uint64) (*HandshakeResponse, bool, error) {
	resp, err := DecodeHandshakeResponse(psk, buf, expectedNonce)
	if err != nil {
		return nil, false, err
	}

	switch len(buf) {
	case HandshakeRespLen:
		return resp, false, nil
	case HandshakeRespLen + DualRespExtLen:
		if buf[HandshakeRespLen] != DualExtType {
			return nil, false, ErrUnknownHandshakeExtension
		}
		return resp, true, nil
	default:
		return nil, false, ErrUnknownHandshakeExtension
	}
}

// AppendDualTrailer appends the dual-path sequence trailer to an inner IPv4
// packet. The same trailer bytes must be produced for both relays.
func AppendDualTrailer(payload []byte, seq, groupHash uint32) []byte {
	out := make([]byte, len(payload)+DualTrailerLen)
	copy(out, payload)
	binary.BigEndian.PutUint32(out[len(payload):], seq)
	binary.BigEndian.PutUint32(out[len(payload)+4:], groupHash)
	return out
}

// SplitDualTrailer removes and parses the dual-path trailer. ok is false when
// the payload is too short to be an IPv4 packet plus trailer.
func SplitDualTrailer(payload []byte) (inner []byte, seq, groupHash uint32, ok bool) {
	if len(payload) < DualMinInnerLen+DualTrailerLen {
		return nil, 0, 0, false
	}
	cut := len(payload) - DualTrailerLen
	inner = payload[:cut]
	seq = binary.BigEndian.Uint32(payload[cut : cut+4])
	groupHash = binary.BigEndian.Uint32(payload[cut+4:])
	return inner, seq, groupHash, true
}

// EncodeDualAck seals the relay's dual-path progress report: the highest dual
// frame sequence the relay has forwarded to the game server. The payload is
// 8 bytes (sequence + 4 reserved bytes) to match the other control payloads.
func (c *SessionCrypto) EncodeDualAck(dst []byte, sessionID uint64, highestSeq uint32) []byte {
	var p [8]byte
	binary.BigEndian.PutUint32(p[0:4], highestSeq)
	return c.SealPacket(dst, TypeDualAck, sessionID, p[:])
}

// DecodeDualAckPayload parses a DualAck payload and returns the highest
// forwarded sequence.
func DecodeDualAckPayload(payload []byte) (uint32, error) {
	if len(payload) < 8 {
		return 0, ErrPacketTooShort
	}
	return binary.BigEndian.Uint32(payload[0:4]), nil
}
