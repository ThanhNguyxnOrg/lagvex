# 🤝 Lagvex Contributor & Maintainer Guide 🚀

> **Comprehensive Developer Handbook for Adding Games, Mapping Server Networks, and Submitting Code**  
> 🔗 [Back to Project README.md](README.md) | [Architecture Overview](docs/ARCHITECTURE.md) | [Game Profiles Guide](docs/PROFILES.md) | [Legal Disclaimer](docs/DISCLAIMER.md)

---

## 📖 Table of Contents 📑

- [🎯 1. Welcome & Core Engineering Invariants](#-1-welcome--core-engineering-invariants)
- [🎮 2. Step-by-Step: How to Add a New Game](#-2-step-by-step-how-to-add-a-new-game)
  - [🔍 Step 2.1: Capture Live Match Network Traffic](#-step-21-capture-live-match-network-traffic)
  - [🏢 Step 2.2: Identify ASN & Datacenter Providers](#-step-22-identify-asn--datacenter-providers)
  - [🎯 Step 2.3: Narrow Subnet CIDRs (The Non-Interference Rule)](#-step-23-narrow-subnet-cidrs-the-non-interference-rule)
  - [📋 Step 2.4: Register in `configs/profiles.json`](#-step-24-register-in-configsprofilesjson)
  - [🎨 Step 2.5: Add Vector SVG Insignia in `web/app.js`](#-step-25-add-vector-svg-insignia-in-webappjs)
  - [🧪 Step 2.6: Validate with Automated Tests](#-step-26-validate-with-automated-tests)
- [🌐 3. Step-by-Step: Adding or Updating Relay Nodes](#-3-step-by-step-adding-or-updating-relay-nodes)
- [🧪 4. Local Development & Quality Gates](#-4-local-development--quality-gates)
- [📬 5. Pull Request (PR) Guidelines & Checklist](#-5-pull-request-pr-guidelines--checklist)
- [⚖️ 6. Nominative Fair Use & Legal Compliance](#️-6-nominative-fair-use--legal-compliance)

---

## 🎯 1. Welcome & Core Engineering Invariants 💡

Thank you for contributing to **Lagvex**! Whether you are an open-source contributor adding your favorite competitive shooter or a core maintainer expanding global relay clusters, this guide provides the exact procedures to follow.

### 🛡️ Non-Negotiable Invariants:
1. **Zero Memory Tampering**: We **never** read, write, or detour game process memory (`ReadProcessMemory`, `WriteProcessMemory`, DLL injection).
2. **Zero Socket Hooking**: We **never** hook game sockets with WinDivert, LSP, or WFP callouts that could trigger anti-cheat heuristic bans.
3. **Strict Route Isolation**: We **only** route game server IP subnets through the tunnel. Discord, Spotify, browser downloads, and local network traffic must remain 100% on the user's direct physical network adapter.
4. **Narrow Subnet Sizing**: Subnets must be kept tight (typically `/20` down to `/24`). Never add wide supernets like `/8` or `/12`.
5. **Clean Code & Fast CI**: Every PR must pass `golangci-lint run` and `go test ./...` in under 60 seconds with zero warnings.

---

## 🎮 2. Step-by-Step: How to Add a New Game 🏆

Adding a new game to Lagvex is simple and requires **no low-level driver programming**. You only need to discover the game's server destination CIDRs, update `configs/profiles.json`, and add an SVG emblem to `web/app.js`.

```text
  ┌─────────────────────────┐
  │ 1. Play Live Match      │ ➔ Use Resource Monitor / Wireshark to record destination UDP IP
  └────────────┬────────────┘
               ▼
  ┌─────────────────────────┐
  │ 2. Query BGP / Whois    │ ➔ Identify ASN (Valve SDR, Riot Direct, AWS, Azure, Tencent)
  └────────────┬────────────┘
               ▼
  ┌─────────────────────────┐
  │ 3. Narrow CIDR Subnets  │ ➔ Cross-reference cloud provider range tables (/20 - /24)
  └────────────┬────────────┘
               ▼
  ┌─────────────────────────┐
  │ 4. Update JSON & SVG    │ ➔ Add to configs/profiles.json and web/app.js
  └────────────┬────────────┘
               ▼
  ┌─────────────────────────┐
  │ 5. Run go test ./...    │ ➔ Confirm clean profile parsing and zero lint errors
  └─────────────────────────┘
```

---

### 🔍 Step 2.1: Capture Live Match Network Traffic 📡

To find where a game's game-state tick packets travel, capture traffic while connected to an active competitive server:

#### Method A: Windows Resource Monitor (Easiest — No Tools Needed)
1. Press `Win + R`, type `resmon.exe`, and press **Enter**.
2. Switch to the **Network** tab.
3. Expand **Processes with Network Activity** and check the box next to your game's executable (e.g., `Marvel-Win64-Shipping.exe`).
4. Expand **Network Activity**.
5. Look for the remote destination address with sustained traffic (typically sending 15–50 packets per second with ~20–60 KB/s transfer rate). Note down the **Remote IP** and **Port**.

#### Method B: Wireshark / TShark
Capture UDP packets filtered by the game process:
```text
udp and ip.dst != 192.168.0.0/16 and ip.dst != 10.0.0.0/8 and not broadcast
```
Look for high-frequency, non-fragmented datagrams exchanged between your PC and a cloud hosting provider.

#### Method C: PowerShell (Administrator)
```powershell
Get-NetUDPEndpoint | Select-Object LocalAddress, LocalPort, OwningProcess | Get-Process -Id { $_.OwningProcess }
```

---

### 🏢 Step 2.2: Identify ASN & Datacenter Providers 🌐

Once you have a candidate server IP (for example, `13.250.45.10`):

1. Run `whois` or query an online BGP intelligence service (e.g. [BGP.he.net](https://bgp.he.net) or [ipinfo.io](https://ipinfo.io)):
   ```bash
   whois 13.250.45.10 | grep -E "OriginAS|NetRange|CIDR|OrgName"
   ```
2. Note the Autonomous System (AS):
   - **AS6507**: Riot Games (Riot Direct)
   - **AS32590**: Valve Corporation (Steam Datagram Relay)
   - **AS16509 / AS14618**: Amazon Web Services (AWS EC2 / Multiplay)
   - **AS8075**: Microsoft Azure
   - **AS15169**: Google Cloud Platform (GCP)
   - **AS13238**: Yandex / G-Core Labs
   - **AS132203**: Tencent Cloud

---

### 🎯 Step 2.3: Narrow Subnet CIDRs (The Non-Interference Rule) 📏

> [!WARNING]
> **CRITICAL RULE**: Do NOT add the publisher's entire `/12` or `/16` network!  
> For instance, adding `13.0.0.0/8` would drag all of Amazon Web Services through the player's relay, breaking Amazon Shopping, Prime Video, and unrelated web apps.

1. **Locate the specific regional cluster prefix**:
   - In AWS `ip-ranges.json`, search for the region (e.g., `ap-southeast-1` for Singapore, `eu-central-1` for Frankfurt) under `service: "EC2"`.
   - Prefer subnets between `/18` and `/24` (e.g., `13.250.0.0/15`, `54.251.0.0/16`, `155.133.254.0/24`).
2. Verify that the subnet CIDRs cleanly cover all servers in that regional match pool without overlapping private LAN blocks (`10.0.0.0/8`, `192.168.0.0/16`).

---

### 📋 Step 2.4: Register in `configs/profiles.json` 📝

Open [`configs/profiles.json`](configs/profiles.json) and append your new game entry under the `"games"` array:

```json
{
  "id": "game_slug",
  "name": "Game Display Title",
  "category": "Tactical FPS",
  "icon": "🎯",
  "processNames": [
    "GameBinary-Win64-Shipping.exe",
    "GameBinary.exe",
    "GameLauncher.exe"
  ],
  "regions": [
    {
      "id": "apac-sg",
      "continent": "apac",
      "name": "Singapore (SEA Cluster)",
      "source": "AWS ap-southeast-1 & Edge Relays",
      "cidrs": [
        "13.250.0.0/15",
        "18.140.0.0/15"
      ]
    },
    {
      "id": "eu-frankfurt",
      "continent": "eu",
      "name": "Frankfurt (EU Central)",
      "source": "AWS eu-central-1",
      "cidrs": [
        "3.120.0.0/14",
        "18.194.0.0/15"
      ]
    }
  ]
}
```

#### Field Specifications:
- `id`: Unique lowercase identifier with underscores (e.g. `marvel_rivals`, `deadlock`).
- `category`: Genre label displayed on game card tags (`Tactical FPS`, `Battle Royale`, `Hero Shooter`, `MOBA`, `Arena FPS`).
- `icon`: Fallback Unicode emoji.
- `processNames`: Array of exact executable names. The **Smart Process Watcher** uses these to automatically inject routes when the game starts and remove them when the game exits.
- `regions`: Regional server groupings. Must include:
  - `continent`: Valid continent code: `"apac"`, `"eu"`, `"na"`, `"sa"`, or `"mena"`. This organizes the UI server dropdown into clean `<optgroup>` clusters!
  - `name`: Clean, human-readable label with city and cloud details.
  - `cidrs`: Array of valid IPv4 CIDR strings.

---

### 🎨 Step 2.5: Add Vector SVG Insignia in `web/app.js` 🖌️

Lagvex uses crisp, resolution-independent SVG emblems rather than blurry raster bitmaps or generic buttons:

1. Open [`web/app.js`](web/app.js).
2. Locate the function `getGameLogoSvg(gameId)`.
3. Add a new `case "game_slug":` block returning your inline SVG markup:

```javascript
case "game_slug":
  return `
    <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Game Title">
      <rect width="100" height="100" rx="14" fill="#0e121a"/>
      <!-- Clean vector paths here with vibrant accents -->
      <path d="M50 15 L85 80 L15 80 Z" fill="#00F0FF"/>
      <text x="50" y="70" text-anchor="middle" font-family="'Outfit', sans-serif" font-weight="900" font-size="16" fill="#FFF">GAME</text>
    </svg>`;
```

#### SVG Guidelines:
- `viewBox`: Always use `0 0 100 100`.
- Class: Include `class="game-svg-logo"`.
- Style: Use rounded squircle backgrounds (`rx="14"`) and colors consistent with the cyber/esports palette (`#00F0FF`, `#00FF88`, `#FF0055`, `#DE9B35`, `#FF4655`).
- Keep markup lightweight (< 1.5 KB per SVG).

---

### 🧪 Step 2.6: Validate with Automated Tests 🔍

Run the test suite to verify your new game profile:

```bash
# In project root:
go test -v ./pkg/profiles/...
```

The test runner will verify:
- `configs/profiles.json` parses successfully with valid JSON syntax.
- All CIDR strings are mathematically valid IPv4 subnets.
- Process names and region mappings are non-empty.

To test with the web dashboard, start the client:
```bash
go run ./cmd/lagvex-client
```
Open `http://127.0.0.1:18888` in your browser. Verify that your game card appears, the SVG renders cleanly, and the server dropdown displays your server regions organized under their respective continent headers.

---

## 🌐 3. Step-by-Step: Adding or Updating Relay Nodes 🛰️

To contribute a new community relay node or update private endpoints:

1. Open [`configs/profiles.json`](configs/profiles.json).
2. Find the `"relays"` array.
3. Add or modify a relay entry:

```json
{
  "id": "relay-syd-1",
  "name": "🇦🇺 Sydney #1 [Community Free]",
  "location": "Sydney, Australia",
  "endpoint": "203.0.113.50:51820",
  "psk": "community_free_access_syd1"
}
```

- `endpoint`: Public IP or DDNS hostname with UDP listening port (`IP:PORT`).
- `psk`: 32-character Pre-Shared Key.
- `location`: Flag emoji + City, Country.

---

## 🧪 4. Local Development & Quality Gates 🛠️

Before opening a pull request, verify that your code adheres to all quality standards:

### 1. Run Unit Tests
```bash
go test -race -v ./...
```

### 2. Run Quality Linter
```bash
golangci-lint run
```
*(Or verify with `go vet ./...`)*

### 3. Check Git Status
Ensure no temporary test binaries or scratch files are untracked:
```bash
git status
```

---

## 📬 5. Pull Request (PR) Guidelines & Checklist 📋

When submitting your contribution:

### 🌿 Branch Naming
- New Game: `feature/add-game-<game_id>` (e.g. `feature/add-game-marvel-rivals`)
- New Relay: `feature/add-relay-<region>` (e.g. `feature/add-relay-sydney`)
- Bug Fix: `fix/<issue-description>` (e.g. `fix/wintun-pointer-cast`)
- Documentation: `docs/<topic>` (e.g. `docs/update-contributing-guide`)

### 💬 Commit Message Format
We follow [Conventional Commits](https://www.conventionalcommits.org/):
```text
feat(profiles): add Marvel Rivals with 4 global server clusters
fix(client): prevent duplicate route addition on reconnect
docs(contributing): clarify Wireshark CIDR capture instructions
```

### ✅ PR Submission Checklist
- [ ] Added game entry to `configs/profiles.json` with valid `processNames` and `continent` tags.
- [ ] Verified destination CIDRs are narrow (`/18` to `/24`) and confirmed via live match packet captures.
- [ ] Added vector SVG emblem into `web/app.js` with `viewBox="0 0 100 100"`.
- [ ] Ran `go test ./...` and confirmed all tests pass.
- [ ] Ran `golangci-lint run` (or `go vet ./...`) with zero errors.
- [ ] Added notes in `CHANGELOG.md` under `## [Unreleased]`.
- [ ] Confirmed compliance with Nominative Fair Use guidelines below.

---

## ⚖️ 6. Nominative Fair Use & Legal Compliance 📜

Lagvex is committed to total legal transparency and intellectual property respect:

1. **Nominative Fair Use**: All game titles, trademarks, and logos are referenced strictly to identify destination game executables and server routing clusters. Lagvex is an independent project not affiliated with or endorsed by any game publisher.
2. **Anti-Cheat Non-Tampering**: Under no circumstances will PRs containing process injection, DLL detouring, memory scraping, or packet payload mutation be accepted.
3. **No Closed-Source Bloat**: All contributions must be 100% open-source under the [MIT License](LICENSE).

---

🔗 **Navigation**:
- 🏠 [**Project README**](README.md)
- 📐 [**Architecture Overview**](docs/ARCHITECTURE.md)
- 📡 [**Protocol Specification**](docs/PROTOCOL.md)
- 🚀 [**Deployment Guide**](docs/DEPLOYMENT.md)
- 🌐 [**Game Profiles & CIDRs**](docs/PROFILES.md)
- 🗺️ [**Engineering Roadmap**](docs/ROADMAP.md)
- ⚖️ [**Legal Disclaimer**](docs/DISCLAIMER.md)
