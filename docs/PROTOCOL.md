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
bits 3..0 : Message Type (0x1 .. 0x9)
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
| `0x8` | `AnchorProbe` 🎯 | Client -> Relay | 41 bytes (Encrypted sequence + anchor id) |
| `0x9` | `AnchorProbeReply` 🎯 | Relay -> Client | 41 bytes (Encrypted sequence + rtt + loss) |
| `0xA` | `DualAck` ♻️ | Relay -> Client | 41 bytes (Encrypted highest forwarded dual seq) |

Old relays drop the `0x8`/`0x9` types, so anchor probing degrades to "not
measured" instead of breaking the tunnel.

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

## 7. 🎯 Anchor Probe (41 Bytes)

The client can ask the connected relay to measure **its own** RTT to one public
cloud anchor so the route detail view can show a measured via-relay estimate.

- `AnchorProbe` payload (8 bytes, sealed): `sequence (u32) | anchor id (u8)`.
- `AnchorProbeReply` payload (8 bytes, sealed): `sequence (u32) | rtt us (u24) | loss % (u8)`.

Safety rules that keep a relay from becoming a scanner or amplifier:

- The relay only accepts ids from the fixed `protocol.AnchorProbeTargets`
  allowlist; unknown ids are rejected and logged. A client can never request an
  arbitrary host or port.
- Each session is rate limited to one probe every 2 seconds and 60 per hour.
- A probe is only two TCP handshakes with a 1.5s timeout each; failures are
  reported as loss, never as invented latency.

---

## 8. 📏 MTU Math & TCP MSS Clamping

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

## 9. ♻️ Dual-Path Dedup (Redundant-lite, Phase 1)

Phase 1 makes relays ready for an opt-in dual-send client (Phase 2): the same
data frame is transmitted to two relays and the first copy seen is forwarded,
while later copies are dropped. **No shipped client sends dual frames yet** —
until Phase 2, every session uses a single path and the dedup state stays empty.

### 9.1 Handshake extension (request)

A dual-capable client appends 17 bytes after the 57-byte handshake request:

| Offset (after byte 57) | Size | Field |
|------------------------|------|-------|
| +0 | 1 | Extension type `0x01` (dual path) |
| +1 | 16 | Connection group ID (random, client-chosen) |

The base 57 bytes and their PSK HMAC are unchanged, so **old relays still
authenticate the handshake** and simply ignore the trailing bytes. A relay that
receives unknown extension bytes fails the handshake closed, which makes a
dual-capable client fall back to a single path on that relay.

### 9.2 Capability echo (response)

A dual-aware relay that accepted a group appends 4 bytes to the 60-byte
handshake response: extension type `0x01` plus three reserved zero bytes. The
base response and HMAC are unchanged, so **old clients ignore it**. A client
must treat a response without the extension as "this relay is single-path only"
and never duplicate frames to it.

### 9.3 Dedup scope

The relay derives the dedup scope as `SHA-256(PSK || group ID)` truncated to
4 bytes. Because the hash is keyed by the relay's PSK, only PSK holders can
join a group: an attacker who can compute a victim's group hash can already
authenticate to the relay. The raw group ID is never used directly and is not
covered by the request HMAC — it does not need to be.

### 9.4 Data frame trailer

A dual session's data frames carry an 8-byte trailer after the inner IPv4
packet:

| Offset (after inner packet) | Size | Field |
|-----------------------------|------|-------|
| +0 | 4 | Frame sequence (u32, shared by both relays) |
| +4 | 4 | Group hash (u32, §9.3) |

The trailer must be identical for both relays. The relay strips it before FEC
processing and TUN delivery, and counts only the inner packet's bytes as tunnel
uplink. A dual session that sends a frame without a valid trailer matching its
group is dropped and counted as a bad frame.

### 9.5 Relay dedup rules

- First copy of `(group hash, sequence)` inside a 2-second TTL is forwarded;
  later copies are counted as suppressed and dropped.
- Per-group window is capped at 4096 sequences; expired entries are evicted
  first, then the oldest.
- Groups expire after 10 minutes without observations; at most 256 groups are
  tracked. The reaper sweeps both every 10 seconds.
- Unregistered groups are never deduped: only a PSK-authenticated handshake
  creates group state.

The relay `/health` endpoint reports `dualSessions`, `dualGroups`,
`dualForwarded` and `dualSuppressed` so operators can verify the counters stay
at zero until Phase 2 exists.

### 9.6 DualAck progress report (Phase 2 groundwork)

For dual sessions the relay sends a `DualAck` (type `0xA`, sealed like Pong)
after forwarding dual frames. The 8-byte payload carries the **highest
contiguous dual frame sequence** the relay has forwarded to the game server;
four bytes are reserved. ACKs are rate-limited to at most one per 20 ms per
session (50/s), and a session that has never forwarded a dual frame never
receives one. Contiguity matters: a highest-seq ACK with gaps below it would
let the client purge frames it still needs to retransmit, so the watermark
only advances across sequences the relay has actually seen — a suppressed
duplicate (a client retransmission of an earlier copy) counts too and can
fill a hole. The out-of-order tracking set behind the watermark is capped at
512 entries; evicting the oldest hole stalls the watermark at that sequence
until the client's retransmission fills it, so the state self-heals.

The client (Phase 2) keeps a small ring of recent dual frames, purges it on
ACK, and retransmits only frames that stay unacked beyond a measured
threshold — the primary relay first (Phase 1 dedup makes that idempotent),
then the standby relay. This mirrors ExitLag's confirm/`detect_lost_packets`
behaviour and keeps the common case free of duplicate packets at the game
server: blind duplicate sends would make both relays forward their own first
copy, since dedup state is per relay and relays do not share state.

### 9.7 Client behaviour (Phase 2, shipped)

- The user opts in from the dashboard; enabling while connected runs a real
  standby handshake, so the toggle's state reflects an actual session.
- Data frames are sent **only to the primary relay**, each carrying the dual
  trailer. FEC parity, Ping/Pong and anchor probes stay primary-only.
- The primary's contiguous `DualAck` watermarks purge the client ring.
- A frame unacknowledged past `max(40 ms, 1.5× measured RTT)` (clamped to
  20–250 ms) is retransmitted with backoff: attempt 1 to the primary relay,
  attempt 2 to the standby relay, at most twice per frame. Ring entries
  expire after 1 second and the ring holds at most 256 frames.
- The standby session receives keepalive Pings; 30 s without a pong, a
  standby Disconnect, or a relay handover tears the dual session down and
  the client records the reason in `/api/status` (`dualPath.reason`).
- Both relays must share one PSK: the dedup scope is
  `SHA-256(PSK ‖ group ID)`, so different PSKs would give each relay a
  different group identity and the single trailer could not satisfy both.
- `/api/status` reports the measured counters (`framed`, `retransmits`,
  `ackPurged`, `expired`, `ringOverflow`, `inFlight`, `ackedSeq`,
  `thresholdMs`) — overhead is what was counted, never a fixed multiplier.

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
