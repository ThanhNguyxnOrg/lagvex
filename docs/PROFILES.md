# 🌐 Lagvex Game Profiles & CIDR Routing Guide 🎮

> **Comprehensive Network Maps for Competitive Esports & Online Titles**  
> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Legal Disclaimer](DISCLAIMER.md)

---

## 1. 📋 Profile Schema (`configs/profiles.json`)

Lagvex utilizes a structured JSON schema to define game metadata, monitored executables, continent categorization, and regional server CIDR blocks:

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
          "id": "apac-sg",
          "continent": "apac",
          "name": "Singapore (SEA - Riot Direct & AWS)",
          "source": "Riot Direct AS6507 & AWS ap-southeast-1",
          "cidrs": [
            "13.250.0.0/15",
            "13.228.0.0/15",
            "18.140.0.0/15",
            "52.220.0.0/15",
            "54.251.0.0/16",
            "54.255.0.0/16"
          ]
        },
        {
          "id": "eu-frankfurt",
          "continent": "eu",
          "name": "Frankfurt (EU Central - AWS & Riot Direct)",
          "source": "Riot Direct & AWS eu-central-1",
          "cidrs": [
            "3.120.0.0/14",
            "18.194.0.0/15",
            "52.28.0.0/15"
          ]
        }
      ]
    }
  ],
  "relays": [
    {
      "id": "relay-sg-1",
      "name": "🇸🇬 Singapore #1 [Community Free]",
      "location": "Singapore, SG",
      "endpoint": "198.51.100.10:51820",
      "psk": "community_free_access_sg1"
    },
    {
      "id": "relay-fra-1",
      "name": "🇩🇪 Frankfurt #1 [Community Free]",
      "location": "Frankfurt, Germany",
      "endpoint": "198.51.100.20:51820",
      "psk": "community_free_access_fra1"
    }
  ]
}
```

### 🌍 Continent Identifiers
The `continent` field in each regional block enables the HUD to group game servers into clean, intuitive `<optgroup>` clusters:
- `apac`: 🌏 **Asia-Pacific** (Singapore, Tokyo, Hong Kong, Seoul, Sydney, Mumbai, Taiwan)
- `eu`: 🌍 **Europe** (Frankfurt, London, Paris, Helsinki, Madrid)
- `na`: 🌎 **North America** (US-East N. Virginia, US-West Oregon/California, US-Central Texas)
- `sa`: 🌎 **South America** (São Paulo Brazil, Santiago Chile)
- `mena`: 🌍 **Middle East & Africa** (Bahrain, Dubai, Johannesburg)

---

## 2. 🎮 Supported Games & Global Server Networks (105 Clusters) 🌐

### 🎯 A. Valorant (Riot Games)
- 🏢 **Infrastructure**: Hybrid deployment of **Riot Direct** (AS6507) and AWS game server clusters.
- 🌐 **Regions (5 Continents)**:
  - `apac-sg`: Singapore (`ap-southeast-1` & AS6507)
  - `apac-jp`: Tokyo (`ap-northeast-1`)
  - `apac-hk`: Hong Kong (`ap-east-1`)
  - `apac-in`: Mumbai (`ap-south-1`)
  - `apac-kr`: Seoul (`ap-northeast-2`)
  - `apac-au`: Sydney (`ap-southeast-2`)
  - `eu-frankfurt`: Frankfurt (`eu-central-1`)
  - `eu-london`: London (`eu-west-2`)
  - `eu-paris`: Paris (`eu-west-3`)
  - `eu-madrid`: Madrid (`eu-south-2`)
  - `na-east`: US-East (N. Virginia `us-east-1`)
  - `na-west`: US-West (Oregon `us-west-2`)
  - `sa-brazil`: São Paulo (`sa-east-1`)
  - `mena-bahrain`: Bahrain (`me-central-1`)
- 🔍 **Processes Monitored**: `VALORANT-Win64-Shipping.exe`, `RiotClientServices.exe`, `VALORANT.exe`.

### 💣 B. Counter-Strike 2 (CS2) & Dota 2 (Valve)
- 🏢 **Infrastructure**: Valve Corporation Autonomous System (AS32590) with **Steam Datagram Relay (SDR)** edge gateways.
- 🌐 **Regions (5 Continents)**:
  - `apac-sg`: Singapore SDR relays (`103.10.124.0/24`, `155.133.254.0/24`, `162.254.197.0/24`)
  - `apac-hk`: Hong Kong SDR relays (`153.254.86.0/24`, `162.254.193.0/24`)
  - `apac-jp`: Tokyo SDR relays (`155.133.239.0/24`, `155.133.245.0/24`)
  - `apac-kr`: Seoul SDR relays (`155.133.234.0/24`, `162.254.196.0/24`)
  - `apac-au`: Sydney SDR relays (`103.10.125.0/24`, `155.133.246.0/24`)
  - `eu-frankfurt`: Frankfurt SDR relays (`155.133.226.0/24`, `155.133.248.0/24`)
  - `eu-london`: London SDR relays (`155.133.230.0/24`, `162.254.197.0/24`)
  - `eu-helsinki`: Helsinki / Stockholm SDR relays (`155.133.242.0/24`)
  - `eu-madrid`: Madrid SDR relays (`155.133.247.0/24`)
  - `na-east`: US-East (Sterling / Virginia `162.254.192.0/24`, `155.133.253.0/24`)
  - `na-west`: US-West (Seattle / Moses Lake `162.254.194.0/24`, `155.133.238.0/24`)
  - `sa-brazil`: São Paulo SDR relays (`155.133.249.0/24`, `162.254.199.0/24`)
- 🔍 **Processes Monitored**: `cs2.exe` and `dota2.exe`.

### 🪂 C. PUBG: BATTLEGROUNDS (Krafton)
- 🏢 **Infrastructure**: Microsoft Azure regional datacenters combined with AWS EC2 game servers.
- 🌐 **Regions**:
  - `apac-sg`: Azure Southeast Asia (`20.24.48.0/20`, `52.139.208.0/20`) & AWS Singapore (`13.212.0.0/15`).
  - `apac-jp`: Tokyo AWS & Azure Japan East (`20.43.64.0/19`).
  - `apac-kr`: Seoul AWS & Azure Korea Central (`20.41.64.0/19`).
  - `eu-frankfurt`: Frankfurt Azure Germany West Central & AWS `eu-central-1`.
  - `eu-ireland`: Dublin Azure North Europe & AWS `eu-west-1`.
  - `na-east`: US-East Azure East US & AWS `us-east-1`.
  - `na-west`: US-West Azure West US & AWS `us-west-2`.
  - `sa-brazil`: São Paulo Azure Brazil South.
- 🔍 **Processes Monitored**: `TslGame.exe` and `TslGame_BE.exe`.

### ⚡ D. Apex Legends (EA / Respawn)
- 🏢 **Infrastructure**: Multiplay (Unity Gaming Services) hosted on Google Cloud Platform and AWS.
- 🌐 **Regions**: Singapore, Tokyo, Taiwan, Sydney, Frankfurt, London, Amsterdam, Oregon, Virginia, Texas, São Paulo.
- 🔍 **Processes Monitored**: `r5apex.exe` and `r5apex_dx12.exe`.

### 🏆 E. The Finals (Embark Studios)
- 🏢 **Infrastructure**: AWS and G-Core high-tickrate arena infrastructure.
- 🌐 **Regions**: Singapore, Tokyo, Frankfurt, London, US-East, US-West, São Paulo.
- 🔍 **Process Monitored**: `Discovery.exe`.

### 🎖️ F. Call of Duty: Warzone / MW3 (Activision)
- 🏢 **Infrastructure**: Activision Demonware global dedicated hosting.
- 🌐 **Regions**: Singapore, Tokyo, Sydney, Frankfurt, London, US-East, US-West, Dallas Texas, São Paulo.
- 🔍 **Processes Monitored**: `cod.exe` and `bootstrapper.exe`.

### 🦅 G. Delta Force: Hawk Ops (Team Jade / Tencent)
- 🏢 **Infrastructure**: Tencent Cloud Global Application Accelerator and AWS regional clusters.
- 🌐 **Regions**: Singapore, Hong Kong, Tokyo, Frankfurt, US-East, US-West.
- 🔍 **Process Monitored**: `DeltaForceClient-Win64-Shipping.exe`.

### 🛡️ H. Overwatch 2 (Blizzard Entertainment)
- 🏢 **Infrastructure**: Blizzard Defense Matrix regional battle.net clusters.
- 🌐 **Regions**: Singapore, Taiwan, Korea, Japan, Australia, Frankfurt, Paris, US-East, US-West, São Paulo.
- 🔍 **Process Monitored**: `Overwatch.exe`.

### 🧱 I. Rainbow Six Siege (Ubisoft)
- 🏢 **Infrastructure**: Ubisoft dedicated servers on Microsoft Azure.
- 🌐 **Regions**: Singapore (SEAU), Japan East, Australia East, North Europe, West Europe, US-East, US-Central, US-West, Brazil South.
- 🔍 **Processes Monitored**: `RainbowSix.exe`, `RainbowSix_Vulkan.exe`.

### ⚔️ J. League of Legends (LoL) & Teamfight Tactics (TFT) (Riot Games)
- 🏢 **Infrastructure**: Riot Direct global private backbone network (AS6507).
- 🌐 **Regions**: Vietnam (VNG), Singapore (Riot Direct), Taiwan (TPE), Korea (KR), Japan (JP), Europe West (EUW), Europe Nordic & East (EUNE), North America (NA), Brazil (BR).
- 🔍 **Processes Monitored**: `LeagueClientUx.exe`, `League of Legends.exe`, `RiotClientServices.exe`.

---

## 3. 🔬 Adding Custom Games & Discovering CIDRs 🔍

Players and developers can add new games through two methods:

### 🌟 Method A: Dynamic In-App Registration (Zero Coding)
1. Open the Web Dashboard at `http://127.0.0.1:18888`.
2. Click **+ Add Custom Game** in the cockpit.
3. Enter your game's process name (e.g. `CustomGame.exe`) and paste the regional IP CIDRs.
4. Custom games are saved immediately to local storage and become selectable for split-tunneling.

### 🛠️ Method B: Permanent Repository Contribution
To permanently contribute a game profile, complete regional CIDR map, and vector SVG insignia to the official Lagvex distribution:
👉 **Follow our exhaustive developer walkthrough in [🤝 CONTRIBUTING.md](../CONTRIBUTING.md)**, which covers:
- Live match packet capture with Resource Monitor, Wireshark, and PowerShell.
- Cloud ASN lookups (Valve SDR, Riot Direct, AWS, Azure, GCP, Tencent Cloud).
- Subnet isolation rules to prevent web traffic hijacking.
- SVG emblem formatting and Go test validation.

---

🔗 **Navigation**:
- 🏠 [**Project README**](../README.md)
- 🤝 [**Contributor Guide**](../CONTRIBUTING.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- ⚖️ [**Legal Disclaimer**](DISCLAIMER.md)
