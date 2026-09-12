#!/usr/bin/env bash
# ==============================================================================
# Lagvex Relay - One-Liner Automated Linux VPS Installer
# Supported OS: Ubuntu, Debian, CentOS, RHEL, Rocky Linux, AlmaLinux, Arch Linux
# ==============================================================================
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
cat << 'EOF'
  _                                  
 | |    __ _  __ ___   _____  __     
 | |   / _` |/ _` \ \ / / _ \ \/ /     Lagvex Relay Installer
 | |__| (_| | (_| |\ V /  __/>  <      High-Performance Game Booster Relay
 |_____\__,_|\__, | \_/ \___/_/\_\     https://github.com/lagvex/lagvex
             |___/                   
EOF
echo -e "${NC}"

if [[ $EUID -ne 0 ]]; then
    echo -e "${RED}[ERROR] This script must be run as root. Please run with sudo.${NC}" >&2
    exit 1
fi

PORT="${LAGVEX_PORT:-51820}"
SUBNET="${LAGVEX_SUBNET:-10.88.0.0/24}"
TUN_DEV="${LAGVEX_TUN:-lagvex0}"
PSK="${LAGVEX_PSK:-}"
INSTALL_DIR="/usr/local/bin"
CONF_DIR="/etc/lagvex"

# Detect WAN interface
WAN_IF="$(ip -4 route show default 2>/dev/null | awk '/default/ {print $5; exit}')"
if [[ -z "$WAN_IF" ]]; then
    WAN_IF="$(ip link | awk -F: '$0 !~ "lo|vir|wl" && /^[0-9]+: / {print $2; exit}' | tr -d ' ')"
fi
if [[ -z "$WAN_IF" ]]; then
    echo -e "${RED}[ERROR] Could not auto-detect WAN network interface. Set LAGVEX_WAN=eth0 and re-run.${NC}" >&2
    exit 1
fi

# Detect Public IP
PUBLIC_IP="$(curl -s4 --max-time 5 https://api.ipify.org || curl -s4 --max-time 5 https://ifconfig.me || ip -4 addr show dev "$WAN_IF" | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n 1)"

echo -e "${CYAN}==> System Info:${NC}"
echo -e "    Public IP: ${BOLD}${PUBLIC_IP}${NC}"
echo -e "    WAN Interface: ${BOLD}${WAN_IF}${NC}"
echo -e "    Relay UDP Port: ${BOLD}${PORT}${NC}"
echo -e "    Tunnel Subnet: ${BOLD}${SUBNET}${NC}"
echo ""

# 1. Check TUN Device
echo -e "${CYAN}==> [1/6] Checking Linux TUN driver...${NC}"
modprobe tun 2>/dev/null || true
mkdir -p /etc/modules-load.d
echo "tun" > /etc/modules-load.d/lagvex.conf

if [[ ! -c /dev/net/tun ]]; then
    echo -e "${RED}[ERROR] /dev/net/tun is not available. Please verify this VPS is KVM/Bare-metal (OpenVZ cannot create TUN devices).${NC}" >&2
    exit 1
fi

# 2. Sysctl Tweaks
echo -e "${CYAN}==> [2/6] Configuring sysctl (IP forwarding & UDP buffers)...${NC}"
cat > /etc/sysctl.d/99-lagvex.conf << 'EOF'
net.ipv4.ip_forward = 1
net.ipv4.conf.all.rp_filter = 2
net.ipv4.conf.default.rp_filter = 2
net.core.rmem_max = 8388608
net.core.wmem_max = 8388608
net.ipv4.udp_rmem_min = 16384
net.ipv4.udp_wmem_min = 16384
EOF
sysctl -q --system 2>/dev/null || true

# 3. Firewall & NAT Masquerade
echo -e "${CYAN}==> [3/6] Setting up firewall & NAT MASQUERADE...${NC}"
if command -v ufw >/dev/null 2>&1 && ufw status | grep -q "Status: active"; then
    ufw allow "${PORT}/udp" comment "Lagvex Relay" >/dev/null
fi

if command -v iptables >/dev/null 2>&1; then
    # Add UDP Port
    iptables -C INPUT -p udp --dport "$PORT" -j ACCEPT 2>/dev/null || \
        iptables -I INPUT 1 -p udp --dport "$PORT" -j ACCEPT

    # Forwarding
    iptables -C FORWARD -s "$SUBNET" -o "$WAN_IF" -j ACCEPT 2>/dev/null || \
        iptables -I FORWARD 1 -s "$SUBNET" -o "$WAN_IF" -j ACCEPT
    iptables -C FORWARD -d "$SUBNET" -i "$WAN_IF" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
        iptables -I FORWARD 2 -d "$SUBNET" -i "$WAN_IF" -m state --state RELATED,ESTABLISHED -j ACCEPT

    # NAT Masquerade
    iptables -t nat -C POSTROUTING -s "$SUBNET" -o "$WAN_IF" -j MASQUERADE 2>/dev/null || \
        iptables -t nat -I POSTROUTING 1 -s "$SUBNET" -o "$WAN_IF" -j MASQUERADE

    # TCP MSS Clamping
    iptables -t mangle -C FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu 2>/dev/null || \
        iptables -t mangle -I FORWARD 1 -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

    # Save rules
    if command -v netfilter-persistent >/dev/null 2>&1; then
        netfilter-persistent save >/dev/null 2>&1 || true
    elif command -v iptables-save >/dev/null 2>&1; then
        mkdir -p /etc/iptables
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
    fi
fi

# 4. Generate Pre-Shared Key (PSK)
echo -e "${CYAN}==> [4/6] Setting up Pre-Shared Key (PSK)...${NC}"
mkdir -p "$CONF_DIR"
if [[ -z "$PSK" ]]; then
    if [[ -f "$CONF_DIR/psk.key" ]]; then
        PSK="$(cat "$CONF_DIR/psk.key" | tr -d '[:space:]')"
        echo -e "    Using existing PSK from $CONF_DIR/psk.key"
    else
        PSK="$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32)"
        echo "$PSK" > "$CONF_DIR/psk.key"
        chmod 600 "$CONF_DIR/psk.key"
        echo -e "    Generated new secure 32-character PSK."
    fi
else
    echo "$PSK" > "$CONF_DIR/psk.key"
    chmod 600 "$CONF_DIR/psk.key"
fi

# 5. Build or Install Binary
echo -e "${CYAN}==> [5/6] Installing Lagvex Relay binary...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." 2>/dev/null && pwd || true)"

if [[ -f "$PROJECT_ROOT/cmd/lagvex-relay/main.go" ]] && command -v go >/dev/null 2>&1; then
    echo "    Compiling lagvex-relay from local source..."
    (cd "$PROJECT_ROOT" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/lagvex-relay" ./cmd/lagvex-relay)
elif [[ -f "$SCRIPT_DIR/lagvex-relay" ]]; then
    install -m 755 "$SCRIPT_DIR/lagvex-relay" "$INSTALL_DIR/lagvex-relay"
else
    # Compile or download release
    if command -v go >/dev/null 2>&1; then
        echo "    Building lagvex-relay via Go..."
        TMP_BUILD="$(mktemp -d)"
        git clone --depth 1 https://github.com/lagvex/lagvex.git "$TMP_BUILD" 2>/dev/null || true
        if [[ -d "$TMP_BUILD/cmd/lagvex-relay" ]]; then
            (cd "$TMP_BUILD" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/lagvex-relay" ./cmd/lagvex-relay)
            rm -rf "$TMP_BUILD"
        fi
    fi
fi

chmod +x "$INSTALL_DIR/lagvex-relay" 2>/dev/null || true

# 6. Install Systemd Service
echo -e "${CYAN}==> [6/6] Configuring systemd service...${NC}"
cat > /etc/systemd/system/lagvex-relay.service << EOF
[Unit]
Description=Lagvex Game Booster Relay Daemon
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/lagvex-relay -listen :${PORT} -tun ${TUN_DEV} -subnet ${SUBNET} -psk-file ${CONF_DIR}/psk.key
Restart=always
RestartSec=3s
LimitNOFILE=65536
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now lagvex-relay.service

echo ""
echo -e "${GREEN}${BOLD}===================================================================${NC}"
echo -e "${GREEN}${BOLD}  LAGVEX RELAY INSTALLED & RUNNING SUCCESSFULLY!                   ${NC}"
echo -e "${GREEN}${BOLD}===================================================================${NC}"
echo -e "  Server IP  : ${BOLD}${PUBLIC_IP}${NC}"
echo -e "  UDP Port   : ${BOLD}${PORT}${NC}"
echo -e "  PSK Key    : ${BOLD}${YELLOW}${PSK}${NC}"
echo ""
echo -e "${CYAN}Copy this configuration into your Lagvex Client:${NC}"
cat << EOF
{
  "id": "my-vps-relay",
  "name": "Custom VPS Relay",
  "endpoint": "${PUBLIC_IP}:${PORT}",
  "psk": "${PSK}"
}
EOF
echo -e "${GREEN}${BOLD}===================================================================${NC}"
