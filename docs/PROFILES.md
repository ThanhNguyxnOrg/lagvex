# 🌐 Lagvex Game Profiles & CIDR Routing Guide 🎮

> **Comprehensive Network Maps for Competitive Esports & Online Titles**  
> 🔗 [Back to Project README.md](../README.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 📋 Profile Schema (`configs/profiles.json`)

Lagvex utilizes a structured JSON schema to define game metadata, monitored executables, and regional server CIDR blocks:

```json
{
  "schemaVersion": 1,
  "updatedUtc": "2026-09-12T00:00:00Z",
  "games": [
    {
      "id": "valorant",
      "name": "Valorant",
      "category": "Tactical FPS",
      "icon": "🎯",
      "processNames": [
        "VALORANT-Win64-Shipping.exe",
        "RiotClientServices.exe",
        "VALORANT.exe"
      ],
      "regions": [
        {
          "id": "asia-sg",
          "name": "Southeast Asia (Singapore)",
          "source": "Riot Direct AS6507 & AWS ap-southeast-1",
          "cidrs": [
            "13.250.0.0/15",
            "13.228.0.0/15",
            "18.140.0.0/15",
            "52.220.0.0/15",
            "54.251.0.0/16",
            "54.255.0.0/16"
          ]
        }
      ]
    }
  ],
  "relays": [
    {
      "id": "relay-1",
      "name": "Singapore Premium VPS",
      "location": "Singapore",
      "endpoint": "198.51.100.10:51820",
      "psk": "sample_pre_shared_key"
    }
  ]
}
```

---

## 2. 🎮 Supported Games & Server Networks 🌐

### 🎯 A. Valorant (Riot Games)
- 🏢 **Infrastructure**: Hybrid deployment of **Riot Direct** (Autonomous System AS6507) and dedicated **AWS** game server clusters.
- 🌐 **Regions**:
  - `asia-sg`: Singapore cluster (`ap-southeast-1`)
  - `asia-hk`: Hong Kong cluster (`ap-east-1`)
  - `asia-jp`: Tokyo cluster (`ap-northeast-1`)
  - `asia-in`: Mumbai cluster (`ap-south-1`)
  - `eu-frankfurt`: Frankfurt cluster (`eu-central-1`)
- 🔍 **Process Monitored**: `VALORANT-Win64-Shipping.exe`.

### 💣 B. Counter-Strike 2 (CS2) & Dota 2 (Valve)
- 🏢 **Infrastructure**: Valve Corporation Autonomous System (AS32590) with **Steam Datagram Relay (SDR)** edge gateways.
- 🌐 **Regions**:
  - `sgp`: Singapore SDR relays (`103.10.124.0/24`, `103.28.54.0/24`, `45.121.184.0/24`, `162.254.197.0/24`, `155.133.254.0/24`)
  - `hkg`: Hong Kong SDR relays (`153.254.86.0/24`, `162.254.193.0/24`, `155.133.244.0/24`)
  - `tyo`: Tokyo SDR relays (`155.133.239.0/24`, `155.133.245.0/24`, `45.121.186.0/24`)
  - `sel`: Seoul SDR relays (`155.133.234.0/24`, `162.254.196.0/24`)
  - `fra`: Frankfurt SDR relays (`155.133.226.0/24`, `155.133.248.0/24`)
- 🔍 **Processes Monitored**: `cs2.exe` and `dota2.exe`.

### 🪂 C. PUBG: BATTLEGROUNDS (Krafton)
- 🏢 **Infrastructure**: Microsoft Azure regional datacenters combined with AWS EC2 game servers.
- 🌐 **Regions**:
  - `asia-sg`: Azure Southeast Asia (`20.24.48.0/20`, `52.139.208.0/20`, `20.198.192.0/19`, `20.197.0.0/18`) and AWS Singapore (`13.212.0.0/15`, `18.140.0.0/15`).
  - `asia-jp`: Tokyo AWS & Azure Japan East.
  - `asia-kr`: Seoul AWS & Azure Korea Central.
- 🔍 **Processes Monitored**: `TslGame.exe` and `TslGame_BE.exe`.

### ⚡ D. Apex Legends (EA / Respawn)
- 🏢 **Infrastructure**: Multiplay (Unity Gaming Services) hosted on Google Cloud Platform and AWS.
- 🌐 **Regions**: Singapore, Tokyo, Taiwan, Oregon.
- 🔍 **Processes Monitored**: `r5apex.exe` and `r5apex_dx12.exe`.

### 🏆 E. The Finals (Embark Studios)
- 🏢 **Infrastructure**: AWS and G-Core infrastructure.
- 🌐 **Regions**: Singapore, Tokyo, Frankfurt.
- 🔍 **Process Monitored**: `Discovery.exe`.

### 🎖️ F. Call of Duty: Warzone / MW3 (Activision)
- 🏢 **Infrastructure**: Activision Demonware global network.
- 🌐 **Regions**: Singapore, Tokyo, US-West.
- 🔍 **Processes Monitored**: `cod.exe` and `bootstrapper.exe`.

### 🦅 G. Delta Force: Hawk Ops (Team Jade / Tencent)
- 🏢 **Infrastructure**: Tencent Cloud and AWS regional game nodes.
- 🌐 **Regions**: Singapore, Hong Kong.
- 🔍 **Process Monitored**: `DeltaForceClient-Win64-Shipping.exe`.

---

## 3. 🔬 Discovering CIDRs for New Games 🔍

To capture and add a new game title to Lagvex:

1. 📊 **Monitor Sockets in Live Match**:
   - Launch Windows **Resource Monitor** (`resmon.exe`) -> **Network** tab.
   - Filter by your game's `.exe` while playing inside a server.
   - Record the destination IP and UDP port transferring game data (typically 10-40 KB/s steady).
2. 🔍 **Look up Autonomous System (ASN)**:
   ```bash
   whois <GAME_SERVER_IP> | grep -E "OriginAS|NetRange|CIDR"
   ```
3. 🎯 **Keep CIDRs Narrow**:
   - Cross-check with official public cloud ranges (AWS `ip-ranges.json`, Azure IP ranges).
   - Prefer `/20` through `/24` subnets to avoid dragging unrelated services through your VPS relay.
4. 🌟 **Register via UI**:
   - Open Dashboard at `http://127.0.0.1:18888`.
   - Click **+ Add Custom Game**, enter process name and paste your CIDRs!

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
