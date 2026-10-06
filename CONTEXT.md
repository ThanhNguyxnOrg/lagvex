# Lagvex Domain Model & Glossary (CONTEXT.md)

This document establishes the ubiquitous language and architectural domain model for **Lagvex**, a high-performance cross-platform gaming VPN split-tunneling booster.

---

## 1. Core Ubiquitous Language

| Term | Definition |
|---|---|
| **Lagvex Client** | The user-space endpoint running on the gamer's computer (`lagvex-client`). Handles adapter orchestration, route pinning, process detection, AEAD encryption, FEC parity injection, and UI telemetry. |
| **Lagvex Relay** | The remote high-speed edge server running on Linux (`lagvex-relay`). Decapsulates UDP datagrams, applies anti-spoofing validation, NAT MASQUERADE, and forwards packets directly to game datacenters. |
| **Virtual Adapter** | The kernel-level virtual network interface on the client machine: **WinTun** on Windows, **utun** on macOS, and **/dev/net/tun** on Linux. |
| **Split-Tunneling** | Selective routing where only target game datacenter CIDRs are directed across the accelerated tunnel, leaving web browsers, Discord, and streaming on the physical ISP connection. |
| **Route Pinning** | Installing a specific `/32` host route for the VPS relay through the physical network gateway to prevent routing loops. |
| **Session Crypto** | ChaCha20-Poly1305 Authenticated Encryption with Associated Data (AEAD) with monotonic 64-bit anti-replay counter, negotiated via HMAC-SHA256 handshake. |
| **Systematic FEC** | Forward Error Correction using adaptive XOR parity blocks (e.g., 8:1 ratio) to recover dropped game packets with zero round-trip latency overhead. |
| **Packet Hedging** | Selective duplication of small critical game input packets ($\le 256$ bytes) during network degradation periods to eliminate jitter. |
| **Hysteresis Failover** | Intelligent relay switching requiring a minimum percentage score improvement (e.g. 15%) and cooldown timer to prevent route flapping. |
| **L4 Smart Filter** | Layer-4 transport inspection that routes UDP gaming loop packets through the tunnel while letting bulky TCP (HTTP/HTTPS) patch downloads bypass the tunnel. |
| **In-Game Mini HUD** | An optional, draggable glassmorphism on-screen widget displaying real-time ping, packet loss, and relay status. |

---

## 2. Invariants & Boundaries

1. **Anti-Spoofing Invariant**: The Relay strictly drops packets whose inner source IPv4 does not match the authenticated session's inner IP.
2. **Replay Invariant**: All post-handshake AEAD packets must possess a monotonically advancing sequence counter within the sliding replay window.
3. **Clean Teardown Invariant**: All installed kernel routes must be unpinned and deleted prior to closing the virtual adapter, avoiding orphan routes or connection loss.
4. **Zero-Lock Data Plane**: High-throughput packet pumps must never allocate heap memory per packet or acquire global mutex locks on the hot path.
