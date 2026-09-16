# Design System & Visual Specification

<!-- impeccable:design-schema 1 -->

## Direction

Calm-dark esports engineering cockpit. Razor-sharp, high-contrast, distraction-free visual environment designed to run alongside fullscreen competitive multiplayer games without causing ambient glare or visual fatigue. Built upon the official Lagvex crystal-cut emblem and oscilloscope wave identity.

## Visual Anchor

- **Primary Brand Mark**: `assets/logo.png` — An angular crystal-cut "L" monogram embedded with a high-voltage lightning bolt and telemetry pulse wave, rendered in electric cyan (`#00F0FF`) and electric violet (`#8B5CF6`).
- **Brand Wordmark**: `LAGVEX` set in `Outfit` Heavy 800 with letter-spacing `0.04em` paired with a high-contrast `PRO` micro-badge.

## Color Tokens

### Surfaces
- `bg-canvas`: `#090B10` — Pure deep matte obsidian canvas; prevents ocular fatigue in dim esports environments.
- `bg-card`: `#111520` — Baseline card container elevation with crisp structural delineation.
- `bg-card-elevated`: `#161C2A` — Hover and focused container surface.
- `bg-input`: `#131722` — Deep neutral well for select boxes and text fields.
- `bg-input-hover`: `#192030` — Interactive input highlight.

### Borders
- `border-subtle`: `rgba(255, 255, 255, 0.08)` — Hairline structural separation.
- `border-card`: `#1F2738` — Defined card container edge.
- `border-interactive`: `#2A344C` — Unfocused interactive element outline.
- `border-hover`: `rgba(0, 240, 255, 0.40)` — High-clarity hover feedback.
- `border-active`: `#00F0FF` — Focused and active element border.

### Brand Accents
- `accent-cyan`: `#00F0FF` — Primary action, primary focus states, live telemetry metrics.
- `accent-cyan-subtle`: `rgba(0, 240, 255, 0.12)` — Badge wells and selected state backgrounds.
- `accent-violet`: `#8B5CF6` — Secondary brand facet derived from the logo's crystal geometry.
- `accent-green`: `#00E599` — Low-latency RTT indicator, active tunnel engagement, and differential gain.
- `accent-amber`: `#F59E0B` — Network jitter warning and prober timeouts.
- `accent-red`: `#EF4444` — Driver disconnect or unreachable relay alert.

### Text Contrast
- `text-primary`: `#F8FAFC` — Primary copy and metric numbers (Contrast Ratio $\approx 16:1$).
- `text-secondary`: `#CBD5E1` — Subtitles, process names, and input labels (Contrast Ratio $\approx 10:1$, passes WCAG AAA).
- `text-muted`: `#94A3B8` — Secondary notes and measurement units (Contrast Ratio $\approx 6.5:1$, passes WCAG AA).
- `text-on-accent`: `#090B10` — High-contrast black text on electric cyan button surfaces.

## Typography Ramp

- **Brand & Headings**: `Outfit`, sans-serif (`font-weight: 700 / 800`).
  - Hero Title: `1.625rem` (26px), line-height `1.2`.
  - Section Headings: `1.375rem` (22px), line-height `1.3`.
  - Modal Titles: `1.125rem` (18px), line-height `1.3`.
- **UI Body & Functional Text**: `Sora`, sans-serif (`font-weight: 400 / 500 / 600`).
  - Base UI text: `0.9375rem` (15px), line-height `1.5`.
  - Input labels: `0.8125rem` (13px), font-weight `600`.
  - Secondary meta: `0.75rem` (12px), line-height `1.35`.
  - Floor rule: strictly $\ge 11$px for functional text.
- **Telemetry & Numerals**: `JetBrains Mono`, monospace (`font-variant-numeric: tabular-nums`).
  - Differential latency values: `2.0rem` (32px), bold, tabular numerals to prevent layout shift during live polling.
  - Telemetry counters: `0.875rem` (14px), font-weight `600`.

## Key Component Patterns

### 1. Header Navigation
A sticky, high-performance banner featuring:
- Official brand avatar (`assets/logo.png`, 36x36px with 1px border).
- Global tunnel engine status chip (`Standby` / `Accelerating`) with semantic status dot.
- Probed RTT counter with cyan tabular numerals.
- Action triggers: "Join Squad" and "Add Custom VPS".

### 2. Differential Latency HUD
A 3-column comparative telemetry grid that makes the accelerator's value immediately obvious:
- **Direct ISP Ping**: Measured baseline (e.g. `68 ms`) under default public transit.
- **Accelerated Ping**: Probed RTT (e.g. `24 ms`) via ChaCha20-Poly1305 encrypted relay tunnel.
- **Net Latency Gain**: Computed differential (e.g. `-44 ms (-64%)`) in high-visibility emerald green.

### 3. Route Hop Pipeline
A high-contrast visual trajectory showing the 4 hops:
`Client Machine (WinTun Kernel)` ➔ `ChaCha20-Poly1305 Tunnel` ➔ `Relay Node (Live RTT)` ➔ `Subsea Fiber` ➔ `Game Cloud Cluster`.

### 4. Game Library Grid
A Bento grid displaying supported game titles with official vector SVG insignias, publisher details, server cluster counts, and instant filter pills.

## Craft Floor Verification

- **Anti-patterns**: 0 (audited via `impeccable detect`).
- **Advisories**: 0 (neutral elevation shadows, clean 1px borders, zero AI glow slop).
- **Accessibility**: 100% WCAG AA compliant on all interactive text and controls.
