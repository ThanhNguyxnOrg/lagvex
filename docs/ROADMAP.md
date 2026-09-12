# 🗺️ Lagvex Engineering Roadmap & Production Plan 🚀

> 💡 **Related Documents**: [📖 Back to README](../README.md) • [🔬 Networking Research](RESEARCH.md) • [📐 Architecture](ARCHITECTURE.md) • [📡 Wire Protocol](PROTOCOL.md) • [⚖️ Disclaimer](DISCLAIMER.md)

---

## 🎯 Strategic Overview 📌

This document outlines the phased engineering roadmap to evolve **Lagvex** from a high-performance Minimum Viable Product (MVP) into an enterprise-grade, cryptographically sound, production-ready game latency reducer.

Our roadmap strictly adheres to the principle of **Evidence-Based Systems Engineering**:
> *"Never add complexity before establishing a measurement baseline. Do not jump to multipath or AI schedulers before solving data-plane cryptographic integrity and OS-level routing robustness."*

```text
 ┌───────────────────────────┐
 │   P0: Security & Crypto   │ ➔ ChaCha20-Poly1305 AEAD, Anti-Replay Nonce Cache, Authenticated Roaming
 └─────────────┬─────────────┘
               ▼
 ┌───────────────────────────┐
 │   P1: Native OS Routing   │ ➔ Win32 IP Helper APIs (no CLI scraping), Session IP Pool pre-cleanup
 └─────────────┬─────────────┘
               ▼
 ┌───────────────────────────┐
 │   P2: E2E Telemetry & HUD │ ➔ Direct vs. Relay comparison, p50/p95/p99 tail latency, Jitter math
 └─────────────┬─────────────┘
               ▼
 ┌───────────────────────────┐
 │   P3: Multi-Relay Engine  │ ➔ Multi-node scoring, Hysteresis cooldown, Automated failover
 └─────────────┬─────────────┘
               ▼
 ┌───────────────────────────┐
 │   P4: Loss Recovery & FEC │ ➔ Adaptive Forward Error Correction, Multi-interface WAN bonding
 └───────────────────────────┘
```

---

## 🚨 Phase P0 — Critical Security & Cryptographic Integrity 🔐

*Target: Immediate Execution. Focus: Eliminating protocol vulnerabilities and securing data transmission.*

### 1. 🛡️ Data-Plane Authenticated Encryption (AEAD)
- **Vulnerability**: While Handshakes are authenticated via HMAC-SHA256, current data packets encapsulate raw IPv4 payloads with only a plaintext `SessionID` and sequence number.
- **Action**: Implement **ChaCha20-Poly1305 AEAD** (`golang.org/x/crypto/chacha20poly1305`) as standardized in WireGuard.
- **Mechanism**:
  - Derive symmetric session keys during the HMAC-SHA256 handshake.
  - Encrypt and authenticate each game datagram with an incremental 64-bit sequence counter as the nonce.
  - Append a 16-byte Poly1305 authentication tag to guarantee both **confidentiality** and **tamper-proof integrity**.

### 2. ⏳ True Anti-Replay Mechanism with Nonce Caching
- **Vulnerability**: A simple timestamp window ($\pm 120$s) allows duplicate handshakes to be replayed within the 2-minute window, creating orphaned sessions.
- **Action**:
  - Deploy a server-side **Sliding-Window Nonce Cache** (`sync.Map` with a 120-second TTL pruner).
  - Reject any handshake whose nonce has already been observed within the validity window.
  - Apply RFC 6479 128-bit sliding-window packet replay checks to data-plane sequence numbers.

### 3. 🌐 Cryptographically Authenticated Dynamic Roaming
- **Vulnerability**: Currently, incoming packets with a valid `SessionID` immediately update `sess.RemoteUDP`, allowing session hijacking if a 4-byte session ID is guessed or sniffed.
- **Action**:
  - Update `sess.RemoteUDP` **strictly AFTER** `aead.Open()` successfully authenticates the Poly1305 tag.
  - Ensure seamless Wi-Fi $\leftrightarrow$ Ethernet $\leftrightarrow$ 4G handoffs remain 100% immune to spoofing.

### 4. 📝 Claim Alignment & Documentation Honesty
- **Action**: Remove absolute guarantees such as "100% Anti-Cheat Safe" and "Sub-ms GC guarantees zero jitter".
- **Status**: Completed in v1.0.0; permanently enforced across all future pull requests.

---

## ⚙️ Phase P1 — Native OS Routing & State Invariants 🪟

*Target: System Reliability. Focus: Replacing CLI text parsing with native OS APIs and strengthening state lifecycles.*

