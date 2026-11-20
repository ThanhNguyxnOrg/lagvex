# 📐 Lagvex Architecture & Network Design ⚡

> **High-Performance Route-Based Split-Tunneling Engine**  
> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 🎯 Core Principles & Design Goals

Lagvex was built from the ground up to solve the latency challenges of online gaming without risking anti-cheat penalties:

1. 🛡️ **Route-Based Split Tunneling, Not Socket Hooking**:
   - Traditional game boosters often use Windows Filtering Platform (WFP), LSP, or packet-divert drivers (e.g. WinDivert) to hook sockets per process. Sitting inside a game's network stack triggers immediate red flags from modern kernel-level anti-cheats (Vanguard, VAC, BattlEye, EAC, RICOCHET).
   - Lagvex operates **purely at the operating system routing table level**. It touches zero game memory, injects zero DLLs, and intercepts zero game sockets. To the game and anti-cheat, traffic flows out of a standard network card.
2. 🔌 **Official Layer-3 Virtual Adapters**:
   - **Windows**: Uses **WinTun**, WireGuard's signed, high-performance kernel TUN driver.
   - **macOS**: Uses native **`utun`** (built into the Darwin kernel; zero third-party drivers needed).
   - **Linux / Steam Deck**: Uses native **/dev/net/tun**.
   - If the kernel adapter cannot be created (missing privileges or no TUN device), the engine keeps a control-only relay session instead of failing. The dashboard labels that state **userspace fallback** and states plainly that game traffic is not routed through the relay.
3. ⚡ **Pure Go Monorepo**:
   - Both the VPS Relay daemon (`lagvex-relay`) and the Client engine (`lagvex-client`) are written in Go. Zero-allocation data pipelines and buffer recycling via `sync.Pool` minimize GC pause times to preserve microsecond-level packet throughput.
4. 🐧 **High-Performance Kernel-Level NAT (Linux relay only)**:
   - The relay server avoids slow user-space packet translation. It writes raw IPv4 frames into `/dev/net/tun` (`lagvex0`) where the Linux kernel executes hardware-speed `iptables` NAT MASQUERADE, conntrack, and TCP MSS clamping. Relay hosting needs Linux; Windows and macOS clients use WinTun/`utun` for the client side only.

---

## 2. 🛣️ The Journey of a Game Packet

```text
   [ 🎮 Game Process: e.g. VALORANT-Win64-Shipping.exe ]
          │
          │ Sends UDP packet to game server: 13.250.45.10:7000
          ▼
   [ 🪟 Operating System Network Stack ]
          │
          │ Routing table check: 13.250.0.0/15 matches on-link -> "Lagvex" adapter
          │ (Unrelated traffic like Discord/Browser matches 0.0.0.0/0 -> Physical NIC)
          ▼
   [ 🔌 Virtual TUN Adapter (WinTun / utun) ]
          │
          │ Reads raw IPv4 packet from user-space ring buffer
          ▼
   [ ⚡ Lagvex Client Engine (lagvex-client) ]
          │
          │ Authenticates & encrypts packet with ChaCha20-Poly1305 AEAD + replay window (17-byte header + 16-byte tag)
          ▼
   [ 📦 Standard UDP Socket ]
          │
          │ Sends to Relay VPS: 198.51.100.5:51820
          │ (Pinned /32 route forces this packet out the physical network gateway)
          ▼
   ═══════════════════════════ International Fiber Backbone ═══════════════════════════
          ▼
   [ ☁️ Lagvex Relay Server (lagvex-relay on Linux VPS) ]
          │
          │ Receives datagram on port 51820
          │ Validates Session ID, checks replay window, and decrypts via ChaCha20-Poly1305
          │ Verifies anti-spoofing inner source IP
          │ Writes raw IPv4 packet into TUN interface: lagvex0
          ▼
   [ 🐧 Linux Kernel (VPS) ]
          │
          │ iptables NAT MASQUERADE rewrites source IP to VPS Public IP
          │ Clamps TCP MSS to PMTU
          │ Forwards packet out physical WAN interface (eth0)
          ▼
   [ 🎯 Target Game Server (Singapore / Tokyo / Frankfurt) ]
```

