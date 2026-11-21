# 📝 Changelog 📜

> 🔗 [Back to Project README.md](README.md) | [Contributor Guide](CONTRIBUTING.md)

All notable changes to the **Lagvex** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> Note: tags `v1.1.1`–`v1.1.3` were created without changelog entries; their contents are
> not reconstructed here. The v1.2.0 section below describes everything shipped since
> v1.1.0.

---

## [Unreleased]

- Docker Compose now requires an explicit relay PSK. Relays using the old public default must rotate their key and update clients.

## [1.10.0] - 2026-11-21

### ✨ Free Mode, Burst-Loss RS-FEC, Anti-DPI & Diagnostics

- **Free Mode (Zero-VPS) — WireGuard & Cloudflare WARP Split-Tunnel**: Integrated
  pure-Go `wireguard-go` bridged directly onto Lagvex's cross-platform virtual
  adapter (WinTun / `utun` / `/dev/net/tun`) and `RouteManager`. Users can now
  activate a free Cloudflare WARP profile in one click (`POST /api/relay/register-warp`)
  or import any standard WireGuard `.conf` (`POST /api/relay/import-wireguard`)
  while keeping CIDR split-tunneling and process watching intact. WireGuard
  private keys and preshared keys are redacted from `/api/relays` and exports.
- **Relay Share URIs & Community Lists**: Added `lagvex://` share URI import/export
  (`POST /api/relay/import-uri`, `POST /api/relay/export-uri`), HTTPS subscription
  import (`POST /api/relay/import-subscription`, SSRF-guarded), and relay deletion
  (`POST /api/relay/delete`).
- **Cauchy Reed-Solomon FEC (`GF(2^8)`)**: Upgraded FEC beyond single-packet XOR
  to multi-parity Reed-Solomon (`RS 6:2`, `RS 4:2`, `RS 4:3`) using a backwards-
  compatible `buf[9]` metadata byte, recovering multi-packet burst losses (5–15%)
  on both client and relay without CGO.
- **Packet Hedging & Independent-PSK Dual-Path**: Small UDP game packets (`≤256B`)
  are hedged during active loss (`≥2%`) or recent FEC recovery with a measured
  `packetsHedged` counter. Dual-Path ARQ now supports independent PSKs per relay.
- **Anti-DPI UDP Obfuscation**: Added opt-in zero-overhead header scrambling
  (salted by the random Poly1305 tag) plus variable-length handshake padding
  (`POST /api/obfuscation/toggle`), transparently auto-detected by relays on the
  same UDP port.
- **Free Local Diagnostics Suite**: Added PingPlotter-style hop-by-hop Route
  Tracer (`POST /api/traceroute`), UDP DNS Resolver Benchmark (`POST /api/dns-benchmark`),
  Auto Path-MTU Discovery (`POST /api/diagnose-mtu`), live `relay→game-server`
  hop population inside `/api/route-detail`, monthly transfer quota tracking, and
  stale-aware RTT + FEC/WireGuard mode tags in the Floating HUD.

---

## [1.9.1] - 2026-11-21

### 🛡️ Security

- Every request body on the relay's public signaling port is capped at 8 KiB
  by the rate-limit middleware. The squad and probe handlers decode small JSON
  objects, so an unbounded body was pure attack surface: an anonymous caller
  could stream JSON that never closes and make each request allocate the whole
  stream. Oversized bodies now fail to decode and are answered with 400.
- The relay probe's rate limit now runs on two limiters of its own: the 2 s
  burst window plus a sustained 600/hour per-IP cap. The burst window alone
  still allowed 1,800 probes/hour, which is enough to turn the endpoint into a
  slow port scanner; the hourly cap bounds it to a fraction of that while
  staying ten times what a real dashboard uses (~60/hour).
- `POST /squad/leave` no longer answers success for a body that does not
  decode: nothing changed, and the caller deserves a 400 instead.

---

## [1.9.0] - 2026-11-20

### 🧪 Test, correctness and docs sync

- The relay integration test is deterministic again: the userspace TUN discards
  writes until the write sink is attached, so the harness now waits for the sink
  before sending traffic (previously the first forwarded packet could be lost).
