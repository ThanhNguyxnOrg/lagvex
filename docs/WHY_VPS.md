# 🌐 Why a Relay Server (VPS) is Required ⚡

> 💡 **Related Documents**: [📖 Back to README](../README.md) • [📐 Architecture Overview](ARCHITECTURE.md) • [🚀 Deployment Guide](DEPLOYMENT.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)

---

## 🎯 Executive Summary 📌

A frequent question from gamers is:  
> *"Why does Lagvex need a remote relay server (VPS)? Can't a software tool optimize my ping directly from my gaming PC without any external server?"*

The short answer is **physical networking reality**: **No software installed on your computer can alter how your domestic Internet Service Provider (ISP) routes packets across the ocean once they leave your home router.** 🌊

To bypass international transit bottlenecks, avoid congested undersea fiber cables, and achieve rock-solid latency, an intermediate relay server strategically positioned near the game's datacenter is fundamentally required. 🛰️

This document explains the technical reasons behind this architecture in a completely vendor-neutral, educational manner. 🔬

---

## 🌊 The Root Problem: How International Game Traffic Travels 🌍

When you play a competitive game whose servers are located abroad (such as Singapore 🇸🇬, Tokyo 🇯🇵, Hong Kong 🇭🇰, or Frankfurt 🇩🇪), your gameplay packets travel through several network layers:

```
[ 🖥️ Gaming PC ]
       │
       ▼ (1. Local Ethernet / Wi-Fi)
[ 📡 Home Router ]
       │
       ▼ (2. Domestic ISP Access Network - Viettel / VNPT / FPT / etc.)
[ 🏢 ISP Regional Aggregation Gateway ]
       │
       ▼ (3. International Transit & Submarine Cables: AAG, APG, IA, AAE-1)
       │    ⚠️ [BOTTLENECK]: Peering congestion, bufferbloat, fiber cuts, sub-optimal BGP routing
       ▼
[ 🏢 Overseas Tier-1 Transit Carrier ]
       │
       ▼ (4. Cloud / Datacenter Ingress: Riot Direct / Valve SDR / AWS / Azure)
[ 🎮 Game Dedicated Server ]
```

### 🚨 Why Your Local PC Cannot Fix This Alone ❌
1. **The BGP Routing Trap**: Domestic ISPs negotiate international transit based on economic cost rather than lowest gamer latency. Your ISP's Border Gateway Protocol (BGP) tables often route packets through indirect hops (e.g., routing through Hong Kong or the US West Coast before reaching Singapore) because those transit contracts are cheaper for the ISP.
2. **Submarine Cable Congestion**: International undersea cables experience high utilization and frequent fiber faults. During peak evening hours (8:00 PM – 11:00 PM), packet buffers fill up on international border routers, causing **bufferbloat, jitter spikes, and packet loss**.
3. **No Local Control Beyond the Gateway**: Any "ping booster" or "optimizer" claiming to speed up your international ping *without* an external relay is technically incapable of affecting packets after they leave your home router.

---

## ⚡ The Solution: The Strategic Relay Proxy 🛰️

Lagvex solves this physical routing limitation by inserting a dedicated high-performance **Relay Server** into the transmission path:

```
                               ┌────────────────────────────────────────────────────────┐
                               │                 ⚡ LAGVEX TUNNEL ⚡                   │
[ 🖥️ Gaming PC ] ─────────────►│ Direct, high-priority UDP stream to Relay Node         │
                               └──────────────────────────┬─────────────────────────────┘
                                                          ▼
                                            [ 🌐 Lagvex Relay Server (VPS) ]
                                            (Located in Singapore / Tokyo / etc.)
                                                          │
                                                          ▼ (< 1ms local fiber transit)
                                            [ 🎮 Game Server (AWS / Valve / Riot) ]
```

### 🏆 Key Benefits of the Relay Model:
- 🚀 **Narrows the Transit Path**: Instead of dozens of unpredictable intermediate hops, your traffic travels via a single optimized UDP stream directly to your relay node.
- 🎯 **Datacenter Proximity**: A relay located in the target region (e.g., Singapore) sits inside a high-bandwidth datacenter with Tier-1 connectivity. The distance from the relay to the game server (AWS `ap-southeast-1`, Valve SDR, or Riot Direct) is virtually negligible (< 1ms to 2ms).
- 🛡️ **Wire-Speed Kernel NAT**: The Lagvex relay leverages Linux kernel `/dev/net/tun` and hardware IP masquerading, decapsulating packets in microseconds without injecting artificial delay.
- 💎 **Eliminates Packet Loss**: By bypassing the congested public peering exchanges that domestic ISPs use, packet loss drops to 0%.

