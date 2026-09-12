# 📝 Changelog 📜

> 🔗 [Back to Project README.md](README.md) | [Contributor Guide](CONTRIBUTING.md)

All notable changes to the **Lagvex** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-09-12 🚀

### 🔥 Added
- 🤝 **Comprehensive Contributor & Maintainer Guide (`CONTRIBUTING.md`)**:
  - Detailed handbook for open-source contributors and maintainers covering how to capture live game traffic (Resmon, Wireshark, PowerShell), identify ASNs, narrow CIDR subnets, register profiles in `configs/profiles.json`, craft vector SVG insignia in `web/app.js`, and validate with Go tests.
- 🎨 **Esports Cockpit HUD Redesign (`web/`)**:
  - Complete front-end overhaul inspired by professional esports command centers and tactical gaming HUDs.
  - High-contrast cyber dark aesthetics (`#0a0e17`) with electric cyan accents (`#00f0ff`), neon emerald indicators (`#00ff88`), warning amber (`#ffaa00`), and hot magenta alerts (`#ff0055`).
  - Ultra-crisp typography using Google Fonts *Outfit* (display) and *JetBrains Mono* (telemetry / code).
- 🛡️ **Authentic Vector SVG Game Insignias**:
  - Handcrafted, resolution-independent SVG vector emblems for all 11 supported titles:
    - 🎯 **Valorant**: Radiant sharp V emblem with cyber neon drop-shadow.
    - 💣 **Counter-Strike 2**: Defuse kit / tactical crosshair strike icon.
    - 🪂 **PUBG: BATTLEGROUNDS**: Spetsnaz Level 3 helmet & parachute dropship silhouette.
    - ⚡ **Apex Legends**: Stylized Apex predator chevron glyph.
    - ⚔️ **League of Legends & TFT**: Winged summoner sword crest.
    - 🎖️ **Call of Duty: Warzone**: Military skull crest with tactical reticle.
    - 🦅 **Delta Force: Hawk Ops**: High-speed tactical stealth chevron.
    - 🛡️ **Overwatch 2**: Blizzard defense matrix twin-ring insignia.
    - 🧱 **Rainbow Six Siege**: Breaching hammer & tactical shield.
    - 🏆 **The Finals**: Holographic cashout cube / arena glyph.
    - 🛡️ **Dota 2**: Ancient Aegis of Champions slab.
  - Full removal of generic emoji buttons or low-resolution bitmap graphics.
- 📡 **Dual-Ring Tactical Radar Scanner**:
  - Animated 360-degree radar sweep line with expanding radial sonar waves triggered during active packet tunneling and game process scanning.
- ⚡ **Live Laser Packet Stream Animation**:
  - Dynamic neon laser path along the route connector with traveling particle blips, visually reflecting real-time UDP game packet routing.
- 🔔 **Floating Glassmorphic Toast Notification System**:
  - Sleek pill notifications with backdrop blur, color-coded icon badges, and smooth sliding entrance/exit animations for ping probes, clipboard copy events, and tunnel status alerts.
- 🖼️ **Translucent Squircle App Icon & Favicon**:
  - Clean transparent SVG/PNG emblem with neon cyan glow, eliminating boxy borders on high-DPI desktop screens and browser tabs.
- 🌐 **Global 5-Continent Relay Network**:
  - Built-in profiles now ship with **14 Worldwide VPS Relay Endpoints** covering all competitive esports zones:
    - 🌏 **Asia-Pacific (APAC)**: Singapore #1, Singapore #2, Tokyo (Japan), Hong Kong, Seoul (South Korea), Sydney (Australia).
    - 🌍 **Europe (EU)**: Frankfurt (Germany), London (United Kingdom), Paris (France), Helsinki (Finland).
    - 🌎 **North America (NA)**: US-East (N. Virginia), US-West (Oregon).
    - 🌎 **South America (SA)**: São Paulo (Brazil).
    - 🌍 **Middle East & Africa (MENA)**: Bahrain / Dubai.
- 🎮 **105 Global Target Server Clusters**:
  - Expanded game routing tables to **105 regional server clusters** across all 11 games.
  - Organized neatly in the HUD dropdown using HTML `<optgroup>` continent separators for instant continent-by-continent server selection.
