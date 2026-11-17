# ⚡ Lagvex 🚀

<p align="center">
  <img src="assets/logo.png" alt="Lagvex logo" width="140" />
</p>

<p align="center">
  <strong>Open-source split-tunnel game latency accelerator</strong><br>
  <em>🧭 Route-based, not hook-based • 📊 Measured telemetry only • 🛠️ Self-hosted relays</em>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="https://github.com/ThanhNguyxnOrg/lagvex/actions/workflows/ci.yml"><img src="https://github.com/ThanhNguyxnOrg/lagvex/actions/workflows/ci.yml/badge.svg" alt="CI Status"></a>
  <a href="https://github.com/ThanhNguyxnOrg/lagvex"><img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue" alt="Platform"></a>
  <a href="docs/DISCLAIMER.md"><img src="https://img.shields.io/badge/Anti--Cheat-Non--Invasive%20(No%20Hooks)-brightgreen" alt="Anti-Cheat: Non-Invasive"></a>
  <a href="CHANGELOG.md"><img src="https://img.shields.io/badge/Changelog-v1.2.0-blueviolet" alt="Changelog"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

---

## 🎯 What Lagvex Is

Lagvex is a free, MIT-licensed alternative to commercial game boosters. It puts only
your game's traffic on a faster route and leaves everything else on your normal
internet connection:

1. You deploy a small **relay** on a Linux VPS close to the game server (or run the
   embedded relay locally on Linux for a loopback test).
2. The client watches your game's server region **CIDR ranges** (and, optionally, the
   exact peers your game process talks to) and installs routes for them.
3. Those packets travel through an authenticated **ChaCha20-Poly1305 encrypted UDP
   tunnel** to the relay, which forwards them to the game server.

Lagvex touches **no game memory, injects no DLLs and hooks no sockets**. It works at
the operating system routing layer, which is why it does not trigger kernel-level
anti-cheat checks the way socket-hooking boosters can.

> **Honesty first:** Lagvex only displays values it actually measured (probe RTT,
> tunnel echo RTT, packet loss, process findings). Everything it cannot measure is
> shown as `—` with the reason, never as an invented number.

## ✨ Features

- 🧭 **CIDR split tunneling** — only game server ranges enter the tunnel; browsers,
  Discord and updates keep using your local ISP path.
- 🎛️ **Peer routing (opt-in)** — reads the OS socket tables, routes `/32` routes for
  the public IPs your game process is actually talking to (Linux/macOS observe UDP and
  TCP; Windows exposes TCP peers only). Off by default.
- 📊 **Measured node list** — per-relay RTT history: latest, average, standard
  deviation, loss and a sparkline built from real probe runs. Hover a sparkline
  point to see its sample number, RTT and change from the previous sample. When
  the OS permits unprivileged ICMP, an echo RTT (`ICMP x ms`) is shown next to
  the tunnel probe.
- 🗺️ **Route detail** — per-hop measurement (gateway, direct relay probe, live tunnel
  echo, region cloud anchor) with explicit "not measured" reasons for segments the
  client cannot observe.
- 🧪 **TCP ping tool** — measures the TCP handshake RTT (min/avg/max/loss) to an
  explicit public game-server IP. Hostnames and private, loopback or link-local
  addresses are refused, and runs are limited to one every 2 seconds and 60 per
  hour. A closed port still reports a real RTT when the host answers with RST.
- 🪟 **Floating HUD** — an always-on-top overlay that reads the same `/api/status`
  values and shows dashes until a sample exists.
- 🧾 **Diagnostics snapshot** — export measured state plus non-secret settings for bug
  reports; import merges them back (relay PSKs are never included).
- 🛡️ **Adaptive FEC, failover and DSCP marking** — loss recovery and relay failover
  with honest status reporting. While connected, one alternative relay is kept
  pre-measured as a standby, so a degradation promotes it without waiting for a sweep.
- 🖥️ **Windows, macOS and Linux client** — WinTun, native `utun` and `/dev/net/tun`
  respectively; one Go codebase, one React dashboard.

## 🚀 Quick Start

### 1. Deploy a relay (Linux VPS — required for real acceleration)

```bash
curl -fsSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash
```

The installer prints the relay endpoint, the UDP port (default `51820`) and a generated
**PSK**. Keep the PSK: the client refuses to connect without it. Relay hosting requires a
Linux VPS with `/dev/net/tun` (KVM, not OpenVZ). Windows and macOS cannot host relays; the
installer fails fast there instead of pretending.

### 2. Build and run the client

```bash
go build -o lagvex-client ./cmd/lagvex-client
./lagvex-client                 # opens the dashboard at http://127.0.0.1:18888
```

**Add node** → enter the relay `host:port` **and its PSK** → pick your game and server
region → **BOOST**. The dashboard shows probes and tunnel state as they are measured.

### 3. Squad rooms

