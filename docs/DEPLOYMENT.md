# 🚀 Lagvex Relay VPS Deployment & Operations ☁️

> **Step-by-Step Operator Guide for Self-Hosting Lagvex**  
> 🔗 [Back to Project README.md](../README.md) | [Why a Relay is Needed](WHY_VPS.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 📋 System Requirements

- 💻 **Virtualization**: KVM, Xen, or Bare Metal. *(OpenVZ and basic LXC containers cannot create TUN interfaces and are not supported).*
- 🐧 **Operating System**: Ubuntu 20.04+, Debian 11+, CentOS 8+, Rocky Linux, AlmaLinux 9+, or Arch Linux.
- ⚙️ **Kernel**: Linux 5.4 or newer with `tun` module support.
- 🛡️ **Firewall**: `iptables` or `firewalld` with root / `sudo` access.
- ⚡ **Hardware**: 1 vCPU, 512 MB RAM, 10 GB Disk (Lagvex is ultra-lightweight; a $3.50/month VPS can easily support 100+ concurrent players).

---

## 2. 🌍 Choosing Optimal VPS Locations

For players based in Southeast Asia / Vietnam:
- 🇸🇬 **Singapore (SGP)**: The primary routing hub for Southeast Asian gaming servers (**Valorant, CS2, PUBG, Apex, LoL**). Providers with direct fiber peering: Vultr, Linode, AWS EC2 (`ap-southeast-1`), Oracle Cloud SG, OVH Singapore.
- 🇯🇵 **Tokyo, Japan (TYO)**: Ideal for connecting to Japanese and East Asian game servers.
- 🇭🇰 **Hong Kong (HKG)**: Low-latency hop for southern China, Taiwan, and Korean gaming clusters.

---

## 3. ⚡ Quick Deployment (One-Liner) 🏁

Run this single command as root on your VPS:

```bash
curl -fsSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash
```

**What the installer does automatically:**
1. 🔌 Verifies kernel TUN support (`/dev/net/tun`).
2. 🔧 Configures `sysctl` for high-speed packet forwarding:
   ```ini
   net.ipv4.ip_forward = 1
   net.ipv4.conf.all.rp_filter = 2
   net.core.rmem_max = 8388608
   net.core.wmem_max = 8388608
   ```
3. 🛡️ Sets up `iptables` / `ufw` NAT MASQUERADE and TCP MSS clamping, saving rules to persist across reboots.
4. 🔑 Generates a cryptographically random 32-character Pre-Shared Key (PSK) stored securely in `/etc/lagvex/psk.key`.
5. 📦 Installs the `lagvex-relay` binary to `/usr/local/bin/lagvex-relay`.
6. ⚙️ Registers and starts the systemd service `lagvex-relay.service`.
7. 📋 Displays your server IP, Port, PSK, and ready-to-use client JSON snippet.

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
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
