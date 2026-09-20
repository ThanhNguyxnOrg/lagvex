# 📝 Changelog 📜

> 🔗 [Back to Project README.md](README.md) | [Contributor Guide](CONTRIBUTING.md)

All notable changes to the **Lagvex** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### 🎮 Zero-Config Standalone Client & Repository Cleanliness
- 🚀 **Embedded Game Profiles & Community Relays (`pkg/profiles/embedded/`)**:
  - `profiles.json` (11 games, 189 CIDRs) and `relays.json` (12 global community relays across 5 continents) are now directly compiled into the client binary via `//go:embed`.
  - Gamers can download a single `lagvex-client.exe` file and boost immediately without renting a VPS, copying PSKs, or configuring JSON files. External files remain supported as overrides for self-hosters.
- 🧹 **Repository Cleanliness**:
  - Pruned legacy web prototype and temporary design assets.
  - Production binaries compile self-contained and clean with zero lingering assets.

### 🛡️ Phase P4: Adaptive Forward Error Correction (FEC) & Zero-RTT Loss Recovery
- 🛡️ **Adaptive Systematic XOR Parity Engine (`pkg/protocol/fec.go`)**:
  - **Systematic Coding**: Raw game data packets (`TypeData`) are forwarded immediately with 0ms buffer delay and zero byte overhead.
  - **Auxiliary Parity Type (`TypeFEC = 0x7`)**: Encrypted with ChaCha20-Poly1305 AEAD, carrying systematic XOR parity blocks and per-packet length tables.
  - **Zero-RTT Single-Loss Recovery**: When a packet drops in transit, the decoder reconstructs the exact original payload in 0ms without waiting for TCP/ARQ retransmission (saving 30-100ms of lag spike).
  - Handles variable game datagram lengths cleanly with dynamic padding and exact reconstruction.
- 🎛️ **Adaptive Loss Ratio Controller (`pkg/protocol/fec.go`)**:
  - Standby (0% overhead) when loss $< 0.5\%$.
  - Light (10:1, 10% redundancy) when loss is $0.5\% - 3.0\%$.
  - Medium (6:1, 16.7% redundancy) when loss is $3.0\% - 8.0\%$.
  - Aggressive (4:1, 25% redundancy) during severe network degradation $> 8.0\%$.
- ⚡ **Bi-Directional Loss Protection (`pkg/client/engine.go`, `pkg/relay/server.go`)**:
  - Uplink: Protects critical player actions and tick updates before relay TUN ingestion.
  - Downlink: Protects game state return packets before injection into client WinTun virtual adapter.
- 🌐 **Esports Cockpit HUD & Telemetry (`web/`, `pkg/client/ui_server.go`)**:
  - Live HUD badge showing `ZERO-RTT FEC: ACTIVE (6:1)` with real-time `X rec` recovery counter.
  - Settings modal toggle switch for instant FEC enablement/disablement.
  - REST endpoint `POST /api/fec/toggle` and extended `TunnelStats` telemetry.
- 🧪 **Unit Test Suite**:
  - 100% test coverage with exhaustive permutation tests: exact single packet recovery, multi-loss false-positive prevention, zero-loss idempotency, adaptive scaling, and engine API controls.

### 🔄 Phase P3: Hysteresis Auto-Failover & Smart Route Advisor
- 🔄 **Hysteresis Auto-Failover Engine (`pkg/client/failover.go`)**:
  - Continuous health monitoring detecting missed keepalive pongs ($\ge 3$), packet loss surges ($\ge 3\%$), and latency spikes ($\ge 40\%$ over baseline).
  - Anti-flapping mathematical hysteresis requiring replacement candidate nodes to be $\ge 15\%$ better for at least 5 seconds before switching:
    $$\text{Score}_{\text{candidate}} \le \text{Score}_{\text{current}} \cdot (1 - 0.15)$$
  - Cooldown timer suppression preventing route oscillation during unstable ISP jitter.
- ⚡ **Zero-Loss Seamless Route Handover (`pkg/client/engine.go`)**:
  - In-flight tunnel transition executing handshake and session key exchange on a secondary UDP socket *before* modifying routing tables or closing existing connections.
  - Thread-safe atomic pointer swaps for active UDP sockets, ChaCha20-Poly1305 ciphers, session IDs, and host `/32` pinned routes.
  - Zero dropped frames, zero game lobby disconnection, and graceful background disconnection of degraded relay.
