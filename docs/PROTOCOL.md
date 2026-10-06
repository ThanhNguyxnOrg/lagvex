# 📡 Lagvex Wire Protocol v1 Specification ⚡

> **Lightweight Binary Wire Protocol for Competitive Gaming**  
> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 🎯 Protocol Principles

- 📦 **Transport**: Standard UDP (one client socket, one server listening port).
- 🌐 **Layer-3 Payload**: Transports whole IPv4 packets. Layer-4 protocols (TCP, UDP, ICMP) are handled by the Linux kernel.
- ⚡ **Minimal Per-Packet Overhead**: Fixed 17-byte secure header + 16-byte Poly1305 tag (33 bytes total overhead) for encrypted data packets.
- 🔒 **Security**: Pre-Shared Key (PSK) authentication via HMAC-SHA256 during handshake. All post-handshake traffic is secured with ChaCha20-Poly1305 AEAD with anti-replay monotonic counters.

---

## 2. 🧩 Header Layout & Message Types

Every packet begins with a 1-byte header:

```text
bits 7..4 : Protocol Version (currently 0x1)
bits 3..0 : Message Type (0x1 .. 0x7)
```

| Type Code | 🏷️ Message Name | ↔️ Direction | 📏 Packet Size |
|---|---|---|---|
| `0x1` | `HandshakeReq` 🤝 | Client -> Relay | 57 bytes |
| `0x2` | `HandshakeResp` 📨 | Relay -> Client | 60 bytes |
| `0x3` | `Data` 🚀 | Bidirectional | 33 bytes + Payload (ChaCha20-Poly1305 AEAD) |
| `0x4` | `Ping` 🏓 | Client -> Relay | 41 bytes (Encrypted timestamp payload) |
| `0x5` | `Pong` 🎾 | Relay -> Client | 41 bytes (Encrypted timestamp payload) |
| `0x6` | `Disconnect` 🛑 | Client -> Relay | 33 bytes (Encrypted notification) |
| `0x7` | `FEC` 🛡️ | Client -> Relay | 33 bytes + Parity Payload |

---

## 3. 🤝 Handshake Protocol

### HandshakeRequest (57 bytes) 📤

Sent by the client to initiate or resume an accelerated session:

```text
Offset  Length  Field
0       1       Header (0x11: Version 1, Type 1)
1       8       Client Nonce (random uint64, big-endian)
9       8       Unix Timestamp in seconds (int64, big-endian)
17      8       ClientID (random uint64 persistent identifier)
25      32      HMAC-SHA256(psk, bytes[0..25))
```

- 🛡️ **Replay Guard**: The relay verifies that `|now - timestamp| <= 120s`. Packets outside this window are dropped.
- 🤫 **Silent Drop**: If HMAC verification fails or clock drift exceeds 120s, the relay drops the packet silently without returning an error (preventing port scanning oracle attacks).
- 🔄 **ClientID Address Reservation**: The relay remembers the inner IP assigned to `ClientID`. If a client briefly drops connection, reconnecting with the same `ClientID` retrieves the identical IP, eliminating the need to tear down or rebuild routing table entries.

### HandshakeResponse (60 bytes) 📥

Returned by the relay upon successful authentication:

```text
Offset  Length  Field
0       1       Header (0x12: Version 1, Type 2)
1       1       Status (0 = OK, 1 = Pool Full, 2 = Server Busy, 3 = Version Mismatch, 4 = Auth Failed)
2       8       SessionID (random uint64 assigned by relay)
10      4       Client Assigned IPv4 (e.g. 10.88.0.2)
14      4       Gateway IPv4 (e.g. 10.88.0.1)
18      2       Recommended Tunnel MTU (uint16 big-endian, default 1400)
20      8       Nonce Echo (must match client's requested Nonce)
28      32      HMAC-SHA256(psk, bytes[0..28))
```

The client verifies the HMAC before trusting any returned fields.

---

## 4. 🚀 Data Packet (33-Byte Secure Overhead + Encrypted IPv4 Payload)

Transports game IP packets between client and relay using ChaCha20-Poly1305 AEAD:

```text
Offset  Length  Field
0       1       Header (0x13: Version 1, Type 3)
1       8       SessionID (uint64, big-endian, plaintext for routing/demux)
9       8       Monotonic Counter (uint64, big-endian nonce)
17      N       Encrypted IPv4 Packet Payload
17+N    16      Poly1305 Authentication Tag (bytes[0..17) used as AAD)
```

### Data Plane Invariants:
1. 🛡️ **Anti-Spoofing**: When receiving decrypted data from a client, the relay extracts the source IPv4 from the inner packet header (`bytes[12..16]`) and drops the packet if it does not match `session.InnerIP`.
2. 🔄 **Dynamic Roaming**: If a client switches from Wi-Fi to Ethernet or cellular, their external UDP address and port change. The relay automatically updates `session.RemoteUDP` upon receiving any valid authenticated packet for that session.
3. 🚫 **Bogon / Private Network Filtering**: The relay drops packets whose inner destination IP belongs to RFC 1918 private subnets, loopback, or cloud instance metadata (`169.254.169.254`).
4. 🛡️ **Replay Protection**: Packets with duplicate or backwards monotonic counters outside the sliding window are rejected before processing.

---

## 5. 🏓 Ping & Pong (41 Bytes)

Used for keepalive, NAT hole punching, and continuous round-trip time (RTT) latency measurement:

```text
Offset  Length  Field
0       1       Header (0x14 for Ping, 0x15 for Pong)
1       8       SessionID (uint64, big-endian)
9       8       Monotonic Counter (uint64, big-endian)
17      8       Encrypted Client Timestamp (uint64 ticks or nanoseconds)
25      16      Poly1305 Authentication Tag
```

- ⏱️ The client transmits a `Ping` every 2 seconds.
- 🔁 The relay echoes the timestamp in an encrypted `Pong` packet.
- 📊 The client measures RTT: `ping_latency_ms = (now - timestamp)`.
- 🧹 If no packet is received for 90 seconds, the relay marks the session as idle and releases its resources.

---

## 6. 🛑 Disconnect (33 Bytes)

Graceful teardown notification sent by the client when shutting down:

```text
Offset  Length  Field
0       1       Header (0x16: Version 1, Type 6)
1       8       SessionID (uint64, big-endian)
9       8       Monotonic Counter (uint64, big-endian)
17      16      Poly1305 Authentication Tag (empty payload)
```

The relay immediately cleans up the session table and frees the IP address in the pool.

---

## 7. 📏 MTU Math & TCP MSS Clamping

```text
Standard Path MTU:               1500 bytes
- Outer IPv4 Header:             - 20 bytes
- UDP Header:                    -  8 bytes
- Lagvex AEAD Overhead:          - 33 bytes (17B header + 16B Poly1305 tag)
-------------------------------------------
Maximum Safe Tunnel MTU:         1439 bytes
Recommended Safe Default MTU:    1400 bytes
```

Configuring an MTU of **1400** leaves ample headroom for PPPoE connections (1492 bytes) and ISP encapsulation protocols without causing packet fragmentation.

TCP connections inside the tunnel are clamped to PMTU on the relay via `iptables`:
```bash
iptables -t mangle -I FORWARD 1 -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu
```

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
