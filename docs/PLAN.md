# 📋 Lagvex Product & Engineering Execution Plan 🚀

> 💡 **Related Documents**: [📖 Back to README](../README.md) • [🗺️ Engineering Roadmap](ROADMAP.md) • [📐 Architecture](ARCHITECTURE.md) • [🎨 UI Design Direction](#-phase-2--ui-redesign-anti-ai-look) • [⚖️ Disclaimer](DISCLAIMER.md)

---

## 🎯 Purpose of This Document

This is the **master execution plan** that turns Lagvex from a working MVP / developer tool into a
polished consumer-grade app comparable in usability to ExitLag or GearUP Booster, while paying down
the technical debt that blocks public relay sharing.

It supersedes the *execution ordering* of [ROADMAP.md](ROADMAP.md) (which remains the technical
deep-dive for each phase). The key re-prioritization: **Multi-Relay selection moves from P3 → Phase 1**
because automatic node selection is a *user-experience requirement*, not an optimization.

### Inputs synthesized into this plan

1. **Code audit** of `pkg/client`, `pkg/relay`, `pkg/protocol` (see Phase 0 findings).
2. **Lessons from the predecessor project** (`GamePingBooster`, C#/Avalonia + Go relay):
   its own README documents the UX failure modes — manual `gpb.conf` + SSH deploy + PSK copy into
   `config.json` + LocalSystem service debugging + unsigned-installer SmartScreen warnings.
   Gamers churn before step 2.
3. **Commercial booster workflow research** (ExitLag, GearUP Booster, LagoFast, Hone, NoPing, Mudfish):
   the winning flow is exactly three steps — *select game → select region/node (auto-tested) → press Boost*.
   Infrastructure is never visible to the end user.
4. **Open-source ecosystem research**: [wireguard-go](https://github.com/WireGuard/wireguard-go) embedding,
   [udp2raw](https://github.com/wangyu-/udp2raw) / [UDPspeeder](https://github.com/wangyu-/udpspeeder) FEC techniques,
   the IPv6 leak class-bug documented across Mudfish/ExitLag deployments.

---

## 🩺 Current State Assessment

| Area | Status | Blocking Issue |
|---|---|---|
| Tunnel data plane | ✅ Works | ❌ Unencrypted, unauthenticated (session hijack / Disconnect spoof DoS via sniffed SessionID) |
| Windows routing | ✅ Works | ❌ `netsh` / `route print` **text parsing breaks on non-English Windows** |
| IPv6 | ❌ Missing | Game traffic over IPv6 **bypasses the tunnel entirely** (silent booster failure) |
| Relay UX | ⚠️ Dev-oriented | Rent VPS → one-liner → copy PSK into config: acceptable for self-hosters, fatal for normal gamers |
| Node selection | ❌ Manual | User must know/ping relays themselves; no auto-probe, no recommendation |
| UI | ⚠️ Feature-complete, taste-poor | "AI-generated" aesthetic: neon-on-neon, radar sweeps, ALL CAPS, 3 font families |
| IP pool lifecycle | ❌ Bug | `Allocate()` runs *before* old-session cleanup → pool exhaustion on reconnect storms |
| Connection health | ❌ Missing | Relay death leaves tunnel stuck in "connected" forever; no pong-timeout, no auto-reconnect |

---

## 🔵 Phase 0 — Bug Hops & Architecture Decision *(effort: 1–2 sessions)*

Small, independent fixes that unblock everything else.

### 0.1 IP-pool exhaustion on reconnect
- **Where**: [`pkg/relay/server.go`](../pkg/relay/server.go) `handleHandshake()` — `pool.Allocate()` (≈line 191)
  runs **before** the same-`ClientID` cleanup loop (≈line 202).
- **Fix**: move the cleanup loop *before* allocation. Verify with a reconnect-storm unit test in `pool_test.go`.

### 0.2 Remove hardcoded dev path
- **Where**: `pkg/client/wintun_windows.go` line ≈46 — `D:\Code\Lagvex\bin\amd64\wintun.dll` candidate list.
- **Fix**: delete the absolute path candidate; keep exe-relative + `bin/amd64` relative lookups.

### 0.3 Client accepts UDP from any source
- **Where**: `pkg/client/engine.go` `pumpUDPToWinTun()` — never checks `ReadFromUDP` source == relay addr.
- **Fix**: drop packets whose source ≠ current `relayAddr`. Cheap defense-in-depth until AEAD lands.

### 0.4 Dead-relay detection
- **Where**: `loopKeepalive()` sends pings but nothing tracks pongs.
- **Fix**: if no valid Pong for N seconds (suggest N=10, i.e. 5 missed keepalives):
  set state → `error`, tear down routes, surface a toast/inline error, offer one-click reconnect.

### 0.5 THE architecture decision: custom AEAD vs embedded wireguard-go
| Option | Pros | Cons |
|---|---|---|
| **A. Keep custom protocol, implement ChaCha20-Poly1305 AEAD** (ROADMAP P0 spec) | Full learning/portfolio value; PSK model preserved; protocol stays tiny | Must implement session-key derivation, 64-bit counter nonces, RFC 6479 replay window, authenticated roaming — all attackable surface |
| **B. Embed [wireguard-go](https://github.com/WireGuard/wireguard-go) as a library** ([official embedding guide](https://www.wireguard.com/embedding/)) | Battle-tested AEAD + replay + roaming for free (the Tailscale approach); Phase 3 mostly evaporates | Loses "own wire protocol" story; key model becomes public-key per client; bigger dependency |
- **Recommendation**: **Option A** — the wire protocol is core to this project's identity; implement AEAD in Phase 3 with the spec already written in ROADMAP P0. Revisit only if Phase 3 stalls.

---

## 🟢 Phase 1 — "Real App" Workflow *(effort: the milestone that matters)*

Goal: **a gamer reaches an active boost without ever seeing the words "VPS", "SSH", or "PSK".**

### Target user flow (the spec)

```text
Launch app ──► Game grid (search + cover art) ──► Click game
   ──► Region & node screen: live RTT probe of every relay, best node pre-selected ✔
   ──► BIG "BOOST" button ──► connecting animation ──► Active HUD (live ping, before/after)
   ──► (optional) "Launch Game" button ──► process watcher auto-manages routes (already built)
```

### 1.1 Bundled public relay directory
- Ship a curated `relays.json` (fetched + cached from GitHub raw at startup, offline fallback bundled).
- Maintainer + community-run relays register via PR (`endpoint`, `region`, `continent`, `capacity`, `PSK tier`).
- **First-run works out of the box** — the app already knows usable relays before the user configures anything.
- Seed with 1–2 maintainer-hosted relays (SG + HK) as proof.

### 1.2 Auto node selection *(pulled forward from ROADMAP P3)*
- On game+region selection, concurrently probe all candidate relays (UDP handshake-RTT, 5 probes each, median).
- Sort by composite score `RTT_med + jitter + loss penalty`; pre-select the winner with a ✓; user can override.
- This reuses the relay-probe loop later for hysteresis failover (ROADMAP P3) — build it as `pkg/client/prober.go` once.

### 1.3 Self-host demoted to "Advanced" + deploy wizard
- VPS settings move behind an *Advanced* section (current UI flow stays available for power users).
- **API-key wizard**: user pastes a DigitalOcean / Vultr / Hetzner API key → app creates the cheapest droplet
  in the chosen region → runs the existing one-liner installer via cloud-init → writes endpoint + PSK back
  into local config → relay appears in the node list. Zero SSH for the user.
- Keep the manual one-liner flow documented ([DEPLOYMENT.md](DEPLOYMENT.md)) as the fallback path.

### 1.4 First-run wizard
- Single admin-elevation prompt; silent WinTun driver placement; language pick; done in ≤ 2 screens.
- Unsigned-binary trust problem (learned from GamePingBooster SmartScreen complaints) → addressed by
  code signing in release pipeline, tracked in Phase 4 ops.

---

## 🟡 Phase 2 — UI Redesign ("anti-AI-look")

The current HUD is feature-complete but over-decorated. The reference class (ExitLag / GearUP) is
**calm dark, one accent, real game artwork**. Redesign rule: **subtract, don't add.**

### Remove
- Mesh glows, grid texture overlays, radar sweep, laser packet stream, glassmorphic toasts
- ALL-CAPS labels everywhere; the "PRO" badge; cyberpunk status chips in the header
- Three simultaneous font families; JetBrains Mono as the default UI voice

### Adopt
- Neutral deep-dark base (near-black, *not* blue-purple tint), **exactly one accent color**
- **Real game cover art** as the visual anchor of the grid/hero (store-sourced or commissioned stylized art)
- Flat cards with 1px subtle borders, normal-case typography, hierarchy via size/weight
- Motion only where it conveys state: connecting spinner, boost on/off, ping delta count-up
- Tabular numerals for all telemetry; one chart style for ping history

### Screens to design (in order)
1. Game grid (home) — search, cover tiles, "recently boosted" row
2. Game detail + node picker — region groups, live RTT per node, ✓ recommendation
3. Active boost HUD — before/after ping, up/down rate, session info, STOP button
4. Settings (incl. demoted Advanced/self-host)

**Process note**: produce 2–3 mockup variants (HTML stills) for review *before* rewriting `web/`.

---

## 🔴 Phase 3 — Security Hardening *(hard gate before any public/shared relay)*

No community relay goes live until all four land:

1. **AEAD data plane** — ChaCha20-Poly1305, session keys derived during the HMAC handshake,
   64-bit monotonic counter nonce per direction (spec: [ROADMAP P0.1](ROADMAP.md)).
2. **Authenticated control messages** — Disconnect / remote-update (roaming) only after tag verification
   → kills both the sniffed-SessionID DoS and the endpoint-hijack vector.
3. **Per-client credentials + quotas** — replace the single shared PSK with per-client keys;
   enforce per-session bandwidth caps + connection limits in the relay (`Config.MaxClients` exists,
   add `MaxMbpsPerSession`). Otherwise a public relay is a free-transit farm.
4. **Anti-replay** — server-side sliding-window nonce cache for handshakes (120 s TTL),
   RFC 6479 window on data counters.

Also in this phase: **IPv6 leak fix** — disable IPv6 binding on the Lagvex WinTun adapter at connect
(and restore at disconnect), the documented fix Mudfish shipped for the same class of bug.

---

## ⚪ Phase 4 — Trust, Telemetry & Ops

- **Smart Route Advisor** (ROADMAP P2.2): simultaneous direct vs relay RTT; if direct wins, say so
  (*"Your direct path is optimal — booster recommended: OFF"*). Nobody does this; it is the strongest
  credibility feature an open-source booster can ship.
- **P50/P95/P99 + jitter HUD** (rolling 120-packet window).
- **Learn-mode route discovery**: observe actual packet destinations in the WinTun ring during play,
  propose/add routes dynamically → retires the hand-maintained CIDR grind (also fixes stale-IP drift).
- **Native `iphlpapi` routing** (ROADMAP P1.1): `GetBestRoute2` / `CreateIpForwardEntry2` via
  `golang.org/x/sys/windows`, deleting all `route print` / `netsh` text parsing.
- **Release ops**: code-signing certificate in the release workflow (SmartScreen), crash/ping telemetry (opt-in).
- Later research slots: adaptive FEC (UDPspeeder-style, ROADMAP P4), multipath duplication, fakeTCP port
  for ISP UDP throttling (udp2raw-style).

---

## 📅 Execution Matrix

| # | Phase | Deliverable | Effort | Unblocks |
|---|---|---|---|---|
| 0 🔵 | Bug hops + decision | Pool fix, path fix, source check, pong-timeout, AEAD-vs-wireguard decision | 1–2 sessions | All |
| 1 🟢 | App workflow | Relay directory, auto node select, deploy wizard, first-run | Core milestone | Public launch |
| 2 🟡 | UI redesign | 4 screens rebuilt calm-dark, mockups approved first | Parallel with 1 | Screenshots/README |
| 3 🔴 | Security | AEAD, auth control msgs, per-client keys+quotas, replay guard, IPv6 fix | Must finish pre-public-relay | Community relays |
| 4 ⚪ | Trust & ops | Advisor, P95 HUD, learn-mode, iphlpapi, signing | Ongoing | Credibility, scale |

**Definition of "complete app" milestone**: Phase 0 + 1.1 + 1.2 + 2 → a new user goes from download to
active boost in under 2 minutes with zero configuration.

---

## 🔗 Research References

- [wireguard-go](https://github.com/WireGuard/wireguard-go) • [WireGuard embedding guide](https://www.wireguard.com/embedding/) • [tun package API](https://pkg.go.dev/golang.zx2c4.com/wireguard/tun)
- [udp2raw](https://github.com/wangyu-/udp2raw) • [UDPspeeder](https://github.com/wangyu-/udpspeeder) • [kcptun](https://github.com/xtaci/kcptun) • [KCP](https://github.com/skywind3000/kcp)
- [GearUP Booster user guide (3-step flow)](https://www.gearupbooster.com/support/mobile-user-guide.html) • [GearUP review ("no technical knowledge required")](https://pixelblockbuster.com/gearup-booster-review/) • [ExitLag — how it works](https://www.exitlag.com/blog/how-exitlag-works/)
- [Mudfish IPv6 leak docs](https://docs.mudfish.net/v3/en/docs/mudfish-product/mudfish-cloud-vpn/release-notes/) • [Mudfish custom route forum](http://mudfish.net/forums/6/topics/41964)
- [Microsoft NIC performance tuning](https://learn.microsoft.com/en-us/windows-server/networking/technologies/network-subsystem/net-sub-performance-tuning-nics) • [SpeedGuide gaming tweaks](https://www.speedguide.net/articles/gaming-tweaks-5812)

---

[⬅️ Back to README](../README.md) • [🗺️ Engineering Roadmap](ROADMAP.md) • [📐 Architecture](ARCHITECTURE.md) • [📡 Protocol Spec](PROTOCOL.md)
