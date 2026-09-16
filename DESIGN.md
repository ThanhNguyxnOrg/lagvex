# Lagvex Pro — Design System & Visual Specification

<!-- impeccable:design-schema 1 -->

## Direction & Architecture

A calm-dark, esports engineering cockpit adhering directly to the custom Figma desktop layout (`Lagvex UI Design.zip`). Built upon the official Lagvex crystal-cut emblem and oscilloscope identity, the layout features a 2-column desktop window (1280px max width) engineered to sit cleanly alongside competitive gameplay without visual distraction or ocular fatigue.

## Visual Anchor

- **Primary Brand Mark**: `assets/logo.png` — An angular crystal-cut "L" monogram embedded with a lightning-fast telemetry pulse wave, rendered in electric cyan (`#00F0FF`) and electric violet (`#8B5CF6`).
- **Brand Wordmark**: `LAGVEX` set in `Outfit` 800 with letter-spacing `0.04em`, paired with a high-contrast `PRO` micro-badge.

## Color Tokens

### Surfaces
- `--color-base`: `#0B0E14` — Deep obsidian canvas.
- `--color-base-elev`: `#0E131C` — Elevated window background.
- `--color-surface`: `#121722` — Deep card surface.
- `--color-surface-2`: `#161D2A` — Hover surface and elevated card highlight.
- `--color-line`: `#232B3B` — Defined container edges and dividers.
- `--color-line-soft`: `#1B2230` — Subtle interior separators.

### Brand Accents
- `--color-cyan`: `#00F0FF` — Primary action, focus borders, active hero boost button.
- `--color-cyan-glow`: `rgba(0, 240, 255, 0.15)` — Active button halo and focus indicators.
- `--color-emerald`: `#00E599` — Differential latency improvement badge (`-Xms IMPROVED`), active tunnel status, and packet loss indicators.
- `--color-amber`: `#FFB020` — Jitter warnings and relay discovery hints.
- `--color-rose`: `#FF5D73` — Error states and critical packet loss.

### Typography Contrast
- `--color-ink`: `#EEF2F8` — Primary copy, game headings, and large metric numerals (Contrast Ratio $\approx 15:1$).
- `--color-ink-dim`: `#CBD5E1` — Subtitles, publisher notes, and dropdown options (Passes WCAG AAA).
- `--color-ink-faint`: `#94A3B8` — Micro-labels, units, and secondary hints (Passes WCAG AA).

## Typography Scale

- **Brand & Headings**: `Outfit`, sans-serif (`font-weight: 700 / 800 / 900`).
  - Hero Game Title: `2.375rem` (38px), bold, tight letter-spacing.
  - Section Headings: `1.25rem` (20px), semi-bold.
  - Window Wordmark: `1.125rem` (18px), heavy 800.
- **UI Body & Functional Text**: `Sora`, sans-serif (`font-weight: 400 / 500 / 600`).
  - Base UI text: `0.875rem` (14px).
  - Field labels & hints: `0.75rem` (12px), uppercase / bold.
  - Floor rule: strictly $\ge 12$px for all body copy and interactive labels.
- **Telemetry Numerals**: `JetBrains Mono`, monospace (`font-variant-numeric: tabular-nums`).
  - Large ping numbers: `2.0rem` (32px), bold, tabular digits to prevent layout shift.
  - Telemetry metrics & units: `0.875rem` to `1.0rem` (14px – 16px).

## 2-Column Layout Components

### 1. Window Header
- Crystal-cut "L" logo badge (`assets/logo.png`, 32x32px) + `LAGVEX PRO`.
- Global engine status pill (`Standby` / `Accelerating`) with glowing green beacon.
- Header actions: `Squad` tunnel sharing and `+ Custom VPS` modal triggers.

### 2. Left Column: Hero Game Stage & Library
- **Hero Card**:
  - Full-bleed game backdrop artwork with horizontal and vertical directional scrims.
  - `Anti-Cheat Safe` shield badge + `SELECTED GAME` micro-label.
  - 38px bold game title + studio & genre details.
  - Dropdown selectors for `Target Region` and `Relay Node` with live `Singapore #1 — 18ms • OPTIMAL` star hint.
  - 176x104px **ACTIVATE BOOST** button transitioning smoothly to emerald green **STOP BOOST**.
- **Game Library**:
  - Header with game count pill, search bar, and `+ Custom` profile creator.
  - Category filter pills (*All Games, Tactical FPS, Battle Royale, MOBA*).
  - 4-column Aspect 3/4 Poster Card grid featuring cover art, anti-cheat safe badges, titles, and `ACTIVE` state outlines.

### 3. Right Column: Telemetry Cockpit Sidebar (360px)
- **Differential Latency Card**: Compares `Direct ISP` baseline (58ms) with `Lagvex Accel.` (18ms–22ms) alongside a vivid emerald green delta badge (`−40ms IMPROVED`).
- **Stability Card**: Real-time Triplet grid showing `P50` median ping, `P95 Tail` spike guard, and `Loss %`.
- **Live Throughput**: Real-time Download and Upload traffic with a 12-bar animated cyan equalizer visualizer.
- **Active Routes & Multi-Path Relay**: Visual indicator of `24 / 24` active subnets with 3 animated multi-path relay pips.
- **Route Breadcrumb**: Live routing trajectory footer (`Route: Asia-Pacific → Singapore #1`).

## Craft Floor Verification

- **Impeccable Detect Audit**: 0 anti-patterns, 0 advisories (`impeccable detect web/index.html web/style.css web/app.js`).
- **Typography & Contrast**: Compliant with WCAG AA/AAA standards; zero sub-11px text.
- **Performance & Polish**: Hardware-accelerated CSS transforms, zero layout shifts on telemetry updates, non-elevated preview server via `preview_server.py`.