### 1. 🔌 Windows Native IP Helper API Integration
- **Problem**: Scraping output from `route print` and `netsh` fails on non-English Windows localizations (Vietnamese, Japanese, German) or systems with complex Hyper-V / WSL virtual switches.
- **Action**: Refactor [routes_windows.go](file:///d:/Code/Lagvex/pkg/client/routes_windows.go) to bind directly to Windows `iphlpapi.dll` via `golang.org/x/sys/windows`:
  - `GetBestRoute2()`: Queries the true physical default gateway route.
  - `GetAdaptersAddresses()`: Retrieves interface metrics and GUIDs without string parsing.
  - `CreateIpForwardEntry2()`: Installs ephemeral `store=active` RAM routes natively.

### 2. 🔄 Atomic Reconnect & Session Pre-Cleanup
- **Problem**: When a client reconnects on an existing `ClientID`, the relay currently checks `MaxClients` and calls `Allocate()` *before* cleaning up the previous session, risking IP exhaustion when the pool is near capacity.
- **Action**: Atomically look up and terminate old sessions for the same `ClientID` *prior* to requesting a new IP reservation.

### 3. 👁️ Process Lifecycle Binding (PID Tracking)
- **Action**: Upgrade [watcher.go](file:///d:/Code/Lagvex/pkg/client/watcher.go) to track active Process IDs (PIDs) using Windows Toolhelp32 snapshots and Linux `/proc` instead of repetitive `tasklist /FO CSV` polling.

---

## 📊 Phase P2 — True E2E Telemetry & Tail Latency HUD 📈

*Target: Observability & User Trust. Focus: Measuring what actually matters for competitive gameplay.*

### 1. 🎯 Direct vs. Relay Real-Time Comparison (Smart Route Advisor)
- **Feature**: Simultaneously probe:
  1. $RTT_{\text{direct}}$: Ping from PC directly to the target game cloud gateway.
  2. $RTT_{\text{relay}}$: Ping from PC to Relay + Relay to Game Server.
- **User Experience**: If the direct path is already optimal ($RTT_{\text{direct}} < RTT_{\text{relay}}$ and 0% loss), the dashboard explicitly advises:
  > *"⚡ Your direct ISP path is currently optimal (26ms vs 34ms). Booster recommended: OFF."*

### 2. 📉 Tail Latency (P50, P95, P99) & Jitter Mathematics
- **Problem**: Mean/average ping hides the micro-freezes and packet spikes that ruin competitive gunfights.
- **Action**: Maintain rolling 120-packet histogram windows in memory to compute:
  - **P50 (Median)**: Typical round-trip time.
  - **P95 / P99**: Tail latency spikes during ISP congestion.
  - **Jitter (Mean Deviation)**: Stability metric displayed live on the Cyberpunk HUD.

---

## 🌐 Phase P3 — Multi-Relay Schedulers & Hysteresis 🛰️

*Target: Routing Intelligence. Focus: Multi-node path selection without route flapping.*

### 1. 🏆 Multi-Relay Path Scoring
- Support concurrent ping probes across multiple deployed relays (e.g., Singapore, Tokyo, Hong Kong).
- Compute composite health score:
  $$\text{Score} = w_{\text{rtt}} \cdot \text{RTT}_{\text{median}} + w_{\text{p95}} \cdot \text{RTT}_{\text{p95}} + w_{\text{loss}} \cdot \text{Loss} + w_{\text{jitter}} \cdot \text{Jitter}$$

### 2. 🎛️ Hysteresis Cooldown (Anti-Flapping)
- Require an alternate relay to show a sustained $\ge 15\%$ improvement for at least 5 seconds before initiating a dynamic route handover.
- Prevent rapid route oscillation during borderline latency periods.

---

## 🚀 Phase P4 — Loss Recovery & Multipath Bonding 🧪

*Target: Extreme Network Resilience. Focus: Forward Error Correction and dual-link support.*

### 1. 🛡️ Adaptive Forward Error Correction (FEC)
- Inspired by `UDPspeeder` mechanics: when packet loss on undersea routes exceeds $1.5\%$, dynamically inject XOR or Reed-Solomon parity packets.
- Enables the receiving end to instantly reconstruct dropped packets without waiting for retransmission.

### 2. 📱 Multipath Failover (Ethernet + Cellular Backup)
- Aggregate secondary connections (e.g., Wi-Fi + 4G USB tethering) using packet duplication or fast failover, ensuring that a home internet outage never disconnects an active ranked match.

---

## 📅 Roadmap Execution Matrix 📋

| Phase | Milestone | Primary Deliverable | Target Impact |
|:---:|:---|:---|:---|
| **P0** 🚨 | Cryptographic Hardening | ChaCha20-Poly1305 AEAD + Nonce Replay Guard | Enterprise-grade security, zero data tampering |
| **P1** ⚙️ | OS Robustness | Win32 IP Helper APIs + Session Cleanup Invariants | Rock-solid routing on all Windows language editions |
| **P2** 📊 | Deep Telemetry | P95/P99 latency histogram + Direct vs. Relay advisor | Transparent, trustworthy gamer HUD |
| **P3** 🌐 | Multi-Relay Engine | Multi-node scoring + Hysteresis anti-flapping | Automated optimal path selection across Asia/EU |
| **P4** 🧪 | Loss Resilience | Adaptive FEC + Dual-link WAN failover | 0% packet loss even during undersea cable cuts |

---

[⬅️ Back to README](../README.md) • [🔬 Read Research Whitepaper](RESEARCH.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)
