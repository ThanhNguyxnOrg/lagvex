# 💸 Running Cost: Free vs. Paid Boosters 💰

> 💡 **Related Documents**: [📖 Back to README](../README.md) • [🌐 Why a VPS is Required](WHY_VPS.md) • [🚀 Deployment Guide](DEPLOYMENT.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)

---

## 🎯 The Short Answer 📌

| | Commercial boosters | Lagvex |
|---|---|---|
| Software | 💵 subscription (see below) | ✅ **100% free, open source** |
| Server network | 💵 included, their VPS fleet | 🫵 **your own VPS** (or a friend's) |
| Typical total | **~$6.50–$12 / month, forever** | **$0 – $6 / month**, cancel anytime |

Lagvex never charges for the client, never gates features behind a plan, and
never asks for a payment method. The **only** optional cost is the relay VPS —
and that is the part you control.

> ℹ️ Competitor prices are **publicly listed subscription prices** observed
> during the 2026 design study, and may change; they are listed only as a size
> reference, not as an offer. The dashboard compares against an illustrative
> **$9/month** reference plan, which is a constant in the UI, not a quote.

---

## 📊 What Gaming Traffic Actually Costs 📉

An accelerated game session is tiny compared to streaming:

| Activity | Data used |
|---|---|
| Video streaming (1080p) | ~3 GB / hour |
| **Accelerated game session** | **~20–40 MB / hour** (planning estimate) |

The 20–40 MB/hour figure is a **planning estimate**, not a measured guarantee:
it depends on the game, the packet rate and how chatty the session is. The
dashboard therefore prefers **your own measured rate** — once a session has
produced at least five samples, the COST & SELF-HOST panel projects the month
from the measured byte rate; before that it falls back to the documented
typical and labels the number as such.

Even a heavy player (4 hours every day) transfers **well under 5 GB/month**
through the relay. Every mainstream VPS includes **500 GB – 4 TB** of monthly
transfer, so the transfer is effectively free — you are only renting a tiny
server's presence in the right city.

---

## 🖥️ Option 1: BYO-VPS (the standard path) 🛰️

One VPS near the game's datacenter, one `lagvex-relay` binary, done:

| Region (example game servers) | Typical smallest VPS | Notes |
|---|---|---|
| Singapore / Tokyo / Seoul | **$3 – $6 / month** | 1 vCPU, 512 MB–1 GB RAM is plenty |
| US East / US West | **$3 – $6 / month** | same |
| Frankfurt / London / Paris | **$3 – $6 / month** | same |
| **Free tiers** | **$0 / month** | e.g. Oracle Cloud Always Free (ARM, ~10 TB egress) — availability varies by region and capacity |

Requirements are deliberately modest (`docs/DEPLOYMENT.md`): 1 vCPU, 512 MB
RAM, 10 GB disk. The relay is a **single static binary** — no Docker required,
no database, no control plane.

---

## 👥 Option 2: Share one relay with friends (Squad math) 🧮

`lagvex-relay -max-clients N` accepts several players on the **same** VPS with
the **same pre-shared key** — treat that PSK like a group secret. Game traffic
per player is small (table above), so a small box is usually enough for a
group; how many sessions fit depends on your uplink and packet rate, so test
before a tournament.

| Squad size | $5 VPS per person |
|---|---|
| 1 | $5.00 |
| 2 | $2.50 |
| 3 | ~$1.67 |
| 5 | **$1.00** |

The dashboard's **Squad** tab coordinates rooms; the cost split is between
humans, not software.

---

## 🧠 Option 3: Need a relay at all? Measure first 🔬

Sometimes the honest answer is "not today":

- **Route Detail** compares a *measured* direct regional-anchor RTT with a
  *via-relay estimate* when you refresh it. The estimate is the sum of the
  measured tunnel RTT and the relay→anchor RTT, so it is a region-anchor
  estimate — never presented as a measured game-server ping.
- The optional `/api/advisor` endpoint returns `RouteAdvice` on request; the
  dashboard does not run it on every sweep.
- The **Nodes** tab ranks relays by measured path quality, so a relay that
  does not actually improve your route will not sit at the top of the list.
- Idle sessions time out after 90 s, and dual-path redundancy is **opt-in** with
  measured retransmit counters — you never pay bandwidth for a feature you did
  not enable.

---

## 🧾 What you do NOT pay for with Lagvex 🚫💰

- ❌ No subscription, no plan tiers, no "premium routes"
- ❌ No per-GB traffic billing (Mudfish-style)
- ❌ No feature paywall: dual-path redundancy, squad, diagnostics, route detail
  and the prober are the same build for everyone
- ❌ No account, no telemetry sold, no payment method collected

---

## 🚀 Quick Start (3 commands on a fresh Ubuntu VPS) ⚡

```bash
# 1. Download the relay binary from the GitHub releases page
wget https://github.com/ThanhNguyxnOrg/lagvex/releases/latest/download/lagvex-relay-linux-amd64 -O lagvex-relay
chmod +x lagvex-relay

# 2. Generate a strong pre-shared key and run it
openssl rand -hex 32 > lagvex.psk
sudo ./lagvex-relay -listen :51820 -psk-file lagvex.psk

# 3. In the Lagvex client: Nodes → + Add Custom Node →
#    endpoint = your-vps-ip:51820, PSK = the contents of lagvex.psk
```

A systemd unit and a Docker Compose example are in
[docs/DEPLOYMENT.md](DEPLOYMENT.md).

---

## ⚖️ Honest Limits 🧯

- A $3.50 VPS gives you **one city**. Commercial boosters sell many cities with
  capacity headroom — that convenience is exactly what their subscription pays
  for. Lagvex trades that convenience for control and price.
- Free-tier instances can be reclaimed by the provider at any time; treat them
  as a lab, not as a tournament-grade dependency.
- Lagvex publishes no infrastructure of its own and never will bundle a paid
  one into the client.
- **Relay hosting is Linux-only** and needs `/dev/net/tun`. Windows and macOS
  can run the client, but they cannot host a relay.
- The traffic figure is a planning estimate and the VPS prices are market
  estimates; the only measured numbers in the panel are your own session's
  byte rate.
