# Lagvex Game Profiles & CIDR Routing Guide

This document describes the structure of game profiles in Lagvex, the mechanics of region-specific routing, and the methodology for discovering and validating new game server CIDR ranges.

---

## 1. Profile Schema (`configs/profiles.json`)

Lagvex uses a clean, extensible JSON schema to define game metadata, processes, and network boundaries:

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

## 2. Supported Games & Server Networks

### A. Valorant (Riot Games)
- **Infrastructure**: Hybrid deployment of **Riot Direct** (Autonomous System AS6507) and dedicated **AWS** game server clusters.
- **Regions**:
  - `asia-sg`: Singapore cluster (`ap-southeast-1`)
  - `asia-hk`: Hong Kong cluster (`ap-east-1`)
  - `asia-jp`: Tokyo cluster (`ap-northeast-1`)
  - `asia-in`: Mumbai cluster (`ap-south-1`)
  - `eu-frankfurt`: Frankfurt cluster (`eu-central-1`)
- **Process Detection**: Watches `VALORANT-Win64-Shipping.exe`.

### B. Counter-Strike 2 (CS2) & Dota 2 (Valve)
- **Infrastructure**: Valve Corporation Autonomous System (AS32590) with **Steam Datagram Relay (SDR)** edge gateways.
- **Regions**:
  - `sgp`: Singapore SDR relays (`103.10.124.0/24`, `103.28.54.0/24`, `45.121.184.0/24`, `162.254.197.0/24`, `155.133.254.0/24`)
  - `hkg`: Hong Kong SDR relays (`153.254.86.0/24`, `162.254.193.0/24`, `155.133.244.0/24`)
  - `tyo`: Tokyo SDR relays (`155.133.239.0/24`, `155.133.245.0/24`, `45.121.186.0/24`)
  - `sel`: Seoul SDR relays (`155.133.234.0/24`, `162.254.196.0/24`)
  - `fra`: Frankfurt SDR relays (`155.133.226.0/24`, `155.133.248.0/24`)
- **Process Detection**: Watches `cs2.exe` and `dota2.exe`.

### C. PUBG: BATTLEGROUNDS (Krafton)
- **Infrastructure**: Microsoft Azure regional datacenters combined with AWS EC2 game servers.
- **Regions**:
  - `asia-sg`: Azure Southeast Asia (`20.24.48.0/20`, `52.139.208.0/20`, `20.198.192.0/19`, `20.197.0.0/18`) and AWS Singapore (`13.212.0.0/15`, `18.140.0.0/15`).
  - `asia-jp`: Tokyo AWS & Azure Japan East.
  - `asia-kr`: Seoul AWS & Azure Korea Central.
- **Process Detection**: Watches `TslGame.exe` and `TslGame_BE.exe`.

### D. Apex Legends (EA / Respawn)
- **Infrastructure**: Multiplay (Unity Gaming Services) hosted on Google Cloud Platform and AWS.
- **Regions**: Singapore, Tokyo, Taiwan, Oregon.
- **Process Detection**: Watches `r5apex.exe` and `r5apex_dx12.exe`.

### E. The Finals (Embark Studios)
- **Infrastructure**: AWS and G-Core infrastructure.
- **Regions**: Singapore, Tokyo, Frankfurt.
- **Process Detection**: Watches `Discovery.exe`.

### F. Call of Duty: Warzone / MW3 (Activision)
- **Infrastructure**: Activision Demonware global network.
- **Regions**: Singapore, Tokyo, US-West.
- **Process Detection**: Watches `cod.exe` and `bootstrapper.exe`.

### G. Delta Force: Hawk Ops (Team Jade / Tencent)
- **Infrastructure**: Tencent Cloud and AWS regional game nodes.
- **Regions**: Singapore, Hong Kong.
- **Process Detection**: Watches `DeltaForceClient-Win64-Shipping.exe`.

---

## 3. Discovering CIDRs for New Games

To add a new game or uncover private server CIDRs:

1. **Capture UDP Sockets in Match**:
   - Open Windows **Resource Monitor** (`resmon.exe`) -> **Network** tab -> **Network Activity** and **Listening Ports**.
   - Filter by the game executable name while playing in a live match.
   - Note the destination IP and UDP port of the active game traffic (typically transmitting 10-40 KB/s steadily).
2. **Determine Autonomous System (ASN)**:
   - Query the destination IP with `whois`:
     ```bash
     whois <GAME_SERVER_IP> | grep -E "OriginAS|NetRange|CIDR"
     ```
3. **Verify Narrow CIDR Range**:
   - Cross-check against official cloud provider IP ranges:
     - **AWS**: `https://ip-ranges.amazonaws.com/ip-ranges.json`
     - **Azure**: `https://www.microsoft.com/en-us/download/details.aspx?id=56519`
     - **Cloudflare**: `https://www.cloudflare.com/ips/`
   - Keep the routed CIDRs as narrow as possible (`/20` to `/24`) to avoid diverting unrelated cloud services through the tunnel.
4. **Register in Lagvex**:
   - Open the Lagvex Dashboard at `http://127.0.0.1:18888`.
   - Click **+ Add Custom Game**.
   - Fill in the Game Name, Executable Name, and paste the verified CIDRs.
