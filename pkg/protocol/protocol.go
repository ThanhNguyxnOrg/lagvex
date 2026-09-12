// Package protocol implements the Lagvex binary wire protocol over UDP.
// It is designed from scratch for minimum packet overhead, high throughput,
// and secure authentication via Pre-Shared Key (HMAC-SHA256).
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

	// Packet lengths in bytes.
	HandshakeReqLen  = 57
	HandshakeRespLen = 60
	DataHeaderLen    = 9
	PingLen          = 17
	PongLen          = 17
	DisconnectLen    = 9

	// MaxPacketSize is the maximum MTU plus protocol header.
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

// --- Data Packet ---

// EncodeDataPacket wraps an IPv4 packet with a 9-byte Lagvex data header.
// buf must have at least len(payload) + DataHeaderLen capacity.
func EncodeDataPacket(buf []byte, sessionID uint64, payload []byte) []byte {
	out := buf[:DataHeaderLen+len(payload)]
	out[0] = HeaderByte(TypeData)
	binary.BigEndian.PutUint64(out[1:9], sessionID)
	copy(out[9:], payload)
	return out
}

// DecodeDataPacket parses a Data packet, extracting the session ID and IPv4 payload.
func DecodeDataPacket(buf []byte) (sessionID uint64, payload []byte, err error) {
	if len(buf) < DataHeaderLen+20 { // 9 bytes header + min 20 bytes IPv4 header
		return 0, nil, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, nil, ErrInvalidVersion
	}
	if mtype != TypeData {
		return 0, nil, ErrInvalidType
	}
	// Fast IPv4 sanity check: first nibble of IP payload must be 4
	if (buf[9] >> 4) != 4 {
		return 0, nil, ErrNotIPv4
	}
	sessionID = binary.BigEndian.Uint64(buf[1:9])
	return sessionID, buf[9:], nil
}

// --- Ping & Pong ---

// EncodePing creates a 17-byte Ping message.
func EncodePing(sessionID uint64, timestamp uint64) []byte {
	buf := make([]byte, PingLen)
	buf[0] = HeaderByte(TypePing)
	binary.BigEndian.PutUint64(buf[1:9], sessionID)
	binary.BigEndian.PutUint64(buf[9:17], timestamp)
	return buf
}

// DecodePing parses a Ping message.
func DecodePing(buf []byte) (sessionID uint64, timestamp uint64, err error) {
	if len(buf) < PingLen {
		return 0, 0, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, 0, ErrInvalidVersion
	}
	if mtype != TypePing {
		return 0, 0, ErrInvalidType
	}
	return binary.BigEndian.Uint64(buf[1:9]), binary.BigEndian.Uint64(buf[9:17]), nil
}

// EncodePong creates a 17-byte Pong message.
func EncodePong(sessionID uint64, timestamp uint64) []byte {
	buf := make([]byte, PongLen)
	buf[0] = HeaderByte(TypePong)
	binary.BigEndian.PutUint64(buf[1:9], sessionID)
	binary.BigEndian.PutUint64(buf[9:17], timestamp)
	return buf
}

// DecodePong parses a Pong message.
func DecodePong(buf []byte) (sessionID uint64, timestamp uint64, err error) {
	if len(buf) < PongLen {
		return 0, 0, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, 0, ErrInvalidVersion
	}
	if mtype != TypePong {
		return 0, 0, ErrInvalidType
	}
	return binary.BigEndian.Uint64(buf[1:9]), binary.BigEndian.Uint64(buf[9:17]), nil
}

// --- Disconnect ---

// EncodeDisconnect creates a 9-byte Disconnect message.
func EncodeDisconnect(sessionID uint64) []byte {
	buf := make([]byte, DisconnectLen)
	buf[0] = HeaderByte(TypeDisconnect)
	binary.BigEndian.PutUint64(buf[1:9], sessionID)
	return buf
}

// DecodeDisconnect parses a Disconnect message.
func DecodeDisconnect(buf []byte) (sessionID uint64, err error) {
	if len(buf) < DisconnectLen {
		return 0, ErrPacketTooShort
	}
	ver, mtype := ParseHeader(buf[0])
	if ver != Version {
		return 0, ErrInvalidVersion
	}
	if mtype != TypeDisconnect {
		return 0, ErrInvalidType
	}
	return binary.BigEndian.Uint64(buf[1:9]), nil
}

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
	default:
		return fmt.Sprintf("Unknown(0x%x)", mtype)
	}
}