---

## 3. 🛡️ Routing Table Safety Rules (Windows)

The routing engine ([`pkg/client/routes_windows.go`](../pkg/client/routes_windows.go)) enforces three strict network safety invariants:

### A. 📌 Pinned `/32` Relay Route (Anti-Loop Guarantee)
Before any game route is installed, a host route (`/32`) pointing to the Relay VPS IP is pinned through the physical network gateway:
```cmd
netsh interface ipv4 add route prefix=<RELAY_IP>/32 interface=<PHYSICAL_INDEX> nexthop=<DEFAULT_GATEWAY> metric=1 store=active
```
> [!IMPORTANT]
> **Why this is critical:** If a game's CIDR range encloses the Relay server's own IP address, packets destined for the relay would get sucked into the tunnel, creating an infinite routing loop that cuts off internet access. The pinned `/32` route guarantees the tunnel endpoint always exits via the physical gateway.

### B. 🔗 On-Link Game Routes
When routes for a game's CIDR blocks are installed, they are configured as **on-link** (no nexthop gateway specified):
```cmd
netsh interface ipv4 add route prefix=<GAME_CIDR> interface=<WINTUN_INDEX> metric=1 store=active
```
> [!NOTE]
> WinTun is an NDIS Layer-3 interface without an ARP/NDIS link layer. Specifying a nexthop leaves Windows waiting on an ARP resolution that can never occur, resulting in silent packet drops.

### C. 💾 `store=active` (RAM-Only Safety Brake)
Every route added by Lagvex is flagged with `store=active`. These entries exist only in volatile memory (RAM). If the process crashes or power is lost, the routing table automatically reverts to a clean state upon reboot.

---

## 4. 👁️ Smart Process Watcher Lifecycle

AWS and Azure cloud blocks encompass thousands of third-party services. Leaving game routes permanently active would route unrelated web services through the relay.

Lagvex solves this using the **Smart Process Watcher** ([`pkg/client/watcher.go`](../pkg/client/watcher.go)):
1. ⏳ **Lobby / Idle**: While the user is connected to a relay and waiting in a lobby, no game routes are installed.
2. 🔍 **Scanning**: The watcher monitors system processes every 2 seconds for known game binaries (e.g. `VALORANT-Win64-Shipping.exe`, `cs2.exe`, `TslGame.exe`).
3. 🚀 **Game Launch Detected**: Routes for the active game and region are dynamically injected into the routing table.
4. 🧹 **Game Exit Detected**: Routes are automatically deleted, returning the network to direct routing immediately.
5. ⚡ **Manual Override**: Users can toggle **Force Route Always** in the Dashboard to bypass the watcher when testing private servers or custom matches.

---

## 5. 🔒 Security & Anti-Abuse Measures

| Threat Vector | 🛡️ Defense Mechanism |
|---|---|
| **Unauthorized Open Relay Abuse** | Handshake authenticated via HMAC-SHA256 over Pre-Shared Key (PSK). Non-matching packets are silently dropped. |
| **Replay Attacks** | Handshake payload contains a Unix timestamp checked against a 120-second clock skew tolerance window. |
| **IP Spoofing Between Players** | The relay verifies that the inner IPv4 source address strictly matches the IP assigned to that session ID. |
| **Intranet & Metadata Scanning** | The relay drops packets destined for RFC 1918 subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback (`127.0.0.0/8`), and cloud metadata (`169.254.169.254`). |
| **Dynamic Roaming (Wi-Fi ↔ 4G)** | Client UDP endpoint (`IP:Port`) is dynamically updated on every valid data packet, preventing session disconnects during network switches. |

