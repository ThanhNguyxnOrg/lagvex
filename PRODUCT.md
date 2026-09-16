# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Competitive esports gamers playing latency-sensitive multiplayer games (Valorant, Counter-Strike 2, Apex Legends, PUBG: BATTLEGROUNDS, League of Legends, Dota 2, Rainbow Six Siege, Overwatch 2) seeking lower round-trip time (RTT), jitter stabilization, and zero packet loss.

## Product Purpose

Lagvex is an open-source, wire-speed Layer-3 game traffic accelerator and ping reducer. It creates an authenticated ChaCha20-Poly1305 AEAD virtual tunnel over UDP, selectively diverting only game server CIDR ranges while keeping regular Internet traffic on the local ISP.

## Positioning

Unlike commercial closed-source boosters that obscure network internals behind black-box marketing, Lagvex provides open-source verifiable routing, zero background bloat, authentic multi-sample RTT probing, and mathematically proven ChaCha20-Poly1305 encryption with anti-replay guarantees.

## Operating Context

Desktop gaming environment (Windows, macOS, Linux). The dashboard runs locally on `127.0.0.1:18888` as a companion HUD while fullscreen games run. It must be effortless to operate before a competitive match: select game, select region, see live probed nodes with the optimal node pre-selected, and engage with a single action.

## Capabilities and Constraints

- High-performance Layer-3 virtual adapter (WinTun on Windows, utun on macOS, /dev/net/tun on Linux).
- Authenticated ChaCha20-Poly1305 AEAD wire protocol with RFC 6479 anti-replay sliding window.
- Multi-sample latency prober calculating median RTT, jitter, packet loss percentage, and composite gaming scores.
- 14 global community relays (North America, Europe, Asia-Pacific, South America).
- In-memory process watcher (`CreateToolhelp32Snapshot`) for instant, anti-cheat safe game detection.
- Windows native IP Helper API route management and IPv6 leak prevention.

## Brand Commitments

- **Brand Name**: Lagvex
- **Brand Logo**: The official crystal-cut angular "L" emblem featuring a high-voltage lightning bolt and an oscilloscope telemetry pulse line, executed in electric cyan (`#00F0FF`), electric violet (`#7E42FF` / `#A366FF`), and deep obsidian (`assets/logo.png`).
- **Aesthetic Direction**: Razor-sharp, calm-dark esports precision. No cheap AI gimmicks, no oversized cheesy radar sweeps, no ALL-CAPS screaming. High-contrast readability, tabular numerals for micro-measurements, and authentic game logos/artwork.

## Evidence on Hand

- Official brand logo assets: `assets/logo.png`, `assets/logo.jpg`, `assets/banner.jpg`, `assets/icon.ico`.
- Verified binary protocol and crypto engine in Go.
- Real-time REST endpoints: `/api/status`, `/api/games`, `/api/relays`, `/api/probe-relays`, `/api/best-relay`, `/api/connect`, `/api/disconnect`, `/api/test-relay`.

## Product Principles

1. **Evidence Over Marketing**: Display real probed millisecond metrics and loss percentages; never fabricate fake ping drops.
2. **Frictionless Activation**: A gamer reaches an active, optimal boost within 2 clicks from app launch.
3. **Respect System Resources**: Minimal CPU/memory consumption; zero interference with game anti-cheat engines.
4. **Visual Authority**: Match the refined, precise aesthetic of elite competitive gaming gear and modern engineering tools.
