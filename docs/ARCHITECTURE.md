# Lagvex Architecture

This document describes the design, implementation, and network principles of **Lagvex**, an open-source game latency reducer and network optimization engine.

---

## 1. Core Principles

The fundamental architectural decisions behind Lagvex:

1. **Route-based Split Tunneling, Not Socket Hooking**:
   - Traditional game boosters or proxy tools often use Windows Filtering Platform (WFP), LSP, or packet-divert drivers (e.g., WinDivert) to intercept sockets per process. Sitting inside a game process's network stack is dangerous: strict kernel-level anti-cheat systems (Riot Vanguard, Valve Anti-Cheat, BattlEye, Easy Anti-Cheat, RICOCHET) routinely flag or ban socket injection and memory hooks.
   - Lagvex operates **purely at the Windows routing table level**. It modifies zero game memory, injects zero DLLs, and touches zero game sockets. From the perspective of anti-cheat software, traffic simply leaves the computer through a standard network adapter.
2. **WinTun for the Virtual Network Adapter**:
   - WinTun is the official high-performance Layer-3 TUN driver developed by the WireGuard project. It is cryptographically signed by WireGuard LLC, runs inside the Windows kernel, provides ultra-low latency ring-buffer I/O in userspace, and automatically destroys the adapter on process exit.
3. **Pure Go Monorepo**:
   - Both the Linux Relay daemon (`lagvex-relay`) and the Windows Client engine (`lagvex-client`) are implemented in pure Go. Go's runtime garbage collector pauses are well under 100 microseconds—negligible compared to physical internet latency.
   - Statically compiled binaries require zero external dependencies (.NET runtimes, C++ redistributables, or Python).
4. **Kernel-Level NAT on the VPS**:
   - The Go relay server does **not** perform user-space NAT or connection tracking. It simply decapsulates packets and writes raw IPv4 packets into a Linux TUN interface (`lagvex0`). The Linux kernel handles `iptables` MASQUERADE, conntrack, and TCP MSS clamping natively at line speed.

---

## 2. The Path of a Packet

```
   [ Game Process: e.g. VALORANT-Win64-Shipping.exe ]
          │  sends UDP packet to game server 13.250.45.10:7000
          ▼
   [ Windows Network Stack ]
          │  Routing table lookup: 13.250.0.0/15 matches on-link -> Interface "Lagvex"
          │  (All unrelated traffic like Discord/Browser matches 0.0.0.0/0 -> Physical NIC)
          ▼
   [ WinTun Virtual Adapter ] ────────┐
                                       │  Reads raw IPv4 packet from ring buffer
                                       ▼
   [ Lagvex Client Engine (lagvex-client.exe) ]
          │  Wraps packet with 9-byte header (Version, Type, SessionID)
          ▼
   [ Standard UDP Socket ]
          │  Sends to Relay VPS (e.g. 198.51.100.5:51820)
          │  (Pinned /32 route forces this packet out the physical network card)
          ▼
   ═════════════════════ International Fiber / VPS ═════════════════════
          ▼
   [ Lagvex Relay Server (lagvex-relay on Linux VPS) ]
          │  Receives UDP datagram on port 51820
          │  Validates Session ID, performs anti-spoofing check
          │  Strips 9-byte header, writes raw IPv4 packet into TUN interface (lagvex0)
          ▼
   [ Linux Kernel (VPS) ]
          │  iptables NAT MASQUERADE rewrites source IP to VPS Public IP
          │  Clamps TCP MSS to PMTU
          │  Forwards packet to physical WAN interface (eth0)
          ▼
   [ Target Game Server (Singapore / Tokyo / Frankfurt) ]
```