- **Anchor mapping fix**: a generic `eu-` prefix no longer shadows specific
  cities, so `eu-london` measures the London anchor and `eu-paris` the Paris
  anchor instead of Frankfurt. Added a US Central (Ohio) anchor and mappings
  for the remaining regions; the client table and the relay allowlist are now
  guarded by a test that walks every shipped game region.
- **Tweak honesty fix**: `/api/tweak` used to treat an absent field as "false",
  so applying tweaks silently disabled the Smart L4 filter and reported it as
  an applied optimisation. Absent fields now mean "leave unchanged", and the
  dashboard exposes the filter explicitly.
- New dashboard unit tests (`node:test`, no new dependencies) cover the CIDR
  and coverage helpers, the display formatters and the cost projection; CI
  gained a dashboard job (tests, type-check, asset build, embedded-asset guard).
- Docs now match the code: the protocol lists the `0xA` message type and the
  16-entry anchor allowlist, the architecture carries the full local API table
  and no longer implies a bundled relay fleet, and the cost guide states that
  relay hosting is Linux-only and that the traffic figure is a planning
  estimate.

### 🛡️ Squad signaling

- Squad rooms actually reach the relay again. The client built the signaling
  URL from the relay's data-plane address, but a dual-stack dialer reports an
  IPv4-mapped IPv6 form (`[::ffff:127.0.0.1]:4433`) that produced an invalid
  URL; every room silently fell back to a local-only room no teammate could
  join. The address is unmapped before the URL is built, with tests for each
  form.

- Room codes are unguessable: the relay generates them with a CSPRNG (and the
  client for its local fallback), replacing a four-digit counter driven by
  `Math.random` in the dashboard. Room creation and lookups are rate limited
  per source IP, the number of live rooms is capped, and codes can be sent in a
  query parameter or a JSON body.

- Relay errors reach the dashboard instead of being swallowed: a 429 rate
  limit, a full room or a room that no longer exists keeps its real status and
  message, and an unreachable signaling port returns an explicit 502. The
  client no longer creates a local-only room when a connected relay rejects the
  request; the deliberately offline fallback still exists, is labelled in the
  squad panel, and only runs when no relay is connected at all.

- The dashboard footer shows the running build version from `/api/system-info`
  instead of a hardcoded `v1.0.0`, and it no longer wraps when the version is
  long.

### 📊 Measured telemetry and route honesty

- Tunnel RTT is only shown once the engine has received a real keepalive Pong
  (`pingSampled`), so a stale ISP measurement can no longer appear as "TUNNEL
  LIVE" latency. Sub-millisecond loopback/LAN RTT is shown as `<1 ms`, and the
  jitter card no longer tells a connected user to "Connect to measure" when the
  measured variation is simply below 1 ms.

- Relay instability now requires repeated recent loss (two of the last three
  runs), so the embedded relay's first sweep before it starts no longer brands
  it UNSTABLE for the session. Sub-millisecond timing noise is ignored when
  judging RTT spread.

- The dashboard header names the game reported by the engine
  (`activeGame`), not the game merely selected in the UI. With a connected
  tunnel but no active game it says "TUNNEL CONNECTED", and the system-ready
  badge shows the real adapter for the OS (WinTun NDIS, macOS utun, Linux
  dev/net/tun) instead of a fixed "kernel accelerator" claim.

- When the kernel adapter cannot be created (for example a non-root Linux
  session), the dashboard now shows an amber **userspace fallback** banner and
  badge: "game traffic is not routed". Ping labels switch from "live
  accelerated ping" to "live relay RTT", CIDR rows read "Not routed", and the
  pipeline shows "No TUN routes". The engine log no longer claims to be
  accelerating in that state.

- Selecting another game while a tunnel is active no longer relabels the live
  RTT as that game's latency: the Boost page states which game the tunnel is
  active for, and that the newly selected game must be boosted after stopping
  the current tunnel.