### Dual-path redundancy (Redundant-lite Phase 1 + 2)

Relays register a PSK-derived connection group from a handshake extension and
keep a bounded first-copy-wins dedup window (`pkg/relay/dedup.go`: 2 s TTL,
4096 sequences per group, 10-minute group TTL, 256 groups max). A dual
session's data frames carry an 8-byte sequence trailer that is stripped before
FEC and TUN delivery, and the relay reports a contiguous ACK watermark
(`DualAck`, at most one per 20 ms) so the client knows what arrived.

The client (`pkg/client/engine_dual.go` + `pkg/client/dual.go`) sends data
frames only to the primary relay and retransmits unacknowledged frames —
primary first, then the standby relay — with the ring, thresholds and measured
counters documented in `docs/PROTOCOL.md` §9. The dashboard toggle
(`/api/dual-path/toggle`) reflects the real standby handshake result, and
`/health` still reports `dualSessions`, `dualGroups`, `dualForwarded` and
`dualSuppressed` so operators can watch the relay side.

---

## 6. 🎨 Esports Cockpit HUD & Telemetry Architecture

The frontend (`ui_source/`) is a React + Vite dashboard built into static assets
that are embedded in the Go binary and served from `web/`:

```text
  [ Embedded Go HTTP Server (:18888) ]
                  │
                  ├─► GET /api/status       (Engine state, active game, active relay)
                  ├─► POST /api/connect     (Triggers handshake & dynamic routing)
                  ├─► POST /api/disconnect  (Tears down routes & releases IP)
                  ├─► GET  /api/probe-relays (Measures & ranks the configured relays)
                  ├─► POST /api/tcping      (TCP handshake RTT to a public IP)
                  ├─► POST /api/dual-path/toggle (Redundant-lite ARQ mode)
                  └─► Static Assets         (React dashboard build)
```

### Local dashboard API

