# Lagvex Wire Protocol v1

This document specifies the wire protocol between the **Lagvex Windows Client** and the **Lagvex Linux Relay**.

---

## 1. Principles

- **Transport**: Standard UDP (one client socket, one server listening port).
- **Payload**: Whole IPv4 packets (Layer-3). Layer-4 headers (TCP, UDP, ICMP) are handled by the Linux kernel.
- **Minimal Per-Packet Overhead**: Fixed 9-byte header for data packets. No Type-Length-Value (TLV) parsing in the data plane.
- **Authentication**: Pre-Shared Key (PSK) authentication using HMAC-SHA256 during handshake. Data packets are authenticated by session ID and source IP verification for minimum latency.

---

## 2. Packet Layout & Message Types

Every packet begins with a 1-byte header:

```
bits 7..4 : Protocol Version (currently 0x1)
bits 3..0 : Message Type (0x1 .. 0x6)
```

| Type Code | Name | Direction | Packet Length |
|---|---|---|---|
| `0x1` | `HandshakeReq` | Client -> Relay | 57 bytes |
| `0x2` | `HandshakeResp` | Relay -> Client | 60 bytes |
| `0x3` | `Data` | Bidirectional | 9 bytes + Payload |
| `0x4` | `Ping` | Client -> Relay | 17 bytes |
| `0x5` | `Pong` | Relay -> Client | 17 bytes |
| `0x6` | `Disconnect` | Client -> Relay | 9 bytes |

---

## 3. Handshake Messages

### HandshakeRequest (57 bytes)

Sent by the client to initiate or resume an accelerated session:

```
Offset  Length  Field
0       1       Header (0x11: Version 1, Type 1)
1       8       Client Nonce (random uint64, big-endian)
9       8       Unix Timestamp in seconds (int64, big-endian)
17      8       ClientID (random uint64 persistent identifier)
25      32      HMAC-SHA256(psk, bytes[0..25))
```

- **Replay Protection**: The relay verifies that `|now - timestamp| <= 120s`. Packets outside this window are dropped.
- **Silent Drop**: If HMAC verification fails or clock drift exceeds 120s, the relay drops the packet silently without returning an error (preventing port scanning oracle attacks).
- **ClientID Address Reservation**: The relay remembers the inner IP assigned to `ClientID`. If a client briefly drops connection, reconnecting with the same `ClientID` retrieves the identical IP, eliminating the need to tear down or rebuild Windows routing table entries.

### HandshakeResponse (60 bytes)

Returned by the relay upon successful authentication:

```
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

## 4. Data Packet (9-Byte Header + Raw IPv4 Payload)

Transports game IP packets between client and relay:

```
Offset  Length  Field
0       1       Header (0x13: Version 1, Type 3)
1       8       SessionID (uint64, big-endian)
9       N       Raw IPv4 Packet (starts with version nibble 0x4)
```

### Data Plane Invariants:
1. **Anti-Spoofing**: When receiving data from a client, the relay extracts the source IPv4 from the inner packet header (`bytes[12..16]`) and drops the packet if it does not match `session.InnerIP`.
2. **Dynamic Roaming**: If a client switches from Wi-Fi to Ethernet or cellular, their external UDP address and port change. The relay automatically updates `session.RemoteUDP` upon receiving any valid data packet for that session.
3. **Bogon / Private Network Filtering**: The relay drops packets whose inner destination IP belongs to RFC 1918 private subnets, loopback, or cloud instance metadata (`169.254.169.254`).

---

## 5. Ping & Pong (17 Bytes)

Used for keepalive, NAT hole punching, and continuous round-trip time (RTT) latency measurement:

```
Offset  Length  Field
0       1       Header (0x14 for Ping, 0x15 for Pong)
1       8       SessionID (uint64, big-endian)
9       8       Client Timestamp (uint64 ticks or nanoseconds)
```

- The client transmits a `Ping` every 2 seconds.
- The relay echoes the timestamp in a `Pong` packet without modification.
- The client measures RTT: `ping_latency_ms = (now - timestamp)`.
- If no packet is received for 90 seconds, the relay marks the session as idle and releases its resources.

---

## 6. Disconnect (9 Bytes)

Graceful teardown notification sent by the client when shutting down:

```
Offset  Length  Field
0       1       Header (0x16: Version 1, Type 6)
1       8       SessionID (uint64, big-endian)
```

The relay immediately cleans up the session table and frees the IP address in the pool.

---

## 7. MTU Calculations & MSS Clamping

```
Standard Path MTU:               1500 bytes
- Outer IPv4 Header:             - 20 bytes
- UDP Header:                    -  8 bytes
- Lagvex Data Header:            -  9 bytes
-------------------------------------------
Maximum Safe Tunnel MTU:         1463 bytes
Recommended Safe Default MTU:    1400 bytes
```

Configuring an MTU of **1400** leaves ample headroom for PPPoE connections (1492 bytes) and ISP encapsulation protocols without causing packet fragmentation.

TCP connections carried inside the tunnel are clamped to PMTU on the relay via `iptables`:
```bash
iptables -t mangle -I FORWARD 1 -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu
```