- The relay gained a game-server TCP probe (`POST /probe/gameserver`, one
  request per IP per two seconds, public IP literals only) and the dashboard
  gained `POST /api/relay-tcpping`, so the Boost page can show the
  **relay→game-server** segment measured by the relay itself — the last
  unmeasured segment of the route (ExitLag's `PingToDS` equivalent). The
  panel also shows the combined "you→relay→server" total only when both
  segments are measured.

- Probe history reports percentiles (`rttP50Ms`, `rttP95Ms`) computed with
  linear interpolation over the retained runs, so tail latency is visible
  next to the average — no value appears without samples.

- RTT samples now carry an age (`pingAgeMs`, `pongTotal`): a sample older than
  three keepalive periods shows as a dash with "last sample is stale" instead
  of being presented as live, mirroring how ExitLag hides a frame stat when
  its own counter and the sample disagree.

- The process scan reports a measured per-core CPU % per game process
  (`/api/detected-processes`: `cpuSupported`, `cpuPercent`), read from the OS's
  own counters (`/proc` on Linux, `GetProcessTimes` on Windows) with no
  injection. macOS reports "not measured" because no portable,
  injection-free counter exists there. Cards and the Boost hop show
  `CPU 12.3%`, or `CPU —` with an explicit tooltip.

- Route advice reports the relay's measured loss instead of a hardcoded "0%
  packet loss"; a lossy but faster relay is described with its actual loss.

- A new client↔relay end-to-end test performs a real UDP handshake, waits for
  an authenticated keepalive Pong, verifies the relay sees and then loses the
  session on disconnect, and covers the pre-flight refusal paths (bad PSK,
  game without a region, missing profiles). Client package coverage rose from
  49.6% to 57.8%.

- Windows and macOS cross-compilation is green again. A console-window fix left
  an unused `os/exec` import in both OS-specific route managers, which only the
  cross-compile jobs caught; local verification now includes `GOOS=windows` and
  `GOOS=darwin` builds and vet passes.


---

## [1.8.0] - 2026-11-18

### ♻️ Dual-path readiness (Redundant-lite Phase 1)

- The relay accepts a dual-handshake extension, derives a PSK-keyed dedup
  scope from the group id, forwards the first copy of a `(group, sequence)`
  pair and suppresses later copies inside a 2-second TTL.
- `DualAck` progress reports the **highest contiguous** forwarded sequence, so
  a client can tell exactly which frames still need retransmission.
- Dual path stays opt-in and both relays must share one PSK.

---

## [1.7.0] - 2026-11-17

### 🛡️ Loopback API guard

- Every state-changing request must carry `X-Lagvex-Request: 1`, and requests
  with an `Origin` header must come from a loopback origin; a cross-site page
  can no longer drive the local dashboard API.

### 🧪 TCP ping tool

- A bounded TCP ping tool (min/avg/max/loss) for explicit public game-server
  IPs, with private/loopback/link-local addresses refused, rate limiting and
  RST-on-closed-port handling. Observed game endpoints can be sent to it with
  one click.

### 📈 Measured node ordering

- Relay nodes can be sorted by measured route, loss or recency, and sparkline
  points show sample number, RTT and change from the previous sample.

---

## [1.6.0] - 2026-11-16

### ⚡ Pre-measured standby relay

- While connected, one alternative relay is kept pre-measured as a standby, so
  a degradation promotes it without waiting for a sweep.

### 🏓 Unprivileged ICMP echo measurement

- When the OS permits unprivileged ICMP, the prober adds an ICMP echo RTT next
  to the tunnel probe; it is labelled with its measurement method.

---

## [1.5.0] - 2026-11-16

### 🔁 Stability-aware ranking and failover history

- Relay ranking now accounts for measured stability (jitter and loss), not just
  the latest RTT, and the dashboard keeps a history of relay handovers with
  the old/new ping and the handover outcome.

---

## [1.4.0] - 2026-11-16

### 📍 Measured per-segment ranking

- Relays are ranked by the measured client→relay **and** relay→region segment
  RTT instead of a flat ping, so the recommended node reflects the real path.

### 🌐 Opt-in full tunnel

- Full-tunnel mode routes all IPv4 traffic through the relay; split tunnelling
  remains the default.

### ⭐ Recent relays and best-route badge

- The Nodes tab keeps a recent-relay shortcut and marks the best measured
  route from the latest sweep.

---

## [1.3.0] - 2026-11-16

### 🎯 Relay-measured anchor probe

- New `AnchorProbe`/`AnchorProbeReply` control messages: the client asks the
  relay to measure its own RTT to one **allowlisted** anchor, and the route
  detail view combines tunnel RTT + relay RTT into a measured via-relay
  estimate with a direct-vs-via-relay improvement value.
- The relay refuses unknown ids, applies a per-session rate limit (2s minimum
  interval, 60 per hour) and only uses the fixed target list, so it cannot be
  pointed at arbitrary hosts.

### 📈 Measured throughput view and endpoint history

- The Boost history tab charts the engine's own up/down byte rates with live
  rates and session totals, and shows a waiting state until samples exist.
- The game-connections card remembers the last three distinct remote IPs the
  selected game was observed talking to; no geolocation and no guessing.

---

## [1.2.0] - 2026-11-15

### 🧭 Honest telemetry overhaul (breaking behavior changes)

- Connect now validates the selected game region against real CIDRs and **requires a
  relay PSK**; there is no fallback key anywhere in the client, the UI or the prober.
- Diagnostics, advisor, bufferbloat and game-telemetry responses only contain measured
  values. Probes that fail return explicit errors instead of placeholder numbers.
- Tweak results list only the changes that actually applied.

### 📊 Measured node history

- The prober retains up to 60 probe runs per endpoint and reports measured average RTT,
  standard deviation, average loss and a trend series.
- The Nodes tab renders a sparkline, latest/avg RTT, loss and the retained sample count.

### 🧭 Route detail (`GET /api/route-detail`)

- Measures the local gateway, a direct relay probe (plus retained avg/std), the live
  tunnel echo and loss, and a public cloud anchor for the selected region.
- The relay → game-server segment is reported as unmeasured instead of guessed, and the
  tunnel overhead appears only when both same-endpoint measurements exist.

### 🔌 Game process awareness

- `GET /api/detected-processes` scans the process table once per poll and the UI marks
  running games; the boost view names the detected executable.
- `GET /api/game-connections` lists the sockets held by running game processes on
  Windows, macOS and Linux (read-only OS socket tables; no injection).
- `POST /api/peer-routing/toggle` (off by default) installs `/32` routes for the public
  peers the game actually talks to, with a 64-route cap and stale-route cleanup.

### 🧾 Diagnostics snapshot

- `GET /api/diagnostics/export` writes measured state plus non-secret settings; relay
  PSKs are never included. `POST /api/diagnostics/import` merges settings back and
  reports skipped entries as warnings.

### 🛡️ Reliability and platform fixes

- FEC parity flushes on schedule, hedging is bounded and failover races are fixed.
- Non-Linux relay hosts fail fast instead of silently dropping packets.
- macOS DNS diagnostics read the live resolver list via `scutil --dns`.
- QoS wording matches reality: the relay socket is tagged with the Windows QoS API or
  `IP_TOS` (best effort) depending on the platform.

### 🖥️ Dashboard

- The legacy `web/app.js` / `web/style.css` frontend was removed; the React/Vite app in
  `ui_source/` builds into `web/` and is embedded in the binary.
- Added the draggable Floating HUD, honest empty states (`—` plus the reason) and a
  single-copy squad room UI without invented Discord links or route sharing.
- Squad rooms now require a host/player name, reject unknown room codes with 404, clear
  stale RTT on heartbeat and use a neutral "In room" status.

### 🗂️ Catalog and relay behavior

- The bundled relay catalog is empty; tests fail if a `lagvex.org` relay ever reappears.
- Relay PSKs are redacted from every API response and never leave the machine.

---

## [1.1.0] - 2026-10-07

### ⚡ Layer-4 Smart Game Tick Filter & Split-Tunneling Inspection
- 🎮 **Layer-4 Game Loop Prioritization (`pkg/client/engine.go`)**:
  - Implemented high-throughput Layer-4 packet classification that selectively accelerates UDP game tick traffic through the tunnel while letting bulky TCP patch downloads and game updates flow directly over the local ISP connection.
  - Eliminates bandwidth saturation during background game downloads and keeps gaming loop jitter near zero.
  - Dynamically toggleable at runtime via the dashboard UI and `POST /api/tweak`.

### 🪟 In-Game Draggable Floating HUD Widget
- 📊 **Real-Time Glassmorphic Cockpit Pill (`web/`, `web/style.css`, `web/index.html`)**:
  - Introduced an optional, draggable in-game overlay widget displaying real-time RTT latency, jitter, live packet loss percentage, and Zero-RTT FEC active status.
  - Smooth drag-and-drop repositioning with automatic browser local storage coordinate persistence.
  - Expandable detailed telemetry metrics card with 1-click toggle controls.

### 🛡️ Adaptive Forward Error Correction (FEC) Standby & Optimization
- 📉 **Zero-Overhead Standby Mode (`pkg/protocol/fec.go`, `pkg/client/engine.go`)**:
  - Corrected return semantics and change tracking in `AdaptiveFECController.UpdateLoss`: returns exact `(blockSize int, changed bool)` to prevent redundant reconfiguration and log spamming.
  - Added zero-redundancy Standby mode (`k = 0`) to `FECEncoder`: when packet loss drops below $0.5\%$, the encoder immediately ceases XOR parity generation and packet emissions, saving up to $17\%$ bandwidth and conserving CPU cycles.
  - Automatically resets pending partial blocks and scales dynamically (10:1 $\to$ 6:1 $\to$ 4:1) upon network degradation.
  - Synchronized `SetFECEnabled` toggle state across controller, encoder, and engine telemetry.

### 🔄 Relay IP Pool 2-Pass Allocation & Sticky IP Affinity
- 🔒 **Hijack-Resistant Client Reservations (`pkg/relay/pool.go`, `pkg/relay/pool_test.go`)**:
  - Implemented a 2-pass allocation algorithm with dedicated `reservedIPs` mapping to guarantee that a temporarily disconnected player's 24-hour reserved inner IP cannot be stolen by new clients during reconnect storms.
  - Pass 1 allocates unused and unreserved IPs; Pass 2 reclaims un-used reservations only when the entire subnet is completely full.

### 🎯 Route Advisor Direct Host Probing
- 🌐 **Valid Host IP Resolution (`pkg/client/ui_server.go`)**:
  - Updated Route Advisor gateway probing to use `netip.ParsePrefix` and `p.Addr().Next()` to probe the first usable host (`.1:443`) rather than the invalid subnet base network address (`.0:443`).
  - Ensures accurate real-time comparison between direct ISP peering and accelerated VPS relay paths.

### 🚀 Cross-Platform Kernel Stability & Concurrency Hardening
- 🍏 **Darwin/macOS Zero-Allocation Data Plane (`pkg/client/tun_darwin.go`)**:
  - Implemented `utunBufPool` (`sync.Pool`) for macOS `ReadPacket` and `WritePacket`, eliminating per-packet heap allocations.
  - Fixed C NUL-terminated string parsing in `getUTunName` to prevent slice underflow and runtime panic.
- 🪟 **Instant WinTun Teardown (`pkg/client/wintun_windows.go`, `pkg/client/engine.go`)**:
  - WinTun `Close()` explicitly signals `readEvent` via `SetEvent`, allowing `ReadPacket` to unblock instantly in 0ms instead of waiting for a 1000ms timeout.
  - Guarded `pumpWinTunToUDP` against high-CPU tight-loop spinning if the virtual adapter closes unexpectedly.
- 🔒 **Race & Deadlock Elimination (`pkg/profiles/manager.go`, `pkg/client/engine.go`)**:
  - Fixed Go `sync.RWMutex` recursive read-lock deadlock by extracting `findGameByIDLocked` in `profiles.Manager`.
  - Eliminated data races in `Engine.Disconnect()` by strictly using local captured variables under lock.
  - Prevented zombie UDP sockets on interrupted failover handovers and aborted connection attempts.
- 🐧 **Linux Relay BBR Congestion Control & Prefix Validation (`scripts/install_relay.sh`, `pkg/relay/config.go`)**:
  - Enabled Linux kernel TCP BBR congestion control in the automated VPS installer script.
  - Enforced strict client subnet prefix validation between `/8` and `/30` to prevent infinite loops in IP allocation.

### 🔍 Universal Case-Insensitive Process Detection
- 🎯 **Cross-Platform Matching (`pkg/client/watcher_windows.go`, `pkg/client/watcher_other.go`)**:
  - Standardized process name comparison using `strings.EqualFold` with transparent handling of with/without `.exe` suffixes across Windows, macOS, and Linux.

### 🌐 100% English Localization & Brand Consistency
- 🌍 **Complete UI & Telemetry Localization**:
  - Refined all dashboard tabs, settings modals, HUD badges, toasts, and notifications to consistent, professional English.

---

## [1.0.0] - 2026-09-22

### 🌐 Enterprise Squad Signaling Hub & Cross-Network Sync
- 📡 **Relay Squad Signaling Hub (`pkg/relay/server.go`)**:
  - Embedded REST signaling hub on relay servers (`/squad/create`, `/squad/join`, `/squad/room`, `/squad/heartbeat`, `/squad/leave`).
  - Enables squad party members in different physical locations to synchronize matchmaking, latency, and route selection over the internet.
  - Automated stale peer pruning with 30-second heartbeat sweeps.
- 🔄 **Relay Squad Client Proxying (`pkg/client/ui_server.go`)**:
  - Client automatically routes squad management requests to the active connected relay server.
  - Graceful fallback to local in-memory party manager when running standalone or disconnected.

### 💾 Persistent User Configuration & Profile Customization
- ⚙️ **User Config Store (`pkg/client/config_store.go`, `%APPDATA%\Lagvex\user_config.json`)**:
  - Atomic JSON persistence for gamer nickname, custom game profiles, custom relay endpoints, and Windows network optimization toggles.
  - Automatically loads and integrates with the active profile catalog on client startup.

### 🪟 Windows TCP MSS Clamping & Network QoS
- ⚡ **TCP MSS Clamping at 1360 Bytes (`pkg/client/routes_windows.go`)**:
  - Clamps WinTun adapter MTU to 1400 bytes, enforcing an effective TCP MSS limit of 1360 bytes (`MTU - 40`).
  - Completely prevents TCP packet fragmentation and eliminates game lobby/matchmaking freeze across fiber and PPPoE networks.

### 🎯 Truthful Game Network Architecture & Zero Struck-Through Numbers
- 💎 **Authentic Peering Metadata**:
  - Completely removed hardcoded fake comparison numbers (`~~80.2ms~~ -> 38.7ms`, `52% Faster`) from game cards on Home and Library tabs.
  - Replaced with genuine server infrastructure routing protocols:
    - **Valorant**: `Riot Direct` (Direct Riot Games Peering)
    - **Counter-Strike 2**: `Valve SDR` (Valve Steam Datagram Relay)
    - **PUBG: BATTLEGROUNDS**: `Steam` (AWS Southeast Asia)
    - **Apex Legends**: `EA / Steam` (Anycast Core)
  - Display sleek `⚡ FastPath Ready` when idle, dynamically switching to `🟢 TUNNEL LIVE: [livePing] ms` with real socket telemetry when boosting.

### ✨ Fluid Micro-Interactions & Animation System
- 🎨 **Natural Spring Easing & Tab Transitions**:
  - Replaced harsh linear transitions with Apple/Linear spring easing: `cubic-bezier(0.16, 1, 0.3, 1)`.
  - Butter-smooth tab entrance transitions (`tab-enter`) across all 6 tabs (Dashboard, Booster, Library, Squad, Nodes, Tweaks).
  - Smooth card hover lift and 700ms eased cover art zoom (`card-smooth`).
  - Backdrop blur fade (`modal-backdrop-animate`) and scale pop-in (`modal-pop-animate`) for all dialogs.
  - Calm glowing state indicators replacing jarring continuous spins and harsh ping blinks.

### 🎮 Zero-Config Standalone Client & Repository Cleanliness
- 🚀 **Embedded Game Profiles & Community Relays (`pkg/profiles/embedded/`)**:
  - `profiles.json` (11 games, 189 CIDRs) and `relays.json` (15 global community relays across 5 continents) compiled directly into the binary via `//go:embed`.
  - Gamers can download a single `lagvex-client.exe` file and boost immediately without configuration files.
- 🧹 **Repository Cleanliness**:
  - Pruned legacy web prototype and temporary design assets.
  - Self-contained production build.

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

## [1.1.1] - 2026-11-09

### Bug Fixes and Performance

- use tag name from create event context ([4dc1ded](https://github.com/ThanhNguyxnOrg/lagvex/commit/4dc1ded6fea78ffafe9d014742b67e2d69c36b05))

## [1.1.2] - 2026-11-10

### Bug Fixes and Performance

- migrate golangci-lint config to v2 ([9214732](https://github.com/ThanhNguyxnOrg/lagvex/commit/9214732e8a5d73fc81613ae333ab0576db2d324b))
- make golangci-lint a required quality gate ([00767e8](https://github.com/ThanhNguyxnOrg/lagvex/commit/00767e853ac9300d49e914ddb44b989a9a41252e))
- address strict Go lint findings ([2735012](https://github.com/ThanhNguyxnOrg/lagvex/commit/27350123e9f27940f80f6781c470ed25b590b2ed))

## [1.1.3] - 2026-11-10

### Bug Fixes and Performance

- grant artifact-read permissions to shared publisher ([1bdac98](https://github.com/ThanhNguyxnOrg/lagvex/commit/1bdac984e043fe66b864d73fcab01c2726dcd6a4))

## [1.4.0] - 2026-11-16

### Features

- rank relays by measured per-segment RTT (client + relay to region) ([da710bf](https://github.com/ThanhNguyxnOrg/lagvex/commit/da710bfcce8a5078a8f4bd087ed97a2225cc0884))
- opt-in full tunnel mode (all IPv4 traffic through the relay) ([d94ddb5](https://github.com/ThanhNguyxnOrg/lagvex/commit/d94ddb5b3e07d3aee0beffb7ab8bde5850f7790c))
- recent-relay shortcut and best-route badge from measured sweeps ([07db59c](https://github.com/ThanhNguyxnOrg/lagvex/commit/07db59c))

## [1.5.0] - 2026-11-16

### Features

- stability-aware relay ranking and session failover history ([1f15aa7](https://github.com/ThanhNguyxnOrg/lagvex/commit/1f15aa750bcabe215ae2c617104a663cdfbd59e9))

## [1.6.0] - 2026-11-16

### Features

- optional unprivileged ICMP echo measurement for relay probes ([4d253c4](https://github.com/ThanhNguyxnOrg/lagvex/commit/4d253c4470bd9a3936fabc4e5832fc677763ad8b))
- pre-measured standby relay for instant failover ([7f97100](https://github.com/ThanhNguyxnOrg/lagvex/commit/7f97100a5d711d1896b88e22c04787db034a2b3c))

## [1.7.0] - 2026-11-17

### Features

- sort relay nodes by measured route, loss or recency ([c172d71](https://github.com/ThanhNguyxnOrg/lagvex/commit/c172d717f8156981211cceaa2de63d1576a54351))
- block cross-site requests to the loopback dashboard API ([c5f0fb0](https://github.com/ThanhNguyxnOrg/lagvex/commit/c5f0fb042c1f393c3b12f24d48027650171089f2))
- add a bounded TCP ping tool for game servers ([6c74955](https://github.com/ThanhNguyxnOrg/lagvex/commit/6c74955f03d866ef729d379a9338900a55140f3b))
- show per-sample RTT details on sparkline hover ([87022d6](https://github.com/ThanhNguyxnOrg/lagvex/commit/87022d609018777578060c2c964758846b742dcc))
- prefill TCP ping from observed game-server endpoints ([4de2562](https://github.com/ThanhNguyxnOrg/lagvex/commit/4de2562498f00b9b144f64f50348f083761be742))

### Bug Fixes and Performance

- check listener close error in TCP ping test ([9926ccc](https://github.com/ThanhNguyxnOrg/lagvex/commit/9926ccc04b127ddadc81f98ce2a0315e4b7c54fa))

## [1.8.0] - 2026-11-17

### Features

- dual-path dedup readiness (Redundant-lite Phase 1) ([3b54e8e](https://github.com/ThanhNguyxnOrg/lagvex/commit/3b54e8ecde73fa8841e95638b4cb17e4699aede0))