Every route below is served by the loopback HTTP server (`pkg/client/ui_server*.go`).

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/status` | Engine state, active relay/game, measured counters |
| GET | `/api/games` | Game catalog with regions and CIDRs |
| GET | `/api/relays` | Configured relays (public + custom) |
| GET | `/api/best-relay` | Best relay by the current measured ranking |
| GET | `/api/probe-relays` | Measures configured relays (`?ids=` subset, `?force=1` bypasses the 20 s cache) |
| GET | `/api/route-detail` | Per-hop route detail for the active session |
| GET | `/api/game-telemetry` | Cached regional reference telemetry |
| GET | `/api/game-connections` | Remote endpoints of running games, read from the local OS socket table |
| GET | `/api/game-server-history` | Persisted per-game measured game-server endpoints (`?gameId=`) |
| GET | `/api/detected-processes` | Known game processes currently running |
| GET | `/api/system-info` | OS, interfaces and tun support |
| GET | `/api/failover/history` | Recent relay handovers |
| GET | `/api/diagnostics/export` | Portable settings snapshot (never contains PSKs) |
| GET | `/api/advisor` | Route advice for the current selection |
| POST | `/api/game-server-history` | Records one measured game-server endpoint (`avgMs > 0`, IP literal only) |
| POST | `/api/connect` | Connects to a relay (`relayEndpoint` + `psk`) |
| POST | `/api/disconnect` | Tears the session down |
| POST | `/api/add-game` / `/api/add-relay` | Adds a custom game or relay |
| POST | `/api/tcping` | TCP handshake RTT to a public IP |
| POST | `/api/ping-relay` / `/api/test-relay` | Relay reachability checks |
| POST | `/api/dual-path/toggle` | Opt-in redundant-lite ARQ mode |
| POST | `/api/fec/toggle` | FEC parity on/off |
| POST | `/api/peer-routing/toggle` | Peer routing on/off |
| POST | `/api/full-tunnel/toggle` | Full tunnel on/off |
| POST | `/api/failover/toggle` | Automatic relay failover on/off |
| POST | `/api/tweak` | Applies a system tweak |
| POST | `/api/optimize-profile` | Runs the profile optimizer |
| POST | `/api/diagnose-network` / `/api/diagnose-bufferbloat` | Network diagnostics |
| POST | `/api/relaunch-admin` | Relaunches the client with elevated rights |
| POST | `/api/shutdown` | Stops the client |
| POST | `/api/diagnostics/import` | Applies a settings snapshot |
| POST | `/api/squad/create` / `/join` / `/leave` / `/room` / `/heartbeat` | Squad signalling |

### Local API guard

The dashboard API is bound to loopback, but loopback alone does not stop a
malicious web page from POSTing to it. Every state-changing request
(`POST`/`PUT`/`PATCH`/`DELETE`) must carry `X-Lagvex-Request: 1`, and a
request that presents an `Origin` header must come from a loopback origin.
Browsers cannot add the custom header cross-site without a CORS preflight,
which the server never grants. The React dashboard sends it automatically;
command-line callers must add it explicitly. Read-only `GET` requests are
unchanged. Endpoints that open outbound sockets on the user's behalf, such as
the TCP ping tool, are POST-only so this guard also covers them.

### Key Front-End Subsystems:
1. 🛡️ **Resolution-Independent Vector Emblems**: Custom inline SVG insignias for all 11 games provide instantaneous rasterization with zero blur on 4K / Ultrawide gaming monitors.
2. 📡 **Tactical Dual-Ring Radar Scanner**: Animated Canvas/SVG radar sweep line with synchronized radial sonar rings provides visual feedback during connection negotiation and process discovery.
3. ⚡ **Active Laser Packet Stream**: When active, a dynamic SVG neon beam with traveling light particles illuminates the connection bridge, demonstrating active packet forwarding at a glance.
4. 🔔 **Glassmorphic Toast Bus**: A centralized event dispatcher queues non-blocking feedback pills (probe results, copy events, route alerts) with smooth spring transitions and automated lifecycle disposal.

---

## 7. 🌐 Relay Topology (self-hosted)

**Lagvex ships no relay fleet.** The bundled relay catalog is empty and the
client never silently routes traffic through infrastructure the project does
not control. A relay is something the user (or a squad) deploys near the game
region — see [docs/DEPLOYMENT.md](DEPLOYMENT.md) and [docs/COST.md](COST.md).

```text
                    ┌──────────────────────────┐
                    │      LAGVEX CLIENT       │
                    └────────────┬─────────────┘
                                 │
              ┌──────────────────┴──────────────────┐
              │  Measured Relay Ranking Selector   │
              ▼                                     ▼
   [ self-hosted relay A ]              [ self-hosted relay B ]
   (e.g. Singapore / Tokyo /             (standby for redundant-lite
    Frankfurt / Sydney …)                 ARQ, opt-in)
```

- **Continent-grouped region selector**: The dashboard groups the catalog's regional server clusters by their `continent` label (Asia-Pacific, Europe, North America, South America, Oceania, Middle East & Africa) and only offers regions that have verified CIDRs.
- **Measured selection, not geography**: relay selection ranks the relays the user configured by **measured** path quality (RTT, jitter, loss and the relay→region-anchor segment), not by geographic proximity. A relay in the wrong city can rank below a nearby one, and the dashboard shows why.
- **Relay hosting is Linux-only**: the relay needs `/dev/net/tun`; Windows and macOS can run the client but cannot host a relay.
- **Squad signaling** runs on the relay's HTTP port (the data-plane port + 1):
  the client proxies `/api/squad/*` to the connected relay, so a squad shares
  one room even when the members are on different networks. The signaling
  port is public and CORS-open, so the relay rate-limits room creation and
  room lookups per source IP, caps the number of live rooms, and generates
  cryptographically random room codes (a room code grants access to the
  roster and must not be guessable).

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- 💸 [**Running Cost & Self-Hosting**](COST.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