Squad rooms share a **roster** (who is in the room plus each member's reported RTT) so a
party can see each other in one place. A room does not change anyone's relay or routes.
Rooms are relayed through the connected relay's signaling endpoint; with no relay
connected, a room exists only on that machine.

## 🖥️ Supported Systems

| OS | Virtual adapter | Notes |
|---|---|---|
| **Windows 10/11** | WinTun (`wintun.dll`) | Elevation needed to create the adapter |
| **macOS 12+** | native `utun` | DNS diagnostics read via `scutil` |
| **Linux / Steam Deck** | `/dev/net/tun` | Also the only OS that can host the embedded local relay |

## 🏗️ How It Works

```text
 Game process               Lagvex client                    Your relay (Linux VPS)         Game server
┌──────────────┐   routes  ┌──────────────────────┐  AEAD UDP  ┌──────────────────┐  UDP   ┌────────────┐
│ VALORANT.exe │ ────────▶ │ WinTun / utun / tun  │ ─────────▶ │ lagvex-relay     │ ─────▶ │ Riot / AWS │
└──────────────┘  (CIDR +  │ ChaCha20-Poly1305    │            │ /dev/net/tun      │        └────────────┘
                  peers)   │ split tunnel + FEC   │            └──────────────────┘
                           └──────────────────────┘
```

- The client connects to the relay with a PSK handshake, receives an inner IP and installs
  routes only for the selected game region and — if enabled — observed peers.
- The relay forwards packets between the tunnel and the internet, and hosts squad
  signaling plus health endpoints.
- See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the packet journey, and
  [`docs/PROTOCOL.md`](docs/PROTOCOL.md) for the wire format.

## 📚 Documentation

| Doc | Contents |
|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Component layout, routing design, anti-cheat safety |
| [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) | Relay VPS installation, hardening and operations |
| [`docs/PROFILES.md`](docs/PROFILES.md) | Game profile schema and how to add a game |
| [`docs/PROTOCOL.md`](docs/PROTOCOL.md) | Packet format, handshake and FEC |
| [`docs/WHY_VPS.md`](docs/WHY_VPS.md) | Why a relay is needed and what it can/cannot fix |
| [`docs/DISCLAIMER.md`](docs/DISCLAIMER.md) | Legal disclaimer and anti-cheat position |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Development setup, adding games, UI builds |
| [`CHANGELOG.md`](CHANGELOG.md) | Release notes |

## 🛠️ Building from Source

```bash
# Go backend (client + relay)
go build ./cmd/lagvex-client
go build ./cmd/lagvex-relay

# React dashboard (required after UI changes; outputs into web/)
npm --prefix ui_source install
npm --prefix ui_source run build

# Tests
go test ./...
npx --prefix ui_source tsc --noEmit
```

The dashboard is a React/Vite app in `ui_source/`; the build writes hashed bundles into
`web/assets/`, which the Go binary embeds. There is no separate legacy HTML/JS UI.

## ⚠️ Known Limits (read this)

- **No public relays are bundled.** The relay list starts empty; you bring your own
  endpoint and PSK. Catalog entries for dead or third-party nodes are deliberately gone.
- **Relay hosting is Linux-only.** The relay needs `/dev/net/tun`; on other systems the
  local relay refuses to start with a clear error.
- **Windows cannot see another process's UDP peers.** Peer routing on Windows works with
  TCP observations only; the UI says so. Game traffic is mostly UDP, so keep the region
  CIDRs enabled there.
- **The relay only probes allowlisted cloud anchors.** On request it measures its own
  RTT to a fixed set of public cloud endpoints (rate-limited per session); the
  "relay → game server" hop itself stays unmeasured and is reported as such rather
  than guessed.
- **No region/device spoofing, no DLL/LSP injection, no anti-cheat evasion.** Those are
  out of scope by design.

## 🤝 Contributing

- 🎮 Add a game profile: capture the destination CIDRs, edit `configs/profiles.json`,
  and run the tests. See [`CONTRIBUTING.md`](CONTRIBUTING.md).
- 🐛 Found a bug? Export a **diagnostics snapshot** from the Tweaker tab and attach it —
  it contains measured values and no secrets.
- 🧪 Run `go test ./...` and `npx --prefix ui_source tsc --noEmit` before opening a PR.

## ⚖️ Disclaimer

Lagvex is a network routing research tool. It does not read, modify or inject into any
game. Game publishers and anti-cheat vendors set their own rules about VPNs and proxies;
you remain responsible for how you use this software. Read the full text in
[`docs/DISCLAIMER.md`](docs/DISCLAIMER.md).

## 🏷️ Trademarks

All game titles and publisher names are trademarks of their respective owners. Lagvex is
not affiliated with or endorsed by any game developer or publisher. Game artwork bundled
in the dashboard is used nominatively to identify game executables and server regions.

## 📜 License

MIT — see [`LICENSE`](LICENSE).
