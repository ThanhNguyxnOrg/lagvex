# ⚡ Lagvex

<p align="center">
  <img src="assets/banner.jpg" alt="Lagvex Banner" width="100%">
</p>

<p align="center">
  <strong>Open-Source Gaming Latency Reducer & Split-Tunneling Ping Booster</strong><br>
  <em>Achieve ultra-low latency, eliminate jitter, and bypass ISP routing bottlenecks with 100% anti-cheat-safe route-based split-tunneling.</em>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="https://github.com/ThanhNguyxnOrg/lagvex"><img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-blue" alt="Platform"></a>
  <a href="https://github.com/ThanhNguyxnOrg/lagvex"><img src="https://img.shields.io/badge/Anti--Cheat-100%25%20Safe%20(No%20Hooks)-brightgreen" alt="Anti-Cheat Safe"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

---

## 📖 Table of Contents

- [Why Lagvex?](#-why-lagvex)
- [Supported Games](#-supported-games)
- [Documentation & Deep Dives](#-documentation--deep-dives)
- [How It Works](#-how-it-works)
- [Quick Start](#-quick-start)
  - [1. Deploy Relay on VPS (One-Liner)](#1-deploy-relay-on-vps-one-liner)
  - [2. Run Client on Windows](#2-run-client-on-windows)
- [Building from Source](#-building-from-source)
- [License](#-license)

---

## 🎯 Why Lagvex?

When playing online multiplayer games (such as Valorant, Counter-Strike 2, PUBG, or Apex Legends), domestic ISPs often route game traffic through suboptimal international transit corridors, causing **high ping, rubberbanding, packet loss, and jitter**.

Traditional workarounds fail in critical ways:
- **Commercial VPNs**: Force 100% of all system traffic through the tunnel. Discord voice regions break, web browsing slows down, and background downloads choke game packets.
- **Hook-based Game Boosters**: Use DLL injection, LSP, or packet-divert drivers (e.g. WinDivert). Modern kernel anti-cheats (**Riot Vanguard, Valve VAC, BattlEye, Easy Anti-Cheat, RICOCHET**) continuously detect and ban third-party socket injection.

**Lagvex solves this cleanly**:
1. **100% Anti-Cheat Safe**: Operates strictly at the OS routing table level using WireGuard's official **WinTun** driver. No memory reading, no DLL injection, and no socket tampering.
2. **Selective Route-Based Split Tunneling**: Only traffic destined for game server IP ranges passes through the tunnel. Discord, browsers, streaming, and Windows updates stay on your physical network at full gigabit speed.
3. **High-Performance Go Relay**: Lightweight server running on any Linux VPS. Features a 9-byte minimal packet header over UDP and hands raw IPv4 packets to the Linux kernel for zero-copy NAT MASQUERADE and MSS clamping.
4. **Out-of-the-Box Multi-Game Support**: Pre-configured with verified server CIDR pools for major competitive shooters and popular online titles.

---

## 🎮 Supported Games

| Game Title | Genre | Server Regions | Anti-Cheat Compatibility |
|---|---|---|---|
| **Valorant** | Tactical FPS | Singapore (SEA), Tokyo (JP), Hong Kong (HK), Mumbai (IN), Frankfurt | ✅ Riot Vanguard |
| **Counter-Strike 2 (CS2)** | Tactical FPS | Singapore, Hong Kong, Tokyo, Seoul, Frankfurt (Valve SDR) | ✅ Valve Anti-Cheat (VAC) |
| **PUBG: BATTLEGROUNDS** | Battle Royale | Singapore, Tokyo, Seoul, Frankfurt (Azure & AWS) | ✅ BattlEye + Zakynthos |
| **Apex Legends** | Battle Royale | Singapore, Tokyo, Taiwan, Oregon (EA Multiplay) | ✅ Easy Anti-Cheat |
| **The Finals** | Arena FPS | Singapore, Tokyo, Frankfurt | ✅ Easy Anti-Cheat |
| **Call of Duty: Warzone / MW3** | FPS / BR | Singapore, Tokyo, US-West (Demonware) | ✅ RICOCHET Anti-Cheat |
| **Delta Force: Hawk Ops** | Tactical Shooter | Singapore, Hong Kong (Tencent Cloud / AWS) | ✅ ACE Anti-Cheat |
| **Overwatch 2** | Hero Shooter | Singapore, Taiwan, Korea, Japan | ✅ Blizzard Defense Matrix |
| **Rainbow Six Siege** | Tactical Shooter | Singapore (SEAU), Japan East (Ubisoft Azure) | ✅ BattlEye |
| **League of Legends (LoL)** | MOBA | Vietnam (VNG), Singapore (Riot Direct), Taiwan | ✅ Riot Vanguard |
| **Dota 2** | MOBA | Singapore (SEA), Japan (Valve SDR) | ✅ Valve VAC |

*Users can also register custom games and private server CIDRs dynamically via the Web Dashboard.*

---

## 📚 Documentation & Deep Dives

| Document | Description |
|---|---|
| 📐 [**Architecture Overview**](docs/ARCHITECTURE.md) | In-depth packet journey, WinTun driver mechanics, Windows routing safety rules, and anti-cheat analysis. |
| 📡 [**Wire Protocol v1 Specification**](docs/PROTOCOL.md) | Binary packet format, HMAC-SHA256 handshake, 9-byte data header, keepalive ping/pong, and MTU arithmetic. |
| 🚀 [**VPS Deployment Guide**](docs/DEPLOYMENT.md) | Step-by-step VPS operator instructions, automated one-liner script, systemd configuration, and Docker Compose. |
| 🌐 [**Game Profiles & CIDRs**](docs/PROFILES.md) | Profile JSON schema, cloud provider network maps (Valve SDR, Riot Direct, AWS, Azure), and capturing new games. |

---

## 🏗️ How It Works

```
[ Game: Valorant / CS2 / PUBG ]
       │  Sends UDP/TCP packets to game server (e.g. 13.250.0.0/15)
       ▼
[ Windows Routing Table ]
       │  Game Server CIDR matches on-link ──> [ WinTun Adapter: Lagvex ] (10.88.0.2)
       │  All other traffic (Discord, Web) ──> [ Default Physical Gateway ] (Direct ISP)
       ▼
[ Lagvex Client Engine (Windows) ]
       │  Reads raw IPv4 packet from WinTun ring buffer
       │  Encloses with 9-byte header (Version, Type, SessionID)
       │  Sends over UDP to Relay VPS (Pinned /32 route guarantees physical path)
       ▼
═════════════════════ International Fiber / VPS ═════════════════════
       ▼
[ Lagvex Relay Server (Go on Linux VPS) ]
       │  Strips 9-byte header, verifies HMAC session, checks anti-spoofing
       │  Writes raw IPv4 packet into Linux TUN device (/dev/net/tun: lagvex0)
       ▼
[ Linux Kernel (VPS) ]
       │  iptables NAT MASQUERADE + TCP MSS Clamping
       ▼
[ Destination Game Server (Singapore / Tokyo / Hong Kong...) ]
```

---

## 🚀 Quick Start

### 1. Deploy Relay on VPS (One-Liner)

Deploy on any Linux VPS (Ubuntu, Debian, CentOS, AlmaLinux, Arch) near your target game servers (e.g., Singapore or Tokyo):

```bash
curl -fsSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash
```

The script automatically:
1. Enables IP forwarding and optimizes kernel UDP buffers in `sysctl`.
2. Sets up `iptables` NAT MASQUERADE and TCP MSS clamping that persist across reboots.
3. Generates a secure random 32-character Pre-Shared Key (PSK).
4. Configures and starts the `lagvex-relay.service` systemd daemon.
5. Prints a client-ready configuration snippet.

*(Or run via container using `docker compose up -d`)*

---

### 2. Run Client on Windows

1. Download the latest release containing `lagvex-client.exe`, `web/`, and `configs/`.
2. Ensure `wintun.dll` (64-bit) is present alongside the executable or in `bin/amd64/`.
3. Launch `lagvex-client.exe` as **Administrator** (required by Windows to create virtual network adapters and insert `store=active` routes):

```cmd
# Launch with Web Dashboard UI (opens http://127.0.0.1:18888 automatically):
lagvex-client.exe

# Or connect directly via CLI:
lagvex-client.exe -connect -relay 123.45.67.89:51820 -psk your_secret_psk -game valorant -region asia-sg
```

---

## 🛠️ Building from Source

Requires **Go 1.22+**:

```bash
# Clone the repository
git clone https://github.com/ThanhNguyxnOrg/lagvex.git
cd lagvex

# Run unit tests
make test

# Build Linux Relay binary
make relay

# Build Windows Client executable
make client
```

---

## 📜 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for more information.
