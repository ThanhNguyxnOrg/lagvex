# 🚀 Lagvex Relay VPS Deployment & Operations ☁️

> **Step-by-Step Operator Guide for Self-Hosting Lagvex**  
> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Why a Relay is Needed](WHY_VPS.md) | [💸 Running Cost](COST.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 📋 System Requirements

- 💻 **Virtualization**: KVM, Xen, or Bare Metal. *(OpenVZ and basic LXC containers cannot create TUN interfaces and are not supported).*
- 🐧 **Operating System**: Ubuntu 20.04+, Debian 11+, CentOS 8+, Rocky Linux, AlmaLinux 9+, or Arch Linux.
- ⚙️ **Kernel**: Linux 5.4 or newer with `tun` module support.
- 🛡️ **Firewall**: `iptables` or `firewalld` with root / `sudo` access.
- ⚡ **Hardware**: 1 vCPU, 512 MB RAM, 10 GB Disk. Lagvex is lightweight, but capacity depends on your uplink bandwidth, MTU and packet rate — no fixed player-count guarantee is benchmarked, so test with your own workload before a tournament.

---

## 2. 🌍 Choosing Optimal Worldwide Relay Locations

**Lagvex ships no relay fleet** — the bundled relay catalog is empty and every
relay is one you (or your squad) deploys. The table below lists the cities that
work best for each game region; pick the one closest to the datacenter your
game actually uses:

| Continent | 📍 Prime Relay Hubs | 🎯 Target Esports & Game Clusters |
|---|---|---|
| **Asia-Pacific (APAC)** 🌏 | 🇸🇬 **Singapore (SGP)**<br>🇯🇵 **Tokyo (TYO)**<br>🇭🇰 **Hong Kong (HKG)**<br>🇰🇷 **Seoul (SEL)**<br>🇦🇺 **Sydney (SYD)** | Valorant SEA/JP/KR, CS2 SDR Singapore/Tokyo, PUBG Asia, Apex SG/Tokyo, LoL VNG/Riot Direct, Delta Force |
| **Europe (EU)** 🌍 | 🇩🇪 **Frankfurt (FRA)**<br>🇬🇧 **London (LON)**<br>🇫🇷 **Paris (CDG)**<br>🇫🇮 **Helsinki (HEL)**<br>🇪🇸 **Madrid (MAD)** | CS2 EU North/West, Valorant EU Central/West, The Finals Frankfurt, Rainbow Six Siege EU, Dota 2 EU |
| **North America (NA)** 🌎 | 🇺🇸 **US-East (N. Virginia - IAD)**<br>🇺🇸 **US-West (Oregon - PDX / San Jose)**<br>🇺🇸 **US-Central (Dallas - DFW)** | Apex NA East/West, Warzone Demonware US, CS2 NA, Valorant NA, Overwatch 2 NA, Rainbow Six NA |
| **South America (SA)** 🌎 | 🇧🇷 **São Paulo (GRU)**<br>🇨🇱 **Santiago (SCL)** | CS2 South America, Valorant Brazil, League of Legends BR, PUBG South America |
| **Middle East & Africa (MENA)** 🌍 | 🇧🇭 **Bahrain (BAH)**<br>🇦🇪 **Dubai (DXB)**<br>🇿🇦 **Johannesburg (JNB)** | Valorant Middle East, Apex Bahrain, CS2 Dubai, Fortnite Middle East |

---

## 3. ⚡ 1-Click Automated Relay Deployment 🏁

### 🐧 Option A: Deploy on Linux VPS (Ubuntu, Debian, CentOS, Rocky, Arch)

Run this single command as `root` (or with `sudo`) in your VPS terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash
```

**Step-by-step what the Linux installer does automatically:**
1. 🔌 **Checks TUN Support**: Ensures `/dev/net/tun` exists in your kernel.
2. 🔧 **Tunes Kernel Network Parameters**:
   ```ini
   net.ipv4.ip_forward = 1
   net.ipv4.conf.all.rp_filter = 2
   net.core.rmem_max = 8388608
   net.core.wmem_max = 8388608
   ```
3. 🛡️ **Sets Up Firewall & Masquerade**: Adds `iptables` NAT MASQUERADE for the `10.88.0.0/24` tunnel subnet and enables TCP MSS clamping so game packets never fragment.
4. 🔑 **Generates Pre-Shared Key (PSK)**: Generates a cryptographically strong 32-character random key stored at `/etc/lagvex/psk.key`.
5. 📦 **Installs Binary**: Downloads and installs `lagvex-relay` to `/usr/local/bin/lagvex-relay`.
6. ⚙️ **Registers Systemd Service**: Creates `/etc/systemd/system/lagvex-relay.service` and activates it immediately.
7. 📋 **Prints Connection Details**: Outputs the relay endpoint, UDP port and the generated PSK. Add these in the dashboard with **+ Add Node**; a relay cannot be probed or used without its PSK.

---

### 🪟 Option B: Windows or macOS hosting — not supported

Relay hosting requires a Linux kernel TUN device (`/dev/net/tun`) plus `iptables`/NAT
forwarding. Windows and macOS cannot host a relay; `scripts/install-relay.ps1` exits with
an explanation instead of installing a broken service. Use a small Linux VPS (KVM) — see
the requirements above.

---

## 4. 🛠️ Manual Deployment Step-by-Step 🧑‍💻

If you prefer configuring your server manually:

### 🔧 Step 1: Enable IP Forwarding
Create `/etc/sysctl.d/99-lagvex.conf`:
```ini
net.ipv4.ip_forward = 1
net.ipv4.conf.all.rp_filter = 2
net.ipv4.conf.default.rp_filter = 2
net.core.rmem_max = 8388608
net.core.wmem_max = 8388608
```
Apply settings:
```bash
sudo sysctl -q --system
```

### 🛡️ Step 2: Configure NAT Masquerade
Assuming your public network interface is `eth0`:
```bash
# Allow incoming UDP on relay port (default 51820)
sudo iptables -I INPUT 1 -p udp --dport 51820 -j ACCEPT

# Forwarding rules between TUN (10.88.0.0/24) and WAN
sudo iptables -I FORWARD 1 -s 10.88.0.0/24 -o eth0 -j ACCEPT
sudo iptables -I FORWARD 2 -d 10.88.0.0/24 -i eth0 -m state --state RELATED,ESTABLISHED -j ACCEPT

# Enable NAT MASQUERADE
sudo iptables -t nat -I POSTROUTING 1 -s 10.88.0.0/24 -o eth0 -j MASQUERADE

# Enable TCP MSS Clamping
sudo iptables -t mangle -I FORWARD 1 -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu
```

### 📦 Step 3: Build & Install Binary
```bash
git clone https://github.com/ThanhNguyxnOrg/lagvex.git
cd lagvex
make relay
sudo install -m 755 bin/lagvex-relay /usr/local/bin/lagvex-relay
```

### ⚙️ Step 4: Configure Systemd Unit
Create `/etc/systemd/system/lagvex-relay.service`:
```ini
[Unit]
Description=Lagvex Game Booster Relay Daemon
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lagvex-relay -listen :51820 -tun lagvex0 -subnet 10.88.0.0/24 -psk-file /etc/lagvex/psk.key
Restart=always
RestartSec=3s
LimitNOFILE=65536
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```
Start and enable:
```bash
sudo mkdir -p /etc/lagvex
echo "your_secret_psk_key_here" | sudo tee /etc/lagvex/psk.key
sudo chmod 600 /etc/lagvex/psk.key

sudo systemctl daemon-reload
sudo systemctl enable --now lagvex-relay
```

---

## 5. 🐳 Docker Deployment 📦

Deploying with Docker Compose:

```bash
git clone https://github.com/ThanhNguyxnOrg/lagvex.git
cd lagvex

# Generate a key once for this relay
LAGVEX_PSK=$(openssl rand -hex 32)
export LAGVEX_PSK

# Launch container in background
docker compose up -d
```

Save the key in your password manager and supply the same `LAGVEX_PSK` for
future Compose commands. Share it privately with clients. If your relay used
the old Compose default, recreate it with a new key and update its clients.

---

## 6. 🔍 Maintenance & Troubleshooting 🩺

### 📜 Inspect Live Service Logs:
```bash
sudo journalctl -u lagvex-relay -f
```

### 👥 Check Active Handshakes & Player Count:
```bash
sudo journalctl -u lagvex-relay | grep "Handshake OK"
```

### 🔌 Verify TUN Interface State:
```bash
ip addr show dev lagvex0
```

---

## 7. 🌐 The Public HTTP Port (squad signalling + relay probe)

The relay serves two open HTTP endpoints on **UDP port + 1** (default
`51821`), because squad rooms must be joinable across networks:

- `/squad/*` — squad room signalling (CORS-open, rate limited per IP, capped
  room count, CSPRNG codes).
- `/probe/gameserver` — the game-server TCP probe the dashboard uses for the
  relay→server route segment. It takes **no PSK**: the measurement is exactly
  what the client's own TCP ping tool performs locally and reveals nothing
  about the relay's configuration.

### Abuse surface, and its limits

| Risk | What an attacker gets | Mitigation |
|---|---|---|
| **Port-scan oracle** — the RTT/loss pair distinguishes open, refused and filtered ports on any public host | a slow, rate-limited probe of third-party hosts | 1 request / 2 s / IP **and** a sustained 600/hour/IP cap (a legitimate dashboard needs ~60/hour); private, loopback, link-local and bogon hosts are refused **before** dialing; a limiter of its own that squad traffic cannot consume |
| **Reflected load** — asking the relay to dial a victim 3× per request | ≤3 TCP handshakes per 2 s per source IP | same limiter; no UDP, no amplification |
| **Resource pinning** | nothing persisted per request | per-IP bookkeeping capped at 4096 entries |

If you do not need squad rooms or the route segment, close the port:
`iptables -A INPUT -p tcp --dport 51821 -j DROP`. The UDP data-plane port is
the only one the tunnel itself requires. The probe endpoint is what makes the
"relay→server" segment measurable, so closing it only removes that display —
never fabricates a value in its place (`—` with a reason instead).

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
