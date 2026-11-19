# 💸 Running Cost: Free vs. Paid Boosters 💰

> 💡 **Related Documents**: [📖 Back to README](../README.md) • [🌐 Why a VPS is Required](WHY_VPS.md) • [🚀 Deployment Guide](DEPLOYMENT.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)

---

## 🎯 The Short Answer 📌

| | Commercial boosters | Lagvex |
|---|---|---|
| Software | 💵 subscription (see below) | ✅ **100% free, open source** |
| Server network | 💵 included, their VPS fleet | 🫵 **your own VPS** (or a friend's) |
| Typical total | **~$6.50–$12 / month, forever** | **$0 – $5 / month**, cancel anytime |

Lagvex never charges for the client, never gates features behind a plan, and
never asks for a payment method. The **only** optional cost is the relay VPS —
and that is the part you control.

> ℹ️ Competitor prices below are their **publicly listed** subscription prices,
> observed during the design study, and may change. They are listed only as a
> size reference, not as an offer.

---

## 📊 What Gaming Traffic Actually Costs 📉

An accelerated game session is tiny compared to streaming:

| Activity | Data used |
|---|---|
| Video streaming (1080p) | ~3 GB / hour |
| **Accelerated game session** | **~20–40 MB / hour** |

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
per player is small (table above), so a 1 vCPU box comfortably serves a group.

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

- The dashboard compares **direct measured RTT vs. best measured relay route**
  on every sweep (`RouteAdvice`). If your direct path is already clean, routing
  through a VPS adds no value — Lagvex tells you so instead of selling you one.
- The relay probe (`via relay: X ms total`) is measured, not estimated, so the
  "is it worth $5?" decision is a number, not a marketing promise.
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

A systemd unit and a Docker run line are in [docs/DEPLOYMENT.md](DEPLOYMENT.md).

---

## ⚖️ Honest Limits 🧯

- A $3.50 VPS gives you **one city**. Commercial boosters sell many cities with
  capacity headroom — that convenience is exactly what their subscription pays
  for. Lagvex trades that convenience for control and price.
- Free-tier instances can be reclaimed by the provider at any time; treat them
  as a lab, not as a tournament-grade dependency.
- Lagvex publishes no infrastructure of its own and never will bundle a paid
  one into the client.