- 🪟 **1-Click Native Windows PowerShell Relay Installer (`scripts/install-relay.ps1`)**:
  - Complete PowerShell automation script for hosting a Lagvex relay on Windows 10/11 or Windows Server.
  - Automatically verifies Administrator elevation, configures WireGuard WinTun driver, enables `Set-NetIPInterface -Forwarding Enabled`, provisions NetNat port forwarding, creates 32-char PSK, and prints 1-click squad invite link (`lagvex://connect?...`).
- 🧪 **Sub-Minute Quality CI/CD Pipeline (`.github/workflows/ci.yml`)**:
  - Fast GitHub Actions workflow running `golangci/golangci-lint-action@v6` with comprehensive lint rules (`govet`, `errcheck`, `staticcheck`, `unused`, `gosec`, `revive`) and matrix unit tests on every commit/PR in under 60 seconds.
- 🚀 **Automated Release Trigger (`.github/workflows/release.yml`)**:
  - Release workflow that automatically triggers on semantic tags (`v*`) or release commits (`release: x.x.x`), parses `CHANGELOG.md` for release notes, and builds multi-arch cross-platform binaries (Windows x64/ARM64, Linux amd64/arm64, macOS arm64/amd64).
- 🏷️ **Nominative Fair Use & Legal Transparency Notice**:
  - Explicit legal disclosure in `README.md` and `docs/DISCLAIMER.md` affirming that all game names and vector logos are used under Nominative Fair Use solely to identify game server routing targets and processes.

### 🔄 Changed
- 🗺️ **Profile Schema Upgrade (`configs/profiles.json`)**:
  - Added optional `continent` field (`apac`, `eu`, `na`, `sa`, `mena`) to region descriptors, allowing client frontends to group server clusters by geographical zone.
- 🔄 **Game ID Standardization**:
  - Normalized `cod_warzone` identifier with backwards compatibility for `warzone` in web frontend and test suites.

### 🛡️ Fixed
- 🐛 **WinTun Driver Unsafe Pointer Alignment**:
  - Fixed Go compiler vet warning in `pkg/client/wintun_windows.go` by properly casting `windows.StringToUTF16Ptr` to avoid unsafe pointer arithmetic. Zero `go vet` warnings across entire codebase.

---

## [1.0.0] - 2026-09-12 🚀

### 🔥 Added
- ⚡ **Cross-Platform Client Engine**: Written in 100% pure Go supporting:
  - 🪟 **Windows 10/11**: Official WireGuard `wintun.dll` Layer-3 ring buffer with RAM-only `store=active` routing.
  - 🐧 **Linux & Steam Deck**: Native Linux kernel `/dev/net/tun` with `ip route` and `CAP_NET_ADMIN` support.
  - 🍎 **macOS (Apple Silicon & Intel)**: Native Darwin `utun` interface and BSD route tables.
- 🎨 **Standalone Desktop App Window**: Built-in launcher supporting Microsoft Edge / Google Chrome App Mode (`1220x840`) with embedded high-resolution `.ico` application icon.
- 🎮 **11 Competitive Game Profiles**: Valorant, CS2, PUBG, Apex Legends, The Finals, COD Warzone, Delta Force, Overwatch 2, Rainbow Six Siege, League of Legends & TFT, and Dota 2.
- 🌐 **Ultra-Lightweight Go Relay Daemon**: Single-binary Linux daemon with Linux kernel IP masquerade NAT, TCP MSS clamping, and HMAC-SHA256 authenticated handshake.
- ⚡ **Wire Protocol v1**: 9-byte minimal data header with microsecond-level packet serialization and zero framing tax.
- 🚀 **1-Command Relay Deployment**: Automated installer script (`scripts/install-relay.sh`) with systemd registration, sysctl tuning, and persistent iptables rules.
- 📊 **Cyberpunk HUD Web Dashboard**: Dark Neon HUD with real-time ping chart, dynamic game filtering, custom game creator, and relay connection manager.
- ⚖️ **Academic Documentation & Legal Waivers**: Comprehensive architectural specs, research whitepaper, vendor-neutral relay FAQ, and zero-liability ban disclaimers.