### Return Path:
1. The game server responds to the VPS Public IP.
2. The Linux kernel un-NATs the packet using conntrack and delivers it to the TUN interface `lagvex0`.
3. `lagvex-relay` reads the packet from `lagvex0`, extracts the destination IPv4 address (e.g., `10.88.0.2`), looks up the client's session, wraps the packet in a 9-byte header, and sends it via UDP back to the client.
4. `lagvex-client` receives the UDP datagram, strips the header, and injects the raw IPv4 packet into the WinTun ring buffer.
5. The Windows network stack delivers the packet to the game process seamlessly.

---

## 3. Windows Routing Table Mechanics

The routing engine ([`pkg/client/routes_windows.go`](file:///D:/Code/Lagvex/pkg/client/routes_windows.go)) adheres to three strict safety rules:

### A. Pinned `/32` Relay Route (Anti-Loop Guarantee)
Before any game route is installed, a host route (`/32`) pointing to the Relay VPS IP is pinned through the physical network interface:
```cmd
netsh interface ipv4 add route prefix=<RELAY_IP>/32 interface=<PHYSICAL_INDEX> nexthop=<DEFAULT_GATEWAY> metric=1 store=active
```
**Why this is critical:** If a game's CIDR range ever overlaps with or encloses the Relay server's own IP address, packets addressed to the relay would get sucked into the virtual tunnel, creating an infinite routing loop that cuts off internet access. The pinned `/32` route guarantees the tunnel endpoint always exits via the physical gateway.

### B. On-Link Game Routes
When routes for a game's CIDR blocks are installed, they are configured as **on-link** (no nexthop gateway specified):
```cmd
netsh interface ipv4 add route prefix=<GAME_CIDR> interface=<WINTUN_INDEX> metric=1 store=active
```
**Why this matters:** WinTun is an NDIS Layer-3 interface without an ARP/NDIS link layer. If a nexthop IP is specified, Windows attempts to resolve ARP neighbor entries that can never exist, causing silent packet drops. On-link routing allows Windows to hand raw IP packets directly to the driver.

### C. `store=active` (RAM-Only Safety Brake)
Every route added by Lagvex is flagged with `store=active`. These entries exist only in volatile memory (RAM). If the Lagvex process crashes, if Windows hangs, or if power is cut, the routing table automatically reverts to a pristine state upon reboot.

---

## 4. Smart Process Watcher Lifecycle

AWS and Azure cloud blocks encompass thousands of third-party services. Leaving game routes permanently active would route unrelated web services through the relay.

Lagvex solves this using the **Smart Process Watcher** ([`pkg/client/watcher.go`](file:///D:/Code/Lagvex/pkg/client/watcher.go)):
1. While the user is connected to a relay and waiting in a lobby, no game routes are installed.
2. The watcher monitors system processes every 2 seconds for known game binaries (e.g., `VALORANT-Win64-Shipping.exe`, `cs2.exe`, `TslGame.exe`).
3. **Game Launch Detected**: Routes for the active game and region are dynamically injected into the routing table.
4. **Game Exit Detected**: Routes are automatically deleted, returning the network to direct routing immediately.
5. Users may also toggle **Force Route Always** in the Dashboard to bypass the watcher when testing private servers or custom matches.

---

## 5. Security Model

| Threat | Mitigation Mechanism |
|---|---|
| **Unauthorized Relay Usage** | Handshake authenticated via HMAC-SHA256 over Pre-Shared Key (PSK). Non-matching handshakes are silently dropped. |
| **Replay Attacks** | Handshake payload contains a Unix timestamp checked against a 120-second clock skew tolerance window. |
| **IP Spoofing Between Clients** | The relay verifies that the inner IPv4 source address strictly matches the IP assigned to that session ID. |
| **Intranet & Metadata Scanning** | The relay blocks outbound packets destined for RFC 1918 subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback (`127.0.0.0/8`), and cloud metadata (`169.254.169.254`). |
| **Client Roaming / NAT Changes** | Client UDP endpoint (`IP:Port`) is dynamically updated on every valid data packet, allowing seamless transitions across Wi-Fi/Ethernet. |
