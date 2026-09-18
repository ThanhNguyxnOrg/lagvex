// Package protocol implements the Lagvex binary wire protocol over UDP.
// The handshake authenticates both peers via a Pre-Shared Key (HMAC-SHA256)
// and bootstraps per-session keys; all post-handshake traffic (data,
// keepalive, disconnect) is protected with ChaCha20-Poly1305 AEAD plus a
// per-direction replay window. See crypto.go for the session cipher.
package protocol

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
	"time"
)

const (
	// Version is the current Lagvex protocol version.
	Version = 1

	// Message types (4-bit nibble).
	TypeHandshakeReq  = 0x1
	TypeHandshakeResp = 0x2
	TypeData          = 0x3
	TypePing          = 0x4
	TypePong          = 0x5
	TypeDisconnect    = 0x6
	TypeFEC           = 0x7

	// Packet lengths in bytes.
	HandshakeReqLen  = 57
	HandshakeRespLen = 60

	// MaxPacketSize is the maximum MTU plus protocol overhead
	// (secure header + Poly1305 tag).
	MaxPacketSize = 2048

	// MaxClockSkew is the allowable drift window for handshake timestamps (replay guard).
	MaxClockSkew = 120 * time.Second

	// DefaultTunnelMTU is the default recommended MTU for the virtual adapter.
	DefaultTunnelMTU = 1400
)

// Handshake status codes.
const (
	StatusOK              = 0
	StatusPoolFull        = 1
	StatusServerBusy      = 2
	StatusVersionMismatch = 3
	StatusAuthFailed      = 4
)

var (
	ErrPacketTooShort    = errors.New("packet too short")
	ErrInvalidVersion    = errors.New("invalid protocol version")
	ErrInvalidType       = errors.New("invalid message type")
	ErrAuthFailed        = errors.New("HMAC verification failed")
	ErrClockSkewExceeded = errors.New("handshake timestamp is outside allowable window")
	ErrNotIPv4           = errors.New("packet payload is not valid IPv4")
	ErrNonceMismatch     = errors.New("handshake response nonce does not match request")
)

// HeaderByte returns the 1-byte wire header for a given message type.
func HeaderByte(msgType byte) byte {
	return (Version << 4) | (msgType & 0x0f)
}

// ParseHeader splits the first byte into version and message type.
func ParseHeader(b byte) (version byte, msgType byte) {
	return b >> 4, b & 0x0f
}

// ComputeHMAC generates an HMAC-SHA256 over data using the provided pre-shared key.
func ComputeHMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// VerifyHMAC verifies that expectedMac matches the HMAC of data in constant time.
func VerifyHMAC(key, data, expectedMac []byte) bool {
	actualMac := ComputeHMAC(key, data)
	return subtle.ConstantTimeCompare(actualMac, expectedMac) == 1
}

// RandomUint64 generates a cryptographically secure random uint64.
func RandomUint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b[:]), nil
}

// --- Handshake Request ---

// HandshakeRequest represents a client connect request.
type HandshakeRequest struct {
	Nonce     uint64
	Timestamp int64
	ClientID  uint64
}

// EncodeHandshakeRequest serializes and signs a HandshakeRequest into a 57-byte buffer.
func EncodeHandshakeRequest(psk []byte, req HandshakeRequest) []byte {
	buf := make([]byte, HandshakeReqLen)
	buf[0] = HeaderByte(TypeHandshakeReq)
	binary.BigEndian.PutUint64(buf[1:9], req.Nonce)
	binary.BigEndian.PutUint64(buf[9:17], uint64(req.Timestamp))
	binary.BigEndian.PutUint64(buf[17:25], req.ClientID)

	mac := ComputeHMAC(psk, buf[:25])
	copy(buf[25:57], mac)
	return buf
}

// DecodeHandshakeRequest validates and parses a 57-byte HandshakeRequest.
func DecodeHandshakeRequest(psk, buf []byte) (*HandshakeRequest, error) {
	if len(buf) < HandshakeReqLen {
		return nil, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return nil, ErrInvalidVersion
	}
	if mtype != TypeHandshakeReq {
		return nil, ErrInvalidType
	}

	if !VerifyHMAC(psk, buf[:25], buf[25:57]) {
		return nil, ErrAuthFailed
	}

	ts := int64(binary.BigEndian.Uint64(buf[9:17]))
	now := time.Now().Unix()
	diff := now - ts
	if diff < -int64(MaxClockSkew.Seconds()) || diff > int64(MaxClockSkew.Seconds()) {
		return nil, ErrClockSkewExceeded
	}

	return &HandshakeRequest{
		Nonce:     binary.BigEndian.Uint64(buf[1:9]),
		Timestamp: ts,
		ClientID:  binary.BigEndian.Uint64(buf[17:25]),
	}, nil
}

// --- Handshake Response ---

// HandshakeResponse represents a server connect reply.
type HandshakeResponse struct {
	Status    byte
	SessionID uint64
	ClientIP  netip.Addr
	GatewayIP netip.Addr
	MTU       uint16
	NonceEcho uint64
}

