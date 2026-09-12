# 🚀 Lagvex Relay VPS Deployment & Operations ☁️

> **Step-by-Step Operator Guide for Self-Hosting Lagvex**  
> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Why a Relay is Needed](WHY_VPS.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 📋 System Requirements

- 💻 **Virtualization**: KVM, Xen, or Bare Metal. *(OpenVZ and basic LXC containers cannot create TUN interfaces and are not supported).*
- 🐧 **Operating System**: Ubuntu 20.04+, Debian 11+, CentOS 8+, Rocky Linux, AlmaLinux 9+, or Arch Linux.
- ⚙️ **Kernel**: Linux 5.4 or newer with `tun` module support.
- 🛡️ **Firewall**: `iptables` or `firewalld` with root / `sudo` access.
- ⚡ **Hardware**: 1 vCPU, 512 MB RAM, 10 GB Disk (Lagvex is ultra-lightweight; a $3.50/month VPS can easily support 100+ concurrent players).

---

## 2. 🌍 Choosing Optimal Worldwide Relay Locations

Lagvex supports a global network across 5 continents. Choose a relay location closest to your target game datacenter:

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
7. 📋 **Prints Squad Invite Link**: Outputs a ready-to-share link (`lagvex://connect?endpoint=IP:51820&psk=KEY&name=MySquad`) that friends can paste directly into their Lagvex HUD!

---

### 🪟 Option B: Host on Windows 10/11 or Windows Server (PowerShell)

If you have a Windows PC or Windows Server with a public IP or port-forwarded router:

1. Right-click the **Start Menu** and choose **PowerShell (Admin)** or **Terminal (Admin)**.
2. Run this command:

```powershell
irm https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.ps1 | iex
```

**Step-by-step what the Windows installer does automatically:**
1. 🛡️ **Checks Elevation**: Verifies script runs with local Administrator privileges.
2. 🔌 **Configures WinTun**: Checks for `wintun.dll` and configures the Layer-3 adapter interface.
3. ⚡ **Enables Routing**: Executes `Set-NetIPInterface -Forwarding Enabled` across network interfaces.
4. 🌐 **Configures NetNat**: Provisions Windows NetNat masquerading for `10.88.0.0/24` to route client packets to the Internet.
5. 🔑 **Generates Security Key**: Creates a 32-character PSK stored securely at `C:\ProgramData\Lagvex\relay.json`.
6. 📋 **Prints 1-Click Link**: Shows your public IP, Port 51820, and generates your squad share link ready to paste!

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

# Set custom Pre-Shared Key
export LAGVEX_PSK="your_custom_psk"

# Launch container in background
docker compose up -d
```

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

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
