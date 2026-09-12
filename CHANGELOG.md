# 📝 Changelog 📜

All notable changes to the **Lagvex** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