// EncodeHandshakeResponse serializes and signs a HandshakeResponse into a 60-byte buffer.
func EncodeHandshakeResponse(psk []byte, resp HandshakeResponse) []byte {
	buf := make([]byte, HandshakeRespLen)
	buf[0] = HeaderByte(TypeHandshakeResp)
	buf[1] = resp.Status
	binary.BigEndian.PutUint64(buf[2:10], resp.SessionID)

	cip := resp.ClientIP.As4()
	copy(buf[10:14], cip[:])

	gip := resp.GatewayIP.As4()
	copy(buf[14:18], gip[:])

	binary.BigEndian.PutUint16(buf[18:20], resp.MTU)
	binary.BigEndian.PutUint64(buf[20:28], resp.NonceEcho)

	mac := ComputeHMAC(psk, buf[:28])
	copy(buf[28:60], mac)
	return buf
}

// DecodeHandshakeResponse validates and parses a 60-byte HandshakeResponse.
func DecodeHandshakeResponse(psk, buf []byte, expectedNonce uint64) (*HandshakeResponse, error) {
	if len(buf) < HandshakeRespLen {
		return nil, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if mtype != TypeHandshakeResp {
		return nil, ErrInvalidType
	}
	// Verify signature first
	if !VerifyHMAC(psk, buf[:28], buf[28:60]) {
		return nil, ErrAuthFailed
	}

	status := buf[1]
	if status == StatusVersionMismatch {
		return &HandshakeResponse{Status: status}, ErrInvalidVersion
	}
	if ver != Version {
		return nil, ErrInvalidVersion
	}

	nonceEcho := binary.BigEndian.Uint64(buf[20:28])
	if expectedNonce != 0 && nonceEcho != expectedNonce {
		return nil, ErrNonceMismatch
	}

	cip, _ := netip.AddrFromSlice(buf[10:14])
	gip, _ := netip.AddrFromSlice(buf[14:18])

	return &HandshakeResponse{
		Status:    status,
		SessionID: binary.BigEndian.Uint64(buf[2:10]),
		ClientIP:  cip,
		GatewayIP: gip,
		MTU:       binary.BigEndian.Uint16(buf[18:20]),
		NonceEcho: nonceEcho,
	}, nil
}

// --- Sealed session messages (Data / Ping / Pong / Disconnect) ---
//
// These message types are encoded and decoded exclusively through
// SessionCrypto (see crypto.go). There is intentionally NO plaintext
// encoding path: accepting one would let an attacker downgrade a session
// by flipping the message-type nibble.

// String provides a human-readable name for a message type.
func TypeString(mtype byte) string {
	switch mtype {
	case TypeHandshakeReq:
		return "HandshakeReq"
	case TypeHandshakeResp:
		return "HandshakeResp"
	case TypeData:
		return "Data"
	case TypePing:
		return "Ping"
	case TypePong:
		return "Pong"
	case TypeDisconnect:
		return "Disconnect"
	case TypeFEC:
		return "FECParity"
	default:
		return fmt.Sprintf("Unknown(0x%x)", mtype)
	}
}

// FECPayload holds structured metadata and XOR parity data for packet recovery.
type FECPayload struct {
	BaseSeq uint64
	Count   uint8
	Lengths []uint16
	Parity  []byte
}

// EncodeFECPayload serializes FEC parity metadata and XOR sum into a binary slice.
func EncodeFECPayload(baseSeq uint64, count uint8, lengths []uint16, parity []byte) []byte {
	hdrLen := 10 + 2*int(count)
	buf := make([]byte, hdrLen+len(parity))
	binary.BigEndian.PutUint64(buf[0:8], baseSeq)
	buf[8] = count
	buf[9] = 0 // reserved
	for i := 0; i < int(count); i++ {
		var l uint16
		if i < len(lengths) {
			l = lengths[i]
		}
		binary.BigEndian.PutUint16(buf[10+i*2:12+i*2], l)
	}
	copy(buf[hdrLen:], parity)
	return buf
}

// DecodeFECPayload deserializes a binary slice into FECPayload.
func DecodeFECPayload(buf []byte) (FECPayload, error) {
	if len(buf) < 10 {
		return FECPayload{}, ErrPacketTooShort
	}
	baseSeq := binary.BigEndian.Uint64(buf[0:8])
	count := buf[8]
	if count == 0 {
		return FECPayload{}, errors.New("fec count cannot be 0")
	}
	hdrLen := 10 + 2*int(count)
	if len(buf) < hdrLen {
		return FECPayload{}, ErrPacketTooShort
	}
	lengths := make([]uint16, count)
	for i := 0; i < int(count); i++ {
		lengths[i] = binary.BigEndian.Uint16(buf[10+i*2 : 12+i*2])
	}
	parity := buf[hdrLen:]
	return FECPayload{
		BaseSeq: baseSeq,
		Count:   count,
		Lengths: lengths,
		Parity:  parity,
	}, nil
}