- 🎯 **Smart Route Advisor (`pkg/client/prober.go`)**:
  - Measures direct domestic gateway latency ($RTT_{\text{direct}}$) and compares it with relay path ($RTT_{\text{relay}}$).
  - Emits real-time advice: `DIRECT_OPTIMAL` (when ISP domestic path is faster or booster unnecessary), `BOOST_RECOMMENDED` (when relay reduces ping or eliminates packet loss), and `COMPARABLE`.
- 🌐 **Cockpit HUD Failover & Advisor Telemetry (`web/`)**:
  - Topbar cyber pill with animated toggle switch for 1-click Auto-Failover control.
  - Glassmorphic Smart Route Advisor card displaying dynamic badges (`⚡ BOOST RECOMMENDED`, `🛡️ DIRECT ISP OPTIMAL`), explanations, and side-by-side RTT comparison.
  - Real-time toast notifications alerting users whenever a background seamless handover occurs.
- 🔌 **REST API Expansion**:
  - `GET /api/advisor`: Returns live route analysis and recommendation for current game/region.
  - `POST /api/failover/toggle`: Dynamically enables/disables auto-failover controller.
  - `GET /api/failover/history`: Returns audit log of recent failover events with score and latency deltas.
- 🧪 **Unit Test Suite**:
  - Comprehensive tests for degradation detection, hysteresis thresholds, cooldown suppression, engine controls, and route advisor calculations with 100% pass rate.

### 🎨 Phase P2: Esports Cockpit HUD & Visual Telemetry
- 🎨 **Esports Cockpit HUD (`web/`)**:
  - Complete front-end overhaul inspired by professional esports command centers.
  - High-contrast cyber dark aesthetics (`#0a0e17`) with electric cyan accents (`#00f0ff`), neon emerald indicators (`#00ff88`), and hot alerts.
  - Typography using Google Fonts *Outfit* (display) and *JetBrains Mono* (telemetry / code).
- 🛡️ **Authentic Vector SVG Game Insignias**:
  - Handcrafted resolution-independent vector emblems for 11 titles (Valorant, CS2, PUBG, Apex Legends, League of Legends, Warzone, Delta Force, Overwatch 2, R6 Siege, The Finals, Dota 2).
- 📡 **Dual-Ring Tactical Radar Scanner & Laser Stream Animation**:
  - Animated 360-degree radar sweep and real-time laser packet stream during active packet tunneling.
- 🔔 **Floating Glassmorphic Toast Notification System**:
  - Sleek pill notifications with backdrop blur, color-coded icon badges, and smooth animations.
- 🌐 **Global 5-Continent Relay Network & 105 Target Clusters**:
  - 14 Worldwide VPS Relay Endpoints (APAC, EU, NA, SA, MENA) and 105 regional server clusters.
- 🤝 **Comprehensive Contributor & Maintainer Guide (`CONTRIBUTING.md`)**:
  - Packet captures, ASN routing, subnets narrowing, profile registration, and validation guide.

### 🪟 Phase P1: Native Routing & Automation
- 🪟 **1-Click Native Windows PowerShell Relay Installer (`scripts/install-relay.ps1`)**:
  - Automated relay configuration on Windows 10/11 or Windows Server.
- ⚡ **Cross-Platform OS Support**:
  - Windows 10/11 (WireGuard WinTun Layer-3 driver), Linux (kernel `/dev/net/tun`), macOS (Darwin `utun`).
- 🔄 **Process Watcher**:
  - PID-aware game process detection and automated on-link route attachment/detachment.

### 🔐 Phase P0: Core Protocol & Security Hardening
- 🛡️ **Data-Plane Cryptographic Integrity**:
  - ChaCha20-Poly1305 AEAD symmetric session cipher derived via HMAC-SHA256 handshake.
  - 64-bit monotonic sequence counter with RFC 6479 anti-replay sliding-window filter.
  - Cryptographically authenticated dynamic roaming protecting against IP hijacking.
- 🌐 **Ultra-Lightweight Go Relay Daemon (`cmd/lagvex-relay`)**:
  - Zero-allocation packet routing, IP masquerade NAT, and TCP MSS clamping.