---

## 👥 One Relay Serves Your Whole Squad / Community 🤝

**You do NOT need a separate relay server for every individual player!** 🚫

Because the Lagvex relay daemon is engineered in lightweight Go and offloads packet forwarding to the Linux kernel network stack:
- 📊 **Resource Footprint**: The relay daemon consumes less than **15 MB of RAM** and negligible CPU overhead (< 2% on a single core).
- 🎮 **Capacity**: A single modest Linux server can comfortably handle **20 to 50 concurrent players** playing simultaneously without breaking a sweat.
- 🤝 **Shared Community Node**: One player, esports team, or gaming guild can operate a single relay node and share the Pre-Shared Key (PSK) with their entire squad or local gaming club.

---

## 🛠️ Relay Server Technical Requirements 📋

Lagvex is 100% open-source, vendor-neutral, and does not partner with or advertise any specific hosting company. You can run the relay on **any standard generic Linux server** that meets the following modest specifications:

| Component | ⚙️ Minimum Specification | 💡 Notes |
|---|---|---|
| **Operating System** | 🐧 Modern 64-bit Linux (Ubuntu 22.04+, Debian 12+, Rocky Linux, Alpine) | Requires root access for `iptables` and `/dev/net/tun`. |
| **Virtualization** | 🖥️ KVM, Dedicated, or Bare Metal | Avoid OpenVZ / LXC containers unless TUN device passthrough is explicitly enabled. |
| **CPU** | ⚡ 1 vCPU (Shared or Dedicated) | Modern x86_64 or ARM64 architecture. |
| **Memory (RAM)** | 💾 512 MB – 1 GB | Extremely lightweight; 512 MB is more than adequate. |
| **Disk Space** | 💽 5 GB – 10 GB SSD | The compiled binary is under 10 MB. |
| **Networking** | 🌐 1 Public IPv4 Address | Static IP with UDP traffic allowed (default port `4433/udp`). |
| **Bandwidth** | 📈 500 GB – 1 TB monthly transfer | An intense gaming session uses only ~20–40 MB per hour. |
| **Location** | 📍 Geographically close to target game servers | E.g., Singapore for SEA servers; Tokyo for JP/KR servers; Frankfurt for EU servers. |

---

## ⚖️ Relay vs. Commercial VPN: Why Split-Tunneling Matters 🥊

Traditional commercial VPN services also use remote servers, but they route **100% of your computer's internet traffic** through their tunnel.

| Comparison Metric | 🌐 Traditional Commercial VPN | ⚡ **Lagvex Split-Tunneling** |
|---|---|---|
| **Traffic Scope** | Everything (Browser, Discord, Steam downloads, Torrents) | 🎯 **Strictly Game Destination CIDRs Only** |
| **Discord / Voice Quality** | Can introduce robotic voice or lag | 🎙️ **Untouched** (flows directly via local ISP) |
| **Web Browsing & YouTube** | Triggers CAPTCHAs, geoblocks, and streaming slowdowns | 🚀 **Untouched** (full local fiber bandwidth) |
| **Privacy & Security** | VPN operator can inspect all personal web browsing | 🔒 **Zero web traffic passes through relay** |
| **Anti-Cheat Interaction** | Frequently flagged as suspicious proxy by anti-cheats | 🛡️ **Non-Invasive** (Layer-3 kernel routing only, zero memory access) |

---

## 🚀 How to Set Up a Relay 📦

Ready to deploy your own relay? It takes under 60 seconds using our automated installer:

```bash
curl -sSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash
```

For complete instructions on ports, firewall configuration, systemd services, and Docker Compose deployment, consult the [🚀 VPS Deployment Guide](DEPLOYMENT.md).

---

[⬅️ Back to README](../README.md) • [📖 Architecture Overview](ARCHITECTURE.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)
