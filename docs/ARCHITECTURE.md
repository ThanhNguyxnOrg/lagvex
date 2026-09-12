# 📐 Lagvex Architecture & Network Design ⚡

> **High-Performance Route-Based Split-Tunneling Engine**  
> 🔗 [Back to Project README.md](../README.md) | [Legal Disclaimer](DISCLAIMER.md)

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
3. ⚡ **Pure Go Monorepo**:
   - Both the VPS Relay daemon (`lagvex-relay`) and the Client engine (`lagvex-client`) are written in Go. Sub-millisecond garbage collection pauses guarantee zero jitter for real-time game packets.
4. 🐧 **Kernel-Level NAT on the VPS**:
   - The Go relay server does **not** perform slow user-space NAT. It writes raw IPv4 packets directly into the Linux TUN interface (`lagvex0`). The Linux kernel handles `iptables` MASQUERADE, conntrack, and TCP MSS clamping at hardware line speed.

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
          │ Wraps packet with 9-byte binary header (Version, Type, SessionID)
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
          │ Validates Session ID & verifies anti-spoofing inner source IP
          │ Strips 9-byte header, writes raw IPv4 packet into TUN interface: lagvex0
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

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
